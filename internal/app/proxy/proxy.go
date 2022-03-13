package proxy

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httputil"

	"sa/internal/app/store"
	"sa/internal/pkg/cert"

	"github.com/sirupsen/logrus"
)

type Proxy struct {
	certificateManager *cert.CertificateManager
	logger             *logrus.Logger
	store              *store.Connection
}

func NewProxy(logger *logrus.Logger, store *store.Connection) *Proxy {
	return &Proxy{
		certificateManager: cert.GetCertificateManager(),
		logger:             logger,
		store:              store,
	}
}

func (p *Proxy) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	reqid, err := p.store.WriteRequest(r)
	if err != nil {
		return
	}
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

	err = p.store.WriteResponse(response, reqid)
	if err != nil {
		p.logger.Error(err)
		return
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
	reqid, err := p.store.WriteRequest(r)
	if err != nil {
		return
	}

	cert_, err := p.certificateManager.GenerateCertificate(r.URL.Hostname())
	if err != nil {
		p.logger.Errorf("[HandleHTTPS:GenerateCertificate] %s", err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*cert_},
		ServerName:   r.URL.Hostname(),
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

	_, err = client.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
	if err != nil {
		client.Close()
		p.logger.Errorf("[HandleHTTPS] cannot send established connection message %s", err.Error())
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
	go p.transferWithStore(tlsConn, destination, r, reqid)
}

func (p *Proxy) transfer(dest io.WriteCloser, source io.ReadCloser) {
	defer dest.Close()
	defer source.Close()
	io.Copy(dest, source)
}

func (p *Proxy) transferWithStore(dest io.WriteCloser, source io.ReadCloser, r *http.Request, reqid string) {
	defer dest.Close()
	defer source.Close()

	// start log answer (response) to database
	buf_reader := bufio.NewReader(source)
	response, err := http.ReadResponse(buf_reader, r)
	if err != nil {
		p.logger.Errorf("[transferWithStore] error in ReadResponse %s", err.Error())
		return
	}
	err = p.store.WriteResponse(response, reqid)
	if err != nil {
		p.logger.Errorf("[transferWithStore] error in WriteResponse %s", err.Error())
		return
	}
	b, err := httputil.DumpResponse(response, true)
	if err != nil {
		p.logger.Errorf("[transferWithStore] error in DumpResponse %s", err.Error())
		return
	}
	// end log

	io.Copy(dest, bytes.NewReader(b))
}
