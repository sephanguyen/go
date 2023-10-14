package bootstrap

import (
	"github.com/manabie-com/backend/internal/golibs/configs"
)

func initStorage(config interface{}, rsc *Resources) error {
	c, err := extract[configs.StorageConfig](config, storageFieldName)
	if err != nil {
		return ignoreErrFieldNotFound(err)
	}
	rsc.storage = c
	return nil
}
