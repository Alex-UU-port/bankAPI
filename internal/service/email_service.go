package service

import (
	"crypto/tls"
	"fmt"
	"os"
	"time"

	"github.com/go-mail/mail/v2"
	"github.com/sirupsen/logrus"
)

type EmailService struct {
	host     string
	port     int
	username string
	password string
	from     string
}

func NewEmailService() *EmailService {
	return &EmailService{
		host:     os.Getenv("SMTP_HOST"),
		port:     587,
		username: os.Getenv("SMTP_USER"),
		password: os.Getenv("SMTP_PASSWORD"),
		from:     os.Getenv("SMTP_FROM"),
	}
}

// Создание сообщения
func (s *EmailService) createMessage(to, subject, body string) *mail.Message {
	m := mail.NewMessage()

	from := s.from
	if from == "" {
		from = s.username
	}

	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html; charset=utf-8", body)

	return m
}

// Создание SMTP диалога
func (s *EmailService) createDialer() *mail.Dialer {
	d := mail.NewDialer(s.host, s.port, s.username, s.password)
	d.TLSConfig = &tls.Config{
		ServerName:         s.host,
		InsecureSkipVerify: false,
	}
	return d
}

// Отправка email
func (s *EmailService) send(d *mail.Dialer, m *mail.Message) error {
	if err := d.DialAndSend(m); err != nil {
		logrus.WithError(err).Error("Ошибка отправки email")
		return fmt.Errorf("ошибка отправки email: %w", err)
	}
	return nil
}

// Основной метод отправки email
func (s *EmailService) SendEmail(to, subject, body string) error {
	if s.host == "" || s.username == "" {
		logrus.Warn("SMTP не настроен, уведомление не отправлено")
		return nil
	}

	m := s.createMessage(to, subject, body)
	d := s.createDialer()

	if err := s.send(d, m); err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"to":      to,
		"subject": subject,
	}).Info("Email успешно отправлен")

	return nil
}

// Отправка приветственного письма
func (s *EmailService) SendWelcomeEmail(email, username string) {
	subject := "Добро пожаловать в Банк!"
	body := fmt.Sprintf(`
        <!DOCTYPE html>
        <html>
        <head>
            <meta charset="UTF-8">
            <style>
                body { font-family: Arial, sans-serif; line-height: 1.6; }
                .container { max-width: 600px; margin: 0 auto; padding: 20px; }
                .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); 
                          color: white; padding: 20px; text-align: center; }
                .content { padding: 20px; background: #f9f9f9; }
                .footer { text-align: center; padding: 10px; font-size: 12px; color: #666; }
            </style>
        </head>
        <body>
            <div class="container">
                <div class="header">
                    <h1>🏦 Добро пожаловать в наш банк!</h1>
                </div>
                <div class="content">
                    <h2>Здравствуйте, %s!</h2>
                    <p>Рады приветствовать вас в нашем банке. Ваш аккаунт успешно создан.</p>
                    <p>Для входа в систему используйте ваш email: <strong>%s</strong></p>
                    <p>Мы предлагаем:</p>
                    <ul>
                        <li>💳 Бесплатное обслуживание карт</li>
                        <li>💰 Выгодные кредитные ставки</li>
                        <li>📱 Удобный интернет-банкинг</li>
                    </ul>
                    <p>С уважением,<br>Команда банка</p>
                </div>
                <div class="footer">
                    <p>Это автоматическое сообщение, пожалуйста, не отвечайте на него.</p>
                </div>
            </div>
        </body>
        </html>
    `, username, email)

	go s.SendEmail(email, subject, body)
}

// Отправка уведомления о платеже
func (s *EmailService) SendPaymentNotification(email, username string, amount float64, cardMask string) {
	subject := "💳 Совершен платеж по карте"
	body := fmt.Sprintf(`
        <!DOCTYPE html>
        <html>
        <head>
            <meta charset="UTF-8">
            <style>
                body { font-family: Arial, sans-serif; }
                .container { max-width: 500px; margin: 0 auto; padding: 20px; }
                .amount { font-size: 24px; color: #28a745; font-weight: bold; }
            </style>
        </head>
        <body>
            <div class="container">
                <h2>Уважаемый(ая) %s!</h2>
                <p>С вашей карты <strong>%s</strong> был совершен платеж на сумму:</p>
                <p class="amount">%.2f RUB</p>
                <p>Если вы не совершали эту операцию, немедленно свяжитесь с поддержкой банка.</p>
                <hr>
                <small>Это автоматическое уведомление об операции по карте.</small>
            </div>
        </body>
        </html>
    `, username, cardMask, amount)

	go s.SendEmail(email, subject, body)
}

// Отправка уведомления о переводе
func (s *EmailService) SendTransferNotification(email, username string, amount float64, direction string) {
	subject := "🔄 Перевод средств"
	body := fmt.Sprintf(`
        <!DOCTYPE html>
        <html>
        <head>
            <meta charset="UTF-8">
            <style>
                body { font-family: Arial, sans-serif; }
                .amount { font-size: 24px; color: #007bff; font-weight: bold; }
            </style>
        </head>
        <body>
            <div class="container">
                <h2>Уважаемый(ая) %s!</h2>
                <p>Был совершен %s перевод на сумму:</p>
                <p class="amount">%.2f RUB</p>
                <p>Статус операции: <strong style="color: green;">Успешно завершена</strong></p>
                <hr>
                <small>Дата операции: %s</small>
            </div>
        </body>
        </html>
    `, username, direction, amount, time.Now().Format("02.01.2006 15:04:05"))

	go s.SendEmail(email, subject, body)
}
