package cart

import (
	"context"
	"errors"
	"fmt"

	"github.com/muhaobing/std-go/go-common/database"

	cartmodel "wdkr-marketplace-service/internal/domain/cart/cart_model"
	"wdkr-marketplace-service/internal/domain/cart/repo"
	"wdkr-marketplace-service/internal/domain/order"
	"wdkr-marketplace-service/internal/domain/sku"
	skumodel "wdkr-marketplace-service/internal/domain/sku/sku_model"
)

type cartServiceImpl struct {
	cartRepo     repo.CartRepo
	skuService   sku.SkuService
	orderService order.OrderService
}

// NewCartService 创建购物车服务实例
func NewCartService(cartRepo repo.CartRepo, skuService sku.SkuService, orderService order.OrderService) CartService {
	return &cartServiceImpl{
		cartRepo:     cartRepo,
		skuService:   skuService,
		orderService: orderService,
	}
}

// AddToCart 添加商品到购物车
func (s *cartServiceImpl) AddToCart(ctx context.Context, req *AddToCartRequest) (*cartmodel.CartItem, error) {
	// 参数校验
	if req.UserId == 0 {
		return nil, errors.New("user_id is required")
	}
	if req.SkuId == 0 {
		return nil, errors.New("sku_id is required")
	}
	if req.Quantity <= 0 {
		return nil, errors.New("quantity must be greater than 0")
	}

	// 检查商品是否存在且已上架
	skuInfo, err := s.skuService.GetSkuById(ctx, req.SkuId)
	if err != nil {
		return nil, fmt.Errorf("failed to get sku: %w", err)
	}
	if skuInfo == nil {
		return nil, errors.New("sku not found")
	}
	if !skuInfo.IsOnline() {
		return nil, errors.New("sku is not available")
	}

	// 检查是否已在购物车中
	existingItem, err := s.cartRepo.GetCartItem(ctx, req.UserId, req.SkuId)
	if err != nil {
		return nil, fmt.Errorf("failed to check cart item: %w", err)
	}

	if existingItem != nil {
		// 已存在，更新数量
		newQuantity := existingItem.Quantity + req.Quantity
		if err := s.cartRepo.UpdateCartItemQuantity(ctx, existingItem.Id, newQuantity); err != nil {
			return nil, fmt.Errorf("failed to update cart item: %w", err)
		}
		existingItem.Quantity = newQuantity
		return existingItem, nil
	}

	// 不存在，创建新项
	item := &cartmodel.CartItem{
		UserId:   req.UserId,
		SkuId:    req.SkuId,
		Quantity: req.Quantity,
	}
	if err := s.cartRepo.AddCartItem(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to add cart item: %w", err)
	}

	return item, nil
}

// RemoveFromCart 从购物车移除商品
func (s *cartServiceImpl) RemoveFromCart(ctx context.Context, req *RemoveFromCartRequest) error {
	if req.UserId == 0 {
		return errors.New("user_id is required")
	}
	if req.SkuId == 0 {
		return errors.New("sku_id is required")
	}

	return s.cartRepo.DeleteCartItemByUserAndSku(ctx, req.UserId, req.SkuId)
}

// UpdateCartItem 更新购物车商品数量
func (s *cartServiceImpl) UpdateCartItem(ctx context.Context, req *UpdateCartItemRequest) (*cartmodel.CartItem, error) {
	if req.UserId == 0 {
		return nil, errors.New("user_id is required")
	}
	if req.SkuId == 0 {
		return nil, errors.New("sku_id is required")
	}

	// 如果数量为0或负数，直接删除
	if req.Quantity <= 0 {
		if err := s.cartRepo.DeleteCartItemByUserAndSku(ctx, req.UserId, req.SkuId); err != nil {
			return nil, fmt.Errorf("failed to delete cart item: %w", err)
		}
		return nil, nil
	}

	// 检查商品是否在购物车中
	item, err := s.cartRepo.GetCartItem(ctx, req.UserId, req.SkuId)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart item: %w", err)
	}
	if item == nil {
		return nil, errors.New("cart item not found")
	}

	// 更新数量
	if err := s.cartRepo.UpdateCartItemQuantity(ctx, item.Id, req.Quantity); err != nil {
		return nil, fmt.Errorf("failed to update cart item: %w", err)
	}
	item.Quantity = req.Quantity

	return item, nil
}

// ClearCart 清空购物车
func (s *cartServiceImpl) ClearCart(ctx context.Context, userId uint64) error {
	if userId == 0 {
		return errors.New("user_id is required")
	}

	return s.cartRepo.ClearCart(ctx, userId)
}

// GetCartList 获取购物车列表
func (s *cartServiceImpl) GetCartList(ctx context.Context, userId uint64) (*CartListResponse, error) {
	if userId == 0 {
		return nil, errors.New("user_id is required")
	}

	// 获取购物车项
	items, err := s.cartRepo.GetCartItemsByUserId(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart items: %w", err)
	}

	if len(items) == 0 {
		return &CartListResponse{
			Items:      []*cartmodel.CartItemWithSku{},
			TotalCount: 0,
			TotalCost:  0,
		}, nil
	}

	// 获取商品ID列表
	skuIds := make([]uint64, 0, len(items))
	for _, item := range items {
		skuIds = append(skuIds, item.SkuId)
	}

	// 批量获取商品信息
	skuMap := make(map[uint64]*skumodel.Sku)
	for _, skuId := range skuIds {
		skuInfo, err := s.skuService.GetSkuById(ctx, skuId)
		if err != nil {
			continue
		}
		if skuInfo != nil {
			skuMap[skuId] = skuInfo
		}
	}

	// 组装结果
	result := make([]*cartmodel.CartItemWithSku, 0, len(items))
	var totalCost float32 = 0
	totalCount := 0

	for _, item := range items {
		skuInfo := skuMap[item.SkuId]
		if skuInfo == nil {
			continue
		}

		itemWithSku := &cartmodel.CartItemWithSku{
			CartItem:           *item,
			SkuCode:            skuInfo.SkuCode,
			SkuName:            skuInfo.SkuName,
			SkuAvatar:          skuInfo.SkuAvatar,
			Cost:               skuInfo.Cost,
			SkuStatus:          skuInfo.SkuStatus,
			DeliveryMethod:     skuInfo.DeliveryMethod,
			FulfillMode:        skuInfo.FulfillMode,
			FulfillEcoinAmount: skuInfo.FulfillEcoinAmount,
		}
		result = append(result, itemWithSku)

		// 只计算已上架商品的总价
		if skuInfo.IsOnline() {
			totalCost += skuInfo.Cost * float32(item.Quantity)
			totalCount += item.Quantity
		}
	}

	return &CartListResponse{
		Items:      result,
		TotalCount: totalCount,
		TotalCost:  totalCost,
	}, nil
}

// Checkout 购物车下单（下单并移除对应商品）
func (s *cartServiceImpl) Checkout(ctx context.Context, req *CheckoutRequest) (*order.CreateOrderResponse, error) {
	if req.UserId == 0 {
		return nil, errors.New("user_id is required")
	}
	if len(req.SkuIds) == 0 {
		return nil, errors.New("sku_ids is required")
	}
	if req.PayType == "" {
		return nil, errors.New("pay_type is required")
	}

	// 获取购物车中的商品
	items, err := s.cartRepo.GetCartItemsByUserId(ctx, req.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart items: %w", err)
	}

	// 过滤出要下单的商品
	skuIdSet := make(map[uint64]bool)
	for _, skuId := range req.SkuIds {
		skuIdSet[skuId] = true
	}

	skuItems := make([]*order.SkuOrderItem, 0)
	for _, item := range items {
		if skuIdSet[item.SkuId] {
			skuItems = append(skuItems, &order.SkuOrderItem{
				SkuId:    item.SkuId,
				Quantity: item.Quantity,
			})
		}
	}

	if len(skuItems) == 0 {
		return nil, errors.New("no valid items to checkout")
	}

	var resp *order.CreateOrderResponse

	// 在事务中执行下单和移除购物车
	err = database.Transaction(ctx, func(ctx context.Context) error {
		// 创建订单
		var orderErr error
		resp, orderErr = s.orderService.CreateOrder(ctx, &order.CreateOrderRequest{
			UserId:   req.UserId,
			SkuItems: skuItems,
			PayType:  req.PayType,
			Remark:   req.Remark,
		})
		if orderErr != nil {
			return fmt.Errorf("failed to create order: %w", orderErr)
		}

		// 从购物车移除已下单的商品
		if err := s.cartRepo.DeleteCartItemsByUserAndSkus(ctx, req.UserId, req.SkuIds); err != nil {
			return fmt.Errorf("failed to remove cart items: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}
