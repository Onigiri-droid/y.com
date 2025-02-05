package handler

import (
	"api-service/model"
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// Добавление комментария к посту
func (h *PostHandler) CommentOnPost(c echo.Context) error {
	// Получаем PostID из параметра маршрута
	postIDParam := c.Param("id")
	postID, err := strconv.Atoi(postIDParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный ID поста"})
	}
  
	// Получаем UserID из контекста
	userID := c.Get("user_id").(int)
  
	// Привязываем тело запроса
	req := new(struct {
		Content string `json:"content"`
	})
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный запрос"})
	}
  
	// Создаём комментарий
	comment := &model.Comment{
		PostID:  postID,
		UserID:  userID,
		Content: req.Content,
	}
  
	// Сохраняем комментарий
	_, err = h.DB.NewInsert().Model(comment).Exec(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при добавлении комментария"})
	}
  
	  // Обновляем счетчик комментариев
	  _, err = h.DB.NewUpdate().
		  Model(&model.Post{}).
		  Set("comments_count = comments_count + 1").
		  Where("id = ?", postID).
		  Exec(c.Request().Context())
	  if err != nil {
		  return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении счетчика комментариев"})
	  }
	  
	return c.NoContent(http.StatusCreated)
}

// Обновление комментария
func (h *PostHandler) UpdateComment(c echo.Context) error {
	commentID, err := strconv.Atoi(c.Param("comment_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid comment ID"})
	}
  
	userID := c.Get("user_id").(int) // Получение `user_id` авторизованного пользователя из контекста
  
	ctx := c.Request().Context()
  
	// Проверяем, что комментарий существует
	comment := &model.Comment{}
	err = h.DB.NewSelect().
		Model(comment).
		Where("id = ?", commentID).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Comment not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve comment"})
	}
  
	// Проверяем, что пользователь — автор комментария
	if comment.UserID != userID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "You are not allowed to update this comment"})
	}
  
	// Привязываем тело запроса
	req := new(struct {
		Content string `json:"content"`
	})
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
  
	// Обновляем комментарий
	_, err = h.DB.NewUpdate().
		Model(comment).
		Set("content = ?", req.Content).
		Where("id = ?", commentID).
		Exec(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update comment"})
	}
  
	return c.JSON(http.StatusOK, map[string]string{"message": "Comment updated successfully"})
}

// Получение комментариев по ID поста
func (h *PostHandler) GetCommentsByPostID(c echo.Context) error {
	postIDParam := c.Param("id")
	postID, err := strconv.Atoi(postIDParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный ID поста"})
	}

	comments := make([]model.Comment, 0)
	err = h.DB.NewSelect().
		Model(&comments).
		Where("post_id = ?", postID).
		Scan(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении комментариев"})
	}

	return c.JSON(http.StatusOK, comments)
}

// Удаление своих комментов
func (h *PostHandler) DeleteComment(c echo.Context) error {
	postID, err := strconv.Atoi(c.Param("post_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid post ID"})
	}
  
	commentID, err := strconv.Atoi(c.Param("comment_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid comment ID"})
	}
  
	userID := c.Get("user_id").(int) // Получение `user_id` авторизованного пользователя из контекста
  
	ctx := c.Request().Context()
  
	// Проверяем, что комментарий существует и принадлежит указанному посту
	comment := &model.Comment{}
	err = h.DB.NewSelect().
		Model(comment).
		Where("id = ?", commentID).
		Where("post_id = ?", postID).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Comment not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve comment"})
	}
  
	// Проверяем, что пользователь — автор комментария
	if comment.UserID != userID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "You are not allowed to delete this comment"})
	}
  
	// Удаляем комментарий
	_, err = h.DB.NewDelete().Model(comment).Where("id = ?", commentID).Exec(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete comment"})
	}
  
	// Обновляем счётчик комментариев
	_, err = h.DB.NewUpdate().
		Model((*model.Post)(nil)).
		Set("comments_count = comments_count - 1").
		Where("id = ?", postID).
		Exec(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update comments count"})
	}
  
	return c.JSON(http.StatusOK, map[string]string{"message": "Comment deleted successfully"})
}