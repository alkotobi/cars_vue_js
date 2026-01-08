<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useApi } from '../../composables/useApi'
import { useEnhancedI18n } from '@/composables/useI18n'

const { t } = useEnhancedI18n()
const { callApi } = useApi()

const props = defineProps({
  selectedCars: {
    type: Array,
    required: true,
  },
  show: {
    type: Boolean,
    default: false,
  },
  isAdmin: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['close', 'save'])

const loading = ref(false)
const error = ref(null)
const isProcessing = ref(false)

// Form data
const selectedUpgradeType = ref(null)
const upgradeValue = ref('')
const availableUpgradeTypes = ref([])
const availableUsers = ref([])
const selectedUserDone = ref(null)

// Add Upgrade Type Dialog
const showAddUpgradeTypeDialog = ref(false)
const isSavingUpgradeType = ref(false)
const upgradeTypeError = ref(null)
const newUpgradeType = ref({
  description: '',
  notes: '',
})

// Get current user
const currentUser = ref(JSON.parse(localStorage.getItem('user') || 'null'))

// Computed
const selectedCarsCount = computed(() => props.selectedCars.length)

// Fetch upgrade types
const fetchUpgradeTypes = async () => {
  try {
    const result = await callApi({
      query: 'SELECT id, description FROM upgrades ORDER BY description',
      params: [],
    })
    if (result.success) {
      availableUpgradeTypes.value = result.data || []
    }
  } catch (err) {
    console.error('Error fetching upgrade types:', err)
    error.value = t('carUpgrades.failed_to_load_upgrade_types') || 'Failed to load upgrade types'
  }
}

// Fetch users
const fetchUsers = async () => {
  try {
    const result = await callApi({
      query: 'SELECT id, username FROM users ORDER BY username',
      params: [],
    })
    if (result.success) {
      availableUsers.value = result.data || []
    }
  } catch (err) {
    console.error('Error fetching users:', err)
  }
}

// Reset form
const resetForm = () => {
  selectedUpgradeType.value = null
  upgradeValue.value = ''
  selectedUserDone.value = null
  error.value = null
  showAddUpgradeTypeDialog.value = false
  upgradeTypeError.value = null
  newUpgradeType.value = {
    description: '',
    notes: '',
  }
}

// Open Add Upgrade Type Dialog
const openAddUpgradeTypeDialog = () => {
  showAddUpgradeTypeDialog.value = true
  upgradeTypeError.value = null
  newUpgradeType.value = {
    description: '',
    notes: '',
  }
}

// Close Add Upgrade Type Dialog
const closeAddUpgradeTypeDialog = () => {
  showAddUpgradeTypeDialog.value = false
  upgradeTypeError.value = null
  newUpgradeType.value = {
    description: '',
    notes: '',
  }
}

// Save Upgrade Type
const saveUpgradeType = async () => {
  if (isSavingUpgradeType.value || !newUpgradeType.value.description?.trim()) return

  isSavingUpgradeType.value = true
  upgradeTypeError.value = null

  try {
    const description = newUpgradeType.value.description.trim()
    const notes = newUpgradeType.value.notes.trim() || null

    // First, try with id_user_owner column (newer schema)
    let result = await callApi({
      query: 'INSERT INTO upgrades (description, notes, id_user_owner) VALUES (?, ?, ?)',
      params: [description, notes, currentUser.value?.id || null],
    })

    // If that fails with column not found error, try without id_user_owner (older schema)
    if (!result.success && result.error?.includes('id_user_owner')) {
      result = await callApi({
        query: 'INSERT INTO upgrades (description, notes) VALUES (?, ?)',
        params: [description, notes],
      })
    }

    if (result.success) {
      // Refresh the upgrade types list
      await fetchUpgradeTypes()
      
      // Select the newly created upgrade type
      if (result.lastInsertId) {
        selectedUpgradeType.value = result.lastInsertId
      } else {
        // Fallback: find by description and id if lastInsertId is not available
        const result2 = await callApi({
          query: 'SELECT id FROM upgrades WHERE description = ? ORDER BY id DESC LIMIT 1',
          params: [description],
        })
        
        if (result2.success && result2.data && result2.data.length > 0) {
          selectedUpgradeType.value = result2.data[0].id
        }
      }
      
      // Close the dialog
      closeAddUpgradeTypeDialog()
    } else {
      upgradeTypeError.value = result.error || t('carUpgrades.failed_to_save_upgrade_type') || 'Failed to save upgrade type'
    }
  } catch (err) {
    console.error('Error saving upgrade type:', err)
    // Try fallback without id_user_owner if error mentions the column
    if (err.message?.includes('id_user_owner')) {
      try {
        const description = newUpgradeType.value.description.trim()
        const notes = newUpgradeType.value.notes.trim() || null
        const result = await callApi({
          query: 'INSERT INTO upgrades (description, notes) VALUES (?, ?)',
          params: [description, notes],
        })
        
        if (result.success) {
          await fetchUpgradeTypes()
          if (result.lastInsertId) {
            selectedUpgradeType.value = result.lastInsertId
          }
          closeAddUpgradeTypeDialog()
          return
        }
      } catch (fallbackErr) {
        upgradeTypeError.value = fallbackErr.message || t('carUpgrades.error_saving_upgrade_type') || 'Error saving upgrade type'
      }
    } else {
      upgradeTypeError.value = err.message || t('carUpgrades.error_saving_upgrade_type') || 'Error saving upgrade type'
    }
  } finally {
    isSavingUpgradeType.value = false
  }
}

// Handle submit
const handleSubmit = async () => {
  if (isProcessing.value || loading.value) return

  // Validation
  if (!selectedUpgradeType.value) {
    error.value = t('carUpgrades.please_select_upgrade_type') || 'Please select an upgrade type'
    return
  }

  if (!upgradeValue.value || parseFloat(upgradeValue.value) <= 0) {
    error.value = t('carUpgrades.please_enter_valid_value') || 'Please enter a valid value'
    return
  }

  isProcessing.value = true
  error.value = null

  try {
    const value = parseFloat(upgradeValue.value)
    const now = new Date()
    const timeCreation = now.toISOString().slice(0, 19).replace('T', ' ')

    // Insert upgrade for each selected car
    const insertPromises = props.selectedCars.map(async (car) => {
      try {
        // Try with id_user_owner first (if column exists)
        const result = await callApi({
          query: `INSERT INTO car_apgrades (id_car, id_upgrade, value, date_done, id_uder_done, id_user_create, time_creation) 
                   VALUES (?, ?, ?, NULL, ?, ?, ?)`,
          params: [
            car.id,
            selectedUpgradeType.value,
            value,
            selectedUserDone.value || null,
            currentUser.value?.id || null,
            timeCreation,
          ],
        })

        if (!result.success) {
          // If error mentions id_user_owner, retry without it
          if (result.error && result.error.includes('id_user_owner')) {
            return await callApi({
              query: `INSERT INTO car_apgrades (id_car, id_upgrade, value, date_done, id_uder_done, id_user_create, time_creation) 
                       VALUES (?, ?, ?, NULL, ?, ?, ?)`,
              params: [
                car.id,
                selectedUpgradeType.value,
                value,
                selectedUserDone.value || null,
                currentUser.value?.id || null,
                timeCreation,
              ],
            })
          }
          return result
        }
        return result
      } catch (err) {
        console.error(`Error adding upgrade for car ${car.id}:`, err)
        return { success: false, error: err.message }
      }
    })

    const results = await Promise.all(insertPromises)
    const failed = results.filter((r) => !r.success)

    if (failed.length > 0) {
      error.value =
        t('carUpgrades.failed_to_add_upgrades', { count: failed.length }) ||
        `Failed to add upgrades for ${failed.length} car(s)`
    } else {
      emit('save', props.selectedCars)
      resetForm()
      emit('close')
    }
  } catch (err) {
    console.error('Error adding bulk upgrades:', err)
    error.value = err.message || t('carUpgrades.error_adding_upgrades') || 'Error adding upgrades'
  } finally {
    isProcessing.value = false
  }
}

// Close modal
const closeModal = () => {
  if (isProcessing.value) return
  resetForm()
  emit('close')
}

// Watch for modal opening
watch(
  () => props.show,
  (isOpen) => {
    if (isOpen) {
      resetForm()
      fetchUpgradeTypes()
      fetchUsers()
    }
  }
)

onMounted(() => {
  if (props.show) {
    fetchUpgradeTypes()
    fetchUsers()
  }
})
</script>

<template>
  <Teleport to="body">
    <div v-if="show" class="modal-overlay" @click="closeModal">
      <div class="modal-content" @click.stop :class="{ 'is-processing': isProcessing }">
        <div class="modal-header">
          <h3>
            <i class="fas fa-wrench"></i>
            {{ t('carUpgradesBulkAdd.add_upgrade_to_cars', { count: selectedCarsCount }) || `Add Upgrade to ${selectedCarsCount} Car(s)` }}
          </h3>
          <button class="close-btn" @click="closeModal" :disabled="isProcessing">
            <i class="fas fa-times"></i>
          </button>
        </div>

        <div class="modal-body">
          <div class="info-section">
            <p>
              <strong>{{ t('carUpgradesBulkAdd.selected_cars') || 'Selected Cars' }}:</strong>
              {{ selectedCarsCount }}
            </p>
          </div>

          <div v-if="error" class="error-message">
            <i class="fas fa-exclamation-circle"></i>
            {{ error }}
          </div>

          <div class="form-group">
            <label for="upgrade-type">
              {{ t('carUpgrades.upgrade_type') || 'Upgrade Type' }}
              <span class="required">*</span>
            </label>
            <div class="input-with-button">
              <select
                id="upgrade-type"
                v-model="selectedUpgradeType"
                :disabled="isProcessing || isSavingUpgradeType"
                class="form-control"
                required
              >
                <option :value="null">
                  {{ t('carUpgrades.select_upgrade_type') || 'Select Upgrade Type' }}
                </option>
                <option v-for="upgrade in availableUpgradeTypes" :key="upgrade.id" :value="upgrade.id">
                  {{ upgrade.description }}
                </option>
              </select>
              <button
                type="button"
                @click="openAddUpgradeTypeDialog"
                class="btn-add-type"
                :disabled="isProcessing || isSavingUpgradeType"
                :title="t('carUpgrades.add_new_upgrade_type') || 'Add New Upgrade Type'"
              >
                <i class="fas fa-plus"></i>
              </button>
            </div>
          </div>

          <div class="form-group">
            <label for="upgrade-value">
              {{ t('carUpgrades.value') || 'Value' }}
              <span class="required">*</span>
            </label>
            <input
              id="upgrade-value"
              v-model="upgradeValue"
              type="number"
              step="0.01"
              min="0"
              :disabled="isProcessing"
              class="form-control"
              :placeholder="t('carUpgrades.enter_value') || 'Enter value'"
            />
          </div>

          <div class="form-group">
            <label for="user-done">
              {{ t('carUpgrades.done_by') || 'Done By' }}
            </label>
            <select
              id="user-done"
              v-model="selectedUserDone"
              :disabled="isProcessing"
              class="form-control"
            >
              <option :value="null">
                {{ t('carUpgrades.select_user') || 'Select User (Optional)' }}
              </option>
              <option v-for="user in availableUsers" :key="user.id" :value="user.id">
                {{ user.username }}
              </option>
            </select>
            <small class="form-help">
              {{ t('carUpgradesBulkAdd.upgrade_will_not_be_done') || 'Upgrade will be created as not done by default' }}
            </small>
          </div>
        </div>

        <div class="modal-footer">
          <button @click="closeModal" class="btn btn-cancel" :disabled="isProcessing">
            {{ t('carUpgrades.cancel') || 'Cancel' }}
          </button>
          <button @click="handleSubmit" class="btn btn-primary" :disabled="isProcessing">
            <i v-if="isProcessing" class="fas fa-spinner fa-spin"></i>
            <i v-else class="fas fa-save"></i>
            {{ isProcessing ? (t('carUpgradesBulkAdd.adding') || 'Adding...') : (t('carUpgradesBulkAdd.add_to_all_cars') || 'Add to All Cars') }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>

  <!-- Add Upgrade Type Dialog Overlay -->
  <Teleport to="body">
    <div v-if="showAddUpgradeTypeDialog" class="dialog-overlay" @click="closeAddUpgradeTypeDialog">
      <div class="add-type-dialog" @click.stop>
        <div class="dialog-header">
          <h4>
            <i class="fas fa-plus-circle"></i>
            {{ t('carUpgrades.add_new_upgrade_type') || 'Add New Upgrade Type' }}
          </h4>
          <button @click="closeAddUpgradeTypeDialog" class="close-dialog-btn" type="button">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <form @submit.prevent="saveUpgradeType" class="dialog-content">
          <div class="form-group">
            <label for="new_upgrade_description">
              {{ t('carUpgrades.description') || 'Description' }}
              <span class="required">*</span>
            </label>
            <input
              id="new_upgrade_description"
              v-model="newUpgradeType.description"
              type="text"
              required
              maxlength="255"
              :disabled="isSavingUpgradeType"
              class="form-control"
              :placeholder="t('carUpgrades.enter_description') || 'Enter upgrade description'"
            />
          </div>

          <div class="form-group">
            <label for="new_upgrade_notes">
              {{ t('carUpgrades.notes') || 'Notes' }}
            </label>
            <textarea
              id="new_upgrade_notes"
              v-model="newUpgradeType.notes"
              :disabled="isSavingUpgradeType"
              rows="3"
              class="form-control textarea"
              :placeholder="t('carUpgrades.enter_notes') || 'Enter notes (optional)'"
            ></textarea>
          </div>

          <div v-if="upgradeTypeError" class="error-message">
            <i class="fas fa-exclamation-circle"></i>
            {{ upgradeTypeError }}
          </div>

          <div class="dialog-actions">
            <button
              type="button"
              @click="closeAddUpgradeTypeDialog"
              class="btn btn-cancel"
              :disabled="isSavingUpgradeType"
            >
              {{ t('carUpgrades.cancel') || 'Cancel' }}
            </button>
            <button
              type="submit"
              class="btn btn-primary"
              :disabled="isSavingUpgradeType || !newUpgradeType.description?.trim()"
            >
              <i v-if="isSavingUpgradeType" class="fas fa-spinner fa-spin"></i>
              {{ t('carUpgrades.save') || 'Save' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </Teleport>
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
  align-items: center;
  justify-content: center;
  z-index: 10000;
  backdrop-filter: blur(4px);
}

.modal-content {
  background: white;
  border-radius: 12px;
  width: 90%;
  max-width: 600px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  display: flex;
  flex-direction: column;
}

.modal-content.is-processing {
  opacity: 0.7;
  pointer-events: none;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #e5e7eb;
}

.modal-header h3 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #1f2937;
  display: flex;
  align-items: center;
  gap: 10px;
}

.modal-header h3 i {
  color: #f59e0b;
}

.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  color: #6b7280;
  cursor: pointer;
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  transition: all 0.2s;
}

.close-btn:hover:not(:disabled) {
  background-color: #f3f4f6;
  color: #1f2937;
}

.close-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.modal-body {
  padding: 24px;
  flex: 1;
  overflow-y: auto;
}

.info-section {
  background-color: #f9fafb;
  padding: 16px;
  border-radius: 8px;
  margin-bottom: 24px;
}

.info-section p {
  margin: 0;
  color: #374151;
  font-size: 14px;
}

.error-message {
  background-color: #fef2f2;
  border: 1px solid #fecaca;
  color: #dc2626;
  padding: 12px 16px;
  border-radius: 8px;
  margin-bottom: 20px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #374151;
  font-size: 14px;
}

.form-group label .required {
  color: #ef4444;
  margin-left: 4px;
}

.form-control {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  transition: all 0.2s;
}

.form-control:focus {
  outline: none;
  border-color: #f59e0b;
  box-shadow: 0 0 0 3px rgba(245, 158, 11, 0.1);
}

.form-control:disabled {
  background-color: #f3f4f6;
  cursor: not-allowed;
}

.form-help {
  display: block;
  margin-top: 6px;
  font-size: 12px;
  color: #6b7280;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 20px 24px;
  border-top: 1px solid #e5e7eb;
  background-color: #f9fafb;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-cancel {
  background-color: #f3f4f6;
  color: #374151;
}

.btn-cancel:hover:not(:disabled) {
  background-color: #e5e7eb;
}

.btn-primary {
  background-color: #f59e0b;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background-color: #d97706;
}

.input-with-button {
  display: flex;
  gap: 8px;
  align-items: stretch;
}

.input-with-button .form-control {
  flex: 1;
}

.btn-add-type {
  padding: 10px 16px;
  background-color: #10b981;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 44px;
}

.btn-add-type:hover:not(:disabled) {
  background-color: #059669;
}

.btn-add-type:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
}

.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10001;
  backdrop-filter: blur(4px);
}

.add-type-dialog {
  background: white;
  border-radius: 12px;
  width: 90%;
  max-width: 500px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  max-height: 90vh;
  overflow-y: auto;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #e5e7eb;
}

.dialog-header h4 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1f2937;
  display: flex;
  align-items: center;
  gap: 10px;
}

.dialog-header h4 i {
  color: #10b981;
}

.close-dialog-btn {
  background: none;
  border: none;
  font-size: 20px;
  color: #6b7280;
  cursor: pointer;
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  transition: all 0.2s;
}

.close-dialog-btn:hover:not(:disabled) {
  background-color: #f3f4f6;
  color: #1f2937;
}

.dialog-content {
  padding: 24px;
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #e5e7eb;
}

.textarea {
  resize: vertical;
  min-height: 80px;
}
</style>

