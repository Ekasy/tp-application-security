package server

import (
	"crypto/tls"
	"log"
	"net/http"
	"sa/internal/app/injector"
	"sa/internal/app/proxy"
	"sa/internal/app/store"
	"strings"

	"github.com/sirupsen/logrus"
)

type Server struct {
	server http.Server
	logger *logrus.Logger
}

func NewServer(addr string, logger *logrus.Logger, store *store.Connection) *Server {
	prx := proxy.NewProxy(logger, store)
	return &Server{
		server: http.Server{
			Addr: addr,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasPrefix(r.URL.Path, "/inject") {
					w = injector.InjectByReqId(r, w, store)
					return
				}

				if r.Method == http.MethodConnect {
					prx.HandleHTTPS(w, r)
				} else {
					prx.HandleHTTP(w, r)
				}
			}),
			TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
		},
		logger: logger,
	}
}

func (s *Server) Start() {
	s.logger.Infof("start server at %s", s.server.Addr)
	log.Fatal(s.server.ListenAndServe())
}
