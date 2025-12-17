<script setup>
import { ref, computed, onMounted, watch } from 'vue'
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
  getMyPhysicalCopies,
  getUsersForTransfer,
  getFileUrl,
  createFileCategory,
  updateFileCategory,
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

// Check if current user is admin
const currentUser = computed(() => {
  const userStr = localStorage.getItem('user')
  return userStr ? JSON.parse(userStr) : null
})

const isAdmin = computed(() => {
  return currentUser.value?.role_id === 1
})

// Group files by category
const filesByCategory = computed(() => {
  const grouped = {}
  if (!categories.value || categories.value.length === 0) {
    console.log('[CarFilesManagement] No categories available')
    return grouped
  }
  categories.value.forEach((cat) => {
    grouped[cat.id] = {
      category: cat,
      files: (files.value || []).filter((f) => f.category_id === cat.id),
    }
  })
  console.log(
    '[CarFilesManagement] filesByCategory computed:',
    grouped,
    'Categories:',
    categories.value.length,
    'Files:',
    files.value.length,
  )
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
    console.error('Load error:', err)
  } finally {
    loading.value = false
  }
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

// Upload file
const uploadFile = async (categoryId) => {
  const file = selectedFiles.value[categoryId]
  if (!file) return

  isProcessing.value = true
  error.value = null

  try {
    const result = await uploadCarFile(file, props.car.id, categoryId)
    success.value = 'File uploaded successfully'

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

// Delete file
const handleDeleteFile = async (fileId) => {
  if (!confirm('Are you sure you want to delete this file?')) return

  loading.value = true
  error.value = null

  try {
    await deleteCarFile(fileId)
    success.value = 'File deleted successfully'
    await loadData()
    emit('save', props.car)
  } catch (err) {
    error.value = err.message || 'Failed to delete file'
  } finally {
    loading.value = false
  }
}

// Check out physical copy
const handleCheckout = async (fileId) => {
  const expectedReturn = prompt('Expected return date (YYYY-MM-DD) - optional:')
  const notes = prompt('Notes (optional):')

  loading.value = true
  error.value = null

  try {
    await checkoutPhysicalCopy(fileId, expectedReturn || null, notes || null)
    success.value = 'File checked out successfully'
    await loadData()
  } catch (err) {
    error.value = err.message || 'Failed to check out file'
  } finally {
    loading.value = false
  }
}

// Check in physical copy
const handleCheckin = async (fileId) => {
  if (!confirm('Check in this physical copy?')) return

  loading.value = true
  error.value = null

  try {
    await checkinPhysicalCopy(fileId)
    success.value = 'File checked in successfully'
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
const availableUsers = ref([])

const openTransferModal = async (file) => {
  fileToTransfer.value = file
  transferToUserId.value = null
  transferNotes.value = ''
  transferExpectedReturn.value = ''

  try {
    const users = await getUsersForTransfer(file.id)
    availableUsers.value = users
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
    success.value = 'File transferred successfully'
    showTransferModal.value = false
    await loadData()
  } catch (err) {
    error.value = err.message || 'Failed to transfer file'
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
  visibility_scope: 'public',
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
      visibility_scope: category.visibility_scope || 'public',
    }
  } else {
    editingCategory.value = null
    categoryForm.value = {
      category_name: '',
      importance_level: 3,
      is_required: false,
      display_order: categories.value.length,
      description: '',
      visibility_scope: 'public',
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
  } catch (err) {
    error.value = err.message || 'Failed to save category'
  } finally {
    loading.value = false
  }
}

const closeModal = () => {
  error.value = null
  success.value = null
  selectedFiles.value = {}
  showTransferModal.value = false
  showCategoryModal.value = false
  emit('close')
}

// Watch for show prop changes
watch(
  () => props.show,
  (newVal) => {
    if (newVal) {
      loadData()
    } else {
      closeModal()
    }
  },
  { immediate: true },
)

// Load data on mount if shown
onMounted(() => {
  if (props.show) {
    loadData()
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
            v-if="isAdmin"
            @click="openCategoryModal()"
            class="btn-add-category"
            title="Add new category"
          >
            <i class="fas fa-plus"></i>
            Add Category
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

        <!-- Debug Info (remove in production) -->
        <div
          v-if="!loading"
          style="padding: 1rem; background: #f0f0f0; margin-bottom: 1rem; font-size: 0.875rem"
        >
          <strong>Debug:</strong> Categories: {{ categories.length }}, Files: {{ files.length }},
          Groups: {{ Object.keys(filesByCategory).length }}
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
                <div v-if="file.physical_status" class="physical-status">
                  <span v-if="file.physical_status === 'available'" class="status-badge available">
                    <i class="fas fa-check-circle"></i> Available
                  </span>
                  <span
                    v-else-if="file.physical_status === 'checked_out'"
                    class="status-badge checked-out"
                  >
                    <i class="fas fa-user"></i>
                    {{ file.current_holder_username || 'Unknown' }}
                  </span>
                </div>

                <!-- Actions -->
                <div class="action-buttons">
                  <button
                    v-if="file.physical_status === 'available'"
                    @click="handleCheckout(file.id)"
                    class="btn-action btn-checkout"
                    title="Check out physical copy"
                  >
                    <i class="fas fa-sign-out-alt"></i>
                  </button>
                  <button
                    v-if="
                      file.physical_status === 'checked_out' &&
                      file.current_holder_id === currentUser?.id
                    "
                    @click="handleCheckin(file.id)"
                    class="btn-action btn-checkin"
                    title="Check in physical copy"
                  >
                    <i class="fas fa-sign-in-alt"></i>
                  </button>
                  <button
                    v-if="
                      file.physical_status === 'checked_out' &&
                      file.current_holder_id === currentUser?.id
                    "
                    @click="openTransferModal(file)"
                    class="btn-action btn-transfer"
                    title="Transfer to another user"
                  >
                    <i class="fas fa-exchange-alt"></i>
                  </button>
                  <button
                    v-if="file.uploaded_by === currentUser?.id || isAdmin"
                    @click="handleDeleteFile(file.id)"
                    class="btn-action btn-delete"
                    title="Delete file"
                  >
                    <i class="fas fa-trash"></i>
                  </button>
                </div>
              </div>
            </div>

            <!-- Upload Section -->
            <div class="upload-section">
              <input
                :id="`file-input-${categoryId}`"
                type="file"
                class="file-input"
                @change="(e) => handleFileSelect(e, categoryId)"
                :disabled="isProcessing"
                accept=".pdf,.jpg,.jpeg,.png,.gif,.webp,.doc,.docx,.xls,.xlsx"
              />
              <button
                @click="uploadFile(categoryId)"
                :disabled="!selectedFiles[categoryId] || isProcessing"
                class="btn-upload"
              >
                <i class="fas fa-upload"></i>
                {{
                  selectedFiles[categoryId]
                    ? `Upload ${selectedFiles[categoryId].name}`
                    : 'Select File'
                }}
              </button>
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
              <option v-for="user in availableUsers" :key="user.id" :value="user.id">
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

          <div class="form-group">
            <label>Visibility Scope</label>
            <select v-model="categoryForm.visibility_scope" class="form-select">
              <option value="public">Public - All users can see</option>
              <option value="department">Department - Same department only</option>
              <option value="role">Role - Same role level only</option>
              <option value="private">Private - Uploader only</option>
            </select>
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
  align-items: center;
  padding: 1rem;
  border: 1px solid #e5e7eb;
  border-radius: 4px;
  margin-bottom: 0.5rem;
  background-color: #ffffff;
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
  align-items: center;
  gap: 1rem;
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
}

.action-buttons {
  display: flex;
  gap: 0.5rem;
}

.btn-action {
  background: none;
  border: 1px solid #e5e7eb;
  border-radius: 4px;
  padding: 0.5rem;
  cursor: pointer;
  color: #6b7280;
  transition: all 0.2s;
}

.btn-action:hover {
  background-color: #f9fafb;
  border-color: #d1d5db;
}

.btn-action.btn-delete {
  color: #dc2626;
}

.btn-action.btn-delete:hover {
  background-color: #fef2f2;
  border-color: #dc2626;
}

.upload-section {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  padding: 1rem;
  background-color: #f9fafb;
  border-radius: 4px;
  margin-top: 0.5rem;
}

.file-input {
  flex: 1;
  padding: 0.5rem;
  border: 1px solid #d1d5db;
  border-radius: 4px;
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
</style>
