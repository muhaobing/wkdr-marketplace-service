package bootstrap

import "wdkr-marketplace-service/internal/resource"

var scheduler *resource.Resources

// SetResources 设置资源引用（在 main 中调用）
func SetResources(r *resource.Resources) {
	scheduler = r
}

func StartUp() error {
	if scheduler != nil && scheduler.Scheduler != nil {
		scheduler.Scheduler.Start()
	}
	return nil
}
