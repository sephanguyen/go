package storage

import (
	"strings"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/payment/utils"
)

func NewStorageService(storageConfig configs.StorageConfig) (storage utils.IStorage, err error) {
	if strings.Contains(storageConfig.Endpoint, "minio") {
		storage, err = NewMinIOStoreService(&storageConfig)
	} else {
		storage, err = NewGCloudStoreService(&storageConfig)
	}
	return
}
