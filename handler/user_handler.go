package handler

import (
	"context"
	"net/http"
	"api-service/model"
	"api-service/utils"
	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
)

type UserHandler struct {
	DB *bun.DB
}

// Проверка наличия email в базе
func (h *UserHandler) isEmailExist(email string) (bool, error) {
	var user model.User
	err := h.DB.NewSelect().Model(&user).Where("email = ?", email).Limit(1).Scan(context.Background())
	if err != nil && err.Error() != "sql: no rows in result set" {
		return false, err
	}
	return user.Email == email, nil
}

// Создание пользователя
func (h *UserHandler) CreateUser(c echo.Context) error {
	req := new(model.CreateUserRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверные данные"})
	}

	// Проверка наличия email
	exists, err := h.isEmailExist(req.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка проверки email"})
	}
	if exists {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email уже существует"})
	}

	// Хэшируем пароль
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка хэширования пароля"})
	}

	user := &model.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
	}

	// Сохраняем пользователя
	_, err = h.DB.NewInsert().Model(user).Exec(context.Background())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка создания пользователя"})
	}

	return c.JSON(http.StatusCreated, user)
}

// Получение списка пользователей
func (h *UserHandler) GetUsers(c echo.Context) error {
	var users []model.User
	err := h.DB.NewSelect().Model(&users).Scan(context.Background())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка получения списка пользователей"})
	}
	return c.JSON(http.StatusOK, users)
}

// Обновление пользователя
func (h *UserHandler) UpdateUser(c echo.Context) error {
    id := c.Param("id")
    req := new(model.UpdateUserRequest)
    if err := c.Bind(req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверные данные"})
    }

    // Собираем данные для обновления
    query := h.DB.NewUpdate().Model(&model.User{}).Where("id = ?", id)

    if req.Name != "" {
        query.Set("name = ?", req.Name)
    }
    if req.Email != "" {
        exists, err := h.isEmailExist(req.Email)
        if err != nil {
            return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка проверки email"})
        }
        if exists {
            return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email уже существует"})
        }
        query.Set("email = ?", req.Email)
    }
    if req.Password != "" {
        hashedPassword, err := utils.HashPassword(req.Password)
        if err != nil {
            return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка хэширования пароля"})
        }
        query.Set("password = ?", hashedPassword)
    }

    // Выполняем обновление
    _, err := query.Exec(context.Background())
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка обновления пользователя"})
    }

    return c.JSON(http.StatusOK, map[string]string{"message": "Пользователь обновлен"})
}

// Удаление пользователя
func (h *UserHandler) DeleteUser(c echo.Context) error {
	id := c.Param("id")
	_, err := h.DB.NewDelete().Model((*model.User)(nil)).Where("id = ?", id).Exec(context.Background())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка удаления пользователя"})
	}
	return c.NoContent(http.StatusNoContent)
}

