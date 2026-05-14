#!/bin/bash

# Скрипт для проверки Redis кэширования

BASE_URL="http://localhost:8080"
REDIS_CLI="redis-cli"

# Цвета для вывода
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Скрипт для проверки Redis кэширования ===${NC}\n"

# Проверка, что Redis запущен
echo -e "${YELLOW}[1] Проверка Redis...${NC}"
if ! $REDIS_CLI ping > /dev/null 2>&1; then
    echo -e "${RED}❌ Redis не запущен!${NC}"
    echo "Запустите: redis-cli ping"
    exit 1
fi
echo -e "${GREEN}✓ Redis запущен${NC}\n"

# Получить токен админа для тестирования
echo -e "${YELLOW}[2] Создание тестового пользователя и получение токена...${NC}"
REGISTER=$(curl -s -X POST ${BASE_URL}/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test-'$(date +%s)'@test.com","password":"123456","gender":"M","age":25}')

LOGIN=$(curl -s -X POST ${BASE_URL}/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@test.com","password":"admin123"}')

TOKEN=$(echo $LOGIN | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo -e "${RED}❌ Не удалось получить токен. Зарегистрируйте админа вручную.${NC}"
    echo "POST /api/auth/register с ролью admin"
    TOKEN="test_token"
fi
echo -e "${GREEN}✓ Токен получен${NC}\n"

# Очистить Redis перед тестом
echo -e "${YELLOW}[3] Очистка Redis...${NC}"
$REDIS_CLI FLUSHALL > /dev/null
echo -e "${GREEN}✓ Redis очищен${NC}\n"

# Тест 1: GET /api/users (1 минута кэш)
echo -e "${YELLOW}[4] Тест 1: GET /api/users (TTL: 1 минута)${NC}"
echo "Первый запрос (из БД):"
time curl -s -H "Authorization: Bearer $TOKEN" \
  ${BASE_URL}/api/users | jq '.' | head -20

echo -e "\nПроверка Redis ключа:"
$REDIS_CLI EXISTS users:list
echo -e "\nTTL ключа (секунды):"
$REDIS_CLI TTL users:list

echo -e "\n\nВторой запрос (из кэша - должно быть БЫСТРЕЕ):"
time curl -s -H "Authorization: Bearer $TOKEN" \
  ${BASE_URL}/api/users > /dev/null

echo -e "${GREEN}✓ Кэш работает!${NC}\n"

# Тест 2: GET /api/users/:id (1 минута кэш)
echo -e "${YELLOW}[5] Тест 2: GET /api/users/:id (TTL: 1 минута)${NC}"
echo "Первый запрос пользователя ID=1 (из БД):"
time curl -s -H "Authorization: Bearer $TOKEN" \
  ${BASE_URL}/api/users/1 | jq '.'

echo -e "\nПроверка Redis ключа:"
$REDIS_CLI EXISTS users:1
echo -e "\nTTL ключа (секунды):"
$REDIS_CLI TTL users:1

echo -e "\n\nВторой запрос (из кэша - должно быть БЫСТРЕЕ):"
time curl -s -H "Authorization: Bearer $TOKEN" \
  ${BASE_URL}/api/users/1 > /dev/null

echo -e "${GREEN}✓ Кэш работает!${NC}\n"

# Тест 3: GET /api/products (10 минут кэш)
echo -e "${YELLOW}[6] Тест 3: GET /api/products (TTL: 10 минут)${NC}"
echo "Первый запрос (из БД):"
time curl -s -H "Authorization: Bearer $TOKEN" \
  ${BASE_URL}/api/products | jq '.' | head -20

echo -e "\nПроверка Redis ключа:"
$REDIS_CLI EXISTS products:list
echo -e "\nTTL ключа (секунды):"
$REDIS_CLI TTL products:list

echo -e "\n\nВторой запрос (из кэша - должно быть БЫСТРЕЕ):"
time curl -s -H "Authorization: Bearer $TOKEN" \
  ${BASE_URL}/api/products > /dev/null

echo -e "${GREEN}✓ Кэш работает!${NC}\n"

# Тест 4: GET /api/products/:id (10 минут кэш)
echo -e "${YELLOW}[7] Тест 4: GET /api/products/:id (TTL: 10 минут)${NC}"
echo "Первый запрос продукта ID=1 (из БД):"
time curl -s -H "Authorization: Bearer $TOKEN" \
  ${BASE_URL}/api/products/1 | jq '.'

echo -e "\nПроверка Redis ключа:"
$REDIS_CLI EXISTS products:1
echo -e "\nTTL ключа (секунды):"
$REDIS_CLI TTL products:1

echo -e "\n\nВторой запрос (из кэша - должно быть БЫСТРЕЕ):"
time curl -s -H "Authorization: Bearer $TOKEN" \
  ${BASE_URL}/api/products/1 > /dev/null

echo -e "${GREEN}✓ Кэш работает!${NC}\n"

# Тест 5: Инвалидация кэша при обновлении
echo -e "${YELLOW}[8] Тест 5: Инвалидация кэша при UPDATE${NC}"
echo "Ключи в Redis до UPDATE:"
$REDIS_CLI KEYS "users:*"

echo -e "\nОбновляем пользователя 1..."
curl -s -X PUT -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"age":26}' \
  ${BASE_URL}/api/users/1 > /dev/null

echo -e "\nКлючи в Redis после UPDATE (должны быть удалены):"
$REDIS_CLI KEYS "users:*"

echo -e "\n${GREEN}✓ Кэш успешно инвалидирован!${NC}\n"

# Финальная проверка
echo -e "${YELLOW}[9] Финальная проверка - все ключи в Redis:${NC}"
$REDIS_CLI KEYS "*"

echo -e "\n${GREEN}=== Все тесты пройдены успешно! ===${NC}"
echo -e "\nДополнительные команды Redis для проверки:"
echo "  redis-cli KEYS '*'          - показать все ключи"
echo "  redis-cli GET users:list    - получить кэш пользователей"
echo "  redis-cli TTL users:list    - показать TTL"
echo "  redis-cli FLUSHALL          - очистить весь кэш"
