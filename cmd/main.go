package main

import (
	"log"

	"github.com/muhaobing-eng/std-go/restserver"
	"github.com/muhaobing-eng/std-go/restserver/handler"
	"github.com/muhaobing-eng/std-go/restserver/middleware/cache"
	"github.com/muhaobing-eng/std-go/restserver/middleware/database"
	"github.com/muhaobing-eng/std-go/restserver/registry"

	"wdkr-marketplace-service/bootstrap"
	"wdkr-marketplace-service/internal/common/config"
	"wdkr-marketplace-service/internal/resource"
	"wdkr-marketplace-service/middleware"
)

func main() {
	// 0. set environments
	if err := config.Init(); err != nil {
		log.Fatalf("init config failed: %v", err)
	}

	// 1. register rest server handler
	handler.RegisterHandler(&database.DatabaseHandler{})
	handler.RegisterHandler(&cache.CacheHandler{})
	handler.RegisterHandler(&middleware.RecoveryHandler{})
	handler.RegisterHandler(&middleware.AuthValidationHandler{})

	// 2. init rest server
	if err := restserver.Init(
		registry.MiddlewareRegistry(
			database.DatabaseHandlerKey,
			cache.CacheHandlerKey,
			middleware.RecoveryHandlerKey,
			middleware.AuthValidationHandlerKey,
		),
		registry.RouterRegistry(
			resource.InitializeResources(),
		),
	); err != nil {
		log.Fatalf("init rest server failed: %v", err)
		return
	}

	// 3. init rest server dependencies
	if err := bootstrap.StartUp(); err != nil {
		log.Fatalf("start bootstrap failed: %v", err)
		return
	}

	// 4. run the rest server
	restserver.Run()
}
