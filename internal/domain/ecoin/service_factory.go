package ecoin

import (
	"wdkr-marketplace-service/internal/domain/ecoin/repo"
)

// ServiceFactory 积分服务工厂
type ServiceFactory struct {
	ecoinRepo    repo.EcoinRepo
	ecoinService EcoinService
}

// NewServiceFactory 创建积分服务工厂
func NewServiceFactory() *ServiceFactory {
	// 创建repo实例
	ecoinRepo := repo.NewEcoinRepo()

	// 创建service实例，注入repo依赖
	ecoinService := NewEcoinService(ecoinRepo)

	return &ServiceFactory{
		ecoinRepo:    ecoinRepo,
		ecoinService: ecoinService,
	}
}

// GetEcoinService 获取积分服务实例
func (f *ServiceFactory) GetEcoinService() EcoinService {
	return f.ecoinService
}

// GetEcoinRepo 获取积分仓储实例
func (f *ServiceFactory) GetEcoinRepo() repo.EcoinRepo {
	return f.ecoinRepo
}
