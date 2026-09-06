package ops

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"wdkr-marketplace-service/internal/common/utils/http_utils"
	"wdkr-marketplace-service/internal/domain/companyecoin"
	"wdkr-marketplace-service/internal/domain/ecoin"
	"wdkr-marketplace-service/internal/domain/order"
	"wdkr-marketplace-service/internal/domain/sku"
	skumodel "wdkr-marketplace-service/internal/domain/sku/sku_model"
	usermodel "wdkr-marketplace-service/internal/domain/user/user_model"
)

const (
	opsEcoinSourceAdjust = "admin"
	opsEcoinSourceGift   = "system"
)

// LookupUser 查询商城用户及金币概况
// GET /ops/users/lookup
func (r *OpsResource) LookupUser(ctx *gin.Context) {
	userIDStr := strings.TrimSpace(ctx.Query("user_id"))
	telNo := strings.TrimSpace(ctx.Query("tel_no"))
	companyIDStr := strings.TrimSpace(ctx.Query("company_id"))
	bizCode := strings.TrimSpace(ctx.Query("biz_code"))
	bizUserIDStr := strings.TrimSpace(ctx.Query("biz_user_id"))

	var user *usermodel.User
	var err error

	if userIDStr != "" {
		id, parseErr := strconv.ParseUint(userIDStr, 10, 64)
		if parseErr != nil || id == 0 {
			http_utils.WriteResponse(ctx, nil, errors.New("invalid user_id"))
			return
		}
		user, err = r.userService.GetUserById(ctx.Request.Context(), uint(id))
	} else if bizCode != "" && bizUserIDStr != "" {
		bizUserID, parseErr := strconv.ParseUint(bizUserIDStr, 10, 64)
		if parseErr != nil || bizUserID == 0 {
			http_utils.WriteResponse(ctx, nil, errors.New("invalid biz_user_id"))
			return
		}
		user, err = r.userService.GetUserByBiz(ctx.Request.Context(), bizCode, bizUserID)
	} else if telNo != "" {
		var companyID uint64
		if companyIDStr != "" {
			companyID, err = strconv.ParseUint(companyIDStr, 10, 64)
			if err != nil {
				http_utils.WriteResponse(ctx, nil, errors.New("invalid company_id"))
				return
			}
		}
		user, err = r.userService.GetUserByIdentity(ctx.Request.Context(), telNo, "", companyID)
	} else {
		http_utils.WriteResponse(ctx, nil, errors.New("请提供 user_id、手机号或 biz_code+biz_user_id"))
		return
	}
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}
	if user == nil {
		http_utils.WriteResponse(ctx, nil, errors.New("用户不存在"))
		return
	}

	bindings, _ := r.userService.GetBindingsByUserId(ctx.Request.Context(), user.Id)
	ecoinSummary, ecoinErr := r.loadUserEcoinSummary(ctx, uint64(user.Id), user.CompanyId)
	if ecoinErr != nil {
		http_utils.WriteResponse(ctx, nil, ecoinErr)
		return
	}

	http_utils.WriteResponse(ctx, gin.H{
		"user":     user,
		"bindings": bindings,
		"ecoin":    ecoinSummary,
	}, nil)
}

func (r *OpsResource) loadUserEcoinSummary(ctx *gin.Context, userID uint64, companyID uint64) (gin.H, error) {
	if companyID > 0 {
		ce, err := r.companyEcoinService.GetCompanyEcoin(ctx.Request.Context(), companyID)
		if err != nil {
			return nil, err
		}
		return gin.H{
			"account_type":    "company",
			"company_id":      companyID,
			"user_id":         userID,
			"available_stock": ce.AvailableStock,
		}, nil
	}
	info, err := r.ecoinService.GetUserEcoin(ctx.Request.Context(), userID)
	if err != nil {
		return nil, err
	}
	if info == nil {
		return gin.H{
			"account_type":    "personal",
			"user_id":         userID,
			"available_stock": float64(0),
		}, nil
	}
	return gin.H{
		"account_type":    "personal",
		"user_id":         userID,
		"available_stock": info.AvailableStock,
	}, nil
}

// AdjustEcoinRequest 运营调整金币
type AdjustEcoinRequest struct {
	UserId        uint64  `json:"user_id" binding:"required"`
	Mode          string  `json:"mode" binding:"required"` // set | add | deduct
	Amount        float64 `json:"amount"`
	TargetBalance float64 `json:"target_balance"`
	Description   string  `json:"description"`
}

// AdjustEcoin 直接调整用户金币（设为指定值 / 增减）
// POST /ops/ecoin/adjust
func (r *OpsResource) AdjustEcoin(ctx *gin.Context) {
	var req AdjustEcoinRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}
	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode != "set" && mode != "add" && mode != "deduct" {
		http_utils.WriteResponse(ctx, nil, errors.New("mode must be set, add or deduct"))
		return
	}

	user, err := r.userService.GetUserById(ctx.Request.Context(), uint(req.UserId))
	if err != nil || user == nil {
		http_utils.WriteResponse(ctx, nil, errors.New("user not found"))
		return
	}

	current, err := r.getAvailableStock(ctx, uint64(user.Id), user.CompanyId)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	var delta float64
	switch mode {
	case "set":
		delta = req.TargetBalance - current
	case "add":
		if req.Amount <= 0 {
			http_utils.WriteResponse(ctx, nil, errors.New("amount must be greater than 0"))
			return
		}
		delta = req.Amount
	case "deduct":
		if req.Amount <= 0 {
			http_utils.WriteResponse(ctx, nil, errors.New("amount must be greater than 0"))
			return
		}
		delta = -req.Amount
	}

	if delta == 0 {
		http_utils.WriteResponse(ctx, gin.H{
			"before_balance": current,
			"after_balance":  current,
			"delta":          0,
		}, nil)
		return
	}

	desc := strings.TrimSpace(req.Description)
	if desc == "" {
		if mode == "set" {
			desc = fmt.Sprintf("运营调整金币至 %.2f", req.TargetBalance)
		} else if delta > 0 {
			desc = fmt.Sprintf("运营增加金币 %.2f", delta)
		} else {
			desc = fmt.Sprintf("运营扣减金币 %.2f", -delta)
		}
	}

	sourceID := fmt.Sprintf("ops_adjust:%d:%d", req.UserId, time.Now().UnixNano())
	var tx interface{}
	if delta > 0 {
		tx, err = r.addUserEcoin(ctx, user, delta, opsEcoinSourceAdjust, sourceID, desc)
	} else {
		tx, err = r.deductUserEcoin(ctx, user, -delta, opsEcoinSourceAdjust, sourceID, desc)
	}
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	after, _ := r.getAvailableStock(ctx, uint64(user.Id), user.CompanyId)
	http_utils.WriteResponse(ctx, gin.H{
		"before_balance": current,
		"after_balance":  after,
		"delta":          delta,
		"transaction":    tx,
	}, nil)
}

// GiftEcoinRequest 运营赠送金币。
// 统一账号后 user_id 与 biz_user_id 都是主站用户 ID；前端可能只传其中一种。
type GiftEcoinRequest struct {
	UserId      any     `json:"user_id"`
	BizCode     string  `json:"biz_code"`
	BizUserId   any     `json:"biz_user_id"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description"`
}

func parseFlexibleUint64(v any) uint64 {
	if v == nil {
		return 0
	}
	switch t := v.(type) {
	case float64:
		if t <= 0 {
			return 0
		}
		return uint64(t)
	case int:
		if t <= 0 {
			return 0
		}
		return uint64(t)
	case int64:
		if t <= 0 {
			return 0
		}
		return uint64(t)
	case uint64:
		return t
	case json.Number:
		n, err := strconv.ParseUint(t.String(), 10, 64)
		if err != nil {
			return 0
		}
		return n
	case string:
		n, err := strconv.ParseUint(strings.TrimSpace(t), 10, 64)
		if err != nil {
			return 0
		}
		return n
	default:
		n, err := strconv.ParseUint(strings.TrimSpace(fmt.Sprint(t)), 10, 64)
		if err != nil {
			return 0
		}
		return n
	}
}

func (r *OpsResource) resolveOpsUser(ctx *gin.Context, userID uint64, bizCode string, bizUserID uint64) (*usermodel.User, error) {
	id := userID
	if id == 0 {
		id = bizUserID
	}
	if id == 0 {
		return nil, errors.New("user_id or biz_user_id is required")
	}
	if strings.TrimSpace(bizCode) != "" {
		return r.userService.GetUserByBiz(ctx.Request.Context(), bizCode, id)
	}
	return r.userService.GetUserById(ctx.Request.Context(), uint(id))
}

// GiftEcoin 赠送金币
// POST /ops/ecoin/gift
func (r *OpsResource) GiftEcoin(ctx *gin.Context) {
	var req GiftEcoinRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	user, err := r.resolveOpsUser(ctx, parseFlexibleUint64(req.UserId), req.BizCode, parseFlexibleUint64(req.BizUserId))
	if err != nil || user == nil {
		if err == nil {
			err = errors.New("user not found")
		}
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	desc := strings.TrimSpace(req.Description)
	if desc == "" {
		desc = fmt.Sprintf("运营赠送金币 %.2f", req.Amount)
	}
	sourceID := fmt.Sprintf("ops_gift:%d:%d", user.Id, time.Now().UnixNano())

	before, _ := r.getAvailableStock(ctx, uint64(user.Id), user.CompanyId)
	tx, err := r.addUserEcoin(ctx, user, req.Amount, opsEcoinSourceGift, sourceID, desc)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}
	after, _ := r.getAvailableStock(ctx, uint64(user.Id), user.CompanyId)
	http_utils.WriteResponse(ctx, gin.H{
		"before_balance": before,
		"after_balance":  after,
		"delta":          req.Amount,
		"transaction":    tx,
	}, nil)
}

// GiftMembershipRequest 运营赠送会员（包月 / 包年）。
// vip_role + period 会映射到商城会员 SKU，再走零元订单自动履约。
type GiftMembershipRequest struct {
	UserId    any    `json:"user_id"`
	BizCode   string `json:"biz_code"`
	BizUserId any    `json:"biz_user_id"`
	VipRole   string `json:"vip_role"`
	Period    string `json:"period"` // monthly | yearly
	SkuCode   string `json:"sku_code"`
	Remark    string `json:"remark"`
}

func membershipPeriodSuffix(period string) (string, string, error) {
	switch strings.ToLower(strings.TrimSpace(period)) {
	case "monthly", "month", "1m", "包月":
		return "_1M", "包月", nil
	case "yearly", "year", "12m", "annual", "包年":
		return "_12M", "包年", nil
	default:
		return "", "", errors.New("period 仅支持 monthly（包月）或 yearly（包年）")
	}
}

func membershipSkuCode(vipRole, suffix string) (string, error) {
	role := strings.ToUpper(strings.TrimSpace(vipRole))
	switch role {
	case "VIP_PRO":
		return "LAWMIND_VIP_PRO" + suffix, nil
	case "VIP_MAX":
		return "LAWMIND_VIP_MAX" + suffix, nil
	case "ENTERPRISE_BASIC":
		return "LAWMIND_ENTERPRISE_BASIC" + suffix, nil
	case "ENTERPRISE_STANDARD":
		return "LAWMIND_ENTERPRISE_STANDARD" + suffix, nil
	case "ENTERPRISE_PRO":
		return "LAWMIND_ENTERPRISE_PRO" + suffix, nil
	case "ENTERPRISE_FLAGSHIP":
		return "LAWMIND_ENTERPRISE_FLAGSHIP" + suffix, nil
	default:
		return "", fmt.Errorf("unsupported vip_role: %s", vipRole)
	}
}

func inferMembershipBizCode(vipRole, bizCode string) string {
	code := strings.TrimSpace(bizCode)
	if code != "" {
		return code
	}
	role := strings.ToUpper(strings.TrimSpace(vipRole))
	if strings.HasPrefix(role, "ENTERPRISE_") {
		return "LawMind_Enterprise"
	}
	return "LawMind_ToC"
}

func (r *OpsResource) findMembershipSku(ctx *gin.Context, bizCode, skuCode string) (*skumodel.Sku, error) {
	bizCode = strings.TrimSpace(bizCode)
	skuCode = strings.TrimSpace(skuCode)
	if bizCode == "" || skuCode == "" {
		return nil, errors.New("biz_code and sku_code are required")
	}
	found, err := r.skuService.GetSkuByCode(ctx.Request.Context(), bizCode, skuCode)
	if err == nil && found != nil {
		return found, nil
	}
	online := skumodel.SkuStatusOnline
	resp, listErr := r.skuService.ListSkus(ctx.Request.Context(), &sku.ListSkuRequest{
		BizCode: bizCode,
		Status:  &online,
		Limit:   100,
	})
	if listErr != nil {
		if err != nil {
			return nil, err
		}
		return nil, listErr
	}
	for _, item := range resp.List {
		if strings.EqualFold(strings.TrimSpace(item.SkuCode), skuCode) {
			return item, nil
		}
	}
	return nil, fmt.Errorf("未找到会员商品 %s / %s", bizCode, skuCode)
}

// GiftMembership 赠送会员
// POST /ops/membership/gift
func (r *OpsResource) GiftMembership(ctx *gin.Context) {
	var req GiftMembershipRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	user, err := r.resolveOpsUser(ctx, parseFlexibleUint64(req.UserId), req.BizCode, parseFlexibleUint64(req.BizUserId))
	if err != nil || user == nil {
		if err == nil {
			err = errors.New("user not found")
		}
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	skuCode := strings.TrimSpace(req.SkuCode)
	periodLabel := ""
	if skuCode == "" {
		suffix, label, periodErr := membershipPeriodSuffix(req.Period)
		if periodErr != nil {
			http_utils.WriteResponse(ctx, nil, periodErr)
			return
		}
		periodLabel = label
		vipRole := strings.TrimSpace(req.VipRole)
		if vipRole == "" {
			if strings.EqualFold(strings.TrimSpace(req.BizCode), "LawMind_Enterprise") {
				vipRole = "ENTERPRISE_BASIC"
			} else {
				vipRole = "VIP_PRO"
			}
		}
		skuCode, err = membershipSkuCode(vipRole, suffix)
		if err != nil {
			http_utils.WriteResponse(ctx, nil, err)
			return
		}
	}
	bizCode := inferMembershipBizCode(req.VipRole, req.BizCode)
	skuInfo, err := r.findMembershipSku(ctx, bizCode, skuCode)
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	remark := strings.TrimSpace(req.Remark)
	if remark == "" {
		if periodLabel != "" {
			remark = fmt.Sprintf("运营赠送%s会员 %s", periodLabel, skuInfo.SkuName)
		} else {
			remark = fmt.Sprintf("运营赠送会员 %s", skuInfo.SkuName)
		}
	}

	resp, err := r.orderService.GiftOrder(ctx.Request.Context(), &order.GiftOrderRequest{
		UserId: uint64(user.Id),
		SkuItems: []*order.SkuOrderItem{{
			SkuId:    skuInfo.Id,
			Quantity: 1,
		}},
		Remark: remark,
	})
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	orderNo := ""
	if resp != nil && resp.Order != nil {
		orderNo = resp.Order.OrderNo
	}
	http_utils.WriteResponse(ctx, gin.H{
		"order":    resp.Order,
		"order_no": orderNo,
		"sku_id":   skuInfo.Id,
		"sku_code": strings.TrimSpace(skuInfo.SkuCode),
		"sku_name": skuInfo.SkuName,
		"user_id":  user.Id,
		"biz_code": bizCode,
	}, nil)
}

// GiftSkuRequest 运营赠送商品
type GiftSkuRequest struct {
	UserId   uint64 `json:"user_id" binding:"required"`
	SkuId    uint64 `json:"sku_id" binding:"required"`
	Quantity int    `json:"quantity" binding:"required,min=1"`
	Remark   string `json:"remark"`
}

// GiftSku 赠送商品（零元订单 + 自动履约）
// POST /ops/orders/gift
func (r *OpsResource) GiftSku(ctx *gin.Context) {
	var req GiftSkuRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	resp, err := r.orderService.GiftOrder(ctx.Request.Context(), &order.GiftOrderRequest{
		UserId: req.UserId,
		SkuItems: []*order.SkuOrderItem{{
			SkuId:    req.SkuId,
			Quantity: req.Quantity,
		}},
		Remark: req.Remark,
	})
	if err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}
	http_utils.WriteResponse(ctx, resp.Order, nil)
}

func (r *OpsResource) getAvailableStock(ctx *gin.Context, userID uint64, companyID uint64) (float64, error) {
	if companyID > 0 {
		ce, err := r.companyEcoinService.GetCompanyEcoin(ctx.Request.Context(), companyID)
		if err != nil {
			return 0, err
		}
		if ce == nil {
			return 0, nil
		}
		return ce.AvailableStock, nil
	}
	info, err := r.ecoinService.GetUserEcoin(ctx.Request.Context(), userID)
	if err != nil {
		return 0, err
	}
	if info == nil {
		return 0, nil
	}
	return info.AvailableStock, nil
}

func (r *OpsResource) addUserEcoin(ctx *gin.Context, user *usermodel.User, amount float64, sourceType, sourceID, desc string) (interface{}, error) {
	if user.CompanyId > 0 {
		return r.companyEcoinService.AddCompanyEcoin(ctx.Request.Context(), &companyecoin.AddCompanyEcoinRequest{
			CompanyId:      user.CompanyId,
			OperatorUserId: uint64(user.Id),
			Amount:         amount,
			SourceType:     sourceType,
			SourceId:       sourceID,
			Description:    desc,
		})
	}
	return r.ecoinService.AddEcoin(ctx.Request.Context(), &ecoin.AddEcoinRequest{
		UserId:      uint64(user.Id),
		Amount:      amount,
		SourceType:  sourceType,
		SourceId:    sourceID,
		Description: desc,
	})
}

func (r *OpsResource) deductUserEcoin(ctx *gin.Context, user *usermodel.User, amount float64, sourceType, sourceID, desc string) (interface{}, error) {
	if amount <= 0 {
		return nil, errors.New("deduct amount must be positive")
	}
	if user.CompanyId > 0 {
		return r.companyEcoinService.DeductCompanyEcoin(ctx.Request.Context(), &companyecoin.DeductCompanyEcoinRequest{
			CompanyId:      user.CompanyId,
			OperatorUserId: uint64(user.Id),
			Amount:         amount,
			SourceType:     sourceType,
			SourceId:       sourceID,
			Description:    desc,
		})
	}
	return r.ecoinService.DeductEcoin(ctx.Request.Context(), &ecoin.DeductEcoinRequest{
		UserId:      uint64(user.Id),
		Amount:      amount,
		SourceType:  sourceType,
		SourceId:    sourceID,
		Description: desc,
	})
}
