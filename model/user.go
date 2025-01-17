package model

// Структура пользователя
type User struct {
	ID       int    `json:"id" bun:",pk,autoincrement"`
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

// Структура для запроса на создание пользователя
type CreateUserRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=10"`
}

// Структура для запроса на обновление пользователя
type UpdateUserRequest struct {
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty" validate:"omitempty,email"`
	Password string `json:"password,omitempty" validate:"omitempty,min=10"`
}

// Структура для запроса на логин
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
