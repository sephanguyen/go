package bootstrap

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/configs"
)

func (b *bootstrapper[T]) initLogger(config interface{}, rsc *Resources) error {
	c, err := extract[configs.CommonConfig](config, commonFieldName)
	if err != nil {
		return fmt.Errorf("failed to extract config.Common: %w", err)
	}

	_ = rsc.WithServiceName(c.Name).WithLoggerC(c)
	return nil
}
