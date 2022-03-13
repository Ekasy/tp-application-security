package store

import (
	"bytes"
	"context"
	"errors"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Connection struct {
	conn   *mongo.Collection
	ctx    context.Context
	logger *logrus.Logger
}

func NewConnection(logger *logrus.Logger) *Connection {
	url := os.Getenv("MONGO_URL")
	if url == "" {
		logger.Fatalln("[NewConnection] cannot get env MONGO_URL")
	}
	ctx := context.TODO()
	clientOptions := options.Client().ApplyURI(url)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		logger.Fatalf("[NewConnection] cannot connect to mongo: %s", err.Error())
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		logger.Fatalf("[NewConnection] cannot ping to mongo: %s", err.Error())
	}

	coll := client.Database("application_security").Collection("requests")
	return &Connection{
		conn:   coll,
		ctx:    ctx,
		logger: logger,
	}
}

var MY_RANDOM *rand.Rand = rand.New(rand.NewSource(time.Now().Unix()))

func generateReqId(n int) string {
	var symbols = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	s := make([]rune, n)
	for i := range s {
		s[i] = symbols[MY_RANDOM.Intn(len(symbols))]
	}
	return string(s)
}

type Request struct {
	Timestamp int                    `bson:"timestamp"`
	Method    string                 `bson:"method"`
	Url       string                 `bson:"url"`
	Path      string                 `bson:"path"`
	CgiParams map[string]string      `bson:"cgi_params"`
	Header    map[string]string      `bson:"headers"`
	Cookies   map[string]interface{} `bson:"cookies"`
	Body      string                 `bson:"body"`
}

type Document struct {
	ReqId     string   `bson:"reqid"`
	Request_  Request  `bson:"request"`
	Response_ Response `bson:"response"`
}

func (conn *Connection) requestToBson(r *http.Request) (string, *bson.D) {
	reqid := generateReqId(64)

	cgies := strings.Split(r.URL.Query().Encode(), "&")
	cgi_params := make(map[string]string)
	for _, cgi := range cgies {
		if cgi == "" {
			continue
		}
		kv_cgi := strings.Split(cgi, "=")
		cgi_params[kv_cgi[0]] = kv_cgi[1]
	}

	header := make(map[string]string)
	for k, value := range r.Header {
		for _, v := range value {
			if v == "" {
				continue
			}
			header[k] = v
		}
	}

	cookies := make(map[string]interface{})
	for _, cookie := range r.Cookies() {
		cookies[cookie.Name] = cookie.Value
	}

	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		conn.logger.Errorf("[requestToBson] cannot read body: %s", err.Error())
		return "", nil
	}

	req := &Request{
		Timestamp: int(time.Now().Unix()),
		Method:    r.Method,
		Url:       r.URL.String(),
		Path:      r.URL.Path,
		CgiParams: cgi_params,
		Header:    header,
		Cookies:   cookies,
		Body:      string(b),
	}

	doc_ := &Document{
		ReqId:    reqid,
		Request_: *req,
	}

	data, err := bson.Marshal(doc_)
	if err != nil {
		conn.logger.Warnf("[requestToBson] cannot marshal request: %s", err.Error())
		return "", nil
	}

	doc := &bson.D{}
	err = bson.Unmarshal(data, doc)
	if err != nil {
		conn.logger.Warnf("[requestToBson] cannot unmarshal: %s", err.Error())
		return "", nil
	}
	return reqid, doc
}

func (conn *Connection) WriteRequest(r *http.Request) (string, error) {
	reqid, doc := conn.requestToBson(r)
	if doc == nil {
		return "", errors.New("[WriteRequest] cannot convert request to bson")
	}

	_, err := conn.conn.InsertOne(conn.ctx, doc)
	if err != nil {
		return "", err
	}

	return reqid, nil
}

type Response struct {
	Timestamp  int               `bson:"timestamp"`
	StatusCode int               `bson:"status_code"`
	Message    string            `bson:"message"`
	Header     map[string]string `bson:"headers"`
	Body       string            `bson:"body"`
}

func (conn *Connection) responseToBson(r *http.Response, reqid string) *bson.D {
	header := make(map[string]string)
	for k, value := range r.Header {
		for _, v := range value {
			if v == "" {
				continue
			}
			header[k] = v
		}
	}

	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		conn.logger.Errorf("[responseToBson] cannot read body: %s", err.Error())
		return nil
	}

	resp := &Response{
		Timestamp:  int(time.Now().Unix()),
		StatusCode: r.StatusCode,
		Message:    r.Status,
		Header:     header,
		Body:       string(b),
	}

	// once read the body, you need to set the buffer back
	r.Body = io.NopCloser(bytes.NewReader(b))

	data, err := bson.Marshal(resp)
	if err != nil {
		conn.logger.Warnf("[responseToBson] cannot marshal request: %s", err.Error())
		return nil
	}

	doc := &bson.D{}
	err = bson.Unmarshal(data, doc)
	if err != nil {
		conn.logger.Warnf("[responseToBson] cannot unmarshal: %s", err.Error())
		return nil
	}
	return doc
}

func (conn *Connection) WriteResponse(r *http.Response, reqid string) error {
	doc := conn.responseToBson(r, reqid)
	if doc == nil {
		return errors.New("[WriteResponse] cannot convert request to bson")
	}

	_, err := conn.conn.UpdateOne(
		conn.ctx,
		bson.M{"reqid": reqid},
		bson.D{primitive.E{
			Key: "$set",
			Value: bson.D{primitive.E{
				Key:   "response",
				Value: doc,
			}},
		}},
	)

	if err != nil {
		conn.logger.Info("[WriteResponse] cannot insert document: %s", err.Error())
		return err
	}

	return nil
}

func (conn *Connection) GetByRequestResponseReqId(reqid string) *Document {
	result := conn.conn.FindOne(conn.ctx, bson.M{"reqid": reqid})
	doc := &Document{}
	err := result.Decode(doc)
	if err != nil {
		conn.logger.Info("[GetByRequestResponseReqId] cannot find document by reqid=%s: %s", reqid, err.Error())
		return nil
	}
	return doc
}
