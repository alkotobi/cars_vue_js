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
            <button
              @click="toggleParameterMode"
              class="btn"
              :class="parameterMode ? 'btn-primary' : 'btn-secondary'"
              :disabled="isProcessing"
            >
              <i class="fas fa-sliders-h"></i>
              {{ parameterMode ? 'Simple Mode' : 'Parameter Mode' }}
            </button>
            <button @click="clearQuery" class="btn btn-secondary" :disabled="isProcessing">
              <i class="fas fa-eraser"></i>
              Clear
            </button>
            <button
              @click="executeQuery"
              class="btn btn-primary"
              :disabled="isProcessing || !sqlQuery.trim() || !areParametersValid"
              :class="{ processing: isProcessing, 'requires-confirmation': isNonSelectQuery }"
            >
              <i class="fas fa-play"></i>
              <span v-if="isNonSelectQuery">Execute (Confirmation Required)</span>
              <span v-else>Execute Query</span>
              <i v-if="isProcessing" class="fas fa-spinner fa-spin loading-indicator"></i>
            </button>
          </div>
        </div>

        <div class="query-input-container">
          <textarea
            v-model="sqlQuery"
            :placeholder="
              parameterMode
                ? 'Enter your SQL query with parameters like :param1, :param2...'
                : 'Enter your SQL query here...'
            "
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
            <span v-if="hasMultipleStatements" class="multi-statement-indicator">
              <i class="fas fa-layer-group"></i>
              Multi-statement query ({{ statementCount }} statements)
            </span>
            <span class="query-length">{{ sqlQuery.length }} characters</span>
          </div>
        </div>

        <!-- Parameter Input Section -->
        <div v-if="parameterMode && detectedParameters.length > 0" class="parameters-section">
          <div class="parameters-header">
            <h4>
              <i class="fas fa-sliders-h"></i>
              Query Parameters ({{ detectedParameters.length }})
              <span class="required-indicator">* Required</span>
              <span v-if="hasLoadedParameters" class="loaded-indicator">
                <i class="fas fa-check-circle"></i>
                Loaded from template
              </span>
            </h4>
            <button
              @click="clearParameters"
              class="btn btn-secondary btn-sm"
              :disabled="isProcessing"
            >
              <i class="fas fa-eraser"></i>
              Clear All
            </button>
          </div>
          <div class="parameters-grid">
            <div v-for="param in detectedParameters" :key="param" class="parameter-input">
              <label :for="`param-${param}`" class="parameter-label">
                {{ param }}
                <span class="required-star">*</span>
              </label>
              <input
                :id="`param-${param}`"
                v-model="parameterValues[param]"
                type="text"
                class="parameter-field"
                :class="{ error: isParameterError(param) }"
                :placeholder="`Enter value for ${param}`"
                :disabled="isProcessing"
                required
              />
              <div v-if="isParameterError(param)" class="parameter-error">
                This parameter is required
              </div>
            </div>
          </div>
        </div>

        <!-- Query Templates Section -->
        <div v-if="parameterMode" class="templates-section">
          <div class="templates-header">
            <h4>
              <i class="fas fa-bookmark"></i>
              Query Templates
            </h4>
            <div class="template-actions">
              <button
                @click="saveAsTemplate"
                class="btn btn-secondary btn-sm"
                :disabled="isProcessing || !sqlQuery.trim()"
              >
                <i class="fas fa-save"></i>
                Save Template
              </button>
              <button
                @click="showTemplatesDialog = true"
                class="btn btn-secondary btn-sm"
                :disabled="isProcessing"
              >
                <i class="fas fa-folder-open"></i>
                Load Template
              </button>
            </div>
          </div>
        </div>

        <div class="safety-warning" v-if="isDangerousQuery || isNonSelectQuery">
          <div class="warning-content">
            <i class="fas fa-exclamation-triangle"></i>
            <div class="warning-text">
              <strong>Warning:</strong>
              <span v-if="isNonSelectQuery">
                This is a {{ queryType }} query that will modify data in the database. You will be
                asked to confirm before execution.
              </span>
              <span v-else>
                This query may modify data. Please review carefully before executing.
              </span>
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

    <!-- Templates Dialog -->
    <div v-if="showTemplatesDialog" class="dialog-overlay" @click="showTemplatesDialog = false">
      <div class="dialog-content" @click.stop>
        <div class="dialog-header">
          <h3>
            <i class="fas fa-bookmark"></i>
            Saved Query Templates
          </h3>
          <button @click="showTemplatesDialog = false" class="dialog-close">
            <i class="fas fa-times"></i>
          </button>
        </div>

        <!-- Filters Section -->
        <div class="dialog-filters">
          <div class="filter-row">
            <div class="filter-group">
              <label for="template-search" class="filter-label">
                <i class="fas fa-search"></i>
                Search
              </label>
              <input
                id="template-search"
                v-model="templateFilters.search"
                type="text"
                class="filter-input"
                placeholder="Search by name, description, or query..."
              />
            </div>

            <div class="filter-group">
              <label for="query-type-filter" class="filter-label">
                <i class="fas fa-filter"></i>
                Query Type
              </label>
              <select
                id="query-type-filter"
                v-model="templateFilters.queryType"
                class="filter-select"
              >
                <option value="">All Types</option>
                <option value="SELECT">SELECT</option>
                <option value="INSERT">INSERT</option>
                <option value="UPDATE">UPDATE</option>
                <option value="DELETE">DELETE</option>
                <option value="CREATE">CREATE</option>
                <option value="ALTER">ALTER</option>
                <option value="DROP">DROP</option>
                <option value="OTHER">Other</option>
              </select>
            </div>

            <div class="filter-group">
              <label for="sort-by" class="filter-label">
                <i class="fas fa-sort"></i>
                Sort By
              </label>
              <select id="sort-by" v-model="templateFilters.sortBy" class="filter-select">
                <option value="newest">Newest First</option>
                <option value="oldest">Oldest First</option>
                <option value="name">Name A-Z</option>
                <option value="name-desc">Name Z-A</option>
                <option value="type">Query Type</option>
              </select>
            </div>
          </div>

          <div class="filter-actions">
            <button
              @click="clearTemplateFilters"
              class="btn btn-secondary btn-sm"
              :disabled="!hasActiveFilters"
            >
              <i class="fas fa-eraser"></i>
              Clear Filters
            </button>
            <span class="filter-count">
              {{ filteredTemplates.length }} of {{ savedTemplates.length }} templates
            </span>
          </div>
        </div>

        <div class="dialog-body">
          <div v-if="savedTemplates.length === 0" class="empty-templates">
            <i class="fas fa-bookmark fa-2x"></i>
            <p>No saved templates yet</p>
          </div>
          <div v-else-if="filteredTemplates.length === 0" class="empty-templates">
            <i class="fas fa-search fa-2x"></i>
            <p>No templates match your filters</p>
            <button @click="clearTemplateFilters" class="btn btn-primary btn-sm">
              Clear Filters
            </button>
          </div>
          <div v-else class="templates-list">
            <div v-for="template in filteredTemplates" :key="template.id" class="template-item">
              <div class="template-info">
                <div class="template-header">
                  <h4>{{ template.name }}</h4>
                  <span class="template-type" :class="getQueryTypeClass(template.query)">
                    {{ getQueryType(template.query) }}
                  </span>
                </div>
                <p class="template-description">{{ template.description }}</p>
                <p class="template-query">
                  {{ template.query.substring(0, 80) }}{{ template.query.length > 80 ? '...' : '' }}
                </p>
                <div
                  v-if="template.paramValues && Object.keys(template.paramValues).length > 0"
                  class="template-params"
                >
                  <strong>Saved Parameters:</strong>
                  <div class="param-values">
                    <span
                      v-for="(value, param) in template.paramValues"
                      :key="param"
                      class="param-value"
                    >
                      {{ param }}: {{ value }}
                    </span>
                  </div>
                </div>
              </div>
              <div class="template-actions">
                <button @click="loadTemplate(template)" class="btn btn-primary btn-sm">
                  <i class="fas fa-edit"></i>
                  Load
                </button>
                <button @click="deleteTemplate(template.id)" class="btn btn-secondary btn-sm">
                  <i class="fas fa-trash"></i>
                  Delete
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Save Template Dialog -->
    <div v-if="showSaveDialog" class="dialog-overlay" @click="showSaveDialog = false">
      <div class="dialog-content save-dialog" @click.stop>
        <div class="dialog-header">
          <h3>
            <i class="fas fa-save"></i>
            Save Query Template
          </h3>
          <button @click="showSaveDialog = false" class="dialog-close">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <div class="dialog-body">
          <form @submit.prevent="saveTemplate" class="save-form">
            <div class="form-group">
              <label for="template-name" class="form-label">
                Template Name <span class="required-star">*</span>
              </label>
              <input
                id="template-name"
                v-model="saveForm.name"
                type="text"
                class="form-input"
                :class="{ error: saveFormErrors.name }"
                placeholder="Enter template name"
                required
              />
              <div v-if="saveFormErrors.name" class="form-error">
                {{ saveFormErrors.name }}
              </div>
            </div>

            <div class="form-group">
              <label for="template-description" class="form-label"> Description </label>
              <textarea
                id="template-description"
                v-model="saveForm.description"
                class="form-textarea"
                placeholder="Enter template description (optional)"
                rows="3"
              ></textarea>
            </div>

            <div class="form-group">
              <label class="form-label">SQL Query</label>
              <div class="query-preview">
                <code>{{ sqlQuery }}</code>
              </div>
            </div>

            <div v-if="detectedParameters.length > 0" class="form-group">
              <label class="form-label">Parameters</label>
              <div class="parameters-preview">
                <div v-for="param in detectedParameters" :key="param" class="param-preview">
                  <span class="param-name">{{ param }}</span>
                  <span class="param-value" v-if="parameterValues[param]">
                    = {{ parameterValues[param] }}
                  </span>
                  <span class="param-empty" v-else> (not set) </span>
                </div>
              </div>
            </div>

            <div class="form-actions">
              <button
                type="button"
                @click="showSaveDialog = false"
                class="btn btn-secondary"
                :disabled="isSaving"
              >
                Cancel
              </button>
              <button
                type="submit"
                class="btn btn-primary"
                :disabled="isSaving || !saveForm.name.trim()"
                :class="{ processing: isSaving }"
              >
                <i class="fas fa-save"></i>
                Save Template
                <i v-if="isSaving" class="fas fa-spinner fa-spin loading-indicator"></i>
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useApi } from '../composables/useApi.js'

const router = useRouter()
const { callApi, loading: apiLoading, error: apiError } = useApi()
const loading = ref(false)
const isProcessing = ref(false)
const sqlQuery = ref('')
const queryResults = ref(null)
const error = ref(null)
const executionTime = ref(null)
const queryHistory = ref([])
const parameterMode = ref(false)
const parameterValues = ref({})
const showTemplatesDialog = ref(false)
const savedTemplates = ref([])
const showSaveDialog = ref(false)
const isSaving = ref(false)
const saveForm = ref({
  name: '',
  description: '',
})
const saveFormErrors = ref({})
const templateFilters = ref({
  search: '',
  queryType: '',
  sortBy: 'newest',
})

// Check if user is admin
const currentUser = computed(() => {
  const userStr = localStorage.getItem('user')
  return userStr ? JSON.parse(userStr) : null
})

const isAdmin = computed(() => {
  return currentUser.value?.role_id === 1
})

// Detect parameters in SQL query
const detectedParameters = computed(() => {
  if (!sqlQuery.value) return []

  const paramRegex = /:([a-zA-Z_][a-zA-Z0-9_]*)/g
  const matches = sqlQuery.value.match(paramRegex) || []
  const uniqueParams = [...new Set(matches.map((match) => match.substring(1)))]

  return uniqueParams
})

// Check if parameter has error (empty or whitespace)
const isParameterError = (param) => {
  return (
    parameterMode.value &&
    detectedParameters.value.includes(param) &&
    (!parameterValues.value[param] || parameterValues.value[param].trim() === '')
  )
}

// Check if all parameters are valid
const areParametersValid = computed(() => {
  if (!parameterMode.value || detectedParameters.value.length === 0) return true
  return detectedParameters.value.every(
    (param) => parameterValues.value[param] && parameterValues.value[param].trim() !== '',
  )
})

// Check if query has multiple statements
const hasMultipleStatements = computed(() => {
  return splitSqlStatements(sqlQuery.value).length > 1
})

// Get statement count
const statementCount = computed(() => {
  return splitSqlStatements(sqlQuery.value).length
})

// Check if parameters are loaded from a template
const hasLoadedParameters = computed(() => {
  if (!parameterMode.value || detectedParameters.value.length === 0) return false
  return detectedParameters.value.some(
    (param) => parameterValues.value[param] && parameterValues.value[param].trim() !== '',
  )
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

// Check if query is a non-SELECT query
const isNonSelectQuery = computed(() => {
  if (!sqlQuery.value.trim()) return false

  const queryType = sqlQuery.value.trim().toUpperCase()
  return (
    !queryType.startsWith('SELECT') &&
    (queryType.startsWith('INSERT') ||
      queryType.startsWith('UPDATE') ||
      queryType.startsWith('DELETE') ||
      queryType.startsWith('DROP') ||
      queryType.startsWith('CREATE') ||
      queryType.startsWith('ALTER') ||
      queryType.startsWith('TRUNCATE'))
  )
})

// Get query type for display
const queryType = computed(() => {
  if (!sqlQuery.value.trim()) return ''

  const query = sqlQuery.value.trim().toUpperCase()
  if (query.startsWith('SELECT')) return 'SELECT'
  if (query.startsWith('INSERT')) return 'INSERT'
  if (query.startsWith('UPDATE')) return 'UPDATE'
  if (query.startsWith('DELETE')) return 'DELETE'
  if (query.startsWith('DROP')) return 'DROP'
  if (query.startsWith('CREATE')) return 'CREATE'
  if (query.startsWith('ALTER')) return 'ALTER'
  if (query.startsWith('TRUNCATE')) return 'TRUNCATE'
  return 'OTHER'
})

// Watch for parameter changes and update parameterValues
watch(
  detectedParameters,
  (newParams) => {
    // Remove parameters that no longer exist
    Object.keys(parameterValues.value).forEach((param) => {
      if (!newParams.includes(param)) {
        delete parameterValues.value[param]
      }
    })

    // Add new parameters with empty values
    newParams.forEach((param) => {
      if (!(param in parameterValues.value)) {
        parameterValues.value[param] = ''
      }
    })
  },
  { immediate: true },
)

// Load data from localStorage and database
onMounted(async () => {
  if (!isAdmin.value) {
    router.push('/dashboard')
    return
  }

  // Load query history from localStorage
  const savedHistory = localStorage.getItem('sqlQueryHistory')
  if (savedHistory) {
    try {
      queryHistory.value = JSON.parse(savedHistory)
    } catch (e) {
      console.error('Error loading query history:', e)
    }
  }

  // Load templates from database
  await loadTemplatesFromDatabase()
})

const loadTemplatesFromDatabase = async () => {
  try {
    const result = await callApi({
      query: 'SELECT id, name, stmt, `desc`, params, param_values FROM adv_sql ORDER BY id DESC',
      requiresAuth: true,
    })

    if (result.success && result.data) {
      savedTemplates.value = result.data.map((template) => {
        const params = template.params ? JSON.parse(template.params) : []
        const paramValues = template.param_values ? JSON.parse(template.param_values) : {}

        return {
          id: template.id.toString(),
          name: template.name || `Template ${template.id}`,
          description: template.desc || '',
          query: template.stmt,
          params: params,
          paramValues: paramValues,
          created: new Date().toISOString(), // We don't have created date in DB, so use current time
        }
      })
    }
  } catch (err) {
    console.error('Error loading templates from database:', err)
    ElMessage.error('Failed to load templates from database')
  }
}

const returnToDashboard = async () => {
  if (isProcessing.value) return
  isProcessing.value = true
  try {
    await router.push('/dashboard')
  } finally {
    isProcessing.value = false
  }
}

const toggleParameterMode = () => {
  parameterMode.value = !parameterMode.value
  if (!parameterMode.value) {
    parameterValues.value = {}
  }
}

const clearQuery = () => {
  sqlQuery.value = ''
  queryResults.value = null
  error.value = null
  executionTime.value = null
  parameterValues.value = {}
}

const clearParameters = () => {
  parameterValues.value = {}
  Object.keys(parameterValues.value).forEach((key) => {
    parameterValues.value[key] = ''
  })
}

const executeQuery = async () => {
  if (isProcessing.value || !sqlQuery.value.trim()) return

  // Check if this is a multi-statement query
  const statements = splitSqlStatements(sqlQuery.value)
  const hasMultipleStatements = statements.length > 1

  // Check if any statement is a non-SELECT query and ask for confirmation
  const hasNonSelectStatements = statements.some((statement) => {
    const queryType = statement.trim().toUpperCase()
    return (
      !queryType.startsWith('SELECT') &&
      (queryType.startsWith('INSERT') ||
        queryType.startsWith('UPDATE') ||
        queryType.startsWith('DELETE') ||
        queryType.startsWith('DROP') ||
        queryType.startsWith('CREATE') ||
        queryType.startsWith('ALTER') ||
        queryType.startsWith('TRUNCATE'))
    )
  })

  if (hasNonSelectStatements) {
    const message = hasMultipleStatements
      ? `This query contains ${statements.length} statements, some of which will modify data in the database. Are you sure you want to execute all statements?`
      : `This query will modify data in the database. Are you sure you want to execute this statement?`

    try {
      await ElMessageBox.confirm(message, 'Confirm Database Modification', {
        confirmButtonText: 'Execute',
        cancelButtonText: 'Cancel',
        type: 'warning',
        dangerouslyUseHTMLString: false,
      })
    } catch (err) {
      // User cancelled
      return
    }
  }

  isProcessing.value = true
  error.value = null
  queryResults.value = null
  executionTime.value = null

  const startTime = Date.now()

  try {
    let finalQuery = sqlQuery.value
    let params = []

    // If in parameter mode, replace parameters with values
    if (parameterMode.value && detectedParameters.value.length > 0) {
      // Check if all parameters have values
      const missingParams = detectedParameters.value.filter(
        (param) => !parameterValues.value[param] || parameterValues.value[param].trim() === '',
      )

      if (missingParams.length > 0) {
        error.value = `Missing parameter values: ${missingParams.join(', ')}`
        ElMessage.error(error.value)
        return
      }

      // For multi-statement queries, we need to handle parameters differently
      const statements = splitSqlStatements(sqlQuery.value)
      const hasMultipleStatements = statements.length > 1

      if (hasMultipleStatements) {
        // For multi-statement queries, replace parameters with actual values
        detectedParameters.value.forEach((param) => {
          const paramValue = parameterValues.value[param]
          // Escape the parameter value to prevent SQL injection
          const escapedValue =
            typeof paramValue === 'string' ? `'${paramValue.replace(/'/g, "''")}'` : paramValue
          finalQuery = finalQuery.replace(new RegExp(`:${param}`, 'g'), escapedValue)
        })
        // No parameters array needed for multi-statement queries with direct value replacement
        params = []
      } else {
        // For single statements, use parameterized queries
        detectedParameters.value.forEach((param, index) => {
          finalQuery = finalQuery.replace(new RegExp(`:${param}`, 'g'), '?')
          params.push(parameterValues.value[param])
        })
      }
    }

    const response = await callApi({
      action: 'execute_multi_sql',
      query: finalQuery,
      params: params,
      requiresAuth: true,
    })

    if (response.success) {
      queryResults.value = response.results || []
      executionTime.value = Date.now() - startTime

      // Add to history
      addToHistory(sqlQuery.value, true)

      if (hasMultipleStatements) {
        const totalAffected = response.totalAffectedRows || 0
        const totalResults = response.totalResults || 0
        ElMessage.success(
          `Multi-statement query executed successfully. ${totalAffected} rows affected, ${totalResults} results returned.`,
        )
      } else if (hasNonSelectStatements) {
        ElMessage.success(
          `Query executed successfully. ${response.affectedRows || 0} rows affected.`,
        )
      } else {
        ElMessage.success(
          `Query executed successfully. ${queryResults.value.length} rows returned.`,
        )
      }
    } else {
      error.value =
        response.message || response.error || 'An error occurred while executing the query.'
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

// Helper function to split SQL statements
const splitSqlStatements = (sql) => {
  if (!sql.trim()) return []

  // Split by semicolon, but handle semicolons inside strings and comments
  const statements = []
  let currentStatement = ''
  let inString = false
  let stringChar = null
  let inComment = false
  let commentType = null // '--' or '/*'
  let i = 0

  while (i < sql.length) {
    const char = sql[i]
    const nextChar = sql[i + 1]

    // Handle comments
    if (!inString && !inComment) {
      if (char === '-' && nextChar === '-') {
        inComment = true
        commentType = '--'
        currentStatement += char + nextChar
        i += 2
        continue
      } else if (char === '/' && nextChar === '*') {
        inComment = true
        commentType = '/*'
        currentStatement += char + nextChar
        i += 2
        continue
      }
    }

    // Handle end of comments
    if (inComment) {
      if (commentType === '--' && char === '\n') {
        inComment = false
        commentType = null
      } else if (commentType === '/*' && char === '*' && nextChar === '/') {
        inComment = false
        commentType = null
        currentStatement += char + nextChar
        i += 2
        continue
      }
      currentStatement += char
      i++
      continue
    }

    // Handle strings
    if (!inComment && (char === "'" || char === '"')) {
      if (!inString) {
        inString = true
        stringChar = char
      } else if (stringChar === char) {
        inString = false
        stringChar = null
      }
    }

    // Handle semicolon (statement separator)
    if (char === ';' && !inString && !inComment) {
      currentStatement += char
      const trimmedStatement = currentStatement.trim()
      if (trimmedStatement) {
        statements.push(trimmedStatement)
      }
      currentStatement = ''
    } else {
      currentStatement += char
    }

    i++
  }

  // Add the last statement if it's not empty
  const trimmedStatement = currentStatement.trim()
  if (trimmedStatement) {
    statements.push(trimmedStatement)
  }

  return statements
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

  // Check if the loaded query has parameters and enable parameter mode
  const paramRegex = /:([a-zA-Z_][a-zA-Z0-9_]*)/g
  const matches = query.match(paramRegex) || []
  if (matches.length > 0) {
    parameterMode.value = true
  }

  // Clear parameter values when loading from history
  parameterValues.value = {}
}

const clearHistory = () => {
  queryHistory.value = []
  localStorage.removeItem('sqlQueryHistory')
  ElMessage.success('Query history cleared.')
}

const saveAsTemplate = async () => {
  if (!sqlQuery.value.trim()) return

  // Check if all parameters have values
  if (parameterMode.value && detectedParameters.value.length > 0) {
    const missingParams = detectedParameters.value.filter(
      (param) => !parameterValues.value[param] || parameterValues.value[param].trim() === '',
    )

    if (missingParams.length > 0) {
      ElMessage.error(`Please fill in all required parameters: ${missingParams.join(', ')}`)
      return
    }
  }

  // Reset form and errors
  saveForm.value = {
    name: '',
    description: '',
  }
  saveFormErrors.value = {}

  // Show save dialog
  showSaveDialog.value = true
}

const loadTemplate = (template) => {
  sqlQuery.value = template.query
  showTemplatesDialog.value = false

  // Enable parameter mode if the template has parameters
  if (template.params && template.params.length > 0) {
    parameterMode.value = true
  }

  // Restore parameter values if they exist
  if (template.paramValues && Object.keys(template.paramValues).length > 0) {
    parameterValues.value = { ...template.paramValues }
  } else {
    // Clear parameter values if no saved values
    parameterValues.value = {}
  }

  ElMessage.success(`Template "${template.name}" loaded.`)
}

const deleteTemplate = async (templateId) => {
  try {
    await ElMessageBox.confirm(
      'Are you sure you want to delete this template?',
      'Delete Template',
      {
        confirmButtonText: 'Delete',
        cancelButtonText: 'Cancel',
        type: 'warning',
      },
    )

    // Delete from database
    const result = await callApi({
      query: 'DELETE FROM adv_sql WHERE id = ?',
      params: [templateId],
      requiresAuth: true,
    })

    if (result.success) {
      ElMessage.success('Template deleted successfully.')
      // Reload templates from database
      await loadTemplatesFromDatabase()
    } else {
      ElMessage.error('Failed to delete template: ' + (result.message || 'Unknown error'))
    }
  } catch (err) {
    if (err.message !== 'User cancelled') {
      ElMessage.error('Failed to delete template: ' + err.message)
    }
  }
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

const saveTemplate = async () => {
  // Validate form
  saveFormErrors.value = {}

  if (!saveForm.value.name.trim()) {
    saveFormErrors.value.name = 'Template name is required'
    return
  }

  isSaving.value = true

  try {
    // Prepare parameter values to save
    const paramValuesToSave = {}
    detectedParameters.value.forEach((param) => {
      if (parameterValues.value[param] && parameterValues.value[param].trim() !== '') {
        paramValuesToSave[param] = parameterValues.value[param].trim()
      }
    })

    // Save to database
    const result = await callApi({
      query:
        'INSERT INTO adv_sql (name, stmt, `desc`, params, param_values) VALUES (?, ?, ?, ?, ?)',
      params: [
        saveForm.value.name,
        sqlQuery.value,
        saveForm.value.description || '',
        JSON.stringify(detectedParameters.value),
        JSON.stringify(paramValuesToSave),
      ],
      requiresAuth: true,
    })

    if (result.success) {
      ElMessage.success('Template saved successfully.')
      showSaveDialog.value = false
      // Reload templates from database
      await loadTemplatesFromDatabase()
    } else {
      ElMessage.error('Failed to save template: ' + (result.message || 'Unknown error'))
    }
  } catch (err) {
    ElMessage.error('Failed to save template: ' + err.message)
  } finally {
    isSaving.value = false
  }
}

const getQueryTypeClass = (query) => {
  const type = query.trim().toUpperCase()
  if (type.startsWith('SELECT')) return 'type-select'
  if (type.startsWith('INSERT')) return 'type-insert'
  if (type.startsWith('UPDATE')) return 'type-update'
  if (type.startsWith('DELETE')) return 'type-delete'
  if (type.startsWith('DROP')) return 'type-drop'
  if (type.startsWith('CREATE')) return 'type-create'
  if (type.startsWith('ALTER')) return 'type-alter'
  if (type.startsWith('TRUNCATE')) return 'type-truncate'
  return 'type-other'
}

const getQueryType = (query) => {
  const type = query.trim().toUpperCase()
  if (type.startsWith('SELECT')) return 'SELECT'
  if (type.startsWith('INSERT')) return 'INSERT'
  if (type.startsWith('UPDATE')) return 'UPDATE'
  if (type.startsWith('DELETE')) return 'DELETE'
  if (type.startsWith('DROP')) return 'DROP'
  if (type.startsWith('CREATE')) return 'CREATE'
  if (type.startsWith('ALTER')) return 'ALTER'
  if (type.startsWith('TRUNCATE')) return 'TRUNCATE'
  return 'OTHER'
}

const filteredTemplates = computed(() => {
  const filters = templateFilters.value
  let filtered = savedTemplates.value.filter((template) => {
    const searchTerm = filters.search.toLowerCase()
    const nameMatch = template.name.toLowerCase().includes(searchTerm)
    const descriptionMatch = template.description.toLowerCase().includes(searchTerm)
    const queryMatch = template.query.toLowerCase().includes(searchTerm)
    const searchMatch = nameMatch || descriptionMatch || queryMatch

    const typeMatch = filters.queryType === '' || getQueryType(template.query) === filters.queryType

    return searchMatch && typeMatch
  })

  // Apply sorting
  switch (filters.sortBy) {
    case 'newest':
      filtered.sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
      break
    case 'oldest':
      filtered.sort((a, b) => new Date(a.created_at) - new Date(b.created_at))
      break
    case 'name':
      filtered.sort((a, b) => a.name.localeCompare(b.name))
      break
    case 'name-desc':
      filtered.sort((a, b) => b.name.localeCompare(a.name))
      break
    case 'type':
      filtered.sort((a, b) => getQueryType(a.query).localeCompare(getQueryType(b.query)))
      break
  }

  return filtered
})

const hasActiveFilters = computed(() => {
  const filters = templateFilters.value
  return filters.search.length > 0 || filters.queryType.length > 0 || filters.sortBy.length > 0
})

const clearTemplateFilters = () => {
  templateFilters.value = {
    search: '',
    queryType: '',
    sortBy: 'newest',
  }
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

.btn-sm {
  padding: 6px 12px;
  font-size: 12px;
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

.btn.requires-confirmation {
  background: #f59e0b;
  border: 2px solid #d97706;
}

.btn.requires-confirmation:hover:not(:disabled) {
  background: #d97706;
  transform: translateY(-1px);
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

.multi-statement-indicator {
  color: #8b5cf6;
  font-size: 12px;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 4px;
  background-color: #f3f4f6;
  padding: 2px 8px;
  border-radius: 4px;
  border: 1px solid #e5e7eb;
}

.parameters-section {
  padding: 20px;
  border-top: 1px solid #e5e7eb;
  background: #f8fafc;
}

.parameters-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.parameters-header h4 {
  margin: 0;
  color: #1f2937;
  font-size: 16px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
}

.parameters-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 16px;
}

.parameter-input {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.parameter-label {
  font-size: 14px;
  font-weight: 500;
  color: #374151;
  display: flex;
  align-items: center;
  gap: 4px;
}

.required-star {
  color: #ef4444;
  font-weight: bold;
}

.required-indicator {
  font-size: 12px;
  color: #ef4444;
  font-weight: 500;
  margin-left: 8px;
}

.loaded-indicator {
  font-size: 12px;
  color: #10b981;
  font-weight: 500;
  margin-left: 8px;
  display: flex;
  align-items: center;
  gap: 4px;
}

.loaded-indicator i {
  font-size: 14px;
}

.parameter-field {
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  transition: border-color 0.2s;
}

.parameter-field:focus {
  outline: none;
  border-color: #3b82f6;
}

.parameter-field.error {
  border-color: #ef4444;
  background-color: #fef2f2;
}

.parameter-field:disabled {
  background: #f9fafb;
  cursor: not-allowed;
}

.parameter-error {
  font-size: 12px;
  color: #ef4444;
  margin-top: 4px;
}

.templates-section {
  padding: 20px;
  border-top: 1px solid #e5e7eb;
  background: #f8fafc;
}

.templates-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.templates-header h4 {
  margin: 0;
  color: #1f2937;
  font-size: 16px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
}

.template-actions {
  display: flex;
  gap: 8px;
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

/* Dialog Styles */
.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.dialog-content {
  background: white;
  border-radius: 12px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
  max-width: 600px;
  width: 90%;
  max-height: 80vh;
  overflow: hidden;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #e5e7eb;
  background: #f9fafb;
}

.dialog-header h3 {
  margin: 0;
  color: #1f2937;
  font-size: 18px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
}

.dialog-close {
  background: none;
  border: none;
  font-size: 18px;
  color: #6b7280;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.2s;
}

.dialog-close:hover {
  background: #e5e7eb;
  color: #374151;
}

.dialog-body {
  padding: 20px;
  max-height: 60vh;
  overflow-y: auto;
}

.empty-templates {
  text-align: center;
  padding: 40px 20px;
  color: #6b7280;
}

.empty-templates i {
  margin-bottom: 16px;
  opacity: 0.5;
}

.empty-templates p {
  margin-bottom: 16px;
  font-size: 16px;
}

.templates-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.template-item {
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 12px;
  background: white;
  transition: all 0.2s;
}

.template-item:hover {
  border-color: #3b82f6;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.1);
}

.template-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.template-header h4 {
  margin: 0;
  color: #1f2937;
  font-size: 16px;
  font-weight: 600;
}

.template-type {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.type-select {
  background-color: #dbeafe;
  color: #1e40af;
}

.type-insert {
  background-color: #dcfce7;
  color: #166534;
}

.type-update {
  background-color: #fef3c7;
  color: #92400e;
}

.type-delete {
  background-color: #fee2e2;
  color: #991b1b;
}

.type-drop {
  background-color: #f3e8ff;
  color: #7c3aed;
}

.type-create {
  background-color: #e0f2fe;
  color: #0c4a6e;
}

.type-alter {
  background-color: #f0f9ff;
  color: #0c4a6e;
}

.type-truncate {
  background-color: #fef2f2;
  color: #dc2626;
}

.type-other {
  background-color: #f3f4f6;
  color: #374151;
}

.template-description {
  color: #6b7280;
  font-size: 14px;
  margin: 8px 0;
}

.template-query {
  font-family: 'Courier New', monospace;
  background-color: #f9fafb;
  padding: 8px;
  border-radius: 4px;
  font-size: 12px;
  color: #374151;
  margin: 8px 0;
  border-left: 3px solid #e5e7eb;
}

.template-params {
  margin-top: 8px;
  font-size: 12px;
}

.param-values {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-top: 4px;
}

.param-value {
  background-color: #f3f4f6;
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 11px;
  color: #374151;
}

.template-actions {
  display: flex;
  gap: 8px;
  margin-top: 12px;
}

.save-dialog {
  max-width: 700px;
}

.save-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-label {
  font-size: 14px;
  font-weight: 500;
  color: #374151;
  display: flex;
  align-items: center;
  gap: 4px;
}

.form-input {
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  transition: border-color 0.2s;
}

.form-input:focus {
  outline: none;
  border-color: #3b82f6;
}

.form-input.error {
  border-color: #ef4444;
  background-color: #fef2f2;
}

.form-textarea {
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  font-family: inherit;
  resize: vertical;
  transition: border-color 0.2s;
}

.form-textarea:focus {
  outline: none;
  border-color: #3b82f6;
}

.form-error {
  font-size: 12px;
  color: #ef4444;
  margin-top: 4px;
}

.query-preview {
  background: #f8fafc;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  padding: 12px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  color: #374151;
  max-height: 100px;
  overflow-y: auto;
  white-space: pre-wrap;
  word-break: break-all;
}

.parameters-preview {
  background: #f8fafc;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.param-preview {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}

.param-name {
  font-weight: 500;
  color: #374151;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
}

.param-value {
  color: #10b981;
  font-weight: 500;
}

.param-empty {
  color: #9ca3af;
  font-style: italic;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #e5e7eb;
}

.dialog-filters {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.filter-row {
  display: flex;
  gap: 12px;
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.filter-label {
  font-size: 14px;
  font-weight: 500;
  color: #374151;
  display: flex;
  align-items: center;
  gap: 4px;
}

.filter-input {
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  transition: border-color 0.2s;
}

.filter-select {
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  transition: border-color 0.2s;
}

.filter-actions {
  display: flex;
  gap: 12px;
}

.filter-count {
  font-size: 12px;
  color: #6b7280;
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

  .parameters-grid {
    grid-template-columns: 1fr;
  }

  .templates-header {
    flex-direction: column;
    gap: 12px;
    align-items: stretch;
  }

  .template-actions {
    justify-content: stretch;
  }

  .template-item {
    flex-direction: column;
    gap: 12px;
  }

  .template-actions {
    justify-content: stretch;
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

  .sql-container {
    flex-direction: column;
  }

  .sql-input {
    width: 100%;
    min-height: 200px;
  }

  .sql-results {
    width: 100%;
    margin-top: 20px;
  }

  .dialog-content {
    width: 95%;
    max-height: 90vh;
  }

  .dialog-filters {
    flex-direction: column;
    gap: 12px;
  }

  .filter-row {
    flex-direction: column;
    gap: 8px;
  }

  .filter-group {
    width: 100%;
  }

  .filter-actions {
    justify-content: space-between;
    width: 100%;
  }

  .template-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .template-actions {
    flex-direction: column;
  }

  .template-actions .btn {
    width: 100%;
  }
}
</style>
