package service

import (
	"fmt"

	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
)

var (
	// Body Email template for password reset email
	// - %[1]s is the app name
	// - %[2]s is the user email
	// - %[3]s is the reset password link

	englishEmailTemplate = EmailTemplate{
		Subject: "Reset your password for %[1]s",
		Body: `
<p>Hello,</p>
<p>Follow this link to reset your %[1]s password for your %[2]s account.</p>
<p><a href='%[3]s'>%[3]s</a></p>
<p>If you didn’t ask to reset your password, you can ignore this email.</p>
<p>Thanks,</p>
<p>Your %[1]s team</p>
`,
	}
	japanEmailTemplate = EmailTemplate{
		Subject: "%[1]s のパスワードを再設定してください",
		Body: `
<p>お客様</p>
<p>%[1]s の %[2]s アカウントのパスワードをリセットするには、次のリンクをクリックしてください。</p>
<p><a href='%[3]s'>%[3]s</a></p>
<p>パスワードのリセットを依頼していない場合は、このメールを無視してください。</p>
<p>よろしくお願いいたします。</p>
<p>%[1]s チーム</p>
`,
	}
)

func GenerateResetPasswordEmail(email string, resetLink string, langCode string) EmailTemplate {
	resetPasswordEmail := englishEmailTemplate

	if langCode == constant.JapanLanguageCode {
		resetPasswordEmail = japanEmailTemplate
	}

	resetPasswordEmail.Subject = fmt.Sprintf(resetPasswordEmail.Subject, constant.AppName)
	resetPasswordEmail.Body = fmt.Sprintf(resetPasswordEmail.Body, constant.AppName, email, resetLink)

	return resetPasswordEmail
}
