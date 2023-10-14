package helper

import (
	"context"
	"fmt"
	"strconv"

	"github.com/manabie-com/backend/features/eibanam/communication/entity"
	"github.com/manabie-com/backend/features/eibanam/communication/util"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"
)

func (h *StatefulHelper) BobNotiReader() bpb.NotificationReaderServiceClient {
	return bpb.NewNotificationReaderServiceClient(h.CommunicationHelper.BobConn)
}

func (h *StatefulHelper) YasuoNotiModifier() ypb.NotificationModifierServiceClient {
	return ypb.NewNotificationModifierServiceClient(h.CommunicationHelper.YasuoConn)
}

func (h *StatefulHelper) BobNotiModifier() bpb.NotificationModifierServiceClient {
	return bpb.NewNotificationModifierServiceClient(h.CommunicationHelper.BobConn)
}

func (h *StatefulHelper) GetCtxToken(ctx context.Context, userID string) (context.Context, error) {
	tok, err := h.GetToken(ctx, userID)
	if err != nil {
		return ctx, err
	}
	return util.ContextWithToken(ctx, tok), nil
}

func (h *StatefulHelper) CreateSchoolAdminWithResourcePathCtx(ctx context.Context) (context.Context, error) {
	_, school, err := h.CommunicationHelper.CreateSchoolAdminAndLoginToCMS(ctx, AccountTypeSchoolAdmin)
	if err != nil {
		return ctx, err
	}
	schoolAdmin := school.Admins[0]
	h.userState.defaultAdmin = schoolAdmin
	h.userState.usersByID[schoolAdmin.ID] = schoolAdmin.ToUser()
	h.userState.school = school
	ctxWithRscPath := golibs.ResourcePathToCtx(ctx, strconv.Itoa(int(school.ID)))

	return h.DoubleCheckResourcePathInContext(ctxWithRscPath, strconv.Itoa(int(school.ID))), nil
}

func (h *StatefulHelper) GetSchoolAdmin() *entity.Admin {
	return h.userState.defaultAdmin
}

func (h *StatefulHelper) GetSchoolAdminToken(ctx context.Context) (string, error) {
	return h.GetToken(ctx, h.userState.defaultAdmin.ID)
}

func (h *StatefulHelper) GetSchoolAdminCtxToken(ctx context.Context) context.Context {
	adminID := h.userState.defaultAdmin.ID
	ctx, err := h.GetCtxToken(ctx, adminID)
	if err != nil {
		panic(fmt.Sprintf("h.GetCtxToken %s", err))
	}
	return ctx
}

func (h *StatefulHelper) GetToken(ctx context.Context, userID string) (string, error) {
	thisUsr, exist := h.userState.usersByID[userID]
	if !exist {
		return "", fmt.Errorf("forgot to assign state of user %s into the pool", userID)
	}
	if thisUsr.Token != "" {
		return h.userState.usersByID[userID].Token, nil
	}
	userID, usrGroup := thisUsr.ID, thisUsr.Group
	tok, err := h.CommunicationHelper.GenerateExchangeTokenCtx(ctx, userID, usrGroup)
	if err != nil {
		return "", fmt.Errorf("h.GenerateExchangeTokenCtx %w", err)
	}
	thisUsr.Token = tok
	return tok, nil
}

func (h *StatefulHelper) CreateStudent(ctx context.Context, numOfParent int) (*entity.User, []*entity.User, error) {
	stu, err := h.CommunicationHelper.CreateStudent(h.userState.defaultAdmin, 1, []string{h.userState.school.DefaultLocation}, numOfParent != 0, numOfParent)
	if err != nil {
		return nil, nil, err
	}
	h.userState.usersByID[stu.User.ID] = &stu.User
	parents := make([]*entity.User, 0, len(stu.Parents))
	if len(stu.Parents) > 0 {
		for _, usr := range stu.Parents {
			h.userState.usersByID[usr.ID] = usr
			parents = append(parents, usr)
		}
	}
	return &stu.User, parents, nil
}

func (h *StatefulHelper) CreateStudentWithTheSameParent(ctx context.Context, numOfParent int) (string, []string, error) {
	stu, err := h.CommunicationHelper.CreateStudent(h.userState.defaultAdmin, 1, []string{h.userState.school.DefaultLocation}, numOfParent != 0, numOfParent)
	if err != nil {
		return "", nil, err
	}
	h.userState.usersByID[stu.User.ID] = &stu.User
	parIDs := make([]string, 0, len(stu.Parents))
	if len(stu.Parents) > 0 {
		for _, usr := range stu.Parents {
			h.userState.usersByID[usr.ID] = usr
			parIDs = append(parIDs, usr.ID)
		}
	}
	return stu.User.ID, parIDs, nil
}

func (h *StatefulHelper) CreateStudentsWithSameParent(ctx context.Context, optP *CreateStudentsWithSameParentOpt) (students []*entity.User, parent *entity.User, err error) {
	tok, err := h.GetSchoolAdminToken(ctx)
	if err != nil {
		return nil, nil, err
	}
	studentsPb, parentPb, err := h.CommunicationHelper.CreateStudentsWithSameParent(ctx, tok, nil, optP)
	if err != nil {
		return nil, nil, err
	}
	for _, studentPb := range studentsPb {
		// studentIDs = append(studentIDs, student.UserProfile.UserId)
		stu := &entity.User{}
		stu.FromStudentPB(studentPb)
		h.userState.usersByID[stu.ID] = stu
		students = append(students, stu)
	}
	parent = &entity.User{}
	parent.FromParentPB(parentPb)
	h.userState.usersByID[parent.ID] = parent
	return students, parent, nil
}

func (h *StatefulHelper) DoubleCheckResourcePathInContext(ctx context.Context, resourcePath string) context.Context {
	claims := interceptors.JWTClaimsFromContext(ctx)
	// defaultResourcePath is 1 for BDD
	// if resourcePath = 1 other queries may fail to select correct data when testing
	if claims.Manabie.ResourcePath == "1" && claims.Manabie.ResourcePath != resourcePath {
		claims.Manabie.ResourcePath = resourcePath
	}
	ctx = interceptors.ContextWithJWTClaims(ctx, claims)
	return ctx
}
