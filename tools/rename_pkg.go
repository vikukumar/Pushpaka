package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	replacements := map[string]string{
		"\"github.com/vikukumar/pushpaka/internal/tunnel\"": "\"github.com/vikukumar/pushpaka/pkg/tunnel\"",
	}

	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "node_modules" || d.Name() == ".git" || d.Name() == "ui" {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".go") {
			return nil
		}

		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		modified := false
		newB := string(b)
		for oldStr, newStr := range replacements {
			if strings.Contains(newB, oldStr) {
				newB = strings.ReplaceAll(newB, oldStr, newStr)
				modified = true
			}
		}

		if modified {
			fmt.Println("Updated", path)
			return os.WriteFile(path, []byte(newB), 0644)
		}
		return nil
	})

	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
