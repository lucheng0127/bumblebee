package consul

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/hashicorp/consul/api"
	log "github.com/sirupsen/logrus"
)

type Consul struct {
	Client *api.Client
}

func NewConsul(ep string) (*Consul, error) {
	epUrl, err := url.Parse(ep)
	if err != nil {
		return nil, err
	}

	cfg := api.DefaultConfig()
	cfg.Scheme = epUrl.Scheme
	cfg.Address = epUrl.Host

	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &Consul{Client: client}, nil
}

func (c *Consul) ServiceRegistration(ctx context.Context, zone, hostname, addr string, timeout, interval int, exit func()) {
	registeration := &api.AgentServiceRegistration{
		ID:      hostname,
		Name:    zone,
		Address: addr,
		Check: &api.AgentServiceCheck{
			// Check field refer in https://developer.hashicorp.com/consul/api-docs/v1.15.x/agent/check#register-check
			CheckID:                        fmt.Sprintf("%s/%s", zone, hostname),
			TTL:                            fmt.Sprintf("%ds", timeout),
			DeregisterCriticalServiceAfter: "1m",
		},
	}

	if err := c.Client.Agent().ServiceRegister(registeration); err != nil {
		panic(err)
	}
	defer c.Client.Agent().ServiceDeregister(registeration.ID)

	ticker := time.NewTicker(time.Second * time.Duration(interval))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Infof("exit signal recived stop to update TTL to consul service %s", zone)
			exit()
			return
		case <-ticker.C:
			if err := c.Client.Agent().UpdateTTL(registeration.Check.CheckID, "TTL check pass", api.HealthPassing); err != nil {
				log.Errorf("failed to update TTL to consul service %s with check ID %s: %s", zone, registeration.Check.CheckID, err.Error())
			}

			log.Debugf("update TTL to consul service %s with check ID %s", zone, registeration.Check.CheckID)
		}
	}
}

func (c *Consul) ServiceDiscovery(zone string) ([]*api.ServiceEntry, error) {
	entries, _, err := c.Client.Health().Service(zone, "", true, &api.QueryOptions{})
	if err != nil {
		return nil, err
	}

	return entries, nil
}
