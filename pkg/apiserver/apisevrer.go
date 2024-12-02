package apiserver

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/emicklei/go-restful"
	"github.com/lucheng0127/bumblebee/pkg/api/ping"
	"github.com/lucheng0127/bumblebee/pkg/client/consul"
	"github.com/lucheng0127/bumblebee/pkg/config"
	"github.com/lucheng0127/bumblebee/pkg/filters"
	"github.com/lucheng0127/bumblebee/pkg/utils/host"
	log "github.com/sirupsen/logrus"
)

type ApiServer struct {
	Config       *config.ApiServerConfig
	apiDocEnable bool
	Server       http.Server
	handler      http.Handler
	Consul       *consul.Consul
}

func NewApiServer(cfg *config.ApiServerConfig, docEnable bool) (*ApiServer, error) {
	svc := &ApiServer{
		Config:       cfg,
		apiDocEnable: docEnable,
		Server:       http.Server{},
	}

	consulClient, err := consul.NewConsul(cfg.Consul.Url)
	if err != nil {
		return nil, err
	}
	svc.Consul = consulClient

	return svc, nil
}

func (svc *ApiServer) initFilters() {
	if svc.Config.Platform.Master {
		svc.handler = filters.WithDispatchByTCP(svc.handler, svc.Consul)
		svc.handler = filters.WithAccessLog(svc.handler)
		svc.handler = filters.WithRequestInfo(svc.handler)
		// Master generate operation id, slave read from header
		svc.handler = filters.WithOperationID(svc.handler)
	} else {
		svc.handler = filters.WithAccessLog(svc.handler)
	}
}

func (svc *ApiServer) installApi() {
	container := restful.NewContainer()
	container.DoNotRecover(false)

	ping.AddToContainer(container, svc.Config.Platform.Zone, svc.Config.Platform.Master)

	svc.handler = container
}

func (svc *ApiServer) PreRun(ctx context.Context) error {
	// Install routes
	svc.installApi()

	// Init filter
	svc.initFilters()

	return nil
}

func (svc *ApiServer) Run(ctx context.Context) error {
	var wg sync.WaitGroup

	wg.Add(1)
	hostIP, err := host.GetOutboundIP()
	if err != nil {
		return err
	}
	go svc.Consul.ServiceRegistration(ctx, svc.Config.Platform.Zone, host.GetHostname(), fmt.Sprintf("%s:%d", hostIP.String(), svc.Config.Platform.Port), 5, 2, func() { wg.Done() })

	svc.Server.Handler = svc.handler

	if svc.Config.Platform.Master {
		svc.Server.Addr = fmt.Sprintf(":%d", svc.Config.Platform.Port)
		log.Infof("run apiserver master on port %d", svc.Config.Platform.Port)
		if err := svc.Server.ListenAndServe(); err != nil {
			return err
		}
	} else {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", svc.Config.Platform.Port))
		if err != nil {
			return err
		}

		log.Infof("run apiserver slave on port %d", svc.Config.Platform.Port)
		if err := svc.Server.Serve(ln); err != nil {
			return err
		}
	}

	wg.Wait()
	return nil
}
