<script setup>
import { ref, onMounted, computed } from 'vue'
import { useApi } from '../composables/useApi'

const dischargePorts = ref([])
const { callApi, error } = useApi()
const showAddDialog = ref(false)
const showEditDialog = ref(false)
const editingPort = ref(null)
const user = ref(null)

const isAdmin = computed(() => user.value?.role_id === 1)

const newPort = ref({
  discharge_port: ''
})

const fetchPorts = async () => {
  const result = await callApi({
    query: 'SELECT * FROM discharge_ports ORDER BY discharge_port ASC',
    params: []
  })
  if (result.success) {
    dischargePorts.value = result.data
  }
}

const addPort = async () => {
  const result = await callApi({
    query: 'INSERT INTO discharge_ports (discharge_port) VALUES (?)',
    params: [newPort.value.discharge_port]
  })
  if (result.success) {
    showAddDialog.value = false
    newPort.value = { discharge_port: '' }
    await fetchPorts()
  }
}

const editPort = (port) => {
  editingPort.value = { ...port }
  showEditDialog.value = true
}

const updatePort = async () => {
  const result = await callApi({
    query: 'UPDATE discharge_ports SET discharge_port = ? WHERE id = ?',
    params: [editingPort.value.discharge_port, editingPort.value.id]
  })
  if (result.success) {
    showEditDialog.value = false
    editingPort.value = null
    await fetchPorts()
  }
}

const deletePort = async (port) => {
  if (confirm('Are you sure you want to delete this discharge port?')) {
    const result = await callApi({
      query: 'DELETE FROM discharge_ports WHERE id = ?',
      params: [port.id]
    })
    if (result.success) {
      await fetchPorts()
    }
  }
}

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
    fetchPorts()
  }
})
</script>

<template>
  <div class="ports-view">
    <div class="header">
      <h2>Discharge Ports Management</h2>
      <button @click="showAddDialog = true" class="add-btn">Add Discharge Port</button>
    </div>
    <div class="content">
      <table class="ports-table">
        <thead>
          <tr>
            <th>Discharge Port</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="port in dischargePorts" :key="port.id">
            <td>{{ port.discharge_port }}</td>
            <td>
              <button @click="editPort(port)" class="btn edit-btn">Edit</button>
              <button 
                v-if="isAdmin"
                @click="deletePort(port)" 
                class="btn delete-btn"
              >Delete</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Add Port Dialog -->
    <div v-if="showAddDialog" class="dialog-overlay">
      <div class="dialog">
        <h3>Add New Discharge Port</h3>
        <div class="form-group">
          <input 
            v-model="newPort.discharge_port" 
            placeholder="Discharge Port Name" 
            class="input-field"
          />
        </div>
        <div class="dialog-actions">
          <button @click="addPort" class="btn save-btn">Save</button>
          <button @click="showAddDialog = false" class="btn cancel-btn">Cancel</button>
        </div>
      </div>
    </div>

    <!-- Edit Port Dialog -->
    <div v-if="showEditDialog" class="dialog-overlay">
      <div class="dialog">
        <h3>Edit Discharge Port</h3>
        <div class="form-group">
          <input 
            v-model="editingPort.discharge_port" 
            placeholder="Discharge Port Name" 
            class="input-field"
          />
        </div>
        <div class="dialog-actions">
          <button @click="updatePort" class="btn save-btn">Save</button>
          <button @click="showEditDialog = false" class="btn cancel-btn">Cancel</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.ports-view {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.ports-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 20px;
}

.ports-table th,
.ports-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #ddd;
}

.ports-table th {
  background-color: #f8f9fa;
  font-weight: 600;
}

.ports-table tbody tr:hover {
  background-color: #f5f5f5;
}

.add-btn {
  padding: 8px 16px;
  background-color: #10b981;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.btn {
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  margin-right: 8px;
}

.edit-btn {
  background-color: #3b82f6;
  color: white;
}

.delete-btn {
  background-color: #ef4444;
  color: white;
}

.save-btn {
  background-color: #10b981;
  color: white;
}

.cancel-btn {
  background-color: #6b7280;
  color: white;
}

.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
}

.dialog {
  background-color: white;
  padding: 20px;
  border-radius: 8px;
  min-width: 400px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 20px;
}

.input-field {
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>