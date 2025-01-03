package router

import (
    "api-service/handler"
    "api-service/middleware"
    "github.com/labstack/echo/v4"
)

func SetupRoutes(e *echo.Echo, userHandler *handler.UserHandler, postHandler *handler.PostHandler) {
    // Публичные маршруты
    e.POST("/register", userHandler.Register)
    e.POST("/login", userHandler.Login)

    // Защищенные маршруты для пользователей
    e.GET("/users", userHandler.GetUsers, middleware.JWTMiddleware) // Получение списка всех пользователей (с JWT)
    e.POST("/users", userHandler.CreateUser, middleware.JWTMiddleware) // Создание нового пользователя (без JWT)
    e.PUT("/users/:id", userHandler.UpdateUser, middleware.JWTMiddleware) // Обновление пользователя (с JWT)
    e.DELETE("/users/:id", userHandler.DeleteUser, middleware.JWTMiddleware) // Удаление пользователя (с JWT)

    // Маршруты для постов
    e.POST("/posts", postHandler.CreatePost, middleware.JWTMiddleware) // Создание поста (с JWT)
    e.GET("/posts", postHandler.GetPosts) // Получение списка всех постов (публично)
    e.GET("/posts/:id", postHandler.GetPostByID) // Получение конкретного поста по ID (публично)
    e.PUT("/posts/:id", postHandler.UpdatePost, middleware.JWTMiddleware) // Обновление поста (с JWT)
    e.DELETE("/posts/:id", postHandler.DeletePost, middleware.JWTMiddleware) // Удаление поста (с JWT)
}


