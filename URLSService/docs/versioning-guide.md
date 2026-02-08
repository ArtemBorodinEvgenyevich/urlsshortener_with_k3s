# Быстрая шпаргалка по версионированию API

## 🚀 Текущие endpoints

### V1 API
```
POST   /api/v1/shorten
GET    /api/v1/urls/{shortCode}
DELETE /api/v1/urls/{shortCode}
GET    /api/v1/health
GET    /api/v1/readiness
```

## 📂 Структура проекта

```
internal/
├── api/
│   ├── middleware/
│   │   └── versioning.go     # Middleware для версионирования
│   └── v1/
│       └── routes.go          # Регистрация V1 routes
├── handler/
│   └── v1/
│       ├── types.go           # V1 DTOs (request/response)
│       ├── url.go             # V1 URL handlers
│       ├── health.go          # V1 Health handlers
│       └── helpers.go         # Общие helper функции
├── service/
│   └── url.go                 # Версионно-независимый service layer
└── repository/
    └── url.go                 # Версионно-независимый data access
```

## 🔄 Как создать V2

### Шаг 1: Создайте структуру

```bash
mkdir -p internal/handler/v2
mkdir -p internal/api/v2
```

### Шаг 2: Определите изменения

Создайте только то, что изменилось:

```go
// internal/handler/v2/types.go
package v2

type CreateURLRequest struct {
    URL         string   `json:"url"`
    TTL         int      `json:"ttl"`
    CustomAlias *string  `json:"custom_alias,omitempty"` // NEW
    Tags        []string `json:"tags,omitempty"`          // NEW
}
```

### Шаг 3: Реализуйте измененные handlers

```go
// internal/handler/v2/url.go
package v2

import (
    "github.com/ArtemBorodinEvgenyevich/URLSService/internal/service"
)

type URLHandler struct {
    service service.URLService  // Тот же service!
}

func (h *URLHandler) Create(w http.ResponseWriter, r *http.Request) {
    var req CreateURLRequest  // V2 request
    // ... парсинг ...

    // Адаптируем V2 → service
    url, err := h.service.CreateShortURL(r.Context(), req.URL, req.TTL)

    // Возвращаем V2 response
    response := URLResponse{...}
    respondWithJSON(w, http.StatusCreated, response)
}
```

### Шаг 4: Настройте routes

```go
// internal/api/v2/routes.go
package v2

import (
    v1 "github.com/ArtemBorodinEvgenyevich/URLSService/internal/handler/v1"
    v2 "github.com/ArtemBorodinEvgenyevich/URLSService/internal/handler/v2"
)

func RegisterRoutes(r chi.Router, cfg *Config) {
    urlHandlerV2 := v2.NewURLHandler(cfg.URLService)
    urlHandlerV1 := v1.NewURLHandler(cfg.URLService)  // Переиспользуем!

    r.Route("/api/v2", func(r chi.Router) {
        r.Use(middleware.APIVersion("v2"))
        r.Use(middleware.CORS())

        // Новая версия
        r.Post("/shorten", urlHandlerV2.Create)

        // Переиспользуем V1 (если не изменились)
        r.Get("/urls/{shortCode}", urlHandlerV1.Get)
        r.Delete("/urls/{shortCode}", urlHandlerV1.Delete)
    })
}
```

### Шаг 5: Регистрируем в main.go

```go
// cmd/main.go

import (
    apiv1 "github.com/.../internal/api/v1"
    apiv2 "github.com/.../internal/api/v2"
)

func main() {
    // ...

    // V1
    apiv1.RegisterRoutes(router, apiConfig)

    // V2
    apiv2.RegisterRoutes(router, apiConfig)

    // Legacy (опционально)
    setupLegacyRoutes(router, ...)
}
```

### Шаг 6: Пометьте V1 как deprecated (опционально)

```go
// internal/api/v1/routes.go

r.Route("/api/v1", func(r chi.Router) {
    r.Use(middleware.APIVersion("v1"))
    r.Use(middleware.Deprecation("2027-12-31", "v2"))  // Добавили
    r.Use(middleware.CORS())
    // ... routes
})
```

## 🧪 Тестирование

```bash
# V1 endpoints
curl -i http://localhost:9091/api/v1/health
# Ожидаем: X-API-Version: v1


# Создание short URL (V1)
curl -X POST http://localhost:9091/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com", "ttl": 3600}'
```
