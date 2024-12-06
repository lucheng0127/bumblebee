package v1

import (
	"crypto/md5"
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/gorilla/schema"
	"github.com/lucheng0127/bumblebee/pkg/models"
	"github.com/lucheng0127/bumblebee/pkg/utils/jwt"
	"xorm.io/xorm"
)

var decoder *schema.Decoder

type authHandler struct {
	authenticator *jwt.JwtAuthenticator
	dbClient      *xorm.Engine
}

type TokenInfo struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func (h *authHandler) login(req *restful.Request, rsp *restful.Response) {
	err := req.Request.ParseForm()
	if err != nil {
		rsp.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}

	username, err := req.BodyParameter("username")
	if err != nil {
		rsp.WriteErrorString(http.StatusBadRequest, "username needed")
		return
	}
	password, err := req.BodyParameter("password")
	if err != nil {
		rsp.WriteErrorString(http.StatusBadRequest, "password needed")
		return
	}
	enPasswd := fmt.Sprintf("%x", md5.Sum([]byte(password)))

	var user models.User
	user.Name = username
	ok, err := h.dbClient.Get(&user)
	if err != nil {
		rsp.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	if !ok {
		rsp.WriteErrorString(http.StatusBadRequest, "invalidate username or password")
		return
	}

	if user.Password != enPasswd {
		rsp.WriteErrorString(http.StatusBadRequest, "invalidate username or password")
		return
	}

	at, err := h.authenticator.NewToken(user.Uuid, jwt.AccessTokenType)
	if err != nil {
		rsp.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}
	rt, err := h.authenticator.NewToken(user.Uuid, jwt.RefreshTokenType)
	if err != nil {
		rsp.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	rsp.WriteEntity(TokenInfo{AccessToken: at, RefreshToken: rt})
}
