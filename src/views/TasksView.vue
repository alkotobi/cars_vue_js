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

    // Simplified query to debug the issue
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
        ORDER BY t.id DESC
      `,
      params: [],
      requiresAuth: true,
    })

    console.log('Tasks result:', result)

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
      <h1><i class="fas fa-tasks"></i> Tasks Management</h1>
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

        <div class="filter-group">
          <label>Assigned To</label>
          <select v-model="filters.assignedTo">
            <option value="">All Users</option>
            <option v-for="user in users" :key="user.id" :value="user.id">
              {{ user.username }}
            </option>
          </select>
        </div>

        <div class="filter-group">
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

    <!-- Tasks Table -->
    <div class="table-container">
      <div v-if="loading" class="loading">
        <i class="fas fa-spinner fa-spin"></i>
        Loading tasks...
      </div>

      <div v-else-if="error" class="error">
        <i class="fas fa-exclamation-circle"></i>
        {{ error }}
      </div>

      <div v-else-if="tasks.length === 0" class="empty-state">
        <i class="fas fa-tasks fa-2x"></i>
        <p>No tasks found</p>
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
            </tr>
          </thead>
          <tbody>
            <tr v-for="task in tasks" :key="task.id" class="task-row">
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
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1" class="pagination">
      <button @click="changePage(currentPage - 1)" :disabled="currentPage === 1" class="btn-page">
        <i class="fas fa-chevron-left"></i>
        Previous
      </button>

      <span class="page-info"> Page {{ currentPage }} of {{ totalPages }} </span>

      <button
        @click="changePage(currentPage + 1)"
        :disabled="currentPage === totalPages"
        class="btn-page"
      >
        Next
        <i class="fas fa-chevron-right"></i>
      </button>
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

.table-container {
  background: white;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
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
</style>
