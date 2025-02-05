package model

import "time"

// Структура пользователя
type User struct {
    ID        int32     `bun:"id,pk,autoincrement"`
    Name      string    `bun:"name,notnull"`
    Email     string    `bun:"email,unique,notnull"`
    Password  string    `bun:"password,notnull"`
    Avatar    string    `bun:"avatar"`
    Status    string    `bun:"status"`
    CreatedAt time.Time `bun:"created_at,default:current_timestamp"`
    LastSeen  time.Time `bun:"last_seen"`
    Role      string    `bun:"role"`
    Bio       string    `bun:"bio"`
}
// type User struct {
// 	ID        int       `json:"id" bson:"_id,omitempty" bun:",pk,autoincrement"`
// 	Name      string    `json:"name,omitempty" bson:"name,omitempty"`
// 	Email     string    `json:"email,omitempty" bson:"email,omitempty"`
// 	Password  string    `json:"password,omitempty" bson:"password,omitempty"`
// 	Avatar    string    `json:"avatar,omitempty" bson:"avatar,omitempty"`
// 	Status    string    `json:"status,omitempty" bson:"status,omitempty"`
// 	CreatedAt time.Time `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
// 	LastSeen  time.Time `json:"lastSeen,omitempty" bson:"lastSeen,omitempty"`
// 	Role      string    `json:"role,omitempty" bson:"role,omitempty"`
// 	Bio       string    `json:"bio,omitempty" bson:"bio,omitempty"`
// }

// Структура для запроса на создание пользователя
type CreateUserRequest struct {
	Name     string `json:"name" validate:"required"`              // Имя пользователя
	Email    string `json:"email" validate:"required,email"`       // Email пользователя
	Password string `json:"password" validate:"required,min=10"`   // Пароль
	Avatar   string `json:"avatar,omitempty" validate:"omitempty"` // Ссылка на аватар (необязательно)
	Status   string `json:"status,omitempty"`                      // Устанавливается по умолчанию, например, "offline"
}

// Структура для запроса на обновление пользователя
type UpdateUserRequest struct {
	Name     string `json:"name,omitempty"`                         // Имя пользователя
	Email    string `json:"email,omitempty" validate:"omitempty,email"` // Email пользователя
	Password string `json:"password,omitempty" validate:"omitempty,min=10"` // Пароль
	Avatar   string `json:"avatar,omitempty"`                       // Ссылка на аватар
	Status   string `json:"status,omitempty"`                       // Изменение статуса
}

// Структура для запроса на логин
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
