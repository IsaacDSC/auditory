package proxy

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type InputAudit struct {
	Method     string
	Path       string
	Query      string
	Headers    http.Header
	Body       []byte
	StatusCode int
}

type AuditorFn func(ctx context.Context, input InputAudit) error

type AuditProxy struct {
	proxy            *httputil.ReverseProxy
	target           *url.URL
	requestCallback  AuditorFn
	responseCallback AuditorFn
}

func NewAuditProxy(target *url.URL, requestCallback AuditorFn, responseCallback AuditorFn) *AuditProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)

	ap := &AuditProxy{
		proxy:            proxy,
		target:           target,
		requestCallback:  requestCallback,
		responseCallback: responseCallback,
	}

	proxy.Director = ap.director
	proxy.ModifyResponse = ap.modifyResponse

	return ap
}

func (ap *AuditProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ap.proxy.ServeHTTP(w, r)
}

func (ap *AuditProxy) director(r *http.Request) {
	r.URL.Scheme = ap.target.Scheme
	r.URL.Host = ap.target.Host
	r.Host = ap.target.Host

	// Store request for audit
	ap.auditRequest(r)
}

func (ap *AuditProxy) auditRequest(r *http.Request) {
	var body []byte
	if r.Body != nil {
		body, _ = io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = time.Now().Format("20060102150405.000000000")
	}

	ctx := r.Context()
	if err := ap.requestCallback(ctx, InputAudit{
		Headers: r.Header,
		Body:    body,
		Method:  r.Method,
		Path:    r.URL.Path,
		Query:   r.URL.RawQuery,
	}); err != nil {
		log.Printf("audit error: %v", err)
	}
}

func (ap *AuditProxy) modifyResponse(resp *http.Response) error {
	var body []byte
	if resp.Body != nil {
		body, _ = io.ReadAll(resp.Body)
		resp.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	ctx := resp.Request.Context()
	if err := ap.responseCallback(ctx, InputAudit{
		Headers:    resp.Header,
		Body:       body,
		StatusCode: resp.StatusCode,
	}); err != nil {
		log.Printf("audit error: %v", err)
	}

	return nil
}
