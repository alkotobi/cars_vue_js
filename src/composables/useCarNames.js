// Composable: car names CRUD.
import { ref } from 'vue'

export function useCarNames({ notify } = {}) {
  const carNames = ref([])
  const loading = ref(false)
  const error = ref('')

  return function withDeps(deps = {}) {
    const { safeApi } = deps

    const fetchCarNames = async () => {
      loading.value = true
      error.value = ''
      const result = await safeApi({
        query: `
          SELECT cn.id, cn.car_name, cn.notes, cn.is_big_car, cn.id_brand, b.brand
          FROM cars_names cn
          LEFT JOIN brands b ON cn.id_brand = b.id
          ORDER BY cn.car_name ASC
        `,
        params: []
      })
      loading.value = false
      if (result.success) {
        carNames.value = result.data
      } else {
        carNames.value = []
        error.value = result.error || 'Car names could not be loaded.'
      }
    }

    const addCarName = async (payload = {}) => {
      const { car_name, notes = '', is_big_car = false, id_brand } = payload
      if (!car_name || !car_name.trim().length > 0 || !id_brand) {
        await notify?.('error', 'Validation error', 'Car name and brand are required.')
        return { success: false }
      }
      const result = await safeApi({
        query: 'INSERT INTO cars_names (car_name, notes, is_big_car, id_brand) VALUES (?, ?, ?, ?)',
        params: [car_name.trim().toUpperCase(), notes || '', is_big_car ? 1 : 0, id_brand]
      })
      if (!result.success) {
        if (result.error && result.error.toLowerCase().includes('duplicate')) {
          await notify?.('error', 'Duplicate car name', 'A car name with this name already exists.')
        } else {
          await notify?.('error', 'Failed to add car name', result.error || 'Unknown error')
        }
        return result
      }
      await fetchCarNames()
      return { success: true }
    }

    const updateCarName = async (payload = {}) => {
      const { id, car_name, notes = '', is_big_car = false, id_brand } = payload
      if (!id || !car_name || !car_name.trim().length > 0 || !id_brand) {
        await notify?.('error', 'Validation error', 'Car name and brand are required.')
        return { success: false }
      }
      const result = await safeApi({
        query: 'UPDATE cars_names SET car_name = ?, notes = ?, is_big_car = ?, id_brand = ? WHERE id = ?',
        params: [car_name.trim().toUpperCase(), notes || '', is_big_car ? 1 : 0, id_brand, id]
      })
      if (!result.success) {
        if (result.error && result.error.toLowerCase().includes('duplicate')) {
          await notify?.('error', 'Duplicate car name', 'A car name with this name already exists.')
        } else {
          await notify?.('error', 'Failed to update car name', result.error || 'Unknown error')
        }
        return result
      }
      await fetchCarNames()
      return { success: true }
    }

    const deleteCarName = async (carName) => {
      const refs = await safeApi({
        query: `
          SELECT
            (SELECT COUNT(*) FROM buy_details WHERE id_car_name = ?) AS buy_refs,
            (SELECT COUNT(*) FROM car_name_media WHERE car_name_id = ?) AS media_total,
            (SELECT COUNT(*) FROM car_name_media WHERE car_name_id = ? AND is_active = 1) AS media_active
        `,
        params: [carName.id, carName.id, carName.id]
      })
      if (!refs.success) {
        await notify?.('error', 'Failed to check car name usage', refs.error || 'Unknown error')
        return { success: false }
      }
      const r = refs.data?.[0] || { buy_refs: 0, media_total: 0, media_active: 0 }
      const buyRefs = Number(r.buy_refs) || 0
      const mediaTotal = Number(r.media_total) || 0
      const mediaActive = Number(r.media_active) || 0
      const hiddenMedia = mediaTotal - mediaActive
      const parts = []
      if (buyRefs) parts.push(`${buyRefs} purchase record${buyRefs === 1 ? '' : 's'}`)
      if (mediaActive) parts.push(`${mediaActive} media item${mediaActive === 1 ? '' : 's'}`)
      if (hiddenMedia) parts.push(`${hiddenMedia} hidden media item${hiddenMedia === 1 ? '' : 's'}`)
      if (buyRefs || mediaTotal) {
        await notify?.(
          'error',
          'Cannot delete car name',
          `${carName.car_name} is still referenced by other records.`,
          parts.join(', ')
        )
        return { success: false, blocked: true }
      }
      const confirmed = await notify?.(
        'warning',
        'Delete car name',
        `Are you sure you want to delete ${carName.car_name}?`,
        'This action cannot be undone.',
        'Delete'
      )
      if (!confirmed) return { success: false, cancelled: true }
      const del = await safeApi({ query: 'DELETE FROM cars_names WHERE id = ?', params: [carName.id] })
      if (!del.success) {
        await notify?.('error', 'Failed to delete car name', del.error || 'Unknown error')
        return del
      }
      await fetchCarNames()
      return { success: true }
    }

    return { carNames, loading, error, fetchCarNames, addCarName, updateCarName, deleteCarName }
  }
}
