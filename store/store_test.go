package main

import (
	"io"
	"strings"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "helloWorld"
	keyPath := CASPathTransformFunc(key)

	expectedFilename := "5395ebfd174b0a5617e6f409dfbb3e064e3fdf0a"
	expectedPath := "5395e/bfd17/4b0a5/617e6/f409d/fbb3e/064e3/fdf0a"

	if keyPath.Filename != expectedFilename {
		t.Errorf("CASPathTransformFunc failed: expected %s, got %s", expectedFilename, keyPath.Filename)
	}
	if keyPath.Pathname != expectedPath {
		t.Errorf("CASPathTransformFunc failed: expected %s, got %s", expectedPath, keyPath.Pathname)
	}
}

func TestStore(t *testing.T) {
	storeOptions := StoreOpts{
		CASPathTransformFunc,
	}
	store := NewStore(storeOptions)
	data := "I am the new file"
	src := strings.NewReader(data)
	if err := store.writeStream(data, src); err != nil {
		t.Fatalf("writeStream failed: %v\n", err)
	}

	r, err := store.Read(data)
	if err != nil {
		t.Fatalf("Read failed: %v\n", err)
	}
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != data {
		t.Fatalf("wrong data: expected %v, got %v\n", data, string(b))
	}
	if err := store.Delete(data); err != nil {
		t.Fatalf("Delete failed: %v\n", err)
	}
}
