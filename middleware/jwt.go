package middleware

import (
    "net/http"
    "github.com/labstack/echo/v4"
    "github.com/golang-jwt/jwt/v4"
)

func JWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
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
