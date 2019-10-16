package ims_test

import (
	"testing"

	"github.com/adobe/ims-go/ims"
)

func TestNewClientNoScheme(t *testing.T) {
	_, err := ims.NewClient(&ims.ClientConfig{
		URL: "ims.endpoint",
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() != "missing URL scheme" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewClientNoHost(t *testing.T) {
	_, err := ims.NewClient(&ims.ClientConfig{
		URL: "https:///path",
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() != "missing URL host" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewClientMalformedURL(t *testing.T) {
	_, err := ims.NewClient(&ims.ClientConfig{
		URL: ":",
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() != "malformed URL" {
		t.Fatalf("unexpected error: %v", err)
	}
}
