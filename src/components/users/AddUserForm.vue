<script setup>
import { ref, onMounted } from 'vue'
import { useApi } from '../../composables/useApi'

const emit = defineEmits(['user-added', 'close'])

const { callApi, uploadFile } = useApi()
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

const fetchRoles = async () => {
  const result = await callApi({
    query: 'SELECT * FROM roles ORDER BY role_name',
  })
  if (result.success) {
    roles.value = result.data
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

  if (type === 'logo') {
    logoFile.value = file
  } else if (type === 'letter_head') {
    letterHeadFile.value = file
  } else if (type === 'stamp') {
    stampFile.value = file
  }
}

const uploadImageFile = async (file, type) => {
  if (!file) return null

  try {
    const timestamp = Date.now()
    const fileExtension = file.name.split('.').pop()
    const filename = `${type}_${timestamp}.${fileExtension}`
    const result = await uploadFile(file, `company_assets/${type}`, filename)

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
  if (!formData.value.username || !formData.value.password || !formData.value.email) {
    error.value = 'Username, email, and password are required'
    return
  }

  loading.value = true
  error.value = null

  try {
    // Upload image files if is_diffrent_company is true
    if (formData.value.is_diffrent_company) {
      if (logoFile.value) {
        uploadingLogo.value = true
        formData.value.path_logo = await uploadImageFile(logoFile.value, 'logo')
        uploadingLogo.value = false
      }
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
    }

    const result = await callApi({
      query:
        'INSERT INTO users (username, email, password, role_id, max_unpayed_created_bills, is_diffrent_company, path_logo, path_letter_head, path_stamp) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)',
      params: [
        formData.value.username,
        formData.value.email,
        formData.value.password,
        formData.value.role_id || null,
        formData.value.max_unpayed_created_bills || 0,
        formData.value.is_diffrent_company ? 1 : 0,
        formData.value.path_logo || null,
        formData.value.path_letter_head || null,
        formData.value.path_stamp || null,
      ],
      action: 'insert_user', // Tell API to hash the password before inserting
    })

    if (result.success) {
      // Reset form
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
      error.value = null
      emit('user-added')
    } else {
      error.value = result.error || 'Failed to add user'
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
  error.value = null
  emit('close')
}

onMounted(() => {
  fetchRoles()
})
</script>

<template>
  <div class="add-user-form-container">
    <div class="form-header">
      <h3>
        <i class="fas fa-user-plus"></i>
        Add New User
      </h3>
      <button @click="handleCancel" class="close-btn" :disabled="loading">
        <i class="fas fa-times"></i>
      </button>
    </div>

    <form @submit.prevent="handleSubmit" class="add-user-form">
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
          Password <span class="required">*</span>
        </label>
        <input
          id="password"
          v-model="formData.password"
          type="password"
          class="form-input"
          placeholder="Enter password"
          required
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
        <p class="form-hint">Enable this to add company-specific assets (logo, letterhead, stamp)</p>
      </div>

      <div v-if="formData.is_diffrent_company" class="company-assets-section">
        <h4 class="section-title">
          <i class="fas fa-images"></i>
          Company Assets
        </h4>

        <div class="form-group">
          <label for="path_logo">
            <i class="fas fa-image"></i>
            Logo
          </label>
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
            Letterhead
          </label>
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
          <i v-else class="fas fa-check"></i>
          {{ loading ? 'Adding...' : 'Add User' }}
        </button>
      </div>
    </form>
  </div>
</template>

<style scoped>
.add-user-form-container {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  padding: 24px;
  margin-bottom: 24px;
}

.form-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding-bottom: 16px;
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
  color: #10b981;
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

.add-user-form {
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
  background-color: #10b981;
  color: white;
}

.btn-submit:hover:not(:disabled) {
  background-color: #059669;
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
</style>

