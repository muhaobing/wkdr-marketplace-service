package user

import (
	"context"

	usermodel "wdkr-marketplace-service/internal/domain/user/user_model"
)

// BindUserRequest 用户绑定请求
type BindUserRequest struct {
	Password  string `json:"password"`    // 用户密码（与商城登录一致，用于校验并生成 secret_key）
	BizCode   string `json:"biz_code"`    // 业务平台代码
	BizUserId uint64 `json:"biz_user_id"` // 业务平台用户ID
	TelNo     string `json:"tel_no"`      // 手机号（优先）
	Email     string `json:"email"`       // 邮箱（tel_no为空时使用）
}

// BindUserResponse 用户绑定响应（绑定成功后即完成登录，返回 session token）
type BindUserResponse struct {
	UserId    uint                       `json:"user_id"`     // 商城中心用户ID
	IsNewUser bool                       `json:"is_new_user"` // 是否为新创建的用户
	Token     string                     `json:"token"`       // session token（Authorization: Bearer）
	User      *usermodel.User            `json:"user"`        // 用户信息（不含 secret_key）
	Bindings  []*usermodel.UserBinding   `json:"bindings"`    // 业务平台绑定列表（与 GET /user/bindings 一致）
}

// UnbindUserRequest 用户解绑请求
type UnbindUserRequest struct {
	UserId  uint   `json:"user_id"`  // 商城中心用户ID
	BizCode string `json:"biz_code"` // 业务平台代码
}

// GetUserByBizRequest 根据业务信息获取用户请求
type GetUserByBizRequest struct {
	BizCode string `json:"biz_code"` // 业务平台代码
	BizId   uint64 `json:"biz_id"`   // 业务平台用户ID
}

// LoginRequest 登录请求（电话号码/邮箱登录）
type LoginRequest struct {
	TelNo  string `json:"tel_no"` // 手机号
	Email  string `json:"email"`  // 邮箱
	Secret string `json:"secret"` // 用户密钥
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token    string                   `json:"token"`    // 登录token
	UserId   uint                     `json:"user_id"`  // 用户ID
	User     *usermodel.User          `json:"user"`     // 用户信息
	Bindings []*usermodel.UserBinding `json:"bindings"` // 业务平台绑定列表（与 GET /user/bindings 一致）
}

// ChangePasswordRequest 修改登录密钥（需验证原密码）
type ChangePasswordRequest struct {
	OldSecret string `json:"old_secret"` // 原密码（与登录密钥一致）
	NewSecret string `json:"new_secret"` // 新密码
}

// UserService 用户服务接口
type UserService interface {
	// BindUser 绑定用户（包含注册逻辑）
	// 如果根据手机号/邮箱找到已有用户，则直接绑定
	// 如果没有找到用户，则创建新用户并绑定，同时初始化积分账户
	BindUser(ctx context.Context, req *BindUserRequest) (*BindUserResponse, error)

	// UnbindUser 解绑用户
	// 移除用户与指定业务平台的绑定关系
	UnbindUser(ctx context.Context, req *UnbindUserRequest) error

	// GetUserById 根据ID获取用户信息
	GetUserById(ctx context.Context, id uint) (*usermodel.User, error)

	// GetUserByIdentity 根据身份标识获取用户（优先手机号，其次邮箱）
	GetUserByIdentity(ctx context.Context, telNo, email string) (*usermodel.User, error)

	// GetUserByBiz 根据业务信息获取用户
	GetUserByBiz(ctx context.Context, bizCode string, bizUserId uint64) (*usermodel.User, error)

	// VerifyUserSecret 验证用户密钥
	VerifyUserSecret(ctx context.Context, userId uint, secret string) (bool, error)

	// Login 用户登录（电话号码/邮箱）
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)

	// GetBindingsByUserId 获取用户所有绑定信息
	GetBindingsByUserId(ctx context.Context, userId uint) ([]*usermodel.UserBinding, error)

	// ChangePassword 修改登录密钥
	ChangePassword(ctx context.Context, userId uint, req *ChangePasswordRequest) error
}
