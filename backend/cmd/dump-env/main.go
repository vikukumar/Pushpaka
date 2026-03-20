package main

import (
	"fmt"
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type EnvVar struct {
	ProjectID string `gorm:"index"`
	Key       string `gorm:"type:varchar(255)"`
	Value     string `gorm:"type:text"`
}

func main() {
	db, err := gorm.Open(sqlite.Open("pushpaka-dev.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	var envVars []EnvVar
	if err := db.Where("project_id = ?", "3dbef396-9a86-4d1d-a8d7-f85069bad347").Find(&envVars).Error; err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Env Vars for 3dbef396:\n")
	for _, ev := range envVars {
		fmt.Printf("%s=%s\n", ev.Key, ev.Value)
	}
}
