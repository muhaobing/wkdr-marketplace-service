// Package err_code 对外 HTTP JSON 的 retcode 约定（与 message 配合供调用方区分场景）
package err_code

const (
	// Success 成功
	Success = 0
	// General 未单独分类的错误（默认）
	General = -1

	// UserBindingNotFound 业务身份 biz_code + biz_user_id 在商城未绑定商城用户
	UserBindingNotFound = 100404
)
