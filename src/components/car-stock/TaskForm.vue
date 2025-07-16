<script setup>
import { ref, onMounted, watch, computed } from 'vue'
import { useEnhancedI18n } from '@/composables/useI18n'
import { useApi } from '../../composables/useApi'
import { useRouter } from 'vue-router'

const { t } = useEnhancedI18n()
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
    validator: (value) =>
      ['car', 'client', 'supplier', 'sell', 'buy', 'loading', 'general'].includes(value),
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
  id_user_receive: null, // Keep for backward compatibility
  id_priority: null,
  is_task_for_car: 0, // Will be set based on entity type
  // New fields for multi-assign and multi-subjects
  assigned_users_ids: [], // Array of user IDs for multi-assign
  subject_ids: [], // Array of subject IDs for multi-subjects
})

const users = ref([])
const priorities = ref([])
const subjects = ref([]) // Add subjects for multi-subjects
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
      label: t('taskForm.car'),
      idField: 'id',
      nameField: 'car_name',
      route: '/cars/stock',
      taskFlag: 1,
      defaultTitle: t('taskForm.car_task'),
    },
    client: {
      icon: 'fas fa-user',
      label: t('taskForm.client'),
      idField: 'id',
      nameField: 'name',
      route: '/clients',
      taskFlag: 0,
      defaultTitle: t('taskForm.client_task'),
    },
    supplier: {
      icon: 'fas fa-truck',
      label: t('taskForm.supplier'),
      idField: 'id',
      nameField: 'name',
      route: '/suppliers',
      taskFlag: 0,
      defaultTitle: t('taskForm.supplier_task'),
    },
    sell: {
      icon: 'fas fa-handshake',
      label: t('taskForm.sell_bill'),
      idField: 'id',
      nameField: 'bill_ref',
      route: '/sells',
      taskFlag: 0,
      defaultTitle: t('taskForm.sell_task'),
    },
    buy: {
      icon: 'fas fa-shopping-cart',
      label: t('taskForm.buy_bill'),
      idField: 'id',
      nameField: 'bill_ref',
      route: '/buy',
      taskFlag: 0,
      defaultTitle: t('taskForm.buy_task'),
    },
    loading: {
      icon: 'fas fa-truck-loading',
      label: t('taskForm.loading_record'),
      idField: 'id',
      nameField: 'shipping_line_name',
      route: '/loading',
      taskFlag: 0,
      defaultTitle: t('taskForm.loading_task'),
    },
    general: {
      icon: 'fas fa-tasks',
      label: t('taskForm.general'),
      idField: null,
      nameField: null,
      route: null,
      taskFlag: 0,
      defaultTitle: t('taskForm.general_task'),
    },
  }
  return configs[props.entityType] || configs.general
})

const entityDisplayName = computed(() => {
  const config = entityConfig.value
  if (props.entityType === 'general') {
    return t('taskForm.general_task')
  }

  const id = props.entityData[config.idField]
  const name = props.entityData[config.nameField] || t('taskForm.unknown')
  return `${config.label} #${id} - ${name}`
})

const entityLink = computed(() => {
  const config = entityConfig.value
  if (props.entityType === 'general' || !config.idField) {
    return null
  }

  const id = props.entityData[config.idField]
  const name = props.entityData[config.nameField] || t('taskForm.unknown')
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
        id_user_receive: null, // Keep for backward compatibility
        id_priority: null,
        is_task_for_car: entityConfig.value.taskFlag,
        // New fields for multi-assign and multi-subjects
        assigned_users_ids: [],
        subject_ids: [],
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
  await fetchSubjects()
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

const fetchSubjects = async () => {
  try {
    const result = await callApi({
      query: `
        SELECT id, name as subject_name
        FROM subjects
        WHERE is_active = 1
        ORDER BY name ASC
      `,
      requiresAuth: true,
    })

    if (result.success && result.data) {
      subjects.value = result.data
    } else {
      // If no subjects found, set empty array
      subjects.value = []
    }
  } catch (err) {
    console.error('Error fetching subjects:', err)
    // Set empty array on error to prevent issues
    subjects.value = []
  }
}

const handleSubmit = async () => {
  if (!formData.value.title.trim()) {
    alert(t('taskForm.please_enter_task_title'))
    return
  }

  // Check for user assignment (support both single and multi-assign)
  const hasUserAssignment =
    formData.value.id_user_receive ||
    (formData.value.assigned_users_ids && formData.value.assigned_users_ids.length > 0)

  if (!hasUserAssignment) {
    alert(t('taskForm.please_select_at_least_one_user'))
    return
  }

  if (!formData.value.id_priority) {
    alert(t('taskForm.please_select_priority'))
    return
  }

  try {
    // Create the entity link if not general task
    let finalTitle = formData.value.title.trim()
    let finalDescription = formData.value.description.trim()

    if (props.entityType === 'car') {
      // For car tasks, add [cars] prefix to indicate it's about cars
      finalTitle = `[cars] ${finalTitle}`
    } else if (props.entityType !== 'general' && entityLink.value) {
      // Add the entity link to the beginning of the title (but not for car tasks)
      finalTitle = `${entityLink.value} ${finalTitle}`
    }

    // For car tasks, add car IDs to the description
    if (props.entityType === 'car') {
      const carIds = Array.isArray(props.entityData.id)
        ? props.entityData.id
        : [props.entityData.id]
      const carIdsText = carIds.map((id) => `#${id}`).join(', ')
      finalDescription = finalDescription
        ? `${finalDescription}\n\n${t('taskForm.car_ids')}: ${carIdsText}`
        : `${t('taskForm.car_ids')}: ${carIdsText}`
    }

    // Prepare user assignment data
    const assignedUsersIds =
      formData.value.assigned_users_ids && formData.value.assigned_users_ids.length > 0
        ? JSON.stringify(formData.value.assigned_users_ids.map((id) => parseInt(id)))
        : formData.value.id_user_receive
          ? JSON.stringify([formData.value.id_user_receive])
          : null

    // Prepare subject data - for car tasks, this should be car IDs
    let subjectIds = null

    if (props.entityType === 'car') {
      // For car tasks, subject_ids should contain the car ID(s)
      const carIds = Array.isArray(props.entityData.id)
        ? props.entityData.id
        : [props.entityData.id]
      subjectIds = JSON.stringify(carIds.map((id) => parseInt(id)))
    } else if (formData.value.subject_ids && formData.value.subject_ids.length > 0) {
      // For other entity types, use the selected subjects
      subjectIds = JSON.stringify(formData.value.subject_ids.map((id) => parseInt(id)))
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
          is_task_for_car,
          assigned_users_ids,
          subject_ids
        ) VALUES (?, ?, UTC_TIMESTAMP(), ?, ?, ?, ?, ?, ?, ?)
      `,
      params: [
        currentUser.value.id,
        formData.value.id_user_receive || null, // Keep for backward compatibility
        formData.value.id_priority,
        finalTitle,
        finalDescription,
        formData.value.notes.trim(),
        props.entityType === 'car' ? 1 : entityConfig.value.taskFlag,
        assignedUsersIds,
        subjectIds,
      ],
      requiresAuth: true,
    })

    if (result.success) {
      console.log('Task created successfully:', result)
      emit('task-created', { success: true, taskId: result.lastInsertId })
    } else {
      console.error('Failed to create task:', result)
      alert(t('taskForm.failed_to_create_task'))
    }
  } catch (err) {
    console.error('Error creating task:', err)
    alert(t('taskForm.error_creating_task'))
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
    alert(t('taskForm.please_enter_priority_name'))
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
      alert(t('taskForm.failed_to_create_priority'))
    }
  } catch (err) {
    console.error('Error creating priority:', err)
    alert(t('taskForm.error_creating_priority'))
  }
}

const cancelAddPriority = () => {
  showPriorityForm.value = false
  newPriority.value = {
    priority: '',
    power: 1,
  }
}

// Handle single user assignment changes
const handleSingleUserChange = () => {
  // If a single user is selected from dropdown, update multi-assign checkboxes
  if (formData.value.id_user_receive) {
    formData.value.assigned_users_ids = [formData.value.id_user_receive]
  } else {
    formData.value.assigned_users_ids = []
  }
}

// Handle multi-user assignment changes
const handleMultiUserChange = () => {
  // If exactly one user is selected in multi-assign, update single assign dropdown
  if (formData.value.assigned_users_ids && formData.value.assigned_users_ids.length === 1) {
    formData.value.id_user_receive = formData.value.assigned_users_ids[0]
  } else {
    // If multiple users or no users are selected, clear single assign
    formData.value.id_user_receive = null
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
              {{ entityConfig.label }}
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
                :title="`${t('taskForm.view')} ${entityConfig.label.toLowerCase()}`"
              >
                <i class="fas fa-external-link-alt"></i>
                {{ t('taskForm.view') }} {{ entityConfig.label }}
              </button>
            </div>
          </div>

          <form @submit.prevent="handleSubmit" class="form">
            <div class="form-group">
              <label for="title">{{ t('taskForm.task_title') }} *</label>
              <div class="title-input-container">
                <input
                  id="title"
                  v-model="formData.title"
                  type="text"
                  :placeholder="`${t('taskForm.enter')} ${entityConfig.defaultTitle.toLowerCase()}`"
                  required
                  maxlength="255"
                  class="title-input"
                />
                <button
                  v-if="entityConfig.route"
                  type="button"
                  @click="openEntityView"
                  class="btn-title-link"
                  :title="`${t('taskForm.open')} ${entityConfig.label.toLowerCase()} ${t('taskForm.view')}`"
                >
                  <i class="fas fa-external-link-alt"></i>
                </button>
              </div>
            </div>

            <div class="form-group">
              <label for="description">{{ t('taskForm.description') }}</label>
              <textarea
                id="description"
                v-model="formData.description"
                :placeholder="t('taskForm.enter_task_description')"
                rows="3"
                maxlength="1000"
              ></textarea>
            </div>

            <div class="form-row">
              <div class="form-group">
                <label for="user_receive">{{ t('taskForm.assign_to') }} *</label>
                <div class="assignment-container">
                  <!-- Single user assignment (for backward compatibility) -->
                  <div class="single-assignment">
                    <select
                      id="user_receive"
                      v-model="formData.id_user_receive"
                      @change="handleSingleUserChange"
                    >
                      <option value="">{{ t('taskForm.select_a_user') }}</option>
                      <option v-for="user in users" :key="user.id" :value="user.id">
                        {{ user.username }}
                        <span v-if="user.email">({{ user.email }})</span>
                      </option>
                    </select>
                  </div>

                  <!-- Multi-user assignment -->
                  <div class="multi-assignment">
                    <label class="multi-label">{{
                      t('taskForm.or_assign_to_multiple_users')
                    }}</label>
                    <div class="multi-select-container">
                      <div v-for="user in users" :key="user.id" class="user-checkbox-item">
                        <input
                          type="checkbox"
                          :id="'user_' + user.id"
                          :value="user.id"
                          v-model="formData.assigned_users_ids"
                          @change="handleMultiUserChange"
                        />
                        <label :for="'user_' + user.id" class="checkbox-label">
                          {{ user.username }}
                          <span v-if="user.email">({{ user.email }})</span>
                        </label>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <div class="form-group">
                <label for="priority">{{ t('taskForm.priority') }} *</label>
                <div class="priority-select-container">
                  <select id="priority" v-model="formData.id_priority" required>
                    <option value="">{{ t('taskForm.select_priority') }}</option>
                    <option v-for="priority in priorities" :key="priority.id" :value="priority.id">
                      {{ priority.priority }}
                    </option>
                  </select>
                  <button
                    type="button"
                    @click="showAddPriorityForm"
                    class="btn-add-priority"
                    :title="t('taskForm.add_new_priority')"
                  >
                    <i class="fas fa-plus"></i>
                  </button>
                </div>
              </div>
            </div>

            <!-- Multi-subjects section -->
            <div class="form-group" v-if="subjects.length > 0">
              <label>{{ t('taskForm.subjects') }} ({{ t('taskForm.optional') }})</label>
              <div class="subjects-container">
                <div v-for="subject in subjects" :key="subject.id" class="subject-checkbox-item">
                  <input
                    type="checkbox"
                    :id="'subject_' + subject.id"
                    :value="subject.id"
                    v-model="formData.subject_ids"
                  />
                  <label :for="'subject_' + subject.id" class="checkbox-label">
                    {{ subject.subject_name }}
                  </label>
                </div>
              </div>
            </div>

            <div class="form-group">
              <label for="notes">{{ t('taskForm.notes') }}</label>
              <textarea
                id="notes"
                v-model="formData.notes"
                :placeholder="t('taskForm.additional_notes')"
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
                {{ t('taskForm.cancel') }}
              </button>
              <button type="submit" class="btn btn-primary" :disabled="loading">
                <i v-if="loading" class="fas fa-spinner fa-spin"></i>
                <span v-else>{{ t('taskForm.create_task') }}</span>
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
              {{ t('taskForm.add_new_priority') }}
            </h3>
          </div>

          <form @submit.prevent="handleAddPriority" class="form">
            <div class="form-group">
              <label for="new_priority_name">{{ t('taskForm.priority_name') }} *</label>
              <input
                id="new_priority_name"
                v-model="newPriority.priority"
                type="text"
                :placeholder="t('taskForm.enter_priority_name')"
                required
                maxlength="255"
              />
            </div>

            <div class="form-group">
              <label for="new_priority_power">{{ t('taskForm.priority_level') }}</label>
              <select id="new_priority_power" v-model="newPriority.power">
                <option value="1">{{ t('taskForm.low') }} (1)</option>
                <option value="2">{{ t('taskForm.medium') }} (2)</option>
                <option value="3">{{ t('taskForm.high') }} (3)</option>
                <option value="4">{{ t('taskForm.urgent') }} (4)</option>
                <option value="5">{{ t('taskForm.critical') }} (5)</option>
              </select>
            </div>

            <div class="form-actions">
              <button type="button" @click="cancelAddPriority" class="btn btn-secondary">
                {{ t('taskForm.cancel') }}
              </button>
              <button type="submit" class="btn btn-primary">
                <i v-if="loading" class="fas fa-spinner fa-spin"></i>
                <span v-else>{{ t('taskForm.add_priority') }}</span>
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
  z-index: 1002;
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

/* Multi-assignment styles */
.assignment-container {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.single-assignment select {
  width: 100%;
}

.multi-assignment {
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  padding: 12px;
  background-color: #f9fafb;
}

.multi-label {
  font-size: 0.85rem;
  color: #6b7280;
  margin-bottom: 8px;
  display: block;
}

.multi-select-container {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 8px;
  max-height: 120px;
  overflow-y: auto;
}

.user-checkbox-item,
.subject-checkbox-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 0;
}

.user-checkbox-item input[type='checkbox'],
.subject-checkbox-item input[type='checkbox'] {
  margin: 0;
  width: auto;
}

.checkbox-label {
  font-size: 0.9rem;
  color: #374151;
  cursor: pointer;
  margin: 0;
  flex: 1;
}

/* Subjects styles */
.subjects-container {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 8px;
  max-height: 120px;
  overflow-y: auto;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  padding: 12px;
  background-color: #f9fafb;
}

@media (max-width: 640px) {
  .multi-select-container,
  .subjects-container {
    grid-template-columns: 1fr;
  }
}
</style>
