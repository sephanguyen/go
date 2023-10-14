package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateResetPasswordEmail(t *testing.T) {
	type args struct {
		email     string
		resetLink string
		langCode  string
	}
	tests := []struct {
		name string
		args args
		want EmailTemplate
	}{
		{
			name: "with default english language code",
			args: args{
				email:     "email",
				resetLink: "resetLink",
				langCode:  "en",
			},
			want: EmailTemplate{
				Subject: "Reset your password for Manabie",
				Body: `
<p>Hello,</p>
<p>Follow this link to reset your Manabie password for your email account.</p>
<p><a href='resetLink'>resetLink</a></p>
<p>If you didn’t ask to reset your password, you can ignore this email.</p>
<p>Thanks,</p>
<p>Your Manabie team</p>
`,
			},
		},
		{
			name: "with japan language code",
			args: args{
				email:     "email",
				resetLink: "resetLink",
				langCode:  "ja",
			},
			want: EmailTemplate{
				Subject: "Manabie のパスワードを再設定してください",
				Body: `
<p>お客様</p>
<p>Manabie の email アカウントのパスワードをリセットするには、次のリンクをクリックしてください。</p>
<p><a href='resetLink'>resetLink</a></p>
<p>パスワードのリセットを依頼していない場合は、このメールを無視してください。</p>
<p>よろしくお願いいたします。</p>
<p>Manabie チーム</p>
`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, GenerateResetPasswordEmail(tt.args.email, tt.args.resetLink, tt.args.langCode), "GenerateResetPasswordEmail(%v, %v, %v)", tt.args.email, tt.args.resetLink, tt.args.langCode)
		})
	}
}
