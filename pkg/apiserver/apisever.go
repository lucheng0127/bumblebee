package apiserver

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/emicklei/go-restful"
	"github.com/lucheng0127/bumblebee/pkg/api/iam"
	"github.com/lucheng0127/bumblebee/pkg/api/ping"
	"github.com/lucheng0127/bumblebee/pkg/client/authorizer"
	"github.com/lucheng0127/bumblebee/pkg/client/consul"
	"github.com/lucheng0127/bumblebee/pkg/client/trace"
	"github.com/lucheng0127/bumblebee/pkg/config"
	"github.com/lucheng0127/bumblebee/pkg/dispatch/service"
	"github.com/lucheng0127/bumblebee/pkg/filters"
	"github.com/lucheng0127/bumblebee/pkg/models"
	"github.com/lucheng0127/bumblebee/pkg/utils/host"
	"github.com/lucheng0127/bumblebee/pkg/utils/jwt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	otrace "go.opentelemetry.io/otel/sdk/trace"
	"xorm.io/xorm"
)

type ApiServer struct {
	Config        *config.ApiServerConfig
	apiDocEnable  bool
	Server        http.Server
	handler       http.Handler
	Consul        *consul.Consul
	Tracer        *otrace.TracerProvider
	DBClient      *xorm.Engine
	Authenticator *jwt.JwtAuthenticator
	Authorizer    *authorizer.DBAuthorizer
	Nats          *nats.Conn
	DispatchSvc   *service.DispatchServer
}

func NewApiServer(cfg *config.ApiServerConfig, debug bool, database string) (*ApiServer, error) {
	svc := &ApiServer{
		Config:       cfg,
		apiDocEnable: debug,
		Server:       http.Server{},
	}

	consulClient, err := consul.NewConsul(cfg.Consul.Url)
	if err != nil {
		return nil, err
	}
	svc.Consul = consulClient

	if cfg.Trace.Enable {
		tracer, err := trace.NewTrace(cfg.Trace.Addr)
		if err != nil {
			return nil, err
		}

		svc.Tracer = tracer
	}

	natsConn, err := nats.Connect(svc.Config.Nats.Url)
	if err != nil {
		return nil, err
	}

	svc.Nats = natsConn

	if svc.Config.Platform.Master {
		dbClient, err := xorm.NewEngine("sqlite3", database)
		if err != nil {
			return nil, err
		}
		svc.DBClient = dbClient
		if debug {
			svc.DBClient.ShowSQL(true)
		}

		dbAuthorizer, err := authorizer.NewDBAuthorizer(database)
		if err != nil {
			return nil, err
		}
		svc.Authorizer = dbAuthorizer

		svc.Authenticator = jwt.NewJwtAuthenticator(svc.Config.Platform.Secret, 10, 2)

		svc.DispatchSvc = service.NewDispatchServer(fmt.Sprintf("%s/%s", svc.Config.Platform.Zone, host.GetHostname()), svc.Nats)
	}

	return svc, nil
}

func (svc *ApiServer) syncDB() error {
	return svc.DBClient.Sync(new(models.User))
}

func (svc *ApiServer) initFilters() {
	if svc.Config.Platform.Master {
		// svc.handler = filters.WithDispatchByTCP(svc.handler, svc.Consul)
		svc.handler = filters.WithDispatchByNats(svc.handler, svc.Consul, svc.DispatchSvc, svc.Config.Platform.Zone)
		svc.handler = filters.WithAutorizer(svc.handler, svc.Authorizer)
		svc.handler = filters.WithUserInfo(svc.handler, svc.Authenticator)
		svc.handler = filters.WithTrace(svc.handler, svc.Config.Trace.Enable, svc.Config.Platform.Master, svc.Tracer, svc.Config.Platform.Zone)
		svc.handler = filters.WithAccessLog(svc.handler)
		svc.handler = filters.WithRequestInfo(svc.handler)
		// Master generate operation id, slave read from header
		svc.handler = filters.WithOperationID(svc.handler)
	} else {
		svc.handler = filters.WithTrace(svc.handler, svc.Config.Trace.Enable, svc.Config.Platform.Master, svc.Tracer, svc.Config.Platform.Zone)
		svc.handler = filters.WithAccessLog(svc.handler)
	}
}

func (svc *ApiServer) installApi() {
	container := restful.NewContainer()
	container.DoNotRecover(false)

	if svc.Config.Platform.Master {
		iam.AddToContainer(container, svc.Authenticator, svc.DBClient)
	}

	ping.AddToContainer(container, svc.Config.Platform.Zone, svc.Config.Platform.Master)

	svc.handler = container
}

func (svc *ApiServer) initAdminUser() (string, error) {
	var user models.User
	ok, err := svc.DBClient.Get(&user)
	if err != nil {
		return "", err
	}

	if ok {
		return user.Uuid, nil
	}

	au := models.NewUser(svc.Config.Platform.AdminUser.Username, svc.Config.Platform.AdminUser.Password)
	_, err = svc.DBClient.Insert(au)
	if err != nil {
		return "", err
	}

	return au.Uuid, nil
}

func (svc *ApiServer) initAdminPolicy(uid string) error {
	for _, act := range []string{authorizer.PolicyActGet, authorizer.PolicyActPost, authorizer.PolicyActPut, authorizer.PolicyActPatch, authorizer.PolicyActDELETE} {
		if err := svc.Authorizer.CreateRolePolicy(authorizer.RolePolicy{
			RoleName: authorizer.AdminSubPrefix + act,
			Url:      authorizer.AdminPolicyObj,
			Method:   act,
		}); err != nil {
			return err
		}

		if err := svc.Authorizer.AddUserRole(uid, authorizer.AdminSubPrefix+act); err != nil {
			return err
		}
	}

	return nil
}

func (svc *ApiServer) PreRun(ctx context.Context) error {
	// Install routes
	svc.installApi()

	// Init filter
	svc.initFilters()

	if svc.Config.Platform.Master {
		if err := svc.syncDB(); err != nil {
			return err
		}

		adminUid, err := svc.initAdminUser()
		if err != nil {
			return err
		}

		// Init system admin policy
		if err := svc.initAdminPolicy(adminUid); err != nil {
			return err
		}
	}

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
		wg.Add(1)
		go svc.DispatchSvc.Run(ctx, func() { wg.Done() })

		svc.Server.Addr = fmt.Sprintf(":%d", svc.Config.Platform.Port)
		log.Infof("run apiserver master on port %d", svc.Config.Platform.Port)
		if err := svc.Server.ListenAndServe(); err != nil {
			return err
		}
	} else {
		//ln, err := net.Listen("tcp", fmt.Sprintf(":%d", svc.Config.Platform.Port))
		//if err != nil {
		//	return err
		//}
		ln := service.NewClientLinstener(fmt.Sprintf("%s/%s", svc.Config.Platform.Zone, host.GetHostname()), svc.Nats)

		// log.Infof("run apiserver slave on port %d", svc.Config.Platform.Port)
		log.Info("run apiserver slave on nats linstener")
		if err := svc.Server.Serve(ln); err != nil {
			return err
		}
	}

	wg.Wait()
	return nil
}
