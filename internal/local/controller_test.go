package local

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type trackingReadCloser struct {
	io.Reader
	closed bool
}

func (r *trackingReadCloser) Close() error {
	r.closed = true
	return nil
}

func TestValidateAdminPasswordClosesResponseBody(t *testing.T) {
	body := &trackingReadCloser{Reader: strings.NewReader(`{"login":"admin"}`)}
	local := &Local{
		ctx:     context.Background(),
		baseURL: "http://gitea.test",
		httpClient: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if got := req.URL.String(); got != "http://gitea.test/api/v1/user" {
					t.Fatalf("request URL = %q", got)
				}
				if got := req.Header.Get("Authorization"); got != "Basic YWRtaW46c2VjcmV0" {
					t.Fatalf("Authorization header = %q", got)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Status:     "200 OK",
					Body:       body,
					Header:     make(http.Header),
					Request:    req,
				}, nil
			}),
		},
	}

	if err := local.validateAdminPassword("admin", "secret"); err != nil {
		t.Fatalf("validateAdminPassword() error = %v", err)
	}
	if !body.closed {
		t.Fatal("response body was not closed")
	}
}
