/**
 * 是否与后端 Sku.IsEcoinGrantFulfill 一致：积分发放履约（含未写 fulfill_mode 但已配每件积分且无回调）
 */
export function isEcoinGrantSku(p) {
  if (!p || typeof p !== 'object') return false
  if (Number(p.fulfill_mode) === 1) return true
  const amt = Number(p.fulfill_ecoin_amount)
  const dm = String(p.delivery_method ?? '').trim()
  return amt > 0 && !dm
}
