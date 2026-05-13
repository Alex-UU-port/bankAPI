package crypto

import (
	"crypto/rand"
	"strconv"
)

// Генерация номера карты по алгоритму Луна
func GenerateCardNumber() string {
	// Префикс для Visa (4) + 15 случайных цифр
	prefix := "4"
	numbers := prefix + generateRandomNumbers(15)

	// Вычисление контрольной суммы
	checksum := calculateLuhnChecksum(numbers)

	return numbers + strconv.Itoa(checksum)
}

// Генерация случайных цифр
func generateRandomNumbers(length int) string {
	const digits = "0123456789"
	result := make([]byte, length)
	rand.Read(result)
	for i := 0; i < length; i++ {
		result[i] = digits[int(result[i])%10]
	}
	return string(result)
}

// Вычисление контрольной суммы Луна
func calculateLuhnChecksum(number string) int {
	sum := 0
	alternate := false

	for i := len(number) - 1; i >= 0; i-- {
		n := int(number[i] - '0')
		if alternate {
			n *= 2
			if n > 9 {
				n = n%10 + 1
			}
		}
		sum += n
		alternate = !alternate
	}

	return (10 - (sum % 10)) % 10
}

// Проверка валидности номера карты
func ValidateCardNumber(cardNumber string) bool {
	if len(cardNumber) != 16 {
		return false
	}

	checksum := calculateLuhnChecksum(cardNumber[:16])
	return checksum == 0
}

// Генерация CVV кода
func GenerateCVV() string {
	cvv := make([]byte, 3)
	rand.Read(cvv)
	for i := 0; i < 3; i++ {
		cvv[i] = '0' + (cvv[i] % 10)
	}
	return string(cvv)
}

// Генерация срока действия карты (5 лет)
func GenerateExpiry() string {
	// Просто возвращаем дату через 5 лет
	return "12/30" // Упрощенно: декабрь 2030
}
