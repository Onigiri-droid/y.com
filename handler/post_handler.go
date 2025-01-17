package handler

import (
	"api-service/model"
	"database/sql"
	// "encoding/json"
	"errors"
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
	ctx := c.Request().Context()

	req := new(model.CreatePostRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный запрос"})
	}

  // Извлекаем user_id из токена
  userID := c.Get("user_id").(int) // Убедись, что `user_id` добавляется в middleware

  post := &model.Post{
      Title:   req.Title,
      Content: req.Content,
      UserID:  userID,
  }

 // Сохраняем пост
 _, err := h.DB.NewInsert().Model(post).Exec(ctx)
 if err != nil {
   return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при создании поста"})
 }

 // Добавляем теги
 for _, tagName := range req.Tags {
   tag := &model.Tag{}
   err := h.DB.NewSelect().Model(tag).Where("name = ?", tagName).Scan(ctx)

   if err != nil {
     if errors.Is(err, sql.ErrNoRows) {
       // Тег не найден — создаём
       tag = &model.Tag{Name: tagName}
       _, err := h.DB.NewInsert().Model(tag).On("CONFLICT (name) DO NOTHING").Exec(ctx)
       if err != nil {
         return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при добавлении тега"})
       }
       // Заново выбираем ID созданного или существующего тега
       _ = h.DB.NewSelect().Model(tag).Where("name = ?", tagName).Scan(ctx)
     } else {
       return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при проверке тега"})
     }
   }

   postTag := &model.PostTag{
     PostID: post.ID,
     TagID:  tag.ID,
   }
   _, err = h.DB.NewInsert().Model(postTag).Exec(ctx)
   if err != nil {
     return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при связывании поста с тегом"})
   }
 }

 // Добавляем медиафайлы
 for _, media := range req.Media {
   mediaFile := &model.Media{PostID: post.ID, URL: media.URL, Type: media.Type}
   _, err := h.DB.NewInsert().Model(mediaFile).Exec(ctx)
   if err != nil {
     return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при добавлении медиафайлов"})
   }
 }

 return c.JSON(http.StatusCreated, post)
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

// Получение всех тегов
func (h *PostHandler) GetAllTags(c echo.Context) error {
	tags := make([]model.Tag, 0)
	err := h.DB.NewSelect().Model(&tags).Scan(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении тегов"})
	}

	return c.JSON(http.StatusOK, tags)
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

// Обновление поста
func (h *PostHandler) UpdatePost(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный ID"})
	}

	userID := c.Get("user_id").(int)

	// Проверяем, принадлежит ли пост текущему пользователю
	post := &model.Post{}
	err = h.DB.NewSelect().Model(post).Where("id = ?", id).Scan(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Пост не найден"})
	}

	if post.UserID != userID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Вы не можете редактировать этот пост"})
	}

	req := new(model.UpdatePostRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный запрос"})
	}

	// Обновление основного контента поста
	_, err = h.DB.NewUpdate().
		Model(&model.Post{ID: id}).
		Set("title = ?", req.Title).
		Set("content = ?", req.Content).
		Where("id = ?", id).
		Exec(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении поста"})
	}

  // Обновление тегов
  if len(req.Tags) > 0 {
    // Удаляем старые связи поста с тегами
    _, err := h.DB.NewDelete().
        Model((*model.PostTag)(nil)).
        Where("post_id = ?", id).
        Exec(c.Request().Context())
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении тегов"})
    }

    for _, tagName := range req.Tags {
      // Проверяем наличие тега в базе данных
      tag := &model.Tag{}
      err := h.DB.NewSelect().Model(tag).Where("name = ?", tagName).Scan(c.Request().Context())
      if err != nil {
          if errors.Is(err, sql.ErrNoRows) {
              // Тег не найден — создаём
              tag = &model.Tag{Name: tagName}
              _, err = h.DB.NewInsert().Model(tag).Exec(c.Request().Context())
              if err != nil {
                  return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при создании тега"})
              }
          } else {
              return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при поиске тега"})
          }
      }

      // Создаём связь поста с тегом
      postTag := &model.PostTag{
          PostID: id,
          TagID:  tag.ID,
      }
      _, err = h.DB.NewInsert().Model(postTag).Exec(c.Request().Context())
      if err != nil {
          return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при добавлении тега к посту"})
      }
  }
}

  // Обновление медиафайлов
  if len(req.Media) > 0 {
    _, err := h.DB.NewDelete().
      Model((*model.Media)(nil)).
      Where("post_id = ?", id).
      Exec(c.Request().Context())
    if err != nil {
      return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении медиафайлов"})
    }

    for _, media := range req.Media {
      mediaFile := &model.Media{PostID: id, URL: media.URL, Type: media.Type}
      _, err := h.DB.NewInsert().
        Model(mediaFile).
        Exec(c.Request().Context())
      if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при добавлении медиафайлов"})
      }
    }
  }

  return c.NoContent(http.StatusOK)
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
