// Composable: car name media gallery + upload.
// Owns media state for a single car name. Injected notify for feedback.
import { ref, watch } from 'vue'
import {
  MAX_MEDIA_BYTES,
  IMAGE_EXTENSIONS,
  VIDEO_EXTENSIONS,
  extensionOf,
  mediaTypeFor,
  uniqueSuffix,
  formatFileSize,
  formatDate
} from '../lib/carNameMedia'

export function useCarNameMedia({ carName, notify } = {}) {
  const media = ref([])
  const inactiveCount = ref(0)
  const selected = ref([])
  const rejected = ref([])
  const uploading = ref(false)
  const fileInput = ref(null)

  let safeApiRef = null
  let uploadFileRef = null
  let getFileUrlRef = null

  const fetchMedia = async () => {
    if (!carName || !carName.id) {
      media.value = []
      inactiveCount.value = 0
      return
    }
    const res = await safeApiRef?.({
      query: `
        SELECT cm.id, cm.file_name, cm.file_path, cm.file_type, cm.media_type, cm.is_active, cm.uploaded_at,
               u.username AS uploaded_by_name
        FROM car_name_media cm
        LEFT JOIN users u ON cm.uploaded_by = u.id
        WHERE cm.car_name_id = ?
        ORDER BY cm.uploaded_at DESC
      `,
      params: [carName.id]
    })
    if (res?.success) {
      media.value = res.data
      inactiveCount.value = (res.data || []).filter((m) => m.is_active === 0).length
    } else {
      media.value = []
      inactiveCount.value = 0
    }
  }

  const handleFileSelect = (files) => {
    selected.value = []
    rejected.value = []
    const list = Array.from(files || [])
    list.forEach((f) => {
      const ext = extensionOf(f.name)
      const mtype = mediaTypeFor(ext)
      const isAllowed = mtype === 'photo' ? IMAGE_EXTENSIONS.includes(ext) : VIDEO_EXTENSIONS.includes(ext)
      if (!isAllowed) {
        rejected.value.push({ file: f, reason: `Unsupported file type (.${ext || 'unknown'})` })
        return
      }
      if (f.size > MAX_MEDIA_BYTES) {
        rejected.value.push({ file: f, reason: `File too large (${formatFileSize(f.size)} > 100 MB)` })
        return
      }
      selected.value.push({ file: f })
    })
  }

  const upload = async () => {
    if (!carName || !carName.id || selected.value.length === 0) return { success: false }
    uploading.value = true
    let successCount = 0
    let failCount = 0
    for (const s of selected.value) {
      const f = s.file
      const ext = extensionOf(f.name)
      const mtype = mediaTypeFor(ext)
      const suffix = uniqueSuffix()
      const safeName = f.name.replace(/[^A-Za-z0-9._-]/g, '_')
      const newName = `${suffix}.${ext || 'bin'}`
      const uploadRes = await uploadFileRef?.(f, `uploads/car_names/${carName.id}/${newName}`)
      if (!uploadRes?.success) {
        failCount++
        continue
      }
      const pathRel = uploadRes.path
      const insertRes = await safeApiRef?.({
        query: `INSERT INTO car_name_media (car_name_id, file_path, file_name, file_type, media_type, uploaded_by, is_active)
                VALUES (?, ?, ?, ?, ?, ?, 1)`,
        params: [carName.id, pathRel, safeName, f.type || 'application/octet-stream', mtype, 1]
      })
      if (insertRes?.success) successCount++
      else failCount++
    }
    uploading.value = false
    selected.value = []
    if (fileInput.value) fileInput.value.value = ''
    await fetchMedia()
    if (failCount > 0) {
      await notify?.('error', 'Media upload issue', `${successCount} uploaded, ${failCount} failed.`)
      return { success: false, successCount, failCount }
    }
    return { success: true, successCount }
  }

  const remove = async (item) => {
    const confirmed = await notify?.(
      'warning',
      'Delete media',
      `Delete ${item.file_name}?`,
      'This action cannot be undone.',
      'Delete'
    )
    if (!confirmed) return { success: false, cancelled: true }
    const del = await safeApiRef?.({ query: 'DELETE FROM car_name_media WHERE id = ?', params: [item.id] })
    if (!del?.success) {
      await notify?.('error', 'Failed to delete media', del.error || 'Unknown error')
      return del
    }
    await fetchMedia()
    return { success: true }
  }

  const reset = () => {
    media.value = []
    inactiveCount.value = 0
    selected.value = []
    rejected.value = []
    uploading.value = false
    if (fileInput.value) fileInput.value.value = ''
  }

  function withDeps(deps = {}) {
    const { safeApi, uploadFile, getFileUrl } = deps
    safeApiRef = safeApi
    uploadFileRef = uploadFile
    getFileUrlRef = getFileUrl
    return {
      media,
      inactiveCount,
      selected,
      rejected,
      uploading,
      fileInput,
      fetchMedia,
      handleFileSelect,
      upload,
      remove,
      reset,
      getFileUrl: getFileUrlRef,
      formatFileSize,
      formatDate
    }
  }

  watch(
    () => carName?.id,
    () => {
      reset()
    }
  )

  return { withDeps }
}
