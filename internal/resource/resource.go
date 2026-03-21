package resource

import (
	"github.com/gin-gonic/gin"
	"github.com/muhaobing/std-go/restserver/registry"
	"github.com/muhaobing/std-go/restserver/scheduler"

	"wdkr-marketplace-service/internal/resource/healthy"
	"wdkr-marketplace-service/internal/resource/marketplace"
	"wdkr-marketplace-service/internal/resource/mock"
	"wdkr-marketplace-service/internal/resource/openapi"
	"wdkr-marketplace-service/internal/resource/ops"
)

// Resources 聚合所有 Resource
type Resources struct {
	Healthy     *healthy.HealthyResource
	Marketplace *marketplace.MarketplaceResource
	Ops         *ops.OpsResource
	OpenAPI     *openapi.OpenAPIResource
	Mock        *mock.MockResource
	Schedulers  []scheduler.Scheduler
}

// NewResources 构建 Resources
func NewResources(
	healthy *healthy.HealthyResource,
	marketplace *marketplace.MarketplaceResource,
	ops *ops.OpsResource,
	openapi *openapi.OpenAPIResource,
	schedulers ...scheduler.Scheduler,
) *Resources {
	return &Resources{
		Healthy:     healthy,
		Marketplace: marketplace,
		Ops:         ops,
		OpenAPI:     openapi,
		Schedulers:  schedulers,
	}
}

func (r *Resources) Router() registry.Registry {
	return func(engine *gin.Engine) {
		r.Healthy.Router()(engine)
		r.Marketplace.Router()(engine)
		r.Ops.Router()(engine)
		r.OpenAPI.Router()(engine)
		if r.Mock != nil {
			r.Mock.Router()(engine)
		}
	}
}

func (r *Resources) Registries() []registry.Registry {
	return []registry.Registry{
		r.Healthy.Router(),
		r.Marketplace.Router(),
		r.Ops.Router(),
		r.OpenAPI.Router(),
	}
}
