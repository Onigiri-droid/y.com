package handler

import (
	"api-service/model"
	"api-service/utils"
	"context"
	"database/sql"
	"errors"
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
	// Получаем ID текущего пользователя из контекста
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
	// Если транзакция не зафиксируется, откатываем изменения
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	// Шаг 1. Агрегируем данные для обновления счетчиков (до удаления записей)
	type countResult struct {
		PostID int `bun:"post_id"`
		Count  int `bun:"cnt"`
	}

	var likeUpdates []countResult
	err = tx.NewSelect().ColumnExpr("post_id, COUNT(*) AS cnt").
		Model((*model.PostLike)(nil)).
		Where("user_id = ?", userID).
		Group("post_id").
		Scan(ctx, &likeUpdates)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка выборки лайков"})
	}

	var repostUpdates []countResult
	err = tx.NewSelect().ColumnExpr("original_post_id AS post_id, COUNT(*) AS cnt").
		Model((*model.Repost)(nil)).
		Where("user_id = ?", userID).
		Group("original_post_id").
		Scan(ctx, &repostUpdates)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка выборки репостов"})
	}

	var commentUpdates []countResult
	err = tx.NewSelect().ColumnExpr("post_id, COUNT(*) AS cnt").
		Model((*model.Comment)(nil)).
		Where("user_id = ?", userID).
		Group("post_id").
		Scan(ctx, &commentUpdates)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка выборки комментариев"})
	}

	// Шаг 2. Обновляем счетчики в постах с учетом найденных данных
	for _, upd := range likeUpdates {
		_, err = tx.NewUpdate().Model((*model.Post)(nil)).
			Set("likes_count = GREATEST(likes_count - ?, 0)", upd.Count).
			Where("id = ?", upd.PostID).
			Exec(ctx)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка обновления счетчика лайков"})
		}
	}

	for _, upd := range repostUpdates {
		_, err = tx.NewUpdate().Model((*model.Post)(nil)).
			Set("reposts_count = GREATEST(reposts_count - ?, 0)", upd.Count).
			Where("id = ?", upd.PostID).
			Exec(ctx)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка обновления счетчика репостов"})
		}
	}

	for _, upd := range commentUpdates {
		_, err = tx.NewUpdate().Model((*model.Post)(nil)).
			Set("comments_count = GREATEST(comments_count - ?, 0)", upd.Count).
			Where("id = ?", upd.PostID).
			Exec(ctx)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка обновления счетчика комментариев"})
		}
	}

	// Шаг 3. Удаляем записи из связанных таблиц, где пользователь является автором
	associatedTables := []struct {
		model interface{}
		query string
	}{
		{&model.PostLike{}, "user_id = ?"},
		{&model.Comment{}, "user_id = ?"},
		{&model.Repost{}, "user_id = ?"},
	}
	for _, table := range associatedTables {
		if _, err = tx.NewDelete().Model(table.model).Where(table.query, userID).Exec(ctx); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("Ошибка удаления данных из таблицы %T: %v", table.model, err),
			})
		}
	}

	// Шаг 4. Получаем ID всех постов пользователя
	var postIDs []int
	err = tx.NewSelect().Column("id").
		Model((*model.Post)(nil)).
		Where("user_id = ?", userID).
		Scan(ctx, &postIDs)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка получения постов пользователя"})
	}

	if len(postIDs) > 0 {
		// Удаляем данные, связанные с постами пользователя, из связанных таблиц
		relatedTables := []struct {
			model interface{}
			query string
		}{
			{&model.PostLike{}, "post_id IN (?)"},
			{&model.Comment{}, "post_id IN (?)"},
			{&model.Repost{}, "original_post_id IN (?)"},
			{&model.PostTag{}, "post_id IN (?)"},
			{&model.Media{}, "post_id IN (?)"},
		}
		for _, table := range relatedTables {
			if _, err = tx.NewDelete().Model(table.model).Where(table.query, bun.In(postIDs)).Exec(ctx); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": fmt.Sprintf("Ошибка удаления из таблицы %T: %v", table.model, err),
				})
			}
		}
		// Удаляем посты пользователя
		if _, err = tx.NewDelete().Model((*model.Post)(nil)).Where("user_id = ?", userID).Exec(ctx); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка удаления постов пользователя"})
		}
	}

	// Шаг 5. Удаляем самого пользователя
	if _, err = tx.NewDelete().Model((*model.User)(nil)).Where("id = ?", userID).Exec(ctx); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка удаления пользователя"})
	}

	// Фиксируем транзакцию
	if err = tx.Commit(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка фиксации транзакции"})
	}

	return c.NoContent(http.StatusNoContent)
}