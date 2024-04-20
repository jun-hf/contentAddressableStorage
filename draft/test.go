package main

import (
	"fmt"
	"os"
	"path/filepath"
)



func deleteFullPath(path string) error {
	if string(path[0]) == "/" {
		path = "." + path
	}
	for {
		if path == "." {
			return nil
		}
		if err := os.RemoveAll(path); err != nil {
			return err
		}
		path = filepath.Dir(path)
	}
}
func main() {
	fmt.Println(deleteFullPath("/hello/ad"))
}