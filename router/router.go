package router

import (
	"api-service/handler"
	"api-service/middleware"

	"github.com/labstack/echo/v4"
)

func SetupRoutes(e *echo.Echo, userHandler *handler.UserHandler, postHandler *handler.PostHandler) {
	SetupUserRoutes(e, userHandler)
	SetupPostRoutes(e, postHandler)
}

func SetupUserRoutes(e *echo.Echo, userHandler *handler.UserHandler) {
	// Публичные маршруты
	e.POST("/register", userHandler.Register)
	e.POST("/login", userHandler.Login)

	// Защищенные маршруты для пользователей
	e.GET("/users", userHandler.GetUsers, middleware.JWTMiddleware)
	e.POST("/users", userHandler.CreateUser, middleware.JWTMiddleware)
	e.PUT("/users/:id", userHandler.UpdateUser, middleware.JWTMiddleware)
	e.DELETE("/users/:id", userHandler.DeleteUser, middleware.JWTMiddleware)
}

func SetupPostRoutes(e *echo.Echo, postHandler *handler.PostHandler) {
	// Публичные маршруты
	e.GET("/posts", postHandler.GetPosts)
	e.GET("/posts/:id", postHandler.GetPostByID)
	e.GET("/posts/:id/comments", postHandler.GetCommentsByPostID) // Новый маршрут для комментариев к посту
	e.GET("/tags", postHandler.GetAllTags)                        // Новый маршрут для тегов
	e.GET("/users/:user_id/posts", postHandler.GetUserPosts)

	// Защищенные маршруты для постов
	e.POST("/posts", postHandler.CreatePost, middleware.JWTMiddleware)
	e.PUT("/posts/:id", postHandler.UpdatePost, middleware.JWTMiddleware)
	e.DELETE("/posts/:id", postHandler.DeletePost, middleware.JWTMiddleware)
	e.POST("/posts/:id/like", postHandler.LikePost, middleware.JWTMiddleware)
	e.POST("/posts/:id/comment", postHandler.CommentOnPost, middleware.JWTMiddleware)
	// e.DELETE("/posts/:id/comment/:id", postHandler.DeleteComment, middleware.JWTMiddleware)
	e.DELETE("/posts/:post_id/comment/:comment_id", postHandler.DeleteComment, middleware.JWTMiddleware)
	e.POST("/posts/:id/repost", postHandler.RepostPost, middleware.JWTMiddleware)
	e.DELETE("/posts/:id/repost", postHandler.DeleteRepost, middleware.JWTMiddleware)
}
