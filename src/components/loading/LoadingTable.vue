<template>
  <div class="loading-table-container">
    <div class="table-header">
      <h3>
        <i class="fas fa-truck-loading"></i>
        Recent Loading Records
      </h3>
      <div class="header-actions">
        <button @click="openAddDialog" class="add-btn" :disabled="loading">
          <i class="fas fa-plus"></i>
          <span>Add Record</span>
        </button>
        <button
          @click="refreshData"
          class="refresh-btn"
          :disabled="loading"
          :class="{ processing: loading }"
        >
          <i class="fas fa-sync-alt"></i>
          <span>Refresh</span>
          <i v-if="loading" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>
      </div>
    </div>

    <div class="table-wrapper">
      <div v-if="loading" class="loading-overlay">
        <i class="fas fa-spinner fa-spin fa-2x"></i>
        <span>Loading...</span>
      </div>

      <div v-else-if="error" class="error-message">
        <i class="fas fa-exclamation-circle"></i>
        {{ error }}
      </div>

      <div v-else-if="loadingRecords.length === 0" class="empty-state">
        <i class="fas fa-inbox fa-2x"></i>
        <p>No loading records found</p>
        <button @click="openAddDialog" class="add-first-btn">
          <i class="fas fa-plus"></i>
          Add First Record
        </button>
      </div>

      <div v-else class="table-content">
        <table class="loading-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>Operation Date</th>
              <th>Shipping Line</th>
              <th>Freight</th>
              <th>Loading Port</th>
              <th>Discharge Port</th>
              <th>EDD</th>
              <th>Loaded Date</th>
              <th>Notes</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="record in loadingRecords"
              :key="record.id"
              @click="selectLoadingRecord(record.id)"
              :class="{ 'selected-row': selectedLoadingId === record.id }"
              class="table-row"
            >
              <td class="id-cell">#{{ record.id }}</td>
              <td class="date-cell">
                {{ record.date_loading ? formatDate(record.date_loading) : '-' }}
              </td>
              <td class="shipping-line-cell">
                {{ record.shipping_line_name || '-' }}
              </td>
              <td class="freight-cell">
                {{ record.freight ? `$${record.freight}` : '-' }}
              </td>
              <td class="port-cell">
                {{ record.loading_port_name || '-' }}
              </td>
              <td class="port-cell">
                {{ record.discharge_port_name || '-' }}
              </td>
              <td class="date-cell">
                {{ record.EDD ? formatDate(record.EDD) : '-' }}
              </td>
              <td class="date-cell">
                {{ record.date_loaded ? formatDate(record.date_loaded) : '-' }}
              </td>
              <td class="notes-cell" :title="record.note">
                {{ record.note ? truncateText(record.note, 30) : '-' }}
              </td>
              <td class="actions-cell">
                <div class="action-buttons">
                  <button
                    @click="editRecord(record)"
                    class="action-btn edit-btn"
                    title="Edit Record"
                  >
                    <i class="fas fa-edit"></i>
                  </button>
                  <button
                    @click="deleteRecord(record)"
                    class="action-btn delete-btn"
                    title="Delete Record"
                  >
                    <i class="fas fa-trash"></i>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="table-footer">
      <span class="record-count">
        Showing {{ Math.min(loadingRecords.length, 5) }} of {{ totalRecords }} records
      </span>
    </div>

    <ContainersTable
      :selectedLoadingId="selectedLoadingId"
      @container-click="handleContainerClick"
      ref="containersTableRef"
    />
    <LoadingAssignedCars
      ref="assignedCarsRef"
      :selectedLoadedContainerId="selectedLoadedContainerId"
      @car-unassigned="handleCarUnassigned"
    />
    <UnassignedCars
      ref="unassignedCarsRef"
      :selectedLoadedContainerId="selectedLoadedContainerId"
      :selectedContainerOnBoard="selectedContainerOnBoard"
      @car-assigned="handleCarAssigned"
    />

    <!-- Add/Edit Dialog -->
    <div v-if="showDialog" class="dialog-overlay" @click.self="closeDialog">
      <div class="dialog">
        <div class="dialog-header">
          <h3>
            <i class="fas fa-truck-loading"></i>
            {{ isEditing ? 'Edit Loading Record' : 'Add New Loading Record' }}
          </h3>
          <button class="close-btn" @click="closeDialog" :disabled="isSubmitting">
            <i class="fas fa-times"></i>
          </button>
        </div>

        <form @submit.prevent="saveRecord" class="dialog-content">
          <div class="form-row">
            <div class="form-group">
              <label for="date_loading">
                <i class="fas fa-calendar"></i>
                Operation Date
              </label>
              <input
                type="date"
                id="date_loading"
                v-model="formData.date_loading"
                required
                :disabled="isSubmitting"
                min="1900-01-01"
                max="2100-12-31"
                @blur="validateDateField('date_loading')"
              />
            </div>

            <div class="form-group">
              <label for="id_shipping_line">
                <i class="fas fa-ship"></i>
                Shipping Line
              </label>
              <div class="input-with-button">
                <select
                  id="id_shipping_line"
                  v-model="formData.id_shipping_line"
                  required
                  :disabled="isSubmitting"
                >
                  <option value="">Select Shipping Line</option>
                  <option v-for="line in shippingLines" :key="line.id" :value="line.id">
                    {{ line.name }}
                  </option>
                </select>
                <button
                  type="button"
                  @click="openAddShippingLineDialog"
                  class="add-item-btn"
                  :disabled="isSubmitting"
                  title="Add New Shipping Line"
                >
                  <i class="fas fa-plus"></i>
                </button>
              </div>
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="freight">
                <i class="fas fa-dollar-sign"></i>
                Freight
              </label>
              <input
                type="number"
                id="freight"
                v-model="formData.freight"
                step="0.01"
                min="0"
                :disabled="isSubmitting"
                placeholder="0.00"
              />
            </div>

            <div class="form-group">
              <label for="id_loading_port">
                <i class="fas fa-anchor"></i>
                Loading Port
              </label>
              <div class="input-with-button">
                <select
                  id="id_loading_port"
                  v-model="formData.id_loading_port"
                  required
                  :disabled="isSubmitting"
                >
                  <option value="">Select Loading Port</option>
                  <option v-for="port in loadingPorts" :key="port.id" :value="port.id">
                    {{ port.loading_port }}
                  </option>
                </select>
                <button
                  type="button"
                  @click="openAddLoadingPortDialog"
                  class="add-item-btn"
                  :disabled="isSubmitting"
                  title="Add New Loading Port"
                >
                  <i class="fas fa-plus"></i>
                </button>
              </div>
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="id_discharge_port">
                <i class="fas fa-anchor"></i>
                Discharge Port
              </label>
              <div class="input-with-button">
                <select
                  id="id_discharge_port"
                  v-model="formData.id_discharge_port"
                  required
                  :disabled="isSubmitting"
                >
                  <option value="">Select Discharge Port</option>
                  <option v-for="port in dischargePorts" :key="port.id" :value="port.id">
                    {{ port.discharge_port }}
                  </option>
                </select>
                <button
                  type="button"
                  @click="openAddDischargePortDialog"
                  class="add-item-btn"
                  :disabled="isSubmitting"
                  title="Add New Discharge Port"
                >
                  <i class="fas fa-plus"></i>
                </button>
              </div>
            </div>

            <div class="form-group">
              <label for="EDD">
                <i class="fas fa-calendar-check"></i>
                EDD (Estimated Delivery Date)
              </label>
              <input type="date" id="EDD" v-model="formData.EDD" :disabled="isSubmitting" />
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="date_loaded">
                <i class="fas fa-check-circle"></i>
                Loaded Date
              </label>
              <input
                type="date"
                id="date_loaded"
                v-model="formData.date_loaded"
                :disabled="isSubmitting"
              />
            </div>
          </div>

          <div class="form-group full-width">
            <label for="note">
              <i class="fas fa-sticky-note"></i>
              Notes
            </label>
            <textarea
              id="note"
              v-model="formData.note"
              rows="3"
              :disabled="isSubmitting"
              placeholder="Enter any additional notes..."
            ></textarea>
          </div>

          <div class="dialog-actions">
            <button type="button" @click="closeDialog" class="cancel-btn" :disabled="isSubmitting">
              <i class="fas fa-times"></i>
              Cancel
            </button>
            <button
              type="submit"
              class="save-btn"
              :disabled="isSubmitting"
              :class="{ processing: isSubmitting }"
            >
              <i class="fas fa-save"></i>
              <span>{{ isSubmitting ? 'Saving...' : isEditing ? 'Update' : 'Save' }}</span>
              <i v-if="isSubmitting" class="fas fa-spinner fa-spin loading-indicator"></i>
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Quick Add Shipping Line Dialog -->
    <div
      v-if="showShippingLineDialog"
      class="dialog-overlay"
      @click.self="showShippingLineDialog = false"
    >
      <div class="quick-add-dialog">
        <div class="dialog-header">
          <h3>
            <i class="fas fa-ship"></i>
            Add New Shipping Line
          </h3>
          <button
            class="close-btn"
            @click="showShippingLineDialog = false"
            :disabled="isAddingItem"
          >
            <i class="fas fa-times"></i>
          </button>
        </div>

        <div class="dialog-content">
          <div class="form-group">
            <label for="shipping_line_name">
              <i class="fas fa-ship"></i>
              Shipping Line Name
            </label>
            <input
              type="text"
              id="shipping_line_name"
              v-model="quickAddForm.shipping_line_name"
              placeholder="Enter shipping line name"
              :disabled="isAddingItem"
              @keyup.enter="saveShippingLine"
            />
          </div>

          <div class="dialog-actions">
            <button
              type="button"
              @click="showShippingLineDialog = false"
              class="cancel-btn"
              :disabled="isAddingItem"
            >
              <i class="fas fa-times"></i>
              Cancel
            </button>
            <button
              type="button"
              @click="saveShippingLine"
              class="save-btn"
              :disabled="isAddingItem"
              :class="{ processing: isAddingItem }"
            >
              <i class="fas fa-save"></i>
              <span>{{ isAddingItem ? 'Adding...' : 'Add' }}</span>
              <i v-if="isAddingItem" class="fas fa-spinner fa-spin loading-indicator"></i>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Quick Add Loading Port Dialog -->
    <div
      v-if="showLoadingPortDialog"
      class="dialog-overlay"
      @click.self="showLoadingPortDialog = false"
    >
      <div class="quick-add-dialog">
        <div class="dialog-header">
          <h3>
            <i class="fas fa-anchor"></i>
            Add New Loading Port
          </h3>
          <button class="close-btn" @click="showLoadingPortDialog = false" :disabled="isAddingItem">
            <i class="fas fa-times"></i>
          </button>
        </div>

        <div class="dialog-content">
          <div class="form-group">
            <label for="loading_port_name">
              <i class="fas fa-anchor"></i>
              Loading Port Name
            </label>
            <input
              type="text"
              id="loading_port_name"
              v-model="quickAddForm.loading_port_name"
              placeholder="Enter loading port name"
              :disabled="isAddingItem"
              @keyup.enter="saveLoadingPort"
            />
          </div>

          <div class="dialog-actions">
            <button
              type="button"
              @click="showLoadingPortDialog = false"
              class="cancel-btn"
              :disabled="isAddingItem"
            >
              <i class="fas fa-times"></i>
              Cancel
            </button>
            <button
              type="button"
              @click="saveLoadingPort"
              class="save-btn"
              :disabled="isAddingItem"
              :class="{ processing: isAddingItem }"
            >
              <i class="fas fa-save"></i>
              <span>{{ isAddingItem ? 'Adding...' : 'Add' }}</span>
              <i v-if="isAddingItem" class="fas fa-spinner fa-spin loading-indicator"></i>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Quick Add Discharge Port Dialog -->
    <div
      v-if="showDischargePortDialog"
      class="dialog-overlay"
      @click.self="showDischargePortDialog = false"
    >
      <div class="quick-add-dialog">
        <div class="dialog-header">
          <h3>
            <i class="fas fa-anchor"></i>
            Add New Discharge Port
          </h3>
          <button
            class="close-btn"
            @click="showDischargePortDialog = false"
            :disabled="isAddingItem"
          >
            <i class="fas fa-times"></i>
          </button>
        </div>

        <div class="dialog-content">
          <div class="form-group">
            <label for="discharge_port_name">
              <i class="fas fa-anchor"></i>
              Discharge Port Name
            </label>
            <input
              type="text"
              id="discharge_port_name"
              v-model="quickAddForm.discharge_port_name"
              placeholder="Enter discharge port name"
              :disabled="isAddingItem"
              @keyup.enter="saveDischargePort"
            />
          </div>

          <div class="dialog-actions">
            <button
              type="button"
              @click="showDischargePortDialog = false"
              class="cancel-btn"
              :disabled="isAddingItem"
            >
              <i class="fas fa-times"></i>
              Cancel
            </button>
            <button
              type="button"
              @click="saveDischargePort"
              class="save-btn"
              :disabled="isAddingItem"
              :class="{ processing: isAddingItem }"
            >
              <i class="fas fa-save"></i>
              <span>{{ isAddingItem ? 'Adding...' : 'Add' }}</span>
              <i v-if="isAddingItem" class="fas fa-spinner fa-spin loading-indicator"></i>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useApi } from '@/composables/useApi'
import ContainersTable from './ContainersTable.vue'
import LoadingAssignedCars from './LoadingAssignedCars.vue'
import UnassignedCars from './UnassignedCars.vue'

const { callApi } = useApi()

const loadingRecords = ref([])
const totalRecords = ref(0)
const loading = ref(false)
const error = ref(null)

// CRUD state
const showDialog = ref(false)
const isEditing = ref(false)
const isSubmitting = ref(false)
const editingRecord = ref(null)

// Quick add dialog states
const showShippingLineDialog = ref(false)
const showLoadingPortDialog = ref(false)
const showDischargePortDialog = ref(false)
const isAddingItem = ref(false)
const selectedLoadingId = ref(null)
const selectedLoadedContainerId = ref(null)
const selectedContainerOnBoard = ref(false)
const assignedCarsRef = ref(null)
const unassignedCarsRef = ref(null)
const containersTableRef = ref(null)

// Form data
const formData = ref({
  date_loading: '',
  id_shipping_line: '',
  freight: '',
  id_loading_port: '',
  id_discharge_port: '',
  EDD: '',
  date_loaded: '',
  note: '',
})

// Quick add form data
const quickAddForm = ref({
  shipping_line_name: '',
  loading_port_name: '',
  discharge_port_name: '',
})

// Reference data
const shippingLines = ref([])
const loadingPorts = ref([])
const dischargePorts = ref([])

const fetchLoadingRecords = async () => {
  loading.value = true
  error.value = null

  try {
    const result = await callApi({
      query: `
        SELECT 
          l.id,
          l.date_loading,
          l.id_shipping_line,
          l.freight,
          l.id_loading_port,
          l.id_discharge_port,
          l.EDD,
          l.date_loaded,
          l.note,
          sl.name as shipping_line_name,
          lp.loading_port as loading_port_name,
          dp.discharge_port as discharge_port_name
        FROM loading l
        LEFT JOIN shipping_lines sl ON l.id_shipping_line = sl.id
        LEFT JOIN loading_ports lp ON l.id_loading_port = lp.id
        LEFT JOIN discharge_ports dp ON l.id_discharge_port = dp.id
        ORDER BY l.date_loading DESC, l.id DESC
        LIMIT 5
      `,
      params: [],
    })

    if (result.success) {
      loadingRecords.value = result.data
    } else {
      error.value = result.error || 'Failed to fetch loading records'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred while fetching loading records'
  } finally {
    loading.value = false
  }
}

const fetchTotalRecords = async () => {
  try {
    const result = await callApi({
      query: 'SELECT COUNT(*) as total FROM loading',
      params: [],
    })

    if (result.success) {
      totalRecords.value = result.data[0].total
    }
  } catch (err) {
    console.error('Failed to fetch total records:', err)
  }
}

const fetchReferenceData = async () => {
  try {
    // Fetch shipping lines
    const shippingLinesResult = await callApi({
      query: 'SELECT id, name FROM shipping_lines ORDER BY name',
      params: [],
    })
    if (shippingLinesResult.success) {
      shippingLines.value = shippingLinesResult.data
    }

    // Fetch loading ports
    const loadingPortsResult = await callApi({
      query: 'SELECT id, loading_port FROM loading_ports ORDER BY loading_port',
      params: [],
    })
    if (loadingPortsResult.success) {
      loadingPorts.value = loadingPortsResult.data
    }

    // Fetch discharge ports
    const dischargePortsResult = await callApi({
      query: 'SELECT id, discharge_port FROM discharge_ports ORDER BY discharge_port',
      params: [],
    })
    if (dischargePortsResult.success) {
      dischargePorts.value = dischargePortsResult.data
    }
  } catch (err) {
    console.error('Failed to fetch reference data:', err)
  }
}

const refreshData = async () => {
  await Promise.all([fetchLoadingRecords(), fetchTotalRecords()])
}

const openAddDialog = () => {
  isEditing.value = false
  editingRecord.value = null
  resetForm()
  showDialog.value = true
}

const editRecord = (record) => {
  console.log('Editing record:', record)
  isEditing.value = true
  editingRecord.value = record
  formData.value = {
    date_loading: record.date_loading || '',
    id_shipping_line: record.id_shipping_line || '',
    freight: record.freight || '',
    id_loading_port: record.id_loading_port || '',
    id_discharge_port: record.id_discharge_port || '',
    EDD: record.EDD || '',
    date_loaded: record.date_loaded || '',
    note: record.note || '',
  }
  console.log('Form data set to:', formData.value)
  showDialog.value = true
}

const deleteRecord = async (record) => {
  if (!confirm(`Are you sure you want to delete loading record #${record.id}?`)) {
    return
  }

  try {
    const result = await callApi({
      query: 'DELETE FROM loading WHERE id = ?',
      params: [record.id],
    })

    if (result.success) {
      await refreshData()
    } else {
      alert('Failed to delete record: ' + result.error)
    }
  } catch (err) {
    alert('Error deleting record: ' + err.message)
  }
}

const validateDateInput = (event) => {
  const input = event.target
  const value = input.value

  // Check if the value matches the expected format
  if (value && !value.match(/^\d{4}-\d{2}-\d{2}$/)) {
    alert('Invalid date format. Please use YYYY-MM-DD format.')
    input.value = ''
    return
  }

  // Check year range
  if (value) {
    const year = parseInt(value.substring(0, 4))
    if (year < 1900 || year > 2100) {
      alert('Year must be between 1900 and 2100.')
      input.value = ''
      return
    }
  }
}

const saveRecord = async () => {
  if (isSubmitting.value) return

  console.log('=== DEBUG: saveRecord called ===')
  console.log('Form data:', JSON.stringify(formData.value, null, 2))
  console.log('Date loading value:', formData.value.date_loading)
  console.log('Date loading type:', typeof formData.value.date_loading)
  console.log('Date loading length:', formData.value.date_loading?.length)

  // Pre-validation for date formats
  if (formData.value.date_loading) {
    console.log('Validating date_loading:', formData.value.date_loading)
    console.log('Date regex test:', formData.value.date_loading.match(/^\d{4}-\d{2}-\d{2}$/))

    if (!formData.value.date_loading.match(/^\d{4}-\d{2}-\d{2}$/)) {
      console.log('Date format validation failed')
      alert('Invalid loading date format. Please use YYYY-MM-DD format.')
      return
    }
    const year = parseInt(formData.value.date_loading.substring(0, 4))
    console.log('Year extracted:', year)
    if (year < 1900 || year > 2100) {
      console.log('Year validation failed:', year)
      alert('Loading date year must be between 1900 and 2100.')
      return
    }
  }

  if (formData.value.EDD) {
    if (!formData.value.EDD.match(/^\d{4}-\d{2}-\d{2}$/)) {
      alert('Invalid EDD format. Please use YYYY-MM-DD format.')
      return
    }
    const year = parseInt(formData.value.EDD.substring(0, 4))
    if (year < 1900 || year > 2100) {
      alert('EDD year must be between 1900 and 2100.')
      return
    }
  }

  if (formData.value.date_loaded) {
    if (!formData.value.date_loaded.match(/^\d{4}-\d{2}-\d{2}$/)) {
      alert('Invalid loaded date format. Please use YYYY-MM-DD format.')
      return
    }
    const year = parseInt(formData.value.date_loaded.substring(0, 4))
    if (year < 1900 || year > 2100) {
      alert('Loaded date year must be between 1900 and 2100.')
      return
    }
  }

  // Validate and format dates
  const validateAndFormatDate = (dateString) => {
    if (!dateString || !dateString.trim()) return null
    const date = new Date(dateString)
    if (isNaN(date.getTime())) {
      throw new Error('Invalid date format')
    }
    return date.toISOString().split('T')[0]
  }

  try {
    // Validate required fields
    if (!formData.value.date_loading) {
      alert('Loading date is required')
      return
    }

    // Format dates - convert empty strings to null
    const formattedDateLoading = validateAndFormatDate(formData.value.date_loading)
    const formattedEDD = validateAndFormatDate(formData.value.EDD)
    const formattedDateLoaded = validateAndFormatDate(formData.value.date_loaded)

    console.log('Formatted dates:', {
      date_loading: formattedDateLoading,
      EDD: formattedEDD,
      date_loaded: formattedDateLoaded,
    })

    isSubmitting.value = true
    let result

    if (isEditing.value) {
      // Update existing record
      result = await callApi({
        query: `
          UPDATE loading 
          SET date_loading = ?, id_shipping_line = ?, freight = ?, 
              id_loading_port = ?, id_discharge_port = ?, EDD = ?, 
              date_loaded = ?, note = ?
          WHERE id = ?
        `,
        params: [
          formattedDateLoading,
          formData.value.id_shipping_line || null,
          formData.value.freight || null,
          formData.value.id_loading_port || null,
          formData.value.id_discharge_port || null,
          formattedEDD,
          formattedDateLoaded,
          formData.value.note || null,
          editingRecord.value.id,
        ],
      })
    } else {
      // Insert new record
      const insertParams = [
        formattedDateLoading,
        formData.value.id_shipping_line || null,
        formData.value.freight || null,
        formData.value.id_loading_port || null,
        formData.value.id_discharge_port || null,
        formattedEDD,
        formattedDateLoaded,
        formData.value.note || null,
      ]

      console.log('Insert parameters:', insertParams)

      result = await callApi({
        query: `
          INSERT INTO loading 
          (date_loading, id_shipping_line, freight, id_loading_port, 
           id_discharge_port, EDD, date_loaded, note)
          VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        `,
        params: insertParams,
      })
    }

    if (result.success) {
      closeDialog()
      await refreshData()
    } else {
      alert('Failed to save record: ' + result.error)
    }
  } catch (err) {
    if (err.message === 'Invalid date format') {
      alert('Please check the date format. Dates should be in YYYY-MM-DD format.')
    } else {
      alert('Error saving record: ' + err.message)
    }
  } finally {
    isSubmitting.value = false
  }
}

const closeDialog = () => {
  showDialog.value = false
  resetForm()
}

const resetForm = () => {
  const today = new Date().toISOString().split('T')[0]
  formData.value = {
    date_loading: today,
    id_shipping_line: '',
    freight: '',
    id_loading_port: '',
    id_discharge_port: '',
    EDD: '',
    date_loaded: '',
    note: '',
  }
}

const formatDate = (dateString) => {
  if (!dateString) return '-'
  const date = new Date(dateString)
  return date.toLocaleDateString('en-GB', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  })
}

const truncateText = (text, maxLength) => {
  if (!text) return '-'
  if (text.length <= maxLength) return text
  return text.substring(0, maxLength) + '...'
}

const viewAllRecords = () => {
  // TODO: Implement view all records functionality
  console.log('View all records clicked')
}

const openAddShippingLineDialog = () => {
  showShippingLineDialog.value = true
  quickAddForm.value.shipping_line_name = ''
}

const openAddLoadingPortDialog = () => {
  showLoadingPortDialog.value = true
  quickAddForm.value.loading_port_name = ''
}

const openAddDischargePortDialog = () => {
  showDischargePortDialog.value = true
  quickAddForm.value.discharge_port_name = ''
}

const saveShippingLine = async () => {
  if (!quickAddForm.value.shipping_line_name.trim()) {
    alert('Please enter a shipping line name')
    return
  }

  isAddingItem.value = true
  try {
    const result = await callApi({
      query: 'INSERT INTO shipping_lines (name) VALUES (?)',
      params: [quickAddForm.value.shipping_line_name.trim()],
    })

    if (result.success) {
      // Refresh shipping lines data
      await fetchReferenceData()
      showShippingLineDialog.value = false
      quickAddForm.value.shipping_line_name = ''
    } else {
      alert('Failed to add shipping line: ' + result.error)
    }
  } catch (err) {
    alert('Error adding shipping line: ' + err.message)
  } finally {
    isAddingItem.value = false
  }
}

const saveLoadingPort = async () => {
  if (!quickAddForm.value.loading_port_name.trim()) {
    alert('Please enter a loading port name')
    return
  }

  isAddingItem.value = true
  try {
    const result = await callApi({
      query: 'INSERT INTO loading_ports (loading_port) VALUES (?)',
      params: [quickAddForm.value.loading_port_name.trim()],
    })

    if (result.success) {
      // Refresh loading ports data
      await fetchReferenceData()
      showLoadingPortDialog.value = false
      quickAddForm.value.loading_port_name = ''
    } else {
      alert('Failed to add loading port: ' + result.error)
    }
  } catch (err) {
    alert('Error adding loading port: ' + err.message)
  } finally {
    isAddingItem.value = false
  }
}

const saveDischargePort = async () => {
  if (!quickAddForm.value.discharge_port_name.trim()) {
    alert('Please enter a discharge port name')
    return
  }

  isAddingItem.value = true
  try {
    const result = await callApi({
      query: 'INSERT INTO discharge_ports (discharge_port) VALUES (?)',
      params: [quickAddForm.value.discharge_port_name.trim()],
    })

    if (result.success) {
      // Refresh discharge ports data
      await fetchReferenceData()
      showDischargePortDialog.value = false
      quickAddForm.value.discharge_port_name = ''
    } else {
      alert('Failed to add discharge port: ' + result.error)
    }
  } catch (err) {
    alert('Error adding discharge port: ' + err.message)
  } finally {
    isAddingItem.value = false
  }
}

// Watchers
watch(
  formData,
  (newVal) => {
    // Validate date formats as user types
    if (newVal.date_loading && newVal.date_loading.length === 10) {
      if (!newVal.date_loading.match(/^\d{4}-\d{2}-\d{2}$/)) {
        alert('Invalid date format. Please use YYYY-MM-DD format.')
        formData.value.date_loading = ''
        return
      }

      const year = parseInt(newVal.date_loading.substring(0, 4))
      if (year < 1900 || year > 2100) {
        alert('Year must be between 1900 and 2100.')
        formData.value.date_loading = ''
        return
      }
    }

    if (newVal.EDD && newVal.EDD.length === 10) {
      if (!newVal.EDD.match(/^\d{4}-\d{2}-\d{2}$/)) {
        alert('Invalid EDD format. Please use YYYY-MM-DD format.')
        formData.value.EDD = ''
        return
      }

      const year = parseInt(newVal.EDD.substring(0, 4))
      if (year < 1900 || year > 2100) {
        alert('EDD year must be between 1900 and 2100.')
        formData.value.EDD = ''
        return
      }
    }

    if (newVal.date_loaded && newVal.date_loaded.length === 10) {
      if (!newVal.date_loaded.match(/^\d{4}-\d{2}-\d{2}$/)) {
        alert('Invalid loaded date format. Please use YYYY-MM-DD format.')
        formData.value.date_loaded = ''
        return
      }

      const year = parseInt(newVal.date_loaded.substring(0, 4))
      if (year < 1900 || year > 2100) {
        alert('Loaded date year must be between 1900 and 2100.')
        formData.value.date_loaded = ''
        return
      }
    }
  },
  { deep: true },
)

const selectLoadingRecord = (id) => {
  selectedLoadingId.value = id
}

const handleContainerClick = async (containerId) => {
  selectedLoadedContainerId.value = containerId

  // Fetch the container's on board status
  try {
    const result = await callApi({
      query: 'SELECT date_on_board FROM loaded_containers WHERE id = ?',
      params: [containerId],
    })

    if (result.success && result.data && result.data.length > 0) {
      selectedContainerOnBoard.value = !!result.data[0].date_on_board
    } else {
      selectedContainerOnBoard.value = false
    }
  } catch (err) {
    console.error('Error fetching container on board status:', err)
    selectedContainerOnBoard.value = false
  }
}

const handleCarUnassigned = () => {
  // Refresh unassigned cars table when a car is unassigned
  console.log('Car unassigned - refreshing unassigned cars')
  if (unassignedCarsRef.value) {
    unassignedCarsRef.value.refreshData()
  }
  // Refresh containers table to update on board button states
  if (containersTableRef.value) {
    containersTableRef.value.fetchContainers()
  }
  // Refresh selected container on board status
  if (selectedLoadedContainerId.value) {
    refreshSelectedContainerStatus()
  }
}

const handleCarAssigned = () => {
  // Refresh assigned cars table when a car is assigned
  console.log('Car assigned - refreshing assigned cars')
  if (assignedCarsRef.value) {
    assignedCarsRef.value.refreshData()
  }
  // Refresh containers table to update on board button states
  if (containersTableRef.value) {
    containersTableRef.value.fetchContainers()
  }
  // Refresh selected container on board status
  if (selectedLoadedContainerId.value) {
    refreshSelectedContainerStatus()
  }
}

const refreshSelectedContainerStatus = async () => {
  try {
    const result = await callApi({
      query: 'SELECT date_on_board FROM loaded_containers WHERE id = ?',
      params: [selectedLoadedContainerId.value],
    })

    if (result.success && result.data && result.data.length > 0) {
      selectedContainerOnBoard.value = !!result.data[0].date_on_board
    } else {
      selectedContainerOnBoard.value = false
    }
  } catch (err) {
    console.error('Error refreshing container on board status:', err)
    selectedContainerOnBoard.value = false
  }
}

onMounted(() => {
  Promise.all([refreshData(), fetchReferenceData()])
})
</script>

<style scoped>
.loading-table-container {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #e5e7eb;
  background-color: #f8fafc;
}

.table-header h3 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #1f2937;
  font-size: 1.1rem;
}

.table-header h3 i {
  color: #3498db;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.add-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: #3498db;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9rem;
  transition: all 0.2s ease;
}

.add-btn:hover:not(:disabled) {
  background: #2980b9;
}

.add-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.refresh-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: #3498db;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9rem;
  transition: all 0.2s ease;
}

.refresh-btn:hover:not(:disabled) {
  background: #2980b9;
}

.refresh-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.refresh-btn.processing {
  background: #95a5a6;
}

.loading-indicator {
  font-size: 0.8rem;
}

.table-wrapper {
  position: relative;
  max-height: 200px;
  overflow-y: auto;
}

.loading-overlay {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
  color: #6b7280;
  gap: 12px;
}

.error-message {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 16px 20px;
  background-color: #fee2e2;
  color: #dc2626;
  border-radius: 4px;
  margin: 16px;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
  color: #6b7280;
  gap: 12px;
}

.empty-state i {
  color: #d1d5db;
}

.table-content {
  overflow-x: auto;
}

.loading-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;
}

.loading-table th {
  background-color: #f8fafc;
  padding: 12px 8px;
  text-align: left;
  font-weight: 600;
  color: #374151;
  border-bottom: 1px solid #e5e7eb;
  position: sticky;
  top: 0;
  z-index: 10;
}

.loading-table td {
  padding: 12px 8px;
  border-bottom: 1px solid #f3f4f6;
  color: #374151;
}

.loading-table tr:hover {
  background-color: #f9fafb;
}

.selected-row {
  background-color: #e3f2fd !important;
  border-left: 3px solid #3498db;
}

.table-row {
  cursor: pointer;
}

.id-cell {
  font-weight: 600;
  color: #3498db;
  width: 60px;
}

.date-cell {
  width: 100px;
  white-space: nowrap;
}

.shipping-line-cell {
  width: 120px;
}

.freight-cell {
  width: 80px;
  text-align: right;
  font-weight: 600;
}

.port-cell {
  width: 120px;
}

.notes-cell {
  max-width: 150px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.actions-cell {
  width: 100px;
  text-align: right;
}

.action-buttons {
  display: flex;
  justify-content: flex-end;
  gap: 4px;
}

.action-btn {
  padding: 4px 8px;
  background: none;
  border: none;
  cursor: pointer;
  transition: color 0.2s ease;
}

.action-btn:hover {
  color: #3498db;
}

.table-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  border-top: 1px solid #e5e7eb;
  background-color: #f8fafc;
  font-size: 0.85rem;
}

.record-count {
  color: #6b7280;
}

.view-all-link a {
  color: #3498db;
  text-decoration: none;
  display: flex;
  align-items: center;
  gap: 4px;
  transition: color 0.2s ease;
}

.view-all-link a:hover {
  color: #2980b9;
  text-decoration: underline;
}

/* Scrollbar styling */
.table-wrapper::-webkit-scrollbar {
  width: 8px;
}

.table-wrapper::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 4px;
}

.table-wrapper::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 4px;
}

.table-wrapper::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

/* Responsive design */
@media (max-width: 768px) {
  .table-header {
    flex-direction: column;
    gap: 12px;
    align-items: flex-start;
  }

  .table-footer {
    flex-direction: column;
    gap: 8px;
    align-items: flex-start;
  }

  .loading-table {
    font-size: 0.8rem;
  }

  .loading-table th,
  .loading-table td {
    padding: 8px 4px;
  }
}

/* Dialog styling */
.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
}

.dialog {
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  width: 90%;
  max-width: 600px;
  max-height: 90vh;
  overflow-y: auto;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid #e5e7eb;
  background-color: #f8fafc;
  border-top-left-radius: 8px;
  border-top-right-radius: 8px;
}

.dialog-header h3 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #1f2937;
  font-size: 1.1rem;
}

.dialog-header h3 i {
  color: #3498db;
}

.close-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 1.2rem;
  color: #6b7280;
  padding: 8px;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.close-btn:hover:not(:disabled) {
  background-color: #f3f4f6;
  color: #374151;
}

.close-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.dialog-content {
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
}

.form-group.full-width {
  grid-column: 1 / -1;
}

.form-group label {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 8px;
  color: #374151;
  font-weight: 500;
  font-size: 0.9rem;
}

.form-group label i {
  color: #6b7280;
  font-size: 0.8rem;
}

.form-group input,
.form-group select,
.form-group textarea {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 0.9rem;
  transition: all 0.2s ease;
  background-color: white;
}

.form-group input:focus,
.form-group select:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #3498db;
  box-shadow: 0 0 0 3px rgba(52, 152, 219, 0.1);
}

.form-group input:disabled,
.form-group select:disabled,
.form-group textarea:disabled {
  background-color: #f9fafb;
  color: #6b7280;
  cursor: not-allowed;
}

.form-group textarea {
  resize: vertical;
  min-height: 80px;
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 16px;
  border-top: 1px solid #e5e7eb;
}

.cancel-btn,
.save-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.9rem;
  font-weight: 500;
  transition: all 0.2s ease;
}

.cancel-btn {
  background-color: #f3f4f6;
  color: #374151;
}

.cancel-btn:hover:not(:disabled) {
  background-color: #e5e7eb;
  color: #1f2937;
}

.cancel-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.save-btn {
  background-color: #3498db;
  color: white;
}

.save-btn:hover:not(:disabled) {
  background-color: #2980b9;
}

.save-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.save-btn.processing {
  background-color: #95a5a6;
}

.loading-indicator {
  font-size: 0.8rem;
}

.add-first-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  background: none;
  border: none;
  cursor: pointer;
  color: #3498db;
  font-size: 0.9rem;
  text-decoration: underline;
  transition: color 0.2s ease;
}

.add-first-btn:hover {
  color: #2980b9;
}

/* Responsive design for dialog */
@media (max-width: 768px) {
  .dialog {
    width: 95%;
    margin: 20px;
  }

  .form-row {
    grid-template-columns: 1fr;
  }

  .dialog-actions {
    flex-direction: column;
  }

  .cancel-btn,
  .save-btn {
    width: 100%;
    justify-content: center;
  }
}

.input-with-button {
  display: flex;
  gap: 8px;
  align-items: flex-end;
}

.input-with-button select {
  flex: 1;
}

.add-item-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 38px;
  background-color: #3498db;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 0.8rem;
}

.add-item-btn:hover:not(:disabled) {
  background-color: #2980b9;
}

.add-item-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.quick-add-dialog {
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  width: 90%;
  max-width: 400px;
  max-height: 90vh;
  overflow-y: auto;
}
</style>
