<script setup>
import { ref, onMounted, watch, defineProps, defineEmits, computed } from 'vue'
import { useApi } from '../../composables/useApi'
import { useEnhancedI18n } from '@/composables/useI18n'

const { t } = useEnhancedI18n()

const props = defineProps({
  carData: {
    type: Object,
    required: true,
  },
})

const user = ref(null)
const can_edit_vin = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_edit_vin')
})

const can_edit_car_client_name = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_edit_car_client_name')
})

const can_edit_cars_sell_price = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_edit_cars_sell_price')
})

const can_edit_cars_sell_rate = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_edit_cars_sell_rate')
})

const can_edit_cars_discharge_port = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some((p) => p.permission_name === 'can_edit_cars_discharge_port')
})

const emit = defineEmits(['save', 'cancel'])

const { callApi, uploadFile, getFileUrl } = useApi()
const loading = ref(false)
const error = ref(null)

// File upload refs
const documentsFile = ref(null)
const sellPiFile = ref(null)
const buyPiFile = ref(null)

// Reference data
const clients = ref([])
const loadingPorts = ref([])
const dischargePorts = ref([])
const buyDetails = ref([])
const warehouses = ref([])

const formData = ref({
  id: null,
  notes: '',
  id_buy_details: null,
  date_sell: null,
  id_client: null,
  price_cell: null,
  freight: null,
  id_port_loading: null,
  id_port_discharge: null,
  vin: '',
  path_documents: '',
  date_loding: null,
  date_send_documents: null,
  id_sell_pi: '',
  sell_pi_path: '',
  buy_pi_path: '',
  id_sell: null,
  export_lisence_ref: '',
  id_warehouse: null,
  in_wharhouse_date: null,
  date_get_documents_from_supp: null,
  date_get_keys_from_supp: null,
  rate: null,
})

// Handle file changes
const handleFileChange = (event, fileType) => {
  const file = event.target.files[0]
  if (!file) return

  switch (fileType) {
    case 'documents':
      documentsFile.value = file
      break
    case 'sellPi':
      sellPiFile.value = file
      break
    case 'buyPi':
      buyPiFile.value = file
      break
  }
}

// Upload a single file and return its path
const uploadSingleFile = async (file, folder, customFilename = '') => {
  if (!file) return null

  try {
    // Get original file extension
    const originalExt = file.name.split('.').pop().toLowerCase()

    // Create a meaningful filename based on car details
    let filename = ''
    if (formData.value.vin) {
      // If VIN exists, use it as the primary identifier
      filename = `${formData.value.vin}`
    } else if (formData.value.id) {
      // If no VIN but ID exists, use car ID
      filename = `car_${formData.value.id}`
    } else {
      // Fallback to timestamp if neither exists
      filename = `new_car_${Date.now()}`
    }

    // Add folder-specific prefix
    switch (folder) {
      case 'documents':
        filename = `docs_${filename}`
        break
      case 'sell_pi':
        filename = `sell_pi_${filename}`
        break
      case 'buy_pi':
        filename = `buy_pi_${filename}`
        break
    }

    // Add date for versioning
    const date = new Date()
    const dateStr = `${date.getFullYear()}${String(date.getMonth() + 1).padStart(2, '0')}${String(date.getDate()).padStart(2, '0')}`
    filename = `${filename}_${dateStr}.${originalExt}`

    const result = await uploadFile(file, folder, filename)
    if (result.success) {
      return result.relativePath
    }
    throw new Error(result.message || 'Upload failed')
  } catch (err) {
    console.error(`Error uploading file to ${folder}:`, err)
    throw err
  }
}

// Watch for changes in carData prop
watch(
  () => props.carData,
  (newData) => {
    if (newData) {
      // Create a new object with all fields from newData
      formData.value = {
        ...newData,
        // Convert rate to number if it exists, otherwise null
        rate: newData.rate ? Number(newData.rate) : null,
      }

      // Format dates for input fields
      const dateFields = [
        'date_sell',
        'date_loding',
        'date_send_documents',
        'in_wharhouse_date',
        'date_get_documents_from_supp',
        'date_get_keys_from_supp',
      ]

      dateFields.forEach((field) => {
        if (formData.value[field]) {
          const date = new Date(formData.value[field])
          if (!isNaN(date)) {
            formData.value[field] = date.toISOString().split('T')[0]
          }
        }
      })
    }
  },
  { immediate: true },
)

const fetchReferenceData = async () => {
  try {
    // Fetch clients
    const clientsResult = await callApi({
      query: 'SELECT id, name FROM clients ORDER BY name ASC',
      params: [],
    })
    if (clientsResult.success) {
      clients.value = clientsResult.data
    }

    // Fetch loading ports
    const loadingPortsResult = await callApi({
      query: 'SELECT id, loading_port FROM loading_ports ORDER BY loading_port ASC',
      params: [],
    })
    if (loadingPortsResult.success) {
      loadingPorts.value = loadingPortsResult.data
    }

    // Fetch discharge ports
    const dischargePortsResult = await callApi({
      query: 'SELECT id, discharge_port FROM discharge_ports ORDER BY discharge_port ASC',
      params: [],
    })
    if (dischargePortsResult.success) {
      dischargePorts.value = dischargePortsResult.data
    }

    // Fetch buy details
    const buyDetailsResult = await callApi({
      query: `
        SELECT bd.id, cn.car_name, clr.color 
        FROM buy_details bd
        LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
        LEFT JOIN colors clr ON bd.id_color = clr.id
        ORDER BY bd.id DESC
      `,
      params: [],
    })
    if (buyDetailsResult.success) {
      buyDetails.value = buyDetailsResult.data
    }

    // Fetch warehouses if needed
    const warehousesResult = await callApi({
      query: 'SELECT id, warhouse_name as name FROM warehouses ORDER BY warhouse_name ASC',
      params: [],
    })
    if (warehousesResult.success) {
      warehouses.value = warehousesResult.data
    }
  } catch (err) {
    error.value = t('carStockForm.failedToFetchReferenceData')
  }
}

const saveCar = async () => {
  loading.value = true
  error.value = null

  try {
    // Handle file uploads first
    if (documentsFile.value) {
      formData.value.path_documents = await uploadSingleFile(documentsFile.value, 'documents')
    }

    if (sellPiFile.value) {
      formData.value.sell_pi_path = await uploadSingleFile(sellPiFile.value, 'sell_pi')
    }

    if (buyPiFile.value) {
      formData.value.buy_pi_path = await uploadSingleFile(buyPiFile.value, 'buy_pi')
    }

    const result = await callApi({
      query: `
        UPDATE cars_stock
        SET 
          notes = ?,
          id_buy_details = ?,
          date_sell = ?,
          id_client = ?,
          price_cell = ?,
          freight = ?,
          id_port_loading = ?,
          id_port_discharge = ?,
          vin = ?,
          path_documents = ?,
          date_loding = ?,
          date_send_documents = ?,
          id_sell_pi = ?,
          sell_pi_path = ?,
          buy_pi_path = ?,
          id_sell = ?,
          export_lisence_ref = ?,
          id_warehouse = ?,
          in_wharhouse_date = ?,
          date_get_documents_from_supp = ?,
          date_get_keys_from_supp = ?,
          rate = ?
        WHERE id = ?
      `,
      params: [
        formData.value.notes || null,
        formData.value.id_buy_details || null,
        formData.value.date_sell || null,
        formData.value.id_client || null,
        formData.value.price_cell || null,
        formData.value.freight || null,
        formData.value.id_port_loading || null,
        formData.value.id_port_discharge || null,
        formData.value.vin || null,
        formData.value.path_documents || null,
        formData.value.date_loding || null,
        formData.value.date_send_documents || null,
        formData.value.id_sell_pi || null,
        formData.value.sell_pi_path || null,
        formData.value.buy_pi_path || null,
        formData.value.id_sell || null,
        formData.value.export_lisence_ref || null,
        formData.value.id_warehouse || null,
        formData.value.in_wharhouse_date || null,
        formData.value.date_get_documents_from_supp || null,
        formData.value.date_get_keys_from_supp || null,
        formData.value.rate || null,
        formData.value.id,
      ],
    })

    if (result.success) {
      emit('save')
    } else {
      throw new Error(result.error || t('carStockForm.failedToSaveCar'))
    }
  } catch (err) {
    error.value = err.message
    console.error('Error saving car:', err)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
  }
  fetchReferenceData()
})
</script>

<template>
  <div class="car-stock-form">
    <h3>{{ formData.id ? t('carStockForm.editCarStock') : t('carStockForm.addCarStock') }}</h3>

    <div class="form-grid">
      <!-- Basic Information -->
      <div class="form-section">
        <h4>{{ t('carStockForm.basicInformation') }}</h4>

        <div class="form-group">
          <label for="vin">{{ t('carStockForm.vin') }}:</label>
          <input :disabled="!can_edit_vin" type="text" id="vin" v-model="formData.vin" />
        </div>

        <div class="form-group">
          <label for="id_buy_details">{{ t('carStockForm.buyDetails') }}:</label>
          <select id="id_buy_details" v-model="formData.id_buy_details">
            <option value="">{{ t('carStockForm.selectBuyDetails') }}</option>
            <option v-for="detail in buyDetails" :key="detail.id" :value="detail.id">
              {{ detail.car_name }} - {{ detail.color }} (ID: {{ detail.id }})
            </option>
          </select>
        </div>

        <div class="form-group">
          <label for="notes">{{ t('carStockForm.notes') }}:</label>
          <textarea id="notes" v-model="formData.notes"></textarea>
        </div>
      </div>

      <!-- Pricing and Client -->
      <div class="form-section">
        <h4>{{ t('carStockForm.pricingAndClient') }}</h4>

        <div class="form-group">
          <label for="price_cell">{{ t('carStockForm.priceCell') }}:</label>
          <input
            :disabled="!can_edit_cars_sell_price"
            type="number"
            id="price_cell"
            v-model="formData.price_cell"
            step="0.01"
          />
        </div>

        <div class="form-group">
          <label for="freight">{{ t('carStockForm.freight') }}:</label>
          <input type="number" id="freight" v-model="formData.freight" step="0.01" />
        </div>

        <div class="form-group">
          <label for="rate">{{ t('carStockForm.rate') }}:</label>
          <input
            :disabled="!can_edit_cars_sell_rate"
            type="number"
            id="rate"
            v-model="formData.rate"
            step="0.01"
          />
        </div>

        <div class="form-group">
          <label for="id_client">{{ t('carStockForm.client') }}:</label>
          <select :disabled="!can_edit_car_client_name" id="id_client" v-model="formData.id_client">
            <option value="">{{ t('carStockForm.selectClient') }}</option>
            <option v-for="client in clients" :key="client.id" :value="client.id">
              {{ client.name }}
            </option>
          </select>
        </div>

        <div class="form-group">
          <label for="date_sell">{{ t('carStockForm.sellDate') }}:</label>
          <input type="date" id="date_sell" v-model="formData.date_sell" />
        </div>
      </div>

      <!-- Ports and Shipping -->
      <div class="form-section">
        <h4>{{ t('carStockForm.portsAndShipping') }}</h4>

        <div class="form-group">
          <label for="id_port_loading">{{ t('carStockForm.loadingPort') }}:</label>
          <select id="id_port_loading" v-model="formData.id_port_loading">
            <option value="">{{ t('carStockForm.selectLoadingPort') }}</option>
            <option v-for="port in loadingPorts" :key="port.id" :value="port.id">
              {{ port.loading_port }}
            </option>
          </select>
        </div>

        <div class="form-group">
          <label for="id_port_discharge">{{ t('carStockForm.dischargePort') }}:</label>
          <select
            :disabled="!can_edit_cars_discharge_port"
            id="id_port_discharge"
            v-model="formData.id_port_discharge"
          >
            <option value="">{{ t('carStockForm.selectDischargePort') }}</option>
            <option v-for="port in dischargePorts" :key="port.id" :value="port.id">
              {{ port.discharge_port }}
            </option>
          </select>
        </div>

        <div class="form-group">
          <label for="date_loding">{{ t('carStockForm.operationDate') }}:</label>
          <input type="date" id="date_loding" v-model="formData.date_loding" />
        </div>

        <div class="form-group">
          <label for="export_lisence_ref">{{ t('carStockForm.exportLicenseRef') }}:</label>
          <input type="text" id="export_lisence_ref" v-model="formData.export_lisence_ref" />
        </div>
      </div>

      <!-- Documents Upload -->
      <div class="form-group">
        <label>{{ t('carStockForm.documents') }}:</label>
        <div class="file-upload-container">
          <input
            type="file"
            @change="handleFileChange($event, 'documents')"
            accept=".pdf"
            class="file-input"
          />
          <a
            v-if="formData.path_documents"
            :href="getFileUrl(formData.path_documents)"
            target="_blank"
            class="current-file-link"
          >
            {{ t('carStockForm.viewCurrentDocument') }}
          </a>
        </div>
      </div>

      <!-- Sell PI Upload -->
      <div class="form-group">
        <label>{{ t('carStockForm.sellPi') }}:</label>
        <div class="file-upload-container">
          <input
            type="file"
            @change="handleFileChange($event, 'sellPi')"
            accept=".pdf"
            class="file-input"
          />
          <a
            v-if="formData.sell_pi_path"
            :href="getFileUrl(formData.sell_pi_path)"
            target="_blank"
            class="current-file-link"
          >
            {{ t('carStockForm.viewSellPi') }}
          </a>
        </div>
      </div>

      <!-- Buy PI Upload -->
      <div class="form-group">
        <label>{{ t('carStockForm.buyPi') }}:</label>
        <div class="file-upload-container">
          <input
            type="file"
            @change="handleFileChange($event, 'buyPi')"
            accept=".pdf"
            class="file-input"
          />
          <a
            v-if="formData.buy_pi_path"
            :href="getFileUrl(formData.buy_pi_path)"
            target="_blank"
            class="current-file-link"
          >
            {{ t('carStockForm.viewBuyPi') }}
          </a>
        </div>
      </div>
    </div>

    <div class="form-actions">
      <button @click="saveCar" :disabled="loading" class="btn save-btn">
        {{ loading ? t('carStockForm.saving') : t('carStockForm.save') }}
      </button>
      <button @click="$emit('cancel')" class="btn cancel-btn">
        {{ t('carStockForm.cancel') }}
      </button>
    </div>

    <div v-if="error" class="error-message">{{ error }}</div>
  </div>
</template>

<style scoped>
.car-stock-form {
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;
}

.error {
  color: #ef4444;
  margin-bottom: 16px;
  padding: 8px;
  background-color: #fee2e2;
  border-radius: 4px;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: 24px;
  margin-bottom: 24px;
}

.form-section {
  background-color: #f8f9fa;
  padding: 16px;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.form-section h4 {
  margin-top: 0;
  margin-bottom: 16px;
  color: #4b5563;
  border-bottom: 1px solid #e5e7eb;
  padding-bottom: 8px;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 4px;
  font-weight: 500;
  color: #4b5563;
}

.form-group input,
.form-group select,
.form-group textarea {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 14px;
}

.form-group textarea {
  min-height: 80px;
  resize: vertical;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
}

.cancel-btn {
  background-color: #f3f4f6;
  color: #4b5563;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  padding: 8px 16px;
  cursor: pointer;
  font-size: 14px;
}

.save-btn {
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 8px 16px;
  cursor: pointer;
  font-size: 14px;
}

.save-btn:hover {
  background-color: #2563eb;
}

.save-btn:disabled {
  background-color: #93c5fd;
  cursor: not-allowed;
}

.file-upload-container {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.file-input {
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  width: 100%;
}

.current-file-link {
  color: #3b82f6;
  text-decoration: none;
  font-size: 0.9em;
  display: inline-block;
  padding: 4px 0;
}

.current-file-link:hover {
  text-decoration: underline;
}

.error-message {
  color: #ef4444;
  margin-top: 16px;
  font-size: 0.9em;
}
</style>
