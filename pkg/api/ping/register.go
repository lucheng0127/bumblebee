package ping

import (
	"github.com/emicklei/go-restful"
	v1 "github.com/lucheng0127/bumblebee/pkg/api/ping/v1"
)

func AddToContainer(container *restful.Container, zone string, master bool) {
	role := "slave"
	if master {
		role = "master"
	}

	v1.AddToContainer(container, zone, role)
}
