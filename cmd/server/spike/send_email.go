package spike

import (
	"context"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/spike/configurations"
	spb "github.com/manabie-com/backend/pkg/manabuf/spike/v1"

	"google.golang.org/grpc/metadata"
)

var (
	subject        string
	content        string
	recipients     string
	organizationID string
)

func init() {
	bootstrap.RegisterJob("send_email", RunSendEmail).
		StringVar(&subject, "subject", "", "email subject").
		StringVar(&content, "content", "", "email content").
		StringVar(&recipients, "recipients", "", "email recipients, separated by semicolons").
		StringVar(&organizationID, "org_id", "", "organization id")
}

func SendEmail(ctx context.Context, rsc *bootstrap.Resources, subject, content, recipients string) {
	ctx = metadata.NewOutgoingContext(ctx, metadata.New(make(map[string]string)))
	ctx = metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0")
	spikeConn := rsc.GRPCDial("spike")

	zapLogger := rsc.Logger()
	zapLogger.Sugar().Info("Start send email...")

	emailReq := spb.SendEmailRequest{
		Subject: subject,
		Content: &spb.SendEmailRequest_EmailContent{
			PlainText: content,
			HTML:      content,
		},
		Recipients:     strings.Split(recipients, ";"),
		OrganizationId: organizationID,
	}

	res, err := spb.NewEmailModifierServiceClient(spikeConn).SendEmail(ctx, &emailReq)
	if err != nil {
		zapLogger.Sugar().Error(err.Error())
	}

	zapLogger.Sugar().Infof("Your email_id is, it's queued: [%s]", res.EmailId)
}

func RunSendEmail(ctx context.Context, _ configurations.Config, rsc *bootstrap.Resources) error {
	SendEmail(ctx, rsc, subject, content, recipients)
	return nil
}
