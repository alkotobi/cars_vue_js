<template>
  <div class="advanced-sql-view">
    <div class="header">
      <div class="header-content">
        <h1>
          <i class="fas fa-database"></i>
          Advanced SQL Interface
        </h1>
        <p>Execute custom SQL queries for database management and analysis</p>
      </div>
      <button
        @click="returnToDashboard"
        class="return-btn"
        :disabled="isProcessing"
        :class="{ processing: isProcessing }"
      >
        <i class="fas fa-arrow-left"></i>
        <span>Return to Dashboard</span>
        <i v-if="isProcessing" class="fas fa-spinner fa-spin loading-indicator"></i>
      </button>
    </div>

    <div class="main-content">
      <div class="sql-editor-section">
        <div class="editor-header">
          <h3>
            <i class="fas fa-code"></i>
            SQL Query Editor
          </h3>
          <div class="editor-actions">
            <button @click="clearQuery" class="btn btn-secondary" :disabled="isProcessing">
              <i class="fas fa-eraser"></i>
              Clear
            </button>
            <button
              @click="executeQuery"
              class="btn btn-primary"
              :disabled="isProcessing || !sqlQuery.trim()"
              :class="{ processing: isProcessing }"
            >
              <i class="fas fa-play"></i>
              Execute Query
              <i v-if="isProcessing" class="fas fa-spinner fa-spin loading-indicator"></i>
            </button>
          </div>
        </div>

        <div class="query-input-container">
          <textarea
            v-model="sqlQuery"
            placeholder="Enter your SQL query here..."
            class="sql-textarea"
            :disabled="isProcessing"
            @keydown.ctrl.enter="executeQuery"
            @keydown.meta.enter="executeQuery"
          ></textarea>
          <div class="query-info">
            <span class="query-tip">
              <i class="fas fa-info-circle"></i>
              Press Ctrl+Enter (or Cmd+Enter) to execute
            </span>
            <span class="query-length">{{ sqlQuery.length }} characters</span>
          </div>
        </div>

        <div class="safety-warning" v-if="isDangerousQuery">
          <div class="warning-content">
            <i class="fas fa-exclamation-triangle"></i>
            <div class="warning-text">
              <strong>Warning:</strong> This query may modify data. Please review carefully before
              executing.
            </div>
          </div>
        </div>
      </div>

      <div class="results-section" v-if="queryResults || error">
        <div class="results-header">
          <h3>
            <i class="fas fa-table"></i>
            Query Results
          </h3>
          <div class="results-actions" v-if="queryResults">
            <button @click="exportToCSV" class="btn btn-secondary" :disabled="isProcessing">
              <i class="fas fa-download"></i>
              Export CSV
            </button>
            <button @click="copyResults" class="btn btn-secondary" :disabled="isProcessing">
              <i class="fas fa-copy"></i>
              Copy Results
            </button>
          </div>
        </div>

        <div v-if="error" class="error-message">
          <div class="error-content">
            <i class="fas fa-times-circle"></i>
            <div class="error-text"><strong>Error:</strong> {{ error }}</div>
          </div>
        </div>

        <div v-if="queryResults" class="results-container">
          <div class="results-info">
            <span class="result-count">
              <i class="fas fa-list"></i>
              {{ queryResults.length }} rows returned
            </span>
            <span class="execution-time" v-if="executionTime">
              <i class="fas fa-clock"></i>
              {{ executionTime }}ms
            </span>
          </div>

          <div class="table-container">
            <table class="results-table" v-if="queryResults.length > 0">
              <thead>
                <tr>
                  <th
                    v-for="column in Object.keys(queryResults[0])"
                    :key="column"
                    class="table-header"
                  >
                    {{ column }}
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(row, index) in queryResults" :key="index" class="table-row">
                  <td
                    v-for="column in Object.keys(queryResults[0])"
                    :key="column"
                    class="table-cell"
                  >
                    <span v-if="row[column] === null" class="null-value">NULL</span>
                    <span v-else-if="typeof row[column] === 'boolean'">
                      {{ row[column] ? 'Yes' : 'No' }}
                    </span>
                    <span v-else>{{ row[column] }}</span>
                  </td>
                </tr>
              </tbody>
            </table>
            <div v-else class="no-results">
              <i class="fas fa-inbox"></i>
              <p>No results returned</p>
            </div>
          </div>
        </div>
      </div>

      <div class="query-history-section" v-if="queryHistory.length > 0">
        <div class="history-header">
          <h3>
            <i class="fas fa-history"></i>
            Query History
          </h3>
          <button @click="clearHistory" class="btn btn-secondary" :disabled="isProcessing">
            <i class="fas fa-trash"></i>
            Clear History
          </button>
        </div>

        <div class="history-list">
          <div
            v-for="(item, index) in queryHistory"
            :key="index"
            class="history-item"
            @click="loadQuery(item.query)"
          >
            <div class="history-query">
              <span class="query-preview"
                >{{ item.query.substring(0, 100) }}{{ item.query.length > 100 ? '...' : '' }}</span
              >
            </div>
            <div class="history-meta">
              <span class="history-time">{{ formatTime(item.timestamp) }}</span>
              <span class="history-status" :class="item.success ? 'success' : 'error'">
                {{ item.success ? 'Success' : 'Error' }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'

const router = useRouter()
const loading = ref(false)
const isProcessing = ref(false)
const sqlQuery = ref('')
const queryResults = ref(null)
const error = ref(null)
const executionTime = ref(null)
const queryHistory = ref([])

// Check if user is admin
const currentUser = computed(() => {
  const userStr = localStorage.getItem('user')
  return userStr ? JSON.parse(userStr) : null
})

const isAdmin = computed(() => {
  return currentUser.value?.role_id === 1
})

// Check if query is potentially dangerous
const isDangerousQuery = computed(() => {
  const query = sqlQuery.value.toLowerCase()
  const dangerousKeywords = [
    'delete',
    'drop',
    'truncate',
    'alter',
    'create',
    'insert',
    'update',
    'grant',
    'revoke',
    'backup',
    'restore',
  ]
  return dangerousKeywords.some((keyword) => query.includes(keyword))
})

// Load query history from localStorage
onMounted(() => {
  if (!isAdmin.value) {
    router.push('/dashboard')
    return
  }

  const savedHistory = localStorage.getItem('sqlQueryHistory')
  if (savedHistory) {
    try {
      queryHistory.value = JSON.parse(savedHistory)
    } catch (e) {
      console.error('Error loading query history:', e)
    }
  }
})

const returnToDashboard = async () => {
  if (isProcessing.value) return
  isProcessing.value = true
  try {
    await router.push('/dashboard')
  } finally {
    isProcessing.value = false
  }
}

const clearQuery = () => {
  sqlQuery.value = ''
  queryResults.value = null
  error.value = null
  executionTime.value = null
}

const executeQuery = async () => {
  if (isProcessing.value || !sqlQuery.value.trim()) return

  isProcessing.value = true
  error.value = null
  queryResults.value = null
  executionTime.value = null

  const startTime = Date.now()

  try {
    const response = await fetch('/api/api.php', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        action: 'execute_sql',
        query: sqlQuery.value,
      }),
    })

    const data = await response.json()

    if (data.success) {
      queryResults.value = data.results || []
      executionTime.value = Date.now() - startTime

      // Add to history
      addToHistory(sqlQuery.value, true)

      ElMessage.success(`Query executed successfully. ${queryResults.value.length} rows returned.`)
    } else {
      error.value = data.message || 'An error occurred while executing the query.'
      addToHistory(sqlQuery.value, false)
      ElMessage.error(error.value)
    }
  } catch (err) {
    error.value = 'Network error: Unable to execute query.'
    addToHistory(sqlQuery.value, false)
    ElMessage.error(error.value)
  } finally {
    isProcessing.value = false
  }
}

const addToHistory = (query, success) => {
  const historyItem = {
    query: query,
    timestamp: new Date().toISOString(),
    success: success,
  }

  queryHistory.value.unshift(historyItem)

  // Keep only last 20 queries
  if (queryHistory.value.length > 20) {
    queryHistory.value = queryHistory.value.slice(0, 20)
  }

  // Save to localStorage
  localStorage.setItem('sqlQueryHistory', JSON.stringify(queryHistory.value))
}

const loadQuery = (query) => {
  sqlQuery.value = query
  queryResults.value = null
  error.value = null
  executionTime.value = null
}

const clearHistory = () => {
  queryHistory.value = []
  localStorage.removeItem('sqlQueryHistory')
  ElMessage.success('Query history cleared.')
}

const exportToCSV = () => {
  if (!queryResults.value || queryResults.value.length === 0) return

  const headers = Object.keys(queryResults.value[0])
  const csvContent = [
    headers.join(','),
    ...queryResults.value.map((row) =>
      headers
        .map((header) => {
          const value = row[header]
          if (value === null || value === undefined) return ''
          return `"${String(value).replace(/"/g, '""')}"`
        })
        .join(','),
    ),
  ].join('\n')

  const blob = new Blob([csvContent], { type: 'text/csv' })
  const url = window.URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `sql_results_${new Date().toISOString().split('T')[0]}.csv`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  window.URL.revokeObjectURL(url)

  ElMessage.success('Results exported to CSV.')
}

const copyResults = async () => {
  if (!queryResults.value || queryResults.value.length === 0) return

  try {
    const headers = Object.keys(queryResults.value[0])
    const textContent = [
      headers.join('\t'),
      ...queryResults.value.map((row) => headers.map((header) => row[header] || '').join('\t')),
    ].join('\n')

    await navigator.clipboard.writeText(textContent)
    ElMessage.success('Results copied to clipboard.')
  } catch (err) {
    ElMessage.error('Failed to copy results.')
  }
}

const formatTime = (timestamp) => {
  const date = new Date(timestamp)
  return date.toLocaleString()
}
</script>

<style scoped>
.advanced-sql-view {
  min-height: 100vh;
  background-color: #f9fafb;
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
  padding: 20px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.header-content h1 {
  margin: 0 0 8px 0;
  color: #1f2937;
  font-size: 28px;
  font-weight: 700;
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-content p {
  margin: 0;
  color: #6b7280;
  font-size: 16px;
}

.return-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.return-btn:hover:not(:disabled) {
  background: #2563eb;
  transform: translateY(-1px);
}

.return-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.return-btn.processing {
  background: #6b7280;
}

.loading-indicator {
  margin-left: 8px;
}

.main-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.sql-editor-section {
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #e5e7eb;
  background: #f9fafb;
}

.editor-header h3 {
  margin: 0;
  color: #1f2937;
  font-size: 18px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
}

.editor-actions {
  display: flex;
  gap: 12px;
}

.btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: #10b981;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #059669;
  transform: translateY(-1px);
}

.btn-secondary {
  background: #6b7280;
  color: white;
}

.btn-secondary:hover:not(:disabled) {
  background: #4b5563;
  transform: translateY(-1px);
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn.processing {
  background: #6b7280;
}

.query-input-container {
  padding: 20px;
}

.sql-textarea {
  width: 100%;
  min-height: 200px;
  padding: 16px;
  border: 2px solid #e5e7eb;
  border-radius: 8px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 14px;
  line-height: 1.5;
  resize: vertical;
  transition: border-color 0.2s;
}

.sql-textarea:focus {
  outline: none;
  border-color: #3b82f6;
}

.sql-textarea:disabled {
  background: #f9fafb;
  cursor: not-allowed;
}

.query-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12px;
  font-size: 12px;
  color: #6b7280;
}

.query-tip {
  display: flex;
  align-items: center;
  gap: 4px;
}

.query-length {
  font-weight: 500;
}

.safety-warning {
  margin: 0 20px 20px 20px;
  padding: 16px;
  background: #fef3c7;
  border: 1px solid #f59e0b;
  border-radius: 8px;
}

.warning-content {
  display: flex;
  align-items: center;
  gap: 12px;
  color: #92400e;
}

.warning-content i {
  font-size: 18px;
  color: #f59e0b;
}

.results-section {
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.results-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #e5e7eb;
  background: #f9fafb;
}

.results-header h3 {
  margin: 0;
  color: #1f2937;
  font-size: 18px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
}

.results-actions {
  display: flex;
  gap: 12px;
}

.error-message {
  margin: 20px;
  padding: 16px;
  background: #fef2f2;
  border: 1px solid #f87171;
  border-radius: 8px;
}

.error-content {
  display: flex;
  align-items: center;
  gap: 12px;
  color: #991b1b;
}

.error-content i {
  font-size: 18px;
  color: #ef4444;
}

.results-container {
  padding: 20px;
}

.results-info {
  display: flex;
  gap: 20px;
  margin-bottom: 16px;
  font-size: 14px;
  color: #6b7280;
}

.result-count,
.execution-time {
  display: flex;
  align-items: center;
  gap: 6px;
}

.table-container {
  overflow-x: auto;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
}

.results-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}

.table-header {
  background: #f9fafb;
  padding: 12px 16px;
  text-align: left;
  font-weight: 600;
  color: #374151;
  border-bottom: 1px solid #e5e7eb;
  white-space: nowrap;
}

.table-row:nth-child(even) {
  background: #f9fafb;
}

.table-cell {
  padding: 12px 16px;
  border-bottom: 1px solid #e5e7eb;
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.null-value {
  color: #9ca3af;
  font-style: italic;
}

.no-results {
  padding: 40px;
  text-align: center;
  color: #6b7280;
}

.no-results i {
  font-size: 48px;
  margin-bottom: 16px;
  opacity: 0.5;
}

.query-history-section {
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #e5e7eb;
  background: #f9fafb;
}

.history-header h3 {
  margin: 0;
  color: #1f2937;
  font-size: 18px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
}

.history-list {
  max-height: 400px;
  overflow-y: auto;
}

.history-item {
  padding: 16px 20px;
  border-bottom: 1px solid #e5e7eb;
  cursor: pointer;
  transition: background-color 0.2s;
}

.history-item:hover {
  background: #f9fafb;
}

.history-item:last-child {
  border-bottom: none;
}

.history-query {
  margin-bottom: 8px;
}

.query-preview {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  color: #374151;
  line-height: 1.4;
}

.history-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
}

.history-time {
  color: #6b7280;
}

.history-status {
  padding: 2px 8px;
  border-radius: 12px;
  font-weight: 500;
  font-size: 11px;
}

.history-status.success {
  background: #d1fae5;
  color: #065f46;
}

.history-status.error {
  background: #fee2e2;
  color: #991b1b;
}

@media (max-width: 768px) {
  .advanced-sql-view {
    padding: 10px;
  }

  .header {
    flex-direction: column;
    gap: 16px;
    align-items: stretch;
  }

  .editor-header,
  .results-header,
  .history-header {
    flex-direction: column;
    gap: 12px;
    align-items: stretch;
  }

  .editor-actions,
  .results-actions {
    justify-content: stretch;
  }

  .btn {
    flex: 1;
    justify-content: center;
  }

  .results-info {
    flex-direction: column;
    gap: 8px;
  }

  .history-meta {
    flex-direction: column;
    gap: 4px;
    align-items: flex-start;
  }
}
</style>
