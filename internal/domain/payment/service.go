package payment

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/muhaobing/std-go/go-common/database"

	"wdkr-marketplace-service/internal/common/config"
	"wdkr-marketplace-service/internal/domain/payment/channel"
	"wdkr-marketplace-service/internal/domain/payment/payment_model"
	"wdkr-marketplace-service/internal/domain/payment/repo"
)

// paymentServiceImpl 支付服务实现
type paymentServiceImpl struct {
	paymentRepo repo.PaymentRepo
	channels    map[string]channel.PaymentChannel
}

// NewPaymentService 创建支付服务实例
func NewPaymentService(paymentRepo repo.PaymentRepo, channels ...channel.PaymentChannel) PaymentService {
	channelMap := make(map[string]channel.PaymentChannel)
	for _, ch := range channels {
		channelMap[ch.GetChannelCode()] = ch
	}

	return &paymentServiceImpl{
		paymentRepo: paymentRepo,
		channels:    channelMap,
	}
}

// getChannel 获取支付渠道
func (s *paymentServiceImpl) getChannel(channelCode string) (channel.PaymentChannel, error) {
	ch, exists := s.channels[channelCode]
	if !exists {
		return nil, fmt.Errorf("unsupported payment channel: %s", channelCode)
	}
	return ch, nil
}

// CreatePayment 创建支付订单
func (s *paymentServiceImpl) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	// 参数校验
	if req.BizOrderNo == "" {
		return nil, errors.New("biz_order_no is required")
	}
	if req.UserId == 0 {
		return nil, errors.New("user_id is required")
	}
	if req.Channel == "" {
		return nil, errors.New("channel is required")
	}
	if req.PayMethod == "" {
		return nil, errors.New("pay_method is required")
	}
	if req.Amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	if req.Description == "" {
		return nil, errors.New("description is required")
	}

	// 获取支付渠道
	ch, err := s.getChannel(req.Channel)
	if err != nil {
		return nil, err
	}

	// 生成支付订单号
	orderNo := s.generateOrderNo()

	// 计算过期时间（默认30分钟）
	expireMinutes := req.ExpireMinutes
	if expireMinutes <= 0 {
		expireMinutes = 30
	}
	expireTime := uint32(time.Now().Add(time.Duration(expireMinutes) * time.Minute).Unix())

	// 获取回调地址
	notifyUrl := s.getNotifyUrl(req.Channel)

	var response *CreatePaymentResponse
	err = database.Transaction(ctx, func(ctx context.Context) error {
		// 检查是否已存在相同业务订单号的待支付订单
		existingOrder, err := s.paymentRepo.GetPaymentOrderByBizOrderNo(ctx, req.BizOrderNo)
		if err != nil {
			return fmt.Errorf("failed to check existing order: %w", err)
		}
		if existingOrder != nil && existingOrder.IsPending() {
			// 如果已存在待支付订单，返回该订单信息
			// 可以考虑重新调用渠道获取支付凭证
			return errors.New("pending payment order already exists for this biz_order_no")
		}

		// 创建支付订单
		order := &payment_model.PaymentOrder{
			OrderNo:    orderNo,
			BizOrderNo: req.BizOrderNo,
			BizType:    req.BizType,
			UserId:     req.UserId,
			Channel:    req.Channel,
			PayMethod:  req.PayMethod,
			Amount:     req.Amount,
			Status:     payment_model.PaymentStatusPending,
			ExpireTime: expireTime,
			NotifyUrl:  notifyUrl,
		}

		if err := s.paymentRepo.CreatePaymentOrder(ctx, order); err != nil {
			return fmt.Errorf("failed to create payment order: %w", err)
		}

		// 调用支付渠道创建支付
		channelReq := &channel.CreatePaymentRequest{
			OrderNo:     orderNo,
			Amount:      req.Amount,
			Description: req.Description,
			PayMethod:   req.PayMethod,
			NotifyUrl:   notifyUrl,
			ExpireTime:  int64(expireTime),
			ClientIP:    req.ClientIP,
			OpenId:      req.OpenId,
		}

		channelResp, err := ch.CreatePayment(ctx, channelReq)
		if err != nil {
			return fmt.Errorf("failed to create payment on channel: %w", err)
		}

		response = &CreatePaymentResponse{
			OrderNo:   orderNo,
			CodeUrl:   channelResp.CodeUrl,
			H5Url:     channelResp.H5Url,
			PrepayId:  channelResp.PrepayId,
			AppId:     channelResp.AppId,
			TimeStamp: channelResp.TimeStamp,
			NonceStr:  channelResp.NonceStr,
			Package:   channelResp.Package,
			SignType:  channelResp.SignType,
			PaySign:   channelResp.PaySign,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return response, nil
}

// HandleNotify 处理支付回调
func (s *paymentServiceImpl) HandleNotify(ctx context.Context, channelCode string, data []byte) (*PaymentNotifyResult, error) {
	ch, err := s.getChannel(channelCode)
	if err != nil {
		return nil, err
	}

	notify, err := ch.VerifyNotify(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("failed to verify notify: %w", err)
	}

	var result *PaymentNotifyResult
	err = database.Transaction(ctx, func(ctx context.Context) error {
		order, err := s.paymentRepo.GetPaymentOrderForUpdate(ctx, notify.OrderNo)
		if err != nil {
			return fmt.Errorf("failed to get payment order: %w", err)
		}
		if order == nil {
			return fmt.Errorf("payment order not found: %s", notify.OrderNo)
		}

		if !order.IsPending() {
			return nil
		}

		var newStatus uint8
		if notify.Status == channel.PayStatusSuccess {
			newStatus = payment_model.PaymentStatusPaid
		} else {
			newStatus = payment_model.PaymentStatusClosed
		}

		if err := s.paymentRepo.UpdatePaymentOrderStatus(ctx, notify.OrderNo, newStatus, notify.ChannelOrderNo, uint32(notify.PayTime)); err != nil {
			return fmt.Errorf("failed to update payment order status: %w", err)
		}

		result = &PaymentNotifyResult{
			BizOrderNo: order.BizOrderNo,
			Status:     newStatus,
			PayTime:    uint32(notify.PayTime),
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// QueryPayment 查询支付状态
func (s *paymentServiceImpl) QueryPayment(ctx context.Context, orderNo string) (*payment_model.PaymentOrder, error) {
	if orderNo == "" {
		return nil, errors.New("order_no is required")
	}

	order, err := s.paymentRepo.GetPaymentOrderByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment order: %w", err)
	}
	if order == nil {
		return nil, errors.New("payment order not found")
	}

	return order, nil
}

// SyncPaymentStatus 同步支付状态
func (s *paymentServiceImpl) SyncPaymentStatus(ctx context.Context, orderNo string) (*payment_model.PaymentOrder, error) {
	if orderNo == "" {
		return nil, errors.New("order_no is required")
	}

	// 获取订单
	order, err := s.paymentRepo.GetPaymentOrderByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment order: %w", err)
	}
	if order == nil {
		return nil, errors.New("payment order not found")
	}

	// 如果订单不是待支付状态，直接返回
	if !order.IsPending() {
		return order, nil
	}

	// 获取支付渠道
	ch, err := s.getChannel(order.Channel)
	if err != nil {
		return nil, err
	}

	// 查询渠道支付状态
	queryResp, err := ch.QueryPayment(ctx, orderNo)
	if err != nil {
		return nil, fmt.Errorf("failed to query payment from channel: %w", err)
	}

	// 根据渠道状态更新订单
	var newStatus uint8
	switch queryResp.Status {
	case channel.PayStatusSuccess:
		newStatus = payment_model.PaymentStatusPaid
	case channel.PayStatusClosed:
		newStatus = payment_model.PaymentStatusClosed
	case channel.PayStatusRefund:
		newStatus = payment_model.PaymentStatusRefunded
	default:
		// 仍然是未支付状态，检查是否过期
		if order.ExpireTime > 0 && uint32(time.Now().Unix()) > order.ExpireTime {
			// 订单已过期，尝试关闭
			_ = s.ClosePayment(ctx, orderNo)
			newStatus = payment_model.PaymentStatusClosed
		} else {
			return order, nil
		}
	}

	// 更新订单状态
	if err := s.paymentRepo.UpdatePaymentOrderStatus(ctx, orderNo, newStatus, queryResp.ChannelOrderNo, uint32(queryResp.PayTime)); err != nil {
		return nil, fmt.Errorf("failed to update payment order status: %w", err)
	}

	// 重新获取订单
	return s.paymentRepo.GetPaymentOrderByOrderNo(ctx, orderNo)
}

// ClosePayment 关闭支付订单
func (s *paymentServiceImpl) ClosePayment(ctx context.Context, orderNo string) error {
	if orderNo == "" {
		return errors.New("order_no is required")
	}

	return database.Transaction(ctx, func(ctx context.Context) error {
		// 获取订单（加锁）
		order, err := s.paymentRepo.GetPaymentOrderForUpdate(ctx, orderNo)
		if err != nil {
			return fmt.Errorf("failed to get payment order: %w", err)
		}
		if order == nil {
			return errors.New("payment order not found")
		}

		// 检查订单状态
		if !order.IsPending() {
			return fmt.Errorf("cannot close payment order with status: %d", order.Status)
		}

		// 获取支付渠道
		ch, err := s.getChannel(order.Channel)
		if err != nil {
			return err
		}

		// 调用渠道关闭订单
		if err := ch.ClosePayment(ctx, orderNo); err != nil {
			// 关闭失败不阻塞本地状态更新
			// 可能订单在渠道已经是关闭状态
		}

		// 更新本地订单状态
		if err := s.paymentRepo.UpdatePaymentOrderToClosed(ctx, orderNo); err != nil {
			return fmt.Errorf("failed to update payment order status: %w", err)
		}

		return nil
	})
}

// Refund 申请退款
func (s *paymentServiceImpl) Refund(ctx context.Context, req *RefundRequest) (*payment_model.PaymentRefund, error) {
	// 参数校验
	if req.OrderNo == "" {
		return nil, errors.New("order_no is required")
	}
	if req.RefundFee <= 0 {
		return nil, errors.New("refund_fee must be positive")
	}

	var refund *payment_model.PaymentRefund
	err := database.Transaction(ctx, func(ctx context.Context) error {
		// 获取订单（加锁）
		order, err := s.paymentRepo.GetPaymentOrderForUpdate(ctx, req.OrderNo)
		if err != nil {
			return fmt.Errorf("failed to get payment order: %w", err)
		}
		if order == nil {
			return errors.New("payment order not found")
		}

		// 检查订单状态
		if !order.IsPaid() {
			return fmt.Errorf("cannot refund payment order with status: %d", order.Status)
		}

		// 检查退款金额
		if req.RefundFee > order.Amount {
			return errors.New("refund_fee exceeds order amount")
		}

		// 生成退款单号
		refundNo := s.generateRefundNo()

		// 获取支付渠道
		ch, err := s.getChannel(order.Channel)
		if err != nil {
			return err
		}

		// 调用渠道申请退款
		channelReq := &channel.RefundRequest{
			OrderNo:   req.OrderNo,
			RefundNo:  refundNo,
			TotalFee:  order.Amount,
			RefundFee: req.RefundFee,
			Reason:    req.Reason,
		}

		channelResp, err := ch.Refund(ctx, channelReq)
		if err != nil {
			return fmt.Errorf("failed to refund on channel: %w", err)
		}

		// 转换退款状态
		var refundStatus uint8
		switch channelResp.Status {
		case channel.RefundStatusSuccess:
			refundStatus = payment_model.RefundStatusSuccess
		case channel.RefundStatusFailed:
			refundStatus = payment_model.RefundStatusFailed
		default:
			refundStatus = payment_model.RefundStatusProcessing
		}

		// 创建退款记录
		refund = &payment_model.PaymentRefund{
			RefundNo:        refundNo,
			OrderNo:         req.OrderNo,
			Channel:         order.Channel,
			Amount:          req.RefundFee,
			Reason:          req.Reason,
			Status:          refundStatus,
			ChannelRefundNo: channelResp.ChannelRefundNo,
		}

		if err := s.paymentRepo.CreatePaymentRefund(ctx, refund); err != nil {
			return fmt.Errorf("failed to create payment refund: %w", err)
		}

		// 如果退款成功，更新订单状态
		if refundStatus == payment_model.RefundStatusSuccess {
			if err := s.paymentRepo.UpdatePaymentOrderToRefunded(ctx, req.OrderNo); err != nil {
				return fmt.Errorf("failed to update payment order status: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return refund, nil
}

// QueryRefund 查询退款状态
func (s *paymentServiceImpl) QueryRefund(ctx context.Context, refundNo string) (*payment_model.PaymentRefund, error) {
	if refundNo == "" {
		return nil, errors.New("refund_no is required")
	}

	refund, err := s.paymentRepo.GetPaymentRefundByRefundNo(ctx, refundNo)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment refund: %w", err)
	}
	if refund == nil {
		return nil, errors.New("payment refund not found")
	}

	return refund, nil
}

// SyncRefundStatus 同步退款状态
func (s *paymentServiceImpl) SyncRefundStatus(ctx context.Context, refundNo string) (*payment_model.PaymentRefund, error) {
	if refundNo == "" {
		return nil, errors.New("refund_no is required")
	}

	// 获取退款记录
	refund, err := s.paymentRepo.GetPaymentRefundByRefundNo(ctx, refundNo)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment refund: %w", err)
	}
	if refund == nil {
		return nil, errors.New("payment refund not found")
	}

	// 如果退款不是处理中状态，直接返回
	if !refund.IsProcessing() {
		return refund, nil
	}

	// 获取支付渠道
	ch, err := s.getChannel(refund.Channel)
	if err != nil {
		return nil, err
	}

	// 查询渠道退款状态
	queryResp, err := ch.QueryRefund(ctx, refundNo)
	if err != nil {
		return nil, fmt.Errorf("failed to query refund from channel: %w", err)
	}

	// 转换退款状态
	var newStatus uint8
	switch queryResp.Status {
	case channel.RefundStatusSuccess:
		newStatus = payment_model.RefundStatusSuccess
	case channel.RefundStatusFailed:
		newStatus = payment_model.RefundStatusFailed
	default:
		return refund, nil
	}

	// 更新退款状态
	if err := s.paymentRepo.UpdatePaymentRefundStatus(ctx, refundNo, newStatus, queryResp.ChannelRefundNo); err != nil {
		return nil, fmt.Errorf("failed to update payment refund status: %w", err)
	}

	// 如果退款成功，更新订单状态
	if newStatus == payment_model.RefundStatusSuccess {
		if err := s.paymentRepo.UpdatePaymentOrderToRefunded(ctx, refund.OrderNo); err != nil {
			return nil, fmt.Errorf("failed to update payment order status: %w", err)
		}
	}

	// 重新获取退款记录
	return s.paymentRepo.GetPaymentRefundByRefundNo(ctx, refundNo)
}

// ListPaymentOrders 获取支付订单列表
func (s *paymentServiceImpl) ListPaymentOrders(ctx context.Context, req *ListPaymentOrdersRequest) (*ListPaymentOrdersResponse, error) {
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
	orders, err := s.paymentRepo.ListPaymentOrdersByUserId(ctx, req.UserId, req.Offset, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list payment orders: %w", err)
	}

	// 获取总数
	total, err := s.paymentRepo.CountPaymentOrdersByUserId(ctx, req.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to count payment orders: %w", err)
	}

	return &ListPaymentOrdersResponse{
		Total: total,
		List:  orders,
	}, nil
}

// GetPaymentOrderByBizOrderNo 根据业务订单号获取支付订单
func (s *paymentServiceImpl) GetPaymentOrderByBizOrderNo(ctx context.Context, bizOrderNo string) (*payment_model.PaymentOrder, error) {
	if bizOrderNo == "" {
		return nil, errors.New("biz_order_no is required")
	}

	order, err := s.paymentRepo.GetPaymentOrderByBizOrderNo(ctx, bizOrderNo)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment order: %w", err)
	}
	if order == nil {
		return nil, errors.New("payment order not found")
	}

	return order, nil
}

// generateOrderNo 生成支付订单号
func (s *paymentServiceImpl) generateOrderNo() string {
	// 格式: PAY + 年月日时分秒 + 6位随机数
	return fmt.Sprintf("PAY%s%06d",
		time.Now().Format("20060102150405"),
		time.Now().UnixNano()%1000000)
}

// generateRefundNo 生成退款单号
func (s *paymentServiceImpl) generateRefundNo() string {
	// 格式: REF + 年月日时分秒 + 6位随机数
	return fmt.Sprintf("REF%s%06d",
		time.Now().Format("20060102150405"),
		time.Now().UnixNano()%1000000)
}

// getNotifyUrl 获取回调地址
func (s *paymentServiceImpl) getNotifyUrl(channelCode string) string {
	cfg := config.GetWechatPayConfig()
	if cfg != nil && channelCode == payment_model.ChannelWechat && cfg.NotifyURL != "" {
		return cfg.NotifyURL
	}
	return fmt.Sprintf("https://your-domain.com/openapi/callback/%s/pay", channelCode)
}
