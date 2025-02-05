package handler

import (
	"api-service/model"
	"api-service/utils"
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"

	proto "api-service/proto/auth-service/proto"
	chatpb "api-service/proto/chat-service/proto"
)

type UserHandler struct {
	DB               *bun.DB
	AuthServiceClient proto.AuthServiceClient
	ChatServiceClient chatpb.ChatServiceClient // Добавьте это поле
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
    // Получаем ID текущего пользователя из контекста
    userID, ok := c.Get("user_id").(int)
    if !ok {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Не удалось получить ID пользователя"})
    }

    req := new(model.UpdateUserRequest)
    if err := c.Bind(req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверные данные"})
    }

    // Собираем данные для обновления
    query := h.DB.NewUpdate().Model(&model.User{}).Where("id = ?", userID)

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
    if req.Avatar != "" {
        query.Set("avatar = ?", req.Avatar)
    }
    if req.Status != "" {
        query.Set("status = ?", req.Status)
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
    userID, ok := c.Get("user_id").(int)
    if !ok {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Не удалось получить ID пользователя"})
    }

    ctx := c.Request().Context()

    // Начинаем транзакцию
    tx, err := h.DB.BeginTx(ctx, nil)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка начала транзакции"})
    }
    defer tx.Rollback()

    // Таблицы, связанные с пользователем
    tables := []struct {
        model interface{}
        where string
    }{
        {(*model.PostLike)(nil), "user_id = ?"},
        {(*model.Comment)(nil), "user_id = ?"},
        {(*model.Repost)(nil), "user_id = ?"},
    }

    // Удаляем данные из связанных таблиц
    for _, table := range tables {
        if _, err := tx.NewDelete().Model(table.model).Where(table.where, userID).Exec(ctx); err != nil {
            return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Ошибка удаления из таблицы %T", table.model)})
        }
    }

    // Удаляем посты пользователя и связанные данные
    var postIDs []int
    if err := tx.NewSelect().Column("id").Model((*model.Post)(nil)).Where("user_id = ?", userID).Scan(ctx, &postIDs); err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка получения постов пользователя"})
    }

    if len(postIDs) > 0 {
        relatedTables := []struct {
            model interface{}
            where string
        }{
            {(*model.PostLike)(nil), "post_id IN (?)"},
            {(*model.Comment)(nil), "post_id IN (?)"},
            {(*model.Repost)(nil), "original_post_id IN (?)"},
            {(*model.PostTag)(nil), "post_id IN (?)"},
            {(*model.Media)(nil), "post_id IN (?)"},
        }

        for _, table := range relatedTables {
            if _, err := tx.NewDelete().Model(table.model).Where(table.where, bun.In(postIDs)).Exec(ctx); err != nil {
                return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Ошибка удаления из таблицы %T", table.model)})
            }
        }

        // Удаляем посты пользователя
        if _, err := tx.NewDelete().Model((*model.Post)(nil)).Where("user_id = ?", userID).Exec(ctx); err != nil {
            return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка удаления постов пользователя"})
        }
    }

    // Удаляем самого пользователя
    if _, err := tx.NewDelete().Model((*model.User)(nil)).Where("id = ?", userID).Exec(ctx); err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка удаления пользователя"})
    }

    // Фиксируем транзакцию
    if err := tx.Commit(); err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка фиксации транзакции"})
    }

    return c.NoContent(http.StatusNoContent)
}