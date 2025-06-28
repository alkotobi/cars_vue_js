<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useApi } from '../composables/useApi'

const router = useRouter()
const { callApi } = useApi()

const tasks = ref([])
const loading = ref(true)
const error = ref(null)
const user = ref(null)

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

    console.log('Test result:', testResult)

    if (!testResult.success) {
      error.value = 'Tasks table not accessible: ' + (testResult.error || 'Unknown error')
      return
    }

    // Build the WHERE clause based on user role
    let whereClause = ''
    let params = []

    if (!isAdmin.value) {
      // For non-admin users, show only tasks created by or assigned to them
      whereClause = 'WHERE (t.id_user_create = ? OR t.id_user_receive = ?)'
      params = [user.value.id, user.value.id]
    }

    // Main query with role-based filtering
    const result = await callApi({
      query: `
        SELECT
          t.*,
          creator.username as creator_name,
          receiver.username as receiver_name,
          p.priority as priority_name,
          p.power as priority_power
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

    console.log('Tasks result:', result)
    console.log('User role:', user.value?.role_id)
    console.log('Is admin:', isAdmin.value)
    console.log('Where clause:', whereClause)
    console.log('Params:', params)

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
  console.log('Marking task as done:', task)
  console.log('Current user ID:', user.value?.id)
  console.log('Task assigned to:', task.id_user_receive)

  if (!confirm(`Are you sure you want to mark task #${task.id} as done?`)) {
    return
  }

  try {
    console.log('Executing SQL to mark task as done...')
    const result = await callApi({
      query: `
        UPDATE tasks
        SET date_declare_done = UTC_TIMESTAMP()
        WHERE id = ?
      `,
      params: [task.id],
      requiresAuth: true,
    })

    console.log('SQL result:', result)

    if (result.success) {
      console.log('Task marked as done successfully')
      alert('Task marked as done successfully!')
      // Refresh the tasks list
      await fetchTasks()
    } else {
      console.error('Failed to mark task as done:', result)
      alert('Failed to mark task as done. Please try again.')
    }
  } catch (err) {
    console.error('Error marking task as done:', err)
    alert('Error marking task as done. Please try again.')
  }
}

const markTaskAsUndone = async (task) => {
  console.log('Marking task as undone:', task)
  console.log('Current user ID:', user.value?.id)
  console.log('Task assigned to:', task.id_user_receive)

  if (!confirm(`Are you sure you want to mark task #${task.id} as undone?`)) {
    return
  }

  try {
    console.log('Executing SQL to mark task as undone...')
    const result = await callApi({
      query: `
        UPDATE tasks
        SET date_declare_done = NULL
        WHERE id = ?
      `,
      params: [task.id],
      requiresAuth: true,
    })

    console.log('SQL result:', result)

    if (result.success) {
      console.log('Task marked as undone successfully')
      alert('Task marked as undone successfully!')
      // Refresh the tasks list
      await fetchTasks()
    } else {
      console.error('Failed to mark task as undone:', result)
      alert('Failed to mark task as undone. Please try again.')
    }
  } catch (err) {
    console.error('Error marking task as undone:', err)
    alert('Error marking task as undone. Please try again.')
  }
}

const confirmTaskAsDone = async (task) => {
  console.log('Confirming task as done:', task)
  console.log('Current user ID:', user.value?.id)
  console.log('Is admin:', isAdmin.value)

  if (
    !confirm(
      `Are you sure you want to confirm task #${task.id} as done? This will mark it as confirmed.`,
    )
  ) {
    return
  }

  try {
    console.log('Executing SQL to confirm task as done...')
    const result = await callApi({
      query: `
        UPDATE tasks
        SET date_confirm_done = UTC_TIMESTAMP(),
            id_user_confirm_done = ?
        WHERE id = ?
      `,
      params: [user.value.id, task.id],
      requiresAuth: true,
    })

    console.log('SQL result:', result)

    if (result.success) {
      console.log('Task confirmed as done successfully')
      alert('Task confirmed as done successfully!')
      // Refresh the tasks list
      await fetchTasks()
    } else {
      console.error('Failed to confirm task as done:', result)
      alert('Failed to confirm task as done. Please try again.')
    }
  } catch (err) {
    console.error('Error confirming task as done:', err)
    alert('Error confirming task as done. Please try again.')
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
  return new Date(dateString).toLocaleString()
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
      <button @click="router.push('/')" class="btn-back">
        <i class="fas fa-arrow-left"></i>
        Back to Dashboard
      </button>
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
                <td class="task-description">{{ task.desciption || '-' }}</td>
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
                <td>{{ task.receiver_name || '-' }}</td>
                <td>{{ task.creator_name || '-' }}</td>
                <td>{{ formatDate(task.date_create) }}</td>
                <td class="task-notes">{{ task.notes || '-' }}</td>
                <td class="actions-cell">
                  <button
                    v-if="task.id_user_receive === user?.id"
                    @click="markTaskAsDone(task)"
                    class="btn-done"
                    title="Mark as done"
                  >
                    <i class="fas fa-check"></i>
                    Done
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
                <td class="task-description">{{ task.desciption || '-' }}</td>
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
                <td>{{ task.receiver_name || '-' }}</td>
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
                    v-if="task.id_user_receive === user?.id"
                    @click="markTaskAsUndone(task)"
                    class="btn-undone"
                    title="Mark as undone"
                  >
                    <i class="fas fa-undo"></i>
                    Undone
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.tasks-view {
  padding: 20px;
  max-width: 1400px;
  margin: 0 auto;
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
  font-weight: 500;
  color: #1f2937;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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
</style>
