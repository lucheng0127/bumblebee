package filters

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

func WithAccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Infof("%s [%s] %s", r.Header.Get(HeaderOperationID), r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
