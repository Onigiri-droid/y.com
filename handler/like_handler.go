package handler

import (
	"api-service/model"
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// Добавление лайка к посту
func (h *PostHandler) LikePost(c echo.Context) error {
	// Получаем ID поста
	postIDParam := c.Param("id")
	postID, err := strconv.Atoi(postIDParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный ID поста"})
	}

	// Получаем ID пользователя
	userID := c.Get("user_id").(int)
	ctx := c.Request().Context()

	// Начинаем транзакцию
	tx, err := h.DB.BeginTx(ctx, nil)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка начала транзакции"})
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// Проверяем, существует ли лайк
	existingLike := new(model.PostLike)
	err = tx.NewSelect().
		Model(existingLike).
		Where("post_id = ? AND user_id = ?", postID, userID).
		Scan(ctx)

	if err == nil { // Лайк уже есть — удаляем
		_, err = tx.NewDelete().
			Model((*model.PostLike)(nil)).
			Where("post_id = ? AND user_id = ?", postID, userID).
			Exec(ctx)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при удалении лайка"})
		}

		// Уменьшаем счетчик
		_, err = tx.NewUpdate().
			Model((*model.Post)(nil)).
			Set("likes_count = GREATEST(likes_count - 1, 0)").
			Where("id = ?", postID).
			Exec(ctx)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении счетчика лайков"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Лайк удалён"})
	}

	if !errors.Is(err, sql.ErrNoRows) { // Неизвестная ошибка
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при проверке лайка"})
	}

	// Лайка нет — добавляем
	like := &model.PostLike{
		PostID: postID,
		UserID: userID,
	}
	_, err = tx.NewInsert().Model(like).Exec(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при добавлении лайка"})
	}

	// Увеличиваем счетчик
	_, err = tx.NewUpdate().
		Model((*model.Post)(nil)).
		Set("likes_count = likes_count + 1").
		Where("id = ?", postID).
		Exec(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении счетчика лайков"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Лайк добавлен"})
}
			