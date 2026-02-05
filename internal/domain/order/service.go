package order

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/muhaobing-eng/std-go/go-common/database"

	"wdkr-marketplace-service/internal/domain/ecoin"
	ordermodel "wdkr-marketplace-service/internal/domain/order/order_model"
	"wdkr-marketplace-service/internal/domain/order/repo"
	"wdkr-marketplace-service/internal/domain/payment"
	"wdkr-marketplace-service/internal/domain/payment/payment_model"
	"wdkr-marketplace-service/internal/domain/sku"
)

// orderServiceImpl 订单服务实现
type orderServiceImpl struct {
	orderRepo  repo.OrderRepo
	skuService sku.SkuService
	ecoinSvc   ecoin.EcoinService
	paymentSvc payment.PaymentService
}

// NewOrderService 创建订单服务实例
func NewOrderService(
	orderRepo repo.OrderRepo,
	skuService sku.SkuService,
	ecoinSvc ecoin.EcoinService,
	paymentSvc payment.PaymentService,
) OrderService {
	return &orderServiceImpl{
		orderRepo:  orderRepo,
		skuService: skuService,
		ecoinSvc:   ecoinSvc,
		paymentSvc: paymentSvc,
	}
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
	if len(req.SkuItems) == 0 {
		return nil, errors.New("at least one sku_item is required")
	}

	// 构建订单和明细
	order, orderItems, err := s.buildOrder(ctx, req)
	if err != nil {
		return nil, err
	}

	// 积分支付需要立即扣除积分
	if req.PayType == ordermodel.PayTypeEcoin {
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

			// 扣除积分
			_, err := s.ecoinSvc.DeductEcoin(ctx, &ecoin.DeductEcoinRequest{
				UserId:      req.UserId,
				Amount:      float64(order.PayAmount),
				SourceType:  "order",
				SourceId:    order.OrderNo,
				Description: fmt.Sprintf("购买商品 - 订单号: %s", order.OrderNo),
			})
			if err != nil {
				return fmt.Errorf("failed to deduct ecoin: %w", err)
			}

			// 更新订单为已支付
			payTime := uint32(time.Now().Unix())
			if err := s.orderRepo.UpdateOrderToPaid(ctx, order.OrderNo, payTime); err != nil {
				return fmt.Errorf("failed to update order to paid: %w", err)
			}
			order.Status = ordermodel.OrderStatusPaid
			order.PayTime = payTime

			return nil
		})
		if err != nil {
			return nil, err
		}
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

// PayOrder 支付订单（货币支付）
func (s *orderServiceImpl) PayOrder(ctx context.Context, req *PayOrderRequest) (*PayOrderResponse, error) {
	if req.OrderNo == "" {
		return nil, errors.New("order_no is required")
	}
	if req.Channel == "" {
		return nil, errors.New("channel is required")
	}
	if req.PayMethod == "" {
		return nil, errors.New("pay_method is required")
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

	// 验证支付类型
	if !order.IsMoneyPay() {
		return nil, errors.New("this order uses ecoin payment, not money payment")
	}

	// 如果已有支付订单，直接查询返回
	if order.PaymentOrderNo != "" {
		paymentOrder, err := s.paymentSvc.QueryPayment(ctx, order.PaymentOrderNo)
		if err == nil && paymentOrder != nil && paymentOrder.IsPending() {
			// 支付订单仍然有效，返回已有的支付信息
			// 这里可能需要重新获取支付凭证
		}
	}

	// 创建支付订单
	paymentResp, err := s.paymentSvc.CreatePayment(ctx, &payment.CreatePaymentRequest{
		BizOrderNo:    req.OrderNo,
		BizType:       payment_model.BizTypePurchase,
		UserId:        order.UserId,
		Channel:       req.Channel,
		PayMethod:     req.PayMethod,
		Amount:        int64(order.PayAmount * 100), // 转换为分
		Description:   fmt.Sprintf("商城订单 - %s", req.OrderNo),
		ClientIP:      req.ClientIP,
		OpenId:        req.OpenId,
		ExpireMinutes: 30,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// 更新订单的支付订单号
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

// HandlePaymentSuccess 处理支付成功回调
func (s *orderServiceImpl) HandlePaymentSuccess(ctx context.Context, orderNo string, payTime uint32) error {
	if orderNo == "" {
		return errors.New("order_no is required")
	}

	return database.Transaction(ctx, func(ctx context.Context) error {
		// 获取订单（加锁）
		order, err := s.orderRepo.GetOrderForUpdate(ctx, orderNo)
		if err != nil {
			return fmt.Errorf("failed to get order: %w", err)
		}
		if order == nil {
			return errors.New("order not found")
		}

		// 检查订单状态
		if !order.IsPending() {
			// 订单已处理，忽略重复回调
			return nil
		}

		// 更新订单为已支付
		if err := s.orderRepo.UpdateOrderToPaid(ctx, orderNo, payTime); err != nil {
			return fmt.Errorf("failed to update order to paid: %w", err)
		}

		return nil
	})
}

// FulfillOrder 履约订单
func (s *orderServiceImpl) FulfillOrder(ctx context.Context, orderNo string, bizUserId string) error {
	if orderNo == "" {
		return errors.New("order_no is required")
	}
	if bizUserId == "" {
		return errors.New("biz_user_id is required")
	}

	// 获取订单
	order, err := s.orderRepo.GetOrderByOrderNo(ctx, orderNo)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}
	if order == nil {
		return errors.New("order not found")
	}

	// 验证订单状态（已支付才能履约）
	if !order.IsPaid() {
		return fmt.Errorf("order is not paid, current status: %d", order.Status)
	}

	// 获取待履约的订单明细
	items, err := s.orderRepo.GetPendingFulfillItems(ctx, orderNo)
	if err != nil {
		return fmt.Errorf("failed to get pending fulfill items: %w", err)
	}

	if len(items) == 0 {
		// 所有商品已履约，更新订单状态
		fulfillTime := uint32(time.Now().Unix())
		return s.orderRepo.UpdateOrderToFulfilled(ctx, orderNo, fulfillTime)
	}

	// 逐个履约
	allSuccess := true
	for _, item := range items {
		// 调用SKU履约接口
		resp, err := s.skuService.FulfillSku(ctx, &sku.FulfillSkuRequest{
			SkuId:     item.SkuId,
			BizUserId: bizUserId,
		})

		fulfillTime := uint32(time.Now().Unix())
		var fulfillStatus uint8
		var fulfillMsg string

		if err != nil {
			fulfillStatus = ordermodel.FulfillStatusFailed
			fulfillMsg = fmt.Sprintf("fulfill error: %v", err)
			allSuccess = false
		} else if !resp.Success {
			fulfillStatus = ordermodel.FulfillStatusFailed
			fulfillMsg = resp.Message
			allSuccess = false
		} else {
			fulfillStatus = ordermodel.FulfillStatusSuccess
			fulfillMsg = resp.Message
		}

		// 更新明细履约状态
		if err := s.orderRepo.UpdateOrderItemFulfillStatus(ctx, item.Id, fulfillStatus, fulfillTime, fulfillMsg); err != nil {
			return fmt.Errorf("failed to update order item fulfill status: %w", err)
		}
	}

	// 如果全部履约成功，更新订单状态
	if allSuccess {
		fulfillTime := uint32(time.Now().Unix())
		if err := s.orderRepo.UpdateOrderToFulfilled(ctx, orderNo, fulfillTime); err != nil {
			return fmt.Errorf("failed to update order to fulfilled: %w", err)
		}
	}

	return nil
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
			// 积分支付退款：返还积分
			_, err := s.ecoinSvc.AddEcoin(ctx, &ecoin.AddEcoinRequest{
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

// SyncOrderStatus 同步订单状态
func (s *orderServiceImpl) SyncOrderStatus(ctx context.Context, orderNo string) (*ordermodel.Order, error) {
	if orderNo == "" {
		return nil, errors.New("order_no is required")
	}

	// 获取订单
	order, err := s.orderRepo.GetOrderByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	if order == nil {
		return nil, errors.New("order not found")
	}

	// 如果订单不是待支付状态，直接返回
	if !order.IsPending() {
		return order, nil
	}

	// 如果是货币支付且有支付订单号，同步支付状态
	if order.IsMoneyPay() && order.PaymentOrderNo != "" {
		paymentOrder, err := s.paymentSvc.SyncPaymentStatus(ctx, order.PaymentOrderNo)
		if err != nil {
			return nil, fmt.Errorf("failed to sync payment status: %w", err)
		}

		if paymentOrder.IsPaid() {
			// 支付成功，更新订单状态
			if err := s.HandlePaymentSuccess(ctx, orderNo, paymentOrder.PayTime); err != nil {
				return nil, err
			}
			order.Status = ordermodel.OrderStatusPaid
			order.PayTime = paymentOrder.PayTime
		} else if paymentOrder.IsClosed() {
			// 支付关闭，取消订单
			cancelTime := uint32(time.Now().Unix())
			if err := s.orderRepo.UpdateOrderToCancelled(ctx, orderNo, cancelTime, "支付超时关闭"); err != nil {
				return nil, fmt.Errorf("failed to cancel order: %w", err)
			}
			order.Status = ordermodel.OrderStatusCancelled
			order.CancelTime = cancelTime
			order.CancelReason = "支付超时关闭"
		}
	}

	return order, nil
}

// generateOrderNo 生成订单号
func (s *orderServiceImpl) generateOrderNo() string {
	// 格式: ORD + 年月日时分秒 + 6位随机数
	return fmt.Sprintf("ORD%s%06d",
		time.Now().Format("20060102150405"),
		time.Now().UnixNano()%1000000)
}
