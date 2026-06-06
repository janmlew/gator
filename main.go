package main

import (
	"fmt"
	"log"

	"github.com/janmlew/gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	if err := cfg.SetUser("jml"); err != nil {
		log.Fatalf("error setting user: %v", err)
	}

	cfg, err = config.Read()
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	fmt.Printf("%+v\n", cfg)
}
