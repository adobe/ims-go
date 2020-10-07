package ims_test

import (
	"fmt"
	"github.com/adobe/ims-go/ims"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetProfile(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("invalid method: %v", r.Method)
		}
		if v := r.Header.Get("authorization"); v != "Bearer accessToken" {
			t.Fatalf("invalid authorization header: %v", v)
		}

		fmt.Fprint(w, `{"foo":"bar"}`)
	}))
	defer s.Close()

	c, err := ims.NewClient(&ims.ClientConfig{
		URL: s.URL,
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	res, err := c.GetProfile(&ims.GetProfileRequest{
		AccessToken: "accessToken",
	})
	if err != nil {
		t.Fatalf("get profile: %v", err)
	}

	if body := string(res.Body); body != `{"foo":"bar"}` {
		t.Fatalf("invalid body: %v", body)
	}
}

func TestGetProfileEmptyErrorResponse(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer s.Close()

	c, err := ims.NewClient(&ims.ClientConfig{
		URL: s.URL,
	})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	res, err := c.GetProfile(&ims.GetProfileRequest{})
	if res != nil {
		t.Fatalf("non-nil response returned")
	}
	if err == nil {
		t.Fatalf("nil error returned")
	}
	if _, ok := ims.IsError(err); !ok {
		t.Fatalf("invalid error type: %v", err)
	}
}
