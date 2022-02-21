package server

import (
	"crypto/tls"
	"log"
	"net/http"
	"sa/internal/app/proxy"

	"github.com/sirupsen/logrus"
)

type Server struct {
	server http.Server
	logger *logrus.Logger
}

func NewServer(addr string, logger *logrus.Logger) *Server {
	prx := proxy.NewProxy(logger)
	return &Server{
		server: http.Server{
			Addr: addr,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
