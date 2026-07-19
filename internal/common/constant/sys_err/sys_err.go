// Package sys_err 可配合 errors.Is 在 controller 层识别的业务错误
package sys_err

import "errors"

// ErrUserBindingNotFound 表示 biz_code + biz_user_id 无对应绑定，无法解析商城 user_id
var ErrUserBindingNotFound = errors.New("user binding not found")

// ErrInsufficientEcoin 个人/企业金币库存不足，无法完成扣除（可用 errors.Is 识别）
var ErrInsufficientEcoin = errors.New("金币不足")

// ErrEcoinBillNotFound 金币账单不存在或无权访问
var ErrEcoinBillNotFound = errors.New("points bill not found")

// ErrEcoinBillInvalidState 金币账单状态不允许该操作
var ErrEcoinBillInvalidState = errors.New("points bill invalid status")
