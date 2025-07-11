<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useApi } from '../composables/useApi'
import TaskForm from '../components/car-stock/TaskForm.vue'

const router = useRouter()
const { callApi } = useApi()

const tasks = ref([])
const loading = ref(true)
const error = ref(null)
const user = ref(null)

// TaskForm states
const showTaskForm = ref(false)
const currentEntityType = ref('general')
const currentEntityData = ref({})

// Description popup states
const showDescriptionPopup = ref(false)
const selectedTaskDescription = ref('')

// Filter states
const filters = ref({
  status: '',
  priority: '',
  assignedTo: '',
  createdBy: '',
  search: '',
})

const priorities = ref([])
const users = ref([])

// Pagination
const currentPage = ref(1)
const itemsPerPage = ref(20)
const totalItems = ref(0)

// Sorting
const sortConfig = ref({
  key: 'date_create',
  direction: 'desc',
})

const fetchTasks = async () => {
  loading.value = true
  error.value = null

  try {
    // First, let's test if the table exists and has data
    const testResult = await callApi({
      query: 'SELECT COUNT(*) as count FROM tasks',
      requiresAuth: true,
    })

    if (!testResult.success) {
      error.value = 'Tasks table not accessible: ' + (testResult.error || 'Unknown error')
      return
    }

    // Build the WHERE clause based on user role and filters
    let whereConditions = []
    let params = []

    // Role-based filtering
    if (!isAdmin.value) {
      // For non-admin users, show only tasks created by or assigned to them
      // Check both single assignment (id_user_receive) and multi-assignment (assigned_users_ids JSON)
      whereConditions.push(
        '(t.id_user_create = ? OR t.id_user_receive = ? OR JSON_CONTAINS(t.assigned_users_ids, ?))',
      )
      params.push(user.value.id, user.value.id, JSON.stringify(user.value.id))
    }

    // Apply user-defined filters
    if (filters.value.status && String(filters.value.status).trim() !== '') {
      if (filters.value.status === 'pending') {
        whereConditions.push('t.date_declare_done IS NULL')
      } else if (filters.value.status === 'completed') {
        whereConditions.push('t.date_declare_done IS NOT NULL AND t.date_confirm_done IS NULL')
      } else if (filters.value.status === 'confirmed') {
        whereConditions.push('t.date_confirm_done IS NOT NULL')
      }
    }

    if (filters.value.priority && String(filters.value.priority).trim() !== '') {
      whereConditions.push('t.id_priority = ?')
      params.push(filters.value.priority)
    }

    if (filters.value.assignedTo && String(filters.value.assignedTo).trim() !== '') {
      // Check both single assignment and multi-assignment for the filter
      whereConditions.push('(t.id_user_receive = ? OR JSON_CONTAINS(t.assigned_users_ids, ?))')
      params.push(filters.value.assignedTo, JSON.stringify(filters.value.assignedTo))
    }

    if (filters.value.createdBy && String(filters.value.createdBy).trim() !== '') {
      whereConditions.push('t.id_user_create = ?')
      params.push(filters.value.createdBy)
    }

    if (filters.value.search && String(filters.value.search).trim() !== '') {
      whereConditions.push('(t.title LIKE ? OR t.desciption LIKE ? OR t.notes LIKE ?)')
      const searchTerm = `%${filters.value.search}%`
      params.push(searchTerm, searchTerm, searchTerm)
    }

    // Combine all conditions
    const whereClause = whereConditions.length > 0 ? `WHERE ${whereConditions.join(' AND ')}` : ''

    // Main query with role-based filtering
    const result = await callApi({
      query: `
        SELECT
          t.*,
          creator.username as creator_name,
          receiver.username as receiver_name,
          p.priority as priority_name,
          p.power as priority_power,
          t.assigned_users_ids,
          t.subject_ids
        FROM tasks t
        LEFT JOIN users creator ON t.id_user_create = creator.id
        LEFT JOIN users receiver ON t.id_user_receive = receiver.id
        LEFT JOIN priorities p ON t.id_priority = p.id
        ${whereClause}
        ORDER BY t.id DESC
      `,
      params,
      requiresAuth: true,
    })

    if (result.success) {
      tasks.value = result.data
      totalItems.value = result.data.length // Temporary, will fix pagination later
    } else {
      console.error('API Error:', result)
      error.value = result.error || 'Failed to fetch tasks'
    }
  } catch (err) {
    console.error('Error fetching tasks:', err)
    error.value = 'Error fetching tasks: ' + err.message
  } finally {
    loading.value = false
  }
}

const fetchPriorities = async () => {
  try {
    const result = await callApi({
      query: 'SELECT id, priority, power FROM priorities ORDER BY power ASC',
      requiresAuth: true,
    })
    if (result.success) {
      priorities.value = result.data
    }
  } catch (err) {
    console.error('Error fetching priorities:', err)
  }
}

const fetchUsers = async () => {
  try {
    const result = await callApi({
      query: 'SELECT id, username, email FROM users ORDER BY username ASC',
      requiresAuth: true,
    })
    if (result.success) {
      users.value = result.data
    }
  } catch (err) {
    console.error('Error fetching users:', err)
  }
}

const toggleSort = (key) => {
  if (sortConfig.value.key === key) {
    sortConfig.value.direction = sortConfig.value.direction === 'asc' ? 'desc' : 'asc'
  } else {
    sortConfig.value.key = key
    sortConfig.value.direction = 'asc'
  }
  fetchTasks()
}

const applyFilters = () => {
  currentPage.value = 1
  fetchTasks()
}

const clearFilters = () => {
  filters.value = {
    status: '',
    priority: '',
    assignedTo: '',
    createdBy: '',
    search: '',
  }
  currentPage.value = 1
  fetchTasks()
}

const changePage = (page) => {
  currentPage.value = page
  fetchTasks()
}

const markTaskAsDone = async (task) => {
  if (!confirm(`Are you sure you want to mark task #${task.id} as done?`)) {
    return
  }

  try {
    const result = await callApi({
      query: `
        UPDATE tasks
        SET date_declare_done = UTC_TIMESTAMP()
        WHERE id = ?
      `,
      params: [task.id],
      requiresAuth: true,
    })

    if (result.success) {
      await fetchTasks()
      alert('Task marked as done successfully!')
    } else {
      console.error('Failed to mark task as done:', result.error)
      alert('Failed to mark task as done: ' + result.error)
    }
  } catch (err) {
    console.error('Error marking task as done:', err)
    alert('Error marking task as done: ' + err.message)
  }
}

const markTaskAsUndone = async (task) => {
  if (!confirm(`Are you sure you want to mark task #${task.id} as undone?`)) {
    return
  }

  try {
    const result = await callApi({
      query: `
        UPDATE tasks
        SET date_declare_done = NULL
        WHERE id = ?
      `,
      params: [task.id],
      requiresAuth: true,
    })

    if (result.success) {
      await fetchTasks()
      alert('Task marked as undone successfully!')
    } else {
      console.error('Failed to mark task as undone:', result.error)
      alert('Failed to mark task as undone: ' + result.error)
    }
  } catch (err) {
    console.error('Error marking task as undone:', err)
    alert('Error marking task as undone: ' + err.message)
  }
}

const confirmTaskAsDone = async (task) => {
  if (!confirm(`Are you sure you want to confirm task #${task.id} as done?`)) {
    return
  }

  try {
    const result = await callApi({
      query: `
        UPDATE tasks
        SET date_confirm_done = UTC_TIMESTAMP()
        WHERE id = ?
      `,
      params: [task.id],
      requiresAuth: true,
    })

    if (result.success) {
      await fetchTasks()
      alert('Task confirmed as done successfully!')
    } else {
      console.error('Failed to confirm task as done:', result.error)
      alert('Failed to confirm task as done: ' + result.error)
    }
  } catch (err) {
    console.error('Error confirming task as done:', err)
    alert('Error confirming task as done: ' + err.message)
  }
}

const getStatusBadge = (task) => {
  if (task.date_confirm_done) {
    return { text: 'Confirmed', class: 'status-confirmed' }
  } else if (task.date_declare_done) {
    return { text: 'Completed', class: 'status-completed' }
  } else {
    return { text: 'Pending', class: 'status-pending' }
  }
}

const getPriorityBadge = (task) => {
  const power = task.priority_power || 1
  if (power >= 4) return { text: task.priority_name, class: 'priority-high' }
  if (power >= 2) return { text: task.priority_name, class: 'priority-medium' }
  return { text: task.priority_name, class: 'priority-low' }
}

const formatDate = (dateString) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleDateString()
}

const openTaskChat = async (task) => {
  try {
    // First, check if a group with this task title already exists
    const checkGroupResult = await callApi({
      query: `
        SELECT id, name
        FROM chat_groups
        WHERE name = ? AND is_active = 1
      `,
      params: [task.title],
      requiresAuth: true,
    })

    let groupId = null

    if (checkGroupResult.success && checkGroupResult.data.length > 0) {
      // Group already exists, use it
      groupId = checkGroupResult.data[0].id
    } else {
      // Create new group with task title
      const createGroupResult = await callApi({
        query: `
          INSERT INTO chat_groups (name, description, id_user_owner, is_active)
          VALUES (?, ?, ?, 1)
        `,
        params: [task.title, `Chat group for task: ${task.title}`, user.value.id],
        requiresAuth: true,
      })

      if (createGroupResult.success && createGroupResult.lastInsertId) {
        groupId = createGroupResult.lastInsertId

        // Add task creator and receiver to the group
        const usersToAdd = []

        // Add task creator if different from current user
        if (task.id_user_create && task.id_user_create !== user.value.id) {
          usersToAdd.push(task.id_user_create)
        }

        // Add task receiver if different from current user and creator
        if (
          task.id_user_receive &&
          task.id_user_receive !== user.value.id &&
          task.id_user_receive !== task.id_user_create
        ) {
          usersToAdd.push(task.id_user_receive)
        }

        // Add current user (group owner)
        usersToAdd.push(user.value.id)

        // Remove duplicates
        const uniqueUsers = [...new Set(usersToAdd)]

        if (uniqueUsers.length > 0) {
          const insertQueries = uniqueUsers
            .map(
              () => 'INSERT INTO chat_users (id_user, id_chat_group, is_active) VALUES (?, ?, 1)',
            )
            .join('; ')

          const params = uniqueUsers.flatMap((userId) => [userId, groupId])

          const addUsersResult = await callApi({
            query: insertQueries,
            params: params,
            requiresAuth: true,
          })

          if (!addUsersResult.success) {
            console.error('Failed to add users to group:', addUsersResult)
          }
        }
      } else {
        console.error('Failed to create group:', createGroupResult)
        alert('Failed to create chat group for this task')
        return
      }
    }

    // Open chat with the group
    const chatUrl = `/mig/chat?group=${groupId}&task=${task.id}&taskName=${encodeURIComponent(task.title)}`
    window.open(chatUrl, '_blank')
  } catch (error) {
    console.error('Error opening task chat:', error)
    alert('Error opening chat for this task')
  }
}

const totalPages = computed(() => {
  return Math.ceil(totalItems.value / itemsPerPage.value)
})

// Check if user is admin
const isAdmin = computed(() => {
  return user.value?.role_id === 1
})

// Separate tasks into undone and done
const undoneTasks = computed(() => {
  return tasks.value.filter((task) => !task.date_declare_done)
})

const doneTasks = computed(() => {
  return tasks.value.filter((task) => task.date_declare_done)
})

// Format assigned users for display
const formatAssignedUsers = (task) => {
  let assignedUsers = []

  // Add single assigned user if exists
  if (task.receiver_name) {
    assignedUsers.push(task.receiver_name)
  }

  // Add multi-assigned users if exists
  if (task.assigned_users_ids) {
    try {
      const assignedIds = JSON.parse(task.assigned_users_ids)
      if (Array.isArray(assignedIds)) {
        // Find usernames for the assigned user IDs
        assignedIds.forEach((userId) => {
          const user = users.value.find((u) => u.id === userId)
          if (user && !assignedUsers.includes(user.username)) {
            assignedUsers.push(user.username)
          }
        })
      }
    } catch (err) {
      console.error('Error parsing assigned_users_ids:', err)
    }
  }

  return assignedUsers.length > 0 ? assignedUsers.join(', ') : '-'
}

// Check if current user is assigned to a task (single or multi-assign)
const isUserAssignedToTask = (task) => {
  // Check single assignment
  if (task.id_user_receive === user.value?.id) {
    return true
  }

  // Check multi-assignment
  if (task.assigned_users_ids) {
    try {
      const assignedIds = JSON.parse(task.assigned_users_ids)
      if (Array.isArray(assignedIds) && assignedIds.includes(user.value?.id)) {
        return true
      }
    } catch (err) {
      console.error('Error parsing assigned_users_ids:', err)
    }
  }

  return false
}

// TaskForm methods
const openNewTaskForm = (entityType = 'general', entityData = {}) => {
  currentEntityType.value = entityType
  currentEntityData.value = entityData
  showTaskForm.value = true
}

const handleTaskSave = async (result) => {
  if (result.success) {
    // Refresh the tasks list
    await fetchTasks()
    showTaskForm.value = false
    // You can add success notification here
    alert('Task created successfully!')
  } else {
    console.error('Failed to create task:', result)
    alert('Failed to create task. Please try again.')
  }
}

const handleTaskCancel = () => {
  showTaskForm.value = false
}

// Description popup methods
const showFullDescription = (description) => {
  selectedTaskDescription.value = description || 'No description available'
  showDescriptionPopup.value = true
}

const closeDescriptionPopup = () => {
  showDescriptionPopup.value = false
  selectedTaskDescription.value = ''
}

const deleteTask = async (task) => {
  // Check if user can delete this task
  const canDelete = isAdmin.value || task.id_user_create === user.value?.id

  if (!canDelete) {
    alert('You do not have permission to delete this task.')
    return
  }

  if (
    !confirm(
      `Are you sure you want to delete task #${task.id} "${task.title}"? This action cannot be undone.`,
    )
  ) {
    return
  }

  try {
    const result = await callApi({
      query: 'DELETE FROM tasks WHERE id = ?',
      params: [task.id],
      requiresAuth: true,
    })

    if (result.success) {
      alert('Task deleted successfully!')
      // Refresh the tasks list
      await fetchTasks()
      // Force update the tasks count in header
      window.dispatchEvent(new CustomEvent('forceUpdateTasks'))
    } else {
      console.error('Failed to delete task:', result)
      alert('Failed to delete task. Please try again.')
    }
  } catch (err) {
    console.error('Error deleting task:', err)
    alert('Error deleting task. Please try again.')
  }
}

onMounted(async () => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
  }

  await Promise.all([fetchTasks(), fetchPriorities(), fetchUsers()])
})
</script>

<template>
  <div class="tasks-view">
    <div class="header">
      <div class="header-info">
        <h1><i class="fas fa-tasks"></i> Tasks Management</h1>
        <p v-if="!isAdmin" class="view-info">
          <i class="fas fa-user"></i>
          Showing your tasks (created by you or assigned to you)
        </p>
        <p v-else class="view-info">
          <i class="fas fa-users"></i>
          Showing all tasks (admin view)
        </p>
      </div>
      <div class="header-actions">
        <button @click="openNewTaskForm()" class="btn-new-task">
          <i class="fas fa-plus"></i>
          New Task
        </button>
        <button @click="router.push('/')" class="btn-back">
          <i class="fas fa-arrow-left"></i>
          Back to Dashboard
        </button>
      </div>
    </div>

    <!-- Filters Section -->
    <div class="filters-section">
      <div class="filters-grid">
        <div class="filter-group">
          <label>Search</label>
          <input
            v-model="filters.search"
            type="text"
            placeholder="Search in title, description, notes..."
            @keyup.enter="applyFilters"
          />
        </div>

        <div class="filter-group">
          <label>Status</label>
          <select v-model="filters.status">
            <option value="">All Status</option>
            <option value="pending">Pending</option>
            <option value="completed">Completed</option>
          </select>
        </div>

        <div class="filter-group">
          <label>Priority</label>
          <select v-model="filters.priority">
            <option value="">All Priorities</option>
            <option v-for="priority in priorities" :key="priority.id" :value="priority.id">
              {{ priority.priority }}
            </option>
          </select>
        </div>

        <div v-if="isAdmin" class="filter-group">
          <label>Assigned To</label>
          <select v-model="filters.assignedTo">
            <option value="">All Users</option>
            <option v-for="user in users" :key="user.id" :value="user.id">
              {{ user.username }}
            </option>
          </select>
        </div>

        <div v-if="isAdmin" class="filter-group">
          <label>Created By</label>
          <select v-model="filters.createdBy">
            <option value="">All Users</option>
            <option v-for="user in users" :key="user.id" :value="user.id">
              {{ user.username }}
            </option>
          </select>
        </div>
      </div>

      <div class="filter-actions">
        <button @click="applyFilters" class="btn btn-primary">
          <i class="fas fa-search"></i>
          Apply Filters
        </button>
        <button @click="clearFilters" class="btn btn-secondary">
          <i class="fas fa-times"></i>
          Clear Filters
        </button>
      </div>
    </div>

    <!-- Tasks Tables -->
    <div class="tables-container">
      <!-- Undone Tasks Table -->
      <div class="table-container">
        <div class="table-header">
          <h3><i class="fas fa-clock"></i> Pending Tasks ({{ undoneTasks.length }})</h3>
        </div>

        <div v-if="loading" class="loading">
          <i class="fas fa-spinner fa-spin"></i>
          Loading tasks...
        </div>

        <div v-else-if="error" class="error">
          <i class="fas fa-exclamation-circle"></i>
          {{ error }}
        </div>

        <div v-else-if="undoneTasks.length === 0" class="empty-state">
          <i class="fas fa-check-circle fa-2x"></i>
          <p>No pending tasks</p>
        </div>

        <div v-else class="table-wrapper">
          <table class="tasks-table">
            <thead>
              <tr>
                <th @click="toggleSort('id')" class="sortable">
                  ID
                  <span v-if="sortConfig.key === 'id'" class="sort-indicator">
                    {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
                  </span>
                </th>
                <th @click="toggleSort('title')" class="sortable">
                  Title
                  <span v-if="sortConfig.key === 'title'" class="sort-indicator">
                    {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
                  </span>
                </th>
                <th>Description</th>
                <th>Priority</th>
                <th>Status</th>
                <th>Assigned To</th>
                <th>Created By</th>
                <th @click="toggleSort('date_create')" class="sortable">
                  Created Date
                  <span v-if="sortConfig.key === 'date_create'" class="sort-indicator">
                    {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
                  </span>
                </th>
                <th>Notes</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="task in undoneTasks" :key="task.id" class="task-row">
                <td>{{ task.id }}</td>
                <td class="task-title">{{ task.title }}</td>
                <td class="task-description">
                  <div class="description-content">
                    <span class="description-text">{{
                      task.description || task.desciption || task.desc || '-'
                    }}</span>
                    <button
                      v-if="task.description || task.desciption || task.desc"
                      @click="showFullDescription(task.description || task.desciption || task.desc)"
                      class="btn-view-description"
                      title="View full description"
                    >
                      <i class="fas fa-eye"></i>
                    </button>
                  </div>
                </td>
                <td>
                  <span :class="getPriorityBadge(task).class" class="priority-badge">
                    {{ getPriorityBadge(task).text }}
                  </span>
                </td>
                <td>
                  <span :class="getStatusBadge(task).class" class="status-badge">
                    {{ getStatusBadge(task).text }}
                  </span>
                </td>
                <td>{{ formatAssignedUsers(task) }}</td>
                <td>{{ task.creator_name || '-' }}</td>
                <td>{{ formatDate(task.date_create) }}</td>
                <td class="task-notes">{{ task.notes || '-' }}</td>
                <td class="actions-cell">
                  <button @click="openTaskChat(task)" class="btn-chat" title="Open task chat">
                    <i class="fas fa-comments"></i>
                    Chat
                  </button>
                  <button
                    v-if="isUserAssignedToTask(task)"
                    @click="markTaskAsDone(task)"
                    class="btn-done"
                    title="Mark as done"
                  >
                    <i class="fas fa-check"></i>
                    Done
                  </button>
                  <button
                    v-if="isAdmin || task.id_user_create === user?.id"
                    @click="deleteTask(task)"
                    class="btn-delete"
                    title="Delete task"
                  >
                    <i class="fas fa-trash"></i>
                    Delete
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Done Tasks Table -->
      <div class="table-container">
        <div class="table-header">
          <h3><i class="fas fa-check-double"></i> Completed Tasks ({{ doneTasks.length }})</h3>
        </div>

        <div v-if="doneTasks.length === 0" class="empty-state">
          <i class="fas fa-tasks fa-2x"></i>
          <p>No completed tasks</p>
        </div>

        <div v-else class="table-wrapper">
          <table class="tasks-table">
            <thead>
              <tr>
                <th @click="toggleSort('id')" class="sortable">
                  ID
                  <span v-if="sortConfig.key === 'id'" class="sort-indicator">
                    {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
                  </span>
                </th>
                <th @click="toggleSort('title')" class="sortable">
                  Title
                  <span v-if="sortConfig.key === 'title'" class="sort-indicator">
                    {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
                  </span>
                </th>
                <th>Description</th>
                <th>Priority</th>
                <th>Status</th>
                <th>Assigned To</th>
                <th>Created By</th>
                <th @click="toggleSort('date_create')" class="sortable">
                  Created Date
                  <span v-if="sortConfig.key === 'date_create'" class="sort-indicator">
                    {{ sortConfig.direction === 'asc' ? '▲' : '▼' }}
                  </span>
                </th>
                <th>Completed Date</th>
                <th>Notes</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="task in doneTasks" :key="task.id" class="task-row">
                <td>{{ task.id }}</td>
                <td class="task-title">{{ task.title }}</td>
                <td class="task-description">
                  <div class="description-content">
                    <span class="description-text">{{ task.description || '-' }}</span>
                    <button
                      v-if="task.description"
                      @click="showFullDescription(task.description)"
                      class="btn-view-description"
                      title="View full description"
                    >
                      <i class="fas fa-eye"></i>
                    </button>
                  </div>
                </td>
                <td>
                  <span :class="getPriorityBadge(task).class" class="priority-badge">
                    {{ getPriorityBadge(task).text }}
                  </span>
                </td>
                <td>
                  <span :class="getStatusBadge(task).class" class="status-badge">
                    {{ getStatusBadge(task).text }}
                  </span>
                </td>
                <td>{{ formatAssignedUsers(task) }}</td>
                <td>{{ task.creator_name || '-' }}</td>
                <td>{{ formatDate(task.date_create) }}</td>
                <td>{{ formatDate(task.date_declare_done) }}</td>
                <td class="task-notes">{{ task.notes || '-' }}</td>
                <td class="actions-cell">
                  <button
                    v-if="isAdmin && !task.date_confirm_done"
                    @click="confirmTaskAsDone(task)"
                    class="btn-confirm"
                    title="Confirm as done"
                  >
                    <i class="fas fa-check-double"></i>
                    Confirm Done
                  </button>
                  <button
                    v-if="isUserAssignedToTask(task) && !task.date_confirm_done"
                    @click="markTaskAsUndone(task)"
                    class="btn-undone"
                    title="Mark as undone"
                  >
                    <i class="fas fa-undo"></i>
                    Undone
                  </button>
                  <button
                    v-if="isAdmin || task.id_user_create === user?.id"
                    @click="deleteTask(task)"
                    class="btn-delete"
                    title="Delete task"
                  >
                    <i class="fas fa-trash"></i>
                    Delete
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Mobile Cards Container -->
    <div class="mobile-cards-container">
      <!-- Pending Tasks Cards -->
      <div class="mobile-section">
        <div class="mobile-section-header">
          <h3><i class="fas fa-clock"></i> Pending Tasks ({{ undoneTasks.length }})</h3>
        </div>

        <div v-if="loading" class="mobile-loading">
          <i class="fas fa-spinner fa-spin"></i>
          Loading tasks...
        </div>

        <div v-else-if="error" class="mobile-error">
          <i class="fas fa-exclamation-circle"></i>
          {{ error }}
        </div>

        <div v-else-if="undoneTasks.length === 0" class="mobile-empty-state">
          <i class="fas fa-check-circle fa-2x"></i>
          <p>No pending tasks</p>
        </div>

        <div v-else class="mobile-cards-grid">
          <div v-for="task in undoneTasks" :key="task.id" class="task-card">
            <div class="card-header">
              <div class="card-id">#{{ task.id }}</div>
              <div class="card-badges">
                <span :class="getPriorityBadge(task).class" class="priority-badge">
                  {{ getPriorityBadge(task).text }}
                </span>
                <span :class="getStatusBadge(task).class" class="status-badge">
                  {{ getStatusBadge(task).text }}
                </span>
              </div>
            </div>

            <div class="card-title">{{ task.title }}</div>

            <div class="card-description">
              <div class="description-content">
                <span class="description-text">{{
                  task.description || task.desciption || task.desc || 'No description'
                }}</span>
                <button
                  v-if="task.description || task.desciption || task.desc"
                  @click="showFullDescription(task.description || task.desciption || task.desc)"
                  class="btn-view-description"
                  title="View full description"
                >
                  <i class="fas fa-eye"></i>
                </button>
              </div>
            </div>

            <div class="card-details">
              <div class="detail-row">
                <div class="detail-label">Assigned To:</div>
                <div class="detail-value">{{ formatAssignedUsers(task) }}</div>
              </div>
              <div class="detail-row">
                <div class="detail-label">Created By:</div>
                <div class="detail-value">{{ task.creator_name || '-' }}</div>
              </div>
              <div class="detail-row">
                <div class="detail-label">Created Date:</div>
                <div class="detail-value">{{ formatDate(task.date_create) }}</div>
              </div>
              <div class="detail-row">
                <div class="detail-label">Notes:</div>
                <div class="detail-value">{{ task.notes || '-' }}</div>
              </div>
            </div>

            <div class="card-actions">
              <button @click="openTaskChat(task)" class="btn-chat" title="Open task chat">
                <i class="fas fa-comments"></i>
                Chat
              </button>
              <button
                v-if="isUserAssignedToTask(task)"
                @click="markTaskAsDone(task)"
                class="btn-done"
                title="Mark as done"
              >
                <i class="fas fa-check"></i>
                Done
              </button>
              <button
                v-if="isAdmin || task.id_user_create === user?.id"
                @click="deleteTask(task)"
                class="btn-delete"
                title="Delete task"
              >
                <i class="fas fa-trash"></i>
                Delete
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Completed Tasks Cards -->
      <div class="mobile-section">
        <div class="mobile-section-header">
          <h3><i class="fas fa-check-double"></i> Completed Tasks ({{ doneTasks.length }})</h3>
        </div>

        <div v-if="doneTasks.length === 0" class="mobile-empty-state">
          <i class="fas fa-tasks fa-2x"></i>
          <p>No completed tasks</p>
        </div>

        <div v-else class="mobile-cards-grid">
          <div v-for="task in doneTasks" :key="task.id" class="task-card completed">
            <div class="card-header">
              <div class="card-id">#{{ task.id }}</div>
              <div class="card-badges">
                <span :class="getPriorityBadge(task).class" class="priority-badge">
                  {{ getPriorityBadge(task).text }}
                </span>
                <span :class="getStatusBadge(task).class" class="status-badge">
                  {{ getStatusBadge(task).text }}
                </span>
              </div>
            </div>

            <div class="card-title">{{ task.title }}</div>

            <div class="card-description">
              <div class="description-content">
                <span class="description-text">{{ task.description || 'No description' }}</span>
                <button
                  v-if="task.description"
                  @click="showFullDescription(task.description)"
                  class="btn-view-description"
                  title="View full description"
                >
                  <i class="fas fa-eye"></i>
                </button>
              </div>
            </div>

            <div class="card-details">
              <div class="detail-row">
                <div class="detail-label">Assigned To:</div>
                <div class="detail-value">{{ formatAssignedUsers(task) }}</div>
              </div>
              <div class="detail-row">
                <div class="detail-label">Created By:</div>
                <div class="detail-value">{{ task.creator_name || '-' }}</div>
              </div>
              <div class="detail-row">
                <div class="detail-label">Created Date:</div>
                <div class="detail-value">{{ formatDate(task.date_create) }}</div>
              </div>
              <div class="detail-row">
                <div class="detail-label">Completed Date:</div>
                <div class="detail-value">{{ formatDate(task.date_declare_done) }}</div>
              </div>
              <div class="detail-row">
                <div class="detail-label">Notes:</div>
                <div class="detail-value">{{ task.notes || '-' }}</div>
              </div>
            </div>

            <div class="card-actions">
              <button
                v-if="isAdmin && !task.date_confirm_done"
                @click="confirmTaskAsDone(task)"
                class="btn-confirm"
                title="Confirm as done"
              >
                <i class="fas fa-check-double"></i>
                Confirm Done
              </button>
              <button
                v-if="isUserAssignedToTask(task) && !task.date_confirm_done"
                @click="markTaskAsUndone(task)"
                class="btn-undone"
                title="Mark as undone"
              >
                <i class="fas fa-undo"></i>
                Undone
              </button>
              <button
                v-if="isAdmin || task.id_user_create === user?.id"
                @click="deleteTask(task)"
                class="btn-delete"
                title="Delete task"
              >
                <i class="fas fa-trash"></i>
                Delete
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- TaskForm Modal -->
    <TaskForm
      :entity-data="currentEntityData"
      :entity-type="currentEntityType"
      :is-visible="showTaskForm"
      @task-created="handleTaskSave"
      @cancel="handleTaskCancel"
    />

    <!-- Description Popup Modal -->
    <div
      v-if="showDescriptionPopup"
      class="description-popup-overlay"
      @click="closeDescriptionPopup"
    >
      <div class="description-popup-content" @click.stop>
        <div class="description-popup-header">
          <h3><i class="fas fa-file-alt"></i> Task Description</h3>
          <button @click="closeDescriptionPopup" class="btn-close-popup">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <div class="description-popup-body">
          <p>{{ selectedTaskDescription }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.tasks-view {
  padding: 20px;
  max-width: 20000px;
  margin: auto;
}
.actions-cell {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.header-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.header h1 {
  display: flex;
  align-items: center;
  gap: 12px;
  color: #1f2937;
  margin: 0;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.btn-new-task {
  padding: 10px 16px;
  background-color: #4caf50;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: background-color 0.2s ease;
}

.btn-new-task:hover {
  background-color: #45a049;
}

.btn-back {
  padding: 10px 16px;
  background-color: #6b7280;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: background-color 0.2s ease;
}

.btn-back:hover {
  background-color: #5a6268;
}

.filters-section {
  background: white;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 24px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.filters-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  margin-bottom: 16px;
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.filter-group label {
  font-weight: 500;
  color: #374151;
  font-size: 0.9rem;
}

.filter-group input,
.filter-group select {
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 0.9rem;
}

.filter-actions {
  display: flex;
  gap: 12px;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.9rem;
  transition: all 0.2s ease;
}

.btn-primary {
  background-color: #4caf50;
  color: white;
}

.btn-primary:hover {
  background-color: #45a049;
}

.btn-secondary {
  background-color: #6b7280;
  color: white;
}

.btn-secondary:hover {
  background-color: #5a6268;
}

.tables-container {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.table-container {
  background: white;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.table-header {
  background-color: #f9fafb;
  padding: 16px 20px;
  border-bottom: 1px solid #e5e7eb;
}

.table-header h3 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #374151;
  font-size: 1.1rem;
}

.table-header h3 i {
  color: #4caf50;
}

.loading,
.error,
.empty-state {
  padding: 40px;
  text-align: center;
  color: #6b7280;
}

.loading i,
.error i,
.empty-state i {
  margin-bottom: 12px;
  color: #9ca3af;
}

.error {
  color: #dc2626;
}

.table-wrapper {
  overflow-x: auto;
}

.tasks-table {
  width: 100%;
  border-collapse: collapse;
  table-layout: auto;
}

.tasks-table th,
.tasks-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

.tasks-table th {
  background-color: #f9fafb;
  font-weight: 600;
  color: #374151;
  position: sticky;
  top: 0;
  z-index: 10;
}

.sortable {
  cursor: pointer;
  user-select: none;
}

.sortable:hover {
  background-color: #f3f4f6;
}

.sort-indicator {
  margin-left: 4px;
  color: #4caf50;
}

.task-row:hover {
  background-color: #f9fafb;
}

.task-title {
  white-space: normal;
  word-break: break-word;
  width: 600px;
  min-width: 400px;
}

.task-description,
.task-notes {
  max-width: 250px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #6b7280;
}

.priority-badge,
.status-badge {
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: 500;
  text-transform: uppercase;
}

.priority-low {
  background-color: #d1fae5;
  color: #065f46;
}

.priority-medium {
  background-color: #fef3c7;
  color: #92400e;
}

.priority-high {
  background-color: #fee2e2;
  color: #991b1b;
}

.status-pending {
  background-color: #fef3c7;
  color: #92400e;
}

.status-completed {
  background-color: #d1fae5;
  color: #065f46;
}

.status-confirmed {
  background-color: #dbeafe;
  color: #1e40af;
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 16px;
  margin-top: 24px;
  padding: 16px;
}

.btn-page {
  padding: 8px 16px;
  background-color: #f3f4f6;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: all 0.2s ease;
}

.btn-page:hover:not(:disabled) {
  background-color: #e5e7eb;
}

.btn-page:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-info {
  color: #6b7280;
  font-weight: 500;
}

@media (max-width: 768px) {
  .header {
    flex-direction: column;
    gap: 16px;
    align-items: flex-start;
  }

  .filters-grid {
    grid-template-columns: 1fr;
  }

  .filter-actions {
    flex-direction: column;
  }

  .tasks-table {
    font-size: 0.9rem;
  }

  .tasks-table th,
  .tasks-table td {
    padding: 8px;
  }
}

.view-info {
  margin: 0;
  color: #6b7280;
  font-size: 0.9rem;
  display: flex;
  align-items: center;
  gap: 6px;
}

.view-info i {
  color: #4caf50;
}

.actions-cell {
  text-align: center;
  min-width: 80px;
}

.btn-done {
  padding: 6px 12px;
  background-color: #4caf50;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.8rem;
  display: flex;
  align-items: center;
  gap: 4px;
  transition: background-color 0.2s ease;
}

.btn-done:hover {
  background-color: #45a049;
}

.btn-done i {
  font-size: 0.8rem;
}

.btn-undone {
  padding: 6px 12px;
  background-color: #6b7280;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.8rem;
  display: flex;
  align-items: center;
  gap: 4px;
  transition: background-color 0.2s ease;
}

.btn-undone:hover {
  background-color: #5a6268;
}

.btn-undone i {
  font-size: 0.8rem;
}

.btn-confirm {
  padding: 6px 12px;
  background-color: #4caf50;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.8rem;
  display: flex;
  align-items: center;
  gap: 4px;
  transition: background-color 0.2s ease;
}

.btn-confirm:hover {
  background-color: #45a049;
}

.btn-confirm i {
  font-size: 0.8rem;
}

.btn-chat {
  padding: 6px 12px;
  background-color: #4caf50;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.8rem;
  display: flex;
  align-items: center;
  gap: 4px;
  transition: background-color 0.2s ease;
}

.btn-chat:hover {
  background-color: #45a049;
}

.btn-chat i {
  font-size: 0.8rem;
}

.description-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.description-text {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.btn-view-description {
  padding: 4px 8px;
  background-color: #4caf50;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.8rem;
  display: flex;
  align-items: center;
  gap: 4px;
  transition: background-color 0.2s ease;
}

.btn-view-description:hover {
  background-color: #45a049;
}

.btn-view-description i {
  font-size: 0.8rem;
}

/* Description Popup Styles */
.description-popup-overlay {
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
}

.description-popup-content {
  background: white;
  border-radius: 12px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
  max-width: 600px;
  width: 90%;
  max-height: 80vh;
  overflow-y: auto;
  animation: modalSlideIn 0.3s ease-out;
}

.description-popup-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #e5e7eb;
}

.description-popup-header h3 {
  margin: 0;
  color: #1f2937;
  display: flex;
  align-items: center;
  gap: 8px;
}

.description-popup-header h3 i {
  color: #4caf50;
}

.btn-close-popup {
  background: none;
  border: none;
  font-size: 1.2rem;
  color: #6b7280;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.btn-close-popup:hover {
  background-color: #f3f4f6;
  color: #374151;
}

.description-popup-body {
  padding: 24px;
}

.description-popup-body p {
  margin: 0;
  line-height: 1.6;
  color: #374151;
  white-space: pre-wrap;
  word-break: break-word;
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

.btn-delete {
  padding: 6px 12px;
  background-color: #dc3545;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.8rem;
  display: flex;
  align-items: center;
  gap: 4px;
  transition: background-color 0.2s ease;
}

.btn-delete:hover {
  background-color: #c82333;
}

.btn-delete i {
  font-size: 0.8rem;
}

.mobile-cards-container {
  display: none;
}

@media (max-width: 768px) {
  .tables-container {
    display: none;
  }

  .mobile-cards-container {
    display: block;
  }

  /* Mobile Cards Styles */
  .mobile-section {
    margin-bottom: 24px;
  }

  .mobile-section-header {
    background-color: #f9fafb;
    padding: 16px 20px;
    border-radius: 8px 8px 0 0;
    border-bottom: 1px solid #e5e7eb;
    margin-bottom: 16px;
  }

  .mobile-section-header h3 {
    margin: 0;
    display: flex;
    align-items: center;
    gap: 8px;
    color: #374151;
    font-size: 1.1rem;
  }

  .mobile-section-header h3 i {
    color: #4caf50;
  }

  .mobile-loading,
  .mobile-error,
  .mobile-empty-state {
    padding: 40px 20px;
    text-align: center;
    color: #6b7280;
    background: white;
    border-radius: 8px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }

  .mobile-loading i,
  .mobile-error i,
  .mobile-empty-state i {
    margin-bottom: 12px;
    color: #9ca3af;
  }

  .mobile-error {
    color: #dc2626;
  }

  .mobile-cards-grid {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .task-card {
    background: white;
    border-radius: 12px;
    padding: 20px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    border: 1px solid #e5e7eb;
    transition: all 0.2s ease;
  }

  .task-card:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    transform: translateY(-2px);
  }

  .task-card.completed {
    opacity: 0.8;
    background-color: #f9fafb;
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 12px;
  }

  .card-id {
    font-weight: 600;
    color: #6b7280;
    font-size: 0.9rem;
  }

  .card-badges {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }

  .card-title {
    font-size: 1.1rem;
    font-weight: 600;
    color: #1f2937;
    margin-bottom: 12px;
    line-height: 1.4;
    word-break: break-word;
  }

  .card-description {
    margin-bottom: 16px;
  }

  .card-description .description-content {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 8px;
  }

  .card-description .description-text {
    flex: 1;
    color: #6b7280;
    line-height: 1.5;
    word-break: break-word;
    display: -webkit-box;
    -webkit-line-clamp: 3;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .card-details {
    margin-bottom: 16px;
  }

  .detail-row {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    padding: 8px 0;
    border-bottom: 1px solid #f3f4f6;
  }

  .detail-row:last-child {
    border-bottom: none;
  }

  .detail-label {
    font-weight: 500;
    color: #374151;
    font-size: 0.9rem;
    min-width: 100px;
  }

  .detail-value {
    color: #6b7280;
    font-size: 0.9rem;
    text-align: right;
    word-break: break-word;
    flex: 1;
  }

  .card-actions {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
    justify-content: flex-start;
  }

  .card-actions button {
    flex: 1;
    min-width: 80px;
    max-width: 120px;
    padding: 8px 12px;
    font-size: 0.85rem;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    border: none;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.2s ease;
    font-weight: 500;
  }

  .card-actions .btn-chat {
    background-color: #4caf50;
    color: white;
  }

  .card-actions .btn-chat:hover {
    background-color: #45a049;
  }

  .card-actions .btn-done {
    background-color: #4caf50;
    color: white;
  }

  .card-actions .btn-done:hover {
    background-color: #45a049;
  }

  .card-actions .btn-undone {
    background-color: #6b7280;
    color: white;
  }

  .card-actions .btn-undone:hover {
    background-color: #5a6268;
  }

  .card-actions .btn-confirm {
    background-color: #4caf50;
    color: white;
  }

  .card-actions .btn-confirm:hover {
    background-color: #45a049;
  }

  .card-actions .btn-delete {
    background-color: #dc3545;
    color: white;
  }

  .card-actions .btn-delete:hover {
    background-color: #c82333;
  }

  .card-actions .btn-view-description {
    background-color: #4caf50;
    color: white;
    padding: 6px 10px;
    font-size: 0.8rem;
  }

  .card-actions .btn-view-description:hover {
    background-color: #45a049;
  }

  /* Mobile Priority and Status Badges */
  .mobile-cards-container .priority-badge,
  .mobile-cards-container .status-badge {
    padding: 4px 8px;
    border-radius: 12px;
    font-size: 0.75rem;
    font-weight: 500;
    text-transform: uppercase;
    white-space: nowrap;
  }

  /* Responsive adjustments for very small screens */
  @media (max-width: 480px) {
    .task-card {
      padding: 16px;
    }

    .card-header {
      flex-direction: column;
      align-items: flex-start;
      gap: 8px;
    }

    .card-badges {
      width: 100%;
    }

    .detail-row {
      flex-direction: column;
      align-items: flex-start;
      gap: 4px;
    }

    .detail-label {
      min-width: auto;
    }

    .detail-value {
      text-align: left;
    }

    .card-actions {
      flex-direction: column;
    }

    .card-actions button {
      flex: none;
      max-width: none;
      width: 100%;
    }
  }
}
</style>
