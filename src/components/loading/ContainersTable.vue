<template>
  <div class="containers-table-container">
    <div class="table-header">
      <h3>
        <i class="fas fa-box"></i>
        Containers
      </h3>
    </div>

    <div v-if="!selectedLoadingId" class="empty-state">
      <i class="fas fa-mouse-pointer"></i>
      <p>Click on a loading record above to view its containers</p>
    </div>

    <div v-else class="table-wrapper">
      <div v-if="loading" class="loading-overlay">
        <i class="fas fa-spinner fa-spin"></i>
        <span>Loading containers...</span>
      </div>

      <div v-else-if="error" class="error-message">
        <i class="fas fa-exclamation-triangle"></i>
        <span>{{ error }}</span>
      </div>

      <div v-else-if="containers.length === 0" class="empty-state">
        <i class="fas fa-box-open"></i>
        <p>No containers found for this loading record</p>
      </div>

      <div v-else class="table-content">
        <table class="containers-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>Container</th>
              <th>Reference</th>
              <th>Date Departed</th>
              <th>Date Loaded</th>
              <th>Notes</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="container in containers" :key="container.id" class="table-row">
              <td class="id-cell">#{{ container.id }}</td>
              <td class="container-cell">{{ container.container_name || '-' }}</td>
              <td class="ref-cell">{{ container.ref_container || '-' }}</td>
              <td class="date-cell">{{ formatDate(container.date_departed) }}</td>
              <td class="date-cell">{{ formatDate(container.date_loaded) }}</td>
              <td class="note-cell">{{ truncateText(container.note, 30) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useApi } from '@/composables/useApi'

const props = defineProps({
  selectedLoadingId: {
    type: Number,
    default: null,
  },
})

const { callApi } = useApi()

const containers = ref([])
const loading = ref(false)
const error = ref(null)

const fetchContainers = async () => {
  if (!props.selectedLoadingId) {
    containers.value = []
    return
  }

  console.log('Fetching containers for loading ID:', props.selectedLoadingId)

  loading.value = true
  error.value = null

  try {
    const result = await callApi({
      query: `
        SELECT 
          lc.id,
          c.name as container_name,
          lc.ref_container,
          lc.date_departed,
          lc.date_loaded,
          lc.note
        FROM loaded_containers lc
        LEFT JOIN containers c ON lc.id_container = c.id
        WHERE lc.id_loading = ?
        ORDER BY lc.id DESC
      `,
      params: [props.selectedLoadingId],
    })

    console.log('Containers query result:', result)

    if (result.success) {
      containers.value = result.data || []
      console.log('Containers data:', containers.value)
    } else {
      error.value = result.error || 'Failed to load containers'
      console.error('Containers query failed:', result.error)
    }
  } catch (err) {
    error.value = 'Error loading containers: ' + err.message
    console.error('Containers fetch error:', err)
  } finally {
    loading.value = false
  }
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

const truncateText = (text, maxLength) => {
  if (!text) return '-'
  if (text.length <= maxLength) return text
  return text.substring(0, maxLength) + '...'
}

// Watch for changes in selectedLoadingId
watch(
  () => props.selectedLoadingId,
  (newId) => {
    if (newId) {
      fetchContainers()
    } else {
      containers.value = []
    }
  },
  { immediate: true },
)
</script>

<style scoped>
.containers-table-container {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  margin-top: 16px;
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

.table-wrapper {
  position: relative;
  max-height: 200px;
  overflow-y: auto;
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
  font-size: 2rem;
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
  position: sticky;
  top: 0;
  z-index: 10;
}

.containers-table td {
  padding: 12px 8px;
  border-bottom: 1px solid #f3f4f6;
  color: #374151;
}

.table-row:hover {
  background-color: #f9fafb;
}

.id-cell {
  font-weight: 600;
  color: #6b7280;
  width: 60px;
}

.container-cell {
  font-weight: 500;
}

.ref-cell {
  font-family: monospace;
  background-color: #f3f4f6;
  border-radius: 4px;
  padding: 2px 6px;
}

.date-cell {
  color: #6b7280;
  white-space: nowrap;
}

.note-cell {
  max-width: 200px;
  color: #6b7280;
}
</style>
