# Запуск приложения с Redis кэшированием

## 🚀 Быстрый старт с Docker Compose

### 1. **Запустить все сервисы**
```bash
docker-compose up -d
```

Это запустит:
- **PostgreSQL** на `localhost:5433`
- **Redis** на `localhost:6379`
- **API** на `http://localhost:8080`
- **Frontend** на `http://localhost:5173`

### 2. **Проверить статус сервисов**
```bash
docker-compose ps
```

### 3. **Остановить сервисы**
```bash
docker-compose down
```

---

## 🧪 Проверка Redis кэширования

### Локальный запуск (на macOS)

#### 1. **Установить Redis локально**
```bash
brew install redis
brew services start redis
```

#### 2. **Запустить приложение в терминале**
```bash
cd /Users/maksimblohin/backend21
go run cmd/main.go
```

---

## 📊 Тестирование кэширования

Используйте этот скрипт для проверки, что кэширование работает:

```bash
#!/bin/bash

# Токен для авторизации (нужно сначала залогиниться)
TOKEN="your_jwt_token_here"

echo "=== Тест 1: GET /api/users (кэш 1 минута) ==="
echo "Запрос 1:"
time curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/users

echo -e "\n\nЗапрос 2 (из кэша, должно быть быстрее):"
time curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/users

echo -e "\n\n=== Тест 2: GET /api/products (кэш 10 минут) ==="
echo "Запрос 1:"
time curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/products

echo -e "\n\nЗапрос 2 (из кэша, должно быть быстрее):"
time curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/products
```

---

## 🔍 Проверка Redis прямо

### 1. **Подключиться к Redis**
```bash
redis-cli
```

### 2. **Посмотреть все ключи кэша**
```redis
KEYS *
```

### 3. **Проверить конкретные кэш ключи**
```redis
# Список пользователей
GET users:list

# Конкретный пользователь (ID=1)
GET users:1

# Список продуктов
GET products:list

# Конкретный продукт (ID=1)
GET products:1

# Посмотреть TTL ключа
TTL users:list
```

### 4. **Очистить кэш**
```redis
# Очистить все кэши
FLUSHALL

# Удалить конкретный ключ
DEL users:list
```

---

## 📝 Настройки кэширования

| Endpoint | TTL | Ключ в Redis |
|----------|-----|--------------|
| GET /api/users | 1 минута | `users:list` |
| GET /api/users/:id | 1 минута | `users:{id}` |
| GET /api/products | 10 минут | `products:list` |
| GET /api/products/:id | 10 минут | `products:{id}` |

---

## 🔧 Как работает кэширование

1. **При запросе** (например `GET /api/users`):
   - Проверяем Redis на наличие ключа `users:list`
   - Если есть → возвращаем из кэша
   - Если нет → запрашиваем БД и сохраняем в Redis с TTL

2. **При обновлении/удалении**:
   - Обновляем в БД
   - Инвалидируем (удаляем) соответствующие ключи из кэша
   - При следующем запросе кэш пересчитается

---

## 🐛 Проблемы и решения

### Redis не подключается

**Ошибка**: `redis connection failed`

**Решение**:
```bash
# Проверить, запущен ли Redis
redis-cli ping  # должно вывести "PONG"

# Если не запущен
brew services start redis
```

### Если используешь Docker

```bash
# Проверить, запущен ли Redis контейнер
docker ps | grep redis

# Если не запущен
docker-compose up -d redis
```

### Посмотреть логи Redis контейнера
```bash
docker logs web-redis
```

---

## 📍 Переменные окружения

```env
# .env файл
DB_HOST=localhost
DB_PORT=5433
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=webdb
DB_SSLMODE=disable
APP_PORT=8080
REDIS_ADDR=localhost:6379
JWT_SECRET=dev_secret
```

Для Docker Compose переменные переопределяются в `docker-compose.yml`

---

## ✅ Проверка функциональности API

### 1. Регистрация
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@test.com","password":"password123","gender":"M","age":25}'
```

### 2. Лог-ин (получить токен)
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@test.com","password":"password123"}'
```

### 3. Получить список пользователей (с кэшированием)
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/users
```

### 4. Получить список продуктов (с кэшированием)
```bash
curl http://localhost:8080/api/products
```

---

## 🎯 Итого

✅ Redux кэширование добавлено на все GET запросы
✅ Автоматическая инвалидация кэша при обновлении данных
✅ Настроены разные TTL для разных endpoints
✅ Docker Compose готов к использованию
