package main

import (
	"api-service/db"
	"api-service/handler"
	"api-service/router"
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	authpb "api-service/proto/auth-service/proto" // Импорт для AuthService
	chatpb "api-service/proto/chat-service/proto" // Импорт для ChatService

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const jwtSecret = "supersecretkey"

// Реализация AuthService
type AuthService struct {
	DB *bun.DB
	authpb.UnimplementedAuthServiceServer // Встраиваем UnimplementedAuthServiceServer
}

// Claims представляет структуру данных, хранящихся в токене
type Claims struct {
	UserID   int32  `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// ParseToken парсит JWT-токен и возвращает claims
func ParseToken(tokenString string) (*Claims, error) {
	secretKey := []byte(jwtSecret)

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		log.Printf("Token parsing error: %v", err)
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("невалидный токен")
}

// CheckUserExists проверяет, существует ли пользователь в базе данных
func CheckUserExists(db *bun.DB, userID int32) (bool, error) {
	var exists bool

	// Используем EXISTS для проверки наличия пользователя
	err := db.NewSelect().
		ColumnExpr("EXISTS (SELECT 1 FROM users WHERE id = ?)", userID).
		Scan(context.Background(), &exists)

	if err != nil {
		return false, err
	}
	return exists, nil
}

// Реализация метода ValidateToken
func (s *AuthService) ValidateToken(ctx context.Context, req *authpb.ValidateTokenRequest) (*authpb.ValidateTokenResponse, error) {
	// Парсим токен
	claims, err := ParseToken(req.Token)
	if err != nil {
		return &authpb.ValidateTokenResponse{Valid: false}, nil
	}

	// Проверяем, существует ли пользователь в БД
	userExists, err := CheckUserExists(s.DB, claims.UserID)
	if err != nil || !userExists {
		return &authpb.ValidateTokenResponse{Valid: false}, nil
	}

	return &authpb.ValidateTokenResponse{
		Valid:    true,
		UserId:   claims.UserID,
		Username: claims.Username,
	}, nil
}

func main() {
	e := echo.New()

	// Подключение к базе данных
	dbConn := db.InitDB()
	defer dbConn.Close()

	// Инициализация bun.DB
	bunDB := bun.NewDB(dbConn.DB, pgdialect.New())

	// Миграции
	db.RunMigrations(dbConn)

	// Создаём gRPC-сервер
	grpcServer := grpc.NewServer()
	authService := &AuthService{DB: bunDB} // Передаем bunDB
	authpb.RegisterAuthServiceServer(grpcServer, authService)

	// Запускаем gRPC-сервер
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Ошибка запуска gRPC-сервера: %v", err)
	}

	go func() {
		log.Println("gRPC-сервер запущен на порту 50051")
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Ошибка работы gRPC-сервера: %v", err)
		}
	}()

	// Подключение к Chat-service
	chatConn, err := grpc.Dial(
		"chat-service:50052",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(10*1024*1024), // 10 MB
			grpc.MaxCallSendMsgSize(10*1024*1024), // 10 MB
		),
	)
	if err != nil {
		log.Fatalf("Не удалось подключиться к Chat-service: %v", err)
	}
	defer chatConn.Close()

	chatServiceClient := chatpb.NewChatServiceClient(chatConn)

	// Создаём обработчики
	userHandler := &handler.UserHandler{
		DB:                bunDB,
		ChatServiceClient: chatServiceClient, // Используем правильный клиент
	}

	postHandler := &handler.PostHandler{
		DB: bunDB,
	}

	// Настройка маршрутов
	router.SetupRoutes(e, userHandler, postHandler, bunDB)

	// Запуск HTTP-сервера с поддержкой graceful shutdown
	server := &http.Server{
		Addr:    ":8080",
		Handler: e,
	}

	go func() {
		log.Println("HTTP-сервер запущен на порту 8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка работы HTTP-сервера: %v", err)
		}
	}()

	// Ожидаем сигнал завершения
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	// Корректно завершаем работу
	log.Println("Завершаем работу HTTP-сервера...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Ошибка при завершении HTTP-сервера: %v", err)
	}

	log.Println("Завершаем работу gRPC-сервера...")
	grpcServer.GracefulStop()
	log.Println("Сервер успешно завершил работу")
}