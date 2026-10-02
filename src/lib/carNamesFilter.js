// Pure filtering logic for car names.
// - When brandId is provided (strict pin), returns only rows whose id_brand === brandId.
// - Otherwise, case-insensitively matches car_name OR brand against query.
// - query is trimmed; empty query returns all rows unchanged.
export function filterCarNames(carNames, query, opts = {}) {
  const list = Array.isArray(carNames) ? carNames : []
  const q = typeof query === 'string' ? query.trim() : ''
  const brandId = opts.brandId ?? null
  const brandName = opts.brandName ?? null

  if (brandId !== null && brandId !== undefined && brandId !== '') {
    const bid = Number(brandId)
    if (!Number.isNaN(bid)) {
      return list.filter((cn) => cn && cn.id_brand === bid)
    }
  }

  if (q === '') {
    return list
  }

  const needle = q.toLowerCase()
  return list.filter((cn) => {
    if (!cn) return false
    const name = (cn.car_name || '').toLowerCase()
    const brand = (cn.brand || '').toLowerCase()
    if (name.includes(needle)) return true
    if (brand.includes(needle)) return true
    return false
  })
}
