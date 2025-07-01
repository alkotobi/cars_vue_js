<script setup>
import { ref, onMounted, computed } from 'vue'
import { useApi } from '../composables/useApi'
import TaskForm from '../components/car-stock/TaskForm.vue'

const suppliers = ref([])
const { callApi, error } = useApi()
const showAddDialog = ref(false)
const showEditDialog = ref(false)
const editingSupplier = ref(null)
const user = ref(null)

// Add task form state
const showTaskForm = ref(false)
const selectedSupplierForTask = ref(null)

const isAdmin = computed(() => user.value?.role_id === 1)

const newSupplier = ref({
  name: '',
  contact_info: '',
  notes: '',
})

const fetchSuppliers = async () => {
  const result = await callApi({
    query: `
      SELECT * FROM suppliers
      ORDER BY name ASC
    `,
    params: [],
  })
  if (result.success) {
    suppliers.value = result.data
  }
}

const addSupplier = async () => {
  const result = await callApi({
    query: `
      INSERT INTO suppliers (name, contact_info, notes)
      VALUES (?, ?, ?)
    `,
    params: [newSupplier.value.name, newSupplier.value.contact_info, newSupplier.value.notes],
  })
  if (result.success) {
    showAddDialog.value = false
    newSupplier.value = {
      name: '',
      contact_info: '',
      notes: '',
    }
    await fetchSuppliers()
  } else {
    error.value = result.error
    console.error('Error adding supplier:', result.error)
  }
}

const editSupplier = (supplier) => {
  editingSupplier.value = { ...supplier }
  showEditDialog.value = true
}

const updateSupplier = async () => {
  const result = await callApi({
    query: `
      UPDATE suppliers 
      SET name = ?, contact_info = ?, notes = ?
      WHERE id = ?
    `,
    params: [
      editingSupplier.value.name,
      editingSupplier.value.contact_info,
      editingSupplier.value.notes,
      editingSupplier.value.id,
    ],
  })
  if (result.success) {
    showEditDialog.value = false
    editingSupplier.value = null
    await fetchSuppliers()
  } else {
    error.value = result.error
    console.error('Error updating supplier:', result.error)
  }
}

const deleteSupplier = async (supplier) => {
  if (confirm('Are you sure you want to delete this supplier?')) {
    const result = await callApi({
      query: 'DELETE FROM suppliers WHERE id = ?',
      params: [supplier.id],
    })
    if (result.success) {
      await fetchSuppliers()
    }
  }
}

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
    fetchSuppliers()
  }
})

// Add task handling methods
const openTaskForSupplier = (supplier) => {
  console.log('openTaskForSupplier called with supplier:', supplier)
  selectedSupplierForTask.value = supplier
  showTaskForm.value = true
}

const handleTaskCreated = () => {
  showTaskForm.value = false
  // Don't set selectedSupplierForTask to null to avoid prop validation errors
  // Optionally refresh data if needed
}
</script>

<template>
  <div class="suppliers-view">
    <div class="header">
      <h2>Suppliers Management</h2>
      <button @click="showAddDialog = true" class="add-btn">Add Supplier</button>
    </div>
    <div class="content">
      <table class="suppliers-table">
        <thead>
          <tr>
            <th>Name</th>
            <th>Contact Info</th>
            <th>Notes</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="supplier in suppliers" :key="supplier.id">
            <td>{{ supplier.name }}</td>
            <td>{{ supplier.contact_info }}</td>
            <td>{{ supplier.notes }}</td>
            <td>
              <button @click="editSupplier(supplier)" class="btn edit-btn">Edit</button>
              <button v-if="isAdmin" @click="deleteSupplier(supplier)" class="btn delete-btn">
                Delete
              </button>
              <button 
                @click="openTaskForSupplier(supplier)"
                class="btn task-btn"
                title="Add New Task"
              >
                <i class="fas fa-tasks"></i>
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Add Supplier Dialog -->
    <div v-if="showAddDialog" class="dialog-overlay">
      <div class="dialog">
        <h3>Add New Supplier</h3>
        <div class="form-group">
          <input v-model="newSupplier.name" placeholder="Name" class="input-field" />
          <textarea 
            v-model="newSupplier.contact_info" 
            placeholder="Contact Information" 
            class="input-field textarea"
          ></textarea>
          <textarea 
            v-model="newSupplier.notes" 
            placeholder="Notes" 
            class="input-field textarea"
          ></textarea>
        </div>
        <div class="dialog-actions">
          <button @click="addSupplier" class="btn save-btn">Save</button>
          <button @click="showAddDialog = false" class="btn cancel-btn">Cancel</button>
        </div>
      </div>
    </div>

    <!-- Edit Supplier Dialog -->
    <div v-if="showEditDialog" class="dialog-overlay">
      <div class="dialog">
        <h3>Edit Supplier</h3>
        <div class="form-group">
          <input v-model="editingSupplier.name" placeholder="Name" class="input-field" />
          <textarea 
            v-model="editingSupplier.contact_info" 
            placeholder="Contact Information" 
            class="input-field textarea"
          ></textarea>
          <textarea 
            v-model="editingSupplier.notes" 
            placeholder="Notes" 
            class="input-field textarea"
          ></textarea>
        </div>
        <div class="dialog-actions">
          <button @click="updateSupplier" class="btn save-btn">Save</button>
          <button @click="showEditDialog = false" class="btn cancel-btn">Cancel</button>
        </div>
      </div>
    </div>

    <!-- Task Form -->
    <TaskForm
      v-if="selectedSupplierForTask"
      :entityType="'supplier'"
      :entityData="selectedSupplierForTask"
      :isVisible="showTaskForm"
      @task-created="handleTaskCreated"
      @cancel="showTaskForm = false"
    />
  </div>
</template>

<style scoped>
.suppliers-view {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.suppliers-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 20px;
}

.suppliers-table th,
.suppliers-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #ddd;
}

.suppliers-table th {
  background-color: #f8f9fa;
  font-weight: 600;
}

.suppliers-table tbody tr:hover {
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
  background-color: #dc3545;
  color: white;
}

.delete-btn:hover {
  background-color: #c82333;
}

.task-btn {
  background-color: #8b5cf6;
  color: white;
}

.task-btn:hover {
  background-color: #7c3aed;
}

.task-btn i {
  font-size: 0.9rem;
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

.textarea {
  min-height: 100px;
  resize: vertical;
}
</style>
