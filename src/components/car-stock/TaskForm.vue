<script setup>
import { ref, onMounted, watch, computed } from 'vue'
import { useApi } from '../../composables/useApi'
import { useRouter } from 'vue-router'

const router = useRouter()

const props = defineProps({
  // Entity data - can be car, client, supplier, sell, buy, etc.
  entityData: {
    type: Object,
    required: true,
  },
  // Entity type to determine behavior and routing
  entityType: {
    type: String,
    required: true,
    validator: (value) => ['car', 'client', 'supplier', 'sell', 'buy', 'general'].includes(value),
  },
  isVisible: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['save', 'cancel', 'task-created'])

const { callApi, loading, error } = useApi()

const formData = ref({
  title: '',
  description: '',
  notes: '',
  id_user_receive: null,
  id_priority: null,
  is_task_for_car: 0, // Will be set based on entity type
})

const users = ref([])
const priorities = ref([])
const currentUser = ref(null)

// Add new state for priority creation
const showPriorityForm = ref(false)
const newPriority = ref({
  priority: '',
  power: 1,
})

// Computed properties for dynamic behavior
const entityConfig = computed(() => {
  const configs = {
    car: {
      icon: 'fas fa-car',
      label: 'Car',
      idField: 'id',
      nameField: 'car_name',
      route: '/cars/stock',
      taskFlag: 1,
      defaultTitle: 'Car Task',
    },
    client: {
      icon: 'fas fa-user',
      label: 'Client',
      idField: 'id',
      nameField: 'name',
      route: '/clients',
      taskFlag: 0,
      defaultTitle: 'Client Task',
    },
    supplier: {
      icon: 'fas fa-truck',
      label: 'Supplier',
      idField: 'id',
      nameField: 'name',
      route: '/suppliers',
      taskFlag: 0,
      defaultTitle: 'Supplier Task',
    },
    sell: {
      icon: 'fas fa-handshake',
      label: 'Sell Bill',
      idField: 'id',
      nameField: 'bill_ref',
      route: '/sells',
      taskFlag: 0,
      defaultTitle: 'Sell Task',
    },
    buy: {
      icon: 'fas fa-shopping-cart',
      label: 'Buy Bill',
      idField: 'id',
      nameField: 'bill_number',
      route: '/buy',
      taskFlag: 0,
      defaultTitle: 'Buy Task',
    },
    general: {
      icon: 'fas fa-tasks',
      label: 'General',
      idField: null,
      nameField: null,
      route: null,
      taskFlag: 0,
      defaultTitle: 'General Task',
    },
  }
  return configs[props.entityType] || configs.general
})

const entityDisplayName = computed(() => {
  const config = entityConfig.value
  if (props.entityType === 'general') {
    return 'General Task'
  }

  const id = props.entityData[config.idField]
  const name = props.entityData[config.nameField] || 'Unknown'
  return `${config.label} #${id} - ${name}`
})

const entityLink = computed(() => {
  const config = entityConfig.value
  if (props.entityType === 'general' || !config.idField) {
    return null
  }

  const id = props.entityData[config.idField]
  const name = props.entityData[config.nameField] || 'Unknown'
  return `[${config.label} #${id} - ${name}]`
})

// Watch for modal visibility to reset form
watch(
  () => props.isVisible,
  (newValue) => {
    if (newValue) {
      // Reset form data when modal opens
      formData.value = {
        title: '',
        description: '',
        notes: '',
        id_user_receive: null,
        id_priority: null,
        is_task_for_car: entityConfig.value.taskFlag,
      }
      // Clear any error messages
      error.value = null
    }
  },
)

// Watch for entity data changes
watch(
  () => props.entityData,
  (newEntityData) => {
    if (newEntityData && props.entityType !== 'general') {
      console.log('Entity data updated:', newEntityData)
    }
  },
  { immediate: true },
)

onMounted(async () => {
  await getCurrentUser()
  await fetchUsers()
  await fetchPriorities()
})

const getCurrentUser = () => {
  try {
    const userStr = localStorage.getItem('user')
    if (userStr) {
      currentUser.value = JSON.parse(userStr)
    }
  } catch (err) {
    console.error('Error getting current user:', err)
  }
}

const fetchUsers = async () => {
  try {
    const result = await callApi({
      query: `
        SELECT id, username, email 
        FROM users 
        ORDER BY username ASC
      `,
      requiresAuth: true,
    })

    if (result.success && result.data) {
      users.value = result.data
    }
  } catch (err) {
    console.error('Error fetching users:', err)
  }
}

const fetchPriorities = async () => {
  try {
    const result = await callApi({
      query: `
        SELECT id, priority, power 
        FROM priorities 
        ORDER BY power ASC
      `,
      requiresAuth: true,
    })

    if (result.success && result.data) {
      priorities.value = result.data
    }
  } catch (err) {
    console.error('Error fetching priorities:', err)
  }
}

const handleSubmit = async () => {
  if (!formData.value.title.trim()) {
    alert('Please enter a task title')
    return
  }

  if (!formData.value.id_user_receive) {
    alert('Please select a user to assign the task to')
    return
  }

  if (!formData.value.id_priority) {
    alert('Please select a priority')
    return
  }

  try {
    // Create the entity link if not general task
    let finalTitle = formData.value.title.trim()

    if (props.entityType !== 'general' && entityLink.value) {
      // Add the entity link to the beginning of the title
      finalTitle = `${entityLink.value} ${finalTitle}`
    }

    const result = await callApi({
      query: `
        INSERT INTO tasks (
          id_user_create, 
          id_user_receive, 
          date_create, 
          id_priority, 
          title, 
          desciption, 
          notes, 
          is_task_for_car
        ) VALUES (?, ?, UTC_TIMESTAMP(), ?, ?, ?, ?, ?)
      `,
      params: [
        currentUser.value.id,
        formData.value.id_user_receive,
        formData.value.id_priority,
        finalTitle,
        formData.value.description.trim(),
        formData.value.notes.trim(),
        entityConfig.value.taskFlag,
      ],
      requiresAuth: true,
    })

    if (result.success) {
      console.log('Task created successfully:', result)
      emit('task-created', { success: true, taskId: result.lastInsertId })
    } else {
      console.error('Failed to create task:', result)
      alert('Failed to create task. Please try again.')
    }
  } catch (err) {
    console.error('Error creating task:', err)
    alert('Error creating task. Please try again.')
  }
}

const handleCancel = () => {
  emit('cancel')
}

const handleBackdropClick = (event) => {
  if (event.target === event.currentTarget) {
    handleCancel()
  }
}

// Add new methods for priority creation
const showAddPriorityForm = () => {
  showPriorityForm.value = true
  newPriority.value = {
    priority: '',
    power: 1,
  }
}

const handleAddPriority = async () => {
  if (!newPriority.value.priority.trim()) {
    alert('Please enter a priority name')
    return
  }

  try {
    const result = await callApi({
      query: `
        INSERT INTO priorities (priority, power) 
        VALUES (?, ?)
      `,
      params: [newPriority.value.priority.trim(), newPriority.value.power],
      requiresAuth: true,
    })

    if (result.success) {
      console.log('Priority created successfully:', result)
      // Refresh priorities list
      await fetchPriorities()
      // Set the new priority as selected
      formData.value.id_priority = result.lastInsertId
      showPriorityForm.value = false
    } else {
      console.error('Failed to create priority:', result)
      alert('Failed to create priority. Please try again.')
    }
  } catch (err) {
    console.error('Error creating priority:', err)
    alert('Error creating priority. Please try again.')
  }
}

const cancelAddPriority = () => {
  showPriorityForm.value = false
  newPriority.value = {
    priority: '',
    power: 1,
  }
}

// Function to open entity in its respective view
const openEntityView = () => {
  const config = entityConfig.value
  if (config.route) {
    router.push(config.route)
    handleCancel()
  }
}
</script>

<template>
  <Teleport to="body">
    <div v-if="isVisible" class="modal-overlay" @click="handleBackdropClick">
      <div class="modal-content">
        <div class="task-form">
          <div class="form-header">
            <h3>
              <i :class="entityConfig.icon"></i>
              Create Task for {{ entityConfig.label }}
            </h3>
            <div v-if="entityType !== 'general'" class="entity-info-container">
              <p class="entity-info">
                <i :class="entityConfig.icon"></i>
                {{ entityDisplayName }}
              </p>
              <button
                v-if="entityConfig.route"
                type="button"
                @click="openEntityView"
                class="btn-view-entity"
                :title="`View ${entityConfig.label.toLowerCase()}`"
              >
                <i class="fas fa-external-link-alt"></i>
                View {{ entityConfig.label }}
              </button>
            </div>
          </div>

          <form @submit.prevent="handleSubmit" class="form">
            <div class="form-group">
              <label for="title">Task Title *</label>
              <div class="title-input-container">
                <input
                  id="title"
                  v-model="formData.title"
                  type="text"
                  :placeholder="`Enter ${entityConfig.defaultTitle.toLowerCase()} title`"
                  required
                  maxlength="255"
                  class="title-input"
                />
                <button
                  v-if="entityConfig.route"
                  type="button"
                  @click="openEntityView"
                  class="btn-title-link"
                  :title="`Open ${entityConfig.label.toLowerCase()} view`"
                >
                  <i class="fas fa-external-link-alt"></i>
                </button>
              </div>
            </div>

            <div class="form-group">
              <label for="description">Description</label>
              <textarea
                id="description"
                v-model="formData.description"
                placeholder="Enter task description"
                rows="3"
                maxlength="1000"
              ></textarea>
            </div>

            <div class="form-row">
              <div class="form-group">
                <label for="user_receive">Assign To *</label>
                <select id="user_receive" v-model="formData.id_user_receive" required>
                  <option value="">Select a user</option>
                  <option v-for="user in users" :key="user.id" :value="user.id">
                    {{ user.username }}
                    <span v-if="user.email">({{ user.email }})</span>
                  </option>
                </select>
              </div>

              <div class="form-group">
                <label for="priority">Priority *</label>
                <div class="priority-select-container">
                  <select id="priority" v-model="formData.id_priority" required>
                    <option value="">Select priority</option>
                    <option v-for="priority in priorities" :key="priority.id" :value="priority.id">
                      {{ priority.priority }}
                    </option>
                  </select>
                  <button
                    type="button"
                    @click="showAddPriorityForm"
                    class="btn-add-priority"
                    title="Add new priority"
                  >
                    <i class="fas fa-plus"></i>
                  </button>
                </div>
              </div>
            </div>

            <div class="form-group">
              <label for="notes">Notes</label>
              <textarea
                id="notes"
                v-model="formData.notes"
                placeholder="Additional notes"
                rows="2"
                maxlength="1000"
              ></textarea>
            </div>

            <div class="form-actions">
              <button
                type="button"
                @click="handleCancel"
                class="btn btn-secondary"
                :disabled="loading"
              >
                Cancel
              </button>
              <button type="submit" class="btn btn-primary" :disabled="loading">
                <i v-if="loading" class="fas fa-spinner fa-spin"></i>
                <span v-else>Create Task</span>
              </button>
            </div>

            <div v-if="error" class="error-message">
              <i class="fas fa-exclamation-triangle"></i>
              {{ error }}
            </div>
          </form>
        </div>
      </div>
    </div>
  </Teleport>

  <!-- Priority Creation Modal -->
  <Teleport to="body">
    <div v-if="showPriorityForm" class="modal-overlay priority-modal" @click="cancelAddPriority">
      <div class="modal-content priority-modal-content" @click.stop>
        <div class="priority-form">
          <div class="form-header">
            <h3>
              <i class="fas fa-plus-circle"></i>
              Add New Priority
            </h3>
          </div>

          <form @submit.prevent="handleAddPriority" class="form">
            <div class="form-group">
              <label for="new_priority_name">Priority Name *</label>
              <input
                id="new_priority_name"
                v-model="newPriority.priority"
                type="text"
                placeholder="Enter priority name"
                required
                maxlength="255"
              />
            </div>

            <div class="form-group">
              <label for="new_priority_power">Priority Level</label>
              <select id="new_priority_power" v-model="newPriority.power">
                <option value="1">Low (1)</option>
                <option value="2">Medium (2)</option>
                <option value="3">High (3)</option>
                <option value="4">Urgent (4)</option>
                <option value="5">Critical (5)</option>
              </select>
            </div>

            <div class="form-actions">
              <button type="button" @click="cancelAddPriority" class="btn btn-secondary">
                Cancel
              </button>
              <button type="submit" class="btn btn-primary">
                <i v-if="loading" class="fas fa-spinner fa-spin"></i>
                <span v-else>Add Priority</span>
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(2px);
}

.modal-content {
  background: white;
  border-radius: 12px;
  box-shadow:
    0 20px 25px -5px rgba(0, 0, 0, 0.1),
    0 10px 10px -5px rgba(0, 0, 0, 0.04);
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

.task-form {
  padding: 24px;
}

.form-header {
  margin-bottom: 24px;
  text-align: center;
}

.form-header h3 {
  margin: 0 0 8px 0;
  color: #1f2937;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.form-header h3 i {
  color: #4caf50;
}

.entity-info-container {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  flex-wrap: wrap;
}

.entity-info {
  margin: 0;
  color: #6b7280;
  font-size: 0.9rem;
  display: flex;
  align-items: center;
  gap: 6px;
}

.form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

label {
  font-weight: 500;
  color: #374151;
  font-size: 0.9rem;
}

input,
select,
textarea {
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 0.9rem;
  transition: border-color 0.2s ease;
}

input:focus,
select:focus,
textarea:focus {
  outline: none;
  border-color: #4caf50;
  box-shadow: 0 0 0 3px rgba(76, 175, 80, 0.1);
}

textarea {
  resize: vertical;
  min-height: 60px;
}

.form-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 8px;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  gap: 6px;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-primary {
  background-color: #4caf50;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background-color: #45a049;
}

.btn-secondary {
  background-color: #6b7280;
  color: white;
}

.btn-secondary:hover:not(:disabled) {
  background-color: #5a6268;
}

.error-message {
  background-color: #fef2f2;
  border: 1px solid #fecaca;
  color: #dc2626;
  padding: 12px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.9rem;
}

@media (max-width: 640px) {
  .modal-content {
    width: 95%;
    margin: 10px;
  }

  .task-form {
    padding: 16px;
  }

  .form-row {
    grid-template-columns: 1fr;
  }

  .form-actions {
    flex-direction: column;
  }
}

/* Priority styles */
.priority-select-container {
  display: flex;
  gap: 8px;
  align-items: flex-end;
}

.priority-select-container select {
  flex: 1;
}

.btn-add-priority {
  padding: 10px 12px;
  background-color: #4caf50;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 40px;
  height: 42px;
}

.btn-add-priority:hover {
  background-color: #45a049;
}

.btn-add-priority i {
  font-size: 0.9rem;
}

/* Priority modal styles */
.priority-modal .modal-content {
  max-width: 400px;
}

.priority-modal-content {
  background: white;
  border-radius: 12px;
  box-shadow:
    0 20px 25px -5px rgba(0, 0, 0, 0.1),
    0 10px 10px -5px rgba(0, 0, 0, 0.04);
  max-width: 400px;
  width: 90%;
  max-height: 90vh;
  overflow-y: auto;
  animation: modalSlideIn 0.3s ease-out;
}

.priority-form {
  padding: 24px;
}

.priority-form .form-header h3 {
  margin: 0 0 8px 0;
  color: #1f2937;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.priority-form .form-header h3 i {
  color: #4caf50;
}

@media (max-width: 640px) {
  .priority-modal-content {
    width: 95%;
    margin: 10px;
  }

  .priority-form {
    padding: 16px;
  }
}

.btn-view-entity {
  padding: 8px 16px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.2s ease;
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.9rem;
}

.btn-view-entity:hover {
  background-color: #2563eb;
}

.btn-view-entity i {
  font-size: 0.8rem;
}

/* Title input with link styles */
.title-input-container {
  display: flex;
  gap: 8px;
  align-items: center;
}

.title-input {
  flex: 1;
}

.btn-title-link {
  padding: 10px 12px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 40px;
  height: 42px;
}

.btn-title-link:hover {
  background-color: #2563eb;
}

.btn-title-link i {
  font-size: 0.9rem;
}
</style>
