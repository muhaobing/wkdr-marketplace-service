package user

import (
	"context"
	"errors"
	"strings"

	"github.com/muhaobing/std-go/go-common/database"

	bizcoderepo "wdkr-marketplace-service/internal/domain/bizcode/repo"
	companyrepo "wdkr-marketplace-service/internal/domain/company/repo"
	"wdkr-marketplace-service/internal/domain/companyecoin"
	"wdkr-marketplace-service/internal/domain/ecoin"
	"wdkr-marketplace-service/internal/domain/user/repo"
	usermodel "wdkr-marketplace-service/internal/domain/user/user_model"
)

// 统一账号说明：
// 商城不再维护独立的 user_tab / user_binding_tab。
// 主站(law 库)的 personal_user / company_user 即唯一用户表，
// 商城的所有业务数据（积分/订单/购物车等）直接以主站 userId / companyId 为键。
// 本服务退化为一层“身份解析器”：把主站 userId 解析为商城所需的 User 视图，
// 企业归属(companyId)来源于主站 law.company_user。

const (
	bizCodeLawMindToC        = "LawMind_ToC"
	bizCodeLawMindEnterprise = "LawMind_Enterprise"
)

// errUserMgmtDisabled 商城自有账号管理（注册/登录/改密/绑定）已随统一账号下线。
var errUserMgmtDisabled = errors.New("商城已与主站统一账号，独立用户登录/注册/绑定能力已下线")

type userServiceImpl struct {
	userRepo     repo.UserRepo
	bindingRepo  repo.UserBindingRepo
	ecoinSvc     ecoin.EcoinService
	companyEcoin companyecoin.CompanyEcoinService
	companyRepo  companyrepo.CompanyRepo
	bizCodeRepo  bizcoderepo.BizCodeRepo
}

// NewUserService 创建用户服务实例（构造签名保持不变，便于 wire 注入；多数依赖在统一账号模式下不再使用）
func NewUserService(
	userRepo repo.UserRepo,
	bindingRepo repo.UserBindingRepo,
	ecoinSvc ecoin.EcoinService,
	companyEcoin companyecoin.CompanyEcoinService,
	companyRepo companyrepo.CompanyRepo,
	bizCodeRepo bizcoderepo.BizCodeRepo,
) UserService {
	return &userServiceImpl{
		userRepo:     userRepo,
		bindingRepo:  bindingRepo,
		ecoinSvc:     ecoinSvc,
		companyEcoin: companyEcoin,
		companyRepo:  companyRepo,
		bizCodeRepo:  bizCodeRepo,
	}
}

// lawCompanyIdByUserId 解析主站用户所属企业 ID（company_user 优先，其次 operation_user.company_id）。
func lawCompanyIdByUserId(ctx context.Context, userId uint64) (uint64, error) {
	if userId == 0 {
		return 0, nil
	}
	var companyId uint64
	err := database.FromContext(ctx).
		Raw("SELECT company_id FROM law.company_user WHERE user_id = ? AND is_deleted = 0 LIMIT 1", userId).
		Scan(&companyId).Error
	if err != nil {
		return 0, err
	}
	if companyId != 0 {
		return companyId, nil
	}
	err = database.FromContext(ctx).
		Raw("SELECT company_id FROM law.operation_user WHERE user_id = ? AND is_deleted = 0 AND company_id IS NOT NULL LIMIT 1", userId).
		Scan(&companyId).Error
	if err != nil {
		return 0, err
	}
	return companyId, nil
}

// lawCompanyName 主站 law.company 企业名称（仅用于展示）。
func lawCompanyName(ctx context.Context, companyId uint64) string {
	if companyId == 0 {
		return ""
	}
	var name string
	_ = database.FromContext(ctx).
		Raw("SELECT company_name FROM law.company WHERE id = ? LIMIT 1", companyId).
		Scan(&name).Error
	return name
}

// buildIdentity 以主站 userId / companyId 构造商城所需的 User 视图（不落库）。
func (s *userServiceImpl) buildIdentity(ctx context.Context, userId uint64, companyId uint64) *usermodel.User {
	u := &usermodel.User{
		Id:        uint(userId),
		CompanyId: companyId,
		Role:      usermodel.RoleUser,
	}
	if companyId > 0 {
		u.CompanyName = lawCompanyName(ctx, companyId)
	}
	return u
}

// ResolveOrCreateByBiz 主站 JWT 身份 → 商城身份。
// 统一账号后：直接采用主站 userId 作为商城用户标识，companyId 取自 JWT（缺省时回查 law.company_user）。
func (s *userServiceImpl) ResolveOrCreateByBiz(ctx context.Context, req *ResolveOrCreateByBizRequest) (*usermodel.User, error) {
	if req == nil || req.BizUserId == 0 {
		return nil, errors.New("biz_user_id is required")
	}
	companyId := req.MainCompanyId
	if companyId == 0 && req.BizCode == bizCodeLawMindEnterprise {
		if c, err := lawCompanyIdByUserId(ctx, req.BizUserId); err == nil {
			companyId = c
		}
	}
	return s.buildIdentity(ctx, req.BizUserId, companyId), nil
}

// GetUserById 以主站 userId 解析商城用户视图（企业归属来自 law.company_user）。
func (s *userServiceImpl) GetUserById(ctx context.Context, id uint) (*usermodel.User, error) {
	if id == 0 {
		return nil, errors.New("user id is required")
	}
	companyId, err := lawCompanyIdByUserId(ctx, uint64(id))
	if err != nil {
		return nil, err
	}
	return s.buildIdentity(ctx, uint64(id), companyId), nil
}

// GetUserByBiz 统一账号下 biz_user_id 即主站 userId。
// 个人 biz_code 强制走个人积分账户；企业 biz_code 取 law.company_user 的 companyId。
func (s *userServiceImpl) GetUserByBiz(ctx context.Context, bizCode string, bizUserId uint64) (*usermodel.User, error) {
	if bizUserId == 0 {
		return nil, errors.New("biz_user_id is required")
	}
	if strings.TrimSpace(bizCode) == bizCodeLawMindToC {
		return s.buildIdentity(ctx, bizUserId, 0), nil
	}
	companyId, err := lawCompanyIdByUserId(ctx, bizUserId)
	if err != nil {
		return nil, err
	}
	return s.buildIdentity(ctx, bizUserId, companyId), nil
}

// GetUserByIdentity 运营按手机号查询主站用户（个人优先，其次企业员工）。
func (s *userServiceImpl) GetUserByIdentity(ctx context.Context, telNo, email string, companyId uint64) (*usermodel.User, error) {
	tel := strings.TrimSpace(telNo)
	if tel == "" {
		return nil, errors.New("tel_no is required")
	}
	db := database.FromContext(ctx)

	var personalUserId uint64
	if err := db.Raw(
		"SELECT user_id FROM law.personal_user WHERE phone = ? AND is_deleted = 0 LIMIT 1", tel,
	).Scan(&personalUserId).Error; err != nil {
		return nil, err
	}
	if personalUserId != 0 {
		return s.buildIdentity(ctx, personalUserId, 0), nil
	}

	var row struct {
		UserId    uint64
		CompanyId uint64
	}
	var err error
	if companyId > 0 {
		err = db.Raw(
			"SELECT user_id AS user_id, company_id AS company_id FROM law.company_user WHERE phone = ? AND company_id = ? AND is_deleted = 0 LIMIT 1",
			tel, companyId,
		).Scan(&row).Error
	} else {
		err = db.Raw(
			"SELECT user_id AS user_id, company_id AS company_id FROM law.company_user WHERE phone = ? AND is_deleted = 0 LIMIT 1",
			tel,
		).Scan(&row).Error
	}
	if err != nil {
		return nil, err
	}
	if row.UserId != 0 {
		return s.buildIdentity(ctx, row.UserId, row.CompanyId), nil
	}
	return nil, nil
}

// GetBindingsByUserId 统一账号下不再有绑定表；返回由主站身份推导出的单条“虚拟绑定”，
// 供订单履约回调时确定 biz_code + biz_user_id。
func (s *userServiceImpl) GetBindingsByUserId(ctx context.Context, userId uint) ([]*usermodel.UserBinding, error) {
	if userId == 0 {
		return nil, errors.New("user_id is required")
	}
	companyId, err := lawCompanyIdByUserId(ctx, uint64(userId))
	if err != nil {
		return nil, err
	}
	code := bizCodeLawMindToC
	if companyId > 0 {
		code = bizCodeLawMindEnterprise
	}
	return []*usermodel.UserBinding{{
		UserId:    userId,
		BizCode:   code,
		BizUserId: uint64(userId),
	}}, nil
}

// ---- 以下为商城自有账号管理接口，统一账号后全部下线 ----

func (s *userServiceImpl) BindUser(ctx context.Context, req *BindUserRequest) (*BindUserResponse, error) {
	return nil, errUserMgmtDisabled
}

func (s *userServiceImpl) UnbindUser(ctx context.Context, req *UnbindUserRequest) error {
	return errUserMgmtDisabled
}

func (s *userServiceImpl) VerifyUserSecret(ctx context.Context, userId uint, secret string) (bool, error) {
	return false, errUserMgmtDisabled
}

func (s *userServiceImpl) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	return nil, errUserMgmtDisabled
}

func (s *userServiceImpl) ChangePassword(ctx context.Context, userId uint, req *ChangePasswordRequest) error {
	return errUserMgmtDisabled
}

func (s *userServiceImpl) UpdateProfile(ctx context.Context, userId uint, req *UpdateProfileRequest, sessionId string) (*usermodel.User, error) {
	return nil, errUserMgmtDisabled
}
