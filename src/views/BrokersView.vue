<script setup>
import { ref, onMounted, computed } from 'vue'
import { useApi } from '../composables/useApi'

const clients = ref([])
const { callApi, error } = useApi()
const showAddDialog = ref(false)
const showEditDialog = ref(false)
const editingClient = ref(null)
const user = ref(null)
const validationError = ref('')

// Check if user is admin by getting role from localStorage
const isAdmin = computed(() => user.value?.role_id === 1)

const validateEmail = (email) => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailRegex.test(email)
}

const newClient = ref({
  name: '',
  address: '',
  email: '',
  mobiles: '',
  id_no: ''
})

const fetchClients = async () => {
  const result = await callApi({
    query: `
      SELECT * FROM clients where is_broker = 1
      ORDER BY name ASC
    `,
    params: []
  })
  if (result.success) {
    clients.value = result.data
  }
}

const addClient = async () => {
  if (!validateEmail(newClient.value.email)) {
    validationError.value = 'Please enter a valid email address'
    return
  }

  const result = await callApi({
    query: `
      INSERT INTO clients (name, address, email, mobiles, id_no, is_broker)
      VALUES (?, ?, ?, ?, ?, 1)
    `,
    params: [
      newClient.value.name,
      newClient.value.address,
      newClient.value.email,
      newClient.value.mobiles,
      newClient.value.id_no,
    ]
  })
  if (result.success) {
    showAddDialog.value = false
    validationError.value = ''
    newClient.value = {
      name: '',
      address: '',
      email: '',
      mobiles: '',
      id_no: ''
    }
    await fetchClients()
  }
  else {
    error.value = result.error
    console.error('Error adding client:', result.error)
  }
}

const editClient = (client) => {
  editingClient.value = { ...client }
  showEditDialog.value = true
}

const updateClient = async () => {
  if (!validateEmail(editingClient.value.email)) {
    validationError.value = 'Please enter a valid email address'
    return
  }

  const result = await callApi({
    query: `
      UPDATE clients 
      SET name = ?, address = ?, email = ?, mobiles = ?, id_no = ?, is_broker = 1
      WHERE id = ?
    `,
    params: [
      editingClient.value.name,
      editingClient.value.address,
      editingClient.value.email,
      editingClient.value.mobiles,
      editingClient.value.id_no,
      editingClient.value.id
    ]
  })
  if (result.success) {
    showEditDialog.value = false
    editingClient.value = null
    validationError.value = ''
    await fetchClients()
  }
  else {
    error.value = result.error
    console.error('Error updating client:', result.error)
  }
}

const deleteClient = async (client) => {
  if (confirm('Are you sure you want to delete this client?')) {
    const result = await callApi({
      query: 'DELETE FROM clients WHERE id = ?',
      params: [client.id]
    })
    if (result.success) {
      await fetchClients()
    }
  }
}

onMounted(() => {
    const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
    fetchClients()
}
}
)
</script>

<template>
  <div class="clients-view">
    <div class="header">
      <h2>Brokers Management</h2>
      <button @click="showAddDialog = true" class="add-btn">Add Client</button>
    </div>
    <div class="content">
      <table class="clients-table">
        <thead>
          <tr>
            <th>Name</th>
            <th>Address</th>
            <th>Email</th>
            <th>Mobile</th>
            <th>ID No</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="client in clients" :key="client.id">
            <td>{{ client.name }}</td>
            <td>{{ client.address }}</td>
            <td>{{ client.email }}</td>
            <td>{{ client.mobiles }}</td>
            <td>{{ client.id_no }}</td>
            <td>
              <button @click="editClient(client)" class="btn edit-btn">Edit</button>
              <button 
                v-if="isAdmin"
                @click="deleteClient(client)" 
                class="btn delete-btn"
              >Delete</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Add Client Dialog -->
    <div v-if="showAddDialog" class="dialog-overlay">
      <div class="dialog">
        <h3>Add New Client</h3>
        <div class="form-group">
          <input 
            v-model="newClient.name" 
            placeholder="Name" 
            class="input-field"
         
          />
          <input 
            v-model="newClient.address" 
            placeholder="Address" 
            class="input-field"
            
          />
          <input 
            v-model="newClient.email" 
            placeholder="Email" 
            class="input-field"
            :class="{ 'error': validationError && newClient.email }"
           
          />
          <input 
            v-model="newClient.mobiles" 
            placeholder="Mobile" 
            class="input-field"
            
          />
          <input 
            v-model="newClient.id_no" 
            placeholder="ID No" 
            class="input-field"
            
          />
          <div v-if="validationError" class="error-message">{{ validationError }}</div>
        </div>
        <div class="dialog-actions">
          <button @click="addClient" class="btn save-btn" :disabled="!isAdmin">Add</button>
          <button @click="showAddDialog = false" class="btn cancel-btn">Cancel</button>
        </div>
      </div>
    </div>

    <!-- Edit Client Dialog -->
    <div v-if="showEditDialog" class="dialog-overlay">
      <div class="dialog">
        <h3>Edit Client</h3>
        <div class="form-group">
          <input 
            v-model="editingClient.name" 
            placeholder="Name" 
            class="input-field"
            :readonly="!isAdmin" 
          />
          <input 
            v-model="editingClient.address" 
            placeholder="Address" 
            class="input-field"
           
          />
          <input 
            v-model="editingClient.email" 
            placeholder="Email" 
            class="input-field"
            :class="{ 'error': validationError && editingClient.email }"
           
          />
          <input 
            v-model="editingClient.mobiles" 
            placeholder="Mobile" 
            class="input-field"
          
          />
          <input 
            v-model="editingClient.id_no" 
            placeholder="ID No" 
            class="input-field"
           
          />
          <div v-if="validationError" class="error-message">{{ validationError }}</div>
        </div>
        <div class="dialog-actions">
          <button @click="updateClient" class="btn save-btn" :disabled="!isAdmin">Save</button>
          <button @click="showEditDialog = false" class="btn cancel-btn">Cancel</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.clients-view {
  padding: 20px;
  width: 1000px;
}

.content {
  margin-top: 20px;
}

.clients-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 20px;
}

.clients-table th,
.clients-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #ddd;
}

.clients-table th {
  background-color: #f8f9fa;
  font-weight: 600;
}

.clients-table tbody tr:hover {
  background-color: #f5f5f5;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
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

.edit-input {
  width: 100%;
  padding: 4px 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
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

.error {
  border-color: #ef4444;
}

.error-message {
  color: #ef4444;
  font-size: 0.875rem;
  margin-top: 4px;
}

.checkbox-field {
  display: flex;
  align-items: center;
  gap: 8px;
}

.checkbox-field input[type="checkbox"] {
  width: 16px;
  height: 16px;
  cursor: pointer;
}

.edit-input.checkbox {
  width: auto;
}

.input-field[readonly] {
  background-color: #f3f4f6;
  cursor: not-allowed;
}
</style>