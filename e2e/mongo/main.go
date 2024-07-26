package main

import (
	"context"
	"log"
)

func main() {
	ctx := context.Background()
	if err := bulk(ctx); err != nil {
		log.Fatal(err)
	}
}
