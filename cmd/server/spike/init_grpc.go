package spike

import (
	email_grpc "github.com/manabie-com/backend/internal/spike/modules/email/controller/grpc"
	spb "github.com/manabie-com/backend/pkg/manabuf/spike/v1"

	"google.golang.org/grpc"
)

func initNewEmailServer(
	server grpc.ServiceRegistrar,
	emailModifierSvc *email_grpc.EmailModifierService,
) {
	spb.RegisterEmailModifierServiceServer(server, emailModifierSvc)
}
