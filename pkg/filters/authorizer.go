package filters

import (
	"fmt"
	"net/http"

	"github.com/lucheng0127/bumblebee/pkg/client/authorizer"
)

func WithAutorizer(next http.Handler, authorizer *authorizer.DBAuthorizer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		uid := r.Header.Get(HeaderUserInfo)
		if uid == "" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("no user info in request"))
			return
		}

		ok, err := authorizer.Authenticate(uid, r.URL.Path, r.Method)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(err.Error()))
			return
		}

		if !ok {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(fmt.Sprintf("no access to %s with %s method", r.URL.Path, r.Method)))
			return
		}

		next.ServeHTTP(w, r)
	})
}
