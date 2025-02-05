package handler

import (
	"api-service/model"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
)

type PostHandler struct {
	DB *bun.DB
}

// Хелпер для обработки ошибок базы данных
func (h *PostHandler) respondWithError(c echo.Context, status int, message string, err error) error {
	if err != nil {
		return c.JSON(status, map[string]string{"error": message, "details": err.Error()})
	}
	return c.JSON(status, map[string]string{"error": message})
}

// Хелпер для работы с тегами
func (h *PostHandler) manageTags(ctx echo.Context, postID int, tags []string) error {
    // Начинаем транзакцию
    tx, err := h.DB.BeginTx(ctx.Request().Context(), nil)
    if err != nil {
        log.Printf("Ошибка при начале транзакции: %v", err)
        return fmt.Errorf("failed to start transaction: %w", err)
    }
    defer tx.Rollback() // Откат транзакции в случае ошибки

    for _, tagName := range tags {
        tag := &model.Tag{}
        err := tx.NewSelect().Model(tag).Where("name = ?", tagName).Scan(ctx.Request().Context())
        if err != nil && !errors.Is(err, sql.ErrNoRows) {
            log.Printf("Ошибка при проверке тега: %v", err)
            return fmt.Errorf("failed to check tag: %w", err)
        }

        // Если тег не найден, создаем его
        if errors.Is(err, sql.ErrNoRows) {
            tag = &model.Tag{Name: tagName}
            _, err := tx.NewInsert().Model(tag).Exec(ctx.Request().Context())
            if err != nil {
                log.Printf("Ошибка при создании тега: %v", err)
                return fmt.Errorf("failed to create tag: %w", err)
            }
        }

        // Проверяем, существует ли уже связь между постом и тегом
        exists, err := tx.NewSelect().
            Model((*model.PostTag)(nil)).
            Where("post_id = ? AND tag_id = ?", postID, tag.ID).
            Limit(1).
            Exists(ctx.Request().Context())
        if err != nil {
            log.Printf("Ошибка при проверке связи пост-тег: %v", err)
            return fmt.Errorf("failed to check post-tag relationship: %w", err)
        }

        // Если связь не существует, создаем её
        if !exists {
            postTag := &model.PostTag{PostID: postID, TagID: tag.ID}
            _, err = tx.NewInsert().Model(postTag).Exec(ctx.Request().Context())
            if err != nil {
                log.Printf("Ошибка при создании связи пост-тег: %v", err)
                return fmt.Errorf("failed to link tag to post: %w", err)
            }
        }
    }

    // Фиксируем транзакцию
    if err := tx.Commit(); err != nil {
        log.Printf("Ошибка при фиксации транзакции: %v", err)
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    return nil
}

// Хелпер для работы с медиа
func (h *PostHandler) manageMedia(ctx echo.Context, postID int, media []model.Media) error {
	for _, mediaItem := range media {
		mediaFile := &model.Media{PostID: postID, URL: mediaItem.URL, Type: mediaItem.Type}
		_, err := h.DB.NewInsert().Model(mediaFile).Exec(ctx.Request().Context())
		if err != nil {
			return errors.New("failed to add media")
		}
	}
	return nil
}

func (h *PostHandler) CreatePost(c echo.Context) error {
    var request struct {
        Title   string   `json:"title"`
        Content string   `json:"content"`
        Tags    []string `json:"tags"`
    }

    if err := c.Bind(&request); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
    }

    // Извлекаем user_id из контекста (например, из JWT токена)
    userID, ok := c.Get("user_id").(int)
    if !ok {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "user not authenticated"})
    }

    // Проверяем, что user_id не равен 0
    if userID == 0 {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user ID"})
    }

    // Проверяем, существует ли пользователь
    exists, err := h.DB.NewSelect().
        Model((*model.User)(nil)).
        Where("id = ?", userID).
        Exists(c.Request().Context())
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to check user existence"})
    }
    if !exists {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "user does not exist"})
    }

    // Создаем пост
    post := &model.Post{
        Title:   request.Title,
        Content: request.Content,
        UserID:  userID,
    }

    if _, err := h.DB.NewInsert().Model(post).Exec(c.Request().Context()); err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create post"})
    }

    // Управляем тегами
    if err := h.manageTags(c, post.ID, request.Tags); err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }

    return c.JSON(http.StatusCreated, post)
}

// Обновление поста
func (h *PostHandler) UpdatePost(c echo.Context) error {
    // Извлекаем ID поста из параметра URL
    idParam := c.Param("id")
    id, err := strconv.Atoi(idParam)
    if err != nil {
        return h.respondWithError(c, http.StatusBadRequest, "Invalid post ID", err)
    }

    // Извлекаем userID из контекста
    userID, ok := c.Get("user_id").(int)
    if !ok {
        log.Printf("Не удалось извлечь userID из контекста")
        return h.respondWithError(c, http.StatusInternalServerError, "Failed to get user ID", nil)
    }

    // Получаем пост из базы данных
    post := &model.Post{}
    err = h.DB.NewSelect().Model(post).Where("id = ?", id).Scan(c.Request().Context())
    if err != nil {
        return h.respondWithError(c, http.StatusNotFound, "Post not found", err)
    }

    // Проверяем, что пользователь имеет право редактировать пост
    if post.UserID != userID {
        return h.respondWithError(c, http.StatusForbidden, "You cannot edit this post", nil)
    }

    // Парсим запрос на обновление
    req := new(model.UpdatePostRequest)
    if err := c.Bind(req); err != nil {
        return h.respondWithError(c, http.StatusBadRequest, "Invalid request", err)
    }

    // Обновляем пост
    _, err = h.DB.NewUpdate().Model(&model.Post{ID: id}).
        Set("title = ?", req.Title).
        Set("content = ?", req.Content).
        Where("id = ?", id).
        Exec(c.Request().Context())
    if err != nil {
        return h.respondWithError(c, http.StatusInternalServerError, "Failed to update post", err)
    }

    // Удаляем старые теги
    _, err = h.DB.NewDelete().
        Model((*model.PostTag)(nil)).
        Where("post_id = ?", id).
        Exec(c.Request().Context())
    if err != nil {
        return h.respondWithError(c, http.StatusInternalServerError, "Failed to delete old tags", err)
    }

    // Управляем тегами
    if err := h.manageTags(c, id, req.Tags); err != nil {
        return h.respondWithError(c, http.StatusInternalServerError, "Failed to update tags", err)
    }

    // Управляем медиа
    if err := h.manageMedia(c, id, req.Media); err != nil {
        return h.respondWithError(c, http.StatusInternalServerError, "Failed to update media", err)
    }

    // Возвращаем успешный ответ
    return c.NoContent(http.StatusOK)
}

// Получение списка постов
func (h *PostHandler) GetPosts(c echo.Context) error {
	posts := make([]model.Post, 0)
	err := h.DB.NewSelect().Model(&posts).
	  Relation("Tags").
	  Relation("Media").
	  Scan(c.Request().Context())
	if err != nil {
	  return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении постов"})
	}
  
	return c.JSON(http.StatusOK, posts)
}

// Получение поста по ID (включая комментарии, теги, медиафайлы и репосты)
func (h *PostHandler) GetPostByID(c echo.Context) error {
	id := c.Param("id")
	post := new(model.Post)
  
	err := h.DB.NewSelect().
	  Model(post).
	  Relation("Comments").
	  Relation("Tags").
	  Relation("Media").
	  Where("post.id = ?", id).
	  Scan(c.Request().Context())
	if err != nil {
	  return c.JSON(http.StatusNotFound, map[string]string{"error": "Пост не найден"})
	}
  
	return c.JSON(http.StatusOK, post)
}

// GetUserPosts возвращает посты и репосты пользователя
func (h *PostHandler) GetUserPosts(c echo.Context) error {
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	var posts []model.Post
	err = h.DB.NewSelect().
		Model(&posts).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Scan(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve posts"})
	}

	var reposts []model.Repost
	err = h.DB.NewSelect().
		Model(&reposts).
		Relation("OriginalPost").
		Where(`"repost"."user_id" = ?`, userID).
		Order("repost.created_at DESC").
		Scan(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve reposts"})
	}

	response := struct {
		Posts   []model.Post   `json:"posts"`
		Reposts []model.Repost `json:"reposts"`
	}{
		Posts:   posts,
		Reposts: reposts,
	}

	return c.JSON(http.StatusOK, response)
}

// Удаление поста
func (h *PostHandler) DeletePost(c echo.Context) error {
	userID := c.Get("user_id").(int) // ID текущего пользователя
	postID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid post ID"})
	}
  
	// Проверяем, что пост принадлежит текущему пользователю
	post := &model.Post{}
	err = h.DB.NewSelect().
		Model(post).
		Where("id = ?", postID).
		Where("user_id = ?", userID).
		Scan(c.Request().Context())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Post not found or access denied"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve post"})
	}
  
	ctx := c.Request().Context()
  
	// Удаляем комментарии, связанные с постом
	_, err = h.DB.NewDelete().
		Model((*model.Comment)(nil)).
		Where("post_id = ?", postID).
		Exec(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete comments"})
	}
  
	// Удаляем репосты, связанные с постом
	_, err = h.DB.NewDelete().
		Model((*model.Repost)(nil)).
		Where("original_post_id = ?", postID).
		Exec(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete reposts"})
	}
  
	// Удаляем лайки, связанные с постом
	_, err = h.DB.NewDelete().
		Model((*model.PostLike)(nil)).
		Where("post_id = ?", postID).
		Exec(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete likes"})
	}
  
	// Удаляем сам пост
	_, err = h.DB.NewDelete().
		Model(post).
		Where("id = ?", postID).
		Exec(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete post"})
	}
  
	return c.JSON(http.StatusOK, map[string]string{"message": "Post and related data deleted successfully"})
}
