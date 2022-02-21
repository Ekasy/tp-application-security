package proxy

import (
	"crypto/tls"
	"io"
	"net/http"

	"sa/internal/pkg/cert"

	"github.com/sirupsen/logrus"
)

type Proxy struct {
	mainCertificate *cert.Certificate
	logger          *logrus.Logger
}

func NewProxy(logger *logrus.Logger) *Proxy {
	return &Proxy{
		mainCertificate: cert.GetSertificate(),
		logger:          logger,
	}
}

func (p *Proxy) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	response, err := p.DoRequest(r)
	if err != nil {
		return
	}

	for key, values := range response.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(response.StatusCode)

	defer response.Body.Close()
	_, err = io.Copy(w, response.Body)
	if err != nil {
		p.logger.Error(err)
	}
}

func (p *Proxy) DoRequest(r *http.Request) (*http.Response, error) {
	// prepare request
	request, err := http.NewRequest(r.Method, r.RequestURI, r.Body)
	if err != nil {
		p.logger.Errorf("[DoRequest:NewRequest] %s", err.Error())
		return nil, err
	}

	// copy header
	request.Header = r.Header.Clone()
	request.Header.Del("Proxy-Connection")
	r.RequestURI = ""

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// do request
	response, err := client.Do(request)
	if err != nil {
		p.logger.Errorf("[DoRequest:Do] %s", err.Error())
		return nil, err
	}
	return response, err
}

func (p *Proxy) HandleHTTPS(w http.ResponseWriter, r *http.Request) {
	cert_, err := cert.CreateLeafCertificate(r.Host)
	if err != nil {
		p.logger.Errorf("[HandleHTTPS:CreateLeafCertificate] %s", err.Error())
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*cert_},
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			return cert.CreateLeafCertificate(info.ServerName)
		},
	}

	destination, err := tls.Dial("tcp", r.Host, tlsConfig)
	if err != nil {
		p.logger.Errorf("[HandleHTTPS:Dial] %s", err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	hr, ok := w.(http.Hijacker)
	if !ok {
		p.logger.Error("[HandleHTTPS] hijacker not supported]")
		http.Error(w, "Hijacker not supported", http.StatusServiceUnavailable)
		return
	}

	client, _, err := hr.Hijack()
	if err != nil {
		p.logger.Errorf("[HandleHTTPS:Hijack] %s", err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	tlsConn := tls.Server(client, tlsConfig)
	err = tlsConn.Handshake()
	if err != nil {
		p.logger.Errorf("[HandleHTTPS:Handshake] %s", err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	go p.transfer(destination, tlsConn)
	go p.transfer(tlsConn, destination)
}

func (p *Proxy) transfer(dest io.WriteCloser, source io.ReadCloser) {
	defer dest.Close()
	defer source.Close()
	io.Copy(dest, source)
}
