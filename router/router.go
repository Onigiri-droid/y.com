package router

import (
	"api-service/handler"
	"api-service/middleware"
	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
)

func SetupRoutes(e *echo.Echo, userHandler *handler.UserHandler, postHandler *handler.PostHandler, db *bun.DB) {
	// Публичные маршруты
	e.POST("/register", userHandler.Register)
	e.POST("/login", userHandler.Login)

	// Группа защищенных маршрутов
	authGroup := e.Group("")
	authGroup.Use(middleware.JWTMiddleware(db)) // Добавляем middleware для всех защищенных маршрутов

	// Защищенные маршруты для пользователей
	authGroup.GET("/users", userHandler.GetUsers)          // Получить список пользователей
	authGroup.PUT("/users", userHandler.UpdateUser)        // Обновить текущего пользователя
	authGroup.DELETE("/users", userHandler.DeleteUser)     // Удалить текущего пользователя

	// Публичные маршруты для постов
	e.GET("/posts", postHandler.GetPosts)
	e.GET("/posts/:id", postHandler.GetPostByID)
	e.GET("/posts/:id/comments", postHandler.GetCommentsByPostID)
	e.GET("/tags", postHandler.GetAllTags)
	e.GET("/users/:user_id/posts", postHandler.GetUserPosts)

	// Защищенные маршруты для постов
	authGroup.POST("/posts", postHandler.CreatePost)
	authGroup.PUT("/posts/:id", postHandler.UpdatePost)
	authGroup.DELETE("/posts/:id", postHandler.DeletePost)
	authGroup.POST("/posts/:id/like", postHandler.LikePost)
	authGroup.POST("/posts/:id/comment", postHandler.CommentOnPost)
	authGroup.PUT("/comments/:comment_id", postHandler.UpdateComment)
	authGroup.DELETE("/posts/:post_id/comment/:comment_id", postHandler.DeleteComment)
	authGroup.POST("/posts/:id/repost", postHandler.RepostPost)
	authGroup.DELETE("/posts/:id/repost", postHandler.DeleteRepost)
}

// package router

// import (
// 	"api-service/handler"
// 	"api-service/middleware"

// 	"github.com/labstack/echo/v4"
// )

// func SetupRoutes(e *echo.Echo, userHandler *handler.UserHandler, postHandler *handler.PostHandler) {
// 	SetupUserRoutes(e, userHandler)
// 	SetupPostRoutes(e, postHandler)
// }

// func SetupUserRoutes(e *echo.Echo, userHandler *handler.UserHandler) {
// 	// Публичные маршруты
// 	e.POST("/register", userHandler.Register)
// 	e.POST("/login", userHandler.Login)

// 	// Защищенные маршруты для пользователей
// 	e.GET("/users", userHandler.GetUsers, middleware.JWTMiddleware)
// 	e.POST("/users", userHandler.CreateUser, middleware.JWTMiddleware)
// 	e.PUT("/users/:id", userHandler.UpdateUser, middleware.JWTMiddleware)
// 	e.DELETE("/users/:id", userHandler.DeleteUser, middleware.JWTMiddleware)
// }

// func SetupPostRoutes(e *echo.Echo, postHandler *handler.PostHandler) {
// 	// Публичные маршруты
// 	e.GET("/posts", postHandler.GetPosts)
// 	e.GET("/posts/:id", postHandler.GetPostByID)
// 	e.GET("/posts/:id/comments", postHandler.GetCommentsByPostID) // Новый маршрут для комментариев к посту
// 	e.GET("/tags", postHandler.GetAllTags)                        // Новый маршрут для тегов
// 	e.GET("/users/:user_id/posts", postHandler.GetUserPosts)

// 	// Защищенные маршруты для постов
// 	e.POST("/posts", postHandler.CreatePost, middleware.JWTMiddleware)
// 	e.PUT("/posts/:id", postHandler.UpdatePost, middleware.JWTMiddleware)
// 	e.DELETE("/posts/:id", postHandler.DeletePost, middleware.JWTMiddleware)
// 	e.POST("/posts/:id/like", postHandler.LikePost, middleware.JWTMiddleware)
// 	e.POST("/posts/:id/comment", postHandler.CommentOnPost, middleware.JWTMiddleware)
// 	e.PUT("/comments/:comment_id", postHandler.UpdateComment, middleware.JWTMiddleware)
// 	e.DELETE("/posts/:post_id/comment/:comment_id", postHandler.DeleteComment, middleware.JWTMiddleware)
// 	e.POST("/posts/:id/repost", postHandler.RepostPost, middleware.JWTMiddleware)
// 	e.DELETE("/posts/:id/repost", postHandler.DeleteRepost, middleware.JWTMiddleware)
// }
