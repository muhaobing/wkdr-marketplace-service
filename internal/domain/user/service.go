package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/muhaobing/std-go/go-common/cache"
	"github.com/muhaobing/std-go/go-common/database"

	"wdkr-marketplace-service/internal/common/config"
	"wdkr-marketplace-service/internal/common/utils/auth_utils"
	"wdkr-marketplace-service/internal/domain/ecoin"
	"wdkr-marketplace-service/internal/domain/user/repo"
	usermodel "wdkr-marketplace-service/internal/domain/user/user_model"
)

type userServiceImpl struct {
	userRepo    repo.UserRepo
	bindingRepo repo.UserBindingRepo
	ecoinSvc    ecoin.EcoinService
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo repo.UserRepo, bindingRepo repo.UserBindingRepo, ecoinSvc ecoin.EcoinService) UserService {
	return &userServiceImpl{
		userRepo:    userRepo,
		bindingRepo: bindingRepo,
		ecoinSvc:    ecoinSvc,
	}
}

// BindUser 绑定用户（包含注册逻辑）；成功后签发 session，返回 token
func (s *userServiceImpl) BindUser(ctx context.Context, req *BindUserRequest) (*BindUserResponse, error) {
	if req.BizCode == "" {
		return nil, errors.New("biz_code is required")
	}
	if req.BizUserId == 0 {
		return nil, errors.New("biz_user_id is required")
	}
	if req.TelNo == "" && req.Email == "" {
		return nil, errors.New("tel_no or email is required")
	}
	if req.Password == "" {
		return nil, errors.New("password is required")
	}

	existingBinding, err := s.bindingRepo.GetBindingByBiz(ctx, req.BizCode, req.BizUserId)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing binding: %w", err)
	}
	if existingBinding != nil {
		user, err := s.userRepo.GetUserById(ctx, existingBinding.UserId)
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}
		if user == nil {
			return nil, errors.New("user not found")
		}
		if !user.VerifySecretKey(req.Password) {
			return nil, errors.New("invalid password")
		}
		loginResp, err := s.generateLoginResponse(ctx, user)
		if err != nil {
			return nil, err
		}
		return &BindUserResponse{
			UserId:    user.Id,
			IsNewUser: false,
			Token:     loginResp.Token,
			User:      loginResp.User,
		}, nil
	}

	var response *BindUserResponse
	err = database.Transaction(ctx, func(ctx context.Context) error {
		user, err := s.findUserByIdentity(ctx, req.TelNo, req.Email)
		if err != nil {
			return fmt.Errorf("failed to find user by identity: %w", err)
		}

		isNewUser := false
		if user == nil {
			user, err = s.createUser(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}
			isNewUser = true
			if _, err = s.ecoinSvc.InitUserEcoin(ctx, uint64(user.Id)); err != nil {
				return fmt.Errorf("failed to init user ecoin: %w", err)
			}
		} else {
			if !user.VerifySecretKey(req.Password) {
				return errors.New("invalid password")
			}
		}

		binding := &usermodel.UserBinding{
			UserId:    user.Id,
			BizCode:   req.BizCode,
			BizUserId: req.BizUserId,
		}
		if err = s.bindingRepo.CreateBinding(ctx, binding); err != nil {
			return fmt.Errorf("failed to create binding: %w", err)
		}

		response = &BindUserResponse{
			UserId:    user.Id,
			IsNewUser: isNewUser,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetUserById(ctx, response.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	loginResp, err := s.generateLoginResponse(ctx, user)
	if err != nil {
		return nil, err
	}
	response.Token = loginResp.Token
	response.User = loginResp.User
	return response, nil
}

// UnbindUser 解绑用户
func (s *userServiceImpl) UnbindUser(ctx context.Context, req *UnbindUserRequest) error {
	// 参数校验
	if req.UserId == 0 {
		return errors.New("user_id is required")
	}
	if req.BizCode == "" {
		return errors.New("biz_code is required")
	}

	// 检查绑定是否存在
	binding, err := s.bindingRepo.GetBindingByUserAndBiz(ctx, req.UserId, req.BizCode)
	if err != nil {
		return fmt.Errorf("failed to get binding: %w", err)
	}
	if binding == nil {
		return fmt.Errorf("user not bindded to biz_code %s", req.BizCode)
	}

	// 删除绑定记录
	if err = s.bindingRepo.DeleteBinding(ctx, req.UserId, req.BizCode); err != nil {
		return fmt.Errorf("failed to delete binding: %w", err)
	}

	return nil
}

// GetUserById 根据ID获取用户信息
func (s *userServiceImpl) GetUserById(ctx context.Context, id uint) (*usermodel.User, error) {
	if id == 0 {
		return nil, errors.New("user id is required")
	}

	user, err := s.userRepo.GetUserById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// GetUserByIdentity 根据身份标识获取用户（优先手机号，其次邮箱）
func (s *userServiceImpl) GetUserByIdentity(ctx context.Context, telNo, email string) (*usermodel.User, error) {
	if telNo == "" && email == "" {
		return nil, errors.New("tel_no or email is required")
	}

	return s.findUserByIdentity(ctx, telNo, email)
}

// GetUserByBiz 根据业务信息获取用户
func (s *userServiceImpl) GetUserByBiz(ctx context.Context, bizCode string, bizUserId uint64) (*usermodel.User, error) {
	if bizCode == "" {
		return nil, errors.New("biz_code is required")
	}
	if bizUserId == 0 {
		return nil, errors.New("biz_user_id is required")
	}

	binding, err := s.bindingRepo.GetBindingByBiz(ctx, bizCode, bizUserId)
	if err != nil {
		return nil, fmt.Errorf("failed to get binding: %w", err)
	}
	if binding == nil {
		return nil, nil
	}

	return s.userRepo.GetUserById(ctx, binding.UserId)
}

// VerifyUserSecret 验证用户密钥
func (s *userServiceImpl) VerifyUserSecret(ctx context.Context, userId uint, secret string) (bool, error) {
	if userId == 0 {
		return false, errors.New("user id is required")
	}
	if secret == "" {
		return false, errors.New("secret is required")
	}

	user, err := s.userRepo.GetUserById(ctx, userId)
	if err != nil {
		return false, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return false, errors.New("user not found")
	}

	return user.VerifySecretKey(secret), nil
}

// Login 用户登录（电话号码/邮箱）
func (s *userServiceImpl) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// 参数校验
	if req.TelNo == "" && req.Email == "" {
		return nil, errors.New("tel_no or email is required")
	}
	if req.Secret == "" {
		return nil, errors.New("secret is required")
	}

	// 查找用户
	user, err := s.findUserByIdentity(ctx, req.TelNo, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// 验证密钥
	if !user.VerifySecretKey(req.Secret) {
		return nil, errors.New("invalid secret")
	}

	// 生成登录响应
	return s.generateLoginResponse(ctx, user)
}

// GetBindingsByUserId 获取用户所有绑定信息
func (s *userServiceImpl) GetBindingsByUserId(ctx context.Context, userId uint) ([]*usermodel.UserBinding, error) {
	if userId == 0 {
		return nil, errors.New("user_id is required")
	}
	return s.bindingRepo.GetBindingsByUserId(ctx, userId)
}

// generateLoginResponse 生成登录响应（生成session、token并写入redis）
func (s *userServiceImpl) generateLoginResponse(ctx context.Context, user *usermodel.User) (*LoginResponse, error) {
	// 生成 session id: "session:$user_id:$login_timestamp"
	loginTimestamp := time.Now().Unix()
	sessionId := fmt.Sprintf("session:%d:%d", user.Id, loginTimestamp)

	// 生成 session: user 序列化为 JSON
	session, err := user.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize user: %w", err)
	}

	// 获取配置
	conf := config.GetConf()
	if conf == nil {
		return nil, errors.New("config not initialized")
	}

	// 生成 token
	token, err := auth_utils.GenAuthToken(sessionId, conf.Auth.AesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// 将 session 写入 redis
	expiration := time.Duration(conf.Auth.Expiration) * time.Second
	redis := cache.FromContext(ctx)
	if redis == nil {
		return nil, errors.New("redis client not initialized")
	}
	if err = redis.Set(ctx, sessionId, session, expiration).Err(); err != nil {
		return nil, fmt.Errorf("failed to save session to redis: %w", err)
	}

	return &LoginResponse{
		Token:  token,
		UserId: user.Id,
		User:   user,
	}, nil
}

// findUserByIdentity 根据身份标识查找用户（优先手机号，其次邮箱）
func (s *userServiceImpl) findUserByIdentity(ctx context.Context, telNo, email string) (*usermodel.User, error) {
	// 优先通过手机号查找
	if telNo != "" {
		user, err := s.userRepo.GetUserByTelNo(ctx, telNo)
		if err != nil {
			return nil, err
		}
		if user != nil {
			return user, nil
		}
	}

	// 其次通过邮箱查找
	if email != "" {
		user, err := s.userRepo.GetUserByEmail(ctx, email)
		if err != nil {
			return nil, err
		}
		if user != nil {
			return user, nil
		}
	}

	return nil, nil
}

// createUser 创建新用户
func (s *userServiceImpl) createUser(ctx context.Context, req *BindUserRequest) (*usermodel.User, error) {
	user := &usermodel.User{
		TelNo: req.TelNo,
		Email: req.Email,
		Role:  usermodel.RoleUser, // 默认为普通用户
	}

	// 先创建用户以获取ID
	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	// 生成并更新密钥（基于 password + id）
	secretKey := usermodel.GenerateSecretKey(req.Password, user.Id)
	if err := s.userRepo.UpdateUserSecretKey(ctx, user.Id, secretKey); err != nil {
		return nil, err
	}
	user.SecretKey = secretKey

	return user, nil
}
