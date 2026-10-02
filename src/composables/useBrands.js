// Composable: brands CRUD.
// Keeps the exact SQL and guard semantics from CarModelsView.
import { ref } from 'vue'

const MAX_LISTED_NAMES = 10

function summariseNames(names) {
  if (!Array.isArray(names) || names.length === 0) return ''
  if (names.length <= MAX_LISTED_NAMES) {
    return names.join(', ')
  }
  const listed = names.slice(0, MAX_LISTED_NAMES).join(', ')
  const remaining = names.length - MAX_LISTED_NAMES
  return `${listed} + ${remaining} more`
}

export function useBrands({ notify, onBrandsChanged } = {}) {
  const brands = ref([])
  const loading = ref(false)
  const error = ref('')

  let safeApiRef = null
  let uploadFileRef = null
  let getFileUrlRef = null

  const ensureDeps = () => {
    if (!safeApiRef) {
      // lazy require to avoid circular if needed, but injected pattern not used here
      // instead, we expect deps passed via init? simpler: accept deps
    }
  }

  // allow late init pattern by returning init function? easier: factory takes deps
  // but better: second param { safeApi, uploadFile, getFileUrl }
  return function withDeps(deps = {}) {
    const { safeApi, uploadFile, getFileUrl } = deps
    safeApiRef = safeApi
    uploadFileRef = uploadFile
    getFileUrlRef = getFileUrl

    const fetchBrands = async () => {
      loading.value = true
      error.value = ''
      const result = await safeApi({
        query: 'SELECT id, brand, logo_path FROM brands ORDER BY brand ASC',
        params: []
      })
      loading.value = false
      if (result.success) {
        brands.value = result.data
      } else {
        brands.value = []
        error.value = result.error || 'Brands could not be loaded.'
      }
    }

    const addBrand = async (payload = {}) => {
      const { brand, logoFile } = payload
      if (!brand || !brand.trim()) {
        await notify?.('error', 'Validation error', 'Brand name is required.')
        return { success: false, error: 'Brand name is required.' }
      }
      let logoPath = null
      if (logoFile) {
        const uploadResult = await uploadFile(logoFile)
        if (!uploadResult.success) {
          await notify?.('error', 'Logo upload failed', uploadResult.error || 'Could not upload logo.')
          return { success: false, error: uploadResult.error || 'Logo upload failed' }
        }
        logoPath = uploadResult.path
      }
      const result = await safeApi({
        query: logoPath
          ? 'INSERT INTO brands (brand, logo_path) VALUES (?, ?)'
          : 'INSERT INTO brands (brand) VALUES (?)',
        params: logoPath ? [brand.trim(), logoPath] : [brand.trim()]
      })
      if (!result.success) {
        await notify?.('error', 'Failed to add brand', result.error || 'Unknown error')
        return result
      }
      await fetchBrands()
      if (typeof onBrandsChanged === 'function') await onBrandsChanged()
      return { success: true }
    }

    const updateBrand = async (payload = {}) => {
      const { brand, id, logoFile, editingBrandLogoFile } = payload
      if (!brand || !brand.trim()) {
        await notify?.('error', 'Validation error', 'Brand name is required.')
        return { success: false, error: 'Brand name is required.' }
      }
      let logoPath = null
      if (logoFile) {
        const uploadResult = await uploadFile(logoFile)
        if (!uploadResult.success) {
          await notify?.('error', 'Logo upload failed', uploadResult.error || 'Could not upload logo.')
          return { success: false, error: uploadResult.error || 'Logo upload failed' }
        }
        logoPath = uploadResult.path
      }
      let result = await safeApi({
        query: logoPath
          ? 'UPDATE brands SET brand = ?, logo_path = ? WHERE id = ?'
          : 'UPDATE brands SET brand = ? WHERE id = ?',
        params: logoPath ? [brand.trim(), logoPath, id] : [brand.trim(), id]
      })
      if (!result.success && result.error && result.error.toLowerCase().includes('logo_path')) {
        const retryResult = await safeApi({
          query: 'UPDATE brands SET brand = ? WHERE id = ?',
          params: [brand.trim(), id]
        })
        if (retryResult.success) {
          if (logoPath && editingBrandLogoFile) {
            await notify?.(
              'warning',
              'Brand renamed without its logo',
              'The database has no logo_path column, so the uploaded logo was not saved.',
              'Run api/migrations/012_add_brand_logos.sql to enable logo support.'
            )
          }
          await fetchBrands()
          if (typeof onBrandsChanged === 'function') await onBrandsChanged()
          return { success: true }
        }
        await notify?.('error', 'Failed to update brand', retryResult.error || 'Unknown error')
        return retryResult
      }
      if (result.success) {
        await fetchBrands()
        if (typeof onBrandsChanged === 'function') await onBrandsChanged()
      } else {
        await notify?.('error', 'Failed to update brand', result.error || 'Unknown error')
      }
      return result
    }

    const deleteBrand = async (brand) => {
      const usage = await safeApi({
        query: 'SELECT car_name FROM cars_names WHERE id_brand = ? ORDER BY car_name',
        params: [brand.id]
      })
      if (!usage.success) {
        await notify?.('error', 'Failed to check brand usage', usage.error || 'Unknown error')
        return { success: false }
      }
      const names = (usage.data || []).map((r) => r.car_name).filter(Boolean)
      if (names.length > 0) {
        const summary = summariseNames(names)
        await notify?.(
          'error',
          'Cannot delete brand',
          `${brand.brand} is still assigned to ${names.length} car name${names.length === 1 ? '' : 's'}.`,
          summary
        )
        return { success: false, blocked: true }
      }
      const confirmed = await notify?.(
        'warning',
        'Delete brand',
        `Are you sure you want to delete ${brand.brand}?`,
        'This action cannot be undone.',
        'Delete'
      )
      if (!confirmed) return { success: false, cancelled: true }
      const del = await safeApi({ query: 'DELETE FROM brands WHERE id = ?', params: [brand.id] })
      if (!del.success) {
        await notify?.('error', 'Failed to delete brand', del.error || 'Unknown error')
        return del
      }
      await fetchBrands()
      if (typeof onBrandsChanged === 'function') await onBrandsChanged()
      return { success: true }
    }

    return { brands, loading, error, fetchBrands, addBrand, updateBrand, deleteBrand }
  }
}
