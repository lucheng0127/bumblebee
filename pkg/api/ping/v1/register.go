package v1

import (
	"github.com/emicklei/go-restful"
	"github.com/lucheng0127/bumblebee/pkg/utils/runtime"
)

func AddToContainer(container *restful.Container, zone, role string) {
	ws := runtime.NewApiWebService("ping", "v1")

	handler := &pingHandler{zone: zone, role: role}

	ws.Route(
		ws.GET("/ping").To(handler.ping),
	)

	container.Add(ws)
}
