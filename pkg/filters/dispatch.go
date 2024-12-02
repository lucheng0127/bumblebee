package filters

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/lucheng0127/bumblebee/pkg/client/consul"
	log "github.com/sirupsen/logrus"
)

func WithDispatchByTCP(next http.Handler, consul *consul.Consul) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestInfo := RequestInfoFromHttpRequest(r)

		if requestInfo == nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("no request info in request"))
			return
		}

		if !requestInfo.ZoneRequest {
			next.ServeHTTP(w, r)
			return
		}

		zoneBackends, err := consul.ServiceDiscovery(requestInfo.Zone)
		if err != nil {
			log.Errorf("failed to get zone %s backends: %s", requestInfo.Zone, err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(zoneBackends) == 0 {
			log.Errorf("no available backend in zone %s", requestInfo.Zone)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("invalidate zone %s", requestInfo.Zone)))
			return
		}

		// Random choice one from available backends
		target := zoneBackends[rand.Intn(len(zoneBackends))]

		targetURLPath := strings.Replace(r.URL.Path, fmt.Sprintf("/zones/%s", requestInfo.Zone), "", 1)
		log.Infof("proxy request %s to zone: %s backend: %s address: %s with request: %s", r.URL.Path, requestInfo.Zone, target.Service.ID, target.Service.Address, targetURLPath)
		targetUrl := r.URL
		targetUrl.Path = targetURLPath
		targetUrl.Host = "127.0.0.1"
		targetUrl.Scheme = "http"

		targetConn, err := net.Dial("tcp", target.Service.Address)
		if err != nil {
			log.Errorf("failed to dial zone %s backend %s with address %s: %s", requestInfo.Zone, target.Service.ID, target.Service.Address, err.Error())
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		proxy := &httputil.ReverseProxy{
			Director: func(request *http.Request) {
				request.URL = targetUrl
				request.Body = r.Body
				request.Header = r.Header
			},
			Transport: &http.Transport{
				DialContext: func(_ context.Context, network, addr string) (net.Conn, error) {
					return targetConn, nil
				},
			},
		}

		proxy.ServeHTTP(w, r)
	})
}
