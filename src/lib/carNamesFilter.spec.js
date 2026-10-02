import { describe, it, expect } from 'vitest'
import { filterCarNames } from './carNamesFilter'

describe('carNamesFilter', () => {
  const data = [
    { id: 1, car_name: 'TIGO 7', brand: 'CHERY', id_brand: 9 },
    { id: 2, car_name: 'TACOUA', brand: 'VW', id_brand: 12 },
    { id: 3, car_name: 'UNASSIGNED', brand: null, id_brand: null },
    { id: 4, car_name: 'DASHING', brand: 'JETOUR', id_brand: 14 }
  ]

  it('returns all when query empty and no brand pin', () => {
    expect(filterCarNames(data, '')).toHaveLength(4)
    expect(filterCarNames(data, '  ')).toHaveLength(4)
  })

  it('matches car_name case-insensitively', () => {
    expect(filterCarNames(data, 'tigo')).toHaveLength(1)
    expect(filterCarNames(data, 'TIGO 7')).toHaveLength(1)
  })

  it('matches brand case-insensitively', () => {
    expect(filterCarNames(data, 'chery')).toHaveLength(1)
    expect(filterCarNames(data, 'vw')).toHaveLength(1)
  })

  it('strict brand pin overrides query', () => {
    const res = filterCarNames(data, 'vw', { brandId: 9 })
    expect(res).toHaveLength(1)
    expect(res[0].car_name).toBe('TIGO 7')
  })

  it('strict brand pin returns only matching id_brand', () => {
    const res = filterCarNames(data, '', { brandId: 12 })
    expect(res).toHaveLength(1)
    expect(res[0].car_name).toBe('TACOUA')
  })

  it('pin by numeric string works', () => {
    const res = filterCarNames(data, '', { brandId: '9' })
    expect(res).toHaveLength(1)
  })

  it('unassigned never matches brand pin', () => {
    const res = filterCarNames(data, '', { brandId: 9 })
    expect(res.some(r => r.id_brand === null)).toBe(false)
  })

  it('handles missing brand gracefully', () => {
    const res = filterCarNames([{ car_name: 'X' }], 'x')
    expect(res).toHaveLength(1)
  })

  it('returns empty array safely', () => {
    expect(filterCarNames(null, 'x')).toHaveLength(0)
    expect(filterCarNames(undefined, '')).toHaveLength(0)
  })
})
