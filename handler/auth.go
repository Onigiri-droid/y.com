// package handler

// import (
//     "net/http"
//     "time"
//     "api-service/model"
//     "github.com/labstack/echo/v4"
//     "github.com/golang-jwt/jwt/v4"
//     "golang.org/x/crypto/bcrypt"
// )

// var jwtSecret = []byte("supersecretkey")

// // Структура для ответа с токеном
// type TokenResponse struct {
//     Token string `json:"token"`
// }

// // Хэширование пароля
// func hashPassword(password string) (string, error) {
//     hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
//     if err != nil {
//         return "", err
//     }
//     return string(hashedPassword), nil
// }

// // Регистрация пользователя
// func (h *UserHandler) Register(c echo.Context) error {
//     user := new(model.User)
//     if err := c.Bind(user); err != nil {
//         return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверные данные"})
//     }

//     // Проверка на наличие email в базе
//     exists, err := h.isEmailExist(user.Email)
//     if err != nil {
//         return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка проверки email"})
//     }
//     if exists {
//         return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email уже существует"})
//     }

//     // Валидация email
//     if user.Email == "" || !isValidEmail(user.Email) {
//         return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный email"})
//     }

//     // Валидация пароля (например, минимум 10 символов)
//     if len(user.Password) < 10 {
//         return c.JSON(http.StatusBadRequest, map[string]string{"error": "Пароль должен содержать минимум 10 символов"})
//     }

//     // Хэшируем пароль перед сохранением
//     hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
//     if err != nil {
//         return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка хэширования пароля"})
//     }
//     user.Password = string(hashedPassword)

//     _, err = h.DB.NewInsert().Model(user).Exec(c.Request().Context())
//     if err != nil {
//         return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка создания пользователя"})
//     }

//     return c.JSON(http.StatusCreated, user)
// }


// // Логин пользователя
// func (h *UserHandler) Login(c echo.Context) error {
//     login := new(model.LoginRequest)
//     if err := c.Bind(login); err != nil {
//         return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверные данные"})
//     }

//     var user model.User
//     err := h.DB.NewSelect().Model(&user).Where("email = ?", login.Email).Scan(c.Request().Context())
//     if err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password)) != nil {
//         return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Неверный логин или пароль"})
//     }

//     // Генерация JWT токена
//     token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
//         "user_id": user.ID,
//         "exp":     time.Now().Add(time.Hour * 12).Unix(),
//     })

//     tokenString, err := token.SignedString(jwtSecret)
//     if err != nil {
//         return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Не удалось сгенерировать токен"})
//     }

//     return c.JSON(http.StatusOK, TokenResponse{Token: tokenString})
// }


package handler

import (
    "net/http"
    "time"
    "api-service/model"
    "api-service/utils"
    "github.com/labstack/echo/v4"
    "github.com/golang-jwt/jwt/v4"
)

var jwtSecret = []byte("supersecretkey")

// Структура для ответа с токеном
type TokenResponse struct {
    Token string `json:"token"`
}

// Регистрация пользователя
func (h *UserHandler) Register(c echo.Context) error {
    req := new(model.CreateUserRequest)
    if err := c.Bind(req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверные данные"})
    }

    // Проверка на наличие email
    exists, err := h.isEmailExist(req.Email)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка проверки email"})
    }
    if exists {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email уже существует"})
    }

    // Хэшируем пароль
    hashedPassword, err := utils.HashPassword(req.Password)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка хэширования пароля"})
    }

    user := &model.User{
        Name:     req.Name,
        Email:    req.Email,
        Password: hashedPassword,
    }

    // Сохраняем пользователя
    _, err = h.DB.NewInsert().Model(user).Exec(c.Request().Context())
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка создания пользователя"})
    }

    return c.JSON(http.StatusCreated, user)
}

// Логин пользователя
func (h *UserHandler) Login(c echo.Context) error {
    req := new(model.LoginRequest)
    if err := c.Bind(req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверные данные"})
    }

    // Проверяем наличие пользователя
    var user model.User
    err := h.DB.NewSelect().Model(&user).Where("email = ?", req.Email).Scan(c.Request().Context())
    if err != nil {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Неверный логин или пароль"})
    }

    // Проверяем пароль
    if !utils.CheckPassword(user.Password, req.Password) {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Неверный логин или пароль"})
    }

    // Генерируем JWT токен
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": user.ID,
        "exp":     time.Now().Add(12 * time.Hour).Unix(),
    })

    tokenString, err := token.SignedString(jwtSecret)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Не удалось сгенерировать токен"})
    }

    return c.JSON(http.StatusOK, TokenResponse{Token: tokenString})
}
