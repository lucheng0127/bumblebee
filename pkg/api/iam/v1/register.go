package v1

import (
	"github.com/emicklei/go-restful"
	"github.com/lucheng0127/bumblebee/pkg/utils/jwt"
	"github.com/lucheng0127/bumblebee/pkg/utils/runtime"
	"xorm.io/xorm"
)

func AddToContainer(container *restful.Container, jwtAuth *jwt.JwtAuthenticator, dbClient *xorm.Engine) {
	ws := runtime.NewApiWebService("iam", "v1")
	ws.Produces(restful.MIME_JSON)

	handler := &authHandler{authenticator: jwtAuth, dbClient: dbClient}

	ws.Route(
		ws.POST("/login").
			Consumes("application/x-www-form-urlencoded").
			To(handler.login),
	)

	container.Add(ws)
}
