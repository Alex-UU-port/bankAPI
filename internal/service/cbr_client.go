package service

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/beevik/etree"
	"github.com/sirupsen/logrus"
)

type CBRClient struct {
	client *http.Client
	url    string
	cache  map[string]float64
}

func NewCBRClient() *CBRClient {
	return &CBRClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		url:   "https://www.cbr.ru/DailyInfoWebServ/DailyInfo.asmx",
		cache: make(map[string]float64),
	}
}

// Формирование SOAP запроса
func (c *CBRClient) buildSOAPRequest() string {
	// Запрашиваем ставки за последние 30 дней
	fromDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	toDate := time.Now().Format("2006-01-02")

	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
        <soap12:Envelope xmlns:soap12="http://www.w3.org/2003/05/soap-envelope">
            <soap12:Body>
                <KeyRate xmlns="http://web.cbr.ru/">
                    <fromDate>%s</fromDate>
                    <ToDate>%s</ToDate>
                </KeyRate>
            </soap12:Body>
        </soap12:Envelope>`, fromDate, toDate)
}

// Отправка SOAP запроса
func (c *CBRClient) sendRequest(soapRequest string) ([]byte, error) {
	req, err := http.NewRequest(
		"POST",
		c.url,
		bytes.NewBuffer([]byte(soapRequest)),
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	// Установка заголовков для SOAP
	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")
	req.Header.Set("SOAPAction", "http://web.cbr.ru/KeyRate")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки запроса: %w", err)
	}
	defer resp.Body.Close()

	// Чтение ответа
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("сервер вернул статус %d: %s", resp.StatusCode, string(rawBody))
	}

	return rawBody, nil
}

// Парсинг XML ответа ЦБ РФ
func (c *CBRClient) parseXMLResponse(rawBody []byte) (float64, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(rawBody); err != nil {
		return 0, fmt.Errorf("ошибка парсинга XML: %w", err)
	}

	logrus.WithField("xml_preview", string(rawBody[:min(500, len(rawBody))])).Debug("Получен XML от ЦБ РФ")

	// Поиск элементов с ключевой ставкой
	// Путь: /Envelope/Body/KeyRateResponse/KeyRateResult/diffgram/KeyRate/KR
	krElements := doc.FindElements("//diffgram/KeyRate/KR")
	if len(krElements) == 0 {
		// Пробуем альтернативный путь
		krElements = doc.FindElements("//KeyRate/KR")
	}

	if len(krElements) == 0 {
		return 0, errors.New("данные по ключевой ставке не найдены в ответе ЦБ РФ")
	}

	// Берем последнюю актуальную ставку
	latestKR := krElements[len(krElements)-1]

	// Ищем тег Rate
	rateElement := latestKR.FindElement("./Rate")
	if rateElement == nil {
		return 0, errors.New("тег Rate отсутствует в ответе")
	}

	// Конвертация строки в число
	rateStr := rateElement.Text()
	var rate float64
	if _, err := fmt.Sscanf(rateStr, "%f", &rate); err != nil {
		return 0, fmt.Errorf("ошибка конвертации ставки '%s': %w", rateStr, err)
	}

	return rate, nil
}

// Получение ключевой ставки ЦБ РФ
func (c *CBRClient) GetKeyRate() (float64, error) {
	logrus.Info("Запрос ключевой ставки к ЦБ РФ")

	// Проверка кэша (актуально на сегодня)
	today := time.Now().Format("2006-01-02")
	if rate, ok := c.cache[today]; ok {
		logrus.WithField("rate", rate).Debug("Ключевая ставка получена из кэша")
		return rate, nil
	}

	// Формирование и отправка SOAP запроса
	soapRequest := c.buildSOAPRequest()
	logrus.WithField("soap_request", soapRequest).Debug("SOAP запрос к ЦБ РФ")

	rawBody, err := c.sendRequest(soapRequest)
	if err != nil {
		logrus.WithError(err).Warn("Ошибка при запросе к ЦБ РФ")
		return 0, err
	}

	// Парсинг ответа
	rate, err := c.parseXMLResponse(rawBody)
	if err != nil {
		logrus.WithError(err).Warn("Ошибка парсинга ответа ЦБ РФ")
		return 0, err
	}

	// Сохраняем в кэш
	c.cache[today] = rate

	logrus.WithFields(logrus.Fields{
		"rate":      rate,
		"cache_key": today,
	}).Info("Ключевая ставка ЦБ РФ успешно получена")

	return rate, nil
}

// Получение ставки с маржой банка
func (c *CBRClient) GetBankRate(marginPercent float64) (float64, error) {
	keyRate, err := c.GetKeyRate()
	if err != nil {
		logrus.WithError(err).Warn("Не удалось получить ключевую ставку, используется значение по умолчанию")
		return 16.0 + marginPercent, nil
	}

	bankRate := keyRate + marginPercent
	logrus.WithFields(logrus.Fields{
		"key_rate":  keyRate,
		"margin":    marginPercent,
		"bank_rate": bankRate,
	}).Info("Рассчитана процентная ставка банка")

	return bankRate, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
