<template>
  <div class="containers-crud-container">
    <div class="table-header">
      <h3>
        <i class="fas fa-box"></i>
        Containers
      </h3>
      <div class="header-actions">
        <button @click="openAddDialog" class="add-btn" :disabled="loading">
          <i class="fas fa-plus"></i>
          <span>Add Container</span>
        </button>
        <button
          @click="refreshData"
          class="refresh-btn"
          :disabled="loading"
          :class="{ processing: loading }"
        >
          <i class="fas fa-sync-alt"></i>
          <span>Refresh</span>
          <i v-if="loading" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>
      </div>
    </div>

    <div class="table-wrapper">
      <div v-if="loading" class="loading-overlay">
        <i class="fas fa-spinner fa-spin fa-2x"></i>
        <span>Loading containers...</span>
      </div>

      <div v-else-if="error" class="error-message">
        <i class="fas fa-exclamation-circle"></i>
        {{ error }}
      </div>

      <div v-else-if="containers.length === 0" class="empty-state">
        <i class="fas fa-box-open fa-2x"></i>
        <p>No containers found</p>
        <button @click="openAddDialog" class="add-first-btn">
          <i class="fas fa-plus"></i>
          Add First Container
        </button>
      </div>

      <div v-else class="table-content">
        <table class="containers-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>Container Name</th>
              <th>Usage Count</th>
              <th>Last Used</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="container in containers"
              :key="container.id"
              class="table-row"
              :class="{ 'selected-row': selectedContainerId === container.id }"
              @click="handleContainerClick(container)"
            >
              <td class="id-cell">#{{ container.id }}</td>
              <td class="name-cell">{{ container.name }}</td>
              <td class="usage-cell">
                <span class="usage-badge">{{ container.usage_count || 0 }}</span>
              </td>
              <td class="date-cell">
                {{ container.last_used ? formatDate(container.last_used) : 'Never' }}
              </td>
              <td class="actions-cell">
                <div class="action-buttons">
                  <button
                    @click.stop="editContainer(container)"
                    class="action-btn edit-btn"
                    title="Edit Container"
                  >
                    <i class="fas fa-edit"></i>
                  </button>
                  <button
                    @click.stop="deleteContainer(container)"
                    class="action-btn delete-btn"
                    title="Delete Container"
                    :disabled="container.usage_count > 0"
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
      <span class="record-count"> Showing {{ containers.length }} containers </span>
    </div>

    <!-- Add/Edit Dialog -->
    <div v-if="showDialog" class="dialog-overlay" @click.self="closeDialog">
      <div class="dialog">
        <div class="dialog-header">
          <h3>
            <i class="fas fa-box"></i>
            {{ isEditing ? 'Edit Container' : 'Add New Container' }}
          </h3>
          <button class="close-btn" @click="closeDialog" :disabled="isSubmitting">
            <i class="fas fa-times"></i>
          </button>
        </div>

        <form @submit.prevent="saveContainer" class="dialog-content">
          <div class="form-group">
            <label for="name">
              <i class="fas fa-tag"></i>
              Container Name
            </label>
            <input
              type="text"
              id="name"
              v-model="formData.name"
              required
              :disabled="isSubmitting"
              maxlength="30"
              placeholder="Enter container name"
              @blur="validateName"
            />
            <div v-if="nameError" class="error-text">{{ nameError }}</div>
          </div>

          <div class="form-actions">
            <button type="button" @click="closeDialog" class="cancel-btn" :disabled="isSubmitting">
              Cancel
            </button>
            <button
              type="submit"
              class="save-btn"
              :disabled="isSubmitting || !!nameError"
              :class="{ processing: isSubmitting }"
            >
              <i v-if="isSubmitting" class="fas fa-spinner fa-spin"></i>
              <span>{{ isEditing ? 'Update' : 'Create' }}</span>
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
            Confirm Delete
          </h3>
          <button class="close-btn" @click="closeDeleteDialog" :disabled="isDeleting">
            <i class="fas fa-times"></i>
          </button>
        </div>

        <div class="dialog-content">
          <div class="delete-message">
            <p>
              Are you sure you want to delete container
              <strong>"{{ containerToDelete?.name }}"</strong>?
            </p>
            <p v-if="containerToDelete?.usage_count > 0" class="warning-text">
              <i class="fas fa-exclamation-triangle"></i>
              This container has been used {{ containerToDelete.usage_count }} time(s) and cannot be
              deleted.
            </p>
          </div>

          <div class="form-actions">
            <button
              type="button"
              @click="closeDeleteDialog"
              class="cancel-btn"
              :disabled="isDeleting"
            >
              Cancel
            </button>
            <button
              type="button"
              @click="confirmDelete"
              class="delete-confirm-btn"
              :disabled="isDeleting || containerToDelete?.usage_count > 0"
              :class="{ processing: isDeleting }"
            >
              <i v-if="isDeleting" class="fas fa-spinner fa-spin"></i>
              <span>Delete</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useApi } from '@/composables/useApi'

const { callApi } = useApi()

const containers = ref([])
const loading = ref(false)
const error = ref(null)
const showDialog = ref(false)
const showDeleteDialog = ref(false)
const isEditing = ref(false)
const isSubmitting = ref(false)
const isDeleting = ref(false)
const containerToDelete = ref(null)
const nameError = ref('')
const selectedContainerId = ref(null)

const formData = ref({
  name: '',
})

const fetchContainers = async () => {
  loading.value = true
  error.value = null

  try {
    const result = await callApi({
      query: `
        SELECT 
          c.id,
          c.name,
          COUNT(lc.id) as usage_count,
          MAX(lc.date_loaded) as last_used
        FROM containers c
        LEFT JOIN loaded_containers lc ON c.id = lc.id_container
        GROUP BY c.id, c.name
        ORDER BY c.name ASC
      `,
    })

    if (result.success) {
      containers.value = result.data || []
    } else {
      error.value = result.error || 'Failed to load containers'
    }
  } catch (err) {
    error.value = 'Error loading containers: ' + err.message
  } finally {
    loading.value = false
  }
}

const openAddDialog = () => {
  isEditing.value = false
  formData.value = { name: '' }
  nameError.value = ''
  showDialog.value = true
}

const editContainer = (container) => {
  isEditing.value = true
  formData.value = { name: container.name, id: container.id }
  nameError.value = ''
  showDialog.value = true
}

const closeDialog = () => {
  showDialog.value = false
  formData.value = { name: '' }
  nameError.value = ''
  isSubmitting.value = false
}

const validateName = async () => {
  nameError.value = ''

  if (!formData.value.name.trim()) {
    nameError.value = 'Container name is required'
    return
  }

  if (formData.value.name.length < 2) {
    nameError.value = 'Container name must be at least 2 characters'
    return
  }

  // Check for duplicate names (excluding current container if editing)
  const existingContainer = containers.value.find(
    (c) =>
      c.name.toLowerCase() === formData.value.name.toLowerCase() &&
      (!isEditing.value || c.id !== formData.value.id),
  )

  if (existingContainer) {
    nameError.value = 'A container with this name already exists'
    return
  }
}

const saveContainer = async () => {
  if (isSubmitting.value) return

  await validateName()
  if (nameError.value) return

  isSubmitting.value = true

  try {
    const query = isEditing.value
      ? 'UPDATE containers SET name = ? WHERE id = ?'
      : 'INSERT INTO containers (name) VALUES (?)'

    const params = isEditing.value
      ? [formData.value.name, formData.value.id]
      : [formData.value.name]

    const result = await callApi({ query, params })

    if (result.success) {
      await fetchContainers()
      closeDialog()
    } else {
      error.value = result.error || 'Failed to save container'
    }
  } catch (err) {
    error.value = 'Error saving container: ' + err.message
  } finally {
    isSubmitting.value = false
  }
}

const deleteContainer = (container) => {
  containerToDelete.value = container
  showDeleteDialog.value = true
}

const closeDeleteDialog = () => {
  showDeleteDialog.value = false
  containerToDelete.value = null
  isDeleting.value = false
}

const confirmDelete = async () => {
  if (isDeleting.value || !containerToDelete.value) return

  isDeleting.value = true

  try {
    const result = await callApi({
      query: 'DELETE FROM containers WHERE id = ?',
      params: [containerToDelete.value.id],
    })

    if (result.success) {
      await fetchContainers()
      closeDeleteDialog()
    } else {
      error.value = result.error || 'Failed to delete container'
    }
  } catch (err) {
    error.value = 'Error deleting container: ' + err.message
  } finally {
    isDeleting.value = false
  }
}

const refreshData = () => {
  fetchContainers()
}

const handleContainerClick = (container) => {
  selectedContainerId.value = container.id
}

const formatDate = (dateString) => {
  if (!dateString) return '-'
  const date = new Date(dateString)
  return date.toLocaleDateString('en-GB', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  })
}

onMounted(() => {
  fetchContainers()
})
</script>

<style scoped>
.containers-crud-container {
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
  color: #3498db;
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

.containers-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;
}

.containers-table th {
  background-color: #f8fafc;
  padding: 12px 8px;
  text-align: left;
  font-weight: 600;
  color: #374151;
  border-bottom: 1px solid #e5e7eb;
}

.containers-table td {
  padding: 12px 8px;
  border-bottom: 1px solid #f3f4f6;
  color: #374151;
}

.table-row:hover {
  background-color: #f9fafb;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.table-row:active {
  background-color: #e5e7eb;
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

.name-cell {
  font-weight: 500;
}

.usage-cell {
  text-align: center;
}

.usage-badge {
  display: inline-block;
  padding: 2px 8px;
  background-color: #e5e7eb;
  color: #374151;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: 500;
}

.date-cell {
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
  color: #3498db;
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

.form-group input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 0.9rem;
  transition: border-color 0.2s;
}

.form-group input:focus {
  outline: none;
  border-color: #3498db;
  box-shadow: 0 0 0 3px rgba(52, 152, 219, 0.1);
}

.form-group input:disabled {
  background-color: #f9fafb;
  color: #6b7280;
}

.error-text {
  color: #dc2626;
  font-size: 0.8rem;
  margin-top: 4px;
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
