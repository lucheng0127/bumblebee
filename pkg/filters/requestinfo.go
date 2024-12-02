package filters

import (
	"context"
	"net/http"
	"strings"
)

type ctxKey int

const (
	requestInfoKey ctxKey = iota
)

type RequestInfo struct {
	ZoneRequest bool
	Zone        string
}

func RequestInfoFromHttpRequest(r *http.Request) *RequestInfo {
	info := r.Context().Value(requestInfoKey)
	if info == nil {
		return nil
	}

	return info.(*RequestInfo)
}

func WithRequestInfo(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowdUrlPrefix := []string{"api"}
		currentPaths := strings.Split(r.URL.Path, "/")
		requestInfo := new(RequestInfo)

		// /api/<group.version>/<resources>
		if len(currentPaths) < 4 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalidate request URL"))
			return
		}

		found := false
		for _, prefix := range allowdUrlPrefix {
			if currentPaths[1] == prefix {
				found = true
				break
			}
		}
		if !found {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalidate request URL"))
			return
		}

		if currentPaths[2] == "zones" {
			requestInfo.ZoneRequest = true
			requestInfo.Zone = currentPaths[3]
			if len(currentPaths) < 6 {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("invalidate request URL"))
				return
			}
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, requestInfoKey, requestInfo)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
