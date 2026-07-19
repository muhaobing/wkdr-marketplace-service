package order

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/muhaobing/std-go/go-common/database"

	"wdkr-marketplace-service/bootstrap"
	"wdkr-marketplace-service/internal/common/config"
	"wdkr-marketplace-service/internal/common/utils"
	"wdkr-marketplace-service/internal/domain/companyecoin"
	"wdkr-marketplace-service/internal/domain/ecoin"
	ecoin_model "wdkr-marketplace-service/internal/domain/ecoin/ecoin_model"
	ordermodel "wdkr-marketplace-service/internal/domain/order/order_model"
	"wdkr-marketplace-service/internal/domain/order/repo"
	"wdkr-marketplace-service/internal/domain/payment"
	"wdkr-marketplace-service/internal/domain/payment/payment_model"
	"wdkr-marketplace-service/internal/domain/sku"
	skumodel "wdkr-marketplace-service/internal/domain/sku/sku_model"
)

const (
	fulfillLockPrefix = "lock:fulfill:"
	fulfillLockTTL    = 60 * time.Second
)

// orderServiceImpl 订单服务实现
type orderServiceImpl struct {
	orderRepo       repo.OrderRepo
	skuService      sku.SkuService
	ecoinSvc        ecoin.EcoinService
	companyEcoinSvc companyecoin.CompanyEcoinService
	paymentSvc      payment.PaymentService
	userSvc         UserServiceForOrder
}

// UserServiceForOrder 订单服务所需的用户服务接口（避免直接依赖 user 包）
type UserServiceForOrder interface {
	GetBindingsByUserId(ctx context.Context, userId uint) ([]UserBinding, error)
	GetUserCompanyId(ctx context.Context, userId uint) (uint64, error)
}

// UserBinding 用于履约时查找业务用户ID
type UserBinding struct {
	BizCode   string
	BizUserId uint64
}

// NewOrderService 创建订单服务实例
func NewOrderService(
	orderRepo repo.OrderRepo,
	skuService sku.SkuService,
	ecoinSvc ecoin.EcoinService,
	companyEcoinSvc companyecoin.CompanyEcoinService,
	paymentSvc payment.PaymentService,
	userSvc UserServiceForOrder,
) OrderService {
	return &orderServiceImpl{
		orderRepo:       orderRepo,
		skuService:      skuService,
		ecoinSvc:        ecoinSvc,
		companyEcoinSvc: companyEcoinSvc,
		paymentSvc:      paymentSvc,
		userSvc:         userSvc,
	}
}

func (s *orderServiceImpl) deductEcoinForOrder(ctx context.Context, userId uint64, req *ecoin.DeductEcoinRequest) error {
	cid, err := s.userSvc.GetUserCompanyId(ctx, uint(userId))
	if err != nil {
		return err
	}
	if cid == 0 {
		_, err := s.ecoinSvc.DeductEcoin(ctx, req)
		return err
	}
	_, err = s.companyEcoinSvc.DeductCompanyEcoin(ctx, &companyecoin.DeductCompanyEcoinRequest{
		CompanyId:      cid,
		OperatorUserId: userId,
		Amount:         req.Amount,
		SourceType:     req.SourceType,
		SourceId:       req.SourceId,
		Description:    req.Description,
	})
	return err
}

func (s *orderServiceImpl) addEcoinForOrder(ctx context.Context, userId uint64, req *ecoin.AddEcoinRequest) error {
	cid, err := s.userSvc.GetUserCompanyId(ctx, uint(userId))
	if err != nil {
		return err
	}
	if cid == 0 {
		_, err := s.ecoinSvc.AddEcoin(ctx, req)
		return err
	}
	_, err = s.companyEcoinSvc.AddCompanyEcoin(ctx, &companyecoin.AddCompanyEcoinRequest{
		CompanyId:      cid,
		OperatorUserId: userId,
		Amount:         req.Amount,
		SourceType:     req.SourceType,
		SourceId:       req.SourceId,
		Description:    req.Description,
	})
	return err
}

// CreateOrder 创建订单
func (s *orderServiceImpl) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error) {
	// 参数校验
	if req.UserId == 0 {
		return nil, errors.New("user_id is required")
	}
	if req.PayType == "" {
		return nil, errors.New("pay_type is required")
	}
	if req.PayType != ordermodel.PayTypeEcoin && req.PayType != ordermodel.PayTypeMoney {
		return nil, errors.New("invalid pay_type, must be ecoin or money")
	}
	if !req.IsEcoinRecharge && len(req.SkuItems) == 0 {
		return nil, errors.New("at least one sku_item is required")
	}
	if req.IsEcoinRecharge && req.EcoinUnits <= 0 {
		return nil, errors.New("ecoin stock should be greater than zero")
	}
	if req.IsEcoinRecharge && req.PayType == ordermodel.PayTypeEcoin {
		return nil, errors.New("ecoin recharge not support ecoin pay type")
	}

	// 构建订单和明细
	order, orderItems, err := s.buildOrder(ctx, req)
	if err != nil {
		return nil, err
	}

	// 金币支付需要立即扣除金币
	if req.PayType == ordermodel.PayTypeEcoin {
		err := database.Transaction(ctx, func(ctx context.Context) error {
			if err := s.orderRepo.CreateOrder(ctx, order); err != nil {
				return fmt.Errorf("failed to create order: %w", err)
			}

			for _, item := range orderItems {
				item.OrderId = order.Id
			}
			if err := s.orderRepo.CreateOrderItems(ctx, orderItems); err != nil {
				return fmt.Errorf("failed to create order items: %w", err)
			}

			ecoinAmount := float64(order.PayAmount) / float64(config.GetConf().EcoinUnitPrice)
			err := s.deductEcoinForOrder(ctx, req.UserId, &ecoin.DeductEcoinRequest{
				UserId:      req.UserId,
				Amount:      ecoinAmount,
				SourceType:  "order",
				SourceId:    order.OrderNo,
				Description: fmt.Sprintf("购买商品 - 订单号: %s", order.OrderNo),
			})
			if err != nil {
				return fmt.Errorf("failed to deduct ecoin: %w", err)
			}

			if err := s.orderRepo.UpdateOrderEcoinAmount(ctx, order.OrderNo, ecoinAmount); err != nil {
				return fmt.Errorf("failed to update ecoin amount: %w", err)
			}

			payTime := uint32(time.Now().Unix())
			if err := s.orderRepo.UpdateOrderToPaid(ctx, order.OrderNo, payTime); err != nil {
				return fmt.Errorf("failed to update order to paid: %w", err)
			}
			order.Status = ordermodel.OrderStatusPaid
			order.PayTime = payTime
			order.EcoinAmount = ecoinAmount

			return nil
		})
		if err != nil {
			return nil, err
		}

		go s.asyncAutoFulfill(order.OrderNo)
	} else {
		// 货币支付创建待支付订单
		err := database.Transaction(ctx, func(ctx context.Context) error {
			// 创建订单
			if err := s.orderRepo.CreateOrder(ctx, order); err != nil {
				return fmt.Errorf("failed to create order: %w", err)
			}

			// 创建订单明细
			for _, item := range orderItems {
				item.OrderId = order.Id
			}
			if err := s.orderRepo.CreateOrderItems(ctx, orderItems); err != nil {
				return fmt.Errorf("failed to create order items: %w", err)
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	order.Items = orderItems
	return &CreateOrderResponse{Order: order}, nil
}

// buildOrder 构建订单和订单明细
func (s *orderServiceImpl) buildOrder(ctx context.Context, req *CreateOrderRequest) (*ordermodel.Order, []*ordermodel.OrderItem, error) {
	orderNo := s.generateOrderNo()
	var orderItems []*ordermodel.OrderItem
	var totalAmount float32
	var totalQuantity int

	if req.IsEcoinRecharge {
		unitPrice := config.GetConf().EcoinUnitPrice
		totalAmount = float32(req.EcoinUnits) * unitPrice
		totalQuantity = req.EcoinUnits

		// 校验支付金额不低于1分钱
		if int64(totalAmount*100) < 1 {
			return nil, nil, fmt.Errorf("充值金额过低，最低支付金额为1分钱，当前金币单价为%.4f元，请至少充值%d金币",
				unitPrice, int(math.Ceil(0.01/float64(unitPrice))))
		}

		orderItem := &ordermodel.OrderItem{
			OrderNo:       orderNo,
			SkuId:         0,
			SkuCode:       "ecoin",
			SkuName:       "金币",
			SkuAvatar:     "",
			Quantity:      req.EcoinUnits,
			UnitPrice:     unitPrice,
			TotalPrice:    totalAmount,
			FulfillStatus: ordermodel.FulfillStatusPending,
		}
		orderItems = append(orderItems, orderItem)
	} else {
		companyId, err := s.userSvc.GetUserCompanyId(ctx, uint(req.UserId))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to resolve user company: %w", err)
		}

		// 处理 SKU 单品
		for _, skuItem := range req.SkuItems {
			if skuItem.SkuId == 0 {
				return nil, nil, errors.New("sku_id is required in sku_items")
			}
			if skuItem.Quantity <= 0 {
				skuItem.Quantity = 1
			}

			// 获取SKU信息
			skuInfo, err := s.skuService.GetSkuById(ctx, skuItem.SkuId)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to get sku %d: %w", skuItem.SkuId, err)
			}

			// 验证SKU状态
			if !skuInfo.IsOnline() {
				return nil, nil, fmt.Errorf("sku %d is not available", skuItem.SkuId)
			}

			if skuInfo.SkuScope > skumodel.SkuScopeEnterprise {
				skuInfo.SkuScope = skumodel.SkuScopeUniversal
			}
			if !skuInfo.MatchesUserSkuScope(companyId) {
				if skuInfo.SkuScope == skumodel.SkuScopeEnterprise {
					return nil, nil, errors.New("本商品仅限企业账号购买")
				}
				return nil, nil, errors.New("本商品仅限个人账号购买")
			}

			if req.PayType == ordermodel.PayTypeEcoin && skuInfo.IsEcoinGrantFulfill() {
				return nil, nil, errors.New("金币类商品不支持金币支付，请使用在线支付")
			}

			// 计算价格
			itemTotal := skuInfo.Cost * float32(skuItem.Quantity)
			totalAmount += itemTotal
			totalQuantity += skuItem.Quantity

			// 构建订单明细
			orderItem := &ordermodel.OrderItem{
				OrderNo:       orderNo,
				SkuId:         skuInfo.Id,
				SkuCode:       skuInfo.SkuCode,
				SkuName:       skuInfo.SkuName,
				SkuAvatar:     skuInfo.SkuAvatar,
				Quantity:      skuItem.Quantity,
				UnitPrice:     skuInfo.Cost,
				TotalPrice:    itemTotal,
				FulfillStatus: ordermodel.FulfillStatusPending,
			}
			orderItems = append(orderItems, orderItem)
		}
	}

	if len(orderItems) == 0 {
		return nil, nil, errors.New("no valid items in order")
	}

	// 构建订单
	order := &ordermodel.Order{
		OrderNo:        orderNo,
		UserId:         req.UserId,
		ItemCount:      len(req.SkuItems),
		TotalQuantity:  totalQuantity,
		OriginalAmount: totalAmount,
		PayAmount:      totalAmount,
		PayType:        req.PayType,
		Status:         ordermodel.OrderStatusPending,
		Remark:         req.Remark,
	}

	return order, orderItems, nil
}

// GetOrderByOrderNo 根据订单号获取订单详情
func (s *orderServiceImpl) GetOrderByOrderNo(ctx context.Context, orderNo string) (*ordermodel.Order, error) {
	if orderNo == "" {
		return nil, errors.New("order_no is required")
	}

	order, err := s.orderRepo.GetOrderByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	if order == nil {
		return nil, errors.New("order not found")
	}

	// 获取订单明细
	items, err := s.orderRepo.GetOrderItemsByOrderId(ctx, order.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	order.Items = items

	return order, nil
}

// PayOrder 支付订单
func (s *orderServiceImpl) PayOrder(ctx context.Context, req *PayOrderRequest) (*PayOrderResponse, error) {
	if req.OrderNo == "" {
		return nil, errors.New("order_no is required")
	}
	if req.Channel == "" {
		return nil, errors.New("channel is required")
	}

	// 获取订单
	order, err := s.orderRepo.GetOrderByOrderNo(ctx, req.OrderNo)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	if order == nil {
		return nil, errors.New("order not found")
	}

	// 验证订单状态
	if !order.CanPay() {
		return nil, fmt.Errorf("order cannot be paid, current status: %d", order.Status)
	}

	// 金币支付
	if req.Channel == "ecoin" {
		return s.payWithEcoin(ctx, order)
	}

	// 货币支付
	if req.PayMethod == "" {
		return nil, errors.New("pay_method is required for non-ecoin payment")
	}

	// 创建支付订单（复用已有的 pending 支付单由 payment service 处理）
	paymentResp, err := s.paymentSvc.CreatePayment(ctx, &payment.CreatePaymentRequest{
		BizOrderNo:    req.OrderNo,
		BizType:       payment_model.BizTypePurchase,
		UserId:        order.UserId,
		Channel:       req.Channel,
		PayMethod:     req.PayMethod,
		Amount:        int64(order.PayAmount * 100),
		Description:   fmt.Sprintf("商城订单 - %s", req.OrderNo),
		ClientIP:      req.ClientIP,
		OpenId:        req.OpenId,
		ExpireMinutes: 15,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// 更新订单的支付类型和支付订单号
	if order.PayType != ordermodel.PayTypeMoney {
		_ = s.orderRepo.UpdateOrderPayType(ctx, req.OrderNo, ordermodel.PayTypeMoney)
	}
	if err := s.orderRepo.UpdateOrderPaymentOrderNo(ctx, req.OrderNo, paymentResp.OrderNo); err != nil {
		return nil, fmt.Errorf("failed to update order payment order no: %w", err)
	}

	return &PayOrderResponse{
		OrderNo:   req.OrderNo,
		CodeUrl:   paymentResp.CodeUrl,
		H5Url:     paymentResp.H5Url,
		PrepayId:  paymentResp.PrepayId,
		AppId:     paymentResp.AppId,
		TimeStamp: paymentResp.TimeStamp,
		NonceStr:  paymentResp.NonceStr,
		Package:   paymentResp.Package,
		SignType:  paymentResp.SignType,
		PaySign:   paymentResp.PaySign,
	}, nil
}

// payWithEcoin 金币支付
func (s *orderServiceImpl) payWithEcoin(ctx context.Context, order *ordermodel.Order) (*PayOrderResponse, error) {
	items, err := s.orderRepo.GetOrderItemsByOrderId(ctx, order.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	for _, it := range items {
		if it.SkuId == 0 {
			continue
		}
		skuInfo, err := s.skuService.GetSkuById(ctx, it.SkuId)
		if err != nil {
			return nil, fmt.Errorf("failed to get sku %d: %w", it.SkuId, err)
		}
		if skuInfo == nil {
			return nil, fmt.Errorf("sku %d not found", it.SkuId)
		}
		if skuInfo.IsEcoinGrantFulfill() {
			return nil, errors.New("金币类商品不支持金币支付，请使用在线支付")
		}
	}

	err = database.Transaction(ctx, func(ctx context.Context) error {
		ecoinAmount := float64(order.PayAmount) / float64(config.GetConf().EcoinUnitPrice)
		err := s.deductEcoinForOrder(ctx, order.UserId, &ecoin.DeductEcoinRequest{
			UserId:      order.UserId,
			Amount:      ecoinAmount,
			SourceType:  "order",
			SourceId:    order.OrderNo,
			Description: fmt.Sprintf("购买商品 - 订单号: %s", order.OrderNo),
		})
		if err != nil {
			return fmt.Errorf("failed to deduct ecoin: %w", err)
		}

		if order.PayType != ordermodel.PayTypeEcoin {
			if err := s.orderRepo.UpdateOrderPayType(ctx, order.OrderNo, ordermodel.PayTypeEcoin); err != nil {
				return fmt.Errorf("failed to update order pay type: %w", err)
			}
		}

		if err := s.orderRepo.UpdateOrderEcoinAmount(ctx, order.OrderNo, ecoinAmount); err != nil {
			return fmt.Errorf("failed to update ecoin amount: %w", err)
		}

		payTime := uint32(time.Now().Unix())
		if err := s.orderRepo.UpdateOrderToPaid(ctx, order.OrderNo, payTime); err != nil {
			return fmt.Errorf("failed to update order to paid: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	go s.asyncAutoFulfill(order.OrderNo)

	return &PayOrderResponse{OrderNo: order.OrderNo}, nil
}

// CancelOrder 取消订单
func (s *orderServiceImpl) CancelOrder(ctx context.Context, req *CancelOrderRequest) error {
	if req.OrderNo == "" {
		return errors.New("order_no is required")
	}

	return database.Transaction(ctx, func(ctx context.Context) error {
		// 获取订单（加锁）
		order, err := s.orderRepo.GetOrderForUpdate(ctx, req.OrderNo)
		if err != nil {
			return fmt.Errorf("failed to get order: %w", err)
		}
		if order == nil {
			return errors.New("order not found")
		}

		// 验证是否可以取消
		if !order.CanCancel() {
			return fmt.Errorf("order cannot be cancelled, current status: %d", order.Status)
		}

		// 如果有支付订单，关闭支付订单
		if order.PaymentOrderNo != "" {
			if err := s.paymentSvc.ClosePayment(ctx, order.PaymentOrderNo); err != nil {
				// 关闭失败不阻塞取消流程
			}
		}

		// 更新订单为已取消
		cancelTime := uint32(time.Now().Unix())
		if err := s.orderRepo.UpdateOrderToCancelled(ctx, req.OrderNo, cancelTime, req.Reason); err != nil {
			return fmt.Errorf("failed to cancel order: %w", err)
		}

		return nil
	})
}

// HandlePaymentSuccess 处理支付成功：更新订单状态 + 自动履约
func (s *orderServiceImpl) HandlePaymentSuccess(ctx context.Context, orderNo string, payTime uint32) error {
	if orderNo == "" {
		return errors.New("order_no is required")
	}

	err := database.Transaction(ctx, func(ctx context.Context) error {
		order, err := s.orderRepo.GetOrderForUpdate(ctx, orderNo)
		if err != nil {
			return fmt.Errorf("failed to get order: %w", err)
		}
		if order == nil {
			return errors.New("order not found")
		}

		if !order.IsPending() {
			return nil
		}

		if err := s.orderRepo.UpdateOrderToPaid(ctx, orderNo, payTime); err != nil {
			return fmt.Errorf("failed to update order to paid: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	go s.asyncAutoFulfill(orderNo)

	return nil
}

// FulfillOrder 手动履约订单（运营后台调用）
func (s *orderServiceImpl) FulfillOrder(ctx context.Context, orderNo string, bizUserId string) error {
	if orderNo == "" {
		return errors.New("order_no is required")
	}
	if bizUserId == "" {
		return errors.New("biz_user_id is required")
	}
	return s.doFulfill(ctx, orderNo, bizUserId)
}

// AutoFulfill 自动履约（获取分布式锁 + 执行履约）
// 所有自动履约入口统一调用此方法：微信回调、金币支付、定时任务、前端同步
func (s *orderServiceImpl) AutoFulfill(ctx context.Context, orderNo string) error {
	lock, err := utils.AcquireDistributedLock(ctx, utils.GenKey(":", fulfillLockPrefix, orderNo), fulfillLockTTL)
	if err != nil {
		return err
	}
	if lock == nil {
		return nil
	}
	defer func() {
		_ = lock.Release(ctx)
	}()

	return s.autoFulfill(ctx, orderNo)
}

// asyncAutoFulfill 异步自动履约（构建后台 context 后调用 AutoFulfill）
func (s *orderServiceImpl) asyncAutoFulfill(orderNo string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[WARN] asyncAutoFulfill panic for order %s: %v\n", orderNo, r)
		}
	}()

	ctx := bootstrap.BackgroundContext()
	if err := s.AutoFulfill(ctx, orderNo); err != nil {
		fmt.Printf("[WARN] auto fulfill failed for order %s: %v\n", orderNo, err)
	}
}

// autoFulfill 支付成功后自动履约
func (s *orderServiceImpl) autoFulfill(ctx context.Context, orderNo string) error {
	order, err := s.orderRepo.GetOrderByOrderNo(ctx, orderNo)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}
	if order == nil {
		return errors.New("order not found")
	}
	if !order.IsPaid() {
		return nil
	}

	items, err := s.orderRepo.GetOrderItemsByOrderId(ctx, order.Id)
	if err != nil {
		return fmt.Errorf("failed to get order items: %w", err)
	}

	// 判断是否为金币充值订单
	if s.isEcoinRechargeOrder(items) {
		return s.fulfillEcoinRecharge(ctx, order, items)
	}

	// 普通商品订单：查找用户绑定信息后履约
	bizUserIdMap := s.buildBizUserIdMap(ctx, order.UserId)
	return s.fulfillSkuItems(ctx, order, items, bizUserIdMap, "")
}

// isEcoinRechargeOrder 判断是否为金币充值订单
func (s *orderServiceImpl) isEcoinRechargeOrder(items []*ordermodel.OrderItem) bool {
	return len(items) == 1 && items[0].SkuId == 0 && items[0].SkuCode == "ecoin"
}

// fulfillEcoinRecharge 金币充值履约：给用户增加金币
func (s *orderServiceImpl) fulfillEcoinRecharge(ctx context.Context, order *ordermodel.Order, items []*ordermodel.OrderItem) error {
	item := items[0]
	if !item.IsFulfillPending() {
		return nil
	}

	err := s.addEcoinForOrder(ctx, order.UserId, &ecoin.AddEcoinRequest{
		UserId:      order.UserId,
		Amount:      float64(item.Quantity),
		SourceType:  "recharge",
		SourceId:    order.OrderNo,
		Description: fmt.Sprintf("金币充值 - 订单号: %s", order.OrderNo),
	})

	fulfillTime := uint32(time.Now().Unix())
	if err != nil {
		_ = s.orderRepo.UpdateOrderItemFulfillStatus(ctx, item.Id, ordermodel.FulfillStatusFailed, fulfillTime, err.Error())
		return fmt.Errorf("failed to add ecoin: %w", err)
	}

	if err := s.orderRepo.UpdateOrderItemFulfillStatus(ctx, item.Id, ordermodel.FulfillStatusSuccess, fulfillTime, "金币充值成功"); err != nil {
		return err
	}
	return s.orderRepo.UpdateOrderToFulfilled(ctx, order.OrderNo, fulfillTime)
}

// buildBizUserIdMap 构建 bizCode -> bizUserId 映射
func (s *orderServiceImpl) buildBizUserIdMap(ctx context.Context, userId uint64) map[string]string {
	result := make(map[string]string)
	if s.userSvc == nil {
		return result
	}

	bindings, err := s.userSvc.GetBindingsByUserId(ctx, uint(userId))
	if err != nil {
		fmt.Printf("[WARN] failed to get user bindings for userId %d: %v\n", userId, err)
		return result
	}
	for _, b := range bindings {
		result[b.BizCode] = fmt.Sprintf("%d", b.BizUserId)
	}
	return result
}

// fulfillSkuItems 普通商品履约
// 重试策略：自动履约会处理 pending + failed 明细，便于定时任务对失败项持续重试。
func (s *orderServiceImpl) fulfillSkuItems(ctx context.Context, order *ordermodel.Order, items []*ordermodel.OrderItem, bizUserIdMap map[string]string, _ string) error {
	retryItems := make([]*ordermodel.OrderItem, 0)
	for _, item := range items {
		if item.IsFulfillPending() || item.IsFulfillFailed() {
			retryItems = append(retryItems, item)
		}
	}
	if len(retryItems) == 0 {
		if allOrderItemsFulfilledSuccessfully(items) {
			fulfillTime := uint32(time.Now().Unix())
			return s.orderRepo.UpdateOrderToFulfilled(ctx, order.OrderNo, fulfillTime)
		}
		return nil
	}

	for _, item := range retryItems {
		skuInfo, err := s.skuService.GetSkuById(ctx, item.SkuId)

		fulfillTime := uint32(time.Now().Unix())
		var fulfillStatus uint8
		var fulfillMsg string

		if err != nil {
			fulfillStatus = ordermodel.FulfillStatusFailed
			fulfillMsg = fmt.Sprintf("get sku failed: %v", err)
		} else if skuInfo.IsEcoinGrantFulfill() {
			grant := skuInfo.FulfillEcoinAmount * float64(item.Quantity)
			if grant <= 0 {
				fulfillStatus = ordermodel.FulfillStatusFailed
				fulfillMsg = "未配置金币发放数量"
			} else {
				addErr := s.addEcoinForOrder(ctx, order.UserId, &ecoin.AddEcoinRequest{
					UserId:      order.UserId,
					Amount:      grant,
					SourceType:  ecoin_model.SourceTypeOrder,
					SourceId:    fmt.Sprintf("%s#%d", order.OrderNo, item.Id),
					Description: fmt.Sprintf("商品履约发放金币 - %s x%d", skuInfo.SkuName, item.Quantity),
				})
				if addErr != nil {
					fulfillStatus = ordermodel.FulfillStatusFailed
					fulfillMsg = addErr.Error()
				} else {
					fulfillStatus = ordermodel.FulfillStatusSuccess
					fulfillMsg = "金币发放成功"
				}
			}
		} else if skuInfo.DeliveryMethod == "" {
			fulfillStatus = ordermodel.FulfillStatusFailed
			fulfillMsg = "未配置履约回调接口"
		} else {
			bizUserId := bizUserIdMap[skuInfo.BizCode]
			if bizUserId == "" {
				fulfillStatus = ordermodel.FulfillStatusFailed
				fulfillMsg = fmt.Sprintf("biz user binding not found for biz_code: %s", skuInfo.BizCode)
			} else {
				resp, ferr := s.skuService.FulfillSku(ctx, &sku.FulfillSkuRequest{
					SkuId:     item.SkuId,
					BizUserId: bizUserId,
				})

				if ferr != nil {
					fulfillStatus = ordermodel.FulfillStatusFailed
					fulfillMsg = fmt.Sprintf("fulfill error: %v", ferr)
				} else if !resp.Success {
					fulfillStatus = ordermodel.FulfillStatusFailed
					fulfillMsg = resp.Message
				} else {
					fulfillStatus = ordermodel.FulfillStatusSuccess
					fulfillMsg = resp.Message
				}
			}
		}

		if err := s.orderRepo.UpdateOrderItemFulfillStatus(ctx, item.Id, fulfillStatus, fulfillTime, fulfillMsg); err != nil {
			return fmt.Errorf("failed to update order item fulfill status: %w", err)
		}
		// 同步内存态，避免后续整单状态判断误判。
		item.FulfillStatus = fulfillStatus
		item.FulfillTime = fulfillTime
		item.FulfillMsg = fulfillMsg
	}

	if allOrderItemsFulfilledSuccessfully(items) {
		fulfillTime := uint32(time.Now().Unix())
		if err := s.orderRepo.UpdateOrderToFulfilled(ctx, order.OrderNo, fulfillTime); err != nil {
			return fmt.Errorf("failed to update order to fulfilled: %w", err)
		}
	}

	return nil
}

func allOrderItemsFulfilledSuccessfully(items []*ordermodel.OrderItem) bool {
	if len(items) == 0 {
		return false
	}
	for _, item := range items {
		if item == nil || !item.IsFulfillSuccess() {
			return false
		}
	}
	return true
}

// doFulfill 执行履约（手动调用时 bizUserId 仅对回调模式生效，会覆盖按业务线映射的用户标识）
func (s *orderServiceImpl) doFulfill(ctx context.Context, orderNo string, bizUserId string) error {
	order, err := s.orderRepo.GetOrderByOrderNo(ctx, orderNo)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}
	if order == nil {
		return errors.New("order not found")
	}
	if !order.IsPaid() {
		return fmt.Errorf("order is not paid, current status: %d", order.Status)
	}

	items, err := s.orderRepo.GetOrderItemsByOrderId(ctx, order.Id)
	if err != nil {
		return fmt.Errorf("failed to get order items: %w", err)
	}

	if s.isEcoinRechargeOrder(items) {
		return s.fulfillEcoinRecharge(ctx, order, items)
	}

	bizUserIdMap := s.buildBizUserIdMap(ctx, order.UserId)
	return s.fulfillSkuItems(ctx, order, items, bizUserIdMap, bizUserId)
}

// RefundOrder 退款订单
func (s *orderServiceImpl) RefundOrder(ctx context.Context, req *RefundOrderRequest) error {
	if req.OrderNo == "" {
		return errors.New("order_no is required")
	}

	return database.Transaction(ctx, func(ctx context.Context) error {
		// 获取订单（加锁）
		order, err := s.orderRepo.GetOrderForUpdate(ctx, req.OrderNo)
		if err != nil {
			return fmt.Errorf("failed to get order: %w", err)
		}
		if order == nil {
			return errors.New("order not found")
		}

		// 验证是否可以退款
		if !order.CanRefund() {
			return fmt.Errorf("order cannot be refunded, current status: %d", order.Status)
		}

		if order.IsEcoinPay() {
			// 金币支付退款：返还金币
			err := s.addEcoinForOrder(ctx, order.UserId, &ecoin.AddEcoinRequest{
				UserId:      order.UserId,
				Amount:      float64(order.PayAmount),
				SourceType:  "refund",
				SourceId:    order.OrderNo,
				Description: fmt.Sprintf("订单退款 - 订单号: %s", order.OrderNo),
			})
			if err != nil {
				return fmt.Errorf("failed to refund ecoin: %w", err)
			}
		} else {
			// 货币支付退款：调用支付模块
			if order.PaymentOrderNo == "" {
				return errors.New("payment order not found")
			}

			_, err := s.paymentSvc.Refund(ctx, &payment.RefundRequest{
				OrderNo:   order.PaymentOrderNo,
				RefundFee: int64(order.PayAmount * 100), // 转换为分
				Reason:    req.Reason,
			})
			if err != nil {
				return fmt.Errorf("failed to refund payment: %w", err)
			}
		}

		// 更新订单为已退款
		if err := s.orderRepo.UpdateOrderToRefunded(ctx, req.OrderNo); err != nil {
			return fmt.Errorf("failed to update order to refunded: %w", err)
		}

		return nil
	})
}

// ListOrders 获取订单列表
func (s *orderServiceImpl) ListOrders(ctx context.Context, req *ListOrdersRequest) (*ListOrdersResponse, error) {
	if req.UserId == 0 {
		return nil, errors.New("user_id is required")
	}

	// 设置默认分页参数
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	// 获取订单列表
	orders, err := s.orderRepo.ListOrdersByUserId(ctx, req.UserId, req.Status, req.Offset, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}

	// 获取总数
	total, err := s.orderRepo.CountOrdersByUserId(ctx, req.UserId, req.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to count orders: %w", err)
	}

	return &ListOrdersResponse{
		Total: total,
		List:  orders,
	}, nil
}

// SyncOrderStatus 同步订单状态：待支付且货币支付时会向渠道查单；若查得已支付则更新业务订单并 asyncAutoFulfill（与微信回调里 HandlePaymentSuccess 行为对齐）
func (s *orderServiceImpl) SyncOrderStatus(ctx context.Context, orderNo string) (*ordermodel.Order, error) {
	if orderNo == "" {
		return nil, errors.New("order_no is required")
	}

	order, err := s.orderRepo.GetOrderByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	if order == nil {
		return nil, errors.New("order not found")
	}

	// 非待支付/已支付状态，直接返回
	if !order.IsPending() && order.Status != ordermodel.OrderStatusPaid {
		items, _ := s.orderRepo.GetOrderItemsByOrderId(ctx, order.Id)
		order.Items = items
		return order, nil
	}

	// 待支付 + 货币支付，主动查询支付渠道获取最新支付结果
	if order.IsPending() && order.IsMoneyPay() && order.PaymentOrderNo != "" {
		paymentOrder, err := s.paymentSvc.SyncPaymentStatus(ctx, order.PaymentOrderNo)
		if err != nil {
			return nil, fmt.Errorf("failed to sync payment status: %w", err)
		}

		if paymentOrder.IsPaid() {
			if err := s.orderRepo.UpdateOrderToPaid(ctx, orderNo, paymentOrder.PayTime); err != nil {
				return nil, fmt.Errorf("failed to update order to paid: %w", err)
			}
			// 与微信异步回调路径一致：查单确认支付后立即触发履约（否则仅依赖 order_fulfill_scan 定时任务）
			fmt.Printf("[SyncOrderStatus] order=%s marked paid via payment channel query (payment_order_no=%s pay_time=%d), scheduling auto fulfill\n",
				orderNo, order.PaymentOrderNo, paymentOrder.PayTime)
			go s.asyncAutoFulfill(orderNo)
		} else if paymentOrder.IsClosed() {
			cancelTime := uint32(time.Now().Unix())
			if err := s.orderRepo.UpdateOrderToCancelled(ctx, orderNo, cancelTime, "支付超时关闭"); err != nil {
				return nil, fmt.Errorf("failed to cancel order: %w", err)
			}
		}
	}

	// 返回最新订单状态
	latestOrder, err := s.orderRepo.GetOrderByOrderNo(ctx, orderNo)
	if err == nil && latestOrder != nil {
		items, _ := s.orderRepo.GetOrderItemsByOrderId(ctx, latestOrder.Id)
		latestOrder.Items = items
		return latestOrder, nil
	}

	items, _ := s.orderRepo.GetOrderItemsByOrderId(ctx, order.Id)
	order.Items = items
	return order, nil
}

// GiftOrder 运营赠送商品：创建零元已支付订单并触发履约
func (s *orderServiceImpl) GiftOrder(ctx context.Context, req *GiftOrderRequest) (*CreateOrderResponse, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}
	if req.UserId == 0 {
		return nil, errors.New("user_id is required")
	}
	if len(req.SkuItems) == 0 {
		return nil, errors.New("at least one sku_item is required")
	}

	createReq := &CreateOrderRequest{
		UserId:   req.UserId,
		SkuItems: req.SkuItems,
		PayType:  ordermodel.PayTypeMoney,
		Remark:   strings.TrimSpace(req.Remark),
	}
	order, orderItems, err := s.buildOrder(ctx, createReq)
	if err != nil {
		return nil, err
	}
	order.PayAmount = 0
	if order.Remark != "" {
		order.Remark = "OPS_GIFT: " + order.Remark
	} else {
		order.Remark = "OPS_GIFT"
	}

	payTime := uint32(time.Now().Unix())
	err = database.Transaction(ctx, func(ctx context.Context) error {
		if err := s.orderRepo.CreateOrder(ctx, order); err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}
		for _, item := range orderItems {
			item.OrderId = order.Id
		}
		if err := s.orderRepo.CreateOrderItems(ctx, orderItems); err != nil {
			return fmt.Errorf("failed to create order items: %w", err)
		}
		if err := s.orderRepo.UpdateOrderToPaid(ctx, order.OrderNo, payTime); err != nil {
			return fmt.Errorf("failed to update order to paid: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	order.Status = ordermodel.OrderStatusPaid
	order.PayTime = payTime
	order.Items = orderItems
	go s.asyncAutoFulfill(order.OrderNo)

	return &CreateOrderResponse{Order: order}, nil
}

// generateOrderNo 生成订单号
func (s *orderServiceImpl) generateOrderNo() string {
	// 格式: ORD + 年月日时分秒 + 6位随机数
	return fmt.Sprintf("ORD%s%06d",
		time.Now().Format("20060102150405"),
		time.Now().UnixNano()%1000000)
}
