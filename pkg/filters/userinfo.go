package filters

import (
	"net/http"
	"strings"

	"github.com/lucheng0127/bumblebee/pkg/utils/jwt"
)

const (
	HeaderUserInfo  = "X-User-Info"
	HeaderAuthToken = "Authorization"
)

func WithUserInfo(next http.Handler, jwtAuth *jwt.JwtAuthenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check is skip auth url
		requestInfo := RequestInfoFromHttpRequest(r)

		if requestInfo == nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("no request info in request"))
			return
		}

		if requestInfo.SkipAuthenticate {
			next.ServeHTTP(w, r)
			return
		}

		// Get jwt tokn from http header
		authInfo := r.Header.Get(HeaderAuthToken)
		authItems := strings.Split(authInfo, " ")
		if len(authItems) != 2 || authItems[0] != "Bearer" {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("invalidate auth info"))
			return
		}

		// Verify token and set user info into header
		uid, err := jwtAuth.VerifyToken(authItems[1])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		r.Header.Set(HeaderUserInfo, uid)

		next.ServeHTTP(w, r)
	})
}
