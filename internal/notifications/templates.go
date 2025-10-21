package notifications

import (
	"bytes"
	"fmt"
	"html/template"
	"time"
)

// EmailTemplate represents an email template
type EmailTemplate struct {
	Name    string
	Subject string
	Body    string
	IsHTML  bool
}

// TemplateData contains data for email templates
type TemplateData struct {
	AppName     string
	BaseURL     string
	UserName    string
	UserEmail   string
	Token       string
	ExpiresAt   time.Time
	NoteTitle   string
	NoteID      string
	SharedBy    string
	QuotaUsed   int64
	QuotaTotal  int64
	QuotaPercent int
	BackupFile  string
	ErrorMsg    string
	CertDomain  string
	ExpiryDate  time.Time
	DaysLeft    int
	LoginIP     string
	LoginTime   time.Time
	DeviceInfo  string
	AdminAction string
	PerformedBy string
	ActionTime  time.Time
	AlertMsg    string
	BugReport   string
	ReportedBy  string
}

// Templates per CLAUDE.md Notification Events
var emailTemplates = map[string]*EmailTemplate{
	"welcome": {
		Name:    "welcome",
		Subject: "Welcome to {{.AppName}}!",
		Body: `<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: #282a36; color: #f8f8f2; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
		.content { background: #f4f4f4; padding: 20px; border-radius: 0 0 5px 5px; }
		.button { display: inline-block; padding: 10px 20px; background: #50fa7b; color: #282a36; text-decoration: none; border-radius: 5px; font-weight: bold; }
		.footer { margin-top: 20px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
	</style>
</head>
<body>
	<div class="header">
		<h1>Welcome to {{.AppName}}!</h1>
	</div>
	<div class="content">
		<p>Hi {{.UserName}},</p>
		<p>Thank you for registering with {{.AppName}}. Your account has been successfully created!</p>
		<p>You can now start creating and organizing your notes.</p>
		<p style="text-align: center; margin: 30px 0;">
			<a href="{{.BaseURL}}/users" class="button">Go to Dashboard</a>
		</p>
		<p>If you have any questions, feel free to contact us.</p>
		<div class="footer">
			<p>This email was sent to {{.UserEmail}}</p>
			<p>&copy; {{.AppName}} - Self-hosted notes</p>
		</div>
	</div>
</body>
</html>`,
		IsHTML: true,
	},

	"email_verification": {
		Name:    "email_verification",
		Subject: "Verify your email - {{.AppName}}",
		Body: `<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: #282a36; color: #f8f8f2; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
		.content { background: #f4f4f4; padding: 20px; border-radius: 0 0 5px 5px; }
		.button { display: inline-block; padding: 10px 20px; background: #50fa7b; color: #282a36; text-decoration: none; border-radius: 5px; font-weight: bold; }
		.code { background: #44475a; color: #f8f8f2; padding: 15px; margin: 20px 0; border-radius: 5px; font-size: 24px; text-align: center; letter-spacing: 5px; font-weight: bold; }
		.footer { margin-top: 20px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
	</style>
</head>
<body>
	<div class="header">
		<h1>Verify Your Email</h1>
	</div>
	<div class="content">
		<p>Hi {{.UserName}},</p>
		<p>Please verify your email address to complete your registration.</p>
		<p style="text-align: center; margin: 30px 0;">
			<a href="{{.BaseURL}}/verify?token={{.Token}}" class="button">Verify Email</a>
		</p>
		<p>Or use this verification code:</p>
		<div class="code">{{.Token}}</div>
		<p>This link expires at {{.ExpiresAt.Format "2006-01-02 15:04:05 MST"}}</p>
		<p>If you didn't register for {{.AppName}}, please ignore this email.</p>
		<div class="footer">
			<p>This email was sent to {{.UserEmail}}</p>
			<p>&copy; {{.AppName}} - Self-hosted notes</p>
		</div>
	</div>
</body>
</html>`,
		IsHTML: true,
	},

	"password_reset": {
		Name:    "password_reset",
		Subject: "Reset your password - {{.AppName}}",
		Body: `<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: #282a36; color: #f8f8f2; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
		.content { background: #f4f4f4; padding: 20px; border-radius: 0 0 5px 5px; }
		.button { display: inline-block; padding: 10px 20px; background: #ff5555; color: #f8f8f2; text-decoration: none; border-radius: 5px; font-weight: bold; }
		.footer { margin-top: 20px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
	</style>
</head>
<body>
	<div class="header">
		<h1>Password Reset Request</h1>
	</div>
	<div class="content">
		<p>Hi {{.UserName}},</p>
		<p>We received a request to reset your password. Click the button below to create a new password:</p>
		<p style="text-align: center; margin: 30px 0;">
			<a href="{{.BaseURL}}/reset-password?token={{.Token}}" class="button">Reset Password</a>
		</p>
		<p>This link expires at {{.ExpiresAt.Format "2006-01-02 15:04:05 MST"}}</p>
		<p>If you didn't request a password reset, please ignore this email. Your password will remain unchanged.</p>
		<div class="footer">
			<p>This email was sent to {{.UserEmail}}</p>
			<p>&copy; {{.AppName}} - Self-hosted notes</p>
		</div>
	</div>
</body>
</html>`,
		IsHTML: true,
	},

	"note_shared": {
		Name:    "note_shared",
		Subject: "{{.SharedBy}} shared a note with you - {{.AppName}}",
		Body: `<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: #282a36; color: #f8f8f2; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
		.content { background: #f4f4f4; padding: 20px; border-radius: 0 0 5px 5px; }
		.button { display: inline-block; padding: 10px 20px; background: #8be9fd; color: #282a36; text-decoration: none; border-radius: 5px; font-weight: bold; }
		.footer { margin-top: 20px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
	</style>
</head>
<body>
	<div class="header">
		<h1>Note Shared</h1>
	</div>
	<div class="content">
		<p>Hi {{.UserName}},</p>
		<p><strong>{{.SharedBy}}</strong> shared a note with you:</p>
		<p><strong>{{.NoteTitle}}</strong></p>
		<p style="text-align: center; margin: 30px 0;">
			<a href="{{.BaseURL}}/notes/shared/{{.NoteID}}" class="button">View Note</a>
		</p>
		<div class="footer">
			<p>This email was sent to {{.UserEmail}}</p>
			<p>&copy; {{.AppName}} - Self-hosted notes</p>
		</div>
	</div>
</body>
</html>`,
		IsHTML: true,
	},

	"quota_warning": {
		Name:    "quota_warning",
		Subject: "Storage quota warning - {{.AppName}}",
		Body: `<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: #282a36; color: #f8f8f2; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
		.content { background: #f4f4f4; padding: 20px; border-radius: 0 0 5px 5px; }
		.warning { background: #ffb86c; color: #282a36; padding: 15px; border-radius: 5px; margin: 20px 0; }
		.button { display: inline-block; padding: 10px 20px; background: #50fa7b; color: #282a36; text-decoration: none; border-radius: 5px; font-weight: bold; }
		.footer { margin-top: 20px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
	</style>
</head>
<body>
	<div class="header">
		<h1>Storage Quota Warning</h1>
	</div>
	<div class="content">
		<p>Hi {{.UserName}},</p>
		<div class="warning">
			<p><strong>You've used {{.QuotaPercent}}% of your storage quota</strong></p>
			<p>Used: {{.QuotaUsed}} MB / {{.QuotaTotal}} MB</p>
		</div>
		<p>Consider cleaning up old notes or archiving content to free up space.</p>
		<p style="text-align: center; margin: 30px 0;">
			<a href="{{.BaseURL}}/users/trash" class="button">Manage Storage</a>
		</p>
		<div class="footer">
			<p>This email was sent to {{.UserEmail}}</p>
			<p>&copy; {{.AppName}} - Self-hosted notes</p>
		</div>
	</div>
</body>
</html>`,
		IsHTML: true,
	},

	"backup_complete": {
		Name:    "backup_complete",
		Subject: "Backup completed - {{.AppName}}",
		Body: `<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: #282a36; color: #f8f8f2; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
		.content { background: #f4f4f4; padding: 20px; border-radius: 0 0 5px 5px; }
		.success { background: #50fa7b; color: #282a36; padding: 15px; border-radius: 5px; margin: 20px 0; }
		.footer { margin-top: 20px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
	</style>
</head>
<body>
	<div class="header">
		<h1>Backup Completed</h1>
	</div>
	<div class="content">
		<p>Hi {{.UserName}},</p>
		<div class="success">
			<p><strong>Database backup completed successfully</strong></p>
			<p>Backup file: {{.BackupFile}}</p>
		</div>
		<p>Your data has been securely backed up.</p>
		<div class="footer">
			<p>This email was sent to {{.UserEmail}}</p>
			<p>&copy; {{.AppName}} - Self-hosted notes</p>
		</div>
	</div>
</body>
</html>`,
		IsHTML: true,
	},

	"backup_failed": {
		Name:    "backup_failed",
		Subject: "Backup failed - {{.AppName}}",
		Body: `<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: #282a36; color: #f8f8f2; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
		.content { background: #f4f4f4; padding: 20px; border-radius: 0 0 5px 5px; }
		.error { background: #ff5555; color: #f8f8f2; padding: 15px; border-radius: 5px; margin: 20px 0; }
		.button { display: inline-block; padding: 10px 20px; background: #ff5555; color: #f8f8f2; text-decoration: none; border-radius: 5px; font-weight: bold; }
		.footer { margin-top: 20px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
	</style>
</head>
<body>
	<div class="header">
		<h1>Backup Failed</h1>
	</div>
	<div class="content">
		<p>Hi {{.UserName}},</p>
		<div class="error">
			<p><strong>Database backup failed</strong></p>
			<p>Error: {{.ErrorMsg}}</p>
		</div>
		<p>Please check your server logs and ensure adequate disk space.</p>
		<p style="text-align: center; margin: 30px 0;">
			<a href="{{.BaseURL}}/admin" class="button">View Admin Panel</a>
		</p>
		<div class="footer">
			<p>This email was sent to {{.UserEmail}}</p>
			<p>&copy; {{.AppName}} - Self-hosted notes</p>
		</div>
	</div>
</body>
</html>`,
		IsHTML: true,
	},

	"cert_expiry": {
		Name:    "cert_expiry",
		Subject: "Certificate expiring soon - {{.AppName}}",
		Body: `<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: #282a36; color: #f8f8f2; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
		.content { background: #f4f4f4; padding: 20px; border-radius: 0 0 5px 5px; }
		.warning { background: #ffb86c; color: #282a36; padding: 15px; border-radius: 5px; margin: 20px 0; }
		.button { display: inline-block; padding: 10px 20px; background: #50fa7b; color: #282a36; text-decoration: none; border-radius: 5px; font-weight: bold; }
		.footer { margin-top: 20px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
	</style>
</head>
<body>
	<div class="header">
		<h1>Certificate Expiry Warning</h1>
	</div>
	<div class="content">
		<p>Hi {{.UserName}},</p>
		<div class="warning">
			<p><strong>SSL/TLS certificate expiring soon</strong></p>
			<p>Domain: {{.CertDomain}}</p>
			<p>Expires: {{.ExpiryDate.Format "2006-01-02 15:04:05 MST"}}</p>
			<p>Days left: {{.DaysLeft}}</p>
		</div>
		<p>Please renew your certificate to avoid service interruption.</p>
		<p style="text-align: center; margin: 30px 0;">
			<a href="{{.BaseURL}}/admin/server" class="button">Manage Certificates</a>
		</p>
		<div class="footer">
			<p>This email was sent to {{.UserEmail}}</p>
			<p>&copy; {{.AppName}} - Self-hosted notes</p>
		</div>
	</div>
</body>
</html>`,
		IsHTML: true,
	},

	"new_device_login": {
		Name:    "new_device_login",
		Subject: "New device login detected - {{.AppName}}",
		Body: `<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: #282a36; color: #f8f8f2; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
		.content { background: #f4f4f4; padding: 20px; border-radius: 0 0 5px 5px; }
		.info { background: #8be9fd; color: #282a36; padding: 15px; border-radius: 5px; margin: 20px 0; }
		.button { display: inline-block; padding: 10px 20px; background: #ff5555; color: #f8f8f2; text-decoration: none; border-radius: 5px; font-weight: bold; }
		.footer { margin-top: 20px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
	</style>
</head>
<body>
	<div class="header">
		<h1>New Device Login</h1>
	</div>
	<div class="content">
		<p>Hi {{.UserName}},</p>
		<p>A new device logged into your account:</p>
		<div class="info">
			<p><strong>Time:</strong> {{.LoginTime.Format "2006-01-02 15:04:05 MST"}}</p>
			<p><strong>IP Address:</strong> {{.LoginIP}}</p>
			<p><strong>Device:</strong> {{.DeviceInfo}}</p>
		</div>
		<p>If this was you, you can safely ignore this email.</p>
		<p>If you don't recognize this activity, please secure your account immediately:</p>
		<p style="text-align: center; margin: 30px 0;">
			<a href="{{.BaseURL}}/users/settings" class="button">Secure My Account</a>
		</p>
		<div class="footer">
			<p>This email was sent to {{.UserEmail}}</p>
			<p>&copy; {{.AppName}} - Self-hosted notes</p>
		</div>
	</div>
</body>
</html>`,
		IsHTML: true,
	},

	"admin_action": {
		Name:    "admin_action",
		Subject: "Admin action notification - {{.AppName}}",
		Body: `<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: #282a36; color: #f8f8f2; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
		.content { background: #f4f4f4; padding: 20px; border-radius: 0 0 5px 5px; }
		.audit { background: #bd93f9; color: #f8f8f2; padding: 15px; border-radius: 5px; margin: 20px 0; }
		.button { display: inline-block; padding: 10px 20px; background: #50fa7b; color: #282a36; text-decoration: none; border-radius: 5px; font-weight: bold; }
		.footer { margin-top: 20px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
	</style>
</head>
<body>
	<div class="header">
		<h1>Admin Action Notification</h1>
	</div>
	<div class="content">
		<p>Hi {{.UserName}},</p>
		<p>An administrative action was performed on your account:</p>
		<div class="audit">
			<p><strong>Action:</strong> {{.AdminAction}}</p>
			<p><strong>Performed by:</strong> {{.PerformedBy}}</p>
			<p><strong>Time:</strong> {{.ActionTime.Format "2006-01-02 15:04:05 MST"}}</p>
		</div>
		<p>If you have questions about this action, please contact your administrator.</p>
		<p style="text-align: center; margin: 30px 0;">
			<a href="{{.BaseURL}}/users" class="button">View Dashboard</a>
		</p>
		<div class="footer">
			<p>This email was sent to {{.UserEmail}}</p>
			<p>&copy; {{.AppName}} - Self-hosted notes</p>
		</div>
	</div>
</body>
</html>`,
		IsHTML: true,
	},

	"emergency_alert": {
		Name:    "emergency_alert",
		Subject: "EMERGENCY ALERT - {{.AppName}}",
		Body: `<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: #ff5555; color: #f8f8f2; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
		.content { background: #f4f4f4; padding: 20px; border-radius: 0 0 5px 5px; }
		.emergency { background: #ff5555; color: #f8f8f2; padding: 15px; border-radius: 5px; margin: 20px 0; font-weight: bold; }
		.button { display: inline-block; padding: 10px 20px; background: #ff5555; color: #f8f8f2; text-decoration: none; border-radius: 5px; font-weight: bold; }
		.footer { margin-top: 20px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
	</style>
</head>
<body>
	<div class="header">
		<h1>⚠️ EMERGENCY ALERT ⚠️</h1>
	</div>
	<div class="content">
		<p>Hi {{.UserName}},</p>
		<div class="emergency">
			<p>{{.AlertMsg}}</p>
		</div>
		<p>Immediate action may be required. Please check your server immediately.</p>
		<p style="text-align: center; margin: 30px 0;">
			<a href="{{.BaseURL}}/admin" class="button">Admin Panel</a>
		</p>
		<div class="footer">
			<p>This email was sent to {{.UserEmail}}</p>
			<p>&copy; {{.AppName}} - Self-hosted notes</p>
		</div>
	</div>
</body>
</html>`,
		IsHTML: true,
	},

	"bug_report": {
		Name:    "bug_report",
		Subject: "Bug report received - {{.AppName}}",
		Body: `<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background: #282a36; color: #f8f8f2; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
		.content { background: #f4f4f4; padding: 20px; border-radius: 0 0 5px 5px; }
		.bug { background: #f1fa8c; color: #282a36; padding: 15px; border-radius: 5px; margin: 20px 0; }
		.button { display: inline-block; padding: 10px 20px; background: #50fa7b; color: #282a36; text-decoration: none; border-radius: 5px; font-weight: bold; }
		.footer { margin-top: 20px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #666; }
	</style>
</head>
<body>
	<div class="header">
		<h1>Bug Report</h1>
	</div>
	<div class="content">
		<p>Hi {{.UserName}},</p>
		<p>A bug report was submitted:</p>
		<div class="bug">
			<p><strong>Reported by:</strong> {{.ReportedBy}}</p>
			<p><strong>Description:</strong></p>
			<p>{{.BugReport}}</p>
		</div>
		<p style="text-align: center; margin: 30px 0;">
			<a href="{{.BaseURL}}/admin" class="button">View Admin Panel</a>
		</p>
		<div class="footer">
			<p>This email was sent to {{.UserEmail}}</p>
			<p>&copy; {{.AppName}} - Self-hosted notes</p>
		</div>
	</div>
</body>
</html>`,
		IsHTML: true,
	},
}

// RenderTemplate renders an email template with provided data
func RenderTemplate(templateName string, data *TemplateData) (*EmailMessage, error) {
	tmpl, exists := emailTemplates[templateName]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", templateName)
	}

	// Render subject
	subjectTmpl, err := template.New("subject").Parse(tmpl.Subject)
	if err != nil {
		return nil, fmt.Errorf("failed to parse subject template: %w", err)
	}
	var subjectBuf bytes.Buffer
	if err := subjectTmpl.Execute(&subjectBuf, data); err != nil {
		return nil, fmt.Errorf("failed to execute subject template: %w", err)
	}

	// Render body
	bodyTmpl, err := template.New("body").Parse(tmpl.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse body template: %w", err)
	}
	var bodyBuf bytes.Buffer
	if err := bodyTmpl.Execute(&bodyBuf, data); err != nil {
		return nil, fmt.Errorf("failed to execute body template: %w", err)
	}

	return &EmailMessage{
		To:      []string{data.UserEmail},
		Subject: subjectBuf.String(),
		Body:    bodyBuf.String(),
		IsHTML:  tmpl.IsHTML,
	}, nil
}

// GetTemplateNames returns all available template names
func GetTemplateNames() []string {
	names := make([]string, 0, len(emailTemplates))
	for name := range emailTemplates {
		names = append(names, name)
	}
	return names
}
