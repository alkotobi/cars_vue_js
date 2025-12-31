<script setup>
import { ref, onMounted } from 'vue'
import { useApi } from '../../composables/useApi'
import { BASE_PATH } from '../../composables/useApi'

const { callApi } = useApi()
const loading = ref(false)
const error = ref(null)
const isProcessing = ref(false)
const hasChanges = ref(false)

const contractTerms = ref([])
const enabledLanguages = ref({
  english: false,
  chinese: false,
  french: false,
  arabic: false,
})

const editingTerm = ref(null)
const showAddForm = ref(false)
const showEditForm = ref(false)

// Helper function to get base path
const getBasePath = () => {
  return BASE_PATH || '/'
}

// Load contract terms from JSON file
const loadContractTerms = async () => {
  loading.value = true
  error.value = null
  try {
    const basePath = getBasePath()
    const termsUrl = `${basePath}contract_terms.json`
    const response = await fetch(termsUrl)
    
    if (!response.ok) {
      throw new Error('Failed to load contract terms')
    }
    
    const data = await response.json()
    contractTerms.value = data.terms || []
    enabledLanguages.value = data.enabledLanguages || {
      english: false,
      chinese: false,
      french: false,
      arabic: false,
    }
    hasChanges.value = false
  } catch (err) {
    error.value = err.message
    console.error('Error loading contract terms:', err)
  } finally {
    loading.value = false
  }
}

// Form data for new/edit term
const formData = ref({
  english: '',
  chinese: '',
  french: '',
  arabic: '',
})

const resetForm = () => {
  formData.value = {
    english: '',
    chinese: '',
    french: '',
    arabic: '',
  }
  editingTerm.value = null
  showAddForm.value = false
  showEditForm.value = false
}

const handleAdd = () => {
  resetForm()
  showAddForm.value = true
}

const handleEdit = (term) => {
  editingTerm.value = term
  formData.value = {
    english: term.english || '',
    chinese: term.chinese || '',
    french: term.french || '',
    arabic: term.arabic || '',
  }
  showEditForm.value = true
}

const handleDelete = async (termId) => {
  if (!confirm('Are you sure you want to delete this term?')) {
    return
  }
  
  isProcessing.value = true
  error.value = null
  
  try {
    contractTerms.value = contractTerms.value.filter(t => t.id !== termId)
    hasChanges.value = true
    
    // Auto-save on delete
    await saveContractTerms()
  } catch (err) {
    error.value = err.message
  } finally {
    isProcessing.value = false
  }
}

const handleSave = async () => {
  if (isProcessing.value) return
  isProcessing.value = true
  error.value = null
  
  try {
    if (showEditForm.value && editingTerm.value) {
      // Update existing term
      const index = contractTerms.value.findIndex(t => t.id === editingTerm.value.id)
      if (index !== -1) {
        contractTerms.value[index] = {
          ...contractTerms.value[index],
          ...formData.value,
        }
      }
    } else {
      // Add new term
      const maxId = contractTerms.value.length > 0
        ? Math.max(...contractTerms.value.map(t => t.id))
        : 0
      contractTerms.value.push({
        id: maxId + 1,
        ...formData.value,
      })
    }
    
    hasChanges.value = true
    resetForm()
    
    // Auto-save
    await saveContractTerms()
  } catch (err) {
    error.value = err.message
  } finally {
    isProcessing.value = false
  }
}

const handleLanguageToggle = () => {
  hasChanges.value = true
}

const saveContractTerms = async () => {
  if (isProcessing.value) return
  
  // Get current user to check if admin
  const userStr = localStorage.getItem('user')
  const user = userStr ? JSON.parse(userStr) : null
  const isAdmin = user?.role_id === 1
  
  if (!isAdmin) {
    error.value = 'Admin access required'
    return
  }
  
  isProcessing.value = true
  error.value = null
  
  try {
    const dataToSave = {
      enabledLanguages: enabledLanguages.value,
      terms: contractTerms.value,
    }
    
    const result = await callApi({
      action: 'save_contract_terms',
      content: dataToSave,
      is_admin: isAdmin,
      requiresAuth: true,
    })
    
    if (result.success) {
      hasChanges.value = false
    } else {
      error.value = result.error || result.message || 'Failed to save contract terms'
    }
  } catch (err) {
    error.value = err.message
  } finally {
    isProcessing.value = false
  }
}

const handleCancel = () => {
  resetForm()
  if (hasChanges.value) {
    loadContractTerms()
  }
}

onMounted(() => {
  loadContractTerms()
})
</script>

<template>
  <div class="contract-terms-form">
    <div class="header">
      <h2>Contract Terms Management</h2>
      <p class="description">Manage contract terms and conditions for billing documents</p>
    </div>

    <div v-if="error" class="error-message">
      {{ error }}
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading">
      <i class="fas fa-spinner fa-spin"></i>
      Loading contract terms...
    </div>

    <!-- Content -->
    <div v-else class="content-container">
      <!-- Language Toggles -->
      <div class="section">
        <h3>Enabled Languages</h3>
        <div class="language-toggles">
          <label class="toggle-label">
            <input
              type="checkbox"
              v-model="enabledLanguages.english"
              @change="handleLanguageToggle"
            />
            <span>English</span>
          </label>
          <label class="toggle-label">
            <input
              type="checkbox"
              v-model="enabledLanguages.chinese"
              @change="handleLanguageToggle"
            />
            <span>Chinese</span>
          </label>
          <label class="toggle-label">
            <input
              type="checkbox"
              v-model="enabledLanguages.french"
              @change="handleLanguageToggle"
            />
            <span>French</span>
          </label>
          <label class="toggle-label">
            <input
              type="checkbox"
              v-model="enabledLanguages.arabic"
              @change="handleLanguageToggle"
            />
            <span>Arabic</span>
          </label>
        </div>
        <button
          v-if="hasChanges"
          @click="saveContractTerms"
          class="save-languages-btn"
          :disabled="isProcessing"
        >
          <i v-if="isProcessing" class="fas fa-spinner fa-spin"></i>
          {{ isProcessing ? 'Saving...' : 'Save Language Settings' }}
        </button>
      </div>

      <!-- Terms List -->
      <div class="section">
        <div class="section-header">
          <h3>Terms and Conditions</h3>
          <button @click="handleAdd" class="add-btn" :disabled="showAddForm || showEditForm">
            <i class="fas fa-plus"></i>
            Add New Term
          </button>
        </div>

        <!-- Add/Edit Form -->
        <div v-if="showAddForm || showEditForm" class="form-container">
          <h4>{{ showEditForm ? 'Edit Term' : 'Add New Term' }}</h4>
          <div class="form-grid">
            <div class="form-group" v-if="enabledLanguages.english">
              <label>English</label>
              <textarea
                v-model="formData.english"
                rows="3"
                placeholder="Enter term in English"
              ></textarea>
            </div>
            <div class="form-group" v-if="enabledLanguages.chinese">
              <label>Chinese</label>
              <textarea
                v-model="formData.chinese"
                rows="3"
                placeholder="Enter term in Chinese"
              ></textarea>
            </div>
            <div class="form-group" v-if="enabledLanguages.french">
              <label>French</label>
              <textarea
                v-model="formData.french"
                rows="3"
                placeholder="Enter term in French"
              ></textarea>
            </div>
            <div class="form-group" v-if="enabledLanguages.arabic">
              <label>Arabic</label>
              <textarea
                v-model="formData.arabic"
                rows="3"
                placeholder="Enter term in Arabic"
                dir="rtl"
              ></textarea>
            </div>
          </div>
          <div class="form-actions">
            <button @click="handleCancel" class="cancel-btn" :disabled="isProcessing">
              Cancel
            </button>
            <button @click="handleSave" class="save-btn" :disabled="isProcessing">
              <i v-if="isProcessing" class="fas fa-spinner fa-spin"></i>
              {{ isProcessing ? 'Saving...' : 'Save Term' }}
            </button>
          </div>
        </div>

        <!-- Terms List -->
        <div class="terms-list">
          <div
            v-for="term in contractTerms"
            :key="term.id"
            class="term-item"
          >
            <div class="term-header">
              <span class="term-id">#{{ term.id }}</span>
              <div class="term-actions">
                <button @click="handleEdit(term)" class="edit-btn" :disabled="showAddForm || showEditForm">
                  <i class="fas fa-edit"></i>
                </button>
                <button @click="handleDelete(term.id)" class="delete-btn" :disabled="isProcessing || showAddForm || showEditForm">
                  <i class="fas fa-trash"></i>
                </button>
              </div>
            </div>
            <div class="term-content">
              <div v-if="enabledLanguages.english && term.english" class="term-text">
                <strong>English:</strong> {{ term.english }}
              </div>
              <div v-if="enabledLanguages.chinese && term.chinese" class="term-text">
                <strong>Chinese:</strong> {{ term.chinese }}
              </div>
              <div v-if="enabledLanguages.french && term.french" class="term-text">
                <strong>French:</strong> {{ term.french }}
              </div>
              <div v-if="enabledLanguages.arabic && term.arabic" class="term-text" dir="rtl">
                <strong>Arabic:</strong> {{ term.arabic }}
              </div>
            </div>
          </div>
          
          <div v-if="contractTerms.length === 0" class="empty-state">
            <p>No terms defined yet. Click "Add New Term" to get started.</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.contract-terms-form {
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;
}

.header {
  margin-bottom: 24px;
}

.header h2 {
  margin: 0 0 8px 0;
  font-size: 1.5rem;
  color: #1f2937;
}

.description {
  margin: 0;
  color: #6b7280;
  font-size: 0.9rem;
}

.content-container {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.section {
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  padding: 24px;
}

.section h3 {
  margin: 0 0 16px 0;
  font-size: 1.25rem;
  color: #1f2937;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.language-toggles {
  display: flex;
  gap: 24px;
  flex-wrap: wrap;
  margin-bottom: 16px;
}

.toggle-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 1rem;
  color: #4b5563;
}

.toggle-label input[type="checkbox"] {
  width: 18px;
  height: 18px;
  cursor: pointer;
}

.save-languages-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.875rem;
  transition: background-color 0.2s ease;
}

.save-languages-btn:hover:not(:disabled) {
  background-color: #2563eb;
}

.save-languages-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
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
  transition: background-color 0.2s ease;
}

.add-btn:hover:not(:disabled) {
  background-color: #059669;
}

.add-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.form-container {
  background-color: #f9fafb;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 24px;
}

.form-container h4 {
  margin: 0 0 16px 0;
  font-size: 1.1rem;
  color: #1f2937;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 16px;
  margin-bottom: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
}

.form-group label {
  margin-bottom: 8px;
  color: #4b5563;
  font-weight: 500;
  font-size: 0.875rem;
}

.form-group textarea {
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 0.875rem;
  font-family: inherit;
  resize: vertical;
  min-height: 80px;
}

.form-group textarea:focus {
  outline: none;
  border-color: #3b82f6;
  ring: 2px solid #3b82f6;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.cancel-btn {
  padding: 8px 16px;
  background-color: #f3f4f6;
  color: #4b5563;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.875rem;
  transition: background-color 0.2s ease;
}

.cancel-btn:hover:not(:disabled) {
  background-color: #e5e7eb;
}

.cancel-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.save-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.875rem;
  transition: background-color 0.2s ease;
}

.save-btn:hover:not(:disabled) {
  background-color: #2563eb;
}

.save-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.terms-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.term-item {
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 16px;
  background-color: #ffffff;
  transition: box-shadow 0.2s ease;
}

.term-item:hover {
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.term-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #e5e7eb;
}

.term-id {
  font-weight: 600;
  color: #6366f1;
  font-size: 0.875rem;
}

.term-actions {
  display: flex;
  gap: 8px;
}

.edit-btn,
.delete-btn {
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.875rem;
  transition: background-color 0.2s ease;
}

.edit-btn {
  background-color: #fbbf24;
  color: white;
}

.edit-btn:hover:not(:disabled) {
  background-color: #f59e0b;
}

.delete-btn {
  background-color: #ef4444;
  color: white;
}

.delete-btn:hover:not(:disabled) {
  background-color: #dc2626;
}

.edit-btn:disabled,
.delete-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.term-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.term-text {
  padding: 8px;
  background-color: #f9fafb;
  border-radius: 4px;
  font-size: 0.875rem;
  line-height: 1.6;
  color: #4b5563;
}

.term-text strong {
  color: #1f2937;
  margin-right: 8px;
}

.empty-state {
  text-align: center;
  padding: 48px;
  color: #6b7280;
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

@media (max-width: 768px) {
  .contract-terms-form {
    max-width: 100%;
  }

  .section {
    padding: 16px;
  }

  .form-grid {
    grid-template-columns: 1fr;
  }

  .section-header {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }

  .add-btn {
    width: 100%;
    justify-content: center;
  }
}
</style>

