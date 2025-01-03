package handler

import (
	"api-service/model"
	"context"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
)

type PostHandler struct {
    DB *bun.DB
}

// Создание нового поста
func (h *PostHandler) CreatePost(c echo.Context) error {
    req := new(model.CreatePostRequest)
    if err := c.Bind(req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный запрос"})
    }

    userID, ok := c.Get("user_id").(int)
    if !ok {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Ошибка аутентификации"})
    }

    post := &model.Post{
        Title:   req.Title,
        Content: req.Content,
        UserID:  userID, // Используем user_id из контекста
    }

    _, err := h.DB.NewInsert().Model(post).Exec(context.Background())
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при создании поста"})
    }

    return c.JSON(http.StatusCreated, post)
}


// Получение всех постов
func (h *PostHandler) GetPosts(c echo.Context) error {
    var posts []model.Post
    err := h.DB.NewSelect().Model(&posts).Order("created_at DESC").Scan(context.Background())
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении постов"})
    }

    return c.JSON(http.StatusOK, posts)
}

// Получение поста по ID
func (h *PostHandler) GetPostByID(c echo.Context) error {
    id := c.Param("id")
    post := new(model.Post)

    err := h.DB.NewSelect().Model(post).Where("id = ?", id).Scan(context.Background())
    if err != nil {
        return c.JSON(http.StatusNotFound, map[string]string{"error": "Пост не найден"})
    }

    return c.JSON(http.StatusOK, post)
}

// Обновление поста
func (h *PostHandler) UpdatePost(c echo.Context) error {
    idParam := c.Param("id")
    id, err := strconv.Atoi(idParam)
    if err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный ID"})
    }

    req := new(model.UpdatePostRequest)
    if err := c.Bind(req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный запрос"})
    }

    _, err = h.DB.NewUpdate().
        Model(&model.Post{ID: id}).
        Set("title = ?", req.Title).
        Set("content = ?", req.Content).
        Where("id = ?", id).
        Exec(context.Background())

    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении поста"})
    }

    return c.NoContent(http.StatusOK)
}

// Удаление поста
func (h *PostHandler) DeletePost(c echo.Context) error {
    id := c.Param("id")

    _, err := h.DB.NewDelete().Model((*model.Post)(nil)).Where("id = ?", id).Exec(context.Background())
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при удалении поста"})
    }

    return c.NoContent(http.StatusNoContent)
}
