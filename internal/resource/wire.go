//go:build wireinject
// +build wireinject

package resource

import (
	"github.com/google/wire"

	"wdkr-marketplace-service/internal/domain/cart"
	cartrepo "wdkr-marketplace-service/internal/domain/cart/repo"
	"wdkr-marketplace-service/internal/domain/ecoin"
	ecoinrepo "wdkr-marketplace-service/internal/domain/ecoin/repo"
	"wdkr-marketplace-service/internal/domain/order"
	orderrepo "wdkr-marketplace-service/internal/domain/order/repo"
	"wdkr-marketplace-service/internal/domain/payment"
	paymentrepo "wdkr-marketplace-service/internal/domain/payment/repo"
	"wdkr-marketplace-service/internal/domain/sku"
	skurepo "wdkr-marketplace-service/internal/domain/sku/repo"
	"wdkr-marketplace-service/internal/domain/user"
	userrepo "wdkr-marketplace-service/internal/domain/user/repo"
	"wdkr-marketplace-service/internal/resource/healthy"
	"wdkr-marketplace-service/internal/resource/marketplace"
	"wdkr-marketplace-service/internal/resource/openapi"
	"wdkr-marketplace-service/internal/resource/ops"
)

// RepoSet 提供所有 Repository 实例
var RepoSet = wire.NewSet(
	skurepo.NewSkuRepo,
	orderrepo.NewOrderRepo,
	paymentrepo.NewPaymentRepo,
	ecoinrepo.NewEcoinRepo,
	userrepo.NewUserRepo,
	userrepo.NewUserBindingRepo,
	cartrepo.NewCartRepo,
)

// PaymentSet 提供支付相关实例（支付渠道 + 支付服务）
var PaymentSet = wire.NewSet(
	payment.NewWechatPayConfig,
	payment.NewWechatPayChannel,
	payment.ProvidePaymentChannels,
	payment.ProvidePaymentService,
)

// ServiceSet 提供所有 Service 实例
var ServiceSet = wire.NewSet(
	ecoin.NewEcoinService,
	sku.NewSkuService,
	order.NewOrderService,
	user.NewUserService,
	cart.NewCartService,
)

// ResourceSet 提供所有 Resource 实例
var ResourceSet = wire.NewSet(
	healthy.NewHealthyResource,
	marketplace.NewMarketplaceResource,
	ops.NewOpsResource,
	openapi.NewOpenAPIResource,
)

// InitializeResources 初始化所有 Resources（wire injector）
func InitializeResources() *Resources {
	wire.Build(
		RepoSet,
		PaymentSet,
		ServiceSet,
		ResourceSet,
		NewResources,
	)
	return nil
}
