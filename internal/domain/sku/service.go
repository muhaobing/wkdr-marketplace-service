package sku

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	jsoniter "github.com/json-iterator/go"

	"wdkr-marketplace-service/internal/domain/sku/repo"
	skumodel "wdkr-marketplace-service/internal/domain/sku/sku_model"
)

type skuServiceImpl struct {
	skuRepo    repo.SkuRepo
	httpClient *http.Client
}

// NewSkuService 创建SKU服务实例
func NewSkuService(skuRepo repo.SkuRepo) SkuService {
	return &skuServiceImpl{
		skuRepo: skuRepo,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateSku 创建商品（默认未上架状态）
func (s *skuServiceImpl) CreateSku(ctx context.Context, req *CreateSkuRequest) (*skumodel.Sku, error) {
	// 参数校验
	if req.BizCode == "" {
		return nil, errors.New("biz_code is required")
	}
	if req.SkuCode == "" {
		return nil, errors.New("sku_code is required")
	}
	if req.SkuName == "" {
		return nil, errors.New("sku_name is required")
	}
	if req.Cost < 0 {
		return nil, errors.New("cost cannot be negative")
	}

	// 检查商品编码是否已存在
	existingSku, err := s.skuRepo.GetSkuByCode(ctx, req.BizCode, req.SkuCode)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing sku: %w", err)
	}
	if existingSku != nil {
		return nil, fmt.Errorf("sku with code %s already exists in biz %s", req.SkuCode, req.BizCode)
	}

	// 创建商品
	sku := &skumodel.Sku{
		BizCode:        req.BizCode,
		SkuCode:        req.SkuCode,
		SkuName:        req.SkuName,
		SkuAvatar:      req.SkuAvatar,
		SkuDesc:        req.SkuDesc,
		SkuStatus:      skumodel.SkuStatusOffline, // 默认未上架
		Cost:           req.Cost,
		DeliveryMethod: req.DeliveryMethod,
	}

	if err := s.skuRepo.CreateSku(ctx, sku); err != nil {
		return nil, fmt.Errorf("failed to create sku: %w", err)
	}

	return sku, nil
}

// GetSkuById 根据ID获取商品
func (s *skuServiceImpl) GetSkuById(ctx context.Context, id uint64) (*skumodel.Sku, error) {
	if id == 0 {
		return nil, errors.New("sku id is required")
	}

	sku, err := s.skuRepo.GetSkuById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get sku: %w", err)
	}

	if sku == nil {
		return nil, errors.New("sku not found")
	}

	return sku, nil
}

// GetSkuByCode 根据业务编码和商品编码获取商品
func (s *skuServiceImpl) GetSkuByCode(ctx context.Context, bizCode, skuCode string) (*skumodel.Sku, error) {
	if bizCode == "" {
		return nil, errors.New("biz_code is required")
	}
	if skuCode == "" {
		return nil, errors.New("sku_code is required")
	}

	sku, err := s.skuRepo.GetSkuByCode(ctx, bizCode, skuCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get sku: %w", err)
	}

	if sku == nil {
		return nil, errors.New("sku not found")
	}

	return sku, nil
}

// EditSku 编辑商品信息
func (s *skuServiceImpl) EditSku(ctx context.Context, req *EditSkuRequest) (*skumodel.Sku, error) {
	// 参数校验
	if req.Id == 0 {
		return nil, errors.New("sku id is required")
	}
	if req.SkuName == "" {
		return nil, errors.New("sku_name is required")
	}
	if req.Cost < 0 {
		return nil, errors.New("cost cannot be negative")
	}

	// 获取现有商品
	sku, err := s.skuRepo.GetSkuById(ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get sku: %w", err)
	}
	if sku == nil {
		return nil, errors.New("sku not found")
	}

	// 更新商品信息
	sku.SkuName = req.SkuName
	sku.SkuAvatar = req.SkuAvatar
	sku.SkuDesc = req.SkuDesc
	sku.Cost = req.Cost
	sku.DeliveryMethod = req.DeliveryMethod

	if err := s.skuRepo.UpdateSku(ctx, sku); err != nil {
		return nil, fmt.Errorf("failed to update sku: %w", err)
	}

	return sku, nil
}

// ListingSku 上架商品（将商品状态改为已上架）
func (s *skuServiceImpl) ListingSku(ctx context.Context, id uint64) error {
	if id == 0 {
		return errors.New("sku id is required")
	}

	// 获取现有商品
	sku, err := s.skuRepo.GetSkuById(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get sku: %w", err)
	}
	if sku == nil {
		return errors.New("sku not found")
	}

	// 检查商品是否已上架
	if sku.SkuStatus == skumodel.SkuStatusOnline {
		return errors.New("sku is already listed")
	}

	// 上架前校验：必须有履约方式
	if sku.DeliveryMethod == "" {
		return errors.New("delivery_method is required before listing")
	}

	// 更新状态为已上架
	if err := s.skuRepo.UpdateSkuStatus(ctx, id, skumodel.SkuStatusOnline); err != nil {
		return fmt.Errorf("failed to list sku: %w", err)
	}

	return nil
}

// DelistingSku 下架商品（将商品状态改为未上架）
func (s *skuServiceImpl) DelistingSku(ctx context.Context, id uint64) error {
	if id == 0 {
		return errors.New("sku id is required")
	}

	// 获取现有商品
	sku, err := s.skuRepo.GetSkuById(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get sku: %w", err)
	}
	if sku == nil {
		return errors.New("sku not found")
	}

	// 检查商品是否已下架
	if sku.SkuStatus == skumodel.SkuStatusOffline {
		return errors.New("sku is already delisted")
	}

	// 更新状态为未上架
	if err := s.skuRepo.UpdateSkuStatus(ctx, id, skumodel.SkuStatusOffline); err != nil {
		return fmt.Errorf("failed to delist sku: %w", err)
	}

	return nil
}

// ListSkus 获取商品列表
func (s *skuServiceImpl) ListSkus(ctx context.Context, req *ListSkuRequest) (*ListSkuResponse, error) {
	// 参数校验
	if req.BizCode == "" {
		return nil, errors.New("biz_code is required")
	}

	// 设置默认分页参数
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	// 获取商品列表
	skus, err := s.skuRepo.ListSkusByBizCode(ctx, req.BizCode, req.Status, req.Offset, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list skus: %w", err)
	}

	// 获取总数
	total, err := s.skuRepo.CountSkusByBizCode(ctx, req.BizCode, req.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to count skus: %w", err)
	}

	return &ListSkuResponse{
		Total: total,
		List:  skus,
	}, nil
}

// DeleteSku 删除商品（仅允许删除未上架的商品）
func (s *skuServiceImpl) DeleteSku(ctx context.Context, id uint64) error {
	if id == 0 {
		return errors.New("sku id is required")
	}

	// 获取现有商品
	sku, err := s.skuRepo.GetSkuById(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get sku: %w", err)
	}
	if sku == nil {
		return errors.New("sku not found")
	}

	// 检查商品状态，只能删除未上架的商品
	if sku.SkuStatus == skumodel.SkuStatusOnline {
		return errors.New("cannot delete listed sku, please delist it first")
	}

	// 删除商品
	if err := s.skuRepo.DeleteSku(ctx, id); err != nil {
		return fmt.Errorf("failed to delete sku: %w", err)
	}

	return nil
}

// FulfillSku 商品履约（调用商品的履约回调接口）
func (s *skuServiceImpl) FulfillSku(ctx context.Context, req *FulfillSkuRequest) (*FulfillSkuResponse, error) {
	// 参数校验
	if req.SkuId == 0 {
		return nil, errors.New("sku_id is required")
	}
	if req.BizUserId == "" {
		return nil, errors.New("biz_user_id is required")
	}

	// 获取商品信息
	sku, err := s.skuRepo.GetSkuById(ctx, req.SkuId)
	if err != nil {
		return nil, fmt.Errorf("failed to get sku: %w", err)
	}
	if sku == nil {
		return nil, errors.New("sku not found")
	}

	// 检查商品是否已上架
	if sku.SkuStatus != skumodel.SkuStatusOnline {
		return nil, errors.New("sku is not listed, cannot fulfill")
	}

	// 检查履约方式是否配置
	if sku.DeliveryMethod == "" {
		return nil, errors.New("sku delivery method is not configured")
	}

	// 调用履约回调接口
	fulfillResponse, err := s.callDeliveryMethod(ctx, sku.DeliveryMethod, sku.SkuCode, req.BizUserId)
	if err != nil {
		return &FulfillSkuResponse{
			Success: false,
			Message: fmt.Sprintf("fulfill callback failed: %v", err),
		}, nil
	}

	return fulfillResponse, nil
}

// deliveryCallbackRequest 履约回调请求体
type deliveryCallbackRequest struct {
	SkuCode   string `json:"sku_code"`
	BizUserId string `json:"biz_user_id"`
}

// deliveryCallbackResponse 履约回调响应体
type deliveryCallbackResponse struct {
	Retcode int    `json:"retcode"`
	Message string `json:"message"`
}

// callDeliveryMethod 调用履约回调接口
func (s *skuServiceImpl) callDeliveryMethod(ctx context.Context, url, skuCode, bizUserId string) (*FulfillSkuResponse, error) {
	// 构建请求体
	reqBody := deliveryCallbackRequest{
		SkuCode:   skuCode,
		BizUserId: bizUserId,
	}

	reqBytes, err := jsoniter.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send http request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("delivery callback returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	// 解析响应
	var callbackResp deliveryCallbackResponse
	if err := jsoniter.Unmarshal(respBytes, &callbackResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 根据回调结果返回
	if callbackResp.Retcode != 0 {
		return &FulfillSkuResponse{
			Success: false,
			Message: callbackResp.Message,
		}, nil
	}

	return &FulfillSkuResponse{
		Success: true,
		Message: callbackResp.Message,
	}, nil
}
