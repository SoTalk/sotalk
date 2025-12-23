package email

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/smtp"
	"time"

	"github.com/yourusername/sotalk/pkg/logger"
	"go.uber.org/zap"
)

// SMTPConfig holds SMTP server configuration
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
}

// Client represents an email client
type Client struct {
	config SMTPConfig
}

// NewClient creates a new email client
func NewClient(config SMTPConfig) *Client {
	return &Client{
		config: config,
	}
}

// EmailData represents email data
type EmailData struct {
	To      []string
	Subject string
	Body    string
	IsHTML  bool
}

// SendEmail sends an email using SMTP
func (c *Client) SendEmail(data EmailData) error {
	// Build email headers
	from := fmt.Sprintf("%s <%s>", c.config.FromName, c.config.From)

	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("From: %s\r\n", from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", data.To[0]))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", data.Subject))
	msg.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	msg.WriteString("MIME-Version: 1.0\r\n")

	if data.IsHTML {
		msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	} else {
		msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	}

	msg.WriteString("\r\n")
	msg.WriteString(data.Body)

	// Setup authentication
	auth := smtp.PlainAuth("", c.config.Username, c.config.Password, c.config.Host)

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)

	// Setup TLS config
	tlsConfig := &tls.Config{
		ServerName:         c.config.Host,
		InsecureSkipVerify: false,
	}

	// Try to send with TLS
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		// Fallback to non-TLS (STARTTLS)
		logger.Warn("Failed to connect with TLS, trying STARTTLS",
			zap.String("host", c.config.Host),
			zap.Error(err),
		)

		// Send email using standard SMTP
		if err := smtp.SendMail(addr, auth, c.config.From, data.To, msg.Bytes()); err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}

		logger.Info("Email sent successfully via STARTTLS",
			zap.String("to", data.To[0]),
			zap.String("subject", data.Subject),
		)
		return nil
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, c.config.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Authenticate
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	// Set sender
	if err := client.Mail(c.config.From); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipients
	for _, to := range data.To {
		if err := client.Rcpt(to); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", to, err)
		}
	}

	// Send email body
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	_, err = w.Write(msg.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write email body: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	client.Quit()

	logger.Info("Email sent successfully via TLS",
		zap.String("to", data.To[0]),
		zap.String("subject", data.Subject),
	)

	return nil
}

// SendInvitation sends an invitation email
func (c *Client) SendInvitation(toEmail, inviterName, inviteLink string) error {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            text-align: center;
            border-radius: 10px 10px 0 0;
        }
        .content {
            background: #f9f9f9;
            padding: 30px;
            border-radius: 0 0 10px 10px;
        }
        .button {
            display: inline-block;
            padding: 12px 30px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            text-decoration: none;
            border-radius: 5px;
            margin: 20px 0;
            font-weight: bold;
        }
        .footer {
            text-align: center;
            margin-top: 30px;
            color: #666;
            font-size: 12px;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>🎉 You're Invited to SoTalk!</h1>
    </div>
    <div class="content">
        <p>Hello!</p>

        <p><strong>{{.InviterName}}</strong> has invited you to join <strong>SoTalk</strong>, a secure messaging platform powered by Solana blockchain.</p>

        <p>Experience:</p>
        <ul>
            <li>🔐 End-to-end encrypted messaging</li>
            <li>💰 Built-in wallet and instant payments</li>
            <li>🎯 No phone number required - just your Solana wallet</li>
            <li>⚡ Fast, secure, and decentralized</li>
        </ul>

        <p style="text-align: center;">
            <a href="{{.InviteLink}}" class="button">Accept Invitation</a>
        </p>

        <p style="color: #666; font-size: 14px;">
            Or copy and paste this link in your browser:<br>
            <code style="background: #e0e0e0; padding: 5px 10px; border-radius: 3px; display: inline-block; margin-top: 5px;">{{.InviteLink}}</code>
        </p>

        <p>Join us and start messaging securely on the blockchain!</p>

        <p>Best regards,<br>The SoTalk Team</p>
    </div>
    <div class="footer">
        <p>This invitation was sent by {{.InviterName}} via SoTalk.</p>
        <p>© 2024 SoTalk. All rights reserved.</p>
    </div>
</body>
</html>
`

	// Parse template
	t, err := template.New("invitation").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template
	var body bytes.Buffer
	data := struct {
		InviterName string
		InviteLink  string
	}{
		InviterName: inviterName,
		InviteLink:  inviteLink,
	}

	if err := t.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Send email
	return c.SendEmail(EmailData{
		To:      []string{toEmail},
		Subject: fmt.Sprintf("%s invited you to SoTalk!", inviterName),
		Body:    body.String(),
		IsHTML:  true,
	})
}

// SendWelcomeEmail sends a welcome email to new users
func (c *Client) SendWelcomeEmail(toEmail, username string) error {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            text-align: center;
            border-radius: 10px 10px 0 0;
        }
        .content {
            background: #f9f9f9;
            padding: 30px;
            border-radius: 0 0 10px 10px;
        }
        .feature {
            margin: 15px 0;
            padding: 15px;
            background: white;
            border-radius: 5px;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>Welcome to SoTalk! 👋</h1>
    </div>
    <div class="content">
        <p>Hi {{.Username}},</p>

        <p>Welcome to <strong>SoTalk</strong>! We're excited to have you on board.</p>

        <div class="feature">
            <h3>🔐 Secure Messaging</h3>
            <p>Your messages are end-to-end encrypted. Only you and your recipient can read them.</p>
        </div>

        <div class="feature">
            <h3>💰 Built-in Wallet</h3>
            <p>Send and receive SOL and tokens directly in chat. No need to leave the app!</p>
        </div>

        <div class="feature">
            <h3>⚡ Blockchain Powered</h3>
            <p>Built on Solana for lightning-fast transactions and unmatched security.</p>
        </div>

        <p>Ready to get started? Log in with your Solana wallet and start chatting!</p>

        <p>If you have any questions, feel free to reach out to our support team.</p>

        <p>Best regards,<br>The SoTalk Team</p>
    </div>
</body>
</html>
`

	// Parse template
	t, err := template.New("welcome").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template
	var body bytes.Buffer
	data := struct {
		Username string
	}{
		Username: username,
	}

	if err := t.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Send email
	return c.SendEmail(EmailData{
		To:      []string{toEmail},
		Subject: "Welcome to SoTalk - Secure Messaging on Solana!",
		Body:    body.String(),
		IsHTML:  true,
	})
}

// SendPasswordResetEmail sends a password reset email (for future 2FA/recovery)
func (c *Client) SendPasswordResetEmail(toEmail, resetLink string) error {
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .header {
            background: #dc3545;
            color: white;
            padding: 30px;
            text-align: center;
            border-radius: 10px 10px 0 0;
        }
        .content {
            background: #f9f9f9;
            padding: 30px;
            border-radius: 0 0 10px 10px;
        }
        .button {
            display: inline-block;
            padding: 12px 30px;
            background: #dc3545;
            color: white;
            text-decoration: none;
            border-radius: 5px;
            margin: 20px 0;
            font-weight: bold;
        }
        .warning {
            background: #fff3cd;
            border: 1px solid #ffc107;
            padding: 15px;
            border-radius: 5px;
            margin: 20px 0;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>🔐 Account Recovery</h1>
    </div>
    <div class="content">
        <p>Hello,</p>

        <p>We received a request to reset your account recovery settings.</p>

        <p style="text-align: center;">
            <a href="%s" class="button">Reset Account</a>
        </p>

        <div class="warning">
            <strong>⚠️ Security Notice:</strong><br>
            If you didn't request this, please ignore this email and ensure your wallet is secure.
        </div>

        <p style="color: #666; font-size: 14px;">
            This link will expire in 1 hour.<br>
            <code style="background: #e0e0e0; padding: 5px 10px; border-radius: 3px; display: inline-block; margin-top: 5px;">%s</code>
        </p>

        <p>Stay safe,<br>The SoTalk Team</p>
    </div>
</body>
</html>
`, resetLink, resetLink)

	return c.SendEmail(EmailData{
		To:      []string{toEmail},
		Subject: "SoTalk - Account Recovery Request",
		Body:    body,
		IsHTML:  true,
	})
}
