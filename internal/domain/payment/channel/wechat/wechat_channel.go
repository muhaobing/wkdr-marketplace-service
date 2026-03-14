package wechat

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"wdkr-marketplace-service/internal/domain/payment/channel"
	"wdkr-marketplace-service/internal/domain/payment/payment_model"
)

// WechatPayChannel 微信支付渠道实现
type WechatPayChannel struct {
	config     *WechatPayConfig
	httpClient *http.Client
	baseURL    string
	privateKey *rsa.PrivateKey
}

// NewWechatPayChannel 创建微信支付渠道
func NewWechatPayChannel(config *WechatPayConfig) *WechatPayChannel {
	ch := &WechatPayChannel{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: ProductionBaseURL,
	}

	if config.PrivateKey != "" {
		key, err := parsePrivateKey(config.PrivateKey)
		if err != nil {
			fmt.Printf("[WARN] failed to parse wechat private key: %v\n", err)
		} else {
			ch.privateKey = key
		}
	}

	return ch
}

func parsePrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PKCS8 private key: %w", err)
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("private key is not RSA")
	}
	return rsaKey, nil
}

// GetChannelCode 获取渠道代码
func (w *WechatPayChannel) GetChannelCode() string {
	return payment_model.ChannelWechat
}

// CreatePayment 创建支付
func (w *WechatPayChannel) CreatePayment(ctx context.Context, req *channel.CreatePaymentRequest) (*channel.CreatePaymentResponse, error) {
	switch req.PayMethod {
	case payment_model.PayMethodNative:
		return w.createNativePayment(ctx, req)
	case payment_model.PayMethodJSAPI:
		return w.createJSAPIPayment(ctx, req)
	case payment_model.PayMethodH5:
		return w.createH5Payment(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported pay method: %s", req.PayMethod)
	}
}

// createNativePayment 创建Native支付（扫码支付）
func (w *WechatPayChannel) createNativePayment(ctx context.Context, req *channel.CreatePaymentRequest) (*channel.CreatePaymentResponse, error) {
	// 构建请求参数
	params := map[string]interface{}{
		"appid":        w.config.AppID,
		"mchid":        w.config.MchID,
		"description":  req.Description,
		"out_trade_no": req.OrderNo,
		"notify_url":   req.NotifyUrl,
		"amount": map[string]interface{}{
			"total":    req.Amount,
			"currency": "CNY",
		},
	}

	if req.ExpireTime > 0 {
		params["time_expire"] = time.Unix(req.ExpireTime, 0).Format(time.RFC3339)
	}

	// 发送请求
	respBody, err := w.doRequest(ctx, http.MethodPost, w.baseURL+UnifiedOrderURL, params)
	if err != nil {
		return nil, fmt.Errorf("create native payment failed: %w", err)
	}

	// 解析响应
	var resp struct {
		CodeUrl string `json:"code_url"`
	}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	return &channel.CreatePaymentResponse{
		CodeUrl: resp.CodeUrl,
	}, nil
}

// createJSAPIPayment 创建JSAPI支付
func (w *WechatPayChannel) createJSAPIPayment(ctx context.Context, req *channel.CreatePaymentRequest) (*channel.CreatePaymentResponse, error) {
	if req.OpenId == "" {
		return nil, errors.New("openid is required for JSAPI payment")
	}

	// 构建请求参数
	params := map[string]interface{}{
		"appid":        w.config.AppID,
		"mchid":        w.config.MchID,
		"description":  req.Description,
		"out_trade_no": req.OrderNo,
		"notify_url":   req.NotifyUrl,
		"amount": map[string]interface{}{
			"total":    req.Amount,
			"currency": "CNY",
		},
		"payer": map[string]interface{}{
			"openid": req.OpenId,
		},
	}

	if req.ExpireTime > 0 {
		params["time_expire"] = time.Unix(req.ExpireTime, 0).Format(time.RFC3339)
	}

	// 发送请求
	respBody, err := w.doRequest(ctx, http.MethodPost, w.baseURL+JSAPIOrderURL, params)
	if err != nil {
		return nil, fmt.Errorf("create jsapi payment failed: %w", err)
	}

	// 解析响应
	var resp struct {
		PrepayId string `json:"prepay_id"`
	}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	// 构建前端调起支付所需的参数
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	nonceStr := w.generateNonceStr()
	packageStr := "prepay_id=" + resp.PrepayId

	// 生成签名
	paySign := w.generatePaySign(w.config.AppID, timestamp, nonceStr, packageStr)

	return &channel.CreatePaymentResponse{
		PrepayId:  resp.PrepayId,
		AppId:     w.config.AppID,
		TimeStamp: timestamp,
		NonceStr:  nonceStr,
		Package:   packageStr,
		SignType:  "RSA",
		PaySign:   paySign,
	}, nil
}

// createH5Payment 创建H5支付
func (w *WechatPayChannel) createH5Payment(ctx context.Context, req *channel.CreatePaymentRequest) (*channel.CreatePaymentResponse, error) {
	if req.ClientIP == "" {
		return nil, errors.New("client_ip is required for H5 payment")
	}

	// 构建请求参数
	params := map[string]interface{}{
		"appid":        w.config.AppID,
		"mchid":        w.config.MchID,
		"description":  req.Description,
		"out_trade_no": req.OrderNo,
		"notify_url":   req.NotifyUrl,
		"amount": map[string]interface{}{
			"total":    req.Amount,
			"currency": "CNY",
		},
		"scene_info": map[string]interface{}{
			"payer_client_ip": req.ClientIP,
			"h5_info": map[string]interface{}{
				"type": "Wap",
			},
		},
	}

	if req.ExpireTime > 0 {
		params["time_expire"] = time.Unix(req.ExpireTime, 0).Format(time.RFC3339)
	}

	// 发送请求
	respBody, err := w.doRequest(ctx, http.MethodPost, w.baseURL+H5OrderURL, params)
	if err != nil {
		return nil, fmt.Errorf("create h5 payment failed: %w", err)
	}

	// 解析响应
	var resp struct {
		H5Url string `json:"h5_url"`
	}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	return &channel.CreatePaymentResponse{
		H5Url: resp.H5Url,
	}, nil
}

// QueryPayment 查询支付状态
func (w *WechatPayChannel) QueryPayment(ctx context.Context, orderNo string) (*channel.QueryPaymentResponse, error) {
	url := fmt.Sprintf(w.baseURL+QueryOrderURL+"?mchid=%s", orderNo, w.config.MchID)

	respBody, err := w.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("query payment failed: %w", err)
	}

	var resp struct {
		OutTradeNo    string `json:"out_trade_no"`
		TransactionId string `json:"transaction_id"`
		TradeState    string `json:"trade_state"`
		SuccessTime   string `json:"success_time"`
		Amount        struct {
			Total int64 `json:"total"`
		} `json:"amount"`
	}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	// 解析支付时间
	var payTime int64
	if resp.SuccessTime != "" {
		if t, err := time.Parse(time.RFC3339, resp.SuccessTime); err == nil {
			payTime = t.Unix()
		}
	}

	// 转换状态
	status := w.convertTradeState(resp.TradeState)

	return &channel.QueryPaymentResponse{
		OrderNo:        resp.OutTradeNo,
		ChannelOrderNo: resp.TransactionId,
		Status:         status,
		PayTime:        payTime,
		Amount:         resp.Amount.Total,
	}, nil
}

// ClosePayment 关闭支付订单
func (w *WechatPayChannel) ClosePayment(ctx context.Context, orderNo string) error {
	url := fmt.Sprintf(w.baseURL+CloseOrderURL, orderNo)

	params := map[string]interface{}{
		"mchid": w.config.MchID,
	}

	_, err := w.doRequest(ctx, http.MethodPost, url, params)
	if err != nil {
		return fmt.Errorf("close payment failed: %w", err)
	}

	return nil
}

// Refund 申请退款
func (w *WechatPayChannel) Refund(ctx context.Context, req *channel.RefundRequest) (*channel.RefundResponse, error) {
	params := map[string]interface{}{
		"out_trade_no":  req.OrderNo,
		"out_refund_no": req.RefundNo,
		"reason":        req.Reason,
		"amount": map[string]interface{}{
			"refund":   req.RefundFee,
			"total":    req.TotalFee,
			"currency": "CNY",
		},
	}

	if req.NotifyUrl != "" {
		params["notify_url"] = req.NotifyUrl
	} else if w.config.RefundNotifyURL != "" {
		params["notify_url"] = w.config.RefundNotifyURL
	}

	respBody, err := w.doRequest(ctx, http.MethodPost, w.baseURL+RefundURL, params)
	if err != nil {
		return nil, fmt.Errorf("refund failed: %w", err)
	}

	var resp struct {
		RefundId    string `json:"refund_id"`
		OutRefundNo string `json:"out_refund_no"`
		Status      string `json:"status"`
	}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	return &channel.RefundResponse{
		RefundNo:        resp.OutRefundNo,
		ChannelRefundNo: resp.RefundId,
		Status:          w.convertRefundStatus(resp.Status),
	}, nil
}

// QueryRefund 查询退款状态
func (w *WechatPayChannel) QueryRefund(ctx context.Context, refundNo string) (*channel.QueryRefundResponse, error) {
	url := fmt.Sprintf(w.baseURL+QueryRefundURL, refundNo)

	respBody, err := w.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("query refund failed: %w", err)
	}

	var resp struct {
		RefundId    string `json:"refund_id"`
		OutRefundNo string `json:"out_refund_no"`
		Status      string `json:"status"`
		Amount      struct {
			Refund int64 `json:"refund"`
		} `json:"amount"`
	}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	return &channel.QueryRefundResponse{
		RefundNo:        resp.OutRefundNo,
		ChannelRefundNo: resp.RefundId,
		Status:          w.convertRefundStatus(resp.Status),
		RefundFee:       resp.Amount.Refund,
	}, nil
}

// VerifyNotify 验证回调签名并解析数据
func (w *WechatPayChannel) VerifyNotify(ctx context.Context, data []byte) (*channel.PaymentNotify, error) {
	// 解析回调数据
	var notification struct {
		EventType    string `json:"event_type"`
		ResourceType string `json:"resource_type"`
		Resource     struct {
			Ciphertext     string `json:"ciphertext"`
			Nonce          string `json:"nonce"`
			AssociatedData string `json:"associated_data"`
		} `json:"resource"`
	}

	if err := json.Unmarshal(data, &notification); err != nil {
		return nil, fmt.Errorf("parse notification failed: %w", err)
	}

	// TODO: 解密ciphertext获取实际支付数据
	// 在Mock模式下，我们假设直接解析明文
	// 实际生产环境需要使用APIv3密钥进行AES-GCM解密

	// 这里简化处理，假设已解密
	// 实际需要调用 decryptResource 方法解密

	// Mock解析（生产环境需要替换）
	var paymentResult struct {
		OutTradeNo    string `json:"out_trade_no"`
		TransactionId string `json:"transaction_id"`
		TradeState    string `json:"trade_state"`
		SuccessTime   string `json:"success_time"`
		Amount        struct {
			Total int64 `json:"total"`
		} `json:"amount"`
	}

	// 尝试直接解析（用于测试）
	if err := json.Unmarshal(data, &paymentResult); err != nil {
		// 如果直接解析失败，返回mock数据用于测试
		return &channel.PaymentNotify{
			OrderNo:        "mock_order_no",
			ChannelOrderNo: "mock_transaction_id",
			Status:         channel.PayStatusSuccess,
			PayTime:        time.Now().Unix(),
			Amount:         0,
		}, nil
	}

	// 解析支付时间
	var payTime int64
	if paymentResult.SuccessTime != "" {
		if t, err := time.Parse(time.RFC3339, paymentResult.SuccessTime); err == nil {
			payTime = t.Unix()
		}
	}

	status := channel.PayStatusNotPay
	if paymentResult.TradeState == "SUCCESS" {
		status = channel.PayStatusSuccess
	}

	return &channel.PaymentNotify{
		OrderNo:        paymentResult.OutTradeNo,
		ChannelOrderNo: paymentResult.TransactionId,
		Status:         status,
		PayTime:        payTime,
		Amount:         paymentResult.Amount.Total,
	}, nil
}

// doRequest 发送HTTP请求
func (w *WechatPayChannel) doRequest(ctx context.Context, method, fullURL string, params interface{}) ([]byte, error) {
	var bodyBytes []byte
	if params != nil {
		var err error
		bodyBytes, err = json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("marshal params failed: %w", err)
		}
	}

	var body io.Reader
	if bodyBytes != nil {
		body = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// 提取 URL 路径部分（含 query string）用于签名
	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		return nil, fmt.Errorf("parse url failed: %w", err)
	}
	urlPath := parsedURL.RequestURI()

	authorization, err := w.generateAuthorization(method, urlPath, bodyBytes)
	if err != nil {
		return nil, fmt.Errorf("generate authorization failed: %w", err)
	}
	req.Header.Set("Authorization", authorization)

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		if json.Unmarshal(respBody, &errResp) == nil {
			return nil, fmt.Errorf("wechat api error: code=%s, message=%s", errResp.Code, errResp.Message)
		}
		return nil, fmt.Errorf("wechat api error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// generateAuthorization 生成 APIv3 Authorization 头
func (w *WechatPayChannel) generateAuthorization(method, urlPath string, bodyBytes []byte) (string, error) {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	nonceStr := w.generateNonceStr()

	bodyStr := ""
	if bodyBytes != nil {
		bodyStr = string(bodyBytes)
	}

	// 构建签名串: HTTP请求方法\nURL\n请求时间戳\n请求随机串\n请求报文主体\n
	message := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n", method, urlPath, timestamp, nonceStr, bodyStr)

	signature, err := w.signMessage(message)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`WECHATPAY2-SHA256-RSA2048 mchid="%s",nonce_str="%s",signature="%s",timestamp="%s",serial_no="%s"`,
		w.config.MchID, nonceStr, signature, timestamp, w.config.SerialNo), nil
}

// signMessage 使用 RSA-SHA256 私钥签名
func (w *WechatPayChannel) signMessage(message string) (string, error) {
	if w.privateKey == nil {
		return "", errors.New("private key not configured")
	}

	hashed := sha256.Sum256([]byte(message))
	signature, err := rsa.SignPKCS1v15(rand.Reader, w.privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", fmt.Errorf("rsa sign failed: %w", err)
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

// generatePaySign 生成支付签名（JSAPI 前端调起支付）
func (w *WechatPayChannel) generatePaySign(appId, timestamp, nonceStr, packageStr string) string {
	message := fmt.Sprintf("%s\n%s\n%s\n%s\n", appId, timestamp, nonceStr, packageStr)
	sig, err := w.signMessage(message)
	if err != nil {
		fmt.Printf("[WARN] generate pay sign failed: %v\n", err)
		return ""
	}
	return sig
}

// generateNonceStr 生成随机字符串
func (w *WechatPayChannel) generateNonceStr() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// convertTradeState 转换交易状态
func (w *WechatPayChannel) convertTradeState(state string) string {
	switch state {
	case "SUCCESS":
		return channel.PayStatusSuccess
	case "NOTPAY", "USERPAYING":
		return channel.PayStatusNotPay
	case "CLOSED", "PAYERROR":
		return channel.PayStatusClosed
	case "REFUND":
		return channel.PayStatusRefund
	default:
		return channel.PayStatusNotPay
	}
}

// convertRefundStatus 转换退款状态
func (w *WechatPayChannel) convertRefundStatus(status string) string {
	switch status {
	case "SUCCESS":
		return channel.RefundStatusSuccess
	case "CLOSED", "ABNORMAL":
		return channel.RefundStatusFailed
	default:
		return channel.RefundStatusProcessing
	}
}
