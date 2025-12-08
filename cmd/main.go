package main

import (
	"log"

	"wdkr-marketplace-service/internal/resource/healthy"

	"github.com/muhaobing-eng/std-go/restserver"
	"github.com/muhaobing-eng/std-go/restserver/registry"
)

func main() {
	if err := restserver.Init(
		registry.RouterRegistry(
			healthy.NewHealthyResource(),
		),
		registry.MiddlewareRegistry(),
	); err != nil {
		log.Fatalf("init rest server failed: %v", err)
	}

	restserver.Run()
}
