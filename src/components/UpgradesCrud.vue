<template>
  <div class="upgrades-container">
    <div class="table-header">
      <h3>
        <i class="fas fa-wrench"></i>
        {{ t('upgrades.title') || 'Upgrades' }}
      </h3>
      <div class="header-actions">
        <button
          v-if="canCreate"
          @click="openAddDialog"
          class="add-btn"
          :disabled="loading"
        >
          <i class="fas fa-plus"></i>
          <span>{{ t('upgrades.add_upgrade') || 'Add Upgrade' }}</span>
        </button>
        <button
          @click="refreshData"
          class="refresh-btn"
          :disabled="loading"
          :class="{ processing: loading }"
        >
          <i class="fas fa-sync-alt"></i>
          <span>{{ t('upgrades.refresh') || 'Refresh' }}</span>
          <i v-if="loading" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>
      </div>
    </div>

    <div class="table-wrapper">
      <div v-if="loading" class="loading-overlay">
        <i class="fas fa-spinner fa-spin fa-2x"></i>
        <span>{{ t('upgrades.loading') || 'Loading upgrades...' }}</span>
      </div>

      <div v-else-if="error" class="error-message">
        <i class="fas fa-exclamation-circle"></i>
        {{ error }}
      </div>

      <div v-else-if="upgrades.length === 0" class="empty-state">
        <i class="fas fa-wrench fa-2x"></i>
        <p>{{ t('upgrades.no_upgrades_found') || 'No upgrades found' }}</p>
        <button
          v-if="canCreate"
          @click="openAddDialog"
          class="add-first-btn"
        >
          <i class="fas fa-plus"></i>
          {{ t('upgrades.add_first_upgrade') || 'Add First Upgrade' }}
        </button>
      </div>

      <div v-else class="table-content">
        <table class="upgrades-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>{{ t('upgrades.description') || 'Description' }}</th>
              <th>{{ t('upgrades.notes') || 'Notes' }}</th>
              <th>{{ t('upgrades.created_by') || 'Created By' }}</th>
              <th>{{ t('upgrades.actions') || 'Actions' }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="upgrade in upgrades"
              :key="upgrade.id"
              class="table-row"
              :class="{ 'selected-row': selectedUpgradeId === upgrade.id }"
            >
              <td class="id-cell">#{{ upgrade.id }}</td>
              <td class="description-cell">{{ upgrade.description || '-' }}</td>
              <td class="notes-cell" :title="upgrade.notes || ''">
                {{ truncateText(upgrade.notes || '', 50) }}
              </td>
              <td class="owner-cell">
                {{ upgrade.owner_name || '-' }}
              </td>
              <td class="actions-cell">
                <div class="action-buttons">
                  <button
                    v-if="canEdit(upgrade)"
                    @click.stop="editUpgrade(upgrade)"
                    class="action-btn edit-btn"
                    :title="t('upgrades.edit_upgrade') || 'Edit Upgrade'"
                  >
                    <i class="fas fa-edit"></i>
                  </button>
                  <button
                    v-if="canDelete(upgrade)"
                    @click.stop="deleteUpgrade(upgrade)"
                    class="action-btn delete-btn"
                    :title="t('upgrades.delete_upgrade') || 'Delete Upgrade'"
                    :disabled="isUpgradeInUse(upgrade.id)"
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

    <div class="table-footer">
      <span class="record-count">
        {{ t('upgrades.showing_count', { count: upgrades.length }) || `Showing ${upgrades.length} upgrade(s)` }}
      </span>
    </div>

    <!-- Add/Edit Dialog -->
    <div v-if="showDialog" class="dialog-overlay" @click.self="closeDialog">
      <div class="dialog">
        <div class="dialog-header">
          <h3>
            <i class="fas fa-wrench"></i>
            {{ isEditing ? (t('upgrades.edit_upgrade') || 'Edit Upgrade') : (t('upgrades.add_upgrade') || 'Add New Upgrade') }}
          </h3>
          <button class="close-btn" @click="closeDialog" :disabled="isSubmitting">
            <i class="fas fa-times"></i>
          </button>
        </div>

        <form @submit.prevent="saveUpgrade" class="dialog-content">
          <div class="form-group">
            <label for="description">
              <i class="fas fa-tag"></i>
              {{ t('upgrades.description') || 'Description' }}
              <span class="required">*</span>
            </label>
            <input
              type="text"
              id="description"
              v-model="formData.description"
              required
              :disabled="isSubmitting"
              maxlength="255"
              placeholder="Enter upgrade description"
            />
          </div>

          <div class="form-group">
            <label for="notes">
              <i class="fas fa-sticky-note"></i>
              {{ t('upgrades.notes') || 'Notes' }}
            </label>
            <textarea
              id="notes"
              v-model="formData.notes"
              :disabled="isSubmitting"
              rows="4"
              placeholder="Enter notes (optional)"
              class="textarea-field"
            ></textarea>
          </div>

          <div class="form-actions">
            <button type="button" @click="closeDialog" class="cancel-btn" :disabled="isSubmitting">
              {{ t('upgrades.cancel') || 'Cancel' }}
            </button>
            <button
              type="submit"
              class="save-btn"
              :disabled="isSubmitting || !formData.description.trim()"
              :class="{ processing: isSubmitting }"
            >
              <i v-if="isSubmitting" class="fas fa-spinner fa-spin"></i>
              <span>{{ isEditing ? (t('upgrades.update') || 'Update') : (t('upgrades.create') || 'Create') }}</span>
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Delete Confirmation Dialog -->
    <div v-if="showDeleteDialog" class="dialog-overlay" @click.self="closeDeleteDialog">
      <div class="dialog delete-dialog">
        <div class="dialog-header">
          <h3>
            <i class="fas fa-exclamation-triangle"></i>
            {{ t('upgrades.confirm_delete') || 'Confirm Delete' }}
          </h3>
          <button class="close-btn" @click="closeDeleteDialog" :disabled="isDeleting">
            <i class="fas fa-times"></i>
          </button>
        </div>

        <div class="dialog-content">
          <div class="delete-message">
            <p>
              {{ t('upgrades.confirm_delete_message', { description: upgradeToDelete?.description }) || `Are you sure you want to delete upgrade "${upgradeToDelete?.description}"?` }}
            </p>
            <p v-if="isUpgradeInUse(upgradeToDelete?.id)" class="warning-text">
              <i class="fas fa-exclamation-triangle"></i>
              {{ t('upgrades.cannot_delete_in_use') || 'This upgrade is being used by one or more cars and cannot be deleted.' }}
            </p>
          </div>

          <div class="form-actions">
            <button
              type="button"
              @click="closeDeleteDialog"
              class="cancel-btn"
              :disabled="isDeleting"
            >
              {{ t('upgrades.cancel') || 'Cancel' }}
            </button>
            <button
              type="button"
              @click="confirmDelete"
              class="delete-confirm-btn"
              :disabled="isDeleting || isUpgradeInUse(upgradeToDelete?.id)"
              :class="{ processing: isDeleting }"
            >
              <i v-if="isDeleting" class="fas fa-spinner fa-spin"></i>
              <span>{{ t('upgrades.delete') || 'Delete' }}</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useApi } from '@/composables/useApi'
import { useEnhancedI18n } from '@/composables/useI18n'

const { t } = useEnhancedI18n()
const { callApi } = useApi()

const upgrades = ref([])
const loading = ref(false)
const error = ref(null)
const showDialog = ref(false)
const showDeleteDialog = ref(false)
const isEditing = ref(false)
const isSubmitting = ref(false)
const isDeleting = ref(false)
const upgradeToDelete = ref(null)
const selectedUpgradeId = ref(null)
const upgradeUsageCount = ref({}) // Cache for upgrade usage counts

const formData = ref({
  description: '',
  notes: '',
})

// Get current user
const currentUser = computed(() => {
  const userStr = localStorage.getItem('user')
  return userStr ? JSON.parse(userStr) : null
})

// Check if user is admin
const isAdmin = computed(() => {
  return currentUser.value?.role_id === 1
})

// Check if user can create upgrades (everyone can create)
const canCreate = computed(() => {
  return !!currentUser.value
})

// Check if user can edit an upgrade (admin or owner)
const canEdit = (upgrade) => {
  if (!currentUser.value) return false
  if (isAdmin.value) return true
  return upgrade.id_user_owner === currentUser.value.id
}

// Check if user can delete an upgrade (admin or owner)
const canDelete = (upgrade) => {
  if (!currentUser.value) return false
  if (isAdmin.value) return true
  return upgrade.id_user_owner === currentUser.value.id
}

// Check if upgrade is being used by any cars
const isUpgradeInUse = (upgradeId) => {
  return upgradeUsageCount.value[upgradeId] > 0
}

const fetchUpgrades = async () => {
  loading.value = true
  error.value = null

  try {
    const result = await callApi({
      query: `
        SELECT 
          u.id,
          u.description,
          u.notes,
          u.id_user_owner,
          owner.username as owner_name
        FROM upgrades u
        LEFT JOIN users owner ON u.id_user_owner = owner.id
        ORDER BY u.id DESC
      `,
    })

    if (result.success) {
      upgrades.value = result.data || []
      
      // Fetch usage counts for all upgrades
      if (upgrades.value.length > 0) {
        const upgradeIds = upgrades.value.map(u => u.id)
        const usageResult = await callApi({
          query: `
            SELECT id_upgrade, COUNT(*) as count
            FROM car_apgrades
            WHERE id_upgrade IN (${upgradeIds.map(() => '?').join(',')})
            GROUP BY id_upgrade
          `,
          params: upgradeIds,
        })
        
        if (usageResult.success) {
          upgradeUsageCount.value = {}
          usageResult.data.forEach(row => {
            upgradeUsageCount.value[row.id_upgrade] = row.count
          })
        }
      }
    } else {
      error.value = result.error || t('upgrades.failed_to_load') || 'Failed to load upgrades'
    }
  } catch (err) {
    error.value = t('upgrades.error_loading', { error: err.message }) || `Error loading upgrades: ${err.message}`
  } finally {
    loading.value = false
  }
}

const truncateText = (text, maxLength) => {
  if (!text) return '-'
  if (text.length <= maxLength) return text
  return text.substring(0, maxLength) + '...'
}

const openAddDialog = () => {
  isEditing.value = false
  formData.value = { description: '', notes: '' }
  showDialog.value = true
}

const editUpgrade = (upgrade) => {
  if (!canEdit(upgrade)) {
    alert(t('upgrades.no_permission_to_edit') || 'You do not have permission to edit this upgrade')
    return
  }
  isEditing.value = true
  formData.value = {
    id: upgrade.id,
    description: upgrade.description || '',
    notes: upgrade.notes || '',
  }
  showDialog.value = true
}

const closeDialog = () => {
  showDialog.value = false
  formData.value = { description: '', notes: '' }
  isSubmitting.value = false
}

const saveUpgrade = async () => {
  if (isSubmitting.value || !formData.value.description.trim()) return

  isSubmitting.value = true
  error.value = null

  try {
    const description = formData.value.description.trim()
    const notes = formData.value.notes.trim() || null

    if (isEditing.value) {
      // Update existing upgrade
      const result = await callApi({
        query: 'UPDATE upgrades SET description = ?, notes = ? WHERE id = ?',
        params: [description, notes, formData.value.id],
      })

      if (result.success) {
        await fetchUpgrades()
        closeDialog()
      } else {
        error.value = result.error || t('upgrades.failed_to_update') || 'Failed to update upgrade'
      }
    } else {
      // Create new upgrade - try with id_user_owner first, fallback without it
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
        await fetchUpgrades()
        closeDialog()
      } else {
        error.value = result.error || t('upgrades.failed_to_create') || 'Failed to create upgrade'
      }
    }
  } catch (err) {
    error.value = t('upgrades.error_saving', { error: err.message }) || `Error saving upgrade: ${err.message}`
  } finally {
    isSubmitting.value = false
  }
}

const deleteUpgrade = (upgrade) => {
  if (!canDelete(upgrade)) {
    alert(t('upgrades.no_permission_to_delete') || 'You do not have permission to delete this upgrade')
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

const confirmDelete = async () => {
  if (isDeleting.value || !upgradeToDelete.value) return

  isDeleting.value = true
  error.value = null

  try {
    const result = await callApi({
      query: 'DELETE FROM upgrades WHERE id = ?',
      params: [upgradeToDelete.value.id],
    })

    if (result.success) {
      await fetchUpgrades()
      closeDeleteDialog()
    } else {
      error.value = result.error || t('upgrades.failed_to_delete') || 'Failed to delete upgrade'
    }
  } catch (err) {
    error.value = t('upgrades.error_deleting', { error: err.message }) || `Error deleting upgrade: ${err.message}`
  } finally {
    isDeleting.value = false
  }
}

const refreshData = () => {
  fetchUpgrades()
}

onMounted(() => {
  fetchUpgrades()
})
</script>

<style scoped>
.upgrades-container {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #e5e7eb;
  background-color: #f8fafc;
}

.table-header h3 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #1f2937;
  font-size: 1.1rem;
}

.table-header h3 i {
  color: #f59e0b;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.add-btn,
.refresh-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.add-btn {
  background-color: #3498db;
  color: white;
}

.add-btn:hover:not(:disabled) {
  background-color: #2980b9;
}

.refresh-btn {
  background-color: #f8fafc;
  color: #374151;
  border: 1px solid #d1d5db;
}

.refresh-btn:hover:not(:disabled) {
  background-color: #f3f4f6;
}

.refresh-btn.processing {
  opacity: 0.7;
}

.loading-indicator {
  margin-left: 4px;
}

.table-wrapper {
  position: relative;
  min-height: 200px;
}

.loading-overlay {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
  color: #6b7280;
  gap: 12px;
}

.error-message {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 16px 20px;
  background-color: #fee2e2;
  color: #dc2626;
  border-radius: 4px;
  margin: 16px;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
  color: #6b7280;
  gap: 12px;
}

.empty-state i {
  color: #d1d5db;
}

.add-first-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background-color: #3498db;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 0.9rem;
  cursor: pointer;
  transition: background-color 0.2s;
}

.add-first-btn:hover {
  background-color: #2980b9;
}

.table-content {
  overflow-x: auto;
}

.upgrades-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;
}

.upgrades-table th {
  background-color: #f8fafc;
  padding: 12px 8px;
  text-align: left;
  font-weight: 600;
  color: #374151;
  border-bottom: 1px solid #e5e7eb;
}

.upgrades-table td {
  padding: 12px 8px;
  border-bottom: 1px solid #f3f4f6;
  color: #374151;
}

.table-row:hover {
  background-color: #f9fafb;
  transition: background-color 0.2s ease;
}

.selected-row {
  background-color: #dbeafe !important;
  border-left: 4px solid #3b82f6;
}

.selected-row:hover {
  background-color: #bfdbfe !important;
}

.id-cell {
  font-weight: 600;
  color: #6b7280;
  width: 60px;
}

.description-cell {
  font-weight: 500;
  min-width: 200px;
}

.notes-cell {
  color: #6b7280;
  max-width: 300px;
}

.owner-cell {
  color: #6b7280;
  white-space: nowrap;
}

.actions-cell {
  width: 100px;
}

.action-buttons {
  display: flex;
  gap: 4px;
}

.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 0.8rem;
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

.delete-btn:disabled {
  background-color: #d1d5db;
  color: #9ca3af;
  cursor: not-allowed;
}

.table-footer {
  padding: 12px 20px;
  border-top: 1px solid #e5e7eb;
  background-color: #f8fafc;
}

.record-count {
  color: #6b7280;
  font-size: 0.9rem;
}

/* Dialog Styles */
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
  z-index: 1000;
}

.dialog {
  background: white;
  border-radius: 8px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #e5e7eb;
}

.dialog-header h3 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #1f2937;
}

.dialog-header h3 i {
  color: #f59e0b;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.2rem;
  color: #6b7280;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.close-btn:hover:not(:disabled) {
  background-color: #f3f4f6;
}

.dialog-content {
  padding: 20px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  font-weight: 500;
  color: #374151;
}

.form-group label i {
  color: #3498db;
}

.required {
  color: #dc2626;
}

.form-group input,
.textarea-field {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 0.9rem;
  font-family: inherit;
  transition: border-color 0.2s;
}

.form-group input:focus,
.textarea-field:focus {
  outline: none;
  border-color: #3498db;
  box-shadow: 0 0 0 3px rgba(52, 152, 219, 0.1);
}

.form-group input:disabled,
.textarea-field:disabled {
  background-color: #f9fafb;
  color: #6b7280;
}

.textarea-field {
  resize: vertical;
  min-height: 80px;
}

.form-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 24px;
}

.cancel-btn,
.save-btn,
.delete-confirm-btn {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 8px;
}

.cancel-btn {
  background-color: #f3f4f6;
  color: #374151;
}

.cancel-btn:hover:not(:disabled) {
  background-color: #e5e7eb;
}

.save-btn {
  background-color: #3498db;
  color: white;
}

.save-btn:hover:not(:disabled) {
  background-color: #2980b9;
}

.save-btn.processing {
  opacity: 0.7;
}

.save-btn:disabled {
  background-color: #d1d5db;
  color: #9ca3af;
  cursor: not-allowed;
}

.delete-confirm-btn {
  background-color: #ef4444;
  color: white;
}

.delete-confirm-btn:hover:not(:disabled) {
  background-color: #dc2626;
}

.delete-confirm-btn:disabled {
  background-color: #d1d5db;
  color: #9ca3af;
  cursor: not-allowed;
}

.delete-confirm-btn.processing {
  opacity: 0.7;
}

/* Delete Dialog Specific Styles */
.delete-dialog {
  max-width: 400px;
}

.delete-message {
  margin-bottom: 20px;
}

.delete-message p {
  margin: 0 0 12px 0;
  color: #374151;
}

.warning-text {
  color: #dc2626 !important;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 8px;
}

.warning-text i {
  color: #dc2626;
}
</style>

