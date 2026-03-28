/**
 * 与 LawMind 同源部署时，LawMind 使用 localStorage `token` 存 JWT。
 * 商城使用独立键，避免新 tab 内 401 清理等操作误删 LawMind 登录态。
 */
export const STORAGE_TOKEN_KEY = 'marketplace_token'

/** 当前登录用户业务平台绑定列表 JSON，与 GET /marketplace/user/bindings 一致 */
export const STORAGE_BINDINGS_KEY = 'marketplace_bindings'
