<script setup>
import { ref, onMounted, computed } from 'vue'
import { useApi } from '../composables/useApi'
import { useSubmitGuard } from '../composables/useSubmitGuard'
import TaskForm from '../components/car-stock/TaskForm.vue'

const suppliers = ref([])
const { callApi } = useApi()
const { guard, isBusy } = useSubmitGuard()
const showAddDialog = ref(false)
const showEditDialog = ref(false)
const editingSupplier = ref(null)
const user = ref(null)
const error = ref(null)

// Add task form state
const showTaskForm = ref(false)
const selectedSupplierForTask = ref(null)

const isAdmin = computed(() => user.value?.role_id === 1)

const newSupplier = ref({
  name: '',
  contact_info: '',
  notes: '',
})

// callApi rejects on transport/HTTP failures and returns { success: false, error }
// when the SQL itself fails (executeQuery in api/api.php) - a duplicate name
// (supplier_name_unic) or a delete blocked by buy_bill.id_supplier lands in that
// second case. Both are funnelled into `error` and rendered, so a rejected save
// no longer looks identical to a no-op. Returns the result on success, null
// otherwise, so callers can keep the dialog open and let the message show.
const runQuery = async (payload, failureMessage) => {
  error.value = null
  try {
    const result = await callApi(payload)
    if (!result?.success) {
      error.value = result?.error ? `${failureMessage}: ${result.error}` : failureMessage
    }
    return result?.success ? result : null
  } catch (err) {
    error.value = `${failureMessage}: ${err.message}`
    console.error(failureMessage, err)
    return null
  }
}

const fetchSuppliers = async () => {
  const result = await runQuery(
    {
      query: `
        SELECT * FROM suppliers
        ORDER BY name ASC
      `,
      params: [],
    },
    'Failed to load suppliers',
  )
  if (result) {
    suppliers.value = result.data || []
  }
}

const addSupplier = guard('add', async () => {
  const result = await runQuery(
    {
      query: `
        INSERT INTO suppliers (name, contact_info, notes)
        VALUES (?, ?, ?)
      `,
      params: [newSupplier.value.name, newSupplier.value.contact_info, newSupplier.value.notes],
    },
    'Failed to add supplier',
  )
  if (!result) return
  showAddDialog.value = false
  newSupplier.value = {
    name: '',
    contact_info: '',
    notes: '',
  }
  await fetchSuppliers()
})

const editSupplier = (supplier) => {
  editingSupplier.value = { ...supplier }
  showEditDialog.value = true
}

const updateSupplier = guard('update', async () => {
  const result = await runQuery(
    {
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
    },
    'Failed to update supplier',
  )
  if (!result) return
  showEditDialog.value = false
  editingSupplier.value = null
  await fetchSuppliers()
})

const deleteSupplier = guard('delete', async (supplier) => {
  if (!confirm('Are you sure you want to delete this supplier?')) return
  const result = await runQuery(
    {
      query: 'DELETE FROM suppliers WHERE id = ?',
      params: [supplier.id],
    },
    'Failed to delete supplier',
  )
  if (!result) {
    // buy_bill.id_supplier references suppliers.id, so a supplier used by any
    // purchase bill cannot be removed. Say that in plain words instead of
    // showing the raw SQLSTATE text.
    if (/foreign key|1452/i.test(error.value || '')) {
      error.value = 'This supplier is used by purchase bills and cannot be deleted.'
    }
    return
  }
  await fetchSuppliers()
})

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
    fetchSuppliers()
  }
})

// Task form handling
const openTaskForSupplier = (supplier) => {
  selectedSupplierForTask.value = supplier
  showTaskForm.value = true
}

// Clears both flags so TaskForm unmounts. It fetches users/priorities/subjects
// on mount, so leaving it mounted behind isVisible=false kept that work - and a
// stale supplier object, if the row was deleted - alive for the whole view.
const closeTaskForm = () => {
  showTaskForm.value = false
  selectedSupplierForTask.value = null
}

const handleTaskCreated = () => {
  closeTaskForm()
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
              <button
                v-if="isAdmin"
                @click="deleteSupplier(supplier)"
                class="btn delete-btn"
                :disabled="isBusy('delete')"
              >
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
          <button @click="addSupplier" class="btn save-btn" :disabled="isBusy('add')">Save</button>
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
          <button @click="updateSupplier" class="btn save-btn" :disabled="isBusy('update')">
            Save
          </button>
          <button @click="showEditDialog = false" class="btn cancel-btn">Cancel</button>
        </div>
      </div>
    </div>

    <!-- Task Form -->
    <TaskForm
      v-if="selectedSupplierForTask"
      :entity-data="selectedSupplierForTask"
      entity-type="supplier"
      :is-visible="showTaskForm"
      @task-created="handleTaskCreated"
      @cancel="closeTaskForm"
    />

    <!-- Fixed so a failure is readable while a dialog or the task form is open -->
    <div v-if="error" class="error-banner" role="alert">
      <i class="fas fa-exclamation-triangle"></i>
      <span>{{ error }}</span>
      <button class="error-dismiss" @click="error = null" aria-label="Dismiss">&times;</button>
    </div>
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

.error-banner {
  position: fixed;
  top: 16px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 10002;
  display: flex;
  align-items: center;
  gap: 10px;
  max-width: min(720px, calc(100vw - 32px));
  padding: 12px 14px;
  border: 1px solid #fecaca;
  border-radius: 6px;
  background-color: #fef2f2;
  color: #b91c1c;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.error-banner span {
  flex: 1;
  word-break: break-word;
}

.error-dismiss {
  background: none;
  border: none;
  color: inherit;
  font-size: 1.25rem;
  line-height: 1;
  cursor: pointer;
  padding: 0 4px;
}

.textarea {
  min-height: 100px;
  resize: vertical;
}
</style>
