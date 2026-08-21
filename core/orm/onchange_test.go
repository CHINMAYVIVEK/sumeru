package orm

import (
	"context"
	"testing"
)

func TestRegisterOnchange(t *testing.T) {
	RegisterOnchange("test.model", "name", func(ctx context.Context, values map[string]interface{}, field string) (OnchangeResult, error) {
		return OnchangeResult{Value: map[string]interface{}{"note": "ok"}}, nil
	})
	result, err := RunOnchange(context.Background(), "test.model", "name", map[string]interface{}{"name": "A"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Value["note"] != "ok" {
		t.Fatalf("expected value note=ok, got %v", result.Value)
	}
}
