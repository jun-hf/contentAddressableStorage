package store

import (
	"fmt"
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
	store := createStore()
	for i := 0; i < 88; i++ {
		data := fmt.Sprintf("The data is %v", i)
		src := strings.NewReader(data)
		if _, err := store.Write(data, src); err != nil {
			t.Fatalf("Write failed: %v\n", err)
		}
	
		r, err := store.Read(data)
		if err != nil {
			t.Fatalf("Read failed: %v\n", err)
		}
	
		b, _ := io.ReadAll(r)
		if string(b) != data {
			t.Fatalf("wrong data: expected %v, got %v\n", data, string(b))
		}
	
		if ok := store.Has(data); !ok {
			t.Fatalf("Has failed key should exists\n")
		}
		
		if err := store.Delete(data); err != nil {
			t.Fatalf("Delete failed: %v\n", err)
		}
		
		if ok := store.Has(data); ok {
			t.Fatalf("Data should not existed: %v\n", data)
		}
	}
	tearDown(t, store)
}


func tearDown(t *testing.T, s *Store) {
	if err := s.ClearAll(); err != nil {
		t.Fatalf("ClearAll failed: %v\n", err)
	}
}

func createStore() *Store {
	storeOptions := StoreOpts{
		CASPathTransformFunc,
		"testDir",
	}
	return NewStore(storeOptions)
}