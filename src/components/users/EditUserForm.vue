<script setup>
import { ref, onMounted, watch } from 'vue'
import { useApi } from '../../composables/useApi'

const props = defineProps({
  user: {
    type: Object,
    required: true,
  },
})

const emit = defineEmits(['user-updated', 'close'])

const { callApi, uploadFile, getFileUrl } = useApi()
const loading = ref(false)
const error = ref(null)
const roles = ref([])
const formData = ref({
  username: '',
  email: '',
  password: '',
  role_id: '',
  max_unpayed_created_bills: 0,
  is_diffrent_company: false,
  path_logo: '',
  path_letter_head: '',
  path_stamp: '',
})

// File upload states
const logoFile = ref(null)
const letterHeadFile = ref(null)
const stampFile = ref(null)
const uploadingLogo = ref(false)
const uploadingLetterHead = ref(false)
const uploadingStamp = ref(false)

// Preview URLs for existing images
const logoPreview = ref(null)
const letterHeadPreview = ref(null)
const stampPreview = ref(null)

const fetchRoles = async () => {
  const result = await callApi({
    query: 'SELECT * FROM roles ORDER BY role_name',
  })
  if (result.success) {
    roles.value = result.data
  }
}

// Initialize form data from user prop
const initializeForm = () => {
  if (props.user) {
    // Handle is_diffrent_company - it might be 0, 1, null, or boolean
    const isDifferentCompany = props.user.is_diffrent_company === 1 || 
                               props.user.is_diffrent_company === true || 
                               props.user.is_diffrent_company === '1'
    
    console.log('Initializing form for user:', props.user.id, 'is_diffrent_company from DB:', props.user.is_diffrent_company, 'type:', typeof props.user.is_diffrent_company, 'converted to:', isDifferentCompany)
    
    formData.value = {
      username: props.user.username || '',
      email: props.user.email || '',
      password: '', // Don't populate password for security
      role_id: props.user.role_id || '',
      max_unpayed_created_bills: props.user.max_unpayed_created_bills || 0,
      is_diffrent_company: isDifferentCompany,
      path_logo: props.user.path_logo || '',
      path_letter_head: props.user.path_letter_head || '',
      path_stamp: props.user.path_stamp || '',
    }

    // Set preview URLs for existing images
    if (formData.value.path_logo) {
      logoPreview.value = getFileUrl(formData.value.path_logo)
    }
    if (formData.value.path_letter_head) {
      letterHeadPreview.value = getFileUrl(formData.value.path_letter_head)
    }
    if (formData.value.path_stamp) {
      stampPreview.value = getFileUrl(formData.value.path_stamp)
    }
  }
}

const handleFileSelect = (event, type) => {
  const file = event.target.files[0]
  if (!file) return

  // Validate image file
  if (!file.type.startsWith('image/')) {
    error.value = `${type} must be an image file`
    return
  }

  // Limit file size to 5MB
  if (file.size > 5 * 1024 * 1024) {
    error.value = `${type} file size must be less than 5MB`
    return
  }

  // Create preview URL
  const previewUrl = URL.createObjectURL(file)

  if (type === 'logo') {
    logoFile.value = file
    logoPreview.value = previewUrl
  } else if (type === 'letter_head') {
    letterHeadFile.value = file
    letterHeadPreview.value = previewUrl
  } else if (type === 'stamp') {
    stampFile.value = file
    stampPreview.value = previewUrl
  }
}

const uploadImageFile = async (file, type) => {
  if (!file) return null

  try {
    const timestamp = Date.now()
    const fileExtension = file.name.split('.').pop()
    const filename = `${type}_${timestamp}.${fileExtension}`
    // Upload to files_dir directly (files_dir is used as base_directory in uploadFile)
    const result = await uploadFile(file, type, filename)

    if (result.success) {
      return result.relativePath
    } else {
      throw new Error(result.error || `Failed to upload ${type}`)
    }
  } catch (err) {
    console.error(`Error uploading ${type}:`, err)
    throw err
  }
}

const handleSubmit = async () => {
  if (!formData.value.username || !formData.value.email) {
    error.value = 'Username and email are required'
    return
  }

  // Validate that logo and letterhead are required for different company users
  // For editing: either new file must be uploaded OR existing path must exist
  if (formData.value.is_diffrent_company) {
    const hasLogoFile = logoFile.value !== null
    const hasExistingLogo = formData.value.path_logo && formData.value.path_logo.trim() !== ''
    
    if (!hasLogoFile && !hasExistingLogo) {
      error.value = 'Logo is required for different company users. Please upload a logo or keep the existing one.'
      return
    }

    const hasLetterHeadFile = letterHeadFile.value !== null
    const hasExistingLetterHead = formData.value.path_letter_head && formData.value.path_letter_head.trim() !== ''
    
    if (!hasLetterHeadFile && !hasExistingLetterHead) {
      error.value = 'Letterhead is required for different company users. Please upload a letterhead or keep the existing one.'
      return
    }
  }

  loading.value = true
  error.value = null

  try {
    // Upload new image files if is_diffrent_company is true and new files are selected
    if (formData.value.is_diffrent_company) {
      if (logoFile.value) {
        uploadingLogo.value = true
        formData.value.path_logo = await uploadImageFile(logoFile.value, 'logo')
        uploadingLogo.value = false
      }
      // If no new logo file but user is different company, keep existing path_logo
      if (letterHeadFile.value) {
        uploadingLetterHead.value = true
        formData.value.path_letter_head = await uploadImageFile(letterHeadFile.value, 'letter_head')
        uploadingLetterHead.value = false
      }
      if (stampFile.value) {
        uploadingStamp.value = true
        formData.value.path_stamp = await uploadImageFile(stampFile.value, 'stamp')
        uploadingStamp.value = false
      }
    } else {
      // If user is no longer a different company, clear the paths
      formData.value.path_logo = ''
      formData.value.path_letter_head = ''
      formData.value.path_stamp = ''
    }

    // Update user info
    const isDifferentCompanyValue = formData.value.is_diffrent_company ? 1 : 0
    
    console.log('Updating user:', props.user.id)
    console.log('is_diffrent_company value to save:', isDifferentCompanyValue, 'formData.is_diffrent_company:', formData.value.is_diffrent_company)
    console.log('All form data:', formData.value)
    
    const result = await callApi({
      query:
        'UPDATE users SET username = ?, email = ?, role_id = ?, max_unpayed_created_bills = ?, is_diffrent_company = ?, path_logo = ?, path_letter_head = ?, path_stamp = ? WHERE id = ?',
      params: [
        formData.value.username,
        formData.value.email,
        formData.value.role_id || null,
        formData.value.max_unpayed_created_bills || 0,
        isDifferentCompanyValue,
        formData.value.path_logo || null,
        formData.value.path_letter_head || null,
        formData.value.path_stamp || null,
        props.user.id,
      ],
    })
    
    console.log('Update result:', result)
    
    if (!result.success) {
      console.error('Update failed:', result.error)
    }

    // If there's a password to update, do it in a separate call
    if (formData.value.password && formData.value.password.trim()) {
      const passwordResult = await callApi({
        query: 'UPDATE users SET password = ? WHERE id = ?',
        params: [formData.value.password, props.user.id],
        action: 'hash_password',
      })
      if (!passwordResult.success) {
        error.value = passwordResult.error || 'Failed to update password'
        return
      }
    }

    if (result.success) {
      emit('user-updated')
      emit('close')
    } else {
      error.value = result.error || 'Failed to update user'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
    uploadingLogo.value = false
    uploadingLetterHead.value = false
    uploadingStamp.value = false
  }
}

const handleCancel = () => {
  formData.value = {
    username: '',
    email: '',
    password: '',
    role_id: '',
    max_unpayed_created_bills: 0,
    is_diffrent_company: false,
    path_logo: '',
    path_letter_head: '',
    path_stamp: '',
  }
  logoFile.value = null
  letterHeadFile.value = null
  stampFile.value = null
  logoPreview.value = null
  letterHeadPreview.value = null
  stampPreview.value = null
  error.value = null
  emit('close')
}

// Watch for user prop changes
watch(() => props.user, () => {
  initializeForm()
}, { immediate: true })

onMounted(() => {
  fetchRoles()
  initializeForm()
})
</script>

<template>
  <teleport to="body">
    <div class="modal-overlay" @click="handleCancel">
      <div class="modal-content" @click.stop>
        <div class="form-header">
          <h3>
            <i class="fas fa-user-edit"></i>
            Edit User
          </h3>
          <button @click="handleCancel" class="close-btn" :disabled="loading">
            <i class="fas fa-times"></i>
          </button>
        </div>

        <form @submit.prevent="handleSubmit" class="edit-user-form">
          <div class="form-group">
            <label for="username">
              <i class="fas fa-user"></i>
              Username <span class="required">*</span>
            </label>
            <input
              id="username"
              v-model="formData.username"
              type="text"
              class="form-input"
              placeholder="Enter username"
              required
              :disabled="loading"
            />
          </div>

          <div class="form-group">
            <label for="email">
              <i class="fas fa-envelope"></i>
              Email <span class="required">*</span>
            </label>
            <input
              id="email"
              v-model="formData.email"
              type="email"
              class="form-input"
              placeholder="Enter email"
              required
              :disabled="loading"
            />
          </div>

          <div class="form-group">
            <label for="password">
              <i class="fas fa-lock"></i>
              New Password
            </label>
            <input
              id="password"
              v-model="formData.password"
              type="password"
              class="form-input"
              placeholder="Leave empty to keep current password"
              :disabled="loading"
            />
          </div>

          <div class="form-group">
            <label for="role_id">
              <i class="fas fa-shield-alt"></i>
              Role
            </label>
            <select
              id="role_id"
              v-model="formData.role_id"
              class="form-select"
              :disabled="loading"
            >
              <option value="">Select a role (optional)</option>
              <option v-for="role in roles" :key="role.id" :value="role.id">
                {{ role.role_name }}
              </option>
            </select>
          </div>

          <div class="form-group">
            <label for="max_unpayed_created_bills">
              <i class="fas fa-receipt"></i>
              Max Unpaid Bills
            </label>
            <input
              id="max_unpayed_created_bills"
              v-model.number="formData.max_unpayed_created_bills"
              type="number"
              class="form-input"
              placeholder="0"
              min="0"
              :disabled="loading"
            />
          </div>

          <div class="form-group">
            <label class="checkbox-label">
              <input
                id="is_diffrent_company"
                v-model="formData.is_diffrent_company"
                type="checkbox"
                class="form-checkbox"
                :disabled="loading"
              />
              <span>
                <i class="fas fa-building"></i>
                Different Company User
              </span>
            </label>
            <p class="form-hint">Enable this to manage company-specific assets (logo, letterhead, stamp)</p>
          </div>

          <div v-if="formData.is_diffrent_company" class="company-assets-section">
            <h4 class="section-title">
              <i class="fas fa-images"></i>
              Company Assets
            </h4>

            <div class="form-group">
              <label for="path_logo">
                <i class="fas fa-image"></i>
                Logo <span class="required">*</span>
              </label>
              <div v-if="logoPreview" class="image-preview">
                <img :src="logoPreview" alt="Logo preview" />
                <button
                  type="button"
                  @click.stop="formData.is_diffrent_company ? null : (logoPreview = null, logoFile = null, formData.path_logo = '')"
                  class="remove-image-btn"
                  :disabled="loading || formData.is_diffrent_company"
                  :title="formData.is_diffrent_company ? 'Logo cannot be removed for different company users' : 'Remove logo'"
                  :style="{ opacity: formData.is_diffrent_company ? 0.5 : 1, cursor: formData.is_diffrent_company ? 'not-allowed' : 'pointer' }"
                >
                  <i class="fas fa-times"></i>
                  Remove
                </button>
              </div>
              <div class="file-upload-wrapper">
                <input
                  id="path_logo"
                  type="file"
                  accept="image/*"
                  class="file-input"
                  @change="handleFileSelect($event, 'logo')"
                  :disabled="loading || uploadingLogo"
                />
                <label for="path_logo" class="file-upload-label">
                  <i class="fas fa-upload"></i>
                  {{ logoFile ? logoFile.name : 'Choose Logo Image' }}
                </label>
                <div v-if="uploadingLogo" class="upload-progress">
                  <i class="fas fa-spinner fa-spin"></i>
                  Uploading...
                </div>
              </div>
            </div>

            <div class="form-group">
              <label for="path_letter_head">
                <i class="fas fa-file-image"></i>
                Letterhead <span class="required">*</span>
              </label>
              <div v-if="letterHeadPreview" class="image-preview">
                <img :src="letterHeadPreview" alt="Letterhead preview" />
                <button
                  type="button"
                  @click.stop="formData.is_diffrent_company ? null : (letterHeadPreview = null, letterHeadFile = null, formData.path_letter_head = '')"
                  class="remove-image-btn"
                  :disabled="loading || formData.is_diffrent_company"
                  :title="formData.is_diffrent_company ? 'Letterhead cannot be removed for different company users' : 'Remove letterhead'"
                  :style="{ opacity: formData.is_diffrent_company ? 0.5 : 1, cursor: formData.is_diffrent_company ? 'not-allowed' : 'pointer' }"
                >
                  <i class="fas fa-times"></i>
                  Remove
                </button>
              </div>
              <div class="file-upload-wrapper">
                <input
                  id="path_letter_head"
                  type="file"
                  accept="image/*"
                  class="file-input"
                  @change="handleFileSelect($event, 'letter_head')"
                  :disabled="loading || uploadingLetterHead"
                />
                <label for="path_letter_head" class="file-upload-label">
                  <i class="fas fa-upload"></i>
                  {{ letterHeadFile ? letterHeadFile.name : 'Choose Letterhead Image' }}
                </label>
                <div v-if="uploadingLetterHead" class="upload-progress">
                  <i class="fas fa-spinner fa-spin"></i>
                  Uploading...
                </div>
              </div>
            </div>

            <div class="form-group">
              <label for="path_stamp">
                <i class="fas fa-stamp"></i>
                Stamp
              </label>
              <div v-if="stampPreview" class="image-preview">
                <img :src="stampPreview" alt="Stamp preview" />
                <button
                  type="button"
                  @click="stampPreview = null; stampFile = null; formData.path_stamp = ''"
                  class="remove-image-btn"
                  :disabled="loading"
                >
                  <i class="fas fa-times"></i>
                  Remove
                </button>
              </div>
              <div class="file-upload-wrapper">
                <input
                  id="path_stamp"
                  type="file"
                  accept="image/*"
                  class="file-input"
                  @change="handleFileSelect($event, 'stamp')"
                  :disabled="loading || uploadingStamp"
                />
                <label for="path_stamp" class="file-upload-label">
                  <i class="fas fa-upload"></i>
                  {{ stampFile ? stampFile.name : 'Choose Stamp Image' }}
                </label>
                <div v-if="uploadingStamp" class="upload-progress">
                  <i class="fas fa-spinner fa-spin"></i>
                  Uploading...
                </div>
              </div>
            </div>
          </div>

          <div v-if="error" class="error-message">
            <i class="fas fa-exclamation-circle"></i>
            {{ error }}
          </div>

          <div class="form-actions">
            <button type="button" @click="handleCancel" class="btn btn-cancel" :disabled="loading">
              <i class="fas fa-times"></i>
              Cancel
            </button>
            <button type="submit" class="btn btn-submit" :disabled="loading">
              <i v-if="loading" class="fas fa-spinner fa-spin"></i>
              <i v-else class="fas fa-save"></i>
              {{ loading ? 'Updating...' : 'Update User' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </teleport>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
  backdrop-filter: blur(2px);
}

.modal-content {
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.18);
  min-width: 400px;
  max-width: 600px;
  width: 90%;
  max-height: 90vh;
  overflow-y: auto;
  animation: modalSlideIn 0.3s ease-out;
}

@keyframes modalSlideIn {
  from {
    opacity: 0;
    transform: translateY(-20px) scale(0.95);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.form-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24px;
  border-bottom: 2px solid #e5e7eb;
}

.form-header h3 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #2c3e50;
  display: flex;
  align-items: center;
  gap: 12px;
}

.form-header h3 i {
  color: #3b82f6;
}

.close-btn {
  background: #ef4444;
  color: white;
  border: none;
  border-radius: 6px;
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s ease;
}

.close-btn:hover:not(:disabled) {
  background: #dc2626;
  transform: scale(1.05);
}

.close-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.edit-user-form {
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group label {
  font-weight: 500;
  color: #374151;
  display: flex;
  align-items: center;
  gap: 8px;
}

.form-group label i {
  color: #6b7280;
  width: 16px;
}

.required {
  color: #ef4444;
}

.form-input,
.form-select {
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  transition: all 0.2s ease;
  font-family: inherit;
}

.form-input:focus,
.form-select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-input:disabled,
.form-select:disabled {
  background-color: #f3f4f6;
  cursor: not-allowed;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  user-select: none;
}

.form-checkbox {
  width: 18px;
  height: 18px;
  cursor: pointer;
  accent-color: #10b981;
}

.form-hint {
  margin: 4px 0 0 0;
  font-size: 12px;
  color: #6b7280;
  font-style: italic;
}

.company-assets-section {
  margin-top: 20px;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
  border: 1px solid #e5e7eb;
}

.section-title {
  margin: 0 0 16px 0;
  font-size: 16px;
  font-weight: 600;
  color: #374151;
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-title i {
  color: #6366f1;
}

.image-preview {
  margin-bottom: 12px;
  position: relative;
  display: inline-block;
}

.image-preview img {
  max-width: 200px;
  max-height: 150px;
  border-radius: 6px;
  border: 2px solid #e5e7eb;
  object-fit: contain;
}

.remove-image-btn {
  position: absolute;
  top: -8px;
  right: -8px;
  background: #ef4444;
  color: white;
  border: none;
  border-radius: 50%;
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  font-size: 12px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  transition: all 0.2s ease;
}

.remove-image-btn:hover:not(:disabled) {
  background: #dc2626;
  transform: scale(1.1);
}

.remove-image-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.file-upload-wrapper {
  position: relative;
}

.file-input {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
}

.file-upload-label {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  background: #3b82f6;
  color: white;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 14px;
  font-weight: 500;
}

.file-upload-label:hover {
  background: #2563eb;
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(59, 130, 246, 0.2);
}

.file-input:disabled + .file-upload-label {
  opacity: 0.6;
  cursor: not-allowed;
}

.upload-progress {
  margin-top: 8px;
  color: #3b82f6;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.error-message {
  padding: 12px;
  background-color: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 6px;
  color: #dc2626;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 8px;
  padding-top: 20px;
  border-top: 1px solid #e5e7eb;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-cancel {
  background-color: #f3f4f6;
  color: #374151;
}

.btn-cancel:hover:not(:disabled) {
  background-color: #e5e7eb;
}

.btn-submit {
  background-color: #3b82f6;
  color: white;
}

.btn-submit:hover:not(:disabled) {
  background-color: #2563eb;
}
</style>

