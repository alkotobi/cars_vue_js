<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { ElSelect, ElOption } from 'element-plus'
import 'element-plus/dist/index.css'
import PhysicalCopyTracking from './PhysicalCopyTracking.vue'
import CustomClearanceAgents from './CustomClearanceAgents.vue'
import { useApi } from '../../composables/useApi'
import { useEnhancedI18n } from '@/composables/useI18n'

const { t } = useEnhancedI18n()
const {
  getCarFileCategories,
  getCarFiles,
  uploadCarFile,
  deleteCarFile,
  checkoutPhysicalCopy,
  checkinPhysicalCopy,
  transferPhysicalCopy,
  rollbackCheckout,
  getMyPhysicalCopies,
  getPendingTransfers,
  approveTransfer,
  rejectTransfer,
  getUsersForTransfer,
  getFileTransferHistory,
  getFileUrl,
  createFileCategory,
  updateFileCategory,
  deleteFileCategory,
  getCustomClearanceAgents,
  callApi,
} = useApi()

const props = defineProps({
  car: {
    type: Object,
    required: true,
  },
  show: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['close', 'save'])

// State
const loading = ref(false)
const error = ref(null)
const success = ref(null)
const categories = ref([])
const files = ref([])
const selectedFiles = ref({}) // { categoryId: file }
const isProcessing = ref(false)
const pendingTransfers = ref([])
const showPendingTransfersModal = ref(false)
const dragOverCategory = ref(null) // Track which category is being dragged over

// Check if current user is admin
const currentUser = computed(() => {
  const userStr = localStorage.getItem('user')
  return userStr ? JSON.parse(userStr) : null
})

const isAdmin = computed(() => {
  const adminStatus = currentUser.value?.role_id === 1
  // Debug: uncomment to check admin status
  // console.log('isAdmin:', adminStatus, 'currentUser:', currentUser.value)
  return adminStatus
})

// Check if user has can_upload_car_files permission
const canUploadCarFiles = computed(() => {
  if (!currentUser.value) return false
  // Admin has all permissions
  if (isAdmin.value) return true
  // Check for specific permission
  return currentUser.value.permissions?.some((p) => p.permission_name === 'can_upload_car_files')
})

// Get all files that belong to current user (current holder)
// Files can be: 'available', 'checked_out' (to current user), or null (no tracking record)
const myFiles = computed(() => {
  if (!currentUser.value || !files.value) return []
  const result = files.value.filter((f) => {
    // Skip files with pending transfers
    if (f.has_pending_transfer) return false

    // If file is checked out to current user, it belongs to them
    if (f.physical_status === 'checked_out' && f.current_holder_id === currentUser.value.id) {
      return true
    }

    // If file is available and current user is the holder
    if (f.physical_status === 'available' && f.current_holder_id === currentUser.value.id) {
      return true
    }

    // If file has no tracking record (null status), check if current user is the uploader
    if (f.physical_status === null && f.uploaded_by === currentUser.value.id) {
      return true
    }

    return false
  })

  return result
})

// Check if all files belong to current user and are available for batch operations
// Files can be: 'available', 'checked_out' (to current user), or null (no tracking record)
const canBatchOperate = computed(() => {
  if (!currentUser.value || !files.value || files.value.length === 0) {
    return false
  }

  // Get all files that don't have pending transfers
  // Exclude files checked out to someone else
  const eligibleFiles = files.value.filter((f) => {
    if (f.has_pending_transfer) return false

    // Exclude files checked out to someone else
    if (f.physical_status === 'checked_out' && f.current_holder_id !== currentUser.value.id) {
      return false
    }

    // Include: available, checked_out to current user, or null (no tracking)
    return (
      f.physical_status === 'available' ||
      (f.physical_status === 'checked_out' && f.current_holder_id === currentUser.value.id) ||
      f.physical_status === null
    )
  })

  if (eligibleFiles.length === 0) {
    return false
  }

  // Check if all eligible files belong to current user
  const allBelongToUser = eligibleFiles.every((f) => {
    if (f.physical_status === 'checked_out' || f.physical_status === 'available') {
      return f.current_holder_id === currentUser.value.id
    } else if (f.physical_status === null) {
      // No tracking record - check if user is the uploader
      return f.uploaded_by === currentUser.value.id
    }
    return false
  })

  return allBelongToUser
})

// Group files by category
const filesByCategory = computed(() => {
  const grouped = {}
  if (!categories.value || categories.value.length === 0) {
    return grouped
  }
  categories.value.forEach((cat) => {
    grouped[cat.id] = {
      category: cat,
      files: (files.value || []).filter((f) => f.category_id === cat.id),
    }
  })

  return grouped
})

// Importance level labels and colors
const importanceConfig = {
  1: { label: 'CRITICAL', color: '#dc2626', bgColor: '#fef2f2' },
  2: { label: 'HIGH', color: '#ea580c', bgColor: '#fff7ed' },
  3: { label: 'MEDIUM', color: '#ca8a04', bgColor: '#fefce8' },
  4: { label: 'LOW', color: '#65a30d', bgColor: '#f7fee7' },
  5: { label: 'OPTIONAL', color: '#64748b', bgColor: '#f1f5f9' },
}

// Load categories and files
const loadData = async () => {
  loading.value = true
  error.value = null

  try {
    // Load categories
    const cats = await getCarFileCategories()
    categories.value = cats

    // Load files for this car
    const carFiles = await getCarFiles(props.car.id)
    files.value = carFiles
  } catch (err) {
    error.value = err.message || 'Failed to load data'
  } finally {
    loading.value = false
  }
}

// Load pending transfers
const loadPendingTransfers = async () => {
  try {
    const transfers = await getPendingTransfers()
    pendingTransfers.value = transfers || []
  } catch (err) {
    pendingTransfers.value = []
  }
}

// Check if a file has a pending transfer
const hasPendingTransfer = (file) => {
  // First check if the file object has has_pending_transfer flag from API
  if (file.has_pending_transfer === 1 || file.has_pending_transfer === true) {
    return true
  }
  // Fallback: check pending transfers array
  const hasPending = pendingTransfers.value.some((transfer) => transfer.car_file_id === file.id)
  return hasPending
}

// Check if current user is the recipient of a pending transfer
const isPendingTransferRecipient = (file) => {
  if (file.has_pending_transfer && file.pending_transfer_to_user_id === currentUser?.id) {
    return true
  }
  return pendingTransfers.value.some(
    (transfer) => transfer.car_file_id === file.id && transfer.to_user_id === currentUser?.id,
  )
}

// Approve a pending transfer
const handleApproveTransfer = async (transfer) => {
  loading.value = true
  error.value = null

  try {
    await approveTransfer(transfer.id)
    success.value = 'Transfer approved successfully'
    await loadPendingTransfers()
    await loadData() // Reload files to update status
  } catch (err) {
    error.value = err.message || 'Failed to approve transfer'
  } finally {
    loading.value = false
  }
}

// Reject a pending transfer
const handleRejectTransfer = async (transfer) => {
  if (
    !confirm(`Are you sure you want to reject the transfer request for "${transfer.file_name}"?`)
  ) {
    return
  }

  loading.value = true
  error.value = null

  try {
    await rejectTransfer(transfer.id)
    success.value = 'Transfer rejected'
    await loadPendingTransfers()
  } catch (err) {
    error.value = err.message || 'Failed to reject transfer'
  } finally {
    loading.value = false
  }
}

// Approve all pending transfers
const handleApproveAll = async () => {
  if (pendingTransfers.value.length === 0) {
    error.value = 'No pending transfers to approve'
    return
  }

  if (
    !confirm(
      `Are you sure you want to approve all ${pendingTransfers.value.length} pending transfer(s)?`,
    )
  ) {
    return
  }

  loading.value = true
  error.value = null
  success.value = null

  try {
    const results = []
    const errors = []

    // Loop through all pending transfers and approve each one
    for (const transfer of pendingTransfers.value) {
      try {
        await approveTransfer(transfer.id)
        results.push(transfer.file_name)
      } catch (err) {
        errors.push({ file: transfer.file_name, error: err.message || 'Failed to approve' })
      }
    }

    if (errors.length > 0) {
      error.value = `Batch approval completed with errors. ${results.length} succeeded, ${errors.length} failed.`
    } else {
      success.value = `Successfully approved ${results.length} transfer(s)`
    }

    await loadPendingTransfers()
    await loadData() // Reload files to update status
  } catch (err) {
    error.value = err.message || 'Failed to approve all transfers'
  } finally {
    loading.value = false
  }
}

// Reject all pending transfers
const handleRejectAll = async () => {
  if (pendingTransfers.value.length === 0) {
    error.value = 'No pending transfers to reject'
    return
  }

  if (
    !confirm(
      `Are you sure you want to reject all ${pendingTransfers.value.length} pending transfer(s)?`,
    )
  ) {
    return
  }

  loading.value = true
  error.value = null
  success.value = null

  try {
    const results = []
    const errors = []

    // Loop through all pending transfers and reject each one
    for (const transfer of pendingTransfers.value) {
      try {
        await rejectTransfer(transfer.id)
        results.push(transfer.file_name)
      } catch (err) {
        errors.push({ file: transfer.file_name, error: err.message || 'Failed to reject' })
      }
    }

    if (errors.length > 0) {
      error.value = `Batch rejection completed with errors. ${results.length} succeeded, ${errors.length} failed.`
    } else {
      success.value = `Successfully rejected ${results.length} transfer(s)`
    }

    await loadPendingTransfers()
    await loadData() // Reload files to update status
  } catch (err) {
    error.value = err.message || 'Failed to reject all transfers'
  } finally {
    loading.value = false
  }
}

// Check if category already has a file
const categoryHasFile = (categoryId) => {
  const categoryFiles = files.value.filter((f) => f.category_id === categoryId && f.is_active !== 0)
  return categoryFiles.length > 0
}

// Get existing file for a category
const getExistingFileForCategory = (categoryId) => {
  const categoryFiles = files.value.filter((f) => f.category_id === categoryId && f.is_active !== 0)
  return categoryFiles.length > 0 ? categoryFiles[0] : null
}

// Handle file selection
const handleFileSelect = (event, categoryId) => {
  const file = event.target.files?.[0]
  if (!file) return

  // Validate file type
  const validTypes = [
    'application/pdf',
    'image/jpeg',
    'image/png',
    'image/gif',
    'image/webp',
    'application/msword',
    'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    'application/vnd.ms-excel',
    'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  ]

  if (!validTypes.includes(file.type)) {
    error.value = 'Invalid file type. Please upload PDF, images, or Office documents.'
    event.target.value = ''
    return
  }

  // Check file size (10MB limit)
  if (file.size > 10 * 1024 * 1024) {
    error.value = 'File size must be less than 10MB'
    event.target.value = ''
    return
  }

  selectedFiles.value[categoryId] = file
  error.value = null
}

// Handle drag and drop
const handleDragOver = (event, categoryId) => {
  event.preventDefault()
  event.stopPropagation()
  dragOverCategory.value = categoryId
}

const handleDragLeave = (event, categoryId) => {
  event.preventDefault()
  event.stopPropagation()
  // Only clear if we're leaving the drop zone (not just moving to a child element)
  const rect = event.currentTarget.getBoundingClientRect()
  const x = event.clientX
  const y = event.clientY
  if (x < rect.left || x > rect.right || y < rect.top || y > rect.bottom) {
    dragOverCategory.value = null
  }
}

const handleDrop = (event, categoryId) => {
  event.preventDefault()
  event.stopPropagation()
  dragOverCategory.value = null

  const files = event.dataTransfer.files
  if (!files || files.length === 0) return

  // Only handle the first file
  const file = files[0]
  
  // Create a synthetic event object to reuse handleFileSelect logic
  const syntheticEvent = {
    target: {
      files: [file],
      value: ''
    }
  }
  
  handleFileSelect(syntheticEvent, categoryId)
}

// Upload file
const uploadFile = async (categoryId) => {
  const file = selectedFiles.value[categoryId]
  if (!file) return

  // Check if category already has a file
  const existingFile = getExistingFileForCategory(categoryId)
  if (existingFile) {
    // Ask for confirmation to replace
    const confirmReplace = confirm(
      `This category already has a file: "${existingFile.file_name}".\n\nDo you want to replace it with the new file?`
    )
    
    if (!confirmReplace) {
      // User cancelled - clear selection
      delete selectedFiles.value[categoryId]
      const input = document.getElementById(`file-input-${categoryId}`)
      if (input) input.value = ''
      return
    }

    // User confirmed - try to delete old file first (backend will also handle it if this fails)
    isProcessing.value = true
    error.value = null

    try {
      // Try to delete old file for better UX (immediate feedback)
      // If this fails, backend will handle the replacement automatically
      await deleteCarFile(existingFile.id)
    } catch (err) {
      // Log error but continue - backend will handle replacement
      console.warn('Frontend delete failed, backend will handle replacement:', err)
    }
  }

  isProcessing.value = true
  error.value = null

  try {
    const result = await uploadCarFile(file, props.car.id, categoryId)
    success.value = existingFile 
      ? 'File replaced successfully' 
      : 'File uploaded successfully'

    // Reload files
    await loadData()

    // Clear selection
    delete selectedFiles.value[categoryId]

    // Reset file input
    const input = document.getElementById(`file-input-${categoryId}`)
    if (input) input.value = ''

    emit('save', props.car)
  } catch (err) {
    error.value = err.message || 'Failed to upload file'
  } finally {
    isProcessing.value = false
  }
}

// Delete file (admin only)
const handleDeleteFile = async (file) => {
  if (!isAdmin.value) {
    error.value = 'Only admins can delete files'
    return
  }

  const fileName = file.file_name || 'this file'
  const confirmMessage = `Are you sure you want to delete "${fileName}"?\n\nThis will permanently delete:\n- The file from the server\n- The database record\n\nThis action cannot be undone.`

  if (!confirm(confirmMessage)) return

  loading.value = true
  error.value = null

  try {
    await deleteCarFile(file.id)
    success.value = 'File and database record deleted successfully'
    await loadData()
    emit('save', props.car)
  } catch (err) {
    error.value = err.message || 'Failed to delete file'
  } finally {
    loading.value = false
  }
}

// Check out physical copy modal
const showCheckoutModal = ref(false)
const fileToCheckout = ref(null)
const checkoutType = ref('user') // 'user', 'client', or 'custom_clearance_agent'
const checkoutUserId = ref(null)
const checkoutAgentId = ref(null)
const checkoutClientId = ref(null)
const checkoutClientName = ref('')
const checkoutClientContact = ref('')
const checkoutNotes = ref('')
const checkoutAvailableUsers = ref([])
const availableAgents = ref([])
const availableClients = ref([])
const filteredUsers = ref([])
const filteredClients = ref([])
const loadingUsers = ref(false)
const loadingClients = ref(false)

const openCheckoutModal = async (file) => {
  fileToCheckout.value = file
  checkoutType.value = 'user'
  checkoutUserId.value = null
  checkoutAgentId.value = null
  checkoutClientId.value = null
  checkoutClientName.value = ''
  checkoutClientContact.value = ''
  checkoutNotes.value = ''

  // Load initial data for dropdowns
  try {
    await loadUsers()
    await loadClients()
    const agents = await getCustomClearanceAgents()
    availableAgents.value = agents || []
  } catch (err) {
    error.value = 'Failed to load checkout data: ' + (err.message || err)
  }

  showCheckoutModal.value = true
}

// Load users from database
const loadUsers = async () => {
  loadingUsers.value = true
  try {
    const users = await getUsersForTransfer(fileToCheckout.value?.id || 0)
    checkoutAvailableUsers.value = users || []
    filteredUsers.value = users || []
  } catch (err) {
    checkoutAvailableUsers.value = []
    filteredUsers.value = []
  } finally {
    loadingUsers.value = false
  }
}

// Load clients from database
const loadClients = async () => {
  loadingClients.value = true
  try {
    const result = await callApi({
      query:
        'SELECT id, name, email, mobiles, address FROM clients WHERE is_client = 1 ORDER BY name ASC',
      params: [],
      requiresAuth: true,
    })
    if (result.success) {
      availableClients.value = result.data || []
      filteredClients.value = result.data || []
    }
  } catch (err) {
    availableClients.value = []
    filteredClients.value = []
  } finally {
    loadingClients.value = false
  }
}

// Remote method for user search
const searchUsers = (query) => {
  if (!query) {
    filteredUsers.value = checkoutAvailableUsers.value
    return
  }
  const searchTerm = query.toLowerCase()
  filteredUsers.value = checkoutAvailableUsers.value.filter(
    (user) =>
      user.username?.toLowerCase().includes(searchTerm) ||
      user.email?.toLowerCase().includes(searchTerm),
  )
}

// Remote method for client search
const searchClients = (query) => {
  if (!query) {
    filteredClients.value = availableClients.value
    return
  }
  const searchTerm = query.toLowerCase()
  filteredClients.value = availableClients.value.filter(
    (client) =>
      client.name?.toLowerCase().includes(searchTerm) ||
      client.email?.toLowerCase().includes(searchTerm) ||
      client.mobiles?.toLowerCase().includes(searchTerm),
  )
}

// Handle user selection
const handleUserChange = (userId) => {
  if (userId) {
    const user = checkoutAvailableUsers.value.find((u) => u.id === userId)
    if (user) {
      checkoutClientName.value = ''
      checkoutClientContact.value = ''
      checkoutClientId.value = null
    }
  }
}

// Handle client selection
const handleClientChange = (clientId) => {
  if (clientId) {
    const client = availableClients.value.find((c) => c.id === clientId)
    if (client) {
      checkoutClientId.value = client.id
      checkoutClientName.value = client.name
      checkoutClientContact.value = client.mobiles || client.email || ''
      checkoutUserId.value = null
    }
  }
}

const confirmCheckout = async () => {
  if (!fileToCheckout.value) {
    return
  }

  // Validate based on checkout type
  if (checkoutType.value === 'user' && !checkoutUserId.value) {
    error.value = 'Please select a user'
    return
  }
  if (checkoutType.value === 'custom_clearance_agent' && !checkoutAgentId.value) {
    error.value = 'Veuillez sélectionner un transiteur'
    return
  }
  if (checkoutType.value === 'client' && !checkoutClientId.value) {
    error.value = 'Veuillez sélectionner un client'
    return
  }

  loading.value = true
  error.value = null
  success.value = null

  try {
    const user = currentUser.value

    const result = await callApi({
      action: 'checkout_physical_copy',
      file_id: fileToCheckout.value.id,
      checkout_type: checkoutType.value,
      user_id: checkoutType.value === 'user' ? checkoutUserId.value : null,
      agent_id: checkoutType.value === 'custom_clearance_agent' ? checkoutAgentId.value : null,
      client_id: checkoutType.value === 'client' ? checkoutClientId.value : null,
      client_name: checkoutType.value === 'client' ? checkoutClientName.value : null,
      client_contact: checkoutType.value === 'client' ? checkoutClientContact.value : null,
      notes: checkoutNotes.value || null,
      performed_by: user?.id,
      requiresAuth: true,
    })

    if (result.success) {
      success.value = 'File checked out successfully'
      showCheckoutModal.value = false
      await loadData()
    } else {
      throw new Error(result.error || 'Failed to check out file')
    }
  } catch (err) {
    error.value = err.message || err.error || 'Failed to check out file'
  } finally {
    loading.value = false
  }
}

// Rollback checkout (admin only)
const handleRollbackCheckout = async (file) => {
  if (
    !confirm(
      `Are you sure you want to rollback the last transfer/checkout for "${file.file_name}"? This will restore the previous holder.`,
    )
  ) {
    return
  }

  loading.value = true
  error.value = null
  success.value = null

  try {
    // Pass null to let the API generate the note with the actual username
    await rollbackCheckout(file.id, null)
    success.value = 'Transfer/checkout rolled back successfully'
    await loadData()
  } catch (err) {
    const errorMsg = err.message || 'Failed to rollback checkout'
    if (errorMsg.includes('initial file upload') || errorMsg.includes('Cannot rollback initial')) {
      error.value =
        'Cannot rollback initial file upload. Use delete instead if you want to remove the file.'
    } else {
      error.value = errorMsg
    }
  } finally {
    loading.value = false
  }
}

// Check in physical copy modal
const showCheckinModal = ref(false)
const fileToCheckin = ref(null)
const checkinNotes = ref('')

const openCheckinModal = (file) => {
  fileToCheckin.value = file
  checkinNotes.value = ''
  showCheckinModal.value = true
}

const confirmCheckin = async () => {
  if (!fileToCheckin.value) return

  loading.value = true
  error.value = null

  try {
    await checkinPhysicalCopy(fileToCheckin.value.id, checkinNotes.value || null)
    success.value = 'File checked in successfully'
    showCheckinModal.value = false
    await loadData()
  } catch (err) {
    error.value = err.message || 'Failed to check in file'
  } finally {
    loading.value = false
  }
}

// Transfer physical copy
const showTransferModal = ref(false)
const fileToTransfer = ref(null)
const transferToUserId = ref(null)
const transferNotes = ref('')
const transferExpectedReturn = ref('')
const transferAvailableUsers = ref([])

// Transfer history modal
const showHistoryModal = ref(false)
const fileForHistory = ref(null)
const transferHistory = ref([])

const openTransferModal = async (file) => {
  fileToTransfer.value = file
  transferToUserId.value = null
  transferNotes.value = ''
  transferExpectedReturn.value = ''

  try {
    const users = await getUsersForTransfer(file.id)
    transferAvailableUsers.value = users
    showTransferModal.value = true
  } catch (err) {
    error.value = err.message || 'Failed to load users'
  }
}

const confirmTransfer = async () => {
  if (!fileToTransfer.value || !transferToUserId.value) {
    error.value = 'Please select a user to transfer to'
    return
  }

  loading.value = true
  error.value = null

  try {
    await transferPhysicalCopy(
      fileToTransfer.value.id,
      transferToUserId.value,
      transferNotes.value || null,
      transferExpectedReturn.value || null,
    )
    success.value = 'Transfer request sent. Waiting for recipient approval.'
    showTransferModal.value = false
    // Reload pending transfers first, then reload file data
    await loadPendingTransfers()
    await loadData()
  } catch (err) {
    error.value = err.message || 'Failed to transfer file'
  } finally {
    loading.value = false
  }
}

// Batch Transfer
const showBatchTransferModal = ref(false)
const batchTransferToUserId = ref(null)
const batchTransferNotes = ref('')
const batchTransferExpectedReturn = ref('')
const batchTransferAvailableUsers = ref([])

const openBatchTransferModal = async () => {
  if (!canBatchOperate.value) {
    error.value =
      'Cannot perform batch operation: Not all documents belong to you. Please ensure all documents are available and you are the current holder.'
    return
  }

  // Get eligible files - include files checked out to current user, available, or null
  const eligibleFiles = files.value.filter((f) => {
    if (f.has_pending_transfer) return false

    // Include files checked out to current user
    if (f.physical_status === 'checked_out' && f.current_holder_id === currentUser.value.id) {
      return true
    }

    // Include available files where current user is holder
    if (f.physical_status === 'available' && f.current_holder_id === currentUser.value.id) {
      return true
    }

    // Include files with no tracking record where current user is uploader
    if (f.physical_status === null && f.uploaded_by === currentUser.value.id) {
      return true
    }

    return false
  })

  if (eligibleFiles.length === 0) {
    error.value = 'No files available for batch transfer'
    return
  }

  batchTransferToUserId.value = null
  batchTransferNotes.value = ''
  batchTransferExpectedReturn.value = ''

  try {
    // Use first file to get users (they should be the same for all files)
    const users = await getUsersForTransfer(eligibleFiles[0].id)
    batchTransferAvailableUsers.value = users
    showBatchTransferModal.value = true
  } catch (err) {
    error.value = err.message || 'Failed to load users'
  }
}

const confirmBatchTransfer = async () => {
  if (!batchTransferToUserId.value) {
    error.value = 'Please select a user to transfer to'
    return
  }

  // Get eligible files - include files checked out to current user, available, or null
  const eligibleFiles = files.value.filter((f) => {
    if (f.has_pending_transfer) return false

    // Include files checked out to current user
    if (f.physical_status === 'checked_out' && f.current_holder_id === currentUser.value.id) {
      return true
    }

    // Include available files where current user is holder
    if (f.physical_status === 'available' && f.current_holder_id === currentUser.value.id) {
      return true
    }

    // Include files with no tracking record where current user is uploader
    if (f.physical_status === null && f.uploaded_by === currentUser.value.id) {
      return true
    }

    return false
  })

  if (eligibleFiles.length === 0) {
    error.value = 'No files available for batch transfer'
    return
  }

  // Double-check: ensure all files belong to current user
  const notMyFiles = eligibleFiles.filter((f) => f.current_holder_id !== currentUser.value.id)
  if (notMyFiles.length > 0) {
    error.value =
      'Cannot perform batch transfer: Some documents do not belong to you. Please try again.'
    return
  }

  loading.value = true
  error.value = null
  success.value = null

  try {
    const results = []
    const errors = []

    // Loop through files and transfer each one
    for (const file of eligibleFiles) {
      try {
        await transferPhysicalCopy(
          file.id,
          batchTransferToUserId.value,
          batchTransferNotes.value || null,
          batchTransferExpectedReturn.value || null,
        )
        results.push(file.file_name)
      } catch (err) {
        errors.push({ file: file.file_name, error: err.message || 'Failed to transfer' })
      }
    }

    if (errors.length > 0) {
      error.value = `Batch transfer completed with errors. ${results.length} succeeded, ${errors.length} failed.`
    } else {
      success.value = `Successfully transferred ${results.length} file(s)`
    }

    showBatchTransferModal.value = false
    await loadPendingTransfers()
    await loadData()
  } catch (err) {
    error.value = err.message || 'Failed to perform batch transfer'
  } finally {
    loading.value = false
  }
}

// Batch Checkout
const showBatchCheckoutModal = ref(false)
const batchCheckoutType = ref('user') // 'user', 'client', or 'custom_clearance_agent'
const batchCheckoutUserId = ref(null)
const batchCheckoutAgentId = ref(null)
const batchCheckoutClientId = ref(null)
const batchCheckoutClientName = ref('')
const batchCheckoutClientContact = ref('')
const batchCheckoutNotes = ref('')
const batchCheckoutAvailableUsers = ref([])
const batchAvailableAgents = ref([])
const batchAvailableClients = ref([])

const openBatchCheckoutModal = async () => {
  if (!canBatchOperate.value) {
    error.value =
      'Cannot perform batch operation: Not all documents belong to you. Please ensure all documents are available and you are the current holder.'
    return
  }

  // Get eligible files - include files checked out to current user, available, or null
  const eligibleFiles = files.value.filter((f) => {
    if (f.has_pending_transfer) return false

    // Include files checked out to current user
    if (f.physical_status === 'checked_out' && f.current_holder_id === currentUser.value.id) {
      return true
    }

    // Include available files where current user is holder
    if (f.physical_status === 'available' && f.current_holder_id === currentUser.value.id) {
      return true
    }

    // Include files with no tracking record where current user is uploader
    if (f.physical_status === null && f.uploaded_by === currentUser.value.id) {
      return true
    }

    return false
  })

  if (eligibleFiles.length === 0) {
    error.value = 'No files available for batch checkout'
    return
  }

  batchCheckoutType.value = 'user'
  batchCheckoutUserId.value = null
  batchCheckoutAgentId.value = null
  batchCheckoutClientId.value = null
  batchCheckoutClientName.value = ''
  batchCheckoutClientContact.value = ''
  batchCheckoutNotes.value = ''

  try {
    await loadUsers()
    await loadClients()
    const agents = await getCustomClearanceAgents()
    batchAvailableAgents.value = agents || []
    showBatchCheckoutModal.value = true
  } catch (err) {
    error.value = 'Failed to load checkout data: ' + (err.message || err)
  }
}

const confirmBatchCheckout = async () => {
  // Get eligible files - include files checked out to current user, available, or null
  const eligibleFiles = files.value.filter((f) => {
    if (f.has_pending_transfer) return false

    // Include files checked out to current user
    if (f.physical_status === 'checked_out' && f.current_holder_id === currentUser.value.id) {
      return true
    }

    // Include available files where current user is holder
    if (f.physical_status === 'available' && f.current_holder_id === currentUser.value.id) {
      return true
    }

    // Include files with no tracking record where current user is uploader
    if (f.physical_status === null && f.uploaded_by === currentUser.value.id) {
      return true
    }

    return false
  })

  if (eligibleFiles.length === 0) {
    error.value = 'No files available for batch checkout'
    return
  }

  // Double-check: ensure all files belong to current user
  const notMyFiles = eligibleFiles.filter((f) => f.current_holder_id !== currentUser.value.id)
  if (notMyFiles.length > 0) {
    error.value =
      'Cannot perform batch checkout: Some documents do not belong to you. Please try again.'
    return
  }

  // Validate based on checkout type
  if (batchCheckoutType.value === 'user' && !batchCheckoutUserId.value) {
    error.value = 'Please select a user'
    return
  }
  if (batchCheckoutType.value === 'custom_clearance_agent' && !batchCheckoutAgentId.value) {
    error.value = 'Veuillez sélectionner un transiteur'
    return
  }
  if (batchCheckoutType.value === 'client' && !batchCheckoutClientId.value) {
    error.value = 'Veuillez sélectionner un client'
    return
  }

  loading.value = true
  error.value = null
  success.value = null

  try {
    const user = currentUser.value
    const results = []
    const errors = []

    // Loop through files and checkout each one
    for (const file of eligibleFiles) {
      try {
        const result = await callApi({
          action: 'checkout_physical_copy',
          file_id: file.id,
          checkout_type: batchCheckoutType.value,
          user_id: batchCheckoutType.value === 'user' ? batchCheckoutUserId.value : null,
          agent_id:
            batchCheckoutType.value === 'custom_clearance_agent'
              ? batchCheckoutAgentId.value
              : null,
          client_id: batchCheckoutType.value === 'client' ? batchCheckoutClientId.value : null,
          client_name: batchCheckoutType.value === 'client' ? batchCheckoutClientName.value : null,
          client_contact:
            batchCheckoutType.value === 'client' ? batchCheckoutClientContact.value : null,
          notes: batchCheckoutNotes.value || null,
          performed_by: user?.id,
          requiresAuth: true,
        })

        if (result.success) {
          results.push(file.file_name)
        } else {
          errors.push({ file: file.file_name, error: result.error || 'Failed to checkout' })
        }
      } catch (err) {
        errors.push({ file: file.file_name, error: err.message || 'Failed to checkout' })
      }
    }

    if (errors.length > 0) {
      error.value = `Batch checkout completed with errors. ${results.length} succeeded, ${errors.length} failed.`
    } else {
      success.value = `Successfully checked out ${results.length} file(s)`
    }

    showBatchCheckoutModal.value = false
    await loadData()
  } catch (err) {
    error.value = err.message || 'Failed to perform batch checkout'
  } finally {
    loading.value = false
  }
}

// Transfer history
const openHistoryModal = async (file) => {
  fileForHistory.value = file
  loading.value = true
  error.value = null
  showHistoryModal.value = true // Show modal immediately, even while loading

  try {
    const history = await getFileTransferHistory(file.id)
    transferHistory.value = history || []
  } catch (err) {
    error.value = err.message || 'Failed to load transfer history'
  } finally {
    loading.value = false
  }
}

// Create/Edit Category Modal
const showCategoryModal = ref(false)
const editingCategory = ref(null)
const categoryForm = ref({
  category_name: '',
  importance_level: 3,
  is_required: false,
  display_order: 0,
  description: '',
})

const openCategoryModal = (category = null) => {
  if (category) {
    editingCategory.value = category
    categoryForm.value = {
      category_name: category.category_name,
      importance_level: category.importance_level,
      is_required: category.is_required === 1 || category.is_required === true,
      display_order: category.display_order,
      description: category.description || '',
    }
  } else {
    editingCategory.value = null
    categoryForm.value = {
      category_name: '',
      importance_level: 3,
      is_required: false,
      display_order: categories.value.length,
      description: '',
    }
  }
  showCategoryModal.value = true
}

const saveCategory = async () => {
  if (!categoryForm.value.category_name.trim()) {
    error.value = 'Category name is required'
    return
  }

  loading.value = true
  error.value = null

  try {
    if (editingCategory.value) {
      // Update existing category
      await updateFileCategory(editingCategory.value.id, categoryForm.value)
      success.value = 'Category updated successfully'
    } else {
      // Create new category
      await createFileCategory(categoryForm.value)
      success.value = 'Category created successfully'
    }
    showCategoryModal.value = false
    await loadData()
    // Reload categories table if it's open
    if (showCategoriesTableModal.value) {
      await loadCategoriesForTable()
    }
  } catch (err) {
    error.value = err.message || 'Failed to save category'
  } finally {
    loading.value = false
  }
}

// Categories Table Modal
const showCategoriesTableModal = ref(false)
const categoriesForTable = ref([])
const loadingCategories = ref(false)

const loadCategoriesForTable = async () => {
  loadingCategories.value = true
  try {
    const cats = await getCarFileCategories()
    categoriesForTable.value = cats || []
  } catch (err) {
    error.value = err.message || 'Failed to load categories'
  } finally {
    loadingCategories.value = false
  }
}

const openCategoriesTableModal = async () => {
  showCategoriesTableModal.value = true
  await loadCategoriesForTable()
}

const handleDeleteCategory = async (category) => {
  if (!confirm(`Are you sure you want to delete the category "${category.category_name}"?`)) {
    return
  }

  loading.value = true
  error.value = null

  try {
    await deleteFileCategory(category.id)
    success.value = 'Category deleted successfully'
    await loadData()
    await loadCategoriesForTable()
  } catch (err) {
    error.value = err.message || 'Failed to delete category'
  } finally {
    loading.value = false
  }
}

const handleEditCategoryFromTable = (category) => {
  showCategoriesTableModal.value = false
  openCategoryModal(category)
}

// Physical Copy Tracking Modal
const showTrackingModal = ref(false)

const openTrackingModal = () => {
  showTrackingModal.value = true
}

// Custom Clearance Agents Modal
const showAgentsModal = ref(false)

const closeModal = () => {
  error.value = null
  success.value = null
  selectedFiles.value = {}
  showTransferModal.value = false
  showCategoryModal.value = false
  showCategoriesTableModal.value = false
  showCheckoutModal.value = false
  showCheckinModal.value = false
  showHistoryModal.value = false
  emit('close')
}

// Format date helper
const formatDate = (dateString) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleDateString()
}

// Format file size helper
const formatFileSize = (bytes) => {
  if (!bytes) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}

// Watch for show prop changes
watch(
  () => props.show,
  (newVal) => {
    if (newVal) {
      loadData()
      loadPendingTransfers()
    } else {
      closeModal()
    }
  },
  { immediate: true },
)

// Watch for pending transfers modal to reload when opened
watch(
  () => showPendingTransfersModal.value,
  (newVal) => {
    if (newVal) {
      loadPendingTransfers()
    }
  },
)

// Load data on mount if shown
onMounted(() => {
  if (props.show) {
    loadData()
    loadPendingTransfers()
  }
})
</script>

<template>
  <div v-if="show" class="modal-overlay" @click="closeModal">
    <div class="modal-content" @click.stop>
      <div class="modal-header">
        <h3>
          <i class="fas fa-folder-open"></i>
          Car Files Management
        </h3>
        <div class="header-actions">
          <button
            v-if="canUploadCarFiles"
            @click="showPendingTransfersModal = true"
            class="btn-pending-transfers"
            title="View pending transfers"
            :class="{ 'has-pending': pendingTransfers.length > 0 }"
          >
            <i class="fas fa-clock"></i>
            Pending Transfers
            <span v-if="pendingTransfers.length > 0" class="pending-badge">
              {{ pendingTransfers.length }}
            </span>
          </button>
          <button
            v-if="isAdmin && canUploadCarFiles"
            @click="showAgentsModal = true"
            class="btn-agents"
            title="Gérer les transiteurs"
          >
            <i class="fas fa-user-tie"></i>
            Agents
          </button>
          <button
            v-if="canUploadCarFiles"
            @click="openTrackingModal"
            class="btn-tracking"
            title="View physical copy tracking"
          >
            <i class="fas fa-clipboard-list"></i>
            Tracking
          </button>
          <button
            v-if="isAdmin || true"
            @click="openCategoriesTableModal"
            class="btn-categories"
            title="Manage categories"
          >
            <i class="fas fa-list"></i>
            Categories
          </button>
          <button
            v-if="isAdmin"
            @click="openCategoryModal()"
            class="btn-add-category"
            title="Add new category"
          >
            <i class="fas fa-plus"></i>
            Add Category
          </button>
          <button
            v-if="canUploadCarFiles"
            @click="openBatchTransferModal"
            class="btn-batch-transfer"
            :class="{ disabled: !canBatchOperate }"
            :disabled="!canBatchOperate"
            :title="
              canBatchOperate
                ? 'Batch transfer all documents'
                : 'Batch transfer unavailable: Not all documents belong to you'
            "
          >
            <i class="fas fa-exchange-alt"></i>
            Batch Transfer
            <span v-if="myFiles.length > 0" class="batch-count">({{ myFiles.length }})</span>
          </button>
          <button
            v-if="canUploadCarFiles"
            @click="openBatchCheckoutModal"
            class="btn-batch-checkout"
            :class="{ disabled: !canBatchOperate }"
            :disabled="!canBatchOperate"
            :title="
              canBatchOperate
                ? 'Batch checkout all documents'
                : 'Batch checkout unavailable: Not all documents belong to you'
            "
          >
            <i class="fas fa-sign-out-alt"></i>
            Batch Checkout
            <span v-if="myFiles.length > 0" class="batch-count">({{ myFiles.length }})</span>
          </button>
          <button class="close-btn" @click="closeModal" :disabled="isProcessing">
            <i class="fas fa-times"></i>
          </button>
        </div>
      </div>

      <div class="modal-body">
        <!-- Loading State -->
        <div v-if="loading && categories.length === 0" class="loading-state">
          <i class="fas fa-spinner fa-spin"></i>
          <span>Loading categories...</span>
        </div>

        <!-- Error Message -->
        <div v-if="error" class="error-message">
          <i class="fas fa-exclamation-circle"></i>
          {{ error }}
        </div>

        <!-- Success Message -->
        <div v-if="success" class="success-message">
          <i class="fas fa-check-circle"></i>
          {{ success }}
        </div>

        <!-- Empty State - No Categories -->
        <div v-if="!loading && categories.length === 0" class="empty-state">
          <i class="fas fa-folder-open"></i>
          <p>No file categories found. Please run the database migration first.</p>
        </div>

        <!-- Files by Category -->
        <div
          v-for="(group, categoryId) in filesByCategory"
          :key="categoryId"
          class="category-section"
          :class="`importance-${group.category.importance_level}`"
        >
          <div class="category-header">
            <div class="category-info">
              <span
                class="importance-badge"
                :style="{
                  backgroundColor: importanceConfig[group.category.importance_level].bgColor,
                  color: importanceConfig[group.category.importance_level].color,
                }"
              >
                {{ importanceConfig[group.category.importance_level].label }}
              </span>
              <h4>
                {{ group.category.category_name }}
                <span v-if="group.category.is_required" class="required-badge">Required</span>
              </h4>
              <p v-if="group.category.description" class="category-description">
                {{ group.category.description }}
              </p>
            </div>
          </div>

          <!-- Files List -->
          <div class="files-list">
            <div v-for="file in group.files" :key="file.id" class="file-item">
              <div class="file-info">
                <i class="fas fa-file-alt"></i>
                <div>
                  <a :href="getFileUrl(file.file_path)" target="_blank" class="file-name">
                    {{ file.file_name }}
                  </a>
                  <div class="file-meta">
                    Uploaded by {{ file.uploaded_by_username }} on
                    {{ new Date(file.uploaded_at).toLocaleDateString() }}
                  </div>
                </div>
              </div>

              <div class="file-actions">
                <!-- Physical Copy Status -->
                <div class="physical-status">
                  <!-- Check if file has pending transfer -->
                  <span
                    v-if="hasPendingTransfer(file)"
                    class="status-badge pending-transfer"
                    title="Pending transfer - waiting for approval"
                  >
                    <i class="fas fa-clock"></i> Pending Transfer
                  </span>
                  <span
                    v-else-if="!file.physical_status || file.physical_status === 'available'"
                    class="status-badge available"
                  >
                    <i class="fas fa-check-circle"></i> Available
                  </span>
                  <span
                    v-else-if="file.physical_status === 'checked_out'"
                    class="status-badge checked-out"
                  >
                    <i v-if="file.checkout_type === 'user'" class="fas fa-user"></i>
                    <i
                      v-else-if="file.checkout_type === 'custom_clearance_agent'"
                      class="fas fa-user-tie"
                    ></i>
                    <i v-else-if="file.checkout_type === 'client'" class="fas fa-building"></i>
                    <span class="holder-info">
                      <strong v-if="file.checkout_type === 'user'">
                        Utilisateur: {{ file.current_holder_username || 'Unknown User' }}
                      </strong>
                      <strong v-else-if="file.checkout_type === 'custom_clearance_agent'">
                        Transiteur: {{ file.agent_name || 'Unknown Agent' }}
                      </strong>
                      <strong v-else-if="file.checkout_type === 'client'">
                        Client: {{ file.client_name || 'Unknown Client' }}
                      </strong>
                      <span v-else>
                        {{
                          file.current_holder_username ||
                          file.agent_name ||
                          file.client_name ||
                          'Unknown'
                        }}
                      </span>
                    </span>
                  </span>
                </div>

                <!-- Show message when file has pending transfer -->
                <template v-if="hasPendingTransfer(file)">
                  <div class="pending-transfer-message">
                    <i class="fas fa-clock"></i>
                    <span v-if="isPendingTransferRecipient(file)">
                      Transfer pending - Click "Pending Transfers" button above to approve/reject
                    </span>
                    <span v-else> Transfer pending - waiting for recipient approval </span>
                    <button
                      @click="showPendingTransfersModal = true"
                      class="btn-view-pending"
                      title="View pending transfers"
                    >
                      View
                    </button>
                  </div>
                </template>

                <!-- Actions - Only show if file is available (not checked out) and no pending transfer -->
                <!-- Note: After upload, file should be checked out to uploader, so this should rarely show -->
                <template v-else-if="file.physical_status !== 'checked_out'">
                  <div v-if="canUploadCarFiles" class="action-buttons">
                    <!-- Only current holder (uploader) can checkout -->
                    <button
                      v-if="
                        file.current_holder_id === currentUser?.id ||
                        (file.physical_status === 'available' &&
                          file.uploaded_by === currentUser?.id)
                      "
                      @click="openCheckoutModal(file)"
                      class="btn-action btn-checkout"
                      title="Check out physical copy"
                    >
                      <i class="fas fa-sign-out-alt"></i>
                      Check Out
                    </button>
                    <!-- Only current holder can transfer -->
                    <button
                      v-if="
                        file.current_holder_id === currentUser?.id ||
                        (file.physical_status === 'available' &&
                          file.uploaded_by === currentUser?.id)
                      "
                      @click="openTransferModal(file)"
                      class="btn-action btn-transfer"
                      title="Transfer to another user"
                    >
                      <i class="fas fa-exchange-alt"></i>
                      Transfer
                    </button>
                    <!-- History button - visible to everyone with permission -->
                    <button
                      @click="openHistoryModal(file)"
                      class="btn-action btn-history"
                      title="View transfer history"
                    >
                      <i class="fas fa-history"></i>
                      History
                    </button>
                    <button
                      v-if="isAdmin"
                      @click="handleDeleteFile(file)"
                      class="btn-action btn-delete"
                      title="Delete file (admin only)"
                    >
                      <i class="fas fa-trash"></i>
                    </button>
                  </div>
                  <div v-else class="no-permission-message">
                    <i class="fas fa-lock"></i>
                    <span>You do not have permission to manage car files</span>
                  </div>
                </template>

                <!-- Show actions when checked out -->
                <template v-else>
                  <!-- Only show actions if user has permission -->
                  <template v-if="canUploadCarFiles">
                    <!-- If current user is the holder, show Transfer, Checkout, History, and Rollback buttons -->
                    <div v-if="file.current_holder_id === currentUser?.id" class="action-buttons">
                      <button
                        @click="openTransferModal(file)"
                        class="btn-action btn-transfer"
                        title="Transfer to another user"
                      >
                        <i class="fas fa-exchange-alt"></i>
                        Transfer
                      </button>
                      <button
                        @click="openCheckoutModal(file)"
                        class="btn-action btn-checkout"
                        title="Check out to user/client/agent"
                      >
                        <i class="fas fa-sign-out-alt"></i>
                        Check Out
                      </button>
                      <button
                        @click="openHistoryModal(file)"
                        class="btn-action btn-history"
                        title="View transfer history"
                      >
                        <i class="fas fa-history"></i>
                        History
                      </button>
                      <!-- Rollback - only current holder can rollback -->
                      <button
                        @click="handleRollbackCheckout(file)"
                        class="btn-action btn-rollback"
                        title="Rollback last transfer/checkout"
                      >
                        <i class="fas fa-undo"></i>
                        Rollback
                      </button>
                      <!-- Delete button - admin only -->
                      <button
                        v-if="isAdmin"
                        @click="handleDeleteFile(file)"
                        class="btn-action btn-delete"
                        title="Delete file (admin only)"
                      >
                        <i class="fas fa-trash"></i>
                      </button>
                    </div>
                    <!-- If checked out to client/transiteur (current_holder_id is NULL), show rollback to previous holder -->
                    <div
                      v-else-if="
                        !file.current_holder_id && file.previous_holder_id === currentUser?.id
                      "
                      class="action-buttons"
                    >
                      <!-- Rollback - previous holder can rollback when checked out to client/transiteur -->
                      <button
                        @click="handleRollbackCheckout(file)"
                        class="btn-action btn-rollback"
                        title="Rollback checkout to client/transiteur"
                      >
                        <i class="fas fa-undo"></i>
                        Rollback
                      </button>
                      <!-- History button - visible to everyone with permission -->
                      <button
                        @click="openHistoryModal(file)"
                        class="btn-action btn-history"
                        title="View transfer history"
                      >
                        <i class="fas fa-history"></i>
                        History
                      </button>
                      <!-- Delete button - admin only -->
                      <button
                        v-if="isAdmin"
                        @click="handleDeleteFile(file)"
                        class="btn-action btn-delete"
                        title="Delete file (admin only)"
                      >
                        <i class="fas fa-trash"></i>
                      </button>
                    </div>
                    <!-- If someone else has it or checked out to client/transiteur (but not previous holder), show message with History button -->
                    <div v-else class="checked-out-message">
                      <div style="display: flex; align-items: center; gap: 0.5rem; flex: 1">
                        <i class="fas fa-info-circle"></i>
                        <span>Document vérifié - Aucune action disponible</span>
                      </div>
                      <div style="display: flex; gap: 0.5rem">
                        <!-- History button - visible to everyone with permission -->
                        <button
                          @click="openHistoryModal(file)"
                          class="btn-action btn-history"
                          title="View transfer history"
                        >
                          <i class="fas fa-history"></i>
                          History
                        </button>
                        <!-- Delete button - admin only -->
                        <button
                          v-if="isAdmin"
                          @click="handleDeleteFile(file)"
                          class="btn-action btn-delete"
                          title="Delete file (admin only)"
                        >
                          <i class="fas fa-trash"></i>
                        </button>
                      </div>
                    </div>
                  </template>
                  <!-- No permission message -->
                  <div v-else class="no-permission-message">
                    <i class="fas fa-lock"></i>
                    <span>You do not have permission to manage car files</span>
                  </div>
                </template>
              </div>
            </div>

            <!-- Upload Section - Show if user has permission -->
            <div
              v-if="canUploadCarFiles"
              class="upload-section"
              :class="{
                'has-existing-file': categoryHasFile(categoryId),
                'drag-over': dragOverCategory === categoryId
              }"
              @dragover.prevent="handleDragOver($event, categoryId)"
              @dragleave.prevent="handleDragLeave($event, categoryId)"
              @drop.prevent="handleDrop($event, categoryId)"
            >
              <div class="drop-zone" :class="{ 'drag-active': dragOverCategory === categoryId }">
                <div v-if="!selectedFiles[categoryId]" class="drop-zone-content">
                  <i class="fas fa-cloud-upload-alt"></i>
                  <p class="drop-zone-text">
                    <span class="drop-zone-main-text">Drag & drop a file here</span>
                    <span class="drop-zone-sub-text">or click to browse</span>
                  </p>
                </div>
                <div v-else class="drop-zone-content file-selected">
                  <i class="fas fa-file"></i>
                  <p class="drop-zone-text">
                    <span class="drop-zone-main-text">{{ selectedFiles[categoryId].name }}</span>
                    <span class="drop-zone-sub-text">{{ formatFileSize(selectedFiles[categoryId].size) }}</span>
                  </p>
                  <button
                    @click.stop="
                      delete selectedFiles[categoryId];
                      const input = document.getElementById(`file-input-${categoryId}`);
                      if (input) input.value = '';
                    "
                    class="btn-remove-file"
                    title="Remove file"
                  >
                    <i class="fas fa-times"></i>
                  </button>
                </div>
                <input
                  :id="`file-input-${categoryId}`"
                  type="file"
                  class="file-input-hidden"
                  @change="(e) => handleFileSelect(e, categoryId)"
                  :disabled="isProcessing"
                  accept=".pdf,.jpg,.jpeg,.png,.gif,.webp,.doc,.docx,.xls,.xlsx"
                />
              </div>
              <button
                @click="uploadFile(categoryId)"
                :disabled="!selectedFiles[categoryId] || isProcessing"
                class="btn-upload"
                :class="{ 'btn-replace': categoryHasFile(categoryId) && selectedFiles[categoryId] }"
              >
                <i class="fas" :class="categoryHasFile(categoryId) && selectedFiles[categoryId] ? 'fa-exchange-alt' : 'fa-upload'"></i>
                {{
                  selectedFiles[categoryId]
                    ? categoryHasFile(categoryId)
                      ? `Replace with ${selectedFiles[categoryId].name}`
                      : `Upload ${selectedFiles[categoryId].name}`
                    : 'Select File'
                }}
              </button>
              <div v-if="categoryHasFile(categoryId) && !selectedFiles[categoryId]" class="upload-info-message">
                <i class="fas fa-info-circle"></i>
                <span>This category has a file. Drag & drop or select a new file to replace it.</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn-close" @click="closeModal" :disabled="isProcessing">Close</button>
      </div>
    </div>

    <!-- Transfer Modal -->
    <div v-if="showTransferModal" class="modal-overlay" @click="showTransferModal = false">
      <div class="modal-content transfer-modal" @click.stop>
        <div class="modal-header">
          <h3>Transfer Physical Copy</h3>
          <button @click="showTransferModal = false" class="close-btn">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>File:</label>
            <p>{{ fileToTransfer?.file_name }}</p>
          </div>
          <div class="form-group">
            <label>Transfer to:</label>
            <select v-model="transferToUserId" class="form-select">
              <option value="">Select user...</option>
              <option v-for="user in transferAvailableUsers" :key="user.id" :value="user.id">
                {{ user.username }} ({{ user.email }})
              </option>
            </select>
          </div>
          <div class="form-group">
            <label>Expected Return Date:</label>
            <input v-model="transferExpectedReturn" type="date" class="form-input" />
          </div>
          <div class="form-group">
            <label>Notes:</label>
            <textarea v-model="transferNotes" class="form-textarea" rows="3"></textarea>
          </div>
        </div>
        <div class="modal-footer">
          <button @click="showTransferModal = false" class="btn-cancel">Cancel</button>
          <button
            @click="confirmTransfer"
            :disabled="!transferToUserId || loading"
            class="btn-primary"
          >
            Transfer
          </button>
        </div>
      </div>
    </div>

    <!-- Checkout Modal -->
    <div v-if="showCheckoutModal" class="modal-overlay" @click="showCheckoutModal = false">
      <div class="modal-content action-modal" @click.stop>
        <div class="modal-header">
          <h3>Check Out Physical Copy</h3>
          <button @click="showCheckoutModal = false" class="close-btn">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>File:</label>
            <p>
              <strong>{{ fileToCheckout?.file_name }}</strong>
            </p>
            <p class="text-muted">{{ fileToCheckout?.category_name }}</p>
          </div>

          <div class="form-group">
            <label>Check Out To: *</label>
            <select v-model="checkoutType" class="form-select">
              <option value="user">User</option>
              <option value="client">Client</option>
              <option value="custom_clearance_agent">Transiteur</option>
            </select>
          </div>

          <!-- User Selection -->
          <div v-if="checkoutType === 'user'" class="form-group">
            <label>Select User: *</label>
            <el-select
              v-model="checkoutUserId"
              filterable
              remote
              :remote-method="searchUsers"
              :loading="loadingUsers"
              placeholder="Select or search user..."
              style="width: 100%"
              @change="handleUserChange"
            >
              <el-option
                v-for="user in filteredUsers"
                :key="user.id"
                :label="user.username"
                :value="user.id"
              >
                <i class="fas fa-user"></i>
                {{ user.username }}
                <small v-if="user.email"> ({{ user.email }})</small>
              </el-option>
            </el-select>
          </div>

          <!-- Agent Selection -->
          <div v-if="checkoutType === 'custom_clearance_agent'" class="form-group">
            <label>Sélectionner un Transiteur: *</label>
            <div class="form-row">
              <select v-model="checkoutAgentId" class="form-select" required>
                <option value="">Select agent...</option>
                <option v-for="agent in availableAgents" :key="agent.id" :value="agent.id">
                  {{ agent.name }}
                  <span v-if="agent.contact_person"> - {{ agent.contact_person }}</span>
                </option>
              </select>
              <button
                @click="showAgentsModal = true"
                class="btn-manage-agents"
                title="Manage agents"
              >
                <i class="fas fa-cog"></i>
              </button>
            </div>
          </div>

          <!-- Client Selection -->
          <template v-if="checkoutType === 'client'">
            <div class="form-group">
              <label>Sélectionner un Client: *</label>
              <el-select
                v-model="checkoutClientId"
                filterable
                remote
                :remote-method="searchClients"
                :loading="loadingClients"
                placeholder="Rechercher un client..."
                style="width: 100%"
                @change="handleClientChange"
              >
                <el-option
                  v-for="client in filteredClients"
                  :key="client.id"
                  :label="client.name"
                  :value="client.id"
                >
                  <i class="fas fa-user"></i>
                  {{ client.name }}
                  <small v-if="client.email || client.mobiles">
                    <span v-if="client.email">{{ client.email }}</span>
                    <span v-if="client.mobiles"> - {{ client.mobiles }}</span>
                  </small>
                </el-option>
              </el-select>
            </div>
          </template>
          <div class="form-group">
            <label>Notes (optional):</label>
            <textarea
              v-model="checkoutNotes"
              class="form-textarea"
              rows="3"
              placeholder="Add any notes about the checkout..."
            ></textarea>
          </div>
        </div>
        <div class="modal-footer">
          <button @click="showCheckoutModal = false" class="btn-cancel">Cancel</button>
          <button @click="confirmCheckout" :disabled="loading" class="btn-primary">
            Check Out
          </button>
        </div>
      </div>
    </div>

    <!-- Check In Modal -->
    <div v-if="showCheckinModal" class="modal-overlay" @click="showCheckinModal = false">
      <div class="modal-content action-modal" @click.stop>
        <div class="modal-header">
          <h3>Check In Physical Copy</h3>
          <button @click="showCheckinModal = false" class="close-btn">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>File:</label>
            <p>
              <strong>{{ fileToCheckin?.file_name }}</strong>
            </p>
            <p class="text-muted">{{ fileToCheckin?.category_name }}</p>
          </div>
          <div class="form-group">
            <label>Notes (optional):</label>
            <textarea
              v-model="checkinNotes"
              class="form-textarea"
              rows="3"
              placeholder="Add any notes about the check-in..."
            ></textarea>
          </div>
        </div>
        <div class="modal-footer">
          <button @click="showCheckinModal = false" class="btn-cancel">Cancel</button>
          <button @click="confirmCheckin" :disabled="loading" class="btn-primary">Check In</button>
        </div>
      </div>
    </div>

    <!-- Pending Transfers Modal -->
    <div
      v-if="showPendingTransfersModal"
      class="modal-overlay"
      @click="showPendingTransfersModal = false"
    >
      <div class="modal-content pending-transfers-modal" @click.stop>
        <div class="modal-header">
          <h3>
            <i class="fas fa-clock"></i>
            Pending Transfers
            <span v-if="pendingTransfers.length > 0" class="pending-count">
              ({{ pendingTransfers.length }})
            </span>
          </h3>
          <button @click="showPendingTransfersModal = false" class="close-btn">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <div class="modal-body">
          <div v-if="loading" class="loading-state">
            <i class="fas fa-spinner fa-spin"></i>
            <span>Loading pending transfers...</span>
          </div>

          <div v-else-if="pendingTransfers.length === 0" class="empty-state">
            <i class="fas fa-check-circle"></i>
            <p>No pending transfers</p>
          </div>

          <div v-else class="pending-transfers-list">
            <div
              v-for="transfer in pendingTransfers"
              :key="transfer.id"
              class="pending-transfer-item"
            >
              <div class="transfer-info">
                <div class="transfer-header">
                  <strong>{{ transfer.file_name }}</strong>
                  <span class="transfer-date">{{ formatDate(transfer.transferred_at) }}</span>
                </div>
                <div class="transfer-details">
                  <div class="transfer-from">
                    <i class="fas fa-user"></i>
                    From: {{ transfer.from_username || 'Unknown' }}
                  </div>
                  <div class="transfer-to">
                    <i class="fas fa-user"></i>
                    To: {{ transfer.to_username || 'You' }}
                  </div>
                  <div v-if="transfer.category_name" class="transfer-category">
                    <i class="fas fa-folder"></i>
                    Category: {{ transfer.category_name }}
                  </div>
                  <div v-if="transfer.vin" class="transfer-car">
                    <i class="fas fa-car"></i>
                    VIN: {{ transfer.vin }}
                  </div>
                  <div v-if="transfer.notes" class="transfer-notes">
                    <i class="fas fa-comment"></i>
                    {{ transfer.notes }}
                  </div>
                  <div v-if="transfer.return_expected_date" class="transfer-return-date">
                    <i class="fas fa-calendar"></i>
                    Expected return: {{ formatDate(transfer.return_expected_date) }}
                  </div>
                </div>
              </div>
              <div class="transfer-actions">
                <button
                  @click="handleApproveTransfer(transfer)"
                  class="btn-approve"
                  :disabled="loading"
                  title="Approve transfer"
                >
                  <i class="fas fa-check"></i>
                  Approve
                </button>
                <button
                  @click="handleRejectTransfer(transfer)"
                  class="btn-reject"
                  :disabled="loading"
                  title="Reject transfer"
                >
                  <i class="fas fa-times"></i>
                  Reject
                </button>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <div style="display: flex; gap: 0.5rem; flex: 1; justify-content: flex-start">
            <button
              v-if="pendingTransfers.length > 0"
              @click="handleApproveAll"
              class="btn-approve-all"
              :disabled="loading"
              title="Approve all pending transfers"
            >
              <i class="fas fa-check-double"></i>
              Approve All ({{ pendingTransfers.length }})
            </button>
            <button
              v-if="pendingTransfers.length > 0"
              @click="handleRejectAll"
              class="btn-reject-all"
              :disabled="loading"
              title="Reject all pending transfers"
            >
              <i class="fas fa-times-circle"></i>
              Reject All ({{ pendingTransfers.length }})
            </button>
          </div>
          <button @click="showPendingTransfersModal = false" class="btn-cancel">Close</button>
        </div>
      </div>
    </div>

    <!-- Transfer History Modal -->
    <div v-if="showHistoryModal" class="modal-overlay" @click="showHistoryModal = false">
      <div class="modal-content history-modal" @click.stop>
        <div class="modal-header">
          <h3>Transfer History</h3>
          <button @click="showHistoryModal = false" class="close-btn">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <div class="modal-body">
          <div class="file-info-header">
            <p>
              <strong>{{ fileForHistory?.file_name }}</strong>
            </p>
            <p class="text-muted">{{ fileForHistory?.category_name }}</p>
          </div>

          <div v-if="loading" class="loading-state">
            <i class="fas fa-spinner fa-spin"></i>
            <span>Loading history...</span>
          </div>

          <div v-else-if="transferHistory.length === 0" class="empty-state">
            <i class="fas fa-history"></i>
            <p>No transfer history found</p>
          </div>

          <div v-else class="history-list">
            <div
              v-for="(transfer, index) in transferHistory"
              :key="transfer.id"
              class="history-item"
            >
              <div class="history-header">
                <span class="history-number">#{{ transferHistory.length - index }}</span>
                <span class="history-date">{{ formatDate(transfer.transferred_at) }}</span>
              </div>
              <div class="history-content">
                <div class="history-transfer">
                  <i class="fas fa-arrow-right"></i>
                  <span v-if="transfer.from_username">
                    {{ transfer.from_username }}
                  </span>
                  <span v-else-if="transfer.from_agent_name">
                    Transiteur: {{ transfer.from_agent_name }}
                  </span>
                  <span v-else-if="transfer.from_client_name">
                    Client: {{ transfer.from_client_name }}
                  </span>
                  <span v-else class="text-muted">Available</span>
                  <i class="fas fa-arrow-right"></i>
                  <span v-if="transfer.to_username">
                    {{ transfer.to_username }}
                  </span>
                  <span v-else-if="transfer.to_agent_name">
                    Transiteur: {{ transfer.to_agent_name }}
                  </span>
                  <span v-else-if="transfer.to_client_name">
                    Client: {{ transfer.to_client_name }}
                  </span>
                  <span v-else class="text-muted">Unknown</span>
                </div>
                <div v-if="transfer.notes" class="history-notes">
                  <i class="fas fa-comment"></i>
                  {{ transfer.notes }}
                </div>
                <div v-if="transfer.return_expected_date" class="history-return">
                  <i class="fas fa-calendar"></i>
                  Expected return: {{ formatDate(transfer.return_expected_date) }}
                </div>
                <div v-if="transfer.returned_at" class="history-returned">
                  <i class="fas fa-check-circle"></i>
                  Returned: {{ formatDate(transfer.returned_at) }}
                  <span v-if="transfer.return_notes" class="return-notes">
                    ({{ transfer.return_notes }})
                  </span>
                </div>
                <div class="history-by">
                  <i class="fas fa-user"></i>
                  <span v-if="transfer.returned_at">Checked in by</span>
                  <span v-else>Transferred by</span>: {{ transfer.transferred_by_username }}
                </div>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button @click="showHistoryModal = false" class="btn-close">Close</button>
        </div>
      </div>
    </div>

    <!-- Physical Copy Tracking Modal -->
    <PhysicalCopyTracking :show="showTrackingModal" @close="showTrackingModal = false" />

    <!-- Custom Clearance Agents Modal -->
    <CustomClearanceAgents :show="showAgentsModal" @close="showAgentsModal = false" />

    <!-- Batch Transfer Modal -->
    <div
      v-if="showBatchTransferModal"
      class="modal-overlay"
      @click="showBatchTransferModal = false"
    >
      <div class="modal-content transfer-modal" @click.stop>
        <div class="modal-header">
          <h3>Batch Transfer All Documents</h3>
          <button @click="showBatchTransferModal = false" class="close-btn">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>Files to transfer:</label>
            <p class="text-muted">{{ myFiles.length }} document(s) will be transferred</p>
          </div>
          <div class="form-group">
            <label>Transfer to:</label>
            <select v-model="batchTransferToUserId" class="form-select">
              <option value="">Select user...</option>
              <option v-for="user in batchTransferAvailableUsers" :key="user.id" :value="user.id">
                {{ user.username }} ({{ user.email }})
              </option>
            </select>
          </div>
          <div class="form-group">
            <label>Expected Return Date:</label>
            <input v-model="batchTransferExpectedReturn" type="date" class="form-input" />
          </div>
          <div class="form-group">
            <label>Notes:</label>
            <textarea v-model="batchTransferNotes" class="form-textarea" rows="3"></textarea>
          </div>
        </div>
        <div class="modal-footer">
          <button @click="showBatchTransferModal = false" class="btn-cancel">Cancel</button>
          <button
            @click="confirmBatchTransfer"
            :disabled="!batchTransferToUserId || loading"
            class="btn-primary"
          >
            Transfer All
          </button>
        </div>
      </div>
    </div>

    <!-- Batch Checkout Modal -->
    <div
      v-if="showBatchCheckoutModal"
      class="modal-overlay"
      @click="showBatchCheckoutModal = false"
    >
      <div class="modal-content action-modal" @click.stop>
        <div class="modal-header">
          <h3>Batch Checkout All Documents</h3>
          <button @click="showBatchCheckoutModal = false" class="close-btn">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>Files to checkout:</label>
            <p class="text-muted">{{ myFiles.length }} document(s) will be checked out</p>
          </div>
          <div class="form-group">
            <label>Checkout Type:</label>
            <select v-model="batchCheckoutType" class="form-select">
              <option value="user">To User</option>
              <option value="client">To Client</option>
              <option value="custom_clearance_agent">To Transiteur</option>
            </select>
          </div>

          <!-- User Selection -->
          <div v-if="batchCheckoutType === 'user'" class="form-group">
            <label>Select User:</label>
            <select v-model="batchCheckoutUserId" class="form-select">
              <option value="">Select user...</option>
              <option v-for="user in batchCheckoutAvailableUsers" :key="user.id" :value="user.id">
                {{ user.username }} ({{ user.email }})
              </option>
            </select>
          </div>

          <!-- Client Selection -->
          <div v-if="batchCheckoutType === 'client'" class="form-group">
            <label>Select Client:</label>
            <el-select
              v-model="batchCheckoutClientId"
              filterable
              remote
              :remote-method="searchClients"
              placeholder="Search and select client..."
              class="w-100"
              style="width: 100%"
            >
              <el-option
                v-for="client in filteredClients"
                :key="client.id"
                :label="`${client.name}${client.email ? ' - ' + client.email : ''}`"
                :value="client.id"
              >
                <div>
                  <div>{{ client.name }}</div>
                  <small v-if="client.email" class="text-muted">{{ client.email }}</small>
                </div>
              </el-option>
            </el-select>
          </div>

          <!-- Transiteur Selection -->
          <div v-if="batchCheckoutType === 'custom_clearance_agent'" class="form-group">
            <label>Select Transiteur:</label>
            <select v-model="batchCheckoutAgentId" class="form-select">
              <option value="">Select transiteur...</option>
              <option v-for="agent in batchAvailableAgents" :key="agent.id" :value="agent.id">
                {{ agent.name }} ({{ agent.contact }})
              </option>
            </select>
          </div>

          <div class="form-group">
            <label>Notes:</label>
            <textarea v-model="batchCheckoutNotes" class="form-textarea" rows="3"></textarea>
          </div>
        </div>
        <div class="modal-footer">
          <button @click="showBatchCheckoutModal = false" class="btn-cancel">Cancel</button>
          <button
            @click="confirmBatchCheckout"
            :disabled="
              loading ||
              (batchCheckoutType === 'user' && !batchCheckoutUserId) ||
              (batchCheckoutType === 'client' && !batchCheckoutClientId) ||
              (batchCheckoutType === 'custom_clearance_agent' && !batchCheckoutAgentId)
            "
            class="btn-primary"
          >
            Checkout All
          </button>
        </div>
      </div>
    </div>

    <!-- Category Create/Edit Modal -->
    <div v-if="showCategoryModal" class="modal-overlay" @click="showCategoryModal = false">
      <div class="modal-content category-modal" @click.stop>
        <div class="modal-header">
          <h3>{{ editingCategory ? 'Edit Category' : 'Create New Category' }}</h3>
          <button @click="showCategoryModal = false" class="close-btn">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>Category Name *</label>
            <input
              v-model="categoryForm.category_name"
              type="text"
              class="form-input"
              placeholder="e.g., Insurance Document"
              required
            />
          </div>

          <div class="form-group">
            <label>Importance Level *</label>
            <select v-model.number="categoryForm.importance_level" class="form-select">
              <option :value="1">1 - Critical</option>
              <option :value="2">2 - High</option>
              <option :value="3">3 - Medium</option>
              <option :value="4">4 - Low</option>
              <option :value="5">5 - Optional</option>
            </select>
          </div>

          <div class="form-group">
            <label>
              <input v-model="categoryForm.is_required" type="checkbox" class="form-checkbox" />
              Required
            </label>
          </div>

          <div class="form-group">
            <label>Display Order</label>
            <input
              v-model.number="categoryForm.display_order"
              type="number"
              class="form-input"
              min="0"
            />
            <small class="help-text">Lower numbers appear first</small>
          </div>

          <div class="form-group">
            <label>Description</label>
            <textarea
              v-model="categoryForm.description"
              class="form-textarea"
              rows="3"
              placeholder="Optional description of this file category"
            ></textarea>
          </div>
        </div>
        <div class="modal-footer">
          <button @click="showCategoryModal = false" class="btn-cancel">Cancel</button>
          <button
            @click="saveCategory"
            :disabled="!categoryForm.category_name.trim() || loading"
            class="btn-primary"
          >
            {{ editingCategory ? 'Update' : 'Create' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Categories Table Modal -->
    <div v-if="showCategoriesTableModal" class="modal-overlay" @click="showCategoriesTableModal = false">
      <div class="modal-content categories-table-modal" @click.stop>
        <div class="modal-header">
          <h3>
            <i class="fas fa-list"></i>
            Manage Categories
          </h3>
          <div class="header-actions">
            <button
              @click="openCategoryModal()"
              class="btn-add-category"
              title="Add new category"
            >
              <i class="fas fa-plus"></i>
              Add Category
            </button>
            <button @click="showCategoriesTableModal = false" class="close-btn">
              <i class="fas fa-times"></i>
            </button>
          </div>
        </div>
        <div class="modal-body">
          <!-- Loading State -->
          <div v-if="loadingCategories" class="loading-state">
            <i class="fas fa-spinner fa-spin"></i>
            <span>Loading categories...</span>
          </div>

          <!-- Error Message -->
          <div v-if="error" class="error-message">
            <i class="fas fa-exclamation-circle"></i>
            {{ error }}
          </div>

          <!-- Success Message -->
          <div v-if="success" class="success-message">
            <i class="fas fa-check-circle"></i>
            {{ success }}
          </div>

          <!-- Categories Table -->
          <div v-if="!loadingCategories" class="categories-table-container">
            <table class="categories-table">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Importance</th>
                  <th>Required</th>
                  <th>Display Order</th>
                  <th>Description</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="categoriesForTable.length === 0">
                  <td colspan="6" class="text-center">No categories found</td>
                </tr>
                <tr v-for="category in categoriesForTable" :key="category.id">
                  <td>
                    <strong>{{ category.category_name }}</strong>
                  </td>
                  <td>
                    <span
                      class="importance-badge"
                      :style="{
                        backgroundColor: importanceConfig[category.importance_level]?.bgColor || '#f3f4f6',
                        color: importanceConfig[category.importance_level]?.color || '#374151',
                      }"
                    >
                      {{ importanceConfig[category.importance_level]?.label || category.importance_level }}
                    </span>
                  </td>
                  <td>
                    <span v-if="category.is_required === 1 || category.is_required === true" class="badge-required">
                      <i class="fas fa-check-circle"></i> Required
                    </span>
                    <span v-else class="badge-optional">Optional</span>
                  </td>
                  <td>{{ category.display_order }}</td>
                  <td>
                    <span v-if="category.description" class="description-text">{{ category.description }}</span>
                    <span v-else class="text-muted">-</span>
                  </td>
                  <td>
                    <div class="action-buttons">
                      <button
                        @click="handleEditCategoryFromTable(category)"
                        class="btn-edit"
                        title="Edit category"
                      >
                        <i class="fas fa-edit"></i>
                      </button>
                      <button
                        @click="handleDeleteCategory(category)"
                        class="btn-delete"
                        title="Delete category"
                        :disabled="loading"
                      >
                        <i class="fas fa-trash"></i>
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  border-radius: 8px;
  width: 90%;
  max-width: 900px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid #e5e7eb;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.25rem;
  color: #1f2937;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.btn-agents {
  padding: 0.5rem 1rem;
  background-color: #f59e0b;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  transition: background-color 0.2s;
}

.btn-agents:hover {
  background-color: #d97706;
}

.btn-tracking {
  padding: 0.5rem 1rem;
  background-color: #8b5cf6;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  transition: background-color 0.2s;
}

.btn-tracking:hover {
  background-color: #7c3aed;
}

.btn-add-category {
  padding: 0.5rem 1rem;
  background-color: #10b981;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  transition: background-color 0.2s;
}

.btn-add-category:hover {
  background-color: #059669;
}

.btn-categories {
  padding: 0.5rem 1rem;
  background-color: #6366f1;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  transition: background-color 0.2s;
}

.btn-categories:hover {
  background-color: #4f46e5;
}

.btn-batch-transfer,
.btn-batch-checkout {
  padding: 0.5rem 1rem;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  transition: background-color 0.2s;
}

.btn-batch-transfer:hover:not(:disabled),
.btn-batch-checkout:hover:not(:disabled) {
  background-color: #2563eb;
}

.btn-batch-transfer:disabled,
.btn-batch-checkout:disabled,
.btn-batch-transfer.disabled,
.btn-batch-checkout.disabled {
  opacity: 0.5;
  cursor: not-allowed;
  background-color: #9ca3af;
}

.batch-count {
  font-size: 0.75rem;
  opacity: 0.9;
  font-weight: 600;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: #6b7280;
  padding: 0.5rem;
}

.close-btn:hover {
  color: #374151;
}

.modal-body {
  flex: 1;
  overflow-y: auto;
  padding: 1.5rem;
}

.loading-state {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  justify-content: center;
  padding: 2rem;
  color: #6b7280;
}

.error-message,
.success-message {
  padding: 1rem;
  border-radius: 4px;
  margin-bottom: 1rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.error-message {
  background-color: #fef2f2;
  color: #dc2626;
}

.success-message {
  background-color: #f0fdf4;
  color: #10b981;
}

.category-section {
  margin-bottom: 2rem;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  overflow: hidden;
}

.category-header {
  padding: 1rem;
  background-color: #f9fafb;
  border-bottom: 1px solid #e5e7eb;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.category-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.importance-badge {
  padding: 0.25rem 0.75rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
}

.category-header h4 {
  margin: 0;
  font-size: 1.1rem;
  color: #1f2937;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.required-badge {
  font-size: 0.75rem;
  color: #dc2626;
  font-weight: normal;
}

.category-description {
  margin: 0.5rem 0 0 0;
  font-size: 0.875rem;
  color: #6b7280;
}

.files-list {
  padding: 1rem;
}

.file-item {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 1rem;
  border: 1px solid #e5e7eb;
  border-radius: 4px;
  margin-bottom: 0.5rem;
  background-color: #ffffff;
  flex-wrap: wrap;
  gap: 1rem;
}

.file-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex: 1;
}

.file-info i {
  font-size: 1.5rem;
  color: #3b82f6;
}

.file-name {
  font-weight: 500;
  color: #3b82f6;
  text-decoration: none;
}

.file-name:hover {
  text-decoration: underline;
}

.file-meta {
  font-size: 0.875rem;
  color: #6b7280;
  margin-top: 0.25rem;
}

.file-actions {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  flex-wrap: wrap;
  flex: 1;
  justify-content: flex-end;
}

.physical-status {
  margin-right: 0.5rem;
}

.status-badge {
  padding: 0.25rem 0.75rem;
  border-radius: 4px;
  font-size: 0.875rem;
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
}

.status-badge.available {
  background-color: #f0fdf4;
  color: #10b981;
}

.status-badge.checked-out {
  background-color: #fef2f2;
  color: #dc2626;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.holder-info {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.checked-out-message {
  padding: 0.75rem;
  background-color: #f3f4f6;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  color: #6b7280;
  font-size: 0.875rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  margin-top: 0.5rem;
}

.btn-rollback {
  padding: 0.5rem 1rem;
  background-color: #dc2626;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  transition: background-color 0.2s;
  margin-left: auto;
}

.btn-rollback:hover {
  background-color: #b91c1c;
}

.btn-rollback:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
}

.action-buttons {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
  align-items: center;
}

.btn-action {
  background: none;
  border: 1px solid #e5e7eb;
  border-radius: 4px;
  padding: 0.5rem 0.75rem;
  cursor: pointer;
  color: #6b7280;
  transition: all 0.2s;
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
  white-space: nowrap;
}

.btn-action:hover {
  background-color: #f9fafb;
  border-color: #d1d5db;
  color: #374151;
}

.btn-action.btn-delete {
  color: #dc2626;
}

.btn-action.btn-delete:hover {
  background-color: #fef2f2;
  border-color: #dc2626;
}

.btn-action.btn-history {
  color: #8b5cf6;
}

.btn-action.btn-history:hover {
  background-color: #f3e8ff;
  border-color: #8b5cf6;
}

.upload-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 1rem;
  background-color: #f9fafb;
  border-radius: 4px;
  margin-top: 0.5rem;
  transition: all 0.2s ease;
}

.upload-section.drag-over {
  background-color: #dbeafe;
  border: 2px dashed #3b82f6;
}

.drop-zone {
  position: relative;
  min-height: 120px;
  border: 2px dashed #d1d5db;
  border-radius: 8px;
  background-color: white;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.drop-zone:hover {
  border-color: #3b82f6;
  background-color: #f0f9ff;
}

.drop-zone.drag-active {
  border-color: #3b82f6;
  background-color: #dbeafe;
  border-style: solid;
}

.drop-zone-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 1.5rem;
  width: 100%;
  text-align: center;
}

.drop-zone-content.file-selected {
  flex-direction: row;
  justify-content: space-between;
  padding: 1rem;
}

.drop-zone-content i {
  font-size: 2.5rem;
  color: #9ca3af;
  margin-bottom: 0.5rem;
}

.drop-zone.drag-active .drop-zone-content i,
.drop-zone:hover .drop-zone-content i {
  color: #3b82f6;
}

.drop-zone-text {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  flex: 1;
}

.drop-zone-main-text {
  font-weight: 500;
  color: #1f2937;
  font-size: 0.875rem;
}

.drop-zone-sub-text {
  font-size: 0.75rem;
  color: #6b7280;
}

.file-input-hidden {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  opacity: 0;
  cursor: pointer;
  z-index: 1;
}

.btn-remove-file {
  padding: 0.375rem;
  background-color: #ef4444;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.2s;
  z-index: 2;
  position: relative;
}

.btn-remove-file:hover {
  background-color: #dc2626;
}

.btn-upload {
  padding: 0.5rem 1rem;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 500;
}

.btn-upload:hover:not(:disabled) {
  background-color: #2563eb;
}

.btn-upload:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.upload-section.has-existing-file {
  background-color: #fef3c7;
  border: 1px solid #fbbf24;
}

.upload-info-message {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #92400e;
  font-size: 0.875rem;
  padding: 0.5rem;
  margin-top: 0.5rem;
}

.upload-info-message i {
  color: #f59e0b;
  font-size: 1rem;
}

.btn-upload.btn-replace {
  background-color: #f59e0b;
}

.btn-upload.btn-replace:hover:not(:disabled) {
  background-color: #d97706;
}

.upload-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 1rem;
  background-color: #f9fafb;
  border-radius: 4px;
  margin-top: 0.5rem;
  transition: all 0.2s ease;
}

.upload-section.drag-over {
  background-color: #dbeafe;
  border: 2px dashed #3b82f6;
}

.drop-zone {
  position: relative;
  min-height: 120px;
  border: 2px dashed #d1d5db;
  border-radius: 8px;
  background-color: white;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.drop-zone:hover {
  border-color: #3b82f6;
  background-color: #f0f9ff;
}

.drop-zone.drag-active {
  border-color: #3b82f6;
  background-color: #dbeafe;
  border-style: solid;
}

.drop-zone-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 1.5rem;
  width: 100%;
  text-align: center;
}

.drop-zone-content.file-selected {
  flex-direction: row;
  justify-content: space-between;
  padding: 1rem;
}

.drop-zone-content i {
  font-size: 2.5rem;
  color: #9ca3af;
  margin-bottom: 0.5rem;
}

.drop-zone.drag-active .drop-zone-content i,
.drop-zone:hover .drop-zone-content i {
  color: #3b82f6;
}

.drop-zone-text {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  flex: 1;
}

.drop-zone-main-text {
  font-weight: 500;
  color: #1f2937;
  font-size: 0.875rem;
}

.drop-zone-sub-text {
  font-size: 0.75rem;
  color: #6b7280;
}

.file-input-hidden {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  opacity: 0;
  cursor: pointer;
  z-index: 1;
}

.btn-remove-file {
  padding: 0.375rem;
  background-color: #ef4444;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.2s;
  z-index: 2;
  position: relative;
}

.btn-remove-file:hover {
  background-color: #dc2626;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  padding: 1.5rem;
  border-top: 1px solid #e5e7eb;
  gap: 0.5rem;
}

.btn-close,
.btn-cancel,
.btn-primary {
  padding: 0.75rem 1.5rem;
  border-radius: 4px;
  font-weight: 500;
  cursor: pointer;
  border: none;
}

.btn-close,
.btn-cancel {
  background-color: #f3f4f6;
  color: #374151;
}

.btn-close:hover,
.btn-cancel:hover {
  background-color: #e5e7eb;
}

.btn-primary {
  background-color: #3b82f6;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background-color: #2563eb;
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.transfer-modal {
  max-width: 500px;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
  color: #374151;
}

.form-select,
.form-input,
.form-textarea {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 1rem;
}

.form-textarea {
  resize: vertical;
}

.category-modal {
  max-width: 600px;
}

.form-checkbox {
  margin-right: 0.5rem;
}

.help-text {
  display: block;
  margin-top: 0.25rem;
  font-size: 0.875rem;
  color: #6b7280;
}

.action-modal,
.history-modal {
  max-width: 600px;
}

.file-info-header {
  margin-bottom: 1.5rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid #e5e7eb;
}

.history-list {
  max-height: 400px;
  overflow-y: auto;
}

.history-item {
  padding: 1rem;
  border: 1px solid #e5e7eb;
  border-radius: 4px;
  margin-bottom: 0.75rem;
  background-color: #ffffff;
}

.history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.history-number {
  font-weight: 600;
  color: #3b82f6;
}

.history-date {
  color: #6b7280;
  font-size: 0.875rem;
}

.history-content {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.history-transfer {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 500;
}

.history-notes,
.history-return,
.history-returned,
.history-by {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #6b7280;
  font-size: 0.875rem;
}

.history-returned {
  color: #10b981;
  font-weight: 500;
}

.returned-badge {
  margin-left: 0.5rem;
  padding: 0.25rem 0.5rem;
  background-color: #10b981;
  color: white;
  border-radius: 4px;
  font-size: 0.75rem;
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
}

.return-notes {
  font-style: italic;
  color: #6b7280;
}

.empty-state {
  text-align: center;
  padding: 3rem 1rem;
  color: #6b7280;
}

.empty-state i {
  font-size: 3rem;
  color: #d1d5db;
  margin-bottom: 1rem;
}

.empty-state p {
  margin: 0;
  font-size: 1rem;
}

.form-row {
  display: flex;
  gap: 0.5rem;
  align-items: flex-start;
}

.btn-manage-agents {
  padding: 0.5rem;
  background-color: #f3f4f6;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  cursor: pointer;
  color: #6b7280;
  transition: all 0.2s;
  white-space: nowrap;
}

.btn-manage-agents:hover {
  background-color: #e5e7eb;
  border-color: #9ca3af;
  color: #374151;
}

/* Element Plus select styling */
:deep(.el-select) {
  width: 100%;
}

:deep(.el-select-dropdown__item) {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

:deep(.el-select-dropdown__item i) {
  color: #3b82f6;
}

:deep(.el-select-dropdown__item small) {
  color: #6b7280;
  margin-left: 0.5rem;
}

.btn-pending-transfers {
  padding: 0.5rem 1rem;
  background-color: #f59e0b;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  transition: background-color 0.2s;
  position: relative;
}

.btn-pending-transfers:hover {
  background-color: #d97706;
}

.btn-pending-transfers.has-pending {
  background-color: #dc2626;
  animation: pulse 2s infinite;
}

.btn-pending-transfers.has-pending:hover {
  background-color: #b91c1c;
}

.pending-badge {
  background-color: white;
  color: #dc2626;
  border-radius: 50%;
  width: 1.5rem;
  height: 1.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.75rem;
  font-weight: bold;
  margin-left: 0.25rem;
}

@keyframes pulse {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.8;
  }
}

.pending-transfers-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.pending-transfer-item {
  padding: 1rem;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  background-color: #f9fafb;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
}

.transfer-info {
  flex: 1;
}

.transfer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.transfer-header strong {
  font-size: 1rem;
  color: #1f2937;
}

.transfer-date {
  font-size: 0.875rem;
  color: #6b7280;
}

.transfer-details {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.transfer-details > div {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
  color: #4b5563;
}

.transfer-details i {
  color: #6b7280;
  width: 1rem;
}

.transfer-actions {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.btn-approve {
  padding: 0.5rem 1rem;
  background-color: #10b981;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  transition: background-color 0.2s;
}

.btn-approve:hover {
  background-color: #059669;
}

.btn-approve:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
}

.btn-reject {
  padding: 0.5rem 1rem;
  background-color: #ef4444;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  transition: background-color 0.2s;
}

.btn-reject:hover {
  background-color: #dc2626;
}

.btn-reject:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
}

.btn-approve-all,
.btn-reject-all {
  padding: 0.75rem 1.5rem;
  border-radius: 4px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  transition: background-color 0.2s;
}

.btn-approve-all {
  background-color: #10b981;
  color: white;
}

.btn-approve-all:hover:not(:disabled) {
  background-color: #059669;
}

.btn-approve-all:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-reject-all {
  background-color: #dc2626;
  color: white;
}

.btn-reject-all:hover:not(:disabled) {
  background-color: #b91c1c;
}

.btn-reject-all:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.pending-count {
  font-size: 0.875rem;
  color: #6b7280;
  font-weight: normal;
}

.status-badge.pending-transfer {
  background-color: #fef3c7;
  color: #d97706;
  border: 1px solid #fbbf24;
}

.pending-transfer-message {
  padding: 0.75rem;
  background-color: #fef3c7;
  border: 1px solid #fbbf24;
  border-radius: 4px;
  color: #d97706;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
}

.pending-transfer-message i {
  color: #d97706;
}

.btn-view-pending {
  margin-left: auto;
  padding: 0.25rem 0.75rem;
  background-color: #d97706;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.75rem;
  transition: background-color 0.2s;
}

.btn-view-pending:hover {
  background-color: #b45309;
}

.categories-table-modal {
  max-width: 1200px;
}

.categories-table-container {
  overflow-x: auto;
}

.categories-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 1rem;
}

.categories-table thead {
  background-color: #f9fafb;
  border-bottom: 2px solid #e5e7eb;
}

.categories-table th {
  padding: 0.75rem 1rem;
  text-align: left;
  font-weight: 600;
  color: #374151;
  font-size: 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.categories-table td {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid #e5e7eb;
  color: #1f2937;
}

.categories-table tbody tr:hover {
  background-color: #f9fafb;
}

.categories-table tbody tr:last-child td {
  border-bottom: none;
}

.importance-badge {
  display: inline-block;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
}

.badge-required {
  color: #dc2626;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.badge-optional {
  color: #6b7280;
}

.description-text {
  color: #6b7280;
  font-size: 0.875rem;
  max-width: 300px;
  display: inline-block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.action-buttons {
  display: flex;
  gap: 0.5rem;
}

.btn-edit {
  padding: 0.375rem 0.75rem;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.875rem;
  transition: background-color 0.2s;
}

.btn-edit:hover {
  background-color: #2563eb;
}

.btn-delete {
  padding: 0.375rem 0.75rem;
  background-color: #dc2626;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.875rem;
  transition: background-color 0.2s;
}

.btn-delete:hover:not(:disabled) {
  background-color: #b91c1c;
}

.btn-delete:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.text-center {
  text-align: center;
}

.text-muted {
  color: #9ca3af;
}
</style>
