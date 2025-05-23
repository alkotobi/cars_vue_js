<script setup>
import { ref, onMounted } from 'vue'
import { useApi } from '../composables/useApi'

const { callApi } = useApi()
const warehouses = ref([])
const newWarehouse = ref({
  warhouse_name: '',
  notes: ''
})
const editingWarehouse = ref(null)
const showAddForm = ref(false)
const showEditForm = ref(false)
const isLoading = ref(true)
const error = ref(null)

// Fetch all warehouses
const fetchWarehouses = async () => {
  try {
    const result = await callApi({
      query: 'SELECT * FROM warehouses',
      params: []
    })
    if (result.success) {
      warehouses.value = result.data
    } else {
      error.value = 'Failed to fetch warehouses: ' + result.error
    }
  } catch (err) {
    error.value = 'Error fetching warehouses: ' + err.message
  } finally {
    isLoading.value = false
  }
}

// Create a new warehouse
const createWarehouse = async () => {
  try {
    const result = await callApi({
      query: 'INSERT INTO warehouses (warhouse_name, notes) VALUES (?, ?)',
      params: [newWarehouse.value.warhouse_name, newWarehouse.value.notes]
    })
    if (result.success) {
      newWarehouse.value = { warhouse_name: '', notes: '' }
      showAddForm.value = false
      await fetchWarehouses()
    }
  } catch (err) {
    console.error('Error creating warehouse:', err)
  }
}

// Edit a warehouse
const startEdit = (warehouse) => {
  editingWarehouse.value = { ...warehouse }
  showEditForm.value = true
}

// Update a warehouse
const updateWarehouse = async () => {
  try {
    const result = await callApi({
      query: 'UPDATE warehouses SET warhouse_name = ?, notes = ? WHERE id = ?',
      params: [editingWarehouse.value.warhouse_name, editingWarehouse.value.notes, editingWarehouse.value.id]
    })
    if (result.success) {
      editingWarehouse.value = null
      showEditForm.value = false
      await fetchWarehouses()
    }
  } catch (err) {
    console.error('Error updating warehouse:', err)
  }
}

// Delete a warehouse
const deleteWarehouse = async (id) => {
  if (!confirm('Are you sure you want to delete this warehouse?')) {
    return
  }
  try {
    const result = await callApi({
      query: 'DELETE FROM warehouses WHERE id = ?',
      params: [id]
    })
    if (result.success) {
      await fetchWarehouses()
    }
  } catch (err) {
    console.error('Error deleting warehouse:', err)
  }
}

onMounted(() => {
  fetchWarehouses()
})
</script>

<template>
  <div class="warehouses-view">
    <div class="header">
      <h2>Warehouses</h2>
      <button class="add-button" @click="showAddForm = true">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
        </svg>
        Add Warehouse
      </button>
    </div>

    <!-- Add Form -->
    <div v-if="showAddForm" class="form-card">
      <div class="form-header">
        <h3>Add New Warehouse</h3>
        <button class="close-button" @click="showAddForm = false">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
      <div class="form-content">
        <input v-model="newWarehouse.warhouse_name" placeholder="Warehouse Name" class="input-field" />
        <textarea v-model="newWarehouse.notes" placeholder="Notes" class="input-field"></textarea>
      </div>
      <div class="form-actions">
        <button class="save-button" @click="createWarehouse">Save</button>
        <button class="cancel-button" @click="showAddForm = false">Cancel</button>
      </div>
    </div>

    <!-- Edit Form -->
    <div v-if="showEditForm" class="form-card">
      <div class="form-header">
        <h3>Edit Warehouse</h3>
        <button class="close-button" @click="showEditForm = false">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
      <div class="form-content">
        <input v-model="editingWarehouse.warhouse_name" placeholder="Warehouse Name" class="input-field" />
        <textarea v-model="editingWarehouse.notes" placeholder="Notes" class="input-field"></textarea>
      </div>
      <div class="form-actions">
        <button class="save-button" @click="updateWarehouse">Save</button>
        <button class="cancel-button" @click="showEditForm = false">Cancel</button>
      </div>
    </div>

    <div class="table-container">
      <div v-if="isLoading">Loading warehouses...</div>
      <div v-else-if="error">{{ error }}</div>
      <table v-else class="modern-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>Warehouse Name</th>
            <th>Notes</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="warehouse in warehouses" :key="warehouse.id">
            <td>{{ warehouse.id }}</td>
            <td>{{ warehouse.warhouse_name }}</td>
            <td>{{ warehouse.notes }}</td>
            <td>
              <button class="edit-button" @click="startEdit(warehouse)">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10" />
                </svg>
              </button>
              <button class="delete-button" @click="deleteWarehouse(warehouse.id)">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
                </svg>
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.warehouses-view {
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.add-button {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background-color: #22c55e;
  color: white;
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 0.375rem;
  cursor: pointer;
  transition: background-color 0.3s ease;
}

.add-button:hover {
  background-color: #16a34a;
}

.form-card {
  background-color: white;
  border-radius: 0.375rem;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
  margin-bottom: 2rem;
}

.form-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid #e5e7eb;
}

.close-button {
  background: none;
  border: none;
  cursor: pointer;
  color: #6b7280;
}

.form-content {
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.input-field {
  padding: 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  padding: 1rem 1.5rem;
  border-top: 1px solid #e5e7eb;
}

.save-button {
  background-color: #2563eb;
  color: white;
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 0.375rem;
  cursor: pointer;
  transition: background-color 0.3s ease;
}

.save-button:hover {
  background-color: #1d4ed8;
}

.cancel-button {
  background-color: #f3f4f6;
  color: #374151;
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 0.375rem;
  cursor: pointer;
  transition: background-color 0.3s ease;
}

.cancel-button:hover {
  background-color: #e5e7eb;
}

.table-container {
  overflow-x: auto;
}

.modern-table {
  width: 100%;
  border-collapse: collapse;
  background-color: white;
  border-radius: 0.375rem;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
}

.modern-table th,
.modern-table td {
  padding: 1rem;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

.modern-table th {
  background-color: #f9fafb;
  font-weight: 500;
}

.modern-table tbody tr:hover {
  background-color: #f3f4f6;
}

.edit-button,
.delete-button {
  background: none;
  border: none;
  cursor: pointer;
  color: #6b7280;
  padding: 0.25rem;
}

.edit-button:hover {
  color: #2563eb;
}

.delete-button:hover {
  color: #dc2626;
}
</style>