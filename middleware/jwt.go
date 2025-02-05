// package middleware

// import (
// 	"context"
// 	"net/http"
// 	"github.com/labstack/echo/v4"
// 	"github.com/golang-jwt/jwt/v4"
// 	"github.com/uptrace/bun"
// 	"api-service/model"
// )

// const jwtSecret = "supersecretkey"

// func JWTMiddleware(db *bun.DB) echo.MiddlewareFunc {
// 	return func(next echo.HandlerFunc) echo.HandlerFunc {
// 		return func(c echo.Context) error {
// 			authHeader := c.Request().Header.Get("Authorization")
// 			if authHeader == "" {
// 				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Отсутствует токен аутентификации"})
// 			}

// 			tokenString := authHeader[len("Bearer "):]
// 			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// 				return []byte(jwtSecret), nil
// 			})

// 			if err != nil || !token.Valid {
// 				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Невалидный токен"})
// 			}

// 			claims, ok := token.Claims.(jwt.MapClaims)
// 			if !ok {
// 				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Ошибка обработки токена"})
// 			}

// 			userID, ok := claims["user_id"].(float64)
//             if !ok {
//                 return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Некорректный user_id в токене"})
//             }
//             // Преобразуем float64 в int32
//             userIDInt := int32(userID)
                    
//             // Проверяем, существует ли пользователь в БД
//             exists, err := CheckUserExists(db, userIDInt)
//             if err != nil || !exists {
//                 return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Пользователь не найден"})
//             }
            
//             // Сохраняем user_id в контексте
//             c.Set("user_id", userIDInt)

// 			return next(c)
// 		}
// 	}
// }

// // CheckUserExists проверяет, существует ли пользователь в базе данных
// func CheckUserExists(db *bun.DB, userID int32) (bool, error) {
// 	var exists bool
// 	err := db.NewSelect().
// 		Model((*model.User)(nil)).
// 		Where("id = ?", userID).
// 		Limit(1).
// 		Scan(context.Background(), &exists)
// 	if err != nil {
// 		return false, err
// 	}
// 	return exists, nil
// }

package middleware

import (
    "net/http"
    "github.com/labstack/echo/v4"
    "github.com/golang-jwt/jwt/v4"
    "github.com/uptrace/bun"
)

func JWTMiddleware(db *bun.DB) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            authHeader := c.Request().Header.Get("Authorization")
            if authHeader == "" {
                return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Отсутствует токен аутентификации"})
            }

            tokenString := authHeader[len("Bearer "):]
            token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
                return []byte("supersecretkey"), nil
            })

            if err != nil || !token.Valid {
                return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Невалидный токен"})
            }

            claims, ok := token.Claims.(jwt.MapClaims)
            if !ok {
                return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Ошибка обработки токена"})
            }

            userID, ok := claims["user_id"].(float64) // JWT возвращает числа как float64
            if !ok {
                return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Некорректный user_id в токене"})
            }

            // Сохраняем user_id в контексте
            c.Set("user_id", int(userID))

            return next(c)
        }
    }
}

// func JWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		// Извлекаем заголовок Authorization
// 		authHeader := c.Request().Header.Get("Authorization")
// 		if authHeader == "" {
// 			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Отсутствует токен аутентификации"})
// 		}

// 		// Проверяем формат заголовка (должен начинаться с "Bearer ")
// 		if !strings.HasPrefix(authHeader, "Bearer ") {
// 			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Некорректный формат токена"})
// 		}

// 		// Извлекаем токен
// 		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

// 		// Парсим токен
// 		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// 			// Проверяем алгоритм подписи
// 			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 				return nil, echo.NewHTTPError(http.StatusUnauthorized, "Некорректный метод подписи токена")
// 			}
// 			return []byte("supersecretkey"), nil // Ваш секретный ключ
// 		})

// 		// Проверяем валидность токена
// 		if err != nil || !token.Valid {
// 			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Невалидный токен"})
// 		}

// 		// Получаем claims
// 		claims, ok := token.Claims.(jwt.MapClaims)
// 		if !ok {
// 			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Ошибка обработки токена"})
// 		}

// 		// Извлекаем user_id и проверяем его тип
// 		userID, ok := claims["user_id"].(float64) // JWT возвращает числа как float64
// 		if !ok {
// 			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Некорректный user_id в токене"})
// 		}

// 		// Сохраняем user_id в контексте (конвертируем float64 в int)
// 		c.Set("userID", int(userID))

// 		// Продолжаем выполнение следующего обработчика
// 		return next(c)
// 	}
// }