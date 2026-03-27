package main

import (
	"fmt"

	"github.com/vikukumar/pushpaka/internal/config"
)

func main() {
	cfg := config.Load()
	fmt.Printf("ProjectsDir: %s\n", cfg.ProjectsDir)
	fmt.Printf("BuildsDir: %s\n", cfg.BuildsDir)
	fmt.Printf("DeploysDir: %s\n", cfg.DeploysDir)
}
