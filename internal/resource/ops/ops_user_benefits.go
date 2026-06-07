package ops

import (
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
	usermodel "wdkr-marketplace-service/internal/domain/user/user_model"
)

const (
	opsEcoinSourceAdjust = "admin"
	opsEcoinSourceGift   = "system"
)

// LookupUser 查询商城用户及积分概况
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

// AdjustEcoinRequest 运营调整积分
type AdjustEcoinRequest struct {
	UserId        uint64  `json:"user_id" binding:"required"`
	Mode          string  `json:"mode" binding:"required"` // set | add | deduct
	Amount        float64 `json:"amount"`
	TargetBalance float64 `json:"target_balance"`
	Description   string  `json:"description"`
}

// AdjustEcoin 直接调整用户积分（设为指定值 / 增减）
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
			desc = fmt.Sprintf("运营调整积分至 %.2f", req.TargetBalance)
		} else if delta > 0 {
			desc = fmt.Sprintf("运营增加积分 %.2f", delta)
		} else {
			desc = fmt.Sprintf("运营扣减积分 %.2f", -delta)
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

// GiftEcoinRequest 运营赠送积分
type GiftEcoinRequest struct {
	UserId      uint64  `json:"user_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description"`
}

// GiftEcoin 赠送积分
// POST /ops/ecoin/gift
func (r *OpsResource) GiftEcoin(ctx *gin.Context) {
	var req GiftEcoinRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		http_utils.WriteResponse(ctx, nil, err)
		return
	}

	user, err := r.userService.GetUserById(ctx.Request.Context(), uint(req.UserId))
	if err != nil || user == nil {
		http_utils.WriteResponse(ctx, nil, errors.New("user not found"))
		return
	}

	desc := strings.TrimSpace(req.Description)
	if desc == "" {
		desc = fmt.Sprintf("运营赠送积分 %.2f", req.Amount)
	}
	sourceID := fmt.Sprintf("ops_gift:%d:%d", req.UserId, time.Now().UnixNano())

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
