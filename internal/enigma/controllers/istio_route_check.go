package controllers

import (
	"net/http"
	"sync"

	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/ktr0731/grpc-web-go-client/grpcweb"
	"github.com/manabie-com/backend/internal/enigma/configurations"
	bobPb "github.com/manabie-com/backend/pkg/genproto/bob"
)

type IstioRouteCheckerController struct {
	logger  *zap.Logger
	config  *configurations.Config
	checker TryEndpointer
}

type TryEndpointer interface {
	TryEndpoint(host, service string) (ok bool, err error)
}

func RegisterLBCheckerController(r *gin.RouterGroup, l *zap.Logger, c *configurations.Config) {
	controller := &IstioRouteCheckerController{
		logger:  l,
		config:  c,
		checker: &GRPCWebEndpoint{},
	}

	r.GET("/check", controller.HandleCheck)
}

// Handy Command:
// grep web-api ./deployments/helm/platforms/gateway/*.yaml
// yq e '.apiHttp[].match[].uri.prefix' deployments/helm/manabie-all-in-one/charts/*/values.yaml
// kubectl get vs -A
// grep service $(fd '\.proto$')

const port = ":31400"
const method = "OMGWTFBBQ"

type IstioRouteResult struct {
	Key   string `json:"key"`
	Error string `json:"error,omitempty"`
}

// HandleCheck does a live check on grpc-web endpoints for each (domain, service) pair.
// this validates our routing config, and helps catch issues where we misconfigure a virtual-service
func (c *IstioRouteCheckerController) HandleCheck(ctx *gin.Context) {
	var wg sync.WaitGroup
	resCh := make(chan *IstioRouteResult)

	tryEndpoint := func(host string, service string) {
		key := fmt.Sprintf("%s/%s", host, service)
		ok, err := c.checker.TryEndpoint(host, service)
		res := &IstioRouteResult{
			Key:   key,
			Error: err.Error(),
		}
		if ok {
			res.Error = ""
		}
		resCh <- res
		wg.Done()
	}

	wg.Add(len(c.config.RouteCheckerHosts) * len(c.config.RouteCheckerServices))
	for _, host := range c.config.RouteCheckerHosts {
		for _, service := range c.config.RouteCheckerServices {
			go tryEndpoint(host+port, service)
		}
	}

	go func() {
		wg.Wait()
		close(resCh)
	}()

	results := map[string]string{}
	pass := true

	for res := range resCh {
		if res.Error == "" {
			results[res.Key] = "OK"
		} else {
			pass = false
			results[res.Key] = res.Error
			c.logger.Sugar().Errorf("virtualHost %s failed check: %s", res.Key, res.Error)
		}
	}
	if !pass {
		ctx.JSON(http.StatusBadGateway, results)
		return
	}

	ctx.JSON(http.StatusOK, results)
}

type GRPCWebEndpoint struct{}

// tryEndpoint checks that a grpc service can be found via GRPC web
// if a bad domain is used, it will fail with TCP errors
// if a bad service is used it will fail with an unknown service error.
// if a good domain / service is used it will pass, because we correctly get an unknown method error.
func (i *GRPCWebEndpoint) TryEndpoint(host, service string) (ok bool, err error) {
	client, err := grpcweb.DialContext(host)
	if err != nil {
		return
	}

	uri := fmt.Sprintf("/%s/%s", service, method)
	methodUnimplemented := fmt.Sprintf("rpc error: code = Unimplemented desc = unknown method %s for service %s", method, service)

	err = client.Invoke(
		context.Background(),
		uri,
		&bobPb.GetBasicProfileRequest{},
		&bobPb.GetBasicProfileResponse{},
		grpcweb.CallContentSubtype("proto"),
	)
	if err != nil {
		if err.Error() == methodUnimplemented {
			ok = true
		}
	}

	return
}
