<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useApi } from '../../composables/useApi'
import { useEnhancedI18n } from '@/composables/useI18n'

const { t } = useEnhancedI18n()
const {
  getCustomClearanceAgents,
  createCustomClearanceAgent,
  updateCustomClearanceAgent,
  deleteCustomClearanceAgent,
} = useApi()

const props = defineProps({
  show: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['close'])

// State
const loading = ref(false)
const error = ref(null)
const success = ref(null)
const agents = ref([])
const searchQuery = ref('')

// Modal state
const showModal = ref(false)
const editingAgent = ref(null)
const agentForm = ref({
  name: '',
  contact_person: '',
  phone: '',
  email: '',
  address: '',
  license_number: '',
  notes: '',
  is_active: true,
})

// Filtered agents
const filteredAgents = computed(() => {
  if (!searchQuery.value.trim()) {
    return agents.value
  }
  const query = searchQuery.value.toLowerCase()
  return agents.value.filter(
    (agent) =>
      agent.name?.toLowerCase().includes(query) ||
      agent.contact_person?.toLowerCase().includes(query) ||
      agent.phone?.toLowerCase().includes(query) ||
      agent.email?.toLowerCase().includes(query) ||
      agent.license_number?.toLowerCase().includes(query),
  )
})

// Load agents
const loadAgents = async () => {
  loading.value = true
  error.value = null

  try {
    const data = await getCustomClearanceAgents()
    agents.value = data || []
  } catch (err) {
    error.value = err.message || 'Failed to load agents'
    console.error('Load error:', err)
  } finally {
    loading.value = false
  }
}

// Open modal for create/edit
const openModal = (agent = null) => {
  if (agent) {
    editingAgent.value = agent
    agentForm.value = {
      name: agent.name || '',
      contact_person: agent.contact_person || '',
      phone: agent.phone || '',
      email: agent.email || '',
      address: agent.address || '',
      license_number: agent.license_number || '',
      notes: agent.notes || '',
      is_active: agent.is_active !== 0,
    }
  } else {
    editingAgent.value = null
    agentForm.value = {
      name: '',
      contact_person: '',
      phone: '',
      email: '',
      address: '',
      license_number: '',
      notes: '',
      is_active: true,
    }
  }
  showModal.value = true
}

// Save agent
const saveAgent = async () => {
  if (!agentForm.value.name.trim()) {
    error.value = 'Name is required'
    return
  }

  loading.value = true
  error.value = null

  try {
    if (editingAgent.value) {
      await updateCustomClearanceAgent(editingAgent.value.id, agentForm.value)
      success.value = 'Agent updated successfully'
    } else {
      await createCustomClearanceAgent(agentForm.value)
      success.value = 'Agent created successfully'
    }
    showModal.value = false
    await loadAgents()
  } catch (err) {
    error.value = err.message || 'Failed to save agent'
  } finally {
    loading.value = false
  }
}

// Delete agent
const handleDelete = async (agent) => {
  if (!confirm(`Are you sure you want to delete "${agent.name}"?`)) return

  loading.value = true
  error.value = null

  try {
    await deleteCustomClearanceAgent(agent.id)
    success.value = 'Agent deleted successfully'
    await loadAgents()
  } catch (err) {
    error.value = err.message || 'Failed to delete agent'
  } finally {
    loading.value = false
  }
}

const closeModal = () => {
  error.value = null
  success.value = null
  searchQuery.value = ''
  showModal.value = false
  emit('close')
}

// Watch for show prop
watch(
  () => props.show,
  (newVal) => {
    if (newVal) {
      loadAgents()
    } else {
      closeModal()
    }
  },
  { immediate: true },
)

onMounted(() => {
  if (props.show) {
    loadAgents()
  }
})
</script>

<template>
  <div v-if="show" class="modal-overlay" @click="closeModal">
    <div class="modal-content agents-modal" @click.stop>
      <div class="modal-header">
        <h3>
          <i class="fas fa-user-tie"></i>
          Transiteurs
        </h3>
        <div class="header-actions">
          <button @click="openModal()" class="btn-add" title="Add new agent">
            <i class="fas fa-plus"></i>
            Add Agent
          </button>
          <button class="close-btn" @click="closeModal" :disabled="loading">
            <i class="fas fa-times"></i>
          </button>
        </div>
      </div>

      <div class="modal-body">
        <!-- Error Message -->
        <div v-if="error" class="error-message">
          <i class="fas fa-exclamation-circle"></i>
          {{ error }}
        </div>

        <!-- Success Message -->
        <div v-if="success" class="success-message">
          <i class="fas fa-check-circle"></i>
          {{ success }}
        </div>

        <!-- Search -->
        <div class="search-section">
          <input
            v-model="searchQuery"
            type="text"
            class="search-input"
            placeholder="Rechercher des transiteurs..."
          />
          <i class="fas fa-search search-icon"></i>
        </div>

        <!-- Loading State -->
        <div v-if="loading && agents.length === 0" class="loading-state">
          <i class="fas fa-spinner fa-spin"></i>
          <span>Chargement des transiteurs...</span>
        </div>

        <!-- Empty State -->
        <div v-else-if="filteredAgents.length === 0" class="empty-state">
          <i class="fas fa-user-tie"></i>
          <p>{{ searchQuery ? 'Aucun transiteur trouvé' : 'Aucun transiteur' }}</p>
        </div>

        <!-- Agents Table -->
        <div v-else class="agents-table">
          <table>
            <thead>
              <tr>
                <th>Name</th>
                <th>Contact Person</th>
                <th>Phone</th>
                <th>Email</th>
                <th>License</th>
                <th>Status</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="agent in filteredAgents" :key="agent.id">
                <td>
                  <strong>{{ agent.name }}</strong>
                </td>
                <td>{{ agent.contact_person || '-' }}</td>
                <td>{{ agent.phone || '-' }}</td>
                <td>{{ agent.email || '-' }}</td>
                <td>{{ agent.license_number || '-' }}</td>
                <td>
                  <span :class="agent.is_active ? 'status-active' : 'status-inactive'">
                    {{ agent.is_active ? 'Active' : 'Inactive' }}
                  </span>
                </td>
                <td>
                  <div class="action-buttons">
                    <button @click="openModal(agent)" class="btn-edit" title="Edit">
                      <i class="fas fa-edit"></i>
                    </button>
                    <button @click="handleDelete(agent)" class="btn-delete" title="Delete">
                      <i class="fas fa-trash"></i>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div class="modal-footer">
        <div class="summary">Total: {{ filteredAgents.length }}</div>
        <button class="btn-close" @click="closeModal">Close</button>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <div v-if="showModal" class="modal-overlay" @click="showModal = false">
      <div class="modal-content form-modal" @click.stop>
        <div class="modal-header">
          <h3>{{ editingAgent ? 'Modifier le Transiteur' : 'Nouveau Transiteur' }}</h3>
          <button @click="showModal = false" class="close-btn">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>Name/Company *</label>
            <input
              v-model="agentForm.name"
              type="text"
              class="form-input"
              placeholder="Nom du transiteur ou de l'entreprise"
              required
            />
          </div>

          <div class="form-group">
            <label>Contact Person</label>
            <input
              v-model="agentForm.contact_person"
              type="text"
              class="form-input"
              placeholder="Contact person name"
            />
          </div>

          <div class="form-row">
            <div class="form-group">
              <label>Phone</label>
              <input
                v-model="agentForm.phone"
                type="tel"
                class="form-input"
                placeholder="Phone number"
              />
            </div>
            <div class="form-group">
              <label>Email</label>
              <input
                v-model="agentForm.email"
                type="email"
                class="form-input"
                placeholder="Email address"
              />
            </div>
          </div>

          <div class="form-group">
            <label>License Number</label>
            <input
              v-model="agentForm.license_number"
              type="text"
              class="form-input"
              placeholder="License or registration number"
            />
          </div>

          <div class="form-group">
            <label>Address</label>
            <textarea
              v-model="agentForm.address"
              class="form-textarea"
              rows="2"
              placeholder="Physical address"
            ></textarea>
          </div>

          <div class="form-group">
            <label>Notes</label>
            <textarea
              v-model="agentForm.notes"
              class="form-textarea"
              rows="3"
              placeholder="Additional notes"
            ></textarea>
          </div>

          <div class="form-group">
            <label>
              <input v-model="agentForm.is_active" type="checkbox" class="form-checkbox" />
              Active
            </label>
          </div>
        </div>
        <div class="modal-footer">
          <button @click="showModal = false" class="btn-cancel">Cancel</button>
          <button
            @click="saveAgent"
            :disabled="!agentForm.name.trim() || loading"
            class="btn-primary"
          >
            {{ editingAgent ? 'Update' : 'Create' }}
          </button>
        </div>
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
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
}

.agents-modal {
  width: 95%;
  max-width: 1000px;
  max-height: 90vh;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid #e5e7eb;
}

.modal-header h3 {
  margin: 0;
  font-size: 1.25rem;
  color: #1f2937;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.btn-add {
  padding: 0.5rem 1rem;
  background-color: #10b981;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  transition: background-color 0.2s;
}

.btn-add:hover {
  background-color: #059669;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: #6b7280;
  padding: 0.5rem;
}

.close-btn:hover {
  color: #374151;
}

.modal-body {
  flex: 1;
  overflow-y: auto;
  padding: 1.5rem;
}

.error-message,
.success-message {
  padding: 1rem;
  border-radius: 4px;
  margin-bottom: 1rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.error-message {
  background-color: #fef2f2;
  color: #dc2626;
}

.success-message {
  background-color: #f0fdf4;
  color: #10b981;
}

.search-section {
  position: relative;
  margin-bottom: 1.5rem;
}

.search-input {
  width: 100%;
  padding: 0.5rem 2.5rem 0.5rem 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 0.875rem;
}

.search-icon {
  position: absolute;
  right: 0.75rem;
  top: 50%;
  transform: translateY(-50%);
  color: #6b7280;
}

.loading-state,
.empty-state {
  text-align: center;
  padding: 3rem 1rem;
  color: #6b7280;
}

.empty-state i {
  font-size: 3rem;
  color: #d1d5db;
  margin-bottom: 1rem;
}

.agents-table {
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
}

thead {
  background-color: #f9fafb;
}

th {
  padding: 0.75rem;
  text-align: left;
  font-weight: 600;
  color: #374151;
  font-size: 0.875rem;
  border-bottom: 2px solid #e5e7eb;
}

td {
  padding: 0.75rem;
  border-bottom: 1px solid #e5e7eb;
  font-size: 0.875rem;
}

tr:hover {
  background-color: #f9fafb;
}

.status-active {
  padding: 0.25rem 0.5rem;
  background-color: #f0fdf4;
  color: #10b981;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
}

.status-inactive {
  padding: 0.25rem 0.5rem;
  background-color: #f3f4f6;
  color: #6b7280;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
}

.action-buttons {
  display: flex;
  gap: 0.5rem;
}

.btn-edit,
.btn-delete {
  background: none;
  border: 1px solid #e5e7eb;
  border-radius: 4px;
  padding: 0.5rem;
  cursor: pointer;
  color: #6b7280;
  transition: all 0.2s;
}

.btn-edit:hover {
  background-color: #f0f9ff;
  border-color: #3b82f6;
  color: #3b82f6;
}

.btn-delete:hover {
  background-color: #fef2f2;
  border-color: #dc2626;
  color: #dc2626;
}

.modal-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-top: 1px solid #e5e7eb;
}

.summary {
  color: #6b7280;
  font-size: 0.875rem;
}

.btn-close,
.btn-cancel,
.btn-primary {
  padding: 0.75rem 1.5rem;
  border-radius: 4px;
  font-weight: 500;
  cursor: pointer;
  border: none;
}

.btn-close,
.btn-cancel {
  background-color: #f3f4f6;
  color: #374151;
}

.btn-close:hover,
.btn-cancel:hover {
  background-color: #e5e7eb;
}

.btn-primary {
  background-color: #3b82f6;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background-color: #2563eb;
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.form-modal {
  max-width: 600px;
  width: 90%;
  max-height: 90vh;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
  color: #374151;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.form-input,
.form-textarea {
  width: 100%;
  padding: 0.5rem;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 1rem;
}

.form-textarea {
  resize: vertical;
}

.form-checkbox {
  margin-right: 0.5rem;
}
</style>
