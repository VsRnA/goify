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
- 🎯 **URL параметры**: Поддержка :param и *wildcard параметров
- 📁 **Группы маршрутов**: Организация маршрутов с префиксами и групповыми middleware
- ✅ **Валидация**: Мощная система валидации с struct tags и custom валидаторами
- 📤 **Загрузка файлов**: Поддержка multipart form с валидацией файлов
- 🔄 **Graceful Shutdown**: Корректное завершение сервера с обработкой сигналов
- 🏥 **Health Checks**: Встроенные проверки состояния приложения

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

## Graceful Shutdown и Health Checks

Goify обеспечивает корректное завершение сервера и мониторинг состояния приложения:

### Graceful Shutdown

#### Базовое использование
```go
app := goify.New()

app.GET("/", func(c *goify.Context) {
    c.SendSuccess(goify.H{"message": "Сервер работает"})
})

if err := app.ListenAndServeWithGracefulShutdown(":3000"); err != nil {
    log.Fatal(err)
}
```

#### Настройка shutdown
```go
config := goify.ShutdownConfig{
    Timeout: 30 * time.Second,
}

app.OnShutdown(func() {
    log.Println("Закрытие соединений с базой данных...")
    db.Close()
})

app.OnShutdown(func() {
    log.Println("Сохранение данных в кэше...")
    cache.Save()
})

app.ListenAndServeWithGracefulShutdown(":3000", config)
```

#### Ручное управление
```go
go func() {
    if err := app.Listen(":3000"); err != nil {
        log.Fatal(err)
    }
}()

quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
<-quit

ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := app.Shutdown(ctx); err != nil {
    log.Fatal("Ошибка при завершении сервера:", err)
}
```

### Health Checks

#### Настройка приложения
```go
goify.SetAppInfo("1.0.0", "production")

goify.RegisterHealthCheck("database", goify.DatabaseHealthCheck(func() error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    return db.PingContext(ctx)
}))

goify.RegisterHealthCheck("redis", goify.RedisHealthCheck(func() error {
    return redisClient.Ping().Err()
}))

goify.RegisterHealthCheck("memory", goify.MemoryHealthCheck(500))
goify.RegisterHealthCheck("disk", goify.DiskSpaceHealthCheck("/", 10))
```

#### Health endpoints
```go
app.GET("/health", goify.HealthCheckHandler())

app.GET("/liveness", func(c *goify.Context) {
    c.JSON(200, goify.H{
        "status": "alive",
        "timestamp": time.Now(),
    })
})

app.GET("/readiness", func(c *goify.Context) {
    c.JSON(200, goify.H{
        "status": "ready",
        "timestamp": time.Now(),
    })
})
```

#### Пользовательские health checks
```go
goify.RegisterHealthCheck("external_api", func() goify.HealthCheck {
    client := &http.Client{Timeout: 5 * time.Second}
    resp, err := client.Get("https://api.example.com/status")
    
    if err != nil {
        return goify.HealthCheck{
            Name:    "external_api",
            Status:  goify.StatusUnhealthy,
            Message: err.Error(),
        }
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != 200 {
        return goify.HealthCheck{
            Name:    "external_api",
            Status:  goify.StatusDegraded,
            Message: fmt.Sprintf("API returned %d", resp.StatusCode),
        }
    }
    
    return goify.HealthCheck{
        Name:    "external_api",
        Status:  goify.StatusHealthy,
        Message: "External API is responding",
    }
})
```

### Встроенные Health Checks

#### Database Health Check
```go
goify.RegisterHealthCheck("postgres", goify.DatabaseHealthCheck(func() error {
    return db.Ping()
}))
```

#### Redis Health Check
```go
goify.RegisterHealthCheck("redis", goify.RedisHealthCheck(func() error {
    return redisClient.Ping().Err()
}))
```

#### Memory Health Check
```go
goify.RegisterHealthCheck("memory", goify.MemoryHealthCheck(500))
```

#### Disk Space Health Check
```go
goify.RegisterHealthCheck("disk", goify.DiskSpaceHealthCheck("/var/lib/app", 5))
```

### Health Check Response

Health endpoint возвращает следующую структуру:

```json
{
    "status": "healthy",
    "timestamp": "2024-01-01T12:00:00Z",
    "uptime": "2h 30m 45s",
    "version": "1.0.0",
    "environment": "production",
    "checks": {
        "database": {
            "name": "database",
            "status": "healthy",
            "message": "Database connection is healthy",
            "last_checked": "2024-01-01T12:00:00Z",
            "duration": "2ms"
        },
        "memory": {
            "name": "memory",
            "status": "healthy",
            "message": "Memory usage is normal: 45MB",
            "last_checked": "2024-01-01T12:00:00Z",
            "duration": "1ms",
            "data": {
                "allocated_mb": 45,
                "max_mb": 500
            }
        }
    }
}
```

### Статусы Health Checks

- **healthy** - Все проверки прошли успешно
- **degraded** - Есть проблемы, но сервис работает
- **unhealthy** - Критические проблемы, сервис недоступен


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

// URL параметры
app.GET("/users/:id", func(c *goify.Context) {
    userID := c.Param("id") // Получить URL параметр
    c.SendSuccess(goify.H{"user_id": userID})
})

// Множественные параметры
app.GET("/users/:userId/posts/:postId", func(c *goify.Context) {
    userID := c.Param("userId")
    postID := c.Param("postId")
    // Обработка обоих параметров
})

// Wildcard параметры (захват всего)
app.GET("/files/*filepath", func(c *goify.Context) {
    filepath := c.Param("filepath") // Получает всё после /files/
})
```

### Graceful Shutdown

```go
app := goify.New()

app.GET("/", func(c *goify.Context) {
    c.SendSuccess(goify.H{"message": "Сервер работает"})
})

app.OnShutdown(func() {
    log.Println("Закрытие соединений с БД...")
})

config := goify.ShutdownConfig{
    Timeout: 30 * time.Second,
}

if err := app.ListenAndServeWithGracefulShutdown(":3000", config); err != nil {
    log.Fatal(err)
}
```

### Health Checks

```go
goify.SetAppInfo("1.0.0", "production")

goify.RegisterHealthCheck("database", goify.DatabaseHealthCheck(func() error {
    return db.Ping()
}))

goify.RegisterHealthCheck("redis", goify.RedisHealthCheck(func() error {
    return redisClient.Ping().Err()
}))

goify.RegisterHealthCheck("memory", goify.MemoryHealthCheck(500))

app.GET("/health", goify.HealthCheckHandler())

app.GET("/liveness", func(c *goify.Context) {
    c.JSON(200, goify.H{"status": "alive"})
})
```

### Загрузка файлов

```go
// Одиночная загрузка файла с валидацией
app.POST("/upload", func(c *goify.Context) {
    file, err := c.FormFile("avatar")
    if err != nil {
        c.SendBadRequest("Файл обязателен")
        return
    }
    
    // Валидация файла
    validation := goify.FileValidation{
        MaxSize:      5 * 1024 * 1024, // 5MB
        AllowedTypes: []string{"image/jpeg", "image/png"},
        AllowedExts:  []string{".jpg", ".jpeg", ".png"},
        Required:     true,
    }
    
    if err := c.ValidateFile(file, validation); err != nil {
        c.SendFileUploadError(err)
        return
    }
    
    // Сохранение файла
    savedPath, err := c.SaveUploadedFile(file, "./uploads/")
    if err != nil {
        c.SendInternalError("Ошибка сохранения файла")
        return
    }
    
    c.SendCreated(goify.H{
        "filename":   file.Filename,
        "path":       savedPath,
        "size":       goify.FormatFileSize(file.Size),
        "url":        "/uploads/" + filepath.Base(savedPath),
    })
})

// Загрузка нескольких файлов
app.POST("/upload/multiple", func(c *goify.Context) {
    files, err := c.FormFiles("files")
    if err != nil {
        c.SendBadRequest("Файлы обязательны")
        return
    }
    
    // Валидация всех файлов
    validation := goify.FileValidation{MaxSize: 10 * 1024 * 1024}
    if errors := c.ValidateFiles(files, validation); len(errors) > 0 {
        c.SendFileUploadError(errors)
        return
    }
    
    // Обработка файлов...
})

// Multipart форма с файлами и данными
type UploadRequest struct {
    Title string            `form:"title" validate:"required"`
    File  *goify.FileHeader `form:"file"`
}

app.POST("/upload/form", func(c *goify.Context) {
    var req UploadRequest
    
    if err := c.BindMultipart(&req); err != nil {
        c.SendBadRequest("Некорректные данные формы")
        return
    }
    
    // Валидация и обработка...
})
```

### Группы маршрутов

```go
// Версионирование API
v1 := app.Group("/api/v1")
v1.GET("/users", handler)     // /api/v1/users
v1.POST("/users", handler)    // /api/v1/users

// Вложенные группы
admin := v1.Group("/admin")
admin.GET("/stats", handler)  // /api/v1/admin/stats

// Групповые middleware
api := app.Group("/api")
api.Use(goify.Logger())
api.Use(authMiddleware)
api.GET("/protected", handler)
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

### Валидация запросов

```go
// Определение модели с валидационными тегами
type User struct {
    Name  string `json:"name" validate:"required,min=2,max=50"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"required,min=18,max=120"`
}

// Автоматическая валидация
app.POST("/users", func(c *goify.Context) {
    var user User
    
    // Привязка и валидация одной командой
    if err := c.BindAndValidate(&user); err != nil {
        c.SendValidationError(err)
        return
    }
    
    c.SendCreated(user, "Пользователь создан")
})

// Валидация query параметров
type QueryParams struct {
    Page  int    `query:"page" validate:"min=1,max=1000"`
    Limit int    `query:"limit" validate:"min=1,max=100"`
    Sort  string `query:"sort" validate:"oneof=name email age"`
}

app.GET("/users", func(c *goify.Context) {
    var params QueryParams
    params.Page = 1 // значения по умолчанию
    
    if err := c.ValidateQuery(&params); err != nil {
        c.SendValidationError(err)
        return
    }
    
    // params теперь содержит валидные данные
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
- [Пример Middleware](./examples/middleware/main.go) - Комплексное использование middleware
- [Группы и параметры](./examples/groups-params/main.go) - URL параметры и группы маршрутов
- [Валидация](./examples/validation/main.go) - Валидация запросов с struct tags
- [Загрузка файлов](./examples/upload/main.go) - Загрузка файлов с валидацией
- [Graceful Shutdown](./examples/shutdown/main.go) - Корректное завершение и health checks

## Справочник API

### Методы Router

- `New()` - Создать новый экземпляр роутера
- `Use(middleware...)` - Добавить middleware к роутеру
- `Group(prefix)` - Создать группу маршрутов с префиксом
- `OnShutdown(fn)` - Добавить функцию для выполнения при завершении
- `Shutdown(ctx)` - Корректно завершить сервер
- `ListenAndServeWithGracefulShutdown(addr, config)` - Запуск с graceful shutdown
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
- `BindAndValidate(obj)` - Привязать JSON и валидировать
- `ValidateStruct(obj)` - Валидировать структуру
- `ValidateQuery(obj)` - Валидировать query параметры
- `FormFile(key)` - Получить загруженный файл
- `FormFiles(key)` - Получить множественные файлы
- `BindMultipart(obj)` - Привязать multipart форму к структуре
- `ValidateFile(file, validation)` - Валидировать загруженный файл
- `ValidateFiles(files, validation)` - Валидировать множественные файлы
- `SaveUploadedFile(file, dir)` - Сохранить загруженный файл
- `GetUploadedFileInfo(key)` - Получить информацию о файле
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
- `SendValidationError(errors)` - Отправить ответ 422 с ошибками валидации
- `SendFileUploadError(errors)` - Отправить ошибку загрузки файла
- `SendFieldError(field, message)` - Отправить ошибку для конкретного поля
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
    "github.com/VsRnA/goify"
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
            Name  string `json:"name" validate:"required,min=2,max=50"`
            Email string `json:"email" validate:"required,email"`
        }

        if err := c.BindAndValidate(&user); err != nil {
            c.SendValidationError(err)
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

## URL Параметры

Goify поддерживает несколько типов URL параметров:

### Одиночные параметры
```go
app.GET("/users/:id", func(c *goify.Context) {
    userID := c.Param("id") // Получить значение параметра
})
// Соответствует: /users/123, /users/abc
```

### Множественные параметры
```go
app.GET("/users/:userId/posts/:postId", func(c *goify.Context) {
    userID := c.Param("userId")
    postID := c.Param("postId")
})
// Соответствует: /users/123/posts/456
```

### Wildcard параметры
```go
app.GET("/files/*filepath", func(c *goify.Context) {
    filepath := c.Param("filepath") // Получает весь остальной путь
})
```

## Группы маршрутов

Группы позволяют организовать маршруты с общими префиксами и middleware:

### Базовые группы
```go
api := app.Group("/api")
api.GET("/users", handler)    // /api/users
api.POST("/users", handler)   // /api/users
```

### Вложенные группы
```go
v1 := app.Group("/api/v1")
admin := v1.Group("/admin")
admin.GET("/stats", handler)  // /api/v1/admin/stats
```

### Групповые middleware
```go
api := app.Group("/api")
api.Use(authMiddleware)       // Применяется ко всем маршрутам группы
api.Use(loggingMiddleware)

api.GET("/protected", handler) // Требует аутентификации
```

### Комбинирование с параметрами
```go
users := app.Group("/users")
users.GET("/:id", getUserHandler)           // /users/123
users.GET("/:id/posts", getUserPostsHandler) // /users/123/posts
users.POST("/:id/posts", createPostHandler)  // /users/123/posts
```

## Валидация запросов

Goify предоставляет мощную систему валидации с поддержкой struct tags:

### Встроенные валидаторы

#### Обязательные поля
```go
type User struct {
    Name string `json:"name" validate:"required"`
}
```

#### Длина и размер
```go
type User struct {
    Name     string `json:"name" validate:"min=2,max=50"`
    Age      int    `json:"age" validate:"min=18,max=120"`
    Tags     []string `json:"tags" validate:"max=5"`
}
```

#### Форматы
```go
type User struct {
    Email   string `json:"email" validate:"email"`
    Website string `json:"website" validate:"url"`
    Phone   string `json:"phone" validate:"numeric"`
}
```

#### Ограничения
```go
type User struct {
    Role     string `json:"role" validate:"oneof=admin user moderator"`
    Username string `json:"username" validate:"alphanum"`
    Name     string `json:"name" validate:"alpha"`
}
```

### Пользовательские валидаторы

```go
// Регистрация пользовательского валидатора
goify.RegisterValidator("strong_password", func(value interface{}, param string) error {
    password := value.(string)
    if len(password) < 8 {
        return fmt.Errorf("пароль должен содержать минимум 8 символов")
    }
    // дополнительные проверки...
    return nil
})

// Использование
type User struct {
    Password string `json:"password" validate:"required,strong_password"`
}
```

### Валидация вложенных структур

```go
type Address struct {
    Street  string `json:"street" validate:"required,min=5"`
    City    string `json:"city" validate:"required"`
    Country string `json:"country" validate:"required,min=2"`
}

type User struct {
    Name    string  `json:"name" validate:"required"`
    Address Address `json:"address" validate:"required"`
}
```

### Обработка ошибок валидации

```go
app.POST("/users", func(c *goify.Context) {
    var user User
    
    if err := c.BindAndValidate(&user); err != nil {
        // Автоматически отправляет структурированные ошибки
        c.SendValidationError(err)
        return
    }
    
    // Данные валидны
    c.SendCreated(user)
})
```

### Полный список валидаторов

| Валидатор | Описание | Пример |
|-----------|----------|---------|
| `required` | Обязательное поле | `validate:"required"` |
| `min=N` | Минимальная длина/значение | `validate:"min=2"` |
| `max=N` | Максимальная длина/значение | `validate:"max=50"` |
| `email` | Email формат | `validate:"email"` |
| `url` | URL формат | `validate:"url"` |
| `alpha` | Только буквы | `validate:"alpha"` |
| `alphanum` | Буквы и цифры | `validate:"alphanum"` |
| `numeric` | Только цифры | `validate:"numeric"` |
| `oneof=a b c` | Одно из значений | `validate:"oneof=admin user"` |

## Загрузка файлов

Goify поддерживает загрузку файлов через multipart forms с мощной системой валидации:

### Основные возможности

#### Одиночная загрузка файла
```go
app.POST("/upload", func(c *goify.Context) {
    file, err := c.FormFile("file")
    if err != nil {
        c.SendBadRequest("Файл обязателен")
        return
    }
    
    // Сохранение с автоматическим именем
    savedPath, err := c.SaveUploadedFile(file, "./uploads/")
    if err != nil {
        c.SendInternalError("Ошибка сохранения")
        return
    }
    
    c.SendCreated(goify.H{"path": savedPath})
})
```

#### Множественная загрузка
```go
app.POST("/upload/multiple", func(c *goify.Context) {
    files, err := c.FormFiles("files")
    if err != nil {
        c.SendBadRequest("Файлы обязательны")
        return
    }
    
    var savedFiles []string
    for _, file := range files {
        path, _ := c.SaveUploadedFile(file, "./uploads/")
        savedFiles = append(savedFiles, path)
    }
    
    c.SendCreated(goify.H{"files": savedFiles})
})
```

### Валидация файлов

#### Базовая валидация
```go
validation := goify.FileValidation{
    MaxSize:      5 * 1024 * 1024, // 5MB
    MinSize:      1024,             // 1KB
    AllowedTypes: []string{"image/jpeg", "image/png"},
    AllowedExts:  []string{".jpg", ".jpeg", ".png"},
    Required:     true,
}

if err := c.ValidateFile(file, validation); err != nil {
    c.SendFileUploadError(err)
    return
}
```

#### Валидация по категориям
```go
func getValidationForCategory(category string) goify.FileValidation {
    switch category {
    case "image":
        return goify.FileValidation{
            MaxSize:      5 * 1024 * 1024,
            AllowedTypes: []string{"image/jpeg", "image/png", "image/gif"},
            AllowedExts:  []string{".jpg", ".jpeg", ".png", ".gif"},
        }
    case "document":
        return goify.FileValidation{
            MaxSize:      10 * 1024 * 1024,
            AllowedTypes: []string{"application/pdf", "text/plain"},
            AllowedExts:  []string{".pdf", ".txt", ".doc", ".docx"},
        }
    case "video":
        return goify.FileValidation{
            MaxSize:      50 * 1024 * 1024,
            AllowedTypes: []string{"video/mp4", "video/avi"},
            AllowedExts:  []string{".mp4", ".avi", ".mov"},
        }
    }
    return goify.FileValidation{}
}
```

### Multipart формы с данными

```go
type FileUploadRequest struct {
    Title       string            `form:"title" validate:"required"`
    Description string            `form:"description" validate:"max=500"`
    Category    string            `form:"category" validate:"oneof=image document video"`
    File        *goify.FileHeader `form:"file"`
    IsPublic    bool              `form:"is_public"`
}

app.POST("/upload/form", func(c *goify.Context) {
    var req FileUploadRequest
    
    // Привязка формы (включая файлы)
    if err := c.BindMultipart(&req); err != nil {
        c.SendBadRequest("Некорректные данные формы")
        return
    }
    
    // Валидация данных формы
    if validationErrors := c.ValidateStruct(&req); len(validationErrors) > 0 {
        c.SendValidationError(validationErrors)
        return
    }
    
    // Валидация файла
    validation := getValidationForCategory(req.Category)
    if err := c.ValidateFile(req.File, validation); err != nil {
        c.SendFileUploadError(err)
        return
    }
    
    // Сохранение и обработка...
})
```

### Утилиты для работы с файлами

```go
// Получение информации о файле
fileInfo, err := c.GetUploadedFileInfo("file")
// Возвращает: filename, size, size_human, content_type, is_image

// Проверка существования файла
exists := goify.FileExists("./uploads/file.jpg")

// Получение размера файла
size, err := goify.GetFileSize("./uploads/file.jpg")

// Форматирование размера
humanSize := goify.FormatFileSize(1024576) // "1.0 MB"

// Определение MIME типа
mimeType := goify.GetMimeType("image.jpg") // "image/jpeg"

// Проверка типа изображения
isImage := goify.IsImageFile("image/jpeg") // true

// Удаление файла
err := goify.DeleteFile("./uploads/file.jpg")
```

### Обработка ошибок загрузки

```go
// Специфичные ошибки загрузки файлов
c.SendFileUploadError(err)           // Общая ошибка загрузки
c.SendFileTooBigError(maxSize)       // Файл слишком большой

// Пример обработки разных типов ошибок
if err := c.ValidateFile(file, validation); err != nil {
    if uploadErr, ok := err.(goify.FileUploadError); ok {
        switch uploadErr.Code {
        case "max_size":
            c.SendFileTooBigError(validation.MaxSize)
        case "invalid_type":
            c.SendFileUploadError(goify.FileUploadError{
                Message: "Недопустимый тип файла",
                Code:    "invalid_type",
            })
        default:
            c.SendFileUploadError(err)
        }
    }
    return
}
```

## Производительность

Goify построен на стандартной библиотеке Go и спроектирован для высокой производительности:

- Отсутствие reflection в критических путях
- Минимальные аллокации памяти
- Эффективная обработка middleware
- Быстрая маршрутизация