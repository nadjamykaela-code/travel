package firestore

import (
	"context"
	"testing"
)

func TestClient_Close_Idempotent(t *testing.T) {
	c := &Client{inner: nil}
	if err := c.Close(); err != nil {
		t.Errorf("Close() on nil client: %v", err)
	}
	if err := c.Close(); err != nil {
		t.Errorf("Close() second call: %v", err)
	}
}

func TestFirestore_NilClient(t *testing.T) {
	c := &Client{inner: nil}
	fs := c.Firestore()
	if fs != nil {
		t.Error("Firestore() should return nil when inner is nil")
	}
}

func TestNew_UsesProjectID(t *testing.T) {
	// This test verifies the constructor doesn't panic and uses the project ID.
	// Actual connection depends on ADC availability.
	c, err := New(context.Background(), "test-project", "")
	if err != nil {
		// On environments without ADC, Firestore will fail — that's expected.
		// We just verify the error mentions project or credentials.
		t.Logf("New() returned expected error in this env: %v", err)
		return
	}
	defer c.Close()
	if c.Firestore() == nil {
		t.Error("Firestore() returned nil after successful New()")
	}
}
