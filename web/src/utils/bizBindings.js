/**
 * LawSharp 跳转参数与商城用户绑定关系校验
 */

/**
 * 从路由 query 解析 biz_code、biz_user_id；不完整或非法则返回 null
 */
export function parseBizQueryFromRoute(query) {
  if (!query) return null
  const code = query.biz_code != null ? String(query.biz_code).trim() : ''
  const uidRaw = query.biz_user_id != null ? String(query.biz_user_id).trim() : ''
  if (!code || !uidRaw) return null
  return { biz_code: code, biz_user_id: uidRaw }
}

/**
 * 绑定列表是否包含指定业务身份（与后端 user_binding_tab 一致）
 */
export function bindingListContains(bindings, bizCode, bizUserId) {
  if (!Array.isArray(bindings)) return false
  const uidStr = String(bizUserId).trim()
  return bindings.some(
    (b) => String(b.biz_code).trim() === String(bizCode).trim() && String(b.biz_user_id).trim() === uidStr
  )
}
