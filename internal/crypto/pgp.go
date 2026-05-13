package crypto

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

// Упрощенная версия PGP для демонстрации
// В реальном проекте используйте полноценную библиотеку

var encryptionKey = []byte("01234567890123456789012345678901") // 32 байта для AES

// Шифрование данных (симметричное для демо)
func EncryptData(data string) (string, error) {
	// В реальном проекте здесь должно быть PGP шифрование
	// Для демонстрации используем простое XOR (НЕ для продакшена!)
	key := encryptionKey
	result := make([]byte, len(data))
	for i := 0; i < len(data); i++ {
		result[i] = data[i] ^ key[i%len(key)]
	}

	logrus.Debug("Данные зашифрованы")
	return string(result), nil
}

// Расшифровка данных
func DecryptData(encrypted string) (string, error) {
	key := encryptionKey
	data := []byte(encrypted)
	result := make([]byte, len(data))
	for i := 0; i < len(data); i++ {
		result[i] = data[i] ^ key[i%len(key)]
	}

	logrus.Debug("Данные расшифрованы")
	return string(result), nil
}

// Загрузка PGP ключа из файла
func LoadPGPKey(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("ошибка загрузки PGP ключа: %w", err)
	}
	return string(data), nil
}
