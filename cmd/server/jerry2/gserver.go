package jerry2

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/jerry/configurations"
	"github.com/manabie-com/backend/internal/jerry/services"
	"github.com/manabie-com/backend/internal/jerry/services/hasura"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"go.opencensus.io/stats/view"
)

func init() {
	s := &server{}
	bootstrap.WithHTTP[configurations.Config2](s).
		WithMonitorServicer(s).
		Register(s)
}

type server struct{}

func (s *server) ServerName() string { return "jerry2" }

func (s *server) SetupHTTP(c configurations.Config2, ge *gin.Engine, rsc *bootstrap.Resources) error {
	if err := services.RegisterHasuraCacheService(ge, rsc.Logger(), c.HasuraCacheConfig); err != nil {
		return fmt.Errorf("services.RegisterHasuraCacheService: %s", err)
	}
	services.RegisterHealthCheckService(ge)
	return nil
}

func (s *server) InitDependencies(configurations.Config2, *bootstrap.Resources) error  { return nil }
func (s *server) WithPrometheusCollectors(*bootstrap.Resources) []prometheus.Collector { return nil }
func (s *server) InitMetricsValue()                                                    {}
func (s *server) WithOpencensusViews() []*view.View                                    { return hasura.Views() }
func (s *server) GracefulShutdown(context.Context)                                     {}
