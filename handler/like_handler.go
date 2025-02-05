package handler

import (
	"api-service/model"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// Добавление лайка к посту
func (h *PostHandler) LikePost(c echo.Context) error {
	// Получаем ID поста из параметров маршрута
	postIDParam := c.Param("id")
	postID, err := strconv.Atoi(postIDParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный ID поста"})
	}
  
	// Получаем ID пользователя из контекста
	userID := c.Get("user_id").(int)
  
	ctx := c.Request().Context()
  
	// Проверяем, существует ли лайк
	existingLike := &model.PostLike{}
	err = h.DB.NewSelect().
		Model(existingLike).
		Where("post_id = ? AND user_id = ?", postID, userID).
		Scan(ctx)
  
	if err == nil {
		// Лайк существует — удаляем
		_, err = h.DB.NewDelete().
			Model(existingLike).
			Where("post_id = ? AND user_id = ?", postID, userID).
			Exec(ctx)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при удалении лайка"})
		}
  
		// Уменьшаем счетчик лайков
		_, err = h.DB.NewUpdate().
			Model(&model.Post{}).
			Set("likes_count = likes_count - 1").
			Where("id = ?", postID).
			Exec(ctx)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении счетчика лайков"})
		}
  
		return c.JSON(http.StatusOK, map[string]string{"message": "Лайк удалён"})
	}
  
	// Если лайк не существует — добавляем
	like := &model.PostLike{
		PostID: postID,
		UserID: userID,
	}
	_, err = h.DB.NewInsert().Model(like).Exec(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при добавлении лайка"})
	}
  
	// Увеличиваем счетчик лайков
	_, err = h.DB.NewUpdate().
		Model(&model.Post{}).
		Set("likes_count = likes_count + 1").
		Where("id = ?", postID).
		Exec(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении счетчика лайков"})
	}
  
	return c.JSON(http.StatusOK, map[string]string{"message": "Лайк добавлен"})
}