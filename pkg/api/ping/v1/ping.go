package v1

import (
	"fmt"
	"io"

	"github.com/emicklei/go-restful"
	"github.com/lucheng0127/bumblebee/pkg/utils/host"
)

type pingHandler struct {
	zone string
	role string
}

func (h *pingHandler) ping(req *restful.Request, rsp *restful.Response) {
	io.WriteString(rsp, fmt.Sprintf("pong from zone: %s host: %s role: %s", h.zone, host.GetHostname(), h.role))
}
