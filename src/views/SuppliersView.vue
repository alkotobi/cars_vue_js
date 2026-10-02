<script setup>
import { ref, onMounted, onBeforeUnmount, computed } from 'vue'
import { useApi } from '../composables/useApi'
import { useSubmitGuard } from '../composables/useSubmitGuard'
import { useEnhancedI18n } from '../composables/useI18n'
import { CREDIBILITY_ENABLED } from '../lib/featureFlags'
import { useSupplierCredibility } from '../composables/useSupplierCredibility'
import { latestCheckFor } from '../lib/supplierCredibility'
import TaskForm from '../components/car-stock/TaskForm.vue'
import SupplierCredibilityModal from '../components/car-stock/SupplierCredibilityModal.vue'

const suppliers = ref([])
const { callApi } = useApi()
const { guard, isBusy } = useSubmitGuard()
const showAddDialog = ref(false)
const showEditDialog = ref(false)
const editingSupplier = ref(null)
const user = ref(null)
const error = ref(null)

// Only the credibility strings below are translated; the rest of this view is
// still hardcoded English, as it was before. Migrating it is its own job.
const { t, locale } = useEnhancedI18n()
const credibility = useSupplierCredibility(t, locale)

// Add task form state
const showTaskForm = ref(false)
const selectedSupplierForTask = ref(null)

// Supplier whose credibility modal is open
const credibilitySupplier = ref(null)
const showCredibilityModal = ref(false)

// Row action menu. The button lives inside a scrolling table, so the menu is
// teleported to the body and positioned from the button's own rect, the same way
// ClientsView does it: a dropdown nested in a table gets clipped by overflow.
const openMenuFor = ref(null)
const menuSupplier = ref(null)
const menuPosition = ref({ x: 0, y: 0 })
const menuButton = ref(null)

const toggleMenu = (supplier, event) => {
  if (openMenuFor.value === supplier.id) {
    closeMenu()
    return
  }

  const button = event.currentTarget
  const rect = button.getBoundingClientRect()
  const menuWidth = 200
  const padding = 20

  // Right-align to the button, then keep it on screen: a menu hanging off the
  // right edge is the failure mode when the table is wide.
  let x = rect.right - menuWidth
  if (x < padding) x = padding

  openMenuFor.value = supplier.id
  menuSupplier.value = supplier
  menuButton.value = button
  menuPosition.value = { x, y: rect.bottom + window.scrollY + 4 }
}

const closeMenu = () => {
  openMenuFor.value = null
  menuSupplier.value = null
  menuButton.value = null
}

// Any click outside the button and the menu dismisses it, as does Escape. Bound
// on the document rather than the table so a click on empty space ends it too.
const onDocumentClick = (event) => {
  if (!openMenuFor.value) return
  if (menuButton.value?.contains(event.target)) return
  if (event.target?.closest?.('.row-menu')) return
  closeMenu()
}

const onDocumentKeydown = (event) => {
  if (event.key === 'Escape' && openMenuFor.value) {
    // Return focus to the button that opened the menu, so keyboard users are
    // not dropped back at the top of the document.
    menuButton.value?.focus()
    closeMenu()
  }
}

onMounted(() => {
  document.addEventListener('click', onDocumentClick)
  document.addEventListener('keydown', onDocumentKeydown)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', onDocumentClick)
  document.removeEventListener('keydown', onDocumentKeydown)
})

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
      params: [
        newSupplier.value.name,
        newSupplier.value.contact_info,
        newSupplier.value.notes,
        // Sent as NULL rather than '' so "nothing on record" and "a blank
        // space" stay distinguishable in the column.
      ],
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
  closeMenu()
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
  closeMenu()
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
    // One call for the whole table's badges, and only for admins: the endpoint
    // is admin-only, so asking as anyone else would just be an error banner.
    // Skipped entirely while the feature is off, so a hidden feature makes no
    // requests and cannot raise a banner the user has no way to dismiss.
    if (CREDIBILITY_ENABLED && isAdmin.value) {
      credibility.loadLatest()
    }
  }
})

// The newest check for a row, or null when it has never been checked.
const credibilityFor = (supplier) => latestCheckFor(credibility.checksBySupplier.value, supplier.id)

const openCredibility = async (supplier) => {
  closeMenu()
  credibilitySupplier.value = supplier
  showCredibilityModal.value = true
  await credibility.loadHistory(supplier.id)
}

// Re-runs the check. Every run costs a model call, so the button is the only
// place that can start one and it is disabled while one is in flight.
const runCredibility = async () => {
  const supplier = credibilitySupplier.value
  if (!supplier) return
  const stored = await credibility.assess(supplier)
  if (stored) {
    // The fields the model read may have changed since the last check.
    await credibility.loadHistory(supplier.id)
  }
}

const closeCredibility = () => {
  showCredibilityModal.value = false
  credibilitySupplier.value = null
}

// Task form handling
const openTaskForSupplier = (supplier) => {
  closeMenu()
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
            <th v-if="CREDIBILITY_ENABLED">{{ t('supplierCredibility.column') }}</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="supplier in suppliers" :key="supplier.id">
            <td>{{ supplier.name }}</td>
            <td>{{ supplier.contact_info }}</td>
            <td>{{ supplier.notes }}</td>
            <td v-if="CREDIBILITY_ENABLED">
              <!-- Score stays visible in the table so a row can be scanned
                   without opening anything. Running a new check is an action,
                   so it lives in the menu. -->
              <button
                v-if="isAdmin && credibilityFor(supplier)"
                @click="openCredibility(supplier)"
                class="btn credibility-btn has-check"
                :class="[`risk-${credibilityFor(supplier)?.risk_level}`]"
                :title="t('supplierCredibility.column')"
              >
                <i class="fas fa-shield-halved"></i>
                <span class="badge">{{ credibilityFor(supplier).score ?? '—' }}</span>
              </button>
              <span v-else-if="isAdmin" class="credibility-none">
                {{ t('supplierCredibility.neverCheckedShort') }}
              </span>
            </td>
            <td>
              <button
                @click.stop="toggleMenu(supplier, $event)"
                class="btn row-actions-btn"
                :class="{ active: openMenuFor === supplier.id }"
                :aria-expanded="openMenuFor === supplier.id"
                aria-haspopup="menu"
                title="Actions"
              >
                <i class="fas fa-ellipsis-v"></i>
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Row actions. Teleported to the body because the table is inside a
         scrolling container that would clip a dropdown rendered in place. -->
    <Teleport to="body">
      <div
        v-if="openMenuFor !== null && menuSupplier"
        class="row-menu"
        role="menu"
        :style="{ left: menuPosition.x + 'px', top: menuPosition.y + 'px' }"
      >
        <button
          v-if="CREDIBILITY_ENABLED && isAdmin"
          class="row-menu-item"
          role="menuitem"
          @click="openCredibility(menuSupplier)"
        >
          <i class="fas fa-shield-halved"></i>
          {{ t('supplierCredibility.run') }}
        </button>
        <button class="row-menu-item" role="menuitem" @click="editSupplier(menuSupplier)">
          <i class="fas fa-pen"></i>
          Edit
        </button>
        <button class="row-menu-item" role="menuitem" @click="openTaskForSupplier(menuSupplier)">
          <i class="fas fa-tasks"></i>
          Add New Task
        </button>
        <button
          v-if="isAdmin"
          class="row-menu-item danger"
          role="menuitem"
          :disabled="isBusy('delete')"
          @click="deleteSupplier(menuSupplier)"
        >
          <i class="fas fa-trash"></i>
          Delete
        </button>
      </div>
    </Teleport>

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

    <!-- Failures from background loads. While the modal is open it shows the
         error itself, next to the button that was pressed, so that the same
         failure is never reported in two places at once. -->
    <div
      v-if="CREDIBILITY_ENABLED && credibility.error.value && !showCredibilityModal"
      class="error-banner"
      role="alert"
    >
      <i class="fas fa-exclamation-triangle"></i>
      <span>{{ credibility.error.value }}</span>
      <button class="error-dismiss" @click="credibility.clearError()" aria-label="Dismiss">
        &times;
      </button>
    </div>

    <SupplierCredibilityModal
      v-if="CREDIBILITY_ENABLED && showCredibilityModal && credibilitySupplier"
      :supplier="credibilitySupplier"
      :check="credibilityFor(credibilitySupplier)"
      :history="credibility.history.value"
      :error="credibility.error.value ?? ''"
      :error-code="credibility.errorCode.value ?? ''"
      :is-loading="credibility.isLoadingHistory.value"
      :is-checking="credibility.isAssessing(credibilitySupplier.id)"
      @close="closeCredibility"
      @recheck="runCredibility"
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

/* Solid once a check exists, outlined before one does: the button is a call to
   action until the table has an opinion. */
.credibility-btn {
  background: none;
  color: #6b7280;
  border: 1px solid #d1d5db;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.credibility-btn.has-check {
  color: #fff;
  border-color: transparent;
}

.credibility-btn.has-check.risk-low {
  background-color: #10b981;
}

.credibility-btn.has-check.risk-medium {
  background-color: #f59e0b;
}

.credibility-btn.has-check.risk-high {
  background-color: #dc3545;
}

.credibility-btn .badge {
  font-weight: 700;
}

.credibility-none {
  color: #9ca3af;
  font-size: 0.85rem;
}

.row-actions-btn {
  background: none;
  border: 1px solid transparent;
  color: #6b7280;
  padding: 6px 10px;
  cursor: pointer;
}

.row-actions-btn:hover,
.row-actions-btn.active {
  background: #f3f4f6;
  border-color: #d1d5db;
  color: #1f2937;
}

/* Teleported to <body>, so these are not scoped to any table container. */
.row-menu {
  position: absolute;
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.15);
  padding: 8px;
  /* Above the table and the dialogs: a menu opened from a row has to survive
     the modal it may be about to open. */
  z-index: 20010;
  min-width: 200px;
}

.row-menu-item {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 10px 14px;
  background: none;
  border: none;
  color: #374151;
  font-size: 0.9rem;
  text-align: left;
  cursor: pointer;
  border-radius: 8px;
}

.row-menu-item:hover:not(:disabled) {
  background: #f8fafc;
  color: #1f2937;
}

.row-menu-item:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.row-menu-item i {
  width: 16px;
  text-align: center;
  color: #6b7280;
}

.row-menu-item.danger,
.row-menu-item.danger i {
  color: #dc2626;
}

.row-menu-item.danger:hover:not(:disabled) {
  background: #fef2f2;
  color: #b91c1c;
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
