package iam

import (
	"github.com/emicklei/go-restful"
	v1 "github.com/lucheng0127/bumblebee/pkg/api/iam/v1"
	"github.com/lucheng0127/bumblebee/pkg/utils/jwt"
	"xorm.io/xorm"
)

func AddToContainer(container *restful.Container, jwtAuth *jwt.JwtAuthenticator, dbClient *xorm.Engine) {
	v1.AddToContainer(container, jwtAuth, dbClient)
}
