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

/**
 * The signed-in user's id, read from the same session key useApi writes.
 *
 * Returns null when there is no usable session rather than 0 or 1: the INSERT is
 * NOT NULL on uploaded_by, and a wrong-but-valid id is worse than a failure because
 * it silently attributes the file to the wrong person.
 */
function resolveCurrentUserId() {
  try {
    const raw = localStorage.getItem('user')
    if (!raw) return null
    const parsed = JSON.parse(raw)
    const id = Number(parsed?.id)
    return Number.isInteger(id) && id > 0 ? id : null
  } catch {
    return null
  }
}

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
    // uploaded_by was the literal 1. car_name_media.uploaded_by is a NOT NULL
    // foreign key to users.id, so it only ever credited whoever happens to be user 1
    // - and the gallery shows that name next to the file. It is resolved from the
    // session like every other upload path in the app (see useApi.uploadCarFile).
    const uploadedBy = resolveCurrentUserId()
    if (uploadedBy === null) {
      uploading.value = false
      await notify?.(
        'error',
        'Cannot upload media',
        'No signed-in user was found, so the upload could not be attributed. Please sign in again.'
      )
      return { success: false }
    }
    for (const s of selected.value) {
      const f = s.file
      const ext = extensionOf(f.name)
      const mtype = mediaTypeFor(ext)
      const suffix = uniqueSuffix()
      const safeName = f.name.replace(/[^A-Za-z0-9._-]/g, '_')
      const newName = `${suffix}.${ext || 'bin'}`
      // The filename is the THIRD argument (custom_filename) and the folder is the
      // second. They were concatenated into the second here, so uploadFile treated
      // the whole thing as a destination folder: it created
      // files/uploads/car_names/<id>/<uuid>.png/ as a DIRECTORY and stored a
      // server-named file inside it. The gallery then recorded a path that pointed
      // at a folder, so nothing resolved.
      const uploadRes = await uploadFileRef?.(f, `uploads/car_names/${carName.id}`, newName)
      if (!uploadRes?.success) {
        failCount++
        continue
      }
      // uploadFile returns `relativePath`, not `path` - see useApi.js, which returns
      // { ...result, relativePath }. Reading `path` gave undefined, which went into
      // car_name_media.file_path as SQL NULL and every INSERT failed on the NOT NULL
      // constraint. That is the failure that made this look like a no-op: the file
      // was stored, the row was refused, and fetchMedia() then found nothing.
      const pathRel = uploadRes.relativePath
      if (!pathRel) {
        failCount++
        continue
      }
      const insertRes = await safeApiRef?.({
        query: `INSERT INTO car_name_media (car_name_id, file_path, file_name, file_size, file_type, media_type, uploaded_by, is_active)
                VALUES (?, ?, ?, ?, ?, ?, ?, 1)`,
        params: [carName.id, pathRel, safeName, f.size || null, f.type || 'application/octet-stream', mtype, uploadedBy]
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
