<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useApi } from '../composables/useApi'
import { useSubmitGuard } from '../composables/useSubmitGuard'
import MessageBox from '../components/MessageBox.vue'

const { callApi, uploadFile, getFileUrl, error } = useApi()
const { guard, isBusy } = useSubmitGuard()

const showAddBrandDialog = ref(false)
const showEditBrandDialog = ref(false)
const showAddCarNameDialog = ref(false)
const showEditCarNameDialog = ref(false)
const brands = ref([])
const carNames = ref([])
const editingBrand = ref(null)
const editingCarName = ref(null)
const user = ref(null)
const showMediaDialog = ref(false)
const selectedCarNameForMedia = ref(null)
const carNameMedia = ref([])
const inactiveMediaCount = ref(0)
const uploadingMedia = ref(false)
const selectedMedia = ref([])
const rejectedMedia = ref([])
const mediaFileInput = ref(null)

const loadingBrands = ref(false)
const loadingCarNames = ref(false)
const brandsError = ref('')
const carNamesError = ref('')

const isAdmin = computed(() => user.value?.role_id === 1)

// ---------------------------------------------------------------------------
// Tabs and brand filter
//
// Brands and car names used to stack on one page: 13 brand rows above 35 car name
// rows, each with its own header and action buttons, which is a lot of scrolling to
// reach the second table. They are now separate tabs.
//
// The car name list is kept because the relationship between the two is worth
// exploiting: clicking a brand opens the car names filtered to it, which is how you
// answer "which models does this marque have" without reading the Brand column.
// ---------------------------------------------------------------------------
const activeTab = ref('brands')
const brandFilterId = ref(null)

// Resolved from brands rather than captured at click time, so the chip label
// follows a rename instead of showing the old name.
const brandFilterName = computed(() => {
  if (brandFilterId.value === null) return ''
  return brands.value.find((b) => b.id === brandFilterId.value)?.brand || ''
})

// Filters on id_brand, the id, not the brand name. Renaming a brand refetches both
// lists, so a name-based filter would silently stop matching after a rename.
// A null id_brand is an unassigned name and must never match a brand filter,
// otherwise those names become unreachable while a filter is set.
const filteredCarNames = computed(() => {
  if (brandFilterId.value === null) return carNames.value
  return carNames.value.filter((cn) => cn.id_brand === brandFilterId.value)
})

const viewModelsOfBrand = (brand) => {
  brandFilterId.value = brand.id
  activeTab.value = 'carNames'
}

const clearBrandFilter = () => {
  brandFilterId.value = null
}

const newBrand = ref({
  brand: '',
  logoFile: null
})

const editingBrandLogoFile = ref(null)

const newCarName = ref({
  car_name: '',
  notes: '',
  is_big_car: false,
  id_brand: null
})

// ---------------------------------------------------------------------------
// Feedback
//
// Replaces alert()/confirm(). MessageBox is a promise-based wrapper so a handler
// can await an answer, which native confirm() cannot do without blocking the
// event loop.
// ---------------------------------------------------------------------------
const msgBox = ref({
  show: false,
  type: 'info',
  title: '',
  message: '',
  details: '',
  showCancel: false,
  confirmText: 'OK'
})
let msgResolve = null

const settleMessageBox = (value) => {
  msgBox.value.show = false
  if (msgResolve) {
    const resolve = msgResolve
    msgResolve = null
    resolve(value)
  }
}

const onMsgConfirm = () => settleMessageBox(true)
const onMsgCancel = () => settleMessageBox(false)
const onMsgClose = () => settleMessageBox(false)

const notify = (type, title, message, details = '', showCancel = false) => {
  // Never strand the previous promise if two messages overlap.
  if (msgResolve) settleMessageBox(false)
  msgResolve = null
  return new Promise((resolve) => {
    msgResolve = resolve
    msgBox.value = {
      show: true,
      type,
      title,
      message,
      details,
      showCancel,
      confirmText: showCancel ? 'Yes' : 'OK'
    }
  })
}

const notifyError = (title, message, details = '') => notify('error', title, message, details)

// callApi re-throws on network/HTTP/parse failure instead of returning
// { success: false }. Without this wrapper every mutation below could reject
// unhandled and the user would see nothing at all.
const safeApi = async (payload) => {
  try {
    return await callApi(payload)
  } catch (err) {
    return { success: false, error: err?.message || 'Unexpected error' }
  }
}

// ---------------------------------------------------------------------------
// Validation
// ---------------------------------------------------------------------------
const isNewBrandValid = computed(() => newBrand.value.brand.trim().length > 0)
const isEditBrandValid = computed(
  () => !!editingBrand.value && editingBrand.value.brand.trim().length > 0
)
const isNewCarNameValid = computed(
  () => newCarName.value.car_name.trim().length > 0 && !!newCarName.value.id_brand
)
const isEditCarNameValid = computed(
  () =>
    !!editingCarName.value &&
    editingCarName.value.car_name.trim().length > 0 &&
    !!editingCarName.value.id_brand
)

// ---------------------------------------------------------------------------
// Media
//
// api/upload.php caps a single file at 100 MB and does not validate the
// extension or MIME type, only basename(). The allowlist below is deliberately
// narrower than the server's: it matches the media_type enum, which is only
// ('photo','video'), so an audio or archive file can no longer be mislabelled
// as a video. svg is excluded because it is served same-origin.
// ---------------------------------------------------------------------------
const MAX_MEDIA_BYTES = 100 * 1024 * 1024
const IMAGE_EXTENSIONS = ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'webp']
const VIDEO_EXTENSIONS = ['mp4', 'webm', 'avi', 'mov']
const MAX_LISTED_NAMES = 6

const extensionOf = (name) => String(name.split('.').pop() || '').toLowerCase()

const mediaTypeFor = (file) => {
  const ext = extensionOf(file.name)
  if (IMAGE_EXTENSIONS.includes(ext)) return 'photo'
  if (VIDEO_EXTENSIONS.includes(ext)) return 'video'
  return null
}

// crypto.randomUUID() only exists in a secure context, which plain-http LAN
// origins are not, so keep a fallback rather than throwing mid-upload.
const uniqueSuffix = () =>
  typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function'
    ? crypto.randomUUID()
    : `${Date.now().toString(36)}${Math.random().toString(36).slice(2, 10)}`

const formatFileSize = (bytes) => {
  if (bytes === null || bytes === undefined) return ''
  if (bytes === 0) return '0 B'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

const formatDate = (dateString) => {
  if (!dateString) return ''
  return new Date(dateString).toLocaleDateString()
}

const summariseNames = (names) => {
  if (names.length <= MAX_LISTED_NAMES) return names.join(', ')
  return `${names.slice(0, MAX_LISTED_NAMES).join(', ')} and ${names.length - MAX_LISTED_NAMES} more`
}

// ---------------------------------------------------------------------------
// Fetching
// ---------------------------------------------------------------------------
const fetchBrands = async () => {
  loadingBrands.value = true
  brandsError.value = ''
  const result = await safeApi({
    query: 'SELECT id, brand, logo_path FROM brands ORDER BY brand ASC',
    params: []
  })
  loadingBrands.value = false
  if (result.success) {
    brands.value = result.data
  } else {
    brands.value = []
    brandsError.value = result.error || error.value || 'Brands could not be loaded.'
  }
}

const fetchCarNames = async () => {
  loadingCarNames.value = true
  carNamesError.value = ''
  const result = await safeApi({
    query: `
      SELECT cn.id, cn.car_name, cn.notes, cn.is_big_car, cn.id_brand, b.brand
      FROM cars_names cn
      LEFT JOIN brands b ON cn.id_brand = b.id
      ORDER BY cn.car_name ASC
    `,
    params: []
  })
  loadingCarNames.value = false
  if (result.success) {
    carNames.value = result.data
  } else {
    carNames.value = []
    carNamesError.value = result.error || error.value || 'Car names could not be loaded.'
  }
}

const openMediaDialog = async (carName) => {
  selectedCarNameForMedia.value = carName
  showMediaDialog.value = true
  inactiveMediaCount.value = 0
  selectedMedia.value = []
  rejectedMedia.value = []
  await fetchCarNameMedia(carName.id)
}

const fetchCarNameMedia = async (carNameId) => {
  const [list, hidden] = await Promise.all([
    safeApi({
      query: `
        SELECT m.*, u.username as uploaded_by_username
        FROM car_name_media m
        LEFT JOIN users u ON m.uploaded_by = u.id
        WHERE m.car_name_id = ? AND m.is_active = 1
        ORDER BY m.uploaded_at DESC
      `,
      params: [carNameId]
    }),
    safeApi({
      query: 'SELECT COUNT(*) AS n FROM car_name_media WHERE car_name_id = ? AND is_active = 0',
      params: [carNameId]
    })
  ])

  if (list.success) {
    carNameMedia.value = list.data
  } else {
    carNameMedia.value = []
    await notifyError('Failed to load media', list.error || 'The media list could not be loaded.')
  }
  inactiveMediaCount.value = hidden.success ? Number(hidden.data?.[0]?.n || 0) : 0
}

const closeMediaDialog = () => {
  if (uploadingMedia.value) {
    notify('info', 'Upload in progress', 'Wait for the current upload to finish before closing.')
    return
  }
  showMediaDialog.value = false
  selectedCarNameForMedia.value = null
  carNameMedia.value = []
  inactiveMediaCount.value = 0
  selectedMedia.value = []
  rejectedMedia.value = []
}

// ---------------------------------------------------------------------------
// Media upload
// ---------------------------------------------------------------------------
const handleMediaFileSelect = (event) => {
  const files = Array.from(event.target.files || [])
  const accepted = []
  const rejected = []

  for (const file of files) {
    const mediaType = mediaTypeFor(file)
    if (!mediaType) {
      rejected.push(`${file.name} — unsupported type, use an image or video`)
      continue
    }
    if (file.size > MAX_MEDIA_BYTES) {
      rejected.push(`${file.name} — ${formatFileSize(file.size)} exceeds the 100 MB limit`)
      continue
    }
    accepted.push({ file, mediaType })
  }

  selectedMedia.value = accepted
  rejectedMedia.value = rejected

  if (rejected.length) {
    notify(
      'warning',
      'Some files were not accepted',
      `${rejected.length} of ${files.length} selected file(s) cannot be uploaded.`,
      rejected.join('; ')
    )
  }
}

const uploadCarNameMedia = async () => {
  if (uploadingMedia.value) return // prevent double submission
  if (!selectedMedia.value.length || !selectedCarNameForMedia.value) return
  if (!user.value || !user.value.id) {
    await notifyError('User not logged in', 'Your session has expired. Please log in again.')
    return
  }

  uploadingMedia.value = true
  const uploaded = []
  const failures = []
  const orphans = []

  try {
    for (const { file, mediaType } of selectedMedia.value) {
      const slug = selectedCarNameForMedia.value.car_name
        .toLowerCase()
        .replace(/\s+/g, '_')
        .replace(/[^a-z0-9_]/g, '')
      const filename = `${slug}_${mediaType}_${Date.now()}_${uniqueSuffix()}.${extensionOf(file.name)}`

      try {
        const uploadResult = await uploadFile(
          file,
          `car_names/${selectedCarNameForMedia.value.id}`,
          filename
        )

        if (!uploadResult.success) {
          failures.push(`${file.name}: ${uploadResult.error || 'upload failed'}`)
          continue
        }

        const insertResult = await safeApi({
          query: `
            INSERT INTO car_name_media (car_name_id, file_path, file_name, file_size, file_type, media_type, uploaded_by)
            VALUES (?, ?, ?, ?, ?, ?, ?)
          `,
          params: [
            selectedCarNameForMedia.value.id,
            uploadResult.relativePath,
            file.name,
            file.size,
            file.type || '',
            mediaType,
            user.value.id
          ]
        })

        if (insertResult.success) {
          uploaded.push(file.name)
        } else {
          failures.push(`${file.name}: ${insertResult.error || 'could not be recorded'}`)
          // The file is already on disk and no row points at it.
          orphans.push(uploadResult.relativePath)
        }
      } catch (err) {
        failures.push(`${file.name}: ${err.message}`)
      }
    }

    await fetchCarNameMedia(selectedCarNameForMedia.value.id)

    selectedMedia.value = []
    rejectedMedia.value = []
    if (mediaFileInput.value) mediaFileInput.value.value = ''

    const parts = []
    if (uploaded.length) parts.push(`Uploaded ${uploaded.length} file(s).`)
    if (failures.length) parts.push(`${failures.length} failed: ${failures.join('; ')}`)
    if (orphans.length) {
      parts.push(
        `Already on disk but not recorded, delete manually: ${orphans.join('; ')}`
      )
    }

    if (parts.length) {
      const type = failures.length ? 'warning' : 'info'
      await notify(type, failures.length ? 'Upload finished with errors' : 'Upload complete', parts.join(' '))
    }
  } finally {
    uploadingMedia.value = false
  }
}

const deleteCarNameMedia = guard('delete-media', async (media) => {
  const confirmed = await notify(
    'confirm',
    'Delete media',
    `Delete ${media.file_name}?`,
    'The record is hidden from this gallery, but the file stays on disk.',
    true
  )
  if (!confirmed) return

  const result = await safeApi({
    query: 'UPDATE car_name_media SET is_active = 0 WHERE id = ?',
    params: [media.id]
  })

  if (result.success) {
    await fetchCarNameMedia(selectedCarNameForMedia.value.id)
  } else {
    await notifyError('Failed to delete media', result.error || 'The media item could not be hidden.')
  }
})

// ---------------------------------------------------------------------------
// Brands
// ---------------------------------------------------------------------------
const closeEditBrandDialog = () => {
  showEditBrandDialog.value = false
  editingBrand.value = null
  editingBrandLogoFile.value = null
}

const addBrand = guard('add-brand', async () => {
  const brandName = newBrand.value.brand.trim()
  if (!brandName) {
    await notifyError('Brand name required', 'Enter a brand name before saving.')
    return
  }

  let logoPath = null

  if (newBrand.value.logoFile) {
    const filename = `brand_${brandName.toLowerCase().replace(/\s+/g, '_')}_${Date.now()}.${extensionOf(newBrand.value.logoFile.name)}`
    let uploadResult
    try {
      uploadResult = await uploadFile(newBrand.value.logoFile, 'brands', filename)
    } catch (err) {
      await notifyError('Failed to upload logo', err.message)
      return
    }
    // Previously a failed upload was ignored and the brand was still created
    // without its logo, which is indistinguishable from a working save.
    if (!uploadResult.success) {
      await notifyError(
        'Failed to upload logo',
        uploadResult.error || 'The logo could not be uploaded, so the brand was not created.'
      )
      return
    }
    logoPath = uploadResult.relativePath
  }

  // logo_path fallback retained: brands.logo_path exists in setup.sql and after
  // migration 012, but UpgradesCrud.vue guards the same way for a column that
  // un-migrated databases may lack.
  let query = 'INSERT INTO brands (brand, logo_path) VALUES (?, ?)'
  let params = [brandName, logoPath]

  const result = await safeApi({ query, params })

  if (!result.success && result.error && result.error.includes('logo_path')) {
    const retryResult = await safeApi({
      query: 'INSERT INTO brands (brand) VALUES (?)',
      params: [brandName]
    })
    if (retryResult.success) {
      if (logoPath) {
        await notify(
          'warning',
          'Brand added without its logo',
          'The database has no logo_path column, so the uploaded logo was not saved.',
          'Run api/migrations/012_add_brand_logos.sql to enable logo support.'
        )
      }
      showAddBrandDialog.value = false
      newBrand.value = { brand: '', logoFile: null }
      await fetchBrands()
      return
    }
    await notifyError('Failed to add brand', retryResult.error || 'Unknown error')
    return
  }

  if (result.success) {
    showAddBrandDialog.value = false
    newBrand.value = { brand: '', logoFile: null }
    await fetchBrands()
  } else {
    await notifyError('Failed to add brand', result.error || 'Unknown error')
  }
})

const editBrand = (brand) => {
  editingBrand.value = { ...brand }
  editingBrandLogoFile.value = null
  showEditBrandDialog.value = true
}

const updateBrand = guard('update-brand', async () => {
  const brandName = editingBrand.value.brand.trim()
  if (!brandName) {
    await notifyError('Brand name required', 'Enter a brand name before saving.')
    return
  }

  let logoPath = editingBrand.value.logo_path

  if (editingBrandLogoFile.value) {
    const filename = `brand_${brandName.toLowerCase().replace(/\s+/g, '_')}_${Date.now()}.${extensionOf(editingBrandLogoFile.value.name)}`
    let uploadResult
    try {
      uploadResult = await uploadFile(editingBrandLogoFile.value, 'brands', filename)
    } catch (err) {
      await notifyError('Failed to upload logo', err.message)
      return
    }
    if (!uploadResult.success) {
      await notifyError(
        'Failed to upload logo',
        uploadResult.error || 'The new logo could not be uploaded, so nothing was changed.'
      )
      return
    }
    logoPath = uploadResult.relativePath
  }

  let query = 'UPDATE brands SET brand = ?, logo_path = ? WHERE id = ?'
  let params = [brandName, logoPath, editingBrand.value.id]

  const result = await safeApi({ query, params })

  if (!result.success && result.error && result.error.includes('logo_path')) {
    const retryResult = await safeApi({
      query: 'UPDATE brands SET brand = ? WHERE id = ?',
      params: [brandName, editingBrand.value.id]
    })
    if (retryResult.success) {
      if (logoPath && editingBrandLogoFile.value) {
        await notify(
          'warning',
          'Brand renamed without its logo',
          'The database has no logo_path column, so the uploaded logo was not saved.',
          'Run api/migrations/012_add_brand_logos.sql to enable logo support.'
        )
      }
      closeEditBrandDialog()
      await fetchBrands()
      return
    }
    await notifyError('Failed to update brand', retryResult.error || 'Unknown error')
    return
  }

  if (result.success) {
    closeEditBrandDialog()
    await fetchBrands()
    // cars_names stores id_brand, but the listing resolves the brand name, so
    // a rename has to be reflected there too.
    await fetchCarNames()
  } else {
    await notifyError('Failed to update brand', result.error || 'Unknown error')
  }
})

const deleteBrand = guard('delete-brand', async (brand) => {
  // cars_names.id_brand has no foreign key, so deleting a brand does not
  // cascade and does not null the column: it leaves a dangling id and the
  // Brand column silently renders blank. Refuse instead.
  const usage = await safeApi({
    query: 'SELECT car_name FROM cars_names WHERE id_brand = ? ORDER BY car_name',
    params: [brand.id]
  })

  if (!usage.success) {
    await notifyError(
      'Cannot delete brand',
      usage.error || 'It could not be checked whether this brand is still in use.'
    )
    return
  }

  const inUse = usage.data || []
  if (inUse.length) {
    await notifyError(
      'Brand is still in use',
      `"${brand.brand}" is assigned to ${inUse.length} car name(s) and cannot be deleted.`,
      `Reassign these first: ${summariseNames(inUse.map((c) => c.car_name))}`
    )
    return
  }

  const confirmed = await notify(
    'confirm',
    'Delete brand',
    `Delete "${brand.brand}"?`,
    'No car names use this brand. This cannot be undone.',
    true
  )
  if (!confirmed) return

  const result = await safeApi({
    query: 'DELETE FROM brands WHERE id = ?',
    params: [brand.id]
  })

  if (result.success) {
    await fetchBrands()
  } else {
    await notifyError('Failed to delete brand', result.error || 'Unknown error')
  }
})

// ---------------------------------------------------------------------------
// Car names
// ---------------------------------------------------------------------------
const addCarName = guard('add-car-name', async () => {
  const carName = newCarName.value.car_name.trim()
  if (!carName || !newCarName.value.id_brand) {
    await notifyError('Incomplete car name', 'A car name and a brand are both required.')
    return
  }

  const result = await safeApi({
    query: 'INSERT INTO cars_names (car_name, notes, is_big_car, id_brand) VALUES (?, ?, ?, ?)',
    params: [
      carName,
      newCarName.value.notes,
      newCarName.value.is_big_car ? 1 : 0,
      newCarName.value.id_brand
    ]
  })

  if (result.success) {
    showAddCarNameDialog.value = false
    newCarName.value = { car_name: '', notes: '', is_big_car: false, id_brand: null }
    await fetchCarNames()
  } else {
    await notifyError(
      'Failed to add car name',
      result.error?.includes('Duplicate') || result.error?.includes('duplicate')
        ? `"${carName}" already exists.`
        : result.error || 'Unknown error'
    )
  }
})

const editCarName = (carName) => {
  editingCarName.value = { ...carName }
  showEditCarNameDialog.value = true
}

const updateCarName = guard('update-car-name', async () => {
  const carName = editingCarName.value.car_name.trim()
  if (!carName || !editingCarName.value.id_brand) {
    await notifyError('Incomplete car name', 'A car name and a brand are both required.')
    return
  }

  const result = await safeApi({
    query: 'UPDATE cars_names SET car_name = ?, notes = ?, is_big_car = ?, id_brand = ? WHERE id = ?',
    params: [
      carName,
      editingCarName.value.notes,
      editingCarName.value.is_big_car ? 1 : 0,
      editingCarName.value.id_brand,
      editingCarName.value.id
    ]
  })

  if (result.success) {
    showEditCarNameDialog.value = false
    editingCarName.value = null
    await fetchCarNames()
  } else {
    await notifyError(
      'Failed to update car name',
      result.error?.includes('Duplicate') || result.error?.includes('duplicate')
        ? `"${carName}" already exists.`
        : result.error || 'Unknown error'
    )
  }
})

const deleteCarName = guard('delete-car-name', async (carName) => {
  // car_name_media.car_name_id cascades, which would delete the media rows but
  // leave the files on disk. buy_details.id_car_name has no foreign key at all
  // and would dangle. Refuse while anything points at this row.
  const usage = await safeApi({
    query: `
      SELECT
        (SELECT COUNT(*) FROM buy_details WHERE id_car_name = ?) AS buy_refs,
        (SELECT COUNT(*) FROM car_name_media WHERE car_name_id = ?) AS media_total,
        (SELECT COUNT(*) FROM car_name_media WHERE car_name_id = ? AND is_active = 1) AS media_active
    `,
    params: [carName.id, carName.id, carName.id]
  })

  if (!usage.success) {
    await notifyError(
      'Cannot delete car name',
      usage.error || 'It could not be checked whether this car name is still in use.'
    )
    return
  }

  const row = usage.data?.[0] || {}
  const buyRefs = Number(row.buy_refs || 0)
  const mediaTotal = Number(row.media_total || 0)
  const mediaActive = Number(row.media_active || 0)
  const hidden = mediaTotal - mediaActive

  if (buyRefs || mediaTotal) {
    const parts = []
    if (buyRefs) parts.push(`${buyRefs} purchase record(s)`)
    if (mediaActive) parts.push(`${mediaActive} media item(s)`)
    if (hidden) parts.push(`${hidden} hidden media item(s)`)

    await notifyError(
      'Car name is still in use',
      `"${carName.car_name}" is referenced by ${parts.join(', ')} and cannot be deleted.`,
      buyRefs
        ? 'Purchase history keeps pointing at this name.'
        : 'Remove the media from the gallery first.'
    )
    return
  }

  const confirmed = await notify(
    'confirm',
    'Delete car name',
    `Delete "${carName.car_name}"?`,
    'Nothing references this name. This cannot be undone.',
    true
  )
  if (!confirmed) return

  const result = await safeApi({
    query: 'DELETE FROM cars_names WHERE id = ?',
    params: [carName.id]
  })

  if (result.success) {
    await fetchCarNames()
  } else {
    await notifyError('Failed to delete car name', result.error || 'Unknown error')
  }
})

// ---------------------------------------------------------------------------
// Lifecycle
// ---------------------------------------------------------------------------
const onKeydown = (event) => {
  if (event.key !== 'Escape') return
  if (msgBox.value.show) {
    onMsgCancel()
    return
  }
  if (showMediaDialog.value) {
    closeMediaDialog()
    return
  }
  if (showEditCarNameDialog.value) {
    showEditCarNameDialog.value = false
    editingCarName.value = null
    return
  }
  if (showAddCarNameDialog.value) {
    showAddCarNameDialog.value = false
    return
  }
  if (showEditBrandDialog.value) {
    closeEditBrandDialog()
    return
  }
  if (showAddBrandDialog.value) {
    showAddBrandDialog.value = false
  }
}

onMounted(async () => {
  window.addEventListener('keydown', onKeydown)

  const userStr = localStorage.getItem('user')
  if (!userStr) return

  try {
    user.value = JSON.parse(userStr)
  } catch {
    await notifyError(
      'Session problem',
      'The saved login could not be read, so brands and car names were not loaded.',
      'Log out and log in again.'
    )
    return
  }

  await Promise.all([fetchBrands(), fetchCarNames()])
})

onUnmounted(() => {
  window.removeEventListener('keydown', onKeydown)
  settleMessageBox(false)
})
</script>

<template>
  <div class="models-view">
    <div class="tab-bar" role="tablist" aria-label="Brands and car names">
      <button
        class="tab-btn"
        :class="{ active: activeTab === 'brands' }"
        role="tab"
        :aria-selected="activeTab === 'brands'"
        @click="activeTab = 'brands'"
      >
        Brands <span class="tab-count">{{ brands.length }}</span>
      </button>
      <button
        class="tab-btn"
        :class="{ active: activeTab === 'carNames' }"
        role="tab"
        :aria-selected="activeTab === 'carNames'"
        @click="activeTab = 'carNames'"
      >
        Car Names <span class="tab-count">{{ filteredCarNames.length }}</span>
      </button>
    </div>

    <div class="section" v-if="activeTab === 'brands'">
      <div class="header">
        <h2>Brands Management</h2>
        <button @click="showAddBrandDialog = true" class="add-btn">Add Brand</button>
      </div>
      <table class="data-table">
        <thead>
          <tr>
            <th>Logo</th>
            <th>Brand</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loadingBrands">
            <td colspan="3" class="table-status">Loading brands…</td>
          </tr>
          <tr v-else-if="brandsError">
            <td colspan="3" class="table-status table-status-error">
              {{ brandsError }}
              <button class="btn retry-btn" @click="fetchBrands">Retry</button>
            </td>
          </tr>
          <tr v-else-if="brands.length === 0">
            <td colspan="3" class="table-status">No brands yet.</td>
          </tr>
          <tr
            v-for="brand in brands"
            v-else
            :key="brand.id"
            class="brand-row"
            tabindex="0"
            role="button"
            :aria-label="`View models of ${brand.brand}`"
            @click="viewModelsOfBrand(brand)"
            @keydown.enter.prevent="viewModelsOfBrand(brand)"
            @keydown.space.prevent="viewModelsOfBrand(brand)"
          >
            <td>
              <img 
                v-if="brand.logo_path" 
                :src="getFileUrl(brand.logo_path)" 
                :alt="brand.brand"
                class="brand-logo"
              />
              <span v-else class="no-logo">No logo</span>
            </td>
            <td>{{ brand.brand }}</td>
            <td>
              <button @click.stop="editBrand(brand)" class="btn edit-btn">Edit</button>
              <button 
                v-if="isAdmin"
                @click.stop="deleteBrand(brand)" 
                class="btn delete-btn" :disabled="isBusy('delete-brand')">Delete</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="section" v-if="activeTab === 'carNames'">
      <div class="header">
        <h2>Car Names Management</h2>
        <button @click="showAddCarNameDialog = true" class="add-btn">Add Car Name</button>
      </div>
      <div v-if="brandFilterId !== null" class="filter-chip">
        <span>Showing models of <strong>{{ brandFilterName }}</strong></span>
        <button type="button" class="filter-clear" @click="clearBrandFilter">Show all</button>
      </div>
      <table class="data-table">
        <thead>
          <tr>
            <th>Car Name</th>
            <th>Brand</th>
            <th>Notes</th>
            <th>Big Car</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loadingCarNames">
            <td colspan="5" class="table-status">Loading car names…</td>
          </tr>
          <tr v-else-if="carNamesError">
            <td colspan="5" class="table-status table-status-error">
              {{ carNamesError }}
              <button class="btn retry-btn" @click="fetchCarNames">Retry</button>
            </td>
          </tr>
          <tr v-else-if="carNames.length === 0">
            <td colspan="5" class="table-status">No car names yet.</td>
          </tr>
          <tr v-else-if="filteredCarNames.length === 0">
            <td colspan="5" class="table-status">
              {{ brandFilterName }} has no models.
              <button class="btn retry-btn" @click="clearBrandFilter">Show all</button>
            </td>
          </tr>
          <tr v-for="carName in filteredCarNames" v-else :key="carName.id">
            <td>{{ carName.car_name }}</td>
            <td>
              <span v-if="carName.brand">{{ carName.brand }}</span>
              <span v-else class="no-logo" title="This car name has no brand assigned">No brand</span>
            </td>
            <td>{{ carName.notes }}</td>
            <td>{{ carName.is_big_car ? 'Yes' : 'No' }}</td>
            <td>
              <button @click="editCarName(carName)" class="btn edit-btn">Edit</button>
              <button @click="openMediaDialog(carName)" class="btn media-btn">Photos/Videos</button>
              <button 
                v-if="isAdmin"
                @click="deleteCarName(carName)" 
                class="btn delete-btn" :disabled="isBusy('delete-car-name')">Delete</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Add Brand Dialog -->
    <div v-if="showAddBrandDialog" class="dialog-overlay" @click.self="showAddBrandDialog = false">
      <div class="dialog" role="dialog" aria-modal="true" aria-labelledby="add-brand-title">
        <h3 id="add-brand-title">Add New Brand</h3>
        <div class="form-group">
          <label for="add-brand-name">Brand Name</label>
          <input 
            id="add-brand-name"
            v-model="newBrand.brand" 
            placeholder="Brand Name" 
            class="input-field"
          />
        </div>
        <div class="form-group">
          <label for="add-brand-logo">Brand Logo</label>
          <input 
            id="add-brand-logo"
            type="file"
            accept="image/png,image/jpeg,image/gif,image/webp"
            @change="(e) => newBrand.logoFile = e.target.files?.[0] || null"
            class="input-field"
          />
          <small class="form-hint">Upload a logo image (PNG, JPG, etc.)</small>
        </div>
        <div class="dialog-actions">
          <button @click="addBrand" class="btn save-btn" :disabled="isBusy('add-brand') || !isNewBrandValid">Save</button>
          <button @click="showAddBrandDialog = false" class="btn cancel-btn">Cancel</button>
        </div>
      </div>
    </div>

    <!-- Edit Brand Dialog -->
    <div v-if="showEditBrandDialog" class="dialog-overlay" @click.self="closeEditBrandDialog">
      <div class="dialog" role="dialog" aria-modal="true" aria-labelledby="edit-brand-title">
        <h3 id="edit-brand-title">Edit Brand</h3>
        <div class="form-group">
          <label for="edit-brand-name">Brand Name</label>
          <input 
            id="edit-brand-name"
            v-model="editingBrand.brand" 
            placeholder="Brand Name" 
            class="input-field"
          />
        </div>
        <div class="form-group">
          <label>Current Logo</label>
          <div v-if="editingBrand.logo_path" class="logo-preview">
            <img 
              :src="getFileUrl(editingBrand.logo_path)" 
              :alt="editingBrand.brand"
              class="brand-logo-preview"
            />
          </div>
          <span v-else class="no-logo">No logo uploaded</span>
        </div>
        <div class="form-group">
          <label for="edit-brand-logo">Upload New Logo (optional)</label>
          <input 
            id="edit-brand-logo"
            type="file"
            accept="image/png,image/jpeg,image/gif,image/webp"
            @change="(e) => editingBrandLogoFile = e.target.files?.[0] || null"
            class="input-field"
          />
          <small class="form-hint">Upload a new logo to replace the current one</small>
        </div>
        <div class="dialog-actions">
          <button @click="updateBrand" class="btn save-btn" :disabled="isBusy('update-brand') || !isEditBrandValid">Save</button>
          <button @click="closeEditBrandDialog" class="btn cancel-btn">Cancel</button>
        </div>
      </div>
    </div>

    <!-- Add Car Name Dialog -->
    <div v-if="showAddCarNameDialog" class="dialog-overlay" @click.self="showAddCarNameDialog = false">
      <div class="dialog" role="dialog" aria-modal="true" aria-labelledby="add-car-name-title">
        <h3 id="add-car-name-title">Add New Car Name</h3>
        <div class="form-group">
          <label for="add-car-name-input">Car Name</label>
          <input 
            id="add-car-name-input"
            v-model="newCarName.car_name" 
            placeholder="Car Name"
          >
        </div>
        <div class="form-group">
          <label for="add-car-name-brand">Brand</label>
          <select id="add-car-name-brand" v-model="newCarName.id_brand">
            <option :value="null" disabled>Select Brand</option>
            <option v-for="brand in brands" :key="brand.id" :value="brand.id">
              {{ brand.brand }}
            </option>
          </select>
          <small class="form-hint">A brand is required — an unbranded car name is treated as incomplete data.</small>
        </div>
        <div class="form-group">
          <label for="add-car-name-notes">Notes</label>
          <input 
            id="add-car-name-notes"
            v-model="newCarName.notes" 
            placeholder="Notes"
          >
        </div>
        <div class="form-group">
          <label for="add-car-name-big">
            <input 
              id="add-car-name-big"
              type="checkbox" 
              v-model="newCarName.is_big_car"
            > Big Car
          </label>
        </div>
        <div class="dialog-buttons">
          <button @click="addCarName" class="btn save-btn" :disabled="isBusy('add-car-name') || !isNewCarNameValid">Save</button>
          <button @click="showAddCarNameDialog = false" class="btn cancel-btn">Cancel</button>
        </div>
      </div>
    </div>

    <!-- Edit Car Name Dialog -->
    <div v-if="showEditCarNameDialog" class="dialog-overlay">
      <div class="dialog" role="dialog" aria-modal="true" aria-labelledby="edit-car-name-title">
        <h3 id="edit-car-name-title">Edit Car Name</h3>
        <div class="form-group">
          <label for="edit-car-name-input">Car Name</label>
          <input 
            id="edit-car-name-input"
            v-model="editingCarName.car_name" 
            placeholder="Car Name"
          >
        </div>
        <div class="form-group">
          <label for="edit-car-name-brand">Brand</label>
          <select id="edit-car-name-brand" v-model="editingCarName.id_brand">
            <option :value="null" disabled>Select Brand</option>
            <option v-for="brand in brands" :key="brand.id" :value="brand.id">
              {{ brand.brand }}
            </option>
          </select>
          <small class="form-hint">A brand is required — an unbranded car name is treated as incomplete data.</small>
        </div>
        <div class="form-group">
          <label for="edit-car-name-notes">Notes</label>
          <input 
            id="edit-car-name-notes"
            v-model="editingCarName.notes" 
            placeholder="Notes"
          >
        </div>
        <div class="form-group">
          <label for="edit-car-name-big">
            <input 
              id="edit-car-name-big"
              type="checkbox" 
              v-model="editingCarName.is_big_car"
            > Big Car
          </label>
        </div>
        <div class="dialog-buttons">
          <button @click="updateCarName" class="btn save-btn" :disabled="isBusy('update-car-name') || !isEditCarNameValid">Save</button>
          <button @click="showEditCarNameDialog = false" class="btn cancel-btn">Cancel</button>
        </div>
      </div>
    </div>

    <!-- Media Dialog -->
    <div v-if="showMediaDialog" class="dialog-overlay" @click.self="closeMediaDialog">
      <div class="dialog media-dialog" role="dialog" aria-modal="true" aria-labelledby="media-dialog-title">
        <div class="dialog-header">
          <h3 id="media-dialog-title">Photos &amp; Videos — {{ selectedCarNameForMedia ? selectedCarNameForMedia.car_name : '' }}</h3>
          <button @click="closeMediaDialog" class="close-btn" :disabled="uploadingMedia" aria-label="Close media gallery">&times;</button>
        </div>
        
        <div class="media-upload-section">
          <div class="form-group">
            <label for="media-file-input">Upload Photos/Videos</label>
            <input
              id="media-file-input"
              ref="mediaFileInput"
              type="file"
              accept=".jpg,.jpeg,.png,.gif,.bmp,.webp,.mp4,.webm,.avi,.mov"
              multiple
              @change="handleMediaFileSelect"
              class="input-field"
              :disabled="uploadingMedia"
            />
            <small class="form-hint">
              You can select multiple files at once. Images and video only, up to 100 MB each.
            </small>
            <small v-if="rejectedMedia.length" class="form-hint form-hint-error">
              Not accepted: {{ rejectedMedia.join('; ') }}
            </small>
          </div>
          <button 
            @click="uploadCarNameMedia" 
            class="btn save-btn"
            :disabled="!selectedMedia.length || uploadingMedia"
          >
            <span v-if="uploadingMedia">Uploading...</span>
            <span v-else>Upload {{ selectedMedia.length }} file(s)</span>
          </button>
        </div>

        <div class="media-gallery">
          <h4>Media Gallery ({{ carNameMedia.length }})</h4>
          <p v-if="inactiveMediaCount" class="media-hidden-note">
            {{ inactiveMediaCount }} deleted item(s) are hidden from this gallery. Their rows are
            kept so the reference count stays honest, and their files remain on disk.
          </p>
          <div v-if="carNameMedia.length === 0" class="no-media">
            No photos or videos uploaded yet
          </div>
          <div v-else class="media-grid">
            <div v-for="media in carNameMedia" :key="media.id" class="media-item">
              <div v-if="media.media_type === 'photo'" class="media-photo">
                <img :src="getFileUrl(media.file_path)" :alt="media.file_name" />
              </div>
              <div v-else class="media-video">
                <video controls :src="getFileUrl(media.file_path)"></video>
              </div>
              <div class="media-info">
                <div class="media-name">{{ media.file_name }}</div>
                <div class="media-meta">
                  <span>{{ media.media_type }}</span>
                  <span v-if="media.file_size">{{ formatFileSize(media.file_size) }}</span>
                  <span>{{ formatDate(media.uploaded_at) }}</span>
                </div>
              </div>
              <button @click="deleteCarNameMedia(media)" class="btn delete-btn-small" title="Delete" :aria-label="'Delete ' + media.file_name" :disabled="isBusy('delete-media')">
                <i class="fas fa-trash"></i>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <MessageBox
      :show="msgBox.show"
      :type="msgBox.type"
      :title="msgBox.title"
      :message="msgBox.message"
      :details="msgBox.details"
      :show-cancel="msgBox.showCancel"
      :confirm-text="msgBox.confirmText"
      @confirm="onMsgConfirm"
      @cancel="onMsgCancel"
      @close="onMsgClose"
    />
  </div>
</template>

<style scoped>
.models-view {
  padding: 20px;
}

/* Tab bar. Ported from CarsView.vue, where the same rules exist but are unused:
   that copy is in a scoped block, so it cannot reach this component. */
.tab-bar {
  display: flex;
  gap: 4px;
  background: #e2e8f0;
  padding: 4px;
  border-radius: 8px;
  flex-wrap: wrap;
  margin-bottom: 24px;
}

.tab-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border: none;
  border-radius: 6px;
  background: transparent;
  cursor: pointer;
  font-size: 0.9em;
  color: #475569;
  transition: all 0.15s ease;
}

.tab-btn:hover {
  background: rgba(255, 255, 255, 0.8);
  color: #1e293b;
}

.tab-btn.active {
  background: white;
  color: #1e293b;
  font-weight: 600;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.06);
}

.tab-count {
  display: inline-block;
  padding: 0 6px;
  border-radius: 10px;
  background: rgba(0, 0, 0, 0.07);
  font-size: 0.85em;
  line-height: 1.5;
}

.tab-btn.active .tab-count {
  background: #e2e8f0;
}

/* Clickable brand row. The whole row drills into that brand's models, so it needs
   the pointer and a visible focus ring for keyboard users: tabindex and the
   keydown handlers are on the <tr>, and without an outline the focus position
   would be invisible. */
.brand-row {
  cursor: pointer;
}

.brand-row:focus-visible {
  outline: 2px solid #3b82f6;
  outline-offset: -2px;
}

.filter-chip {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 20px;
  padding: 8px 12px;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  border-radius: 6px;
  font-size: 0.9rem;
  color: #1e40af;
}

.filter-clear {
  border: none;
  background: transparent;
  color: #2563eb;
  cursor: pointer;
  font-size: 0.9rem;
  text-decoration: underline;
  padding: 0;
}

.filter-clear:hover {
  color: #1d4ed8;
}

.section {
  margin-bottom: 40px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 20px;
}

.data-table th,
.data-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #ddd;
}

.data-table th {
  background-color: #f8f9fa;
  font-weight: 600;
}

.data-table tbody tr:hover {
  background-color: #f5f5f5;
}

.table-status {
  text-align: center;
  padding: 24px;
  color: #9ca3af;
  font-style: italic;
}

.table-status-error {
  color: #b91c1c;
  font-style: normal;
}

.retry-btn {
  margin-left: 12px;
  background-color: #3b82f6;
  color: white;
}

.form-hint-error {
  color: #b91c1c;
  font-weight: 500;
}

.media-hidden-note {
  margin: 0 0 12px;
  padding: 8px 12px;
  background-color: #fffbeb;
  border-left: 3px solid #f59e0b;
  border-radius: 4px;
  color: #92400e;
  font-size: 0.8125rem;
}

.close-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.add-btn {
  padding: 8px 16px;
  background-color: #10b981;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.btn {
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  margin-right: 8px;
}

.edit-btn {
  background-color: #3b82f6;
  color: white;
}

.delete-btn {
  background-color: #ef4444;
  color: white;
}

.save-btn {
  background-color: #10b981;
  color: white;
}

.cancel-btn {
  background-color: #6b7280;
  color: white;
}

.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
}

.dialog {
  background-color: white;
  padding: 20px;
  border-radius: 8px;
  min-width: 400px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 20px;
}

.input-field {
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.textarea {
  min-height: 100px;
  resize: vertical;
}

/* Dialog action rows. The four dialogs in this view are siblings, so they share
   one rule: flex, right-aligned, with the same gap and top margin. Both class
   names appear in the template and .dialog-buttons is the repo-wide convention
   (BuyView.vue, BuyDetailsTable.vue, BuyBillPaymentsView.vue all define it);
   previously this view used it without defining it, so the Add and Edit Car Name
   button rows had no flex layout at all and the buttons stacked. */
.dialog-actions,
.dialog-buttons {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
}

.checkbox-field {
  display: flex;
  align-items: center;
  gap: 8px;
}

.checkbox-field input[type="checkbox"] {
  width: 16px;
  height: 16px;
  cursor: pointer;
}

.checkbox-field label {
  cursor: pointer;
}

.brand-logo {
  width: 40px;
  height: 40px;
  object-fit: contain;
  border-radius: 4px;
}

.brand-logo-preview {
  width: 100px;
  height: 100px;
  object-fit: contain;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 8px;
  background: #f9f9f9;
}

.logo-preview {
  margin-bottom: 12px;
}

.no-logo {
  color: #9ca3af;
  font-style: italic;
  font-size: 0.875rem;
}

.form-group label {
  font-weight: 500;
  margin-bottom: 4px;
  display: block;
  color: #374151;
}

.form-hint {
  display: block;
  color: #6b7280;
  font-size: 0.75rem;
  margin-top: 4px;
}

.media-btn {
  background-color: #8b5cf6;
  color: white;
}

.media-dialog {
  max-width: 900px;
  max-height: 90vh;
  overflow-y: auto;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid #e5e7eb;
}

.close-btn {
  background: none;
  border: none;
  font-size: 28px;
  cursor: pointer;
  color: #6b7280;
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
}

.close-btn:hover {
  background-color: #f3f4f6;
  color: #374151;
}

.media-upload-section {
  margin-bottom: 30px;
  padding: 20px;
  background-color: #f9fafb;
  border-radius: 8px;
}

.media-gallery h4 {
  margin-bottom: 16px;
  color: #374151;
}

.no-media {
  text-align: center;
  padding: 40px;
  color: #9ca3af;
  font-style: italic;
}

.media-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 16px;
}

.media-item {
  position: relative;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  overflow: hidden;
  background: white;
}

.media-photo img,
.media-video video {
  width: 100%;
  height: 200px;
  object-fit: cover;
  display: block;
}

.media-info {
  padding: 12px;
}

.media-name {
  font-weight: 500;
  font-size: 0.875rem;
  color: #374151;
  margin-bottom: 8px;
  word-break: break-word;
}

.media-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  font-size: 0.75rem;
  color: #6b7280;
}

.media-meta span {
  padding: 2px 8px;
  background-color: #f3f4f6;
  border-radius: 4px;
  text-transform: capitalize;
}

.delete-btn-small {
  position: absolute;
  top: 8px;
  right: 8px;
  padding: 6px 10px;
  background-color: rgba(239, 68, 68, 0.9);
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.875rem;
}

.delete-btn-small:hover {
  background-color: #ef4444;
}

</style>