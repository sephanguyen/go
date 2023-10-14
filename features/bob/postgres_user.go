package bob

import (
	"context"
	"fmt"
	"strings"

	pb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"google.golang.org/grpc/metadata"
)

func (s *suite) aRequestToGetPostgresUserInfo(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Request = &pb.GetPostgresUserPermissionRequest{}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aValidPostgresUserInfoKey(ctx context.Context) (context.Context, error) {
	var md metadata.MD = make(map[string][]string)
	md["key"] = []string{"omrsBCQumcNYnLHmv094vV9RzZFHgv1oSkML1Ktya+RqHRvV4BU9kfW/3RbU+4tmIXwDCREtxABStN3mnMoh/gTTPwm069awdsVb1pIeYz+382JiWYRM7mrZWwfWFUi1ZnoHSfQ8jHcxzcc0rI5rQAhPQ0DasrUmdTrWHRtGXU6Pu65kBN4tH+091L2KxIYuiIphLt3hrVOcX4GsQfJ7gosxwJ/a/nFkUGBMmGbEQRj/+wk+eRqrXfObr1g61+20F9MEUakOQdXjkWxy1X4EPUHApr+skSbDBr6dolAoR0X6bPf9gNxc520AlvUQORUmXSP+btIQdt4joKHKKLoL0g=="}
	ctx = metadata.NewOutgoingContext(ctx, md)
	return ctx, nil
}

func (s *suite) anInvalidBase64FormatPostgresUserInfoKey(ctx context.Context) (context.Context, error) {
	var md metadata.MD = make(map[string][]string)
	md["key"] = []string{"invalid base64 format key"}
	ctx = metadata.NewOutgoingContext(ctx, md)
	return ctx, nil
}

func (s *suite) anInvalidRSAFormatPostgresUserInfoKey(ctx context.Context) (context.Context, error) {
	var md metadata.MD = make(map[string][]string)
	md["key"] = []string{"U29tZSB0aGluZyBpbnZhbGlk"}
	ctx = metadata.NewOutgoingContext(ctx, md)
	return ctx, nil
}

func (s *suite) anInvalidPostgresUserInfoKey(ctx context.Context) (context.Context, error) {
	var md metadata.MD = make(map[string][]string)
	md["key"] = []string{"Fdm6Rtn8jwuNtoeVht67PfZJ8t5hcRexkDlpdhcy9+tla3UIHjhJgE00wE78hGw9qEs96ziOkIuUBW3Noa/Y03tyZPvvRlKr9J2kzfuEkOjUh+JA5D4t529YGXEOS9StQU/Lun5n9aTBT3rSYxjZ1fxKfsYlb1IewEmqYiliPHUifQiLMD8r8o2hsYS355YyKrxxDMy+bK5pipQHn99lfxeHddsgbI8H6y8HJ3xeP7dAatz2/pWr4zLq6Nhs5FyJqi0E3fjg9wt4BKROFPZNHu6jMLi38LQUcUklfG/5wzFPV/ndEiqScvpNvO7slAqiUxkB4ds+TOdmxsX3wxze2Q=="}
	ctx = metadata.NewOutgoingContext(ctx, md)
	return ctx, nil
}

func (s *suite) aRequestToGetPostgresPrivilege(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Request = &pb.GetPostgresNamespaceRequest{}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) callGetPostgresUserInfoAPI(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Response, stepState.ResponseErr = pb.NewPostgresUserServiceClient(s.Conn).GetPostgresUserPermission(ctx, stepState.Request.(*pb.GetPostgresUserPermissionRequest))
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) postgresUserInfoDataMustContain(ctx context.Context, userName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	usersData := stepState.Response.(*pb.GetPostgresUserPermissionResponse).PostgresUsers
	for i := range usersData {
		if usersData[i].UserName == userName {
			return StepStateToContext(ctx, stepState), nil
		}
	}
	return StepStateToContext(ctx, stepState), fmt.Errorf("not found user: %s", userName)
}

func (s *suite) callGetPostgresPrivilegeAPI(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Response, stepState.ResponseErr = pb.NewPostgresNamespaceServiceClient(s.Conn).GetPostgresNamespace(ctx, stepState.Request.(*pb.GetPostgresNamespaceRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) postgresMustHavePrivilege(ctx context.Context, usename string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	datas := stepState.Response.(*pb.GetPostgresNamespaceResponse).PostgresNamespaces

	for i := range datas {
		data := datas[i]
		for j := range data.AccessPrivileges {
			accessPrivileges := strings.Split(data.AccessPrivileges[j], "=")
			if accessPrivileges[0] == usename {
				return StepStateToContext(ctx, stepState), nil
			}
		}
	}

	return StepStateToContext(ctx, stepState), fmt.Errorf("not found: %s", usename)
}
