package tom

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/godogutil"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	tom_constants "github.com/manabie-com/backend/internal/tom/constants"
	pb "github.com/manabie-com/backend/pkg/genproto/tom"
)

func (s *suite) initLocationConversationConfigInDB(ctx context.Context, configValue string) (context.Context, error) {
	s.CommonSuite.ConfigKeys = []string{tom_constants.ChatConfigKeyStudent, tom_constants.ChatConfigKeyParent}
	valueArgs := make([]interface{}, 0)
	valueStrings := make([]string, 0)
	for i, key := range s.CommonSuite.ConfigKeys {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, now(), now())", i*3+1, i*3+2, i*3+3))
		valueArgs = append(valueArgs, key)
		valueArgs = append(valueArgs, "boolean")
		valueArgs = append(valueArgs, "CONFIGURATION_TYPE_EXTERNAL")
	}
	query := fmt.Sprintf(`
	INSERT INTO configuration_key (config_key, value_type, configuration_type, created_at, updated_at)
	 VALUES %s`, strings.Join(valueStrings, ","))
	_, _ = s.masterMgmtDBTrace.Exec(ctx, query, valueArgs...)

	resourcePath, _ := interceptors.ResourcePathFromContext(ctx)
	domainName := fmt.Sprintf("test%v", resourcePath)

	// Insert new org into org table for mastermgmt db
	insertOrgQuery := `INSERT INTO public.organizations
	(organization_id, tenant_id, "name", resource_path, domain_name, logo_url, country, created_at, updated_at, deleted_at, scrypt_signer_key, scrypt_salt_separator, scrypt_rounds, scrypt_memory_cost)
	VALUES($1, $1, $1, $1, $2, '', '', timezone('utc'::text, now()), timezone('utc'::text, now()), now(), '', '', '', '') ON CONFLICT DO NOTHING;`
	_, err := s.masterMgmtDBTrace.Exec(ctx, insertOrgQuery, resourcePath, domainName)
	if err != nil {
		return ctx, fmt.Errorf("cannot insert organizations for resource path: %v, err: %s", resourcePath, err)
	}

	// Insert external config for current resource path in ctx
	query2 := `INSERT INTO public.external_configuration_value
	(configuration_id, config_key, config_value, config_value_type, last_editor, created_at, updated_at, deleted_at, resource_path)
	VALUES(uuid_generate_v4(), $1, $2, 'boolean'::text, '', now(), now(), now(), $3) ON CONFLICT DO NOTHING;`

	for _, configKey := range s.CommonSuite.ConfigKeys {
		_, err := s.masterMgmtDBTrace.Exec(ctx, query2, configKey, configValue, resourcePath)
		if err != nil {
			return ctx, fmt.Errorf("cannot seed configurations for resource path: %v, err: %s", resourcePath, err)
		}
	}

	query3 := `WITH location_ids AS (
	select location_id from unnest(cast($1 as text[])) as location_id )
	INSERT INTO location_configuration_value_v2 (location_config_id, config_key, location_id, config_value_type, config_value, created_at, updated_at, resource_path)
	 select uuid_generate_v4() AS uuid_generate_v4, e.config_key, li.location_id, e.config_value_type, $3, now(), now(), e.resource_path  
	 from location_ids li CROSS JOIN external_configuration_value e where  e.config_key = any($2) ON CONFLICT ON constraint location_configuration_value_resource_unique_v2
	DO UPDATE SET config_value = $3, updated_at = now() WHERE location_configuration_value_v2.config_key = any($2)`

	orgID := s.CommonSuite.DefaultLocationID
	locationIDs := []string{orgID}

	_, err = s.masterMgmtDBTrace.Exec(ctx, query3, locationIDs, s.CommonSuite.ConfigKeys, configValue)
	if err != nil {
		return ctx, fmt.Errorf("cannot seed configurations, err: %s", err)
	}

	return ctx, nil
}

func (s *suite) studentCanSeeConversations(ctx context.Context, numChat int) (context.Context, error) {
	recentRes := s.Response.(*pb.ConversationListResponse)
	if len(recentRes.GetConversations()) != numChat {
		return ctx, fmt.Errorf("expect %d chats, %d returned", numChat, len(recentRes.GetConversations()))
	}
	studentID := s.childrenIDs[0]
	for _, item := range recentRes.GetConversations() {
		if item.ConversationType != pb.CONVERSATION_STUDENT {
			return ctx, fmt.Errorf("expect conversation student, got %s", item.ConversationType.String())
		}
		if item.StudentId != studentID {
			return ctx, fmt.Errorf("conversation returned with student id %s not equal to current student id %s", item.StudentId, studentID)
		}
	}

	return ctx, nil
}

func (s *suite) aStudentConversation(ctx context.Context) (context.Context, error) {
	var childrenIDs []string
	teacherID := idutil.ULIDNow()

	ctx, err := godogutil.MultiErrChain(ctx,
		s.createStudentConversation,
		s.multipleTeachersJoinConversation, []string{teacherID},
	)
	if err != nil {
		return ctx, err
	}
	childrenIDs = append(childrenIDs, s.studentID)
	s.childrenIDs = childrenIDs
	return ctx, nil
}
