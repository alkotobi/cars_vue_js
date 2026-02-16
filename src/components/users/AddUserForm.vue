<script setup>
import { ref, onMounted } from 'vue'
import { useApi } from '../../composables/useApi'

const emit = defineEmits(['user-added', 'close'])

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
  // Bank fields (all except is_active - new record is never active)
  company_name: '',
  company_address: '',
  bank_name: '',
  swift_code: '',
  bank_account: '',
  bank_address: '',
  company_mobile: '',
  company_email: '',
  company_website: '',
  company_notes: '',
  logo_path: '',
})
const logoFile = ref(null)
const logoPreviewUrl = ref('')

const fetchRoles = async () => {
  const result = await callApi({
    query: 'SELECT * FROM roles ORDER BY role_name',
  })
  if (result.success) {
    roles.value = result.data
  }
}

const handleSubmit = async () => {
  if (!formData.value.username || !formData.value.password || !formData.value.email) {
    error.value = 'Username, email, and password are required'
    return
  }

  if (formData.value.is_diffrent_company) {
    if (!formData.value.company_name || formData.value.company_name.trim() === '') {
      error.value = 'Company name is required for different company users'
      return
    }
    if (!logoFile.value) {
      error.value = 'Logo is required for different company users. Please select or drop an image file.'
      return
    }
  }

  loading.value = true
  error.value = null

  try {
    let bankId = null
    if (formData.value.is_diffrent_company) {
      let logoPath = formData.value.logo_path
      if (logoFile.value) {
        const filename = `bank_logo_${Date.now()}.${logoFile.value.name.split('.').pop() || 'png'}`
        const uploadResult = await uploadFile(logoFile.value, 'banks_logos', filename)
        if (!uploadResult.success || !uploadResult.relativePath) {
          error.value = uploadResult.message || 'Failed to upload logo'
          loading.value = false
          return
        }
        logoPath = uploadResult.relativePath
      }
      const bankResult = await callApi({
        query:
          'INSERT INTO banks (company_name, company_address, bank_name, swift_code, bank_account, bank_address, mobile, email, website, logo_path, notes, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)',
        params: [
          formData.value.company_name?.trim() || null,
          formData.value.company_address?.trim() || null,
          formData.value.bank_name?.trim() || null,
          formData.value.swift_code?.trim() || null,
          formData.value.bank_account?.trim() || null,
          formData.value.bank_address?.trim() || null,
          formData.value.company_mobile?.trim() || null,
          formData.value.company_email?.trim() || null,
          formData.value.company_website?.trim() || null,
          logoPath,
          formData.value.company_notes?.trim() || null,
        ],
      })
      if (!bankResult.success || !bankResult.lastInsertId) {
        error.value = bankResult.error || 'Failed to create company record'
        loading.value = false
        return
      }
      bankId = bankResult.lastInsertId
    }

    const result = await callApi({
      query:
        'INSERT INTO users (username, email, password, role_id, max_unpayed_created_bills, is_diffrent_company, id_bank_account) VALUES (?, ?, ?, ?, ?, ?, ?)',
      params: [
        formData.value.username,
        formData.value.email,
        formData.value.password,
        formData.value.role_id || null,
        formData.value.max_unpayed_created_bills || 0,
        formData.value.is_diffrent_company ? 1 : 0,
        bankId,
      ],
      action: 'insert_user', // Tell API to hash the password before inserting
    })

    if (result.success) {
      formData.value = {
        username: '',
        email: '',
        password: '',
        role_id: '',
        max_unpayed_created_bills: 0,
        is_diffrent_company: false,
        company_name: '',
        company_address: '',
        bank_name: '',
        swift_code: '',
        bank_account: '',
        bank_address: '',
        company_mobile: '',
        company_email: '',
        company_website: '',
        company_notes: '',
        logo_path: '',
      }
      logoFile.value = null
      logoPreviewUrl.value = ''
      error.value = null
      emit('user-added')
    } else {
      error.value = result.error || 'Failed to add user'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
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
    company_name: '',
    company_address: '',
    bank_name: '',
    swift_code: '',
    bank_account: '',
    bank_address: '',
    company_mobile: '',
    company_email: '',
    company_website: '',
    company_notes: '',
    logo_path: '',
  }
  logoFile.value = null
  logoPreviewUrl.value = ''
  error.value = null
  emit('close')
}

const handleLogoFile = (file) => {
  if (!file) return
  if (!file.type.startsWith('image/')) {
    error.value = 'Please select an image file (PNG, JPEG, or WebP)'
    return
  }
  if (file.size > 5 * 1024 * 1024) {
    error.value = 'Logo must be less than 5MB'
    return
  }
  if (logoPreviewUrl.value && logoPreviewUrl.value.startsWith('blob:')) {
    URL.revokeObjectURL(logoPreviewUrl.value)
  }
  logoFile.value = file
  logoPreviewUrl.value = URL.createObjectURL(file)
}

const onLogoDrop = (e) => {
  e.preventDefault()
  e.stopPropagation()
  handleLogoFile(e.dataTransfer?.files?.[0])
}

const onLogoDragOver = (e) => {
  e.preventDefault()
  e.stopPropagation()
}

const clearLogo = () => {
  if (logoPreviewUrl.value && logoPreviewUrl.value.startsWith('blob:')) {
    URL.revokeObjectURL(logoPreviewUrl.value)
  }
  logoFile.value = null
  formData.value.logo_path = ''
  logoPreviewUrl.value = ''
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
        <select id="role_id" v-model="formData.role_id" class="form-select" :disabled="loading">
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
        <p class="form-hint">
          Enable to link this user to a company record (bank). A new bank row is created and is not set active.
        </p>
      </div>

      <div v-if="formData.is_diffrent_company" class="company-assets-section">
        <div class="form-columns">
          <div class="form-column">
            <div class="form-group">
              <label>Logo <span class="required">*</span></label>
              <div
                class="logo-dropzone"
                :class="{ 'has-preview': logoPreviewUrl }"
                @click="$refs.logoInput?.click()"
                @drop="onLogoDrop"
                @dragover="onLogoDragOver"
              >
                <input
                  ref="logoInput"
                  type="file"
                  accept="image/png,image/jpeg,image/jpg,image/webp"
                  class="logo-input-hidden"
                  @change="(e) => handleLogoFile(e.target.files?.[0])"
                />
                <template v-if="logoPreviewUrl">
                  <img :src="logoPreviewUrl" alt="Logo" class="logo-preview" />
                  <button type="button" class="logo-clear-btn" title="Remove logo" @click.stop="clearLogo">
                    <i class="fas fa-times"></i>
                  </button>
                </template>
                <template v-else>
                  <i class="fas fa-cloud-upload-alt logo-placeholder"></i>
                  <span class="logo-hint">Drop logo here or click to select</span>
                </template>
              </div>
            </div>
            <div class="form-group">
              <label for="company_name">
                <i class="fas fa-building"></i>
                Company Name <span class="required">*</span>
              </label>
              <input
                id="company_name"
                v-model="formData.company_name"
                type="text"
                class="form-input"
                placeholder="Enter company name"
                :disabled="loading"
              />
            </div>
            <div class="form-group">
              <label for="company_address">Company Address</label>
              <textarea
                id="company_address"
                v-model="formData.company_address"
                class="form-input"
                placeholder="Company address"
                rows="2"
                :disabled="loading"
              />
            </div>
            <div class="form-group">
              <label for="bank_name">Bank Name</label>
              <input
                id="bank_name"
                v-model="formData.bank_name"
                type="text"
                class="form-input"
                placeholder="Bank name"
                :disabled="loading"
              />
            </div>
            <div class="form-group">
              <label for="swift_code">Swift Code</label>
              <input
                id="swift_code"
                v-model="formData.swift_code"
                type="text"
                class="form-input"
                placeholder="SWIFT code"
                :disabled="loading"
              />
            </div>
          </div>
          <div class="form-column">
            <div class="form-group">
              <label for="bank_account">Account Number</label>
              <input
                id="bank_account"
                v-model="formData.bank_account"
                type="text"
                class="form-input"
                placeholder="Account number"
                :disabled="loading"
              />
            </div>
            <div class="form-group">
              <label for="bank_address">Bank Address</label>
              <input
                id="bank_address"
                v-model="formData.bank_address"
                type="text"
                class="form-input"
                placeholder="Bank address"
                :disabled="loading"
              />
            </div>
            <div class="form-row bank-fields">
              <div class="form-group">
                <label for="company_mobile">Mobile</label>
                <input
                  id="company_mobile"
                  v-model="formData.company_mobile"
                  type="text"
                  class="form-input"
                  placeholder="Phone"
                  :disabled="loading"
                />
              </div>
              <div class="form-group">
                <label for="company_email">Company Email</label>
                <input
                  id="company_email"
                  v-model="formData.company_email"
                  type="email"
                  class="form-input"
                  placeholder="company@example.com"
                  :disabled="loading"
                />
              </div>
            </div>
            <div class="form-group">
              <label for="company_website">Website</label>
              <input
                id="company_website"
                v-model="formData.company_website"
                type="url"
                class="form-input"
                placeholder="https://..."
                :disabled="loading"
              />
            </div>
            <div class="form-group">
              <label for="company_notes">Notes</label>
              <textarea
                id="company_notes"
                v-model="formData.company_notes"
                class="form-input"
                placeholder="Notes"
                rows="2"
                :disabled="loading"
              />
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

.company-assets-section .form-columns {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
}

@media (max-width: 640px) {
  .company-assets-section .form-columns {
    grid-template-columns: 1fr;
  }
  .company-assets-section .form-row.bank-fields {
    grid-template-columns: 1fr;
  }
}

.company-assets-section .form-column {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.company-assets-section .form-row.bank-fields {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.logo-dropzone {
  position: relative;
  border: 2px dashed #d1d5db;
  border-radius: 8px;
  padding: 24px;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s ease;
  background: #f9fafb;
  min-height: 120px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.logo-dropzone:hover {
  border-color: #3b82f6;
  background: #eff6ff;
}

.logo-dropzone.has-preview {
  padding: 12px;
}

.logo-input-hidden {
  position: absolute;
  width: 0;
  height: 0;
  opacity: 0;
  overflow: hidden;
}

.logo-placeholder {
  font-size: 2rem;
  color: #9ca3af;
  margin-bottom: 8px;
}

.logo-hint {
  font-size: 0.875rem;
  color: #6b7280;
}

.logo-preview {
  max-width: 100%;
  max-height: 120px;
  object-fit: contain;
}

.logo-clear-btn {
  position: absolute;
  top: 8px;
  right: 8px;
  width: 28px;
  height: 28px;
  padding: 0;
  border: none;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.6);
  color: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.75rem;
  transition: background 0.2s;
}

.logo-clear-btn:hover {
  background: #ef4444;
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
