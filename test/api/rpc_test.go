package api_test

import (
	"context"
	"testing"

	"sumeru/core/server/api"
)

func TestDispatchRPC_invalidJSON(t *testing.T) {
	_, err := api.DispatchRPC(context.Background(), []byte(`{`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDispatchRPC_missingModel(t *testing.T) {
	_, err := api.DispatchRPC(context.Background(), []byte(`{"method":"search"}`))
	if err == nil {
		t.Fatal("expected error")
	}
}
