<script setup>
import { ref, onMounted, computed } from 'vue'
import { useApi } from '../composables/useApi'

const clients = ref([])
const { callApi, uploadFile, getFileUrl, error } = useApi()
const showAddDialog = ref(false)
const showEditDialog = ref(false)
const editingClient = ref(null)
const user = ref(null)
const validationError = ref('')
const selectedFile = ref(null)
const editSelectedFile = ref(null)

// Check if user is admin by getting role from localStorage
const isAdmin = computed(() => user.value?.role_id === 1)

// Add helper function to check if file is an image
const isImageFile = (path) => {
  if (!path) return false
  const extension = path.split('.').pop().toLowerCase()
  return ['jpg', 'jpeg', 'png', 'gif', 'webp'].includes(extension)
}

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
      SELECT * FROM clients where is_broker = 0
      ORDER BY name ASC
    `,
    params: []
  })
  if (result.success) {
    clients.value = result.data
  }
}

const handleFileChange = (event, isEdit = false) => {
  const file = event.target.files[0]
  if (isEdit) {
    editSelectedFile.value = file
  } else {
    selectedFile.value = file
  }
}

const addClient = async () => {
  if (!validateEmail(newClient.value.email)) {
    validationError.value = 'Please enter a valid email address'
    return
  }

  try {
    // First insert the client to get the ID
  const result = await callApi({
    query: `
      INSERT INTO clients (name, address, email, mobiles, id_no, is_broker)
      VALUES (?, ?, ?, ?, ?, 0)
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
      const clientId = result.insertId

      // If there's a file selected, upload it using the new uploadFile method
      if (selectedFile.value) {
        try {
          const filename = `${clientId}.${selectedFile.value.name.split('.').pop()}`
          const uploadResult = await uploadFile(
            selectedFile.value,
            'ids',
            filename
          )
          
          if (uploadResult.success) {
            // Use the relative path returned from uploadFile
            await callApi({
              query: 'UPDATE clients SET id_copy_path = ? WHERE id = ?',
              params: [uploadResult.relativePath, clientId]
            })
          }
        } catch (err) {
          console.error('Error uploading file:', err)
          error.value = 'Client created but failed to upload ID document'
        }
      }

    showAddDialog.value = false
    validationError.value = ''
      selectedFile.value = null
    newClient.value = {
      name: '',
      address: '',
      email: '',
      mobiles: '',
      id_no: ''
    }
    await fetchClients()
    } else {
    error.value = result.error
    console.error('Error adding client:', result.error)
    }
  } catch (err) {
    error.value = err.message
    console.error('Error in add client process:', err)
  }
}

const editClient = (client) => {
  editingClient.value = { ...client }
  editSelectedFile.value = null
  showEditDialog.value = true
}

const updateClient = async () => {
  if (!validateEmail(editingClient.value.email)) {
    validationError.value = 'Please enter a valid email address'
    return
  }

  try {
    // Update client basic info
  const result = await callApi({
    query: `
      UPDATE clients 
      SET name = ?, address = ?, email = ?, mobiles = ?, id_no = ?, is_broker = 0
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
      // If there's a new file selected, upload it using the new uploadFile method
      if (editSelectedFile.value) {
        try {
          const filename = `${editingClient.value.id}.${editSelectedFile.value.name.split('.').pop()}`
          const uploadResult = await uploadFile(
            editSelectedFile.value,
            'ids',
            filename
          )
          
          if (uploadResult.success) {
            // Use the relative path returned from uploadFile
            await callApi({
              query: 'UPDATE clients SET id_copy_path = ? WHERE id = ?',
              params: [uploadResult.relativePath, editingClient.value.id]
            })
          }
        } catch (err) {
          console.error('Error uploading file:', err)
          error.value = 'Client updated but failed to upload new ID document'
        }
      }

    showEditDialog.value = false
    editingClient.value = null
      editSelectedFile.value = null
    validationError.value = ''
    await fetchClients()
    } else {
    error.value = result.error
    console.error('Error updating client:', result.error)
    }
  } catch (err) {
    error.value = err.message
    console.error('Error in update client process:', err)
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

const handleImageClick = (path) => {
  if (path) {
    window.open(getFileUrl(path), '_blank')
  }
}

onMounted(() => {
    const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
    fetchClients()
}
})
</script>

<template>
  <div class="clients-view">
    <div class="header">
      <h2>Clients Management</h2>
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
            <th>ID Document</th>
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
            <td class="id-document-cell">
              <div v-if="client.id_copy_path && isImageFile(client.id_copy_path)" 
                   class="image-preview"
                   @click="handleImageClick(client.id_copy_path)">
                <img 
                  :src="getFileUrl(client.id_copy_path)" 
                  :alt="'ID for ' + client.name"
                />
              </div>
              <a v-else-if="client.id_copy_path" 
                 :href="getFileUrl(client.id_copy_path)" 
                 target="_blank" 
                 class="document-link">
                View Document
              </a>
              <span v-else>No ID</span>
            </td>
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
          <div class="file-upload">
            <label for="id-document">ID Document:</label>
            <input 
              type="file" 
              id="id-document" 
              @change="handleFileChange($event)"
              accept="image/*,.pdf"
              class="file-input"
          />
          </div>
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
          <div class="file-upload">
            <label for="edit-id-document">ID Document:</label>
            <div v-if="editingClient.id_copy_path" class="current-file">
              Current: <a :href="editingClient.id_copy_path" target="_blank">View ID</a>
            </div>
            <input 
              type="file" 
              id="edit-id-document" 
              @change="handleFileChange($event, true)"
              accept="image/*,.pdf"
              class="file-input"
          />
          </div>
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

.file-upload {
  margin-top: 8px;
}

.file-upload label {
  display: block;
  margin-bottom: 4px;
  font-weight: 500;
}

.file-input {
  width: 100%;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  background-color: white;
}

.current-file {
  margin-bottom: 8px;
  color: #4b5563;
}

.current-file a {
  color: #3b82f6;
  text-decoration: none;
}

.current-file a:hover {
  text-decoration: underline;
}

.input-field[readonly] {
  background-color: #f3f4f6;
  cursor: not-allowed;
}

.id-document-cell {
  width: 120px;
  text-align: center;
}

.image-preview {
  width: 100px;
  height: 60px;
  margin: 0 auto;
  cursor: pointer;
  overflow: hidden;
  border-radius: 4px;
  border: 1px solid #ddd;
}

.image-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.2s ease;
}

.image-preview:hover img {
  transform: scale(1.1);
}

.document-link {
  color: #3b82f6;
  text-decoration: none;
  font-size: 0.9em;
  display: inline-block;
  padding: 4px 8px;
}

.document-link:hover {
  text-decoration: underline;
}
</style>