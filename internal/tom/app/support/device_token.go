package support

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	entities "github.com/manabie-com/backend/internal/tom/domain/core"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type DeviceTokenModifier struct {
	DB                  database.Ext
	JSM                 nats.JetStreamManagement
	Logger              *zap.Logger
	UserDeviceTokenRepo interface {
		Upsert(context.Context, database.QueryExecer, *entities.UserDeviceToken) error
	}
	ConversationRepo interface {
		SetName(ctx context.Context, db database.QueryExecer, cIDs pgtype.TextArray, name pgtype.Text) error
	}
	ConversationLocationRepo interface {
		RemoveLocationsForConversation(ctx context.Context, db database.QueryExecer, conversationID string, locations []string) error
		FindByConversationIDs(ctx context.Context, db database.QueryExecer, conversationIDs pgtype.TextArray) (map[string][]entities.ConversationLocation, error)
		BulkUpsert(ctx context.Context, db database.QueryExecer, locations []entities.ConversationLocation) error
	}
	ConversationStudentRepo interface {
		FindByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray, conversationType pgtype.Text) ([]string, error)
	}
	ConversationMemberRepo interface {
		FindByConversationIDsAndRoles(ctx context.Context, db database.QueryExecer, conversationIDs pgtype.TextArray, roles pgtype.TextArray) (map[string][]string, error)
		SetStatusByConversationAndUserIDs(ctx context.Context, db database.QueryExecer, conversationIDs pgtype.TextArray, userIDs pgtype.TextArray, status pgtype.Text) error
	}

	GrantedPermissionRepo interface {
		FindByUserIDAndPermissionName(ctx context.Context, db database.QueryExecer, userIDs pgtype.TextArray, permissionName pgtype.Text) (map[string][]*entities.GrantedPermission, error)
	}
}

func NewDeviceTokenModifier(logger *zap.Logger, db database.Ext) *DeviceTokenModifier {
	return &DeviceTokenModifier{
		DB:     db,
		Logger: logger,
	}
}

func areTheSameLocations(newLocations []string, oldLocationEnts []entities.ConversationLocation) (equal bool, removedLocations []string) {
	old := make([]string, 0, len(oldLocationEnts))
	for _, item := range oldLocationEnts {
		old = append(old, item.LocationID.String)
	}
	removedLocations = stringutil.SliceElementsDiff(old, newLocations)
	equal = len(newLocations) == len(oldLocationEnts) && len(removedLocations) == 0

	return
}

func (u *DeviceTokenModifier) HandleEvtUserInfo(ctx context.Context, req *upb.EvtUserInfo, updateLocation bool) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	newLocations := req.GetLocationIds()

	user, err := userInfoToEntity(req)
	if err != nil {
		u.Logger.Error("UserService.UserInfoToEntity: invalid args", zap.Error(err))
		return false, err
	}

	var conversationType pgtype.Text
	_ = conversationType.Set(nil)
	conversationIDs, err := u.ConversationStudentRepo.FindByStudentIDs(ctx, u.DB, database.TextArray([]string{user.UserID.String}), conversationType)
	if err != nil {
		u.Logger.Error("ConversationStudentRepo.FindByStudentIDs", zap.Error(err))
		return true, err
	}

	if updateLocation {
		staffConvMembersMap, err := u.ConversationMemberRepo.FindByConversationIDsAndRoles(ctx, u.DB, database.TextArray(conversationIDs), database.TextArray(constant.ConversationStaffRoles))
		if err != nil {
			u.Logger.Error("ConversationMemberRepo.FindByConversationIDsAndRoles", zap.Error(err))
			return true, err
		}

		staffConvMemberIDs := extractAllUniqueConvMemberIDs(staffConvMembersMap)

		convLocationMap, err := u.ConversationLocationRepo.FindByConversationIDs(ctx, u.DB, database.TextArray(conversationIDs))
		if err != nil {
			u.Logger.Error("ConversationLocationRepo.FindByConversationIDs", zap.Error(err))
			return true, err
		}

		var grantedPermissionsMap map[string][]*entities.GrantedPermission
		if len(staffConvMemberIDs) != 0 {
			grantedPermissionsMap, err = u.GrantedPermissionRepo.FindByUserIDAndPermissionName(ctx, u.DB, database.TextArray(staffConvMemberIDs), database.Text("master.location.read"))
			if err != nil {
				u.Logger.Error("GrantedPermissionRepo.FindByUserIDAndPermissionName", zap.Error(err))
				return true, err
			}
		}

		grantedLocationsMap := extractGrantedLocationIDsMap(grantedPermissionsMap)

		err = database.ExecInTx(ctx, u.DB, func(ctx context.Context, tx pgx.Tx) error {
			for _, convID := range conversationIDs {
				locEntities := convLocationMap[convID]
				if len(locEntities) == 0 {
					locEntities = []entities.ConversationLocation{}
				}
				if equal, removedLocations := areTheSameLocations(newLocations, locEntities); !equal {
					if len(newLocations) > 0 {
						newLocationEnts, err := toConversationLocations(convID, newLocations)
						if err != nil {
							return fmt.Errorf("toConversationLocations: %w", err)
						}
						err = u.ConversationLocationRepo.BulkUpsert(ctx, tx, newLocationEnts)
						if err != nil {
							return fmt.Errorf("ConversationLocationRepo.BulkUpsert: %w", err)
						}
					}

					if len(removedLocations) > 0 {
						err = u.ConversationLocationRepo.RemoveLocationsForConversation(ctx, tx, convID, removedLocations)
						if err != nil {
							return fmt.Errorf("ConversationLocationRepo.RemoveLocationsForConversation: %w", err)
						}
					}
				}
			}

			if len(newLocations) == 0 {
				return nil
			}

			inactivatingStaffIDs := getInactivatingConversationMembers(staffConvMemberIDs, grantedLocationsMap, newLocations)
			if len(inactivatingStaffIDs) > 0 {
				err := u.ConversationMemberRepo.SetStatusByConversationAndUserIDs(ctx, tx, database.TextArray(conversationIDs), database.TextArray(inactivatingStaffIDs), database.Text(entities.ConversationStatusInActive))
				if err != nil {
					return fmt.Errorf("ConversationMemberRepo.SetStatusByConversationAndUserIDs: %w", err)
				}
			}

			return nil
		})
		if err != nil {
			return true, fmt.Errorf("ExecInTx: %w", err)
		}
	}

	if err := u.UserDeviceTokenRepo.Upsert(ctx, u.DB, user); err != nil {
		return true, fmt.Errorf("u.UserDeviceTokenRepo: %s", err)
	}
	if len(conversationIDs) != 0 {
		if err := u.ConversationRepo.SetName(ctx, u.DB, database.TextArray(conversationIDs), database.Text(user.UserName.String)); err != nil {
			return true, err
		}
	}

	for _, convID := range conversationIDs {
		event := &tpb.ConversationInternal{
			TriggeredAt: timestamppb.Now(),
			Message: &tpb.ConversationInternal_ConversationUpdated_{
				ConversationUpdated: &tpb.ConversationInternal_ConversationUpdated{
					ConversationId: convID,
				},
			},
		}

		bs, err := proto.Marshal(event)
		if err != nil {
			u.Logger.Warn("proto.Marshal", zap.Error(err))
			return true, err
		}
		_, err = u.JSM.PublishContext(ctx, constants.SubjectChatUpdated, bs)
		if err != nil {
			u.Logger.Warn("c.JSM.PublishContext", zap.Error(err))
			return true, err
		}
	}

	return true, nil
}

func getInactivatingConversationMembers(memberIDs []string, grantedLocationsMap map[string][]string, conversationLocations []string) []string {
	var inactivatingTeacherIDs []string
	for _, userID := range memberIDs {
		grantedLocations, existed := grantedLocationsMap[userID]
		// deactivate staff in conversation if there is no assigned user groups
		// or granted locations and conversation locations are not matched

		if !existed {
			continue
		}

		matchedLocations := sliceutils.Intersect(conversationLocations, grantedLocations)

		if len(matchedLocations) == 0 {
			inactivatingTeacherIDs = append(inactivatingTeacherIDs, userID)
		}
	}
	return inactivatingTeacherIDs
}

func extractAllUniqueConvMemberIDs(staffConvMembersMap map[string][]string) []string {
	var staffConvMemberIDs []string
	for _, userIDs := range staffConvMembersMap {
		staffConvMemberIDs = append(staffConvMemberIDs, userIDs...)
	}

	return golibs.GetUniqueElementStringArray(staffConvMemberIDs)
}
