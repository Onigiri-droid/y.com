package model

import "time"

// Структура поста
type Post struct {
	ID        int          `json:"id" bun:",pk,autoincrement"`
	Title     string       `json:"title" bun:",notnull"`
	Content   string       `json:"content"`
	UserID    int          `json:"user_id" bun:",notnull"`
	User      *User        `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`
	CreatedAt time.Time    `json:"created_at" bun:",nullzero,notnull,default:current_timestamp"`
	Tags      []PostTag    `json:"tags,omitempty" bun:"rel:has-many,join:id=post_id"`
	Media     []Media  `json:"media,omitempty" bun:"rel:has-many,join:id=post_id"`
	Comments []Comment `json:"comments,omitempty" bun:"rel:has-many,join:id=post_id"`
	LikesCount    int `json:"likes_count"`
	RepostsCount  int `json:"reposts_count"`
	CommentsCount int `json:"comments_count"`
}

// Структура медиафайлов
type Media struct {
	ID     int    `json:"id,omitempty" bun:",pk,autoincrement"` // Идентификатор медиа (может быть пустым при загрузке)
	PostID int    `json:"post_id,omitempty"`                   // ID поста, к которому прикреплено медиа
	URL    string `json:"url" validate:"required"`             // Ссылка на медиафайл
	Type   string `json:"type" validate:"required"`            // Тип медиафайла: "image" или "video"
}

// Структура для тегов
type Tag struct {
	ID   int    `json:"id" bun:",pk,autoincrement"`
	Name string `json:"name" bun:",unique,notnull"`
}

// Связь постов и тегов
type PostTag struct {
    PostID int `json:"post_id" bun:",pk,notnull"`
    TagID  int `json:"tag_id" bun:",pk,notnull,column:tag"`
}


// Структура для лайков
type PostLike struct {
	ID     int    `json:"id" bun:",pk,autoincrement"`
	PostID int    `json:"post_id" bun:",notnull"`
	UserID int    `json:"user_id" bun:",notnull"`
	Post   *Post  `json:"post,omitempty" bun:"rel:belongs-to,join:post_id=id"`
	User   *User  `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`
}

// Структура для комментариев
type Comment struct {
	ID        int       `json:"id" bun:",pk,autoincrement"`
	PostID    int       `json:"post_id" bun:",notnull"`
	UserID    int       `json:"user_id" bun:",notnull"`
	Content   string    `json:"content" bun:",notnull"`
	CreatedAt time.Time `json:"created_at" bun:",nullzero,notnull,default:current_timestamp"`
	Post      *Post     `json:"post,omitempty" bun:"rel:belongs-to,join:post_id=id"`
	User      *User     `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`
}

// Структура для репостов
type Repost struct {
    ID             int       `json:"id" bun:",pk,autoincrement"`
    OriginalPostID int       `json:"original_post_id" bun:",notnull"`
    UserID         int       `json:"user_id" bun:",notnull"`
    OriginalPost   *Post     `json:"original_post,omitempty" bun:"rel:belongs-to,join:original_post_id=id"`
    User           *User     `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`
    CreatedAt      time.Time `json:"created_at" bun:",nullzero,notnull,default:current_timestamp"`
}

// Структура для запроса на создание поста
type CreatePostRequest struct {
	UserID  int    	 `json:"user_id" bun:",notnull"`
	User    *User    `json:"user,omitempty" bun:"rel:belongs-to,join:user_id=id"`
	Title   string   `json:"title" validate:"required"`
	Content string   `json:"content" validate:"required"`
	Tags    []string `json:"tags,omitempty"`
	Media   []Media  `json:"media,omitempty"` // Используем объединенную структуру
}

// Структура для запроса на обновление поста
type UpdatePostRequest struct {
	UserID  int    `json:"user_id" bun:",notnull"`
	Title   string   `json:"title,omitempty"`
	Content string   `json:"content,omitempty"`
	Tags    []string `json:"tags,omitempty"`
	Media   []Media  `json:"media,omitempty"` // Используем объединенную структуру
}

