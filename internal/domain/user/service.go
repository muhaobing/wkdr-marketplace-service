package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/muhaobing-eng/std-go/go-common/database"

	"wdkr-marketplace-service/internal/domain/ecoin"
	"wdkr-marketplace-service/internal/domain/user/repo"
	usermodel "wdkr-marketplace-service/internal/domain/user/user_model"
)

type userServiceImpl struct {
	userRepo repo.UserRepo
	ecoinSvc ecoin.EcoinService
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo repo.UserRepo, ecoinSvc ecoin.EcoinService) UserService {
	return &userServiceImpl{
		userRepo: userRepo,
		ecoinSvc: ecoinSvc,
	}
}

// BindUser 绑定用户（包含注册逻辑）
func (s *userServiceImpl) BindUser(ctx context.Context, req *BindUserRequest) (*BindUserResponse, error) {
	// 参数校验
	if req.BizCode == "" {
		return nil, errors.New("biz_code is required")
	}
	if req.BizId == 0 {
		return nil, errors.New("biz_id is required")
	}
	if req.TelNo == "" && req.Email == "" {
		return nil, errors.New("tel_no or email is required")
	}
	if req.Secret == "" {
		return nil, errors.New("secret is required")
	}

	var response *BindUserResponse
	err := database.Transaction(ctx, func(ctx context.Context) error {
		// 根据身份标识查找用户（优先手机号，其次邮箱）
		user, err := s.findUserByIdentity(ctx, req.TelNo, req.Email)
		if err != nil {
			return fmt.Errorf("failed to find user by identity: %w", err)
		}

		isNewUser := false
		if user == nil {
			// 用户不存在，创建新用户
			user, err = s.createUser(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}
			isNewUser = true

			// 为新用户初始化积分账户
			if _, err = s.ecoinSvc.InitUserEcoin(ctx, uint64(user.Id)); err != nil {
				return fmt.Errorf("failed to init user ecoin: %w", err)
			}
		}

		// 检查是否已绑定该业务平台
		if user.HasBinding(req.BizCode) {
			existingBinding, _ := user.GetBinding(req.BizCode)
			if existingBinding.BizId == req.BizId {
				// 已绑定相同的业务ID，直接返回
				response = &BindUserResponse{
					UserId:    user.Id,
					IsNewUser: false,
				}
				return nil
			}
			return fmt.Errorf("user already bindded to biz_code %s with different biz_id", req.BizCode)
		}

		// 绑定业务平台
		user.Bind(req.BizCode, usermodel.BindingInfo{BizId: req.BizId})
		if err = s.userRepo.UpdateUserBinding(ctx, user.Id, user.Binding); err != nil {
			return fmt.Errorf("failed to update user binding: %w", err)
		}

		response = &BindUserResponse{
			UserId:    user.Id,
			IsNewUser: isNewUser,
		}
		// 新用户返回密钥
		if isNewUser {
			response.SecretKey = user.SecretKey
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

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

	return database.Transaction(ctx, func(ctx context.Context) error {
		// 获取用户信息（加锁）
		user, err := s.userRepo.GetUserForUpdate(ctx, req.UserId)
		if err != nil {
			return fmt.Errorf("failed to get user for update: %w", err)
		}
		if user == nil {
			return errors.New("user not found")
		}

		// 检查是否已绑定该业务平台
		if !user.HasBinding(req.BizCode) {
			return fmt.Errorf("user not bindded to biz_code %s", req.BizCode)
		}

		// 解绑业务平台
		user.Unbind(req.BizCode)
		if err = s.userRepo.UpdateUserBinding(ctx, user.Id, user.Binding); err != nil {
			return fmt.Errorf("failed to update user binding: %w", err)
		}

		return nil
	})
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
		TelNo:   req.TelNo,
		Email:   req.Email,
		Binding: make(usermodel.BindingMap),
	}

	// 先创建用户以获取ID
	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	// 生成并更新密钥（基于secret + id）
	secretKey := usermodel.GenerateSecretKey(req.Secret, user.Id)
	if err := s.userRepo.UpdateUserSecretKey(ctx, user.Id, secretKey); err != nil {
		return nil, err
	}
	user.SecretKey = secretKey

	return user, nil
}
