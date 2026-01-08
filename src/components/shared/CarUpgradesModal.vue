<template>
  <teleport to="body">
    <div v-if="show" class="modal-overlay" @click.self="handleClose">
      <div class="modal-container" @click.stop>
        <div class="modal-header">
          <h3>
            <i class="fas fa-wrench"></i>
            {{ title || t('carUpgrades.title') || 'Car Upgrades' }}
          </h3>
          <div class="header-actions">
            <button 
              v-if="!showAddForm"
              @click.stop="openAddForm" 
              class="add-btn"
              type="button"
              :disabled="loading || !props.carId"
            >
              <i class="fas fa-plus"></i>
              {{ t('carUpgrades.add_upgrade') || 'Add Upgrade' }}
            </button>
            <button @click.stop="handleClose" class="close-btn" type="button">
              <i class="fas fa-times"></i>
            </button>
          </div>
        </div>
        
        <div class="modal-content">
          <!-- Add Upgrade Form -->
          <div v-if="showAddForm" class="add-upgrade-form">
            <h4>{{ t('carUpgrades.add_new_upgrade') || 'Add New Upgrade' }}</h4>
            <form @submit.prevent="saveCarUpgrade">
              <div class="form-group">
                <label for="upgrade_type">
                  {{ t('carUpgrades.upgrade_type') || 'Upgrade Type' }}
                  <span class="required">*</span>
                </label>
                <div class="input-with-button">
                  <select
                    id="upgrade_type"
                    v-model="newUpgrade.id_upgrade"
                    required
                    :disabled="isSubmitting || loadingUpgradeTypes"
                    class="form-input"
                  >
                    <option value="">{{ t('carUpgrades.select_upgrade_type') || 'Select upgrade type...' }}</option>
                    <option
                      v-for="upgradeType in availableUpgradeTypes"
                      :key="upgradeType.id"
                      :value="upgradeType.id"
                    >
                      {{ upgradeType.description || `Upgrade #${upgradeType.id}` }}
                    </option>
                  </select>
                  <button
                    type="button"
                    @click="openAddUpgradeTypeDialog"
                    class="btn-add-type"
                    :disabled="isSubmitting"
                    :title="t('carUpgrades.add_new_upgrade_type') || 'Add New Upgrade Type'"
                  >
                    <i class="fas fa-plus"></i>
                  </button>
                </div>
              </div>

              <div class="form-group">
                <label for="value">
                  {{ t('carUpgrades.value') || 'Value' }}
                  <span class="required">*</span>
                </label>
                <input
                  id="value"
                  v-model.number="newUpgrade.value"
                  type="number"
                  step="0.01"
                  min="0"
                  required
                  :disabled="isSubmitting"
                  class="form-input"
                  placeholder="0.00"
                />
              </div>

              <div class="form-group">
                <label for="date_done">
                  {{ t('carUpgrades.date_done') || 'Date Done' }}
                </label>
                <input
                  id="date_done"
                  v-model="newUpgrade.date_done"
                  type="datetime-local"
                  :disabled="isSubmitting"
                  class="form-input"
                />
              </div>

              <div class="form-group">
                <label for="user_done">
                  {{ t('carUpgrades.done_by') || 'Done By' }}
                </label>
                <select
                  id="user_done"
                  v-model.number="newUpgrade.id_uder_done"
                  :disabled="isSubmitting || loadingUsers"
                  class="form-input"
                >
                  <option :value="null">{{ t('carUpgrades.select_user') || 'Select user...' }}</option>
                  <option
                    v-for="user in availableUsers"
                    :key="user.id"
                    :value="user.id"
                  >
                    {{ user.username }}
                  </option>
                </select>
              </div>

              <div class="form-actions">
                <button
                  type="button"
                  @click="cancelAddForm"
                  class="btn-cancel"
                  :disabled="isSubmitting"
                >
                  {{ t('carUpgrades.cancel') || 'Cancel' }}
                </button>
                <button
                  type="submit"
                  class="btn-save"
                  :disabled="isSubmitting || !newUpgrade.id_upgrade || !newUpgrade.value"
                  :class="{ processing: isSubmitting }"
                >
                  <i v-if="isSubmitting" class="fas fa-spinner fa-spin"></i>
                  {{ t('carUpgrades.save') || 'Save' }}
                </button>
              </div>
            </form>
            <div v-if="submitError" class="error-message">
              <i class="fas fa-exclamation-circle"></i>
              {{ submitError }}
            </div>
          </div>

          <!-- Add Upgrade Type Dialog -->
          <div v-if="showAddUpgradeTypeDialog" class="add-type-dialog">
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
                  class="form-input"
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
                  class="form-input textarea"
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
                  class="btn-cancel"
                  :disabled="isSavingUpgradeType"
                >
                  {{ t('carUpgrades.cancel') || 'Cancel' }}
                </button>
                <button
                  type="submit"
                  class="btn-save"
                  :disabled="isSavingUpgradeType || !newUpgradeType.description?.trim()"
                  :class="{ processing: isSavingUpgradeType }"
                >
                  <i v-if="isSavingUpgradeType" class="fas fa-spinner fa-spin"></i>
                  {{ t('carUpgrades.save') || 'Save' }}
                </button>
              </div>
            </form>
          </div>

          <!-- Loading State -->
          <div v-else-if="loading" class="loading-state">
            <i class="fas fa-spinner fa-spin"></i>
            <p>{{ t('carUpgrades.loading') || 'Loading upgrades...' }}</p>
          </div>

          <!-- Error State -->
          <div v-else-if="error" class="error-state">
            <i class="fas fa-exclamation-circle"></i>
            <p>{{ error }}</p>
          </div>

          <!-- Empty State -->
          <div v-else-if="!loading && upgrades.length === 0 && !showAddForm" class="empty-state">
            <i class="fas fa-inbox"></i>
            <p>{{ t('carUpgrades.no_upgrades') || 'No upgrades available for this car' }}</p>
            <button @click="openAddForm" class="btn-add-first">
              <i class="fas fa-plus"></i>
              {{ t('carUpgrades.add_first_upgrade') || 'Add First Upgrade' }}
            </button>
          </div>

          <!-- Upgrades Table -->
          <div v-else-if="!showAddForm" class="upgrades-table-container">
            <!-- Total Amount Summary -->
            <div v-if="upgrades.length > 0" class="total-amount-summary">
              <div class="total-amount-item">
                <span class="total-label">
                  <i class="fas fa-calculator"></i>
                  {{ t('carUpgrades.total_amount') || 'Total Amount' }}:
                </span>
                <span class="total-value">{{ formatValue(totalUpgradesAmount) }}</span>
              </div>
              <div class="total-amount-item">
                <span class="total-label">
                  <i class="fas fa-list"></i>
                  {{ t('carUpgrades.total_upgrades') || 'Total Upgrades' }}:
                </span>
                <span class="total-value">{{ upgrades.length }}</span>
              </div>
            </div>

            <table class="upgrades-table">
              <thead>
                <tr>
                  <th>{{ t('carUpgrades.description') || 'Description' }}</th>
                  <th>{{ t('carUpgrades.value') || 'Value' }}</th>
                  <th>{{ t('carUpgrades.date_done') || 'Date Done' }}</th>
                  <th>{{ t('carUpgrades.user_done') || 'Done By' }}</th>
                  <th>{{ t('carUpgrades.created_by') || 'Created By' }}</th>
                  <th>{{ t('carUpgrades.time_creation') || 'Created At' }}</th>
                  <th v-if="upgrades.some(u => u.upgrade_notes)" class="notes-column">
                    {{ t('carUpgrades.notes') || 'Notes' }}
                  </th>
                  <th class="actions-column">{{ t('carUpgrades.actions') || 'Actions' }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="upgrade in upgrades" :key="upgrade.id" class="upgrade-row">
                  <td class="description-cell">
                    <div class="upgrade-description">
                      <i class="fas fa-tag"></i>
                      {{ upgrade.description || '-' }}
                    </div>
                    <div v-if="upgrade.upgrade_notes" class="upgrade-notes-text">
                      {{ upgrade.upgrade_notes }}
                    </div>
                  </td>
                  <td class="value-cell">
                    <span class="value-badge">
                      {{ formatValue(upgrade.value) }}
                    </span>
                  </td>
                  <td class="date-cell">
                    {{ formatDate(upgrade.date_done) }}
                  </td>
                  <td class="user-cell">
                    {{ getUserName(upgrade.id_uder_done, upgrade.user_done_name) }}
                  </td>
                  <td class="user-cell">
                    {{ getUserName(upgrade.id_user_create, upgrade.user_create_name) }}
                  </td>
                  <td class="date-cell">
                    {{ formatDate(upgrade.time_creation) }}
                  </td>
                  <td v-if="upgrades.some(u => u.upgrade_notes)" class="notes-cell" :title="upgrade.upgrade_notes || ''">
                    {{ truncateText(upgrade.upgrade_notes || '', 50) }}
                  </td>
                  <td class="actions-cell">
                    <div class="action-buttons">
                      <button
                        v-if="canMarkAsDone(upgrade) && !upgrade.date_done"
                        @click="markUpgradeAsDone(upgrade)"
                        class="action-btn done-btn"
                        :title="t('carUpgrades.mark_as_done') || 'Mark as Done'"
                      >
                        <i class="fas fa-check"></i>
                      </button>
                      <button
                        v-if="canEditUpgrade(upgrade)"
                        @click="startEditUpgrade(upgrade)"
                        class="action-btn edit-btn"
                        :title="t('carUpgrades.edit_upgrade') || 'Edit Upgrade'"
                      >
                        <i class="fas fa-edit"></i>
                      </button>
                      <button
                        v-if="canDeleteUpgrade(upgrade)"
                        @click="confirmDeleteUpgrade(upgrade)"
                        class="action-btn delete-btn"
                        :title="t('carUpgrades.delete_upgrade') || 'Delete Upgrade'"
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

        <!-- Edit Upgrade Dialog -->
        <div v-if="showEditDialog" class="edit-upgrade-dialog">
          <div class="dialog-header">
            <h4>
              <i class="fas fa-edit"></i>
              {{ t('carUpgrades.edit_upgrade') || 'Edit Upgrade' }}
            </h4>
            <button @click="closeEditDialog" class="close-dialog-btn" type="button">
              <i class="fas fa-times"></i>
            </button>
          </div>
          <form @submit.prevent="saveEditedUpgrade" class="dialog-content">
            <div class="form-group">
              <label for="edit_upgrade_type">
                {{ t('carUpgrades.upgrade_type') || 'Upgrade Type' }}
                <span class="required">*</span>
              </label>
              <select
                id="edit_upgrade_type"
                v-model="editingUpgrade.id_upgrade"
                required
                :disabled="isSavingEdit"
                class="form-input"
              >
                <option value="">{{ t('carUpgrades.select_upgrade_type') || 'Select upgrade type...' }}</option>
                <option
                  v-for="upgradeType in availableUpgradeTypes"
                  :key="upgradeType.id"
                  :value="upgradeType.id"
                >
                  {{ upgradeType.description || `Upgrade #${upgradeType.id}` }}
                </option>
              </select>
            </div>

            <div class="form-group">
              <label for="edit_value">
                {{ t('carUpgrades.value') || 'Value' }}
                <span class="required">*</span>
              </label>
              <input
                id="edit_value"
                v-model.number="editingUpgrade.value"
                type="number"
                step="0.01"
                min="0"
                required
                :disabled="isSavingEdit"
                class="form-input"
                placeholder="0.00"
              />
            </div>

            <div class="form-group">
              <label for="edit_date_done">
                {{ t('carUpgrades.date_done') || 'Date Done' }}
              </label>
              <input
                id="edit_date_done"
                v-model="editingUpgrade.date_done"
                type="datetime-local"
                :disabled="isSavingEdit"
                class="form-input"
              />
            </div>

            <div class="form-group">
              <label for="edit_user_done">
                {{ t('carUpgrades.done_by') || 'Done By' }}
              </label>
              <select
                id="edit_user_done"
                v-model.number="editingUpgrade.id_uder_done"
                :disabled="isSavingEdit || loadingUsers"
                class="form-input"
              >
                <option :value="null">{{ t('carUpgrades.select_user') || 'Select user...' }}</option>
                <option
                  v-for="user in availableUsers"
                  :key="user.id"
                  :value="user.id"
                >
                  {{ user.username }}
                </option>
              </select>
            </div>

            <div v-if="editError" class="error-message">
              <i class="fas fa-exclamation-circle"></i>
              {{ editError }}
            </div>

            <div class="dialog-actions">
              <button
                type="button"
                @click="closeEditDialog"
                class="btn-cancel"
                :disabled="isSavingEdit"
              >
                {{ t('carUpgrades.cancel') || 'Cancel' }}
              </button>
              <button
                type="submit"
                class="btn-save"
                :disabled="isSavingEdit || !editingUpgrade.id_upgrade || !editingUpgrade.value"
                :class="{ processing: isSavingEdit }"
              >
                <i v-if="isSavingEdit" class="fas fa-spinner fa-spin"></i>
                {{ t('carUpgrades.update') || 'Update' }}
              </button>
            </div>
          </form>
        </div>

        <!-- Delete Confirmation Dialog -->
        <div v-if="showDeleteDialog" class="delete-confirm-dialog">
          <div class="dialog-header">
            <h4>
              <i class="fas fa-exclamation-triangle"></i>
              {{ t('carUpgrades.confirm_delete') || 'Confirm Delete' }}
            </h4>
            <button @click="closeDeleteDialog" class="close-dialog-btn" type="button">
              <i class="fas fa-times"></i>
            </button>
          </div>
          <div class="dialog-content">
            <p>
              {{ t('carUpgrades.confirm_delete_upgrade_message') || 'Are you sure you want to delete this upgrade?' }}
            </p>
            <div v-if="upgradeToDelete" class="upgrade-info">
              <strong>{{ upgradeToDelete.description || '-' }}</strong>
              <span class="value-text">{{ t('carUpgrades.value') || 'Value' }}: {{ formatValue(upgradeToDelete.value) }}</span>
            </div>
            <div class="dialog-actions">
              <button
                type="button"
                @click="closeDeleteDialog"
                class="btn-cancel"
                :disabled="isDeleting"
              >
                {{ t('carUpgrades.cancel') || 'Cancel' }}
              </button>
              <button
                type="button"
                @click="deleteUpgrade"
                class="btn-delete"
                :disabled="isDeleting"
                :class="{ processing: isDeleting }"
              >
                <i v-if="isDeleting" class="fas fa-spinner fa-spin"></i>
                {{ t('carUpgrades.delete') || 'Delete' }}
              </button>
            </div>
          </div>
        </div>
        
        <div class="modal-footer">
          <button @click="handleClose" class="btn-close">
            {{ t('carUpgrades.close') || 'Close' }}
          </button>
        </div>
      </div>
    </div>
  </teleport>
</template>

<script setup>
import { ref, watch, computed, onMounted } from 'vue'
import { useApi } from '../../composables/useApi'
import { useEnhancedI18n } from '../../composables/useI18n'

const { t } = useEnhancedI18n()
const { callApi } = useApi()

const props = defineProps({
  show: {
    type: Boolean,
    default: false,
  },
  carId: {
    type: [Number, String],
    default: null,
  },
  title: {
    type: String,
    default: '',
  },
  users: {
    type: Array,
    default: () => [],
  },
})

const emit = defineEmits(['close', 'upgrade-added'])

const loading = ref(false)
const error = ref(null)
const upgrades = ref([])
const showAddForm = ref(false)
const isSubmitting = ref(false)
const submitError = ref(null)
const loadingUpgradeTypes = ref(false)
const loadingUsers = ref(false)
const availableUpgradeTypes = ref([])
const availableUsers = ref([])
const showAddUpgradeTypeDialog = ref(false)
const isSavingUpgradeType = ref(false)
const upgradeTypeError = ref(null)
const newUpgradeType = ref({
  description: '',
  notes: '',
})
const showEditDialog = ref(false)
const isSavingEdit = ref(false)
const editError = ref(null)
const editingUpgrade = ref({
  id: null,
  id_upgrade: null,
  value: null,
  date_done: null,
  id_uder_done: null,
})
const showDeleteDialog = ref(false)
const isDeleting = ref(false)
const upgradeToDelete = ref(null)

// Check if user is admin
const isAdmin = computed(() => {
  return currentUser.value?.role_id === 1
})

// Check if user can edit an upgrade (admin or creator)
const canEditUpgrade = (upgrade) => {
  if (!currentUser.value) return false
  if (isAdmin.value) return true
  return upgrade.id_user_create === currentUser.value.id
}

// Check if user can delete an upgrade (admin or creator)
const canDeleteUpgrade = (upgrade) => {
  if (!currentUser.value) return false
  if (isAdmin.value) return true
  return upgrade.id_user_create === currentUser.value.id
}

// Check if user can mark upgrade as done (admin or creator)
const canMarkAsDone = (upgrade) => {
  if (!currentUser.value) return false
  if (isAdmin.value) return true
  return upgrade.id_user_create === currentUser.value.id
}

// Calculate total amount of upgrades
const totalUpgradesAmount = computed(() => {
  if (!upgrades.value || upgrades.value.length === 0) return 0
  return upgrades.value.reduce((sum, upgrade) => {
    const value = parseFloat(upgrade.value) || 0
    return sum + value
  }, 0)
})

// Get current user
const currentUser = computed(() => {
  const userStr = localStorage.getItem('user')
  return userStr ? JSON.parse(userStr) : null
})

const newUpgrade = ref({
  id_upgrade: null,
  value: null,
  date_done: null,
  id_uder_done: null,
})

// Define resetForm before it's used in watch
const resetForm = () => {
  newUpgrade.value = {
    id_upgrade: null,
    value: null,
    date_done: null,
    id_uder_done: null,
  }
  submitError.value = null
}

// Watch for show prop changes and carId changes to fetch data
watch(
  () => [props.show, props.carId],
  ([newShow, newCarId]) => {
    if (newShow && newCarId) {
      fetchCarUpgrades()
    } else {
      upgrades.value = []
      error.value = null
      showAddForm.value = false
      resetForm()
    }
  },
  { immediate: true }
)

const fetchCarUpgrades = async () => {
  if (!props.carId) {
    error.value = 'Car ID is required'
    return
  }

  loading.value = true
  error.value = null

  try {
    const result = await callApi({
      query: `
        SELECT 
          ca.id,
          ca.id_car,
          ca.id_upgrade,
          ca.value,
          ca.date_done,
          ca.id_uder_done,
          ca.id_user_create,
          ca.time_creation,
          u.description,
          u.notes as upgrade_notes,
          ud.username as user_done_name,
          uc.username as user_create_name
        FROM car_apgrades ca
        LEFT JOIN upgrades u ON ca.id_upgrade = u.id
        LEFT JOIN users ud ON ca.id_uder_done = ud.id
        LEFT JOIN users uc ON ca.id_user_create = uc.id
        WHERE ca.id_car = ?
        ORDER BY ca.time_creation DESC, ca.date_done DESC
      `,
      params: [props.carId],
    })

    if (result.success) {
      upgrades.value = result.data || []
    } else {
      error.value = result.error || 'Failed to fetch car upgrades'
      upgrades.value = []
    }
  } catch (err) {
    console.error('Error fetching car upgrades:', err)
    error.value = err.message || 'Failed to fetch car upgrades'
    upgrades.value = []
  } finally {
    loading.value = false
  }
}

const formatValue = (value) => {
  if (value === null || value === undefined) return '-'
  // Format as number with 2 decimal places
  return parseFloat(value).toFixed(2)
}

const formatDate = (dateString) => {
  if (!dateString) return '-'
  try {
    const date = new Date(dateString)
    return date.toLocaleString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  } catch (e) {
    return dateString
  }
}

const getUserName = (userId, username) => {
  if (!userId) return '-'
  
  // First try the username from the query result
  if (username) return username
  
  // Fallback to users prop if available
  if (props.users && props.users.length > 0) {
    const user = props.users.find((u) => u.id == userId || u.id === userId)
    if (user) {
      return user.username || `User ${userId}`
    }
  }
  
  return `User ${userId}`
}

const truncateText = (text, maxLength) => {
  if (!text) return '-'
  if (text.length <= maxLength) return text
  return text.substring(0, maxLength) + '...'
}

const handleClose = () => {
  showAddForm.value = false
  resetForm()
  emit('close')
}

const openAddForm = async () => {
  showAddForm.value = true
  submitError.value = null
  await Promise.all([fetchUpgradeTypes(), fetchUsers()])
  // Don't set default date - upgrade is not done by default
  newUpgrade.value.date_done = null
  // Don't set default user - upgrade is not done by default
  newUpgrade.value.id_uder_done = null
}

const cancelAddForm = () => {
  showAddForm.value = false
  resetForm()
}

const fetchUpgradeTypes = async () => {
  loadingUpgradeTypes.value = true
  try {
    const result = await callApi({
      query: 'SELECT id, description, notes FROM upgrades ORDER BY description ASC',
    })
    if (result.success) {
      availableUpgradeTypes.value = result.data || []
    }
  } catch (err) {
    console.error('Error fetching upgrade types:', err)
  } finally {
    loadingUpgradeTypes.value = false
  }
}

const fetchUsers = async () => {
  loadingUsers.value = true
  try {
    const result = await callApi({
      query: 'SELECT id, username FROM users ORDER BY username ASC',
    })
    if (result.success) {
      availableUsers.value = result.data || []
    }
  } catch (err) {
    console.error('Error fetching users:', err)
  } finally {
    loadingUsers.value = false
  }
}

const openAddUpgradeTypeDialog = () => {
  showAddUpgradeTypeDialog.value = true
  upgradeTypeError.value = null
  newUpgradeType.value = {
    description: '',
    notes: '',
  }
}

const closeAddUpgradeTypeDialog = () => {
  showAddUpgradeTypeDialog.value = false
  upgradeTypeError.value = null
  newUpgradeType.value = {
    description: '',
    notes: '',
  }
}

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
        newUpgrade.value.id_upgrade = result.lastInsertId
      } else {
        // Fallback: find by description and id if lastInsertId is not available
        const result2 = await callApi({
          query: 'SELECT id FROM upgrades WHERE description = ? ORDER BY id DESC LIMIT 1',
          params: [description],
        })
        
        if (result2.success && result2.data && result2.data.length > 0) {
          newUpgrade.value.id_upgrade = result2.data[0].id
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
            newUpgrade.value.id_upgrade = result.lastInsertId
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

const startEditUpgrade = async (upgrade) => {
  if (!canEditUpgrade(upgrade)) {
    alert(t('carUpgrades.no_permission_to_edit') || 'You do not have permission to edit this upgrade')
    return
  }
  
  // Load upgrade types and users if not already loaded
  if (availableUpgradeTypes.value.length === 0) {
    await fetchUpgradeTypes()
  }
  if (availableUsers.value.length === 0) {
    await fetchUsers()
  }
  
  editingUpgrade.value = {
    id: upgrade.id,
    id_upgrade: upgrade.id_upgrade,
    value: upgrade.value,
    date_done: upgrade.date_done ? formatDateForInput(upgrade.date_done) : null,
    id_uder_done: upgrade.id_uder_done || null,
  }
  
  showEditDialog.value = true
  editError.value = null
}

const formatDateForInput = (dateString) => {
  if (!dateString) return null
  try {
    const date = new Date(dateString)
    // Format to datetime-local format (YYYY-MM-DDTHH:mm)
    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')
    const hours = String(date.getHours()).padStart(2, '0')
    const minutes = String(date.getMinutes()).padStart(2, '0')
    return `${year}-${month}-${day}T${hours}:${minutes}`
  } catch (e) {
    return null
  }
}

const closeEditDialog = () => {
  showEditDialog.value = false
  editError.value = null
  editingUpgrade.value = {
    id: null,
    id_upgrade: null,
    value: null,
    date_done: null,
    id_uder_done: null,
  }
}

const saveEditedUpgrade = async () => {
  if (isSavingEdit.value || !editingUpgrade.value.id) return

  isSavingEdit.value = true
  editError.value = null

  try {
    // Format date_done for database
    let dateDone = null
    if (editingUpgrade.value.date_done) {
      const date = new Date(editingUpgrade.value.date_done)
      dateDone = date.toISOString().slice(0, 19).replace('T', ' ')
    }

    const result = await callApi({
      query: `
        UPDATE car_apgrades 
        SET id_upgrade = ?, value = ?, date_done = ?, id_uder_done = ?
        WHERE id = ?
      `,
      params: [
        editingUpgrade.value.id_upgrade,
        editingUpgrade.value.value,
        dateDone,
        editingUpgrade.value.id_uder_done || null,
        editingUpgrade.value.id,
      ],
    })

    if (result.success) {
      // Refresh the upgrades list
      await fetchCarUpgrades()
      // Close the dialog
      closeEditDialog()
    } else {
      editError.value = result.error || t('carUpgrades.failed_to_update') || 'Failed to update upgrade'
    }
  } catch (err) {
    console.error('Error updating car upgrade:', err)
    editError.value = err.message || t('carUpgrades.error_updating') || 'Error updating upgrade'
  } finally {
    isSavingEdit.value = false
  }
}

const confirmDeleteUpgrade = (upgrade) => {
  if (!canDeleteUpgrade(upgrade)) {
    alert(t('carUpgrades.no_permission_to_delete') || 'You do not have permission to delete this upgrade')
    return
  }
  upgradeToDelete.value = upgrade
  showDeleteDialog.value = true
}

const closeDeleteDialog = () => {
  showDeleteDialog.value = false
  upgradeToDelete.value = null
  isDeleting.value = false
}

const deleteUpgrade = async () => {
  if (isDeleting.value || !upgradeToDelete.value) return

  isDeleting.value = true
  error.value = null

  try {
    const result = await callApi({
      query: 'DELETE FROM car_apgrades WHERE id = ?',
      params: [upgradeToDelete.value.id],
    })

    if (result.success) {
      // Refresh the upgrades list
      await fetchCarUpgrades()
      // Close the dialog
      closeDeleteDialog()
    } else {
      error.value = result.error || t('carUpgrades.failed_to_delete') || 'Failed to delete upgrade'
    }
  } catch (err) {
    console.error('Error deleting car upgrade:', err)
    error.value = err.message || t('carUpgrades.error_deleting') || 'Error deleting upgrade'
  } finally {
    isDeleting.value = false
  }
}

const markUpgradeAsDone = async (upgrade) => {
  if (!canMarkAsDone(upgrade)) {
    alert(t('carUpgrades.no_permission_to_mark_done') || 'You do not have permission to mark this upgrade as done')
    return
  }

  if (!confirm(t('carUpgrades.confirm_mark_as_done') || 'Mark this upgrade as done?')) {
    return
  }

  try {
    // Set date_done to current date/time and id_uder_done to current user
    const now = new Date()
    const dateDone = now.toISOString().slice(0, 19).replace('T', ' ')

    const result = await callApi({
      query: 'UPDATE car_apgrades SET date_done = ?, id_uder_done = ? WHERE id = ?',
      params: [dateDone, currentUser.value?.id || null, upgrade.id],
    })

    if (result.success) {
      // Refresh the upgrades list
      await fetchCarUpgrades()
    } else {
      error.value = result.error || t('carUpgrades.failed_to_mark_done') || 'Failed to mark upgrade as done'
    }
  } catch (err) {
    console.error('Error marking upgrade as done:', err)
    error.value = err.message || t('carUpgrades.error_marking_done') || 'Error marking upgrade as done'
  }
}

const saveCarUpgrade = async () => {
  if (isSubmitting.value || !props.carId) return

  isSubmitting.value = true
  submitError.value = null

  try {
    // Format date_done for database (convert to datetime or null)
    let dateDone = null
    if (newUpgrade.value.date_done) {
      // Convert datetime-local format to MySQL datetime format
      const date = new Date(newUpgrade.value.date_done)
      dateDone = date.toISOString().slice(0, 19).replace('T', ' ')
    }

    const result = await callApi({
      query: `
        INSERT INTO car_apgrades (id_car, id_upgrade, value, date_done, id_uder_done, id_user_create, time_creation)
        VALUES (?, ?, ?, ?, ?, ?, NOW())
      `,
      params: [
        props.carId,
        newUpgrade.value.id_upgrade,
        newUpgrade.value.value,
        dateDone,
        newUpgrade.value.id_uder_done || null,
        currentUser.value?.id || null,
      ],
    })

    if (result.success) {
      // Refresh the upgrades list
      await fetchCarUpgrades()
      // Reset form and close
      showAddForm.value = false
      resetForm()
      emit('upgrade-added')
    } else {
      submitError.value = result.error || t('carUpgrades.failed_to_save') || 'Failed to save upgrade'
    }
  } catch (err) {
    console.error('Error saving car upgrade:', err)
    submitError.value = err.message || t('carUpgrades.error_saving') || 'Error saving upgrade'
  } finally {
    isSubmitting.value = false
  }
}
</script>

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
  z-index: 2000;
  padding: 20px;
  backdrop-filter: blur(4px);
}

.modal-container {
  background: white;
  border-radius: 12px;
  width: 100%;
  max-width: 1200px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 25px rgba(0, 0, 0, 0.15);
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

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid #e5e7eb;
}

.header-actions {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.add-btn {
  padding: 0.5rem 1rem;
  background-color: #10b981;
  color: white;
  border: none;
  border-radius: 0.375rem;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  transition: background-color 0.2s;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.add-btn:hover:not(:disabled) {
  background-color: #059669;
}

.add-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 600;
  color: #1f2937;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.modal-header h3 i {
  color: #3b82f6;
}

.close-btn {
  background: none;
  border: none;
  color: #6b7280;
  cursor: pointer;
  padding: 0.5rem;
  border-radius: 6px;
  transition: all 0.2s;
  font-size: 1.25rem;
}

.close-btn:hover {
  background: #f3f4f6;
  color: #ef4444;
}

.modal-content {
  flex: 1;
  overflow-y: auto;
  padding: 1.5rem;
}

.loading-state,
.error-state,
.empty-state {
  text-align: center;
  padding: 3rem 2rem;
  color: #6b7280;
}

.loading-state i,
.error-state i,
.empty-state i {
  font-size: 3rem;
  margin-bottom: 1rem;
  display: block;
}

.loading-state i {
  color: #3b82f6;
}

.error-state i {
  color: #ef4444;
}

.empty-state i {
  color: #d1d5db;
}

.loading-state p,
.error-state p,
.empty-state p {
  margin: 0;
  font-size: 1rem;
}

.error-state p {
  color: #ef4444;
}

.total-amount-summary {
  display: flex;
  gap: 2rem;
  padding: 1rem;
  background-color: #f0f9ff;
  border: 1px solid #bae6fd;
  border-radius: 8px;
  margin-bottom: 1rem;
  flex-wrap: wrap;
}

.total-amount-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.total-label {
  font-weight: 600;
  color: #1e40af;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.total-label i {
  color: #3b82f6;
}

.total-value {
  font-weight: 700;
  font-size: 1.125rem;
  color: #1e40af;
  font-family: monospace;
}

.upgrades-table-container {
  overflow-x: auto;
}

.upgrades-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
}

.upgrades-table thead {
  background-color: #f9fafb;
  border-bottom: 2px solid #e5e7eb;
}

.upgrades-table th {
  padding: 0.75rem 1rem;
  text-align: left;
  font-weight: 600;
  color: #374151;
  white-space: nowrap;
}

.upgrades-table th.notes-column {
  min-width: 200px;
}

.upgrades-table tbody tr {
  border-bottom: 1px solid #e5e7eb;
  transition: background-color 0.2s;
}

.upgrades-table tbody tr:hover {
  background-color: #f9fafb;
}

.upgrades-table td {
  padding: 0.75rem 1rem;
  color: #1f2937;
  vertical-align: top;
}

.description-cell {
  min-width: 200px;
}

.upgrade-description {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 500;
  color: #1f2937;
  margin-bottom: 0.25rem;
}

.upgrade-description i {
  color: #3b82f6;
  font-size: 0.75rem;
}

.upgrade-notes-text {
  font-size: 0.75rem;
  color: #6b7280;
  margin-top: 0.25rem;
  font-style: italic;
}

.value-cell {
  text-align: right;
}

.value-badge {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  background-color: #eff6ff;
  color: #1e40af;
  border-radius: 0.375rem;
  font-weight: 600;
  font-family: 'Courier New', monospace;
}

.date-cell {
  white-space: nowrap;
  color: #6b7280;
  font-size: 0.8125rem;
}

.user-cell {
  color: #374151;
}

.notes-cell {
  max-width: 250px;
  color: #6b7280;
  font-size: 0.8125rem;
}

.actions-column {
  width: 100px;
  text-align: center;
}

.actions-cell {
  text-align: center;
  white-space: nowrap;
}

.action-buttons {
  display: flex;
  gap: 0.5rem;
  justify-content: center;
  align-items: center;
}

.action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  padding: 0;
  border: none;
  border-radius: 0.375rem;
  cursor: pointer;
  font-size: 0.875rem;
  transition: all 0.2s;
}

.edit-btn {
  background-color: #fbbf24;
  color: white;
}

.edit-btn:hover {
  background-color: #f59e0b;
}

.delete-btn {
  background-color: #ef4444;
  color: white;
}

.delete-btn:hover {
  background-color: #dc2626;
}

.done-btn {
  background-color: #10b981;
  color: white;
}

.done-btn:hover {
  background-color: #059669;
}

.modal-footer {
  padding: 1rem 1.5rem;
  border-top: 1px solid #e5e7eb;
  display: flex;
  justify-content: flex-end;
}

.btn-close {
  padding: 0.625rem 1.25rem;
  background-color: #6b7280;
  color: white;
  border: none;
  border-radius: 0.375rem;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  transition: background-color 0.2s;
}

.btn-close:hover {
  background-color: #4b5563;
}

/* Add Upgrade Form Styles */
.add-upgrade-form {
  padding: 1.5rem;
  background: #f9fafb;
  border-radius: 0.5rem;
  margin-bottom: 1.5rem;
  position: relative;
}

.add-upgrade-form h4 {
  margin: 0 0 1.5rem 0;
  font-size: 1.125rem;
  font-weight: 600;
  color: #1f2937;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.add-upgrade-form h4 i {
  color: #10b981;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
  color: #374151;
  font-size: 0.875rem;
}

.required {
  color: #ef4444;
}

.form-input {
  width: 100%;
  padding: 0.625rem 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  font-size: 0.875rem;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.form-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-input:disabled {
  background-color: #f3f4f6;
  color: #9ca3af;
  cursor: not-allowed;
}

.form-actions {
  display: flex;
  gap: 0.75rem;
  justify-content: flex-end;
  margin-top: 1.5rem;
}

.btn-cancel,
.btn-save {
  padding: 0.625rem 1.25rem;
  border: none;
  border-radius: 0.375rem;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.btn-cancel {
  background-color: #f3f4f6;
  color: #374151;
}

.btn-cancel:hover:not(:disabled) {
  background-color: #e5e7eb;
}

.btn-save {
  background-color: #3b82f6;
  color: white;
}

.btn-save:hover:not(:disabled) {
  background-color: #2563eb;
}

.btn-save:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-save.processing {
  opacity: 0.7;
}

.error-message {
  margin-top: 1rem;
  padding: 0.75rem;
  background-color: #fee2e2;
  color: #dc2626;
  border-radius: 0.375rem;
  font-size: 0.875rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.error-message i {
  color: #dc2626;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

.btn-add-first {
  padding: 0.625rem 1.25rem;
  background-color: #10b981;
  color: white;
  border: none;
  border-radius: 0.375rem;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  transition: background-color 0.2s;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-top: 1rem;
}

.btn-add-first:hover {
  background-color: #059669;
}

/* Input with button */
.input-with-button {
  display: flex;
  gap: 0.5rem;
  align-items: stretch;
}

.input-with-button .form-input {
  flex: 1;
}

.btn-add-type {
  padding: 0.625rem 1rem;
  background-color: #10b981;
  color: white;
  border: none;
  border-radius: 0.375rem;
  cursor: pointer;
  font-size: 0.875rem;
  transition: background-color 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 42px;
}

.btn-add-type:hover:not(:disabled) {
  background-color: #059669;
}

.btn-add-type:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Add/Edit Upgrade Type Dialog and Edit/Delete Dialogs */
.add-type-dialog,
.edit-upgrade-dialog,
.delete-confirm-dialog {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  background: white;
  border-radius: 0.5rem;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.3);
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  z-index: 10;
  display: flex;
  flex-direction: column;
  border: 2px solid #3b82f6;
}

.edit-upgrade-dialog {
  border-color: #fbbf24;
}

.delete-confirm-dialog {
  border-color: #ef4444;
  max-width: 400px;
}

.add-type-dialog .dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid #e5e7eb;
  background-color: #f9fafb;
  border-radius: 0.5rem 0.5rem 0 0;
}

.add-type-dialog .dialog-header h4 {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  color: #1f2937;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.add-type-dialog .dialog-header h4 i {
  color: #10b981;
}

.close-dialog-btn {
  background: none;
  border: none;
  color: #6b7280;
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 0.25rem;
  transition: all 0.2s;
  font-size: 1.125rem;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
}

.close-dialog-btn:hover {
  background-color: #f3f4f6;
  color: #ef4444;
}

.add-type-dialog .dialog-content {
  padding: 1.5rem;
  overflow-y: auto;
}

.dialog-actions {
  display: flex;
  gap: 0.75rem;
  justify-content: flex-end;
  margin-top: 1.5rem;
}

.form-input.textarea {
  resize: vertical;
  min-height: 70px;
  font-family: inherit;
}

.delete-confirm-dialog .dialog-content {
  padding: 1.5rem;
}

.delete-confirm-dialog .dialog-content p {
  margin: 0 0 1rem 0;
  color: #374151;
}

.upgrade-info {
  background-color: #f9fafb;
  padding: 0.75rem;
  border-radius: 0.375rem;
  margin-bottom: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.upgrade-info strong {
  color: #1f2937;
  font-size: 0.875rem;
}

.value-text {
  color: #6b7280;
  font-size: 0.8125rem;
}

.btn-delete {
  padding: 0.625rem 1.25rem;
  background-color: #ef4444;
  color: white;
  border: none;
  border-radius: 0.375rem;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.btn-delete:hover:not(:disabled) {
  background-color: #dc2626;
}

.btn-delete:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-delete.processing {
  opacity: 0.7;
}

/* Responsive adjustments */
@media (max-width: 768px) {
  .modal-container {
    max-width: 95vw;
  }

  .upgrades-table {
    font-size: 0.75rem;
  }

  .upgrades-table th,
  .upgrades-table td {
    padding: 0.5rem;
  }
}
</style>

