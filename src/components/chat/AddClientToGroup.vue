<script setup>
import { ref, watch } from 'vue'
import { useApi } from '../../composables/useApi'

const props = defineProps({
  isVisible: {
    type: Boolean,
    default: false,
  },
  groupId: {
    type: Number,
    required: true,
  },
})

const emit = defineEmits(['close', 'client-added'])

const { callApi, loading, error } = useApi()
const availableClients = ref([])
const selectedClients = ref(new Set()) // Use Set to track selected client IDs

// Watch for changes in isVisible to fetch clients
watch(
  () => props.isVisible,
  async (isVisible) => {
    console.log('AddClientToGroup watch triggered - isVisible:', isVisible)
    if (isVisible) {
      console.log('Modal is visible, fetching available clients...')
      await fetchAvailableClients()
    }
  },
  { immediate: true },
)

const fetchAvailableClients = async () => {
  try {
    console.log('Fetching available clients for group:', props.groupId)

    // Get all clients who are not already in this chat group
    const result = await callApi({
      query: `
        SELECT 
          c.id,
          c.name,
          c.email,
          c.mobiles,
          c.address
        FROM clients c
        WHERE c.is_client = 1
          AND c.id NOT IN (
            SELECT cu.id_client 
            FROM chat_users cu 
            WHERE cu.id_chat_group = ? AND cu.is_active = 1 AND cu.id_client IS NOT NULL
          )
        ORDER BY c.name ASC
      `,
      params: [props.groupId],
      requiresAuth: true,
    })

    console.log('Available clients result:', result)

    if (result.success && result.data) {
      availableClients.value = result.data
      console.log('Available clients loaded:', availableClients.value)
    } else {
      console.log('No data returned or API call failed:', result)
    }
  } catch (err) {
    console.error('Error fetching available clients:', err)
  }
}

const toggleClientSelection = (clientId) => {
  if (selectedClients.value.has(clientId)) {
    selectedClients.value.delete(clientId)
  } else {
    selectedClients.value.add(clientId)
  }
}

const isClientSelected = (clientId) => {
  return selectedClients.value.has(clientId)
}

const getSelectedClientsCount = () => {
  return selectedClients.value.size
}

const getSelectedClientsInfo = () => {
  return availableClients.value.filter((client) => selectedClients.value.has(client.id))
}

const addClientsToGroup = async () => {
  if (selectedClients.value.size === 0) {
    alert('Please select at least one client to add')
    return
  }

  try {
    // Add all selected clients to the group
    const selectedClientIds = Array.from(selectedClients.value)

    // Create multiple INSERT statements for clients
    const insertQueries = selectedClientIds
      .map(() => 'INSERT INTO chat_users (id_client, id_chat_group, is_active) VALUES (?, ?, 1)')
      .join('; ')

    const params = selectedClientIds.flatMap((clientId) => [clientId, props.groupId])

    const result = await callApi({
      query: insertQueries,
      params: params,
      requiresAuth: true,
    })

    console.log('Add clients result:', result)

    if (result.success) {
      emit('client-added')
      closeModal()
      alert(`${selectedClients.value.size} client(s) added to group successfully`)
    } else {
      alert('Failed to add clients to group')
    }
  } catch (err) {
    console.error('Error adding clients to group:', err)
    alert('Error adding clients to group')
  }
}

const closeModal = () => {
  selectedClients.value.clear()
  emit('close')
}

const handleBackdropClick = (event) => {
  if (event.target === event.currentTarget) {
    closeModal()
  }
}
</script>

<template>
  <div v-if="isVisible" class="modal-overlay" @click="handleBackdropClick">
    <div class="modal-content">
      <div class="modal-header">
        <h3><i class="fas fa-user-plus"></i> Add Clients to Group</h3>
        <button @click="closeModal" class="close-btn">
          <i class="fas fa-times"></i>
        </button>
      </div>

      <div class="modal-body">
        <div v-if="loading" class="loading">
          <i class="fas fa-spinner fa-spin"></i> Loading clients...
        </div>

        <div v-else-if="error" class="error">
          <i class="fas fa-exclamation-triangle"></i> {{ error }}
        </div>

        <div v-else class="client-selection">
          <div class="selection-header">
            <h4><i class="fas fa-users"></i> Select Clients to Add</h4>
            <p class="selected-count">
              <i class="fas fa-check-circle"></i>
              {{ getSelectedClientsCount() }} client{{ getSelectedClientsCount() !== 1 ? 's' : '' }}
              selected
            </p>
          </div>

          <div class="clients-table-container">
            <table class="clients-table">
              <thead>
                <tr>
                  <th class="checkbox-header">
                    <input
                      type="checkbox"
                      :checked="
                        selectedClients.size === availableClients.length && availableClients.length > 0
                      "
                      :indeterminate="
                        selectedClients.size > 0 && selectedClients.size < availableClients.length
                      "
                      @change="
                        (e) => {
                          if (e.target.checked) {
                            availableClients.forEach((client) => selectedClients.add(client.id))
                          } else {
                            selectedClients.clear()
                          }
                        }
                      "
                      class="select-all-checkbox"
                    />
                  </th>
                  <th>Name</th>
                  <th>Email</th>
                  <th>Mobile</th>
                  <th>Address</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="client in availableClients"
                  :key="client.id"
                  class="client-row"
                  :class="{ selected: isClientSelected(client.id) }"
                  @click="toggleClientSelection(client.id)"
                >
                  <td class="checkbox-cell">
                    <input
                      type="checkbox"
                      :checked="isClientSelected(client.id)"
                      @click.stop
                      @change="toggleClientSelection(client.id)"
                      class="client-checkbox"
                    />
                  </td>
                  <td>
                    <i class="fas fa-user-circle"></i>
                    {{ client.name }}
                  </td>
                  <td>{{ client.email || '-' }}</td>
                  <td>{{ client.mobiles || '-' }}</td>
                  <td>{{ client.address || '-' }}</td>
                  <td>
                    <span
                      v-if="isClientSelected(client.id)"
                      class="status-badge selected-status"
                    >
                      <i class="fas fa-check"></i> Selected
                    </span>
                    <span v-else class="status-badge available-status">
                      <i class="fas fa-user"></i> Available
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div v-if="getSelectedClientsCount() > 0" class="selected-clients-summary">
            <h4>Selected Clients:</h4>
            <div class="selected-clients-list">
              <span
                v-for="client in getSelectedClientsInfo()"
                :key="client.id"
                class="selected-client-tag"
              >
                <i class="fas fa-user"></i>
                {{ client.name }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <div class="modal-footer">
        <button @click="closeModal" class="btn btn-secondary">Cancel</button>
        <button
          @click="addClientsToGroup"
          class="btn btn-primary"
          :disabled="getSelectedClientsCount() === 0 || loading"
        >
          <i class="fas fa-plus"></i>
          Add {{ getSelectedClientsCount() > 0 ? `(${getSelectedClientsCount()})` : '' }} Clients
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1001;
}

.modal-content {
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
  width: 90%;
  max-width: 900px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
}

.modal-header {
  padding: 20px;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background-color: #06b6d4;
  color: white;
  border-radius: 8px 8px 0 0;
}

.modal-header h3 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.close-btn {
  background: none;
  border: none;
  color: white;
  font-size: 1.2rem;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.close-btn:hover {
  background-color: rgba(255, 255, 255, 0.2);
}

.modal-body {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
}

.client-selection {
  margin-top: 20px;
}

.selection-header {
  margin-bottom: 20px;
}

.selection-header h4 {
  margin: 0 0 10px 0;
  color: #374151;
  display: flex;
  align-items: center;
  gap: 8px;
}

.selected-count {
  margin: 5px 0;
  color: #059669;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 6px;
}

.clients-table-container {
  overflow-x: auto;
  margin-bottom: 20px;
}

.clients-table {
  width: 100%;
  border-collapse: collapse;
  background-color: white;
  border-radius: 6px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.clients-table th {
  background-color: #f1f5f9;
  padding: 12px 16px;
  text-align: left;
  font-weight: 600;
  color: #475569;
  border-bottom: 1px solid #e2e8f0;
}

.clients-table td {
  padding: 12px 16px;
  border-bottom: 1px solid #f1f5f9;
  color: #374151;
}

.client-row {
  cursor: pointer;
  transition: background-color 0.2s;
}

.client-row:hover {
  background-color: #f8fafc;
}

.client-row.selected {
  background-color: #e0f2fe;
}

.checkbox-header,
.checkbox-cell {
  width: 50px;
  text-align: center;
}

.select-all-checkbox,
.client-checkbox {
  width: 18px;
  height: 18px;
  cursor: pointer;
}

.client-row td:nth-child(2) {
  display: flex;
  align-items: center;
  gap: 8px;
}

.selected-clients-summary {
  margin-top: 20px;
  padding: 15px;
  background-color: #f0fdf4;
  border: 1px solid #bbf7d0;
  border-radius: 6px;
}

.selected-clients-summary h4 {
  margin: 0 0 10px 0;
  color: #166534;
  display: flex;
  align-items: center;
  gap: 8px;
}

.selected-clients-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.selected-client-tag {
  background-color: #dcfce7;
  color: #166534;
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 0.8rem;
  display: flex;
  align-items: center;
  gap: 4px;
}

.loading,
.error {
  text-align: center;
  padding: 40px 20px;
  color: #64748b;
}

.error {
  color: #ef4444;
}

.modal-footer {
  padding: 20px;
  border-top: 1px solid #e2e8f0;
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 6px;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  background-color: #6b7280;
  color: white;
}

.btn-secondary:hover:not(:disabled) {
  background-color: #4b5563;
}

.btn-primary {
  background-color: #06b6d4;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background-color: #0891b2;
}

.status-badge {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 0.8rem;
  font-weight: 500;
  padding: 4px 8px;
  border-radius: 12px;
}

.selected-status {
  background-color: #dcfce7;
  color: #166534;
}

.available-status {
  background-color: #f1f5f9;
  color: #64748b;
}

/* Responsive design */
@media (max-width: 768px) {
  .modal-content {
    width: 95%;
    margin: 10px;
  }

  .clients-table {
    font-size: 0.9rem;
  }

  .clients-table th,
  .clients-table td {
    padding: 8px 12px;
  }

  .modal-footer {
    flex-direction: column;
  }

  .btn {
    width: 100%;
    justify-content: center;
  }

  .selected-clients-list {
    flex-direction: column;
  }
}
</style>

