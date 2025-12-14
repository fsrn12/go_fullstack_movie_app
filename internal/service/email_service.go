package service

import (
	"context"
	"errors"
	"fmt"
	"net/smtp"

	"multipass/config"
	"multipass/pkg/apperror"
	"multipass/pkg/common"
	"multipass/pkg/logging"
)

type Sender interface {
	SendEmail(ctx context.Context, to, subject, body string) error
}

// EmailSender defines the contract for sending account-related emails.
type EmailSender interface {
	SendVerificationEmail(toEmail, userName, tokenPlaintext string) error
	SendPasswordResetEmail(toEmail, userName, tokenPlaintext string) error
}

type EmailService struct {
	logger      logging.Logger
	EmailFrom   string
	SMTPHost    string
	SMTPPort    string
	SMTPUser    string
	SMTPPass    string
	FrontendURL string
}

func NewEmailService(cfg *config.Config, logger logging.Logger) EmailSender {
	if cfg.Email == nil {
		logger.Fatal("Email configuration is missing or nil in the main config", errors.New("Email_Config_Missing"))
		return nil
	}
	return &EmailService{
		logger:      logger,
		EmailFrom:   cfg.Email.FromAddress,
		SMTPHost:    cfg.Email.SMTPHost,
		SMTPPort:    cfg.Email.SMTPPort,
		SMTPUser:    cfg.Email.SMTPUser,
		SMTPPass:    cfg.Email.SMTPPass,
		FrontendURL: cfg.Email.FrontendURL,
	}
}

func (e *EmailService) send(emailTo, subject, body string, metaData common.Envelop) error {
	op := "email_service.send"
	metaData["op"] = op
	metaData["user_email"] = emailTo

	addr := fmt.Sprintf("%s:%s", e.SMTPHost, e.SMTPPort)
	auth := smtp.PlainAuth("", e.SMTPUser, e.SMTPPass, e.SMTPHost)

	msg := []byte(
		"From: " + e.EmailFrom + "\r\n" +
			"To: " + emailTo + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-version: 1.0;\r\n" +
			"Content-Type: text/html; charset=\"UTF-8\";\r\n" +
			"\r\n" +
			body + "\r\n",
	)

	err := smtp.SendMail(addr, auth, e.EmailFrom, []string{emailTo}, msg)
	if err != nil {
		metaData["smtp_error"] = err.Error()
		e.logger.Error("SMTP failed to send email", err, metaData)
		// Return a generic error to the service layer
		return fmt.Errorf("failed to send email: %w", err)
	}

	e.logger.Info("Email successfully sent", metaData)
	return nil
}

// SendVerificationEmail sends the link to the user to verify their account.
func (e *EmailService) SendVerificationEmail(toEmail, userName, tokenPlaintext string) error {
	op := "email_service.SendVerificationEmail"
	metaData := common.Envelop{"op": op, "to_email": toEmail}

	// Assumes your frontend route is /verify-email?token=<token>
	verificationLink := fmt.Sprintf("%s/verify-email?token=%s", e.FrontendURL, tokenPlaintext)

	subject := "Verify your Email Address for Your Movie App"

	// Using a simple HTML body for clarity. In production, use HTML templates.
	body := fmt.Sprintf(`
<html>
<body style="margin:0; padding:0; background:#0a0a0a; font-family:Arial, Helvetica, sans-serif;">

  <!-- Outer background -->
  <table width="100%%" cellpadding="0" cellspacing="0" border="0"
         style="background:#0a0a0a; padding:40px 0;">
    <tr>
      <td align="center">

        <!-- Card container -->
        <table width="600" cellpadding="0" cellspacing="0" border="0"
               style="background:#111; border-radius:16px; overflow:hidden;
                      box-shadow:0 0 40px rgba(0,0,0,0.6);">

          <!-- Header Section -->
          <tr>
            <td style="padding:40px 30px;
                       background:linear-gradient(135deg, rgba(255, 78, 80, 1) 0%%, rgba(249, 212, 35, 1) 100%%);
                       text-align:center;">
              <h1 style="margin:0; font-size:32px; font-weight:900;
                         color:white; text-shadow:0 2px 6px rgba(0, 0, 0, 0.4);">
                🎬 Welcome to The Movie App!
              </h1>
            </td>
          </tr>

          <!-- Body Section -->
          <tr>
            <td style="padding:40px 30px; color:#eee; font-size:17px; line-height:1.7;">

              <p style="margin-top:0; color:#f9d423; font-size:20px; font-weight:bold;">
                Hello %s,
              </p>

              <p style="margin:0 0 20px 0;">
                We're thrilled to have you join our amazing movie community!
                Your adventure begins now — but first, please verify your email address so we can finalize your account.
              </p>

              <!-- CTA Button -->
              <div style="text-align:center; margin:40px 0;">
                <a href="%s"
                   style="display:inline-block; padding:16px 28px; font-size:18px;
                          font-weight:bold; color:#111;
                          background:linear-gradient(135deg,#f9d423,#ff4e50);
                          text-decoration:none; border-radius:50px;
                          box-shadow:0 6px 20px rgba(255, 78, 80, 0.5),
                                     0 0 12px rgba(249,212,35,0.8);
                          transition:all .3s ease;">
                  Verify Email Address ✨
                </a>
              </div>

              <p style="margin:20px 0 10px;">
                ⏳ <strong>This link will expire in 24 hours.</strong>
              </p>

              <p style="margin:0 0 20px 0;">
                If you didn’t sign up for this account, you can safely ignore this email.
              </p>

              <p style="margin:30px 0 0 0;">With excitement,</p>
              <p style="margin:0; font-weight:bold; color:#f9d423;">The Movie App Team 🍿</p>

            </td>
          </tr>

          <!-- Footer Section -->
          <tr>
            <td style="padding:20px; background:#0d0d0d; text-align:center; color:#777; font-size:12px;">
              © 2025 The Movie App — All rights reserved.
            </td>
          </tr>

        </table>

      </td>
    </tr>
  </table>

</body>
</html>
`, userName, verificationLink)

	if err := e.send(toEmail, subject, body, metaData); err != nil {
		return apperror.ErrEmailSendFailed(err, e.logger, metaData)
	}
	return nil
}

// SendPasswordResetEmail sends the link to the user to reset their password.
func (e *EmailService) SendPasswordResetEmail(toEmail, userName, tokenPlaintext string) error {
	op := "email_service.SendPasswordResetEmail"
	metaData := common.Envelop{"op": op, "to_email": toEmail}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s", e.FrontendURL, tokenPlaintext)

	subject := "Password Reset Request for Your Movie App"

	// Using a simple HTML body for clarity.
	body := fmt.Sprintf(`
		<html>
		<body>
			<p>Hello %s,</p>
			<p>We received a request to reset the password for your account. Click the link below to set a new password:</p>
			<p><a href="%s">Reset Your Password</a></p>
			<p>This link will expire in 1 hour.</p>
			<p>If you did not request a password reset, please ignore this email. Your password will remain unchanged.</p>
			<p>Best regards,</p>
			<p>The Movie App Team</p>
		</body>
		</html>
	`, userName, resetLink)

	if err := e.send(toEmail, subject, body, metaData); err != nil {
		return apperror.ErrEmailSendFailed(err, e.logger, metaData)
	}
	return nil
}
