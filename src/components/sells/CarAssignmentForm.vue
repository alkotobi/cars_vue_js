<script setup>
import { ref, onMounted, watch, defineProps, defineEmits, computed } from 'vue'
import { useEnhancedI18n } from '../../composables/useI18n'
import { useApi } from '../../composables/useApi'
import { ElSelect, ElOption, ElDialog, ElButton, ElInput, ElCheckbox } from 'element-plus'
import 'element-plus/dist/index.css'

const { t } = useEnhancedI18n()

const props = defineProps({
  carId: {
    type: Number,
    default: null,
  },
  sellBillId: {
    type: Number,
    default: null,
  },
  visible: {
    type: Boolean,
    default: false,
  },
})
const user = ref(JSON.parse(localStorage.getItem('user')))
const isAdmin = computed(() => user.value?.role_id === 1)
const can_assign_to_tmp_clients = computed(
  () =>
    user.value?.permissions?.some((p) => p.permission_name === 'can_assign_to_tmp_clients') ||
    isAdmin.value,
)
const emit = defineEmits(['close', 'assign-success'])

const { callApi, uploadFile } = useApi()
const loading = ref(false)
const error = ref(null)
const clients = ref([])
const filteredClients = ref([])
const dischargePorts = ref([])
const carDetails = ref(null)

// Add loading state for form submission
const isSubmitting = ref(false)
const isProcessing = ref(false)

const formData = ref({
  id_client: null,
  id_port_discharge: null,
  price_cell: null,
  freight: null,
  rate: null,
  notes: '',
  is_tmp_client: 0,
})

// Add ref for CFR DA input
const cfrDaInput = ref(null)

// New client functionality
const showAddDialog = ref(false)
const selectedFile = ref(null)
const validationError = ref('')
const newClient = ref({
  name: '',
  address: '',
  email: '',
  mobiles: '',
  id_no: '',
  is_client: true,
  is_broker: false,
  notes: '',
})

// Add function to calculate price from CFR DA
const calculatePriceFromCFRDA = (cfrDa, rate, freight) => {
  if (!cfrDa || !rate) return null
  const parsedCfrDa = parseFloat(cfrDa)
  const parsedRate = parseFloat(rate)
  const parsedFreight = parseFloat(freight) || 0
  return (parsedCfrDa / parsedRate - parsedFreight).toFixed(2)
}

// Add function to calculate CFR DA from price
const calculateCFRDAFromPrice = (price, rate, freight) => {
  if (!price || !rate) return null
  const parsedPrice = parseFloat(price)
  const parsedRate = parseFloat(rate)
  const parsedFreight = parseFloat(freight) || 0
  return ((parsedPrice + parsedFreight) * parsedRate).toFixed(2)
}

// Add handlers for input changes
const handlePriceChange = (event) => {
  const newPrice = event.target.value
  if (newPrice && formData.value.rate) {
    // Only calculate CFR DA if it's not manually set by user
    if (!cfrDaInput.value || cfrDaInput.value === '') {
      cfrDaInput.value = calculateCFRDAFromPrice(
        newPrice,
        formData.value.rate,
        formData.value.freight,
      )
    }
  }
}

const handleCFRDAChange = (event) => {
  const newCFRDA = event.target.value
  if (newCFRDA && formData.value.rate) {
    // Always calculate price from CFR DA when user changes it
    formData.value.price_cell = calculatePriceFromCFRDA(
      newCFRDA,
      formData.value.rate,
      formData.value.freight,
    )
  }
}

// Fetch clients for dropdown
const fetchClients = async () => {
  try {
    const result = await callApi({
      query: `
        SELECT id, name
        FROM clients
        ORDER BY name ASC
      `,
      params: [],
    })

    if (result.success) {
      clients.value = result.data
      filteredClients.value = result.data
    } else {
      error.value = result.error || 'Failed to fetch clients'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  }
}

const remoteMethod = (query) => {
  if (query) {
    // Filter the existing clients data
    filteredClients.value = clients.value.filter((client) =>
      client.name.toLowerCase().includes(query.toLowerCase()),
    )
  } else {
    // If no query, show all clients
    filteredClients.value = clients.value
  }
}

const handleClientChange = async () => {
  if (!formData.value.id_client) return

  try {
    const result = await callApi({
      query: `
        SELECT id_copy_path
        FROM clients
        WHERE id = ?
      `,
      params: [formData.value.id_client],
    })

    if (result.success && result.data.length > 0) {
      if (!result.data[0].id_copy_path) {
        error.value = 'Cannot assign to this client because you did not provide ID card for him'
        formData.value.id_client = null
      }
    }
  } catch (err) {
    console.error('Error checking client ID card:', err)
    error.value = 'Failed to verify client ID card'
    formData.value.id_client = null
  }
}

// Fetch discharge ports for dropdown
const fetchDischargePorts = async () => {
  try {
    const result = await callApi({
      query: `
        SELECT id, discharge_port
        FROM discharge_ports
        ORDER BY discharge_port ASC
      `,
      params: [],
    })

    if (result.success) {
      dischargePorts.value = result.data
    } else {
      error.value = result.error || 'Failed to fetch discharge ports'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  }
}

// Add new function to fetch default freight values
const fetchDefaultFreight = async (isSmallCar) => {
  try {
    const result = await callApi({
      query: 'SELECT freight_small, freight_big FROM defaults LIMIT 1',
      params: [],
    })

    if (result.success && result.data.length > 0) {
      const defaultFreight = isSmallCar ? result.data[0].freight_small : result.data[0].freight_big
      formData.value.freight = defaultFreight
    }
  } catch (err) {
    console.error('Error fetching default freight:', err)
  }
}

// Update the fetchCarDetails function to include car size info and set freight
const fetchCarDetails = async () => {
  if (!props.carId) return

  try {
    const result = await callApi({
      query: `
        SELECT 
          cs.id,
          cs.vin,
          cs.price_cell,
          cs.is_big_car,
          cn.car_name,
          clr.color,
          bd.price_sell as buy_price
        FROM cars_stock cs
        LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
        LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
        LEFT JOIN colors clr ON cs.id_color = clr.id
        WHERE cs.id = ?
      `,
      params: [props.carId],
    })

    if (result.success && result.data.length > 0) {
      carDetails.value = result.data[0]
      // Set the CFR DA value if available (from cfr_da field)
      if (carDetails.value.cfr_da) {
        cfrDaInput.value = carDetails.value.cfr_da
      } else if (carDetails.value.price_cell) {
        // Fallback for old records that don't have cfr_da
        cfrDaInput.value = carDetails.value.price_cell
      }
      // Fetch and set default freight based on car size
      await fetchDefaultFreight(!carDetails.value.is_big_car)
      // After fetching default rate, don't calculate initial CFR DA automatically
      await fetchDefaultRate()
      // CFR DA will be calculated only when user enters it manually
    } else {
      error.value = result.error || 'Failed to fetch car details'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  }
}

// Update fetchDefaultRate to not calculate initial CFR DA automatically
const fetchDefaultRate = async () => {
  try {
    const result = await callApi({
      query: 'SELECT rate FROM defaults LIMIT 1',
      params: [],
    })

    if (result.success && result.data.length > 0) {
      formData.value.rate = result.data[0].rate
      // CFR DA will be calculated only when user enters it manually
    }
  } catch (err) {
    console.error('Error fetching default rate:', err)
  }
}

// Add computed property for CFR DA calculation
const calculateCFRDA = computed(() => {
  if (!formData.value.price_cell || !formData.value.rate) return 'N/A'
  const sellPrice = parseFloat(formData.value.price_cell) || 0
  const freight = parseFloat(formData.value.freight) || 0
  const rate = parseFloat(formData.value.rate) || 0
  return ((sellPrice + freight) * rate).toFixed(2)
})

// Add computed property for calculating sell price from CFR DA
const calculateSellPriceFromCFRDA = computed(() => {
  if (!cfrDaInput.value || !formData.value.rate) return 0
  const cfrDa = parseFloat(cfrDaInput.value) || 0
  const rate = parseFloat(formData.value.rate) || 0
  const freight = parseFloat(formData.value.freight) || 0
  return (cfrDa / rate - freight).toFixed(2)
})

// Assign car with collected data
const assignCar = async () => {
  // Prevent multiple submissions
  if (isSubmitting.value) {
    return
  }

  if (!validateForm()) return

  try {
    isSubmitting.value = true
    loading.value = true
    error.value = null

    // First get the bill_ref from the sell_bill
    const billResult = await callApi({
      query: `
        SELECT bill_ref
        FROM sell_bill
        WHERE id = ?
      `,
      params: [props.sellBillId],
    })

    if (!billResult.success || !billResult.data.length) {
      error.value = 'Failed to get bill reference'
      return
    }

    const billRef = billResult.data[0].bill_ref
    const currentDate = new Date().toISOString().split('T')[0]

    // Check if car is still available for assignment
    const checkResult = await callApi({
      query: `
        SELECT id, id_client
        FROM cars_stock
        WHERE id = ?
      `,
      params: [props.carId],
    })

    if (!checkResult.success || !checkResult.data.length) {
      error.value = 'Car not found'
      return
    }

    if (checkResult.data[0].id_client !== null) {
      error.value =
        'This car has already been assigned to another client. Please refresh and try again.'
      return
    }

    // Store the exact CFR DA value that user entered
    const cfrDaValue = cfrDaInput.value

    const result = await callApi({
      query: `
        UPDATE cars_stock 
        SET id_sell = ?,
            id_client = ?,
            id_port_discharge = ?,
            cfr_da = ?,
            freight = ?,
            date_sell = ?,
            id_sell_pi = ?,
            rate = ?,
            notes = ?,
            is_tmp_client = ?,
            date_assigned = NOW()
        WHERE id = ? AND id_client IS NULL
      `,
      params: [
        props.sellBillId,
        formData.value.id_client,
        formData.value.id_port_discharge,
        cfrDaValue, // Store exact CFR DA value only
        formData.value.freight || null,
        currentDate,
        billRef,
        formData.value.rate || null,
        formData.value.notes || null,
        formData.value.is_tmp_client,
        props.carId,
      ],
    })

    if (result.success) {
      if (result.affectedRows === 0) {
        error.value =
          'This car has already been assigned to another client. Please refresh and try again.'
      } else {
        emit('assign-success')
        resetForm()
      }
    } else {
      error.value = result.error || 'Failed to assign car'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
    isSubmitting.value = false
  }
}

// Validate form before submission
const validateForm = () => {
  if (!formData.value.id_client) {
    error.value = 'Please select a client'
    return false
  }

  if (!formData.value.id_port_discharge) {
    error.value = 'Please select a discharge port'
    return false
  }

  if (!cfrDaInput.value) {
    error.value = 'Please enter a CFR DA amount'
    return false
  }

  if (!formData.value.rate) {
    error.value = 'Please enter a rate'
    return false
  }

  return true
}

// Reset form after submission
const resetForm = () => {
  const currentRate = formData.value.rate // Preserve current rate
  formData.value = {
    id_client: null,
    id_port_discharge: null,
    price_cell: null,
    freight: null,
    rate: currentRate, // Keep the default rate
    notes: '',
    is_tmp_client: 0,
  }
  cfrDaInput.value = null // Reset CFR DA input
  error.value = null
}

// Calculate potential profit
const calculateProfit = () => {
  if (!carDetails.value?.buy_price || !formData.value.price_cell) return 'N/A'

  const buyPrice = parseFloat(carDetails.value.buy_price)
  const sellPrice = parseFloat(formData.value.price_cell)
  const freightCost = formData.value.freight ? parseFloat(formData.value.freight) : 0

  return (sellPrice - buyPrice - freightCost).toFixed(2)
}

// New client functionality
const validateEmail = (email) => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailRegex.test(email)
}

const handleFileChange = (event) => {
  const file = event.target.files[0]
  selectedFile.value = file
}

const addClient = async () => {
  if (isSubmitting.value) return

  validationError.value = ''

  if (!newClient.value.mobiles) {
    validationError.value = 'Mobile number is required'
    return
  }

  if (!newClient.value.id_no) {
    validationError.value = 'ID number is required'
    return
  }

  if (!selectedFile.value) {
    validationError.value = 'ID Document is required'
    return
  }

  if (newClient.value.email && !validateEmail(newClient.value.email)) {
    validationError.value = 'Please enter a valid email address'
    return
  }

  try {
    isSubmitting.value = true
    const result = await callApi({
      query: `
        INSERT INTO clients (name, address, email, mobiles, id_no, is_broker, is_client, notes)
        VALUES (?, ?, ?, ?, ?, ?, 1, ?)
      `,
      params: [
        newClient.value.name,
        newClient.value.address,
        newClient.value.email,
        newClient.value.mobiles,
        newClient.value.id_no,
        newClient.value.is_broker ? 1 : 0,
        newClient.value.notes,
      ],
    })

    if (result.success) {
      const clientId = result.lastInsertId
      if (selectedFile.value) {
        try {
          const filename = `${clientId}.${selectedFile.value.name.split('.').pop()}`
          const uploadResult = await uploadFile(selectedFile.value, 'ids', filename)

          if (uploadResult.success) {
            const updateResult = await callApi({
              query: 'UPDATE clients SET id_copy_path = ? WHERE id = ?',
              params: [uploadResult.relativePath, clientId],
            })

            if (updateResult.success) {
              formData.value.id_client = clientId
              await fetchClients()
            }
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
        id_no: '',
        is_client: true,
        is_broker: false,
        notes: '',
      }
    }
  } catch (err) {
    console.error('Error adding client:', err)
    error.value = 'Failed to add client'
  } finally {
    isSubmitting.value = false
  }
}

onMounted(() => {
  fetchClients()
  fetchDischargePorts()
  fetchDefaultRate()
  if (props.carId) {
    fetchCarDetails()
  }
})
</script>

<template>
  <div v-if="visible" class="car-assignment-form-overlay">
    <div class="car-assignment-form">
      <!-- Loading Overlay -->
      <div v-if="loading" class="loading-overlay">
        <i class="fas fa-spinner fa-spin fa-2x"></i>
        <span>{{ isProcessing ? 'Assigning car...' : 'Loading...' }}</span>
      </div>

      <div class="form-header">
        <h3>
          <i class="fas fa-link"></i>
          {{ t('sellBills.assign_car_to_bill') }}
        </h3>
        <button @click="$emit('close')" class="close-btn">&times;</button>
      </div>

      <div v-if="error" class="error">{{ error }}</div>

      <div v-if="carDetails" class="car-details">
        <div class="detail-item">
          <span class="label">
            <i class="fas fa-car"></i>
            {{ t('sellBills.car') }}:
          </span>
          <span class="value">{{ carDetails.car_name }}</span>
        </div>
        <div class="detail-item">
          <span class="label">
            <i class="fas fa-palette"></i>
            {{ t('sellBills.color') }}:
          </span>
          <span class="value">{{ carDetails.color }}</span>
        </div>
        <div class="detail-item">
          <span class="label">
            <i class="fas fa-fingerprint"></i>
            {{ t('sellBills.vin') }}:
          </span>
          <span class="value">{{ carDetails.vin || 'N/A' }}</span>
        </div>
        <div v-if="isAdmin" class="detail-item">
          <span class="label">
            <i class="fas fa-dollar-sign"></i>
            {{ t('sellBills.buy_price') }}:
          </span>
          <span class="value">${{ carDetails.buy_price }}</span>
        </div>
      </div>

      <form @submit.prevent="assignCar">
        <div class="form-group">
          <label for="client">
            <i class="fas fa-user"></i>
            {{ t('sellBills.client') }}: <span class="required">*</span>
          </label>
          <div class="client-select-wrapper">
            <el-select
              v-model="formData.id_client"
              filterable
              remote
              :remote-method="remoteMethod"
              :loading="loading"
              :placeholder="t('sellBills.select_or_search_client')"
              style="width: 100%"
              @change="handleClientChange"
            >
              <el-option
                v-for="client in filteredClients"
                :key="client.id"
                :label="client.name"
                :value="client.id"
              >
                <i class="fas fa-user"></i>
                {{ client.name }}
                <small v-if="client.mobiles">
                  <i class="fas fa-phone"></i>
                  {{ client.mobiles }}
                </small>
              </el-option>
            </el-select>
            <el-button type="primary" @click="showAddDialog = true" class="new-client-btn">
              <i class="fas fa-plus"></i>
              {{ t('sellBills.new') }}
            </el-button>
          </div>
        </div>

        <div class="form-group">
          <label for="discharge-port">
            <i class="fas fa-anchor"></i>
            {{ t('sellBills.discharge_port') }}: <span class="required">*</span>
          </label>
          <select id="discharge-port" v-model="formData.id_port_discharge" required>
            <option value="">{{ t('sellBills.select_discharge_port') }}</option>
            <option v-for="port in dischargePorts" :key="port.id" :value="port.id">
              {{ port.discharge_port }}
            </option>
          </select>
        </div>
        <div class="form-group">
          <label for="cfr-da">
            <i class="fas fa-calculator"></i>
            {{ t('sellBills.cfr_da') }} (Base Value): <span class="required">*</span>
          </label>
          <div class="input-with-info">
            <input
              type="number"
              id="cfr-da"
              v-model="cfrDaInput"
              step="0.01"
              min="0"
              required
              @input="handleCFRDAChange"
              :class="{ 'base-value': true }"
              placeholder="Enter CFR DA amount"
            />
          </div>
          <span class="info-text">This value will be used to calculate the sell price</span>
        </div>
        <div class="form-group">
          <label for="sell-price">
            <i class="fas fa-dollar-sign"></i>
            {{ t('sellBills.sell_price_required') }} (Calculated):
          </label>
          <div class="input-with-info">
            <input
              type="number"
              id="sell-price"
              v-model="formData.price_cell"
              step="0.01"
              min="0"
              required
              @input="handlePriceChange"
              :class="{
                'calculated-value': true,
                'default-value':
                  carDetails?.price_cell && formData.price_cell === carDetails.price_cell,
              }"
              readonly
            />
            <span
              v-if="carDetails?.price_cell && formData.price_cell === carDetails.price_cell"
              class="default-badge"
            >
              {{ t('sellBills.default') }}
            </span>
          </div>
          <span class="info-text">Calculated from CFR DA</span>
        </div>

        <div class="form-group">
          <label for="freight">
            <i class="fas fa-ship"></i>
            {{ t('sellBills.freight_cost') }}:
          </label>
          <div class="input-with-info">
            <input
              type="number"
              id="freight"
              v-model="formData.freight"
              step="0.01"
              min="0"
              :class="{ 'default-value': formData.freight !== null }"
            />
            <span v-if="formData.freight !== null" class="default-badge">
              {{ t('sellBills.default') }} ({{
                carDetails?.is_big_car
                  ? t('sellBills.default_big_car')
                  : t('sellBills.default_small_car')
              }})
            </span>
          </div>
        </div>

        <div class="form-group">
          <label for="rate">
            <i class="fas fa-exchange-alt"></i>
            {{ t('sellBills.rate_required') }}:
          </label>
          <div class="input-with-info">
            <input
              type="number"
              id="rate"
              v-model="formData.rate"
              step="0.01"
              min="0"
              required
              :class="{ 'default-value': formData.rate !== null }"
            />
            <span v-if="formData.rate !== null" class="default-badge">{{
              t('sellBills.default')
            }}</span>
          </div>
        </div>

        <div class="form-group">
          <label for="notes">
            <i class="fas fa-sticky-note"></i>
            {{ t('sellBills.notes') }}:
          </label>
          <textarea
            id="notes"
            v-model="formData.notes"
            :placeholder="t('sellBills.add_notes_about_car')"
            rows="3"
            class="textarea-field"
            :disabled="isSubmitting"
          ></textarea>
        </div>

        <div v-if="can_assign_to_tmp_clients" class="form-group">
          <label class="checkbox-label">
            <input
              type="checkbox"
              v-model="formData.is_tmp_client"
              :true-value="1"
              :false-value="0"
              :disabled="isSubmitting"
            />
            <i class="fas fa-clock"></i>
            {{ t('sellBills.temporary_client') }}
          </label>
          <small class="help-text">{{ t('sellBills.help_text_temporary_client') }}</small>
        </div>

        <div
          v-if="formData.price_cell && carDetails?.buy_price && isAdmin"
          class="profit-calculation"
        >
          <span class="label">
            <i class="fas fa-dollar-sign"></i>
            {{ t('sellBills.estimated_profit') }}:
          </span>
          <span class="value profit">${{ calculateProfit() }}</span>
        </div>

        <div class="form-actions">
          <button type="button" @click="$emit('close')" class="cancel-btn">
            <i class="fas fa-times"></i>
            {{ t('sellBills.cancel') }}
          </button>
          <button type="submit" class="assign-btn" :disabled="isProcessing">
            <i class="fas fa-save"></i>
            {{ isProcessing ? t('sellBills.assigning_car') : t('sellBills.assign_car') }}
          </button>
        </div>
      </form>
    </div>
  </div>

  <!-- New Client Dialog -->
  <el-dialog
    v-model="showAddDialog"
    :title="t('sellBills.add_new_client')"
    width="60%"
    :close-on-click-modal="false"
    append-to-body
  >
    <div v-if="validationError" class="error">{{ validationError }}</div>

    <div class="form-container">
      <div class="form-group">
        <label>{{ t('sellBills.name') }}</label>
        <el-input v-model="newClient.name" :placeholder="t('sellBills.enter_client_name')" />
      </div>

      <div class="form-group">
        <label>{{ t('sellBills.address') }}</label>
        <el-input v-model="newClient.address" :placeholder="t('sellBills.enter_address')" />
      </div>

      <div class="form-group">
        <label>{{ t('sellBills.phone') }}</label>
        <el-input
          v-model="newClient.email"
          :placeholder="t('sellBills.enter_email')"
          type="email"
        />
      </div>

      <div class="form-group">
        <label>{{ t('sellBills.mobile_numbers') }}</label>
        <el-input v-model="newClient.mobiles" :placeholder="t('sellBills.enter_mobile_numbers')" />
      </div>

      <div class="form-group">
        <label>{{ t('sellBills.id_number') }}</label>
        <el-input v-model="newClient.id_no" :placeholder="t('sellBills.enter_id_number')" />
      </div>

      <div class="form-group">
        <label>{{ t('sellBills.id_document_copy') }}</label>
        <input type="file" @change="handleFileChange" accept="image/*,.pdf" class="file-input" />
      </div>

      <div class="form-group">
        <label>{{ t('sellBills.notes') }}</label>
        <el-input
          v-model="newClient.notes"
          type="textarea"
          :placeholder="t('sellBills.enter_additional_notes')"
          :rows="3"
        />
      </div>

      <div class="form-group">
        <el-checkbox v-model="newClient.is_broker">{{ t('sellBills.is_broker') }}</el-checkbox>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="showAddDialog = false">{{ t('sellBills.cancel') }}</el-button>
        <el-button type="primary" @click="addClient" :loading="isSubmitting">
          <i class="fas fa-save"></i>
          {{ t('sellBills.add_client') }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<style scoped>
.car-assignment-form-overlay {
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

.car-assignment-form {
  background-color: white;
  border-radius: 8px;
  padding: 20px;
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.form-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.form-header h3 {
  margin: 0;
  font-size: 1.2rem;
  color: #333;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: #666;
}

.car-details {
  background-color: #f9fafb;
  border-radius: 4px;
  padding: 15px;
  margin-bottom: 20px;
}

.detail-item {
  display: flex;
  margin-bottom: 8px;
}

.detail-item:last-child {
  margin-bottom: 0;
}

.label {
  font-weight: 600;
  width: 100px;
}

.form-group {
  margin-bottom: 15px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
  font-weight: 500;
}

.required {
  color: #ef4444;
}

select,
input {
  width: 100%;
  padding: 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 0.9rem;
}

.profit-calculation {
  background-color: #ecfdf5;
  padding: 10px;
  border-radius: 4px;
  margin: 15px 0;
  display: flex;
  align-items: center;
}

.profit {
  color: #10b981;
  font-weight: 600;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 20px;
}

.cancel-btn {
  background-color: #6b7280;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 8px 16px;
  cursor: pointer;
}

.assign-btn {
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 8px 16px;
  cursor: pointer;
}

.assign-btn:disabled {
  background-color: #93c5fd;
  cursor: not-allowed;
  opacity: 0.6;
}

.spinner {
  display: inline-block;
  width: 16px;
  height: 16px;
  border: 2px solid #ffffff;
  border-radius: 50%;
  border-top-color: transparent;
  animation: spin 1s ease-in-out infinite;
  margin-right: 8px;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.error {
  color: #ef4444;
  background-color: #fee2e2;
  padding: 10px;
  border-radius: 4px;
  margin-bottom: 15px;
}

.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.9);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  z-index: 10;
}

.form-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #1f2937;
  margin-bottom: 1.5rem;
}

.error-message {
  background-color: #fee2e2;
  color: #dc2626;
  padding: 1rem;
  border-radius: 0.5rem;
  margin-bottom: 1rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.form-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-group label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #374151;
  font-weight: 500;
}

.form-group label i {
  color: #6b7280;
}

.car-details {
  background-color: #f9fafb;
  padding: 1rem;
  border-radius: 0.5rem;
  border: 1px solid #e5e7eb;
}

.car-details p {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin: 0.5rem 0;
  color: #374151;
}

.car-details i {
  color: #6b7280;
}

.client-select {
  width: 100%;
}

input[type='number'],
select,
textarea {
  padding: 0.5rem;
  border: 1px solid #d1d5db;
  border-radius: 0.375rem;
  font-size: 1rem;
  transition: border-color 0.2s;
}

input[type='number']:focus,
select:focus,
textarea:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

textarea {
  min-height: 100px;
  resize: vertical;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  margin-top: 1rem;
}

.cancel-btn,
.save-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 0.375rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.cancel-btn {
  background-color: #f3f4f6;
  color: #374151;
}

.cancel-btn:hover:not(:disabled) {
  background-color: #e5e7eb;
}

.save-btn {
  background-color: #10b981;
  color: white;
}

.save-btn:hover:not(:disabled) {
  background-color: #059669;
}

button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Element Plus Select Customization */
:deep(.el-select-dropdown__item) {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

:deep(.el-select-dropdown__item i) {
  color: #6b7280;
}

:deep(.el-select-dropdown__item small) {
  margin-left: auto;
  color: #6b7280;
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

/* Add smooth transitions */
.form-group input,
.form-group select,
.form-group textarea {
  transition: all 0.2s;
}

button i {
  transition: transform 0.2s;
}

button:hover:not(:disabled) i {
  transform: scale(1.1);
}

.calculated-value {
  background-color: #f3f4f6;
  color: #1f2937;
  font-weight: 500;
  font-family: monospace;
}

.base-value {
  background-color: #dbeafe;
  color: #1e40af;
  font-weight: 600;
  border: 2px solid #3b82f6;
}

.info-text {
  color: #6b7280;
  font-size: 0.75rem;
  font-style: italic;
  margin-top: 4px;
}

.input-with-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.input-with-info input {
  flex: 1;
}

.client-select-wrapper {
  display: flex;
  gap: 0.5rem;
  align-items: flex-start;
}

.new-client-btn {
  white-space: nowrap;
  height: 32px;
  padding: 0 12px;
}

.file-input {
  width: 100%;
  padding: 8px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  margin-top: 1rem;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  user-select: none;
  font-weight: 500;
  color: #374151;
}

.checkbox-label input[type='checkbox'] {
  width: 16px;
  height: 16px;
  cursor: pointer;
  accent-color: #3b82f6;
}

.checkbox-label input[type='checkbox']:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.help-text {
  color: #6b7280;
  font-size: 0.875rem;
  margin-top: 0.25rem;
  margin-left: 24px;
}
</style>
