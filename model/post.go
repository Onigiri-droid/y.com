package model

import "time"

// Структура поста
type Post struct {
    ID        int       `json:"id" bun:",pk,autoincrement"`
    Title     string    `json:"title" bun:",notnull"`
    Content   string    `json:"content"`
    UserID    int       `json:"user_id" bun:",notnull"`
    CreatedAt time.Time `json:"created_at" bun:",nullzero,notnull,default:current_timestamp"`
}

// Структура для запроса на создание поста
type CreatePostRequest struct {
    Title   string `json:"title" validate:"required"`
    Content string `json:"content" validate:"required"`
}

// Структура для обновления поста
type UpdatePostRequest struct {
    Title   string `json:"title"`
    Content string `json:"content"`
}
