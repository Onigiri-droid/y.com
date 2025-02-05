package handler

import (
	"api-service/model"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// Репост поста
func (h *PostHandler) RepostPost(c echo.Context) error {
	userID := c.Get("user_id").(int) // Получаем ID текущего пользователя
	postID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid post ID"})
	}
  
	// Проверяем, что пользователь не пытается репостить свой пост
	originalPost := &model.Post{}
	if err := h.DB.NewSelect().Model(originalPost).Where("id = ?", postID).Scan(c.Request().Context()); err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Post not found"})
	}
	if originalPost.UserID == userID {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "You cannot repost your own post"})
	}
  
	// Проверяем, что пользователь еще не репостил этот пост
	existingRepost := &model.Repost{}
	if err := h.DB.NewSelect().
		Model(existingRepost).
		Where("original_post_id = ? AND user_id = ?", postID, userID).
		Scan(c.Request().Context()); err == nil {
		return c.JSON(http.StatusConflict, map[string]string{"error": "You have already reposted this post"})
	}
  
	// Создаем репост
	repost := &model.Repost{
		OriginalPostID: postID,
		UserID:         userID,
	}
	if _, err := h.DB.NewInsert().Model(repost).Exec(c.Request().Context()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create repost"})
	}
  
	// Увеличиваем счетчик репостов
	if _, err := h.DB.NewUpdate().
		Model((*model.Post)(nil)).
		Set("reposts_count = reposts_count + 1").
		Where("id = ?", postID).
		Exec(c.Request().Context()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update repost count"})
	}
  
	return c.JSON(http.StatusCreated, map[string]string{"message": "Repost created successfully"})
}

// удаление репоста
func (h *PostHandler) DeleteRepost(c echo.Context) error {
	userID := c.Get("user_id").(int) // Получаем ID текущего пользователя
	repostID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid repost ID"})
	}
  
	// Проверяем, что репост принадлежит текущему пользователю
	repost := &model.Repost{}
	if err := h.DB.NewSelect().Model(repost).Where("id = ? AND user_id = ?", repostID, userID).Scan(c.Request().Context()); err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Repost not found"})
	}
  
	// Удаляем репост
	if _, err := h.DB.NewDelete().Model(repost).Where("id = ?", repostID).Exec(c.Request().Context()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete repost"})
	}
  
	// Уменьшаем счетчик репостов для оригинального поста
	if _, err := h.DB.NewUpdate().
		Model((*model.Post)(nil)).
		Set("reposts_count = GREATEST(reposts_count - 1, 0)"). // Предотвращаем отрицательные значения
		Where("id = ?", repost.OriginalPostID).
		Exec(c.Request().Context()); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update repost count"})
	}
  
	return c.JSON(http.StatusOK, map[string]string{"message": "Repost deleted successfully"})
}