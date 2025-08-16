# Goify 🚀

Быстрый, легковесный и выразительный HTTP веб-фреймворк для Go, вдохновленный Fastify и Express.js.

## Возможности

- ⚡ **Быстрый**: Построен на стандартной библиотеке Go для максимальной производительности
- 🪶 **Легковесный**: Без внешних зависимостей
- 🎯 **Простой**: Интуитивный API похожий на Express.js/Fastify
- 🔧 **Гибкий**: Легко расширяется и настраивается
- 📝 **Богатый контекст**: Комплексная обработка запросов/ответов
- 🔗 **Middleware**: Мощная система middleware с встроенными middleware
- 🛡️ **Безопасный**: Встроенные CORS, recovery и authentication middleware

## Установка

```bash
go get github.com/VsRnA/goify
```

## Быстрый старт

```go
package main

import (
    "log"
    "github.com/VsRnA/goify"
)

func main() {
    app := goify.New()

    app.GET("/", func(c *goify.Context) {
        c.JSON(200, goify.H{
            "message": "Привет от Goify!",
        })
    })

    log.Fatal(app.Listen(":3000"))
}
```

## Основное использование

### Маршрутизация

```go
app := goify.New()

// HTTP методы
app.GET("/users", getUsersHandler)
app.POST("/users", createUserHandler)
app.PUT("/users/:id", updateUserHandler)
app.DELETE("/users/:id", deleteUserHandler)
app.PATCH("/users/:id", patchUserHandler)
```

### Обработка запросов

```go
app.GET("/users", func(c *goify.Context) {
    // Параметры запроса
    name := c.Query("name")
    age, _ := c.QueryInt("age")
    page := c.QueryDefault("page", "1")
    
    // Заголовки
    userAgent := c.GetHeader("User-Agent")
    
    // Привязка JSON
    var user User
    if err := c.BindJSON(&user); err != nil {
        c.SendBadRequest("Некорректный JSON")
        return
    }
    
    c.SendSuccess(user)
})
```

### Помощники ответов

```go
// JSON ответы
c.JSON(200, goify.H{"key": "value"})
c.SendSuccess(data, "Операция выполнена успешно")
c.SendCreated(newUser, "Пользователь создан")

// Ответы с ошибками
c.SendBadRequest("Некорректные данные")
c.SendNotFound("Пользователь не найден")
c.SendInternalError("Что-то пошло не так")

// Другие ответы
c.String(200, "Привет %s", name)
c.HTML(200, "<h1>Привет мир</h1>")
c.Redirect(302, "/login")
```

### Middleware

```go
app := goify.New()

// Встроенные middleware
app.Use(goify.Logger())
app.Use(goify.Recovery())
app.Use(goify.CORS())
app.Use(goify.RequestID())

// Ограничение частоты запросов
rateLimiter := goify.NewRateLimiter(100, time.Minute)
app.Use(rateLimiter.Middleware())

// Аутентификация
app.Use(goify.BasicAuth("username", "password"))

// Статические файлы
app.Use(goify.Static("/static", "./public"))

// Пользовательский middleware
app.Use(func(c *goify.Context, next func()) {
    // До обработки запроса
    start := time.Now()
    c.Set("start", start)
    
    next() // Переход к следующему middleware/обработчику
    
    // После обработки запроса
    duration := time.Since(start)
    log.Printf("Запрос выполнен за %v", duration)
})
```

## Примеры

Посмотрите папку [examples](./examples/) для более подробных примеров использования:

- [Базовый пример](./examples/basic/main.go) - Простые CRUD операции

## Справочник API

### Методы Router

- `New()` - Создать новый экземпляр роутера
- `Use(middleware...)` - Добавить middleware к роутеру
- `GET(path, handler)` - Зарегистрировать GET маршрут
- `POST(path, handler)` - Зарегистрировать POST маршрут
- `PUT(path, handler)` - Зарегистрировать PUT маршрут
- `DELETE(path, handler)` - Зарегистрировать DELETE маршрут
- `PATCH(path, handler)` - Зарегистрировать PATCH маршрут
- `Listen(addr)` - Запустить сервер

### Методы Context

#### Запрос
- `Query(key)` - Получить параметр запроса
- `QueryDefault(key, default)` - Получить параметр запроса со значением по умолчанию
- `QueryInt(key)` - Получить параметр запроса как целое число
- `Param(key)` - Получить URL параметр
- `GetHeader(key)` - Получить заголовок запроса
- `BindJSON(obj)` - Привязать JSON к структуре
- `Body()` - Получить сырое тело запроса
- `Set(key, value)` - Сохранить значение в контексте
- `Get(key)` - Получить значение из контекста

#### Ответ
- `JSON(code, obj)` - Отправить JSON ответ
- `String(code, format, values...)` - Отправить текстовый ответ
- `HTML(code, html)` - Отправить HTML ответ
- `SendSuccess(data, message?)` - Отправить успешный ответ
- `SendError(code, message, details?)` - Отправить ответ с ошибкой
- `SendCreated(data, message?)` - Отправить ответ 201
- `SendBadRequest(message, details?)` - Отправить ответ 400
- `SendNotFound(message?)` - Отправить ответ 404
- `SetHeader(key, value)` - Установить заголовок ответа
- `Redirect(code, location)` - Отправить редирект

## Встроенные Middleware

### Logger
Логирует HTTP запросы с временем выполнения:
```go
app.Use(goify.Logger())
```

### Recovery
Восстанавливается после паник и отправляет 500 ошибку:
```go
app.Use(goify.Recovery())
```

### CORS
Добавляет CORS заголовки:
```go
// Базовый CORS
app.Use(goify.CORS())

// CORS с настройками
app.Use(goify.CORSWithConfig(goify.CORSConfig{
    AllowOrigins: []string{"http://localhost:3000", "https://myapp.com"},
    AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders: []string{"Content-Type", "Authorization"},
}))
```

### BasicAuth
Базовая HTTP аутентификация:
```go
app.Use(goify.BasicAuth("admin", "secret"))
```

### RateLimit
Ограничение частоты запросов:
```go
rateLimiter := goify.NewRateLimiter(100, time.Minute) // 100 запросов в минуту
app.Use(rateLimiter.Middleware())
```

### Static
Обслуживание статических файлов:
```go
app.Use(goify.Static("/static", "./public"))
```

### RequestID
Добавляет уникальный ID к каждому запросу:
```go
app.Use(goify.RequestID())

// В обработчике
app.GET("/", func(c *goify.Context) {
    requestID, _ := c.Get("requestID")
    // использовать requestID
})
```

## Полный пример

```go
package main

import (
    "log"
    "time"
    "github.com/yourusername/goify"
)

func main() {
    app := goify.New()

    // Глобальные middleware
    app.Use(goify.Logger())
    app.Use(goify.Recovery())
    app.Use(goify.CORS())
    app.Use(goify.RequestID())

    // Ограничение частоты запросов
    rateLimiter := goify.NewRateLimiter(10, time.Minute)
    app.Use(rateLimiter.Middleware())

    // Статические файлы
    app.Use(goify.Static("/static", "./static"))

    // Маршруты
    app.GET("/", func(c *goify.Context) {
        c.SendSuccess(goify.H{
            "message": "Привет от Goify!",
            "version": "1.0.0",
        })
    })

    app.POST("/users", func(c *goify.Context) {
        var user struct {
            Name  string `json:"name"`
            Email string `json:"email"`
        }

        if err := c.BindJSON(&user); err != nil {
            c.SendBadRequest("Некорректный JSON")
            return
        }

        if user.Name == "" || user.Email == "" {
            c.SendBadRequest("Имя и email обязательны")
            return
        }

        c.SendCreated(goify.H{
            "id":    123,
            "name":  user.Name,
            "email": user.Email,
        }, "Пользователь создан успешно")
    })

    // Защищенный маршрут
    adminGroup := app.Group("/admin")
    adminGroup.Use(goify.BasicAuth("admin", "secret"))
    adminGroup.GET("/dashboard", func(c *goify.Context) {
        c.SendSuccess(goify.H{
            "message": "Добро пожаловать в админ панель!",
        })
    })

    log.Println("🚀 Сервер запущен на http://localhost:3000")
    log.Fatal(app.Listen(":3000"))
}
```


## Производительность

Goify построен на стандартной библиотеке Go и спроектирован для высокой производительности:

- Отсутствие reflection в критических путях
- Минимальные аллокации памяти
- Эффективная обработка middleware
- Быстрая маршрутизация

