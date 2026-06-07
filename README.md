# Учебный банковский сервис на Go

REST API для банковского обслуживания с простым веб-интерфейсом. Поддерживает управление счетами, выпуск карт, переводы, кредитование и финансовую аналитику.

## Возможности   

### Пользовательские операции
- Регистрация с проверкой уникальности email и username
- JWT аутентификация (токен живет 24 часа)

### Управление счетами
- Создание банковских счетов
- Пополнение баланса
- Переводы между счетами

### Операции с картами
- Генераци  я виртуальных карт (алгоритм Луна)
- Шифрование данных карт (PGP + HMAC)
- Хеширование CVV (bcrypt)
- Оплата картой

### Кредитные операции
- Оформление кредита с аннуитетными платежами
- Автоматическое списание платежей (шедулер)
- Генерация графика платежей
- Начисление штрафов за просрочку (+10%)

### Аналитика
- Статистика доходов/расходов за месяц
- Аналитика кредитной нагрузки
- Прогноз баланса на N дней (до 365 дней)

### Интеграции
- Получение ключевой ставки ЦБ РФ (SOAP)
- Email уведомления (SMTP)

## Технологии

| Компонент | Технология |
|-----------|------------|
| Язык | Go 1.23+ |
| Маршрутизация | gorilla/mux |
| База данных | PostgreSQL 17 |
| Аутентификация | JWT (golang-jwt) |
| Шифрование | bcrypt, HMAC-SHA256, PGP |
| Логирование | logrus |
| Email | gomail |
| Парсинг XML | etree |

Приложение состоит из нескольких слоев:
- модели данных,
- репозитории для работы с базой данных,
- сервисы для реализации бизнес-логики,
- обработчики запросов.

Каждый из слоев выполняет свою роль, обеспечивая чистую архитектуру и легкость в сопровождении кода.

## Установка и запуск

## 1. Настройка переменных окружения

Создайте файл, в корне приложения `.env` :

```env
# PostgreSQL
DB_CONN_STRING=postgres://myuser:mypassword@localhost:5432/bankDB?sslmode=disable

# JWT (секретный ключ, минимум 32 символа)
JWT_SECRET=your-super-secret-key-min-32-characters

# SMTP (для Mail.ru)
SMTP_HOST=smtp.mail.ru
SMTP_PORT=465
SMTP_USER=your-email@mail.ru
SMTP_PASSWORD=пароль_приложения
SMTP_FROM=noreply@bank.ru

# Сервер
PORT=8080
```

## 2. Запуск PostgreSQL через Docker

```bash
docker compose up -d --build
```
Необходимо создать БД 'bankDB'

## 3. Установка зависимостей
```bash
go mod tidy
```

## 4. Запуск приложения
```bash
go run main.go
```
## 5. Проверка работы

Откройте браузер и перейдите по адресу:

Главная страница: http://localhost:8080/

# Структура проекта

```bash
bankAPI/
├── main.go                          # Точка входа, запуск сервера
├── go.mod                           # Зависимости Go
├── go.sum                           # Контрольные суммы зависимостей
├── .env                             # Переменные окружения
├── docker-compose.yml               # Запуск PostgreSQL
├── test_api.sh                      # Скрипт для тестирования API
├── init-scripts/
│   └── 001_init.sql                 # SQL схема базы данных
├── web/                             # Фронтенд
│   ├── index.html                   # Страница входа/регистрации
│   ├── dashboard.html               # Личный кабинет
│   ├── css/
│   │   └── style.css                # Стили
│   └── js/
│       └── app.js                   # Клиентская логика
└── internal/
    ├── models/
    │   └── models.go                # Все структуры данных
    ├── repository/
    │   ├── user_repo.go             # Работа с таблицей users
    │   ├── account_repo.go          # Работа с таблицей accounts
    │   ├── card_repo.go             # Работа с таблицей cards
    │   ├── transaction_repo.go      # Работа с таблицей transactions
    │   ├── credit_repo.go           # Работа с таблицей credits
    │   └── schedule_repo.go         # Работа с таблицей payment_schedules
    ├── service/
    │   ├── auth_service.go          # Регистрация, логин, JWT
    │   ├── account_service.go       # Управление счетами
    │   ├── card_service.go          # Выпуск и оплата карт
    │   ├── transfer_service.go      # Переводы между счетами
    │   ├── credit_service.go        # Кредиты и платежи
    │   ├── analytics_service.go     # Аналитика и прогнозы
    │   ├── email_service.go         # SMTP уведомления
    │   └── cbr_client.go            # Интеграция с ЦБ РФ
    ├── handler/
    │   ├── auth_handler.go          # /register, /login
    │   ├── account_handler.go       # /api/accounts
    │   ├── card_handler.go          # /api/cards
    │   ├── transfer_handler.go      # /api/transfer
    │   ├── credit_handler.go        # /api/credits
    │   ├── analytics_handler.go     # /api/analytics
    │   └── web_handler.go           # Статические страницы
    ├── middleware/
    │   └── auth.go                  # JWT проверка и логирование
    └── crypto/
        ├── pgp.go                   # PGP шифрование
        ├── hmac.go                  # HMAC подписи
        └── luhn.go                  # Генерация номеров карт
```

# API Эндпоинты

## Публичные (без JWT)

| Метод | Эндпоинт | Описание | Пример тела запроса |
|-----------|------------|------------|------------|
| GET | / | Главная страница | - |
| POST | /register | Регистрация | {"username":"john","email":"john@test.com","password":"123456"} |
| POST | /login | Вход | {"email":"john@test.com","password":"123456"} |


# Защищенные (требуют JWT)
## Заголовок: `Authorization: Bearer <token>`

| Метод | Эндпоинт | Описание | Пример тела запроса |
|-----------|------------|------------|------------|
| POST | /api/accounts | Создать счет | {"currency":"RUB"} |
| GET | /api/accounts |	Все счета |	- |
| POST | /api/accounts/{id}/deposit | Пополнить счет | {"amount":1000} |
| POST | /api/transfer | Перевод | {"from_account_id":"...","to_account_number":"...","amount":100,"description":"..."} |
| POST | /api/cards | Выпустить карту | {"account_id":"..."} |
| GET | /api/cards/{accountId} | Карты счета | - |
| POST | /api/cards/{id}/pay | Оплата картой | {"cvv":"123","amount":1000} |
| POST | /api/credits | Оформить кредит | {"account_id":"...","amount":50000,"term_months":12} |
| GET | /api/credits/{id}/schedule | График платежей | - |
| GET | /api/analytics | Аналитика | - |
| GET | /api/accounts/{id}/predict?days=30 | Прогноз баланса | - |

## Проверка через CURL
```bash
# 1. Регистрация
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username":"ivan","email":"ivan@test.com","password":"123456"}'

# 2. Вход (сохраните токен)
TOKEN=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"ivan@test.com","password":"123456"}' | jq -r '.token')

# 3. Создание счета
ACCOUNT_ID=$(curl -s -X POST http://localhost:8080/api/accounts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"currency":"RUB"}' | jq -r '.id')

# 4. Пополнение счета
curl -X POST http://localhost:8080/api/accounts/$ACCOUNT_ID/deposit \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"amount": 10000}'

# 5. Выпуск карты
curl -X POST http://localhost:8080/api/cards \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"account_id\": \"$ACCOUNT_ID\"}"

# 6. Перевод средств (нужен номер счета получателя)
curl -X POST http://localhost:8080/api/transfer \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"from_account_id\": \"$ACCOUNT_ID\",
    \"to_account_number\": \"40817XXXXXXXXXXXXX\",
    \"amount\": 500,
    \"description\": \"Тестовый перевод\"
  }"

# 7. Оформление кредита
curl -X POST http://localhost:8080/api/credits \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"account_id\": \"$ACCOUNT_ID\",
    \"amount\": 50000,
    \"term_months\": 12
  }"

# 8. Получение аналитики
curl -X GET http://localhost:8080/api/analytics \
  -H "Authorization: Bearer $TOKEN"

# 9. Прогноз баланса
curl -X GET "http://localhost:8080/api/accounts/$ACCOUNT_ID/predict?days=30" \
  -H "Authorization: Bearer $TOKEN"

# 10. Попытка доступа без токена (должен вернуть 401)
curl -X GET http://localhost:8080/api/accounts

# 11. Попытка доступа к чужому счету (должен вернуть 404)
curl -X GET http://localhost:8080/api/accounts/чужой_id \
  -H "Authorization: Bearer $TOKEN"
```

## Таким образом в учебном приложении реализованно:

### Слой моделей:

- Определение структур данных (соответствие таблицам БД)
- Сериализация/десериализация (теги JSON)
- Базовая валидация полей (email, username) — 1 балл.
- Проверка уникальности (email, username) — 1 балл.
- Полная валидация всех полей — 1 балл.

### Cлой репозиториев:

- Инкапсуляция SQL-запросов
- Параметризованные запросы
- Простейшая обработка ошибок БД
- Управление транзакциями
- Обработка сложных ошибок БД

### Cлой сервисов:

- Регистрация и аутентификация
- Создание счетов, пополнение баланса
- Переводы между счетами
- Генерация карт (алгоритм Луна)
- Кредиты: расчет аннуитетных платежей
- Интеграция с SMTP (уведомления)
- Интеграция с ЦБ РФ (SOAP)
- Шедулер для списания платежей
- Логирование через logrus

### Слой обработчиков:

- Валидация входных данных
- Формирование HTTP-ответов (JSON)
- Вызов методов сервисов
- Реализация всех эндпоинтов из ТЗ
- Проверка прав доступа к ресурсам


### Маршрутизации:

- Публичные эндпоинты (/register, /login)
- Защищенные эндпоинты (/accounts, /transfer и другие)


### Реализация Middleware:

- Проверка JWT-токенов
- Блокировка неавторизованных запросов
- Добавление ID пользователя в контекст


### Безопасность:

- Хеширование паролей (bcrypt)
- Шифрование данных карт (PGP + HMAC)
- Хеширование CVV (bcrypt)
- Проверка прав доступа к счетам

### База данных:

- Создание минимальных таблиц