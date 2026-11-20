package bolt

import "testing"

func TestAddGet(t *testing.T) {
	if err := Add("bolt-test-key", []byte("value")); err != nil {
		t.Fatal(err)
	}
	got, err := Get("bolt-test-key")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "value" {
		t.Errorf("Get: got %q, want %q", got, "value")
	}
}

func TestGetMissingKey(t *testing.T) {
	if _, err := Get("bolt-test-missing-key"); err == nil {
		t.Error("Get of a missing key: want an error, got nil")
	}
}

func TestEmptyKey(t *testing.T) {
	if err := Add("", []byte("v")); err == nil {
		t.Error("Add with an empty key: want an error, got nil")
	}
	if _, err := Get(""); err == nil {
		t.Error("Get with an empty key: want an error, got nil")
	}
}
