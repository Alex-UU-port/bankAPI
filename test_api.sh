#!/bin/bash

echo "=== 1. Регистрация пользователя ==="
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@test.com","password":"123456"}'

echo -e "\n=== 2. Вход в систему ==="
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"123456"}')

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token')
echo "Токен: $TOKEN"

echo -e "\n=== 3. Создание счета ==="
ACCOUNT_RESPONSE=$(curl -s -X POST http://localhost:8080/api/accounts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"currency":"RUB"}')

ACCOUNT_ID=$(echo $ACCOUNT_RESPONSE | jq -r '.id')
echo "ID счета: $ACCOUNT_ID"

echo -e "\n=== 4. Пополнение счета ==="
curl -s -X POST http://localhost:8080/api/accounts/$ACCOUNT_ID/deposit \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"amount": 10000}'
echo " - Пополнено 10000 RUB"

echo -e "\n=== 5. Выпуск карты ==="
CARD_RESPONSE=$(curl -s -X POST http://localhost:8080/api/cards \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"account_id\": \"$ACCOUNT_ID\"}")

# Проверяем, что ответ не пустой
if [ -n "$CARD_RESPONSE" ]; then
    echo "$CARD_RESPONSE" | jq '.' 2>/dev/null || echo "Ответ от сервера: $CARD_RESPONSE"
else
    echo "Ошибка: пустой ответ от сервера"
fi

echo -e "\n=== 6. Получение карт счета ==="
curl -s -X GET http://localhost:8080/api/cards/$ACCOUNT_ID \
  -H "Authorization: Bearer $TOKEN" | jq '.'

echo -e "\n=== 7. Регистрация второго пользователя (для теста перевода) ==="
curl -s -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username":"petr","email":"petr@test.com","password":"123456"}' | jq '.'

echo -e "\n=== 8. Вход второго пользователя ==="
LOGIN_RESPONSE2=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"petr@test.com","password":"123456"}')

TOKEN2=$(echo $LOGIN_RESPONSE2 | jq -r '.token')

echo -e "\n=== 9. Создание счета для второго пользователя ==="
ACCOUNT_RESPONSE2=$(curl -s -X POST http://localhost:8080/api/accounts \
  -H "Authorization: Bearer $TOKEN2" \
  -H "Content-Type: application/json" \
  -d '{"currency":"RUB"}')

ACCOUNT_NUMBER2=$(echo $ACCOUNT_RESPONSE2 | jq -r '.account_number')
echo "Номер счета получателя: $ACCOUNT_NUMBER2"

echo -e "\n=== 10. Перевод между счетами ==="
curl -s -X POST http://localhost:8080/api/transfer \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"from_account_id\": \"$ACCOUNT_ID\",
    \"to_account_number\": \"$ACCOUNT_NUMBER2\",
    \"amount\": 500,
    \"description\": \"Тестовый перевод\"
  }" | jq '.'

echo -e "\n=== 11. Проверка баланса после перевода ==="
curl -s -X GET http://localhost:8080/api/accounts \
  -H "Authorization: Bearer $TOKEN" | jq '.[0].balance'

echo -e "\n=== 12. Оформление кредита ==="
CREDIT_RESPONSE=$(curl -s -X POST http://localhost:8080/api/credits \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"account_id\": \"$ACCOUNT_ID\",
    \"amount\": 50000,
    \"term_months\": 12
  }")

CREDIT_ID=$(echo $CREDIT_RESPONSE | jq -r '.credit_id')
echo "Кредит оформлен, ID: $CREDIT_ID"
echo "$CREDIT_RESPONSE" | jq '{amount, interest_rate, monthly_payment}'

echo -e "\n=== 13. Получение графика платежей ==="
curl -s -X GET http://localhost:8080/api/credits/$CREDIT_ID/schedule \
  -H "Authorization: Bearer $TOKEN" | jq '.[0:3]'

echo -e "\n=== 14. Получение аналитики ==="
curl -s -X GET http://localhost:8080/api/analytics \
  -H "Authorization: Bearer $TOKEN" | jq '.'

echo -e "\n=== 15. Прогноз баланса на 30 дней ==="
curl -s -X GET "http://localhost:8080/api/accounts/$ACCOUNT_ID/predict?days=30" \
  -H "Authorization: Bearer $TOKEN" | jq '.[0:5]'

echo -e "\n=== 16. Проверка защиты - доступ к чужому счету ==="
echo "Пытаемся получить счет первого пользователя через токен второго:"
curl -s -X GET http://localhost:8080/api/accounts/$ACCOUNT_ID \
  -H "Authorization: Bearer $TOKEN2" | jq '.'

echo -e "\n=== 17. Проверка защиты - API без токена ==="
echo "Пытаемся получить счета без токена:"
curl -s -X GET http://localhost:8080/api/accounts \
  -H "Content-Type: application/json" | jq '.'

echo -e "\n=== ✅ Тестирование завершено ==="
