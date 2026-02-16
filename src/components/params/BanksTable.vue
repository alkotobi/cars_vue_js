<script setup>
import { ref, onMounted } from 'vue'
import { useApi } from '../../composables/useApi'

const { callApi, uploadFile, getFileUrl } = useApi()
const banks = ref([])
const loading = ref(false)
const error = ref(null)
const showAddDialog = ref(false)
const editingBank = ref(null)
const isProcessing = ref(false)
const logoFile = ref(null)
const logoPreviewUrl = ref('')

// Form data
const formData = ref({
  company_name: '',
  company_address: '',
  bank_name: '',
  swift_code: '',
  bank_account: '',
  bank_address: '',
  mobile: '',
  email: '',
  website: '',
  logo_path: '',
  is_active: true,
  notes: '',
})

const resetForm = () => {
  logoFile.value = null
  logoPreviewUrl.value = ''
  formData.value = {
    company_name: '',
    company_address: '',
    bank_name: '',
    swift_code: '',
    bank_account: '',
    bank_address: '',
    mobile: '',
    email: '',
    website: '',
    logo_path: '',
    is_active: true,
    notes: '',
  }
}

const fetchBanks = async () => {
  loading.value = true
  error.value = null
  try {
    const result = await callApi({
      query: 'SELECT * FROM banks ORDER BY is_active DESC, company_name ASC',
    })
    if (result.success) {
      banks.value = result.data
    } else {
      error.value = 'Failed to fetch banks'
    }
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

const handleAdd = () => {
  resetForm()
  showAddDialog.value = true
  editingBank.value = null
}

const handleEdit = (bank) => {
  logoFile.value = null
  formData.value = {
    company_name: bank.company_name || '',
    company_address: bank.company_address || '',
    bank_name: bank.bank_name || '',
    swift_code: bank.swift_code || '',
    bank_account: bank.bank_account || '',
    bank_address: bank.bank_address || '',
    mobile: bank.mobile || '',
    email: bank.email || '',
    website: bank.website || '',
    logo_path: bank.logo_path || '',
    is_active: bank.is_active === 1 || bank.is_active === true || bank.is_active === '1',
    notes: bank.notes || '',
  }
  logoPreviewUrl.value = formData.value.logo_path ? getFileUrl(formData.value.logo_path) : ''
  showAddDialog.value = true
  editingBank.value = bank
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
  const file = e.dataTransfer?.files?.[0]
  handleLogoFile(file)
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

const uploadLogoFile = async () => {
  if (!logoFile.value) return formData.value.logo_path
  const filename = `bank_logo_${Date.now()}.${logoFile.value.name.split('.').pop() || 'png'}`
  const result = await uploadFile(logoFile.value, 'banks_logos', filename)
  if (result.success && result.relativePath) {
    formData.value.logo_path = result.relativePath
    logoFile.value = null
    if (logoPreviewUrl.value && logoPreviewUrl.value.startsWith('blob:')) {
      URL.revokeObjectURL(logoPreviewUrl.value)
    }
    logoPreviewUrl.value = getFileUrl(result.relativePath)
    return result.relativePath
  }
  throw new Error(result.message || 'Upload failed')
}

const handleDelete = async (bank) => {
  if (!confirm('Are you sure you want to delete this bank?')) return

  loading.value = true
  error.value = null
  try {
    const result = await callApi({
      query: 'DELETE FROM banks WHERE id = ?',
      params: [bank.id],
    })
    if (result.success) {
      await fetchBanks()
    } else {
      error.value = 'Failed to delete bank'
    }
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}

const handleSubmit = async () => {
  if (isProcessing.value) return
  isProcessing.value = true
  error.value = null

  try {
    if (logoFile.value) {
      await uploadLogoFile()
    }
    // When setting a bank active: only one can be active; deactivate all first
    if (formData.value.is_active) {
      const deactivateResult = await callApi({
        query: 'UPDATE banks SET is_active = 0',
        params: [],
      })
      if (!deactivateResult.success) {
        error.value = 'Failed to update bank status'
        return
      }
    }

    let result
    if (editingBank.value) {
      // Update
      result = await callApi({
        query: `
          UPDATE banks 
          SET company_name = ?, company_address = ?, bank_name = ?, swift_code = ?, 
              bank_account = ?, bank_address = ?, mobile = ?, email = ?, website = ?,
              logo_path = ?, is_active = ?, notes = ?
          WHERE id = ?
        `,
        params: [
          formData.value.company_name,
          formData.value.company_address,
          formData.value.bank_name,
          formData.value.swift_code,
          formData.value.bank_account,
          formData.value.bank_address,
          formData.value.mobile,
          formData.value.email,
          formData.value.website,
          formData.value.logo_path,
          formData.value.is_active ? 1 : 0,
          formData.value.notes,
          editingBank.value.id,
        ],
      })
    } else {
      // Insert
      result = await callApi({
        query: `
          INSERT INTO banks 
          (company_name, company_address, bank_name, swift_code, bank_account, bank_address, mobile, email, website, logo_path, is_active, notes)
          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        `,
        params: [
          formData.value.company_name,
          formData.value.company_address,
          formData.value.bank_name,
          formData.value.swift_code,
          formData.value.bank_account,
          formData.value.bank_address,
          formData.value.mobile,
          formData.value.email,
          formData.value.website,
          formData.value.logo_path,
          formData.value.is_active ? 1 : 0,
          formData.value.notes,
        ],
      })
    }

    if (result.success) {
      showAddDialog.value = false
      await fetchBanks()
      window.dispatchEvent(new CustomEvent('companyInfoUpdated'))
    } else {
      error.value = 'Failed to save bank'
    }
  } catch (err) {
    error.value = err.message
  } finally {
    isProcessing.value = false
  }
}

onMounted(() => {
  fetchBanks()
})
</script>

<template>
  <div class="banks-table">
    <div class="header">
      <h2>Company info</h2>
      <button @click="handleAdd" class="add-btn" :disabled="loading">
        <i class="fas fa-plus"></i>
        Add
      </button>
    </div>

    <div v-if="error" class="error-message">
      {{ error }}
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading">
      <i class="fas fa-spinner fa-spin"></i>
      Loading...
    </div>

    <!-- Banks Table -->
    <div v-else class="table-container">
      <table v-if="banks.length > 0">
        <thead>
          <tr>
            <th>Active</th>
            <th>Company Name</th>
            <th>Bank Name</th>
            <th>Swift Code</th>
            <th>Account</th>
            <th>Mobile</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="bank in banks" :key="bank.id" :class="{ inactive: !(bank.is_active === 1 || bank.is_active === true || bank.is_active === '1') }">
            <td>
              <i v-if="bank.is_active === 1 || bank.is_active === true || bank.is_active === '1'" class="fas fa-check-circle active-icon" title="Active"></i>
              <i v-else class="fas fa-times-circle inactive-icon" title="Inactive"></i>
            </td>
            <td>{{ bank.company_name }}</td>
            <td>{{ bank.bank_name }}</td>
            <td>{{ bank.swift_code }}</td>
            <td>{{ bank.bank_account }}</td>
            <td>{{ bank.mobile }}</td>
            <td class="actions">
              <button @click="handleEdit(bank)" class="edit-btn">
                <i class="fas fa-edit"></i>
              </button>
              <button @click="handleDelete(bank)" class="delete-btn">
                <i class="fas fa-trash"></i>
              </button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-else class="empty-state">No company info found. Click "Add" to create one.</div>
    </div>

    <!-- Add/Edit Dialog: teleported to body, above app header and alerts -->
    <teleport to="body">
      <div v-if="showAddDialog" class="dialog-overlay">
        <div class="dialog">
          <div class="dialog-header">
            <h3>{{ editingBank ? 'Edit Company info' : 'Add Company info' }}</h3>
            <button @click="showAddDialog = false" class="close-btn">
              <i class="fas fa-times"></i>
            </button>
          </div>

          <form @submit.prevent="handleSubmit" class="bank-form">
            <div class="form-columns">
              <div class="form-column">
                <div class="form-group">
                  <label for="company_name">Company Name</label>
                  <input
                    id="company_name"
                    v-model="formData.company_name"
                    type="text"
                    required
                    placeholder="Enter company name"
                  />
                </div>
                <div class="form-group">
                  <label for="company_address">Company Address</label>
                  <textarea
                    id="company_address"
                    v-model="formData.company_address"
                    placeholder="Enter company address"
                    rows="2"
                  ></textarea>
                </div>
                <div class="form-group">
                  <label for="bank_name">Bank Name</label>
                  <input
                    id="bank_name"
                    v-model="formData.bank_name"
                    type="text"
                    required
                    placeholder="Enter bank name"
                  />
                </div>
                <div class="form-group">
                  <label for="swift_code">Swift Code</label>
                  <input
                    id="swift_code"
                    v-model="formData.swift_code"
                    type="text"
                    required
                    placeholder="Enter SWIFT code"
                  />
                </div>
                <div class="form-group">
                  <label for="bank_account">Account Number</label>
                  <input
                    id="bank_account"
                    v-model="formData.bank_account"
                    type="text"
                    required
                    placeholder="Enter account number"
                  />
                </div>
                <div class="form-group">
                  <label for="bank_address">Bank Address</label>
                  <input
                    id="bank_address"
                    v-model="formData.bank_address"
                    type="text"
                    placeholder="Enter bank address"
                  />
                </div>
              </div>
              <div class="form-column">
                <div class="form-row">
                  <div class="form-group">
                    <label for="mobile">Mobile</label>
                    <input
                      id="mobile"
                      v-model="formData.mobile"
                      type="text"
                      placeholder="Phone / mobile"
                    />
                  </div>
                  <div class="form-group">
                    <label for="email">Email</label>
                    <input
                      id="email"
                      v-model="formData.email"
                      type="email"
                      placeholder="company@example.com"
                    />
                  </div>
                </div>
                <div class="form-group">
                  <label for="website">Website</label>
                  <input
                    id="website"
                    v-model="formData.website"
                    type="url"
                    placeholder="https://example.com"
                  />
                </div>
                <div class="form-group">
                  <label>Logo</label>
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
                      <button
                        type="button"
                        class="logo-clear-btn"
                        title="Remove logo"
                        @click.stop="clearLogo"
                      >
                        <i class="fas fa-times"></i>
                      </button>
                    </template>
                    <template v-else>
                      <i class="fas fa-cloud-upload-alt logo-placeholder"></i>
                      <span class="logo-hint">Drop logo here or click to select</span>
                    </template>
                  </div>
                </div>
                <div class="form-group checkbox-group">
                  <label>
                    <input
                      v-model="formData.is_active"
                      type="checkbox"
                    />
                    Active
                  </label>
                </div>
                <div class="form-group">
                  <label for="notes">Notes</label>
                  <textarea
                    id="notes"
                    v-model="formData.notes"
                    placeholder="Enter additional notes"
                    rows="3"
                  ></textarea>
                </div>
              </div>
            </div>

            <div class="form-actions">
              <button
                type="button"
                @click="showAddDialog = false"
                class="cancel-btn"
                :disabled="isProcessing"
              >
                Cancel
              </button>
              <button type="submit" class="save-btn" :disabled="isProcessing">
                <i v-if="isProcessing" class="fas fa-spinner fa-spin"></i>
                {{ isProcessing ? 'Saving...' : 'Save' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </teleport>
  </div>
</template>

<style scoped>
.banks-table {
  width: 100%;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.header h2 {
  margin: 0;
  font-size: 1.5rem;
  color: #1f2937;
}

.add-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background-color: #10b981;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.875rem;
  transition: all 0.2s ease;
}

.add-btn:hover {
  background-color: #059669;
}

.table-container {
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th,
td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

th {
  background-color: #f9fafb;
  font-weight: 600;
  color: #4b5563;
}

.actions {
  display: flex;
  gap: 8px;
}

.edit-btn,
.delete-btn {
  padding: 6px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.edit-btn {
  background-color: #3b82f6;
  color: white;
}

.edit-btn:hover {
  background-color: #2563eb;
}

.delete-btn {
  background-color: #ef4444;
  color: white;
}

.delete-btn:hover {
  background-color: #dc2626;
}

.dialog-overlay {
  position: fixed;
  inset: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 2000; /* Above app header (1001) and alerts (999) */
}

.dialog {
  background-color: white;
  border-radius: 8px;
  width: 90%;
  max-width: 900px;
  max-height: 90vh;
  overflow-y: auto;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid #e5e7eb;
}

.dialog-header h3 {
  margin: 0;
  font-size: 1.25rem;
  color: #1f2937;
}

.close-btn {
  background: none;
  border: none;
  color: #6b7280;
  cursor: pointer;
  font-size: 1.25rem;
  padding: 4px;
}

.close-btn:hover {
  color: #1f2937;
}

.bank-form {
  padding: 16px;
}

.form-columns {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
}

.form-column {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.form-row {
  display: flex;
  gap: 16px;
}

.form-row .form-group {
  flex: 1;
}

.form-group {
  margin-bottom: 16px;
}

.checkbox-group label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.checkbox-group input[type="checkbox"] {
  width: auto;
  margin: 0;
}

tr.inactive {
  opacity: 0.6;
}

.active-icon {
  color: #10b981;
}

.inactive-icon {
  color: #9ca3af;
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

.form-group label {
  display: block;
  margin-bottom: 4px;
  color: #4b5563;
  font-weight: 500;
}

.form-group input,
.form-group textarea {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 0.875rem;
}

.form-group input:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #3b82f6;
  ring: 2px solid #3b82f6;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
}

.cancel-btn,
.save-btn {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.875rem;
  transition: all 0.2s ease;
}

.cancel-btn {
  background-color: #e5e7eb;
  color: #4b5563;
}

.cancel-btn:hover {
  background-color: #d1d5db;
}

.save-btn {
  background-color: #3b82f6;
  color: white;
  display: flex;
  align-items: center;
  gap: 8px;
}

.save-btn:hover {
  background-color: #2563eb;
}

.save-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.error-message {
  padding: 12px;
  background-color: #fee2e2;
  border: 1px solid #ef4444;
  border-radius: 6px;
  color: #b91c1c;
  margin-bottom: 16px;
}

.loading {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #6b7280;
  padding: 24px;
  justify-content: center;
}

.empty-state {
  text-align: center;
  padding: 24px;
  color: #6b7280;
}

@media (max-width: 768px) {
  .dialog {
    width: 95%;
    margin: 16px;
  }

  .form-columns {
    grid-template-columns: 1fr;
  }

  .form-row {
    flex-direction: column;
  }

  .form-actions {
    flex-direction: column;
  }

  .form-actions button {
    width: 100%;
  }
}
</style>
