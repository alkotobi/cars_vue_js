<script setup>
import { ref, onMounted, computed } from 'vue'
import { useApi } from '../composables/useApi'

const colors = ref([])
const { callApi, error } = useApi()
const showAddDialog = ref(false)
const showEditDialog = ref(false)
const editingColor = ref(null)
const user = ref(null)

const isAdmin = computed(() => user.value?.role_id === 1)

const newColor = ref({
  color: ''
})

const fetchColors = async () => {
  const result = await callApi({
    query: 'SELECT * FROM colors ORDER BY color ASC',
    params: []
  })
  if (result.success) {
    colors.value = result.data
  }
}

const addColor = async () => {
  const result = await callApi({
    query: 'INSERT INTO colors (color) VALUES (?)',
    params: [newColor.value.color]
  })
  if (result.success) {
    showAddDialog.value = false
    newColor.value = { color: '' }
    await fetchColors()
  }
}

const editColor = (color) => {
  editingColor.value = { ...color }
  showEditDialog.value = true
}

const updateColor = async () => {
  const result = await callApi({
    query: 'UPDATE colors SET color = ? WHERE id = ?',
    params: [editingColor.value.color, editingColor.value.id]
  })
  if (result.success) {
    showEditDialog.value = false
    editingColor.value = null
    await fetchColors()
  }
}

const deleteColor = async (color) => {
  if (confirm('Are you sure you want to delete this color?')) {
    const result = await callApi({
      query: 'DELETE FROM colors WHERE id = ?',
      params: [color.id]
    })
    if (result.success) {
      await fetchColors()
    }
  }
}

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
    fetchColors()
  }
})
</script>

<template>
  <div class="colors-view">
    <div class="header">
      <h2>Colors Management</h2>
      <button @click="showAddDialog = true" class="add-btn">Add Color</button>
    </div>
    <div class="content">
      <table class="colors-table">
        <thead>
          <tr>
            <th>Color</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="color in colors" :key="color.id">
            <td>{{ color.color }}</td>
            <td>
              <button @click="editColor(color)" class="btn edit-btn">Edit</button>
              <button 
                v-if="isAdmin"
                @click="deleteColor(color)" 
                class="btn delete-btn"
              >Delete</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Add Color Dialog -->
    <div v-if="showAddDialog" class="dialog-overlay">
      <div class="dialog">
        <h3>Add New Color</h3>
        <div class="form-group">
          <input 
            v-model="newColor.color" 
            placeholder="Color Name" 
            class="input-field"
          />
        </div>
        <div class="dialog-actions">
          <button @click="addColor" class="btn save-btn">Save</button>
          <button @click="showAddDialog = false" class="btn cancel-btn">Cancel</button>
        </div>
      </div>
    </div>

    <!-- Edit Color Dialog -->
    <div v-if="showEditDialog" class="dialog-overlay">
      <div class="dialog">
        <h3>Edit Color</h3>
        <div class="form-group">
          <input 
            v-model="editingColor.color" 
            placeholder="Color Name" 
            class="input-field"
          />
        </div>
        <div class="dialog-actions">
          <button @click="updateColor" class="btn save-btn">Save</button>
          <button @click="showEditDialog = false" class="btn cancel-btn">Cancel</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.colors-view {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.colors-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 20px;
}

.colors-table th,
.colors-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #ddd;
}

.colors-table th {
  background-color: #f8f9fa;
  font-weight: 600;
}

.colors-table tbody tr:hover {
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