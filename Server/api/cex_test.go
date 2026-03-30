package api_test

import (
	"testing"

	"example.com/go/crypto/api"
)

func TestAPICall(t *testing.T) {
	_, err := api.GetRate("")
	if err == nil {
		t.Fatal("expected an error for empty currency")
	}
}
