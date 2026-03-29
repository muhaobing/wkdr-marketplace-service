package repo

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/muhaobing/std-go/go-common/database"
	"gorm.io/gorm"

	usermodel "wdkr-marketplace-service/internal/domain/user/user_model"
)

type userRepoImpl struct{}

// NewUserRepo 创建用户仓储实现
func NewUserRepo() UserRepo {
	return &userRepoImpl{}
}

// GetUserById 根据ID获取用户
func (r *userRepoImpl) GetUserById(ctx context.Context, id uint) (*usermodel.User, error) {
	var user usermodel.User
	err := database.FromContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByTelNoAndCompany 根据手机号 + company_id 获取用户
func (r *userRepoImpl) GetUserByTelNoAndCompany(ctx context.Context, telNo string, companyId uint64) (*usermodel.User, error) {
	if telNo == "" {
		return nil, nil
	}
	var user usermodel.User
	err := database.FromContext(ctx).Where("tel_no = ? AND company_id = ?", telNo, companyId).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByEmailAndCompany 根据邮箱 + company_id 获取用户
func (r *userRepoImpl) GetUserByEmailAndCompany(ctx context.Context, email string, companyId uint64) (*usermodel.User, error) {
	if email == "" {
		return nil, nil
	}
	var user usermodel.User
	err := database.FromContext(ctx).Where("email = ? AND company_id = ?", email, companyId).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// nullableContactArg 空串写入 NULL，避免 uk_company_tel / uk_company_email 与空串冲突
func nullableContactArg(s string) interface{} {
	t := strings.TrimSpace(s)
	if t == "" {
		return nil
	}
	return t
}

// CreateUser 创建用户（tel_no / email 仅非空时写入，否则为 NULL）
func (r *userRepoImpl) CreateUser(ctx context.Context, user *usermodel.User) error {
	db := database.FromContext(ctx)
	telArg := nullableContactArg(user.TelNo)
	emailArg := nullableContactArg(user.Email)
	now := uint32(time.Now().Unix())
	var uid uint64
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			`INSERT INTO `+usermodel.UserTabName+` (company_id, role, tel_no, email, ctime, mtime) VALUES (?, ?, ?, ?, ?, ?)`,
			user.CompanyId, user.Role, telArg, emailArg, now, now,
		).Error; err != nil {
			return err
		}
		return tx.Raw("SELECT LAST_INSERT_ID()").Scan(&uid).Error
	})
	if err != nil {
		return err
	}
	user.Id = uint(uid)
	return nil
}

// UpdateUserSecretKey 更新用户密钥
func (r *userRepoImpl) UpdateUserSecretKey(ctx context.Context, id uint, secretKey string) error {
	return database.FromContext(ctx).Model(&usermodel.User{}).
		Where("id = ?", id).
		Update("secret_key", secretKey).Error
}

// UpdateUserContact 更新手机号、邮箱（空串写入 NULL）
func (r *userRepoImpl) UpdateUserContact(ctx context.Context, id uint, telNo, email string) error {
	return database.FromContext(ctx).Model(&usermodel.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"tel_no": nullableContactArg(telNo),
			"email":  nullableContactArg(email),
			"mtime":  uint32(time.Now().Unix()),
		}).Error
}

// GetUserForUpdate 获取用户信息（加锁）
func (r *userRepoImpl) GetUserForUpdate(ctx context.Context, id uint) (*usermodel.User, error) {
	var user usermodel.User
	err := database.FromContext(ctx).Set("gorm:query_option", "FOR UPDATE").Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}
