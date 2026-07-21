package email

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"

	"github.com/DmitriiCherkasow/synergyconnect.git/internal/domain"

	"net/smtp"
)

// Config — конфигурация SMTP
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
	UseTLS   bool
}

// Service — сервис для отправки email
type Service struct {
	config Config
}

// NewService создает новый email-сервис
func NewService(config Config) *Service {
	return &Service{config: config}
}

// ReminderEmailData — данные для письма-напоминания
type ReminderEmailData struct {
	StickerTitle  string
	StickerContent string
	Priority      string
	DueDate       string
	BoardTitle    string
	BoardURL      string
	UserName      string
	DaysUntilDue  int
	IsOverdue     bool
}

// SendReminder отправляет письмо-напоминание
func (s *Service) SendReminder(
	to string,
	userName string,
	reminder *domain.Reminder,
	sticker *domain.Sticker,
	board *domain.Board,
) error {
	// Формируем данные для шаблона
	data := ReminderEmailData{
		StickerTitle:   sticker.Title,
		StickerContent: sticker.Content,
		Priority:       string(sticker.Priority),
		UserName:       userName,
		BoardTitle:     board.Title,
		BoardURL:       fmt.Sprintf("http://localhost:8080/boards/%s", board.ID.String()),
	}

	if sticker.DueDate != nil {
		data.DueDate = sticker.DueDate.Format("02 Jan 2006 15:04")
		data.DaysUntilDue = sticker.DaysUntilDue()
		data.IsOverdue = sticker.IsOverdue()
	}

	// Генерируем HTML-письмо
	body, err := s.generateReminderHTML(data)
	if err != nil {
		return err
	}

	// Отправляем
	return s.send(to, "🔔 SynergyConnect — Напоминание о задаче", body)
}

// send отправляет письмо через SMTP
func (s *Service) send(to, subject, body string) error {
	// Формируем заголовки
	from := s.config.From
	if s.config.FromName != "" {
		from = fmt.Sprintf("%s <%s>", s.config.FromName, s.config.From)
	}

	headers := map[string]string{
		"From":         from,
		"To":           to,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": `text/html; charset="UTF-8"`,
	}

	// Собираем письмо
	msg := ""
	for k, v := range headers {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	msg += "\r\n" + body

	// Настройка SMTP
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	// Отправка
	if s.config.UseTLS {
		return s.sendTLS(addr, auth, s.config.From, []string{to}, []byte(msg))
	}
	return smtp.SendMail(addr, auth, s.config.From, []string{to}, []byte(msg))
}

// sendTLS отправляет письмо через TLS
func (s *Service) sendTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		ServerName:         s.config.Host,
		InsecureSkipVerify: false,
	})
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return err
	}
	defer client.Quit()

	if err = client.Auth(auth); err != nil {
		return err
	}
	if err = client.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return nil
}

// generateReminderHTML генерирует HTML-письмо
func (s *Service) generateReminderHTML(data ReminderEmailData) (string, error) {
	// HTML-шаблон
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Arial, sans-serif;
            background-color: #f4f7fa;
            margin: 0;
            padding: 0;
        }
        .container {
            max-width: 600px;
            margin: 40px auto;
            background: #ffffff;
            border-radius: 12px;
            box-shadow: 0 4px 12px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        .header {
            background: #2c3e50;
            color: white;
            padding: 24px 32px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 24px;
            font-weight: 600;
        }
        .content {
            padding: 32px;
        }
        .greeting {
            font-size: 18px;
            color: #2c3e50;
            margin-bottom: 20px;
        }
        .sticker-card {
            background: #f8f9fa;
            border-left: 4px solid #3498db;
            padding: 16px 20px;
            border-radius: 4px;
            margin: 16px 0;
        }
        .sticker-title {
            font-size: 20px;
            font-weight: 600;
            color: #2c3e50;
            margin-bottom: 8px;
        }
        .sticker-content {
            color: #555;
            line-height: 1.6;
        }
        .meta {
            display: flex;
            flex-wrap: wrap;
            gap: 16px;
            margin: 16px 0;
            padding: 12px 0;
            border-top: 1px solid #e9ecef;
            border-bottom: 1px solid #e9ecef;
        }
        .meta-item {
            display: flex;
            align-items: center;
            gap: 6px;
            font-size: 14px;
            color: #666;
        }
        .meta-item .label {
            font-weight: 600;
            color: #2c3e50;
        }
        .priority-low { color: #2ecc71; }
        .priority-medium { color: #f39c12; }
        .priority-high { color: #e67e22; }
        .priority-urgent { color: #e74c3c; font-weight: 700; }
        .overdue { color: #e74c3c; font-weight: 700; }
        .btn {
            display: inline-block;
            background: #3498db;
            color: white;
            padding: 10px 24px;
            text-decoration: none;
            border-radius: 6px;
            margin-top: 16px;
        }
        .btn:hover {
            background: #2980b9;
        }
        .footer {
            text-align: center;
            padding: 16px;
            font-size: 12px;
            color: #999;
            border-top: 1px solid #e9ecef;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📌 SynergyConnect</h1>
        </div>
        <div class="content">
            <div class="greeting">Здравствуйте, <strong>{{.UserName}}</strong>!</div>
            <p style="color: #555;">Напоминаем о задаче на доске <strong>{{.BoardTitle}}</strong>:</p>

            <div class="sticker-card">
                <div class="sticker-title">{{.StickerTitle}}</div>
                <div class="sticker-content">{{.StickerContent}}</div>
            </div>

            <div class="meta">
                <span class="meta-item">
                    <span class="label">Приоритет:</span>
                    <span class="priority-{{.Priority}}">{{.Priority}}</span>
                </span>
                {{if .DueDate}}
                <span class="meta-item">
                    <span class="label">Срок:</span>
                    {{if .IsOverdue}}
                    <span class="overdue">🔴 ПРОСРОЧЕНО!</span>
                    {{else}}
                    <span>{{.DueDate}}</span>
                    <span style="color: #666; font-size: 12px;">(осталось {{.DaysUntilDue}} дн.)</span>
                    {{end}}
                </span>
                {{end}}
            </div>

            <a href="{{.BoardURL}}" class="btn">Перейти к доске</a>
        </div>
        <div class="footer">
            Это письмо отправлено автоматически из SynergyConnect.
            Вы можете управлять напоминаниями в своём профиле.
        </div>
    </div>
</body>
</html>
	`

	t, err := template.New("reminder").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}