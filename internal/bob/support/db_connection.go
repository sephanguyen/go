package support

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	unleash_client "github.com/manabie-com/backend/internal/golibs/unleashclient"
)

type WrapperDBConnection struct {
	bobDB         database.Ext
	lessonDB      database.Ext
	unleashClient unleash_client.ClientInstance
	env           string
}

func (wdb *WrapperDBConnection) GetDB(resourcePath string) (database.Ext, error) {
	isUnleashToggled, err := wdb.unleashClient.IsFeatureEnabledOnOrganization("Lesson_LessonManagement_BackOffice_SwitchNewDBConnection", wdb.env, resourcePath)
	if err != nil {
		return nil, fmt.Errorf("[isSwitchToNewDBEnabled] IsFeatureEnabled() failed: %s", err)
	}
	if isUnleashToggled {
		return wdb.lessonDB, nil
	}
	return wdb.bobDB, nil
}

func InitWrapperDBConnector(bobDB, lessonDB database.Ext, unleashClient unleash_client.ClientInstance, env string) *WrapperDBConnection {
	return &WrapperDBConnection{
		bobDB:         bobDB,
		lessonDB:      lessonDB,
		unleashClient: unleashClient,
		env:           env,
	}
}
