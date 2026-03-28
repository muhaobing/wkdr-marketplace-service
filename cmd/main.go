package main

import (
	"log"

	"github.com/muhaobing/std-go/restserver"
	"github.com/muhaobing/std-go/restserver/handler"
	"github.com/muhaobing/std-go/restserver/middleware/cache"
	"github.com/muhaobing/std-go/restserver/middleware/database"
	"github.com/muhaobing/std-go/restserver/registry"

	"wdkr-marketplace-service/internal/common/config"
	"wdkr-marketplace-service/internal/resource"
	"wdkr-marketplace-service/middleware"
)

func main() {
	// 0. init config
	if err := config.Init(); err != nil {
		log.Fatalf("init config failed: %v", err)
	}

	// 1. register rest server handler
	handler.RegisterHandler(&database.DatabaseHandler{})
	handler.RegisterHandler(&cache.CacheHandler{})
	handler.RegisterHandler(&middleware.RecoveryHandler{})
	handler.RegisterHandler(&middleware.AuthValidationHandler{})
	handler.RegisterHandler(&middleware.JWTValidationHandler{})

	// 2. init resources
	resources := resource.InitializeResources()

	// 3. init rest server (middleware + routes)
	if err := restserver.Init(
		registry.MiddlewareRegistry(
			database.DatabaseHandlerKey,
			cache.CacheHandlerKey,
			middleware.RecoveryHandlerKey,
			middleware.AuthValidationHandlerKey,
			middleware.JWTValidationHandlerKey,
		),
		registry.RouterRegistry(resources),
	); err != nil {
		log.Fatalf("init rest server failed: %v", err)
	}

	// 4. register schedulers
	for _, s := range resources.Schedulers {
		restserver.RegisterScheduler(s)
	}

	// 5. run the rest server (auto starts scheduler)
	restserver.Run()
}
