package filters

import (
	"net/http"

	"k8s.io/apimachinery/pkg/util/uuid"
)

const (
	HeaderOperationID = "X-Operation-ID"
)

func WithOperationID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(HeaderOperationID)

		if id == "" {
			r.Header.Set(HeaderOperationID, string(uuid.NewUUID()))
		}

		next.ServeHTTP(w, r)
	})
}
