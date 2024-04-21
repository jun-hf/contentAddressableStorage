package store

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type KeyPath struct {
	Pathname, Filename string
}

func (k KeyPath) FullPath() string {
	return fmt.Sprintf("%s/%s", k.Pathname, k.Filename)
}

type TransformPathFunc func(string) KeyPath

func DefaultPathTransformFunc(key string) KeyPath {
	return KeyPath{Pathname: key, Filename: key}
}

type StoreOpts struct {
	TransformPathFunc TransformPathFunc
	Root              string
}

type Store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	if len(opts.Root) == 0 {
		opts.Root = defaultRoot
	}
	if opts.TransformPathFunc == nil {
		opts.TransformPathFunc = DefaultPathTransformFunc
	}
	return &Store{StoreOpts: opts}
}

func (s *Store) Has(key string) bool {
	pathKey := s.TransformPathFunc(key)
	filePath := s.BuildStoreFullFilePath(pathKey)

	_, err := os.Stat(filePath)
	return !errors.Is(err, os.ErrNotExist)
}

func (s *Store) Read(key string) (io.Reader, error) {
	keyPath := s.TransformPathFunc(key)
	filePath := s.BuildStoreFullFilePath(keyPath)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}

func (s *Store) Delete(key string) error {
	keyPath := s.TransformPathFunc(key)
	filePath := s.BuildStoreFullFilePath(keyPath)
	return s.deleteFullPath(filePath)
}

func (s *Store) ClearAll() error {
	return os.RemoveAll(s.Root)
}

func (s *Store) BuildStoreFullFilePath(k KeyPath) string {
	return filepath.Join(s.Root, k.Pathname, k.Filename)
}

func (s *Store) BuildStoreFullPath(k KeyPath) string {
	return filepath.Join(s.Root, k.Pathname)
}

func (s *Store) deleteFullPath(path string) error {
	if string(path[0]) == "/" {
		path = "." + path
	}
	for {
		if path == s.Root {
			return nil
		}
		if _, err := os.Stat(path); err != nil {
			return err
		}
		if err := os.RemoveAll(path); err != nil {
			return err
		}
		path = filepath.Dir(path)
	}
}

func (s *Store) writeStream(key string, r io.Reader) error {
	keyPath := s.TransformPathFunc(key)
	path := s.BuildStoreFullPath(keyPath)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	filename := s.BuildStoreFullFilePath(keyPath)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	n, err := io.Copy(file, r)
	if err != nil {
		return err
	}
	fmt.Printf("Written [%d] to file\n", n)
	return nil
}

var defaultRoot = "fileStorage"

func CASPathTransformFunc(key string) KeyPath {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	blockSize := 5
	numBlock := len(hashStr) / blockSize
	paths := make([]string, numBlock)

	for i := 0; i < numBlock; i++ {
		from, to := i*blockSize, i*blockSize+blockSize
		paths[i] = hashStr[from:to]
	}
	return KeyPath{
		Pathname: strings.Join(paths, "/"),
		Filename: hashStr,
	}
}
