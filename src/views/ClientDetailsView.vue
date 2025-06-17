<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useApi } from '../composables/useApi'
import { ElMessage } from 'element-plus'

const route = useRoute()
const { callApi, error, getFileUrl } = useApi()
const client = ref(null)
const clientCars = ref([])
const isLoading = ref(false)
const shareUrl = ref(window.location.href)
const isDev = ref(process.env.NODE_ENV === 'development')
const user = ref(JSON.parse(localStorage.getItem('user')))
const isAdmin = computed(() => user.value?.role_id === 1)
const copyToClipboard = async () => {
  try {
    await navigator.clipboard.writeText(shareUrl.value)
    ElMessage({
      message: 'Link copied to clipboard!',
      type: 'success',
    })
  } catch (err) {
    ElMessage({
      message: 'Failed to copy link',
      type: 'error',
    })
  }
}

const fetchClientDetails = async () => {
  isLoading.value = true
  try {
    const [clientResult, carsResult] = await Promise.all([
      callApi({
        query: `
          SELECT 
            c.id,
            c.name,
            c.email,
            c.mobiles,
            c.id_no,
            c.address,
            c.is_broker,
            COUNT(cs.id) as cars_count
          FROM clients c
          LEFT JOIN cars_stock cs ON c.id = cs.id_client
          WHERE c.id = ?
          GROUP BY c.id
        `,
        params: [route.params.id],
        requiresAuth: false,
      }),
      callApi({
        query: `
          SELECT 
            cs.*,
            w.warhouse_name as warehouse_name,
            bd.amount as buy_price,
            bd.year,
            bd.is_used_car,
            bd.is_big_car,
            cn.car_name as model,
            b.brand,
            c.color,
            dp.discharge_port,
            COALESCE(
              (SELECT SUM(sp.amount_da) 
               FROM sell_bill sb 
               JOIN sell_payments sp ON sp.id_sell_bill = sb.id 
               WHERE sb.id = cs.id_sell),
              0
            ) as total_paid,
            CASE 
              WHEN cs.id_client IS NOT NULL THEN 'Reserved'
              ELSE 'Available'
            END as status
          FROM cars_stock cs
          LEFT JOIN warehouses w ON cs.id_warehouse = w.id
          LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
          LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
          LEFT JOIN brands b ON cn.id_brand = b.id
          LEFT JOIN colors c ON bd.id_color = c.id
          LEFT JOIN discharge_ports dp ON cs.id_port_discharge = dp.id
          WHERE cs.id_client = ? AND cs.hidden = 0
          ORDER BY cs.in_wharhouse_date DESC, cs.id DESC
        `,
        params: [route.params.id],
        requiresAuth: false,
      }),
    ])

    console.log('Client Result:', clientResult)
    console.log('Cars Result:', carsResult)

    if (clientResult.success && clientResult.data.length > 0) {
      client.value = clientResult.data[0]
    } else {
      error.value = 'Client not found'
    }

    if (carsResult.success) {
      clientCars.value = carsResult.data
      console.log('Processed Cars:', clientCars.value)
    }
  } catch (err) {
    console.error('Error fetching client details:', err)
    error.value = 'Failed to load client details'
  } finally {
    isLoading.value = false
  }
}

const formatCurrency = (value) => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(value)
}

const formatDate = (dateString) => {
  return new Date(dateString).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

onMounted(() => {
  fetchClientDetails()
})
</script>

<template>
  <div class="client-details">
    <div class="header">
      <div class="header-left">
        <h2>
          <i class="fas fa-user"></i>
          Client Details
        </h2>
        <router-link to="/clients" class="back-btn" v-if="$route.query.from === 'clients'">
          <i class="fas fa-arrow-left"></i>
          Back to Clients
        </router-link>
      </div>
      <div class="share-section">
        <div class="share-url">
          <input type="text" :value="shareUrl" readonly />
          <button @click="copyToClipboard" class="copy-btn">
            <i class="fas fa-copy"></i>
            Copy Link
          </button>
        </div>
      </div>
    </div>

    <!-- Debug Information -->
    <div v-if="false" class="debug-section">
      <div class="info-section">
        <h3>Debug Information</h3>
        <div class="debug-info">
          <h4>Client ID: {{ $route.params.id }}</h4>
          <pre>{{ client }}</pre>
          <h4>Cars Data:</h4>
          <pre>{{ clientCars }}</pre>
        </div>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="isLoading" class="loading-state">
      <i class="fas fa-spinner fa-spin"></i>
      Loading client details...
    </div>

    <!-- Error Message -->
    <div v-if="error" class="error-message">
      <i class="fas fa-exclamation-circle"></i>
      {{ error }}
    </div>

    <div v-if="client && !isLoading" class="client-info">
      <div class="info-section">
        <h3>Basic Information</h3>
        <div class="info-grid">
          <div class="info-item">
            <label>Client ID:</label>
            <span>{{ client.id }}</span>
          </div>
          <div class="info-item">
            <label>Name:</label>
            <span>{{ client.name }}</span>
          </div>
          <div class="info-item">
            <label>Email:</label>
            <span>{{ client.email || 'Not provided' }}</span>
          </div>
          <div class="info-item">
            <label>Mobile:</label>
            <span>{{ client.mobiles }}</span>
          </div>
          <div class="info-item">
            <label>ID Number:</label>
            <span>{{ client.id_no }}</span>
          </div>
          <div class="info-item">
            <label>Address:</label>
            <span>{{ client.address || 'Not provided' }}</span>
          </div>
          <div class="info-item">
            <label>Status:</label>
            <div class="status-badges">
              <span class="badge client">
                <i class="fas fa-user"></i>
                Client
              </span>
              <span v-if="client.is_broker" class="badge broker">
                <i class="fas fa-user-tie"></i>
                Broker
              </span>
            </div>
          </div>
          <div class="info-item">
            <label>Cars Count:</label>
            <span class="badge cars" :class="{ 'has-cars': client.cars_count > 0 }">
              <i class="fas fa-car"></i>
              {{ client.cars_count }}
            </span>
          </div>
        </div>
      </div>

      <!-- Cars Section -->
      <div v-if="clientCars.length > 0" class="info-section cars-section">
        <h3>
          <i class="fas fa-car"></i>
          Cars
        </h3>
        <div class="cars-grid">
          <div v-for="car in clientCars" :key="car.id" class="car-card">
            <div class="car-details">
              <h4>{{ car.brand || 'Unknown Brand' }} {{ car.model || '' }}</h4>
              <div class="car-info-grid">
                <div class="car-info-item">
                  <label>VIN:</label>
                  <span>{{ car.vin || 'Not available' }}</span>
                </div>
                <div class="car-info-item">
                  <label>Year:</label>
                  <span>{{ car.year || 'Not specified' }}</span>
                </div>
                <div class="car-info-item">
                  <label>Color:</label>
                  <span>{{ car.color || 'Not specified' }}</span>
                </div>
                <div class="car-info-item">
                  <label>Documents Path:</label>
                  <span>{{ car.path_documents || 'Not available' }}</span>
                </div>
                <div class="car-info-item">
                  <label>Loading Date:</label>
                  <div class="status-field">
                    <span>{{
                      car.date_loding ? formatDate(car.date_loding) : 'Not available'
                    }}</span>
                    <i
                      :class="[
                        'fas',
                        car.date_loding ? 'fa-check-circle done' : 'fa-times-circle undone',
                      ]"
                    ></i>
                  </div>
                </div>
                <div class="car-info-item">
                  <label>Documents Sent:</label>
                  <div class="status-field">
                    <span>{{
                      car.date_send_documents
                        ? formatDate(car.date_send_documents)
                        : 'Not available'
                    }}</span>
                    <i
                      :class="[
                        'fas',
                        car.date_send_documents ? 'fa-check-circle done' : 'fa-times-circle undone',
                      ]"
                    ></i>
                  </div>
                </div>
                <div class="car-info-item">
                  <label>Documents Received:</label>
                  <div class="status-field">
                    <span>{{
                      car.date_get_documents_from_supp
                        ? formatDate(car.date_get_documents_from_supp)
                        : 'Not available'
                    }}</span>
                    <i
                      :class="[
                        'fas',
                        car.date_get_documents_from_supp
                          ? 'fa-check-circle done'
                          : 'fa-times-circle undone',
                      ]"
                    ></i>
                  </div>
                </div>
                <div class="car-info-item">
                  <label>BL Date:</label>
                  <div class="status-field">
                    <span>{{
                      car.date_get_bl ? formatDate(car.date_get_bl) : 'Not available'
                    }}</span>
                    <i
                      :class="[
                        'fas',
                        car.date_get_bl ? 'fa-check-circle done' : 'fa-times-circle undone',
                      ]"
                    ></i>
                  </div>
                </div>
                <div class="car-info-item">
                  <label>Freight Payment:</label>
                  <div class="status-field">
                    <span>{{
                      car.date_pay_freight ? formatDate(car.date_pay_freight) : 'Not available'
                    }}</span>
                    <i
                      :class="[
                        'fas',
                        car.date_pay_freight ? 'fa-check-circle done' : 'fa-times-circle undone',
                      ]"
                    ></i>
                  </div>
                </div>
                <div class="car-info-item">
                  <label>Discharge Port:</label>
                  <span>{{ car.discharge_port || 'Not available' }}</span>
                </div>
                <div class="car-info-item">
                  <label>Price:</label>
                  <span>{{ formatCurrency(car.price_cell || 0) }}</span>
                </div>
                <div v-if="false" class="car-info-item">
                  <label>Paid:</label>
                  <span>{{ formatCurrency(car.total_paid || 0) }}</span>
                </div>
                <div v-if="false" class="car-info-item">
                  <label>Balance:</label>
                  <span
                    :class="{ 'text-danger': (car.price_cell || 0) - (car.total_paid || 0) > 0 }"
                  >
                    {{ formatCurrency((car.price_cell || 0) - (car.total_paid || 0)) }}
                  </span>
                </div>
                <div v-if="car.is_used_car" class="car-info-item">
                  <label>Type:</label>
                  <span class="badge used">Used Car</span>
                </div>
                <div v-if="car.is_big_car" class="car-info-item">
                  <label>Size:</label>
                  <span class="badge big">Big Car</span>
                </div>
              </div>
              <div class="car-status" :class="car.status.toLowerCase()">
                <i class="fas fa-circle"></i>
                {{ car.status }}
              </div>
              <div
                v-if="car.path_documents || car.sell_pi_path || car.buy_pi_path"
                class="car-documents"
              >
                <h5>Documents</h5>
                <div class="document-links">
                  <a
                    v-if="car.path_documents"
                    :href="getFileUrl(car.path_documents)"
                    target="_blank"
                    class="document-link"
                  >
                    <i class="fas fa-file-alt"></i>
                    Car Documents
                  </a>
                  <a
                    v-if="car.sell_pi_path"
                    :href="getFileUrl(car.sell_pi_path)"
                    target="_blank"
                    class="document-link"
                  >
                    <i class="fas fa-file-invoice-dollar"></i>
                    Sell PI
                  </a>
                  <a
                    v-if="car.buy_pi_path && isAdmin"
                    :href="getFileUrl(car.buy_pi_path)"
                    target="_blank"
                    class="document-link"
                  >
                    <i class="fas fa-file-invoice"></i>
                    Buy PI
                  </a>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div v-else-if="!isLoading" class="info-section no-cars">
        <div class="empty-state">
          <i class="fas fa-car fa-2x"></i>
          <p>No cars found for this client</p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.client-details {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 16px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.share-section {
  flex-grow: 1;
  max-width: 500px;
}

.share-url {
  display: flex;
  gap: 8px;
  background: white;
  padding: 8px;
  border-radius: 4px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.share-url input {
  flex-grow: 1;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  color: #666;
  background: #f5f5f5;
  cursor: default;
}

.copy-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background-color: #e3f2fd;
  color: #1976d2;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: background-color 0.2s;
}

.copy-btn:hover {
  background-color: #bbdefb;
}

.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background-color: #f0f0f0;
  border-radius: 4px;
  text-decoration: none;
  color: #333;
}

.back-btn:hover {
  background-color: #e0e0e0;
}

.loading-state {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 20px;
  color: #666;
}

.error-message {
  padding: 12px;
  background-color: #fee;
  color: #c00;
  border-radius: 4px;
  margin-bottom: 20px;
}

.client-info {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.info-section {
  background-color: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.info-section h3 {
  margin-bottom: 16px;
  padding-bottom: 8px;
  border-bottom: 1px solid #eee;
  display: flex;
  align-items: center;
  gap: 8px;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 16px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.info-item label {
  font-weight: 500;
  color: #666;
}

.status-badges {
  display: flex;
  gap: 8px;
}

.badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 0.9em;
}

.badge.client {
  background-color: #e3f2fd;
  color: #1976d2;
}

.badge.broker {
  background-color: #fce4ec;
  color: #c2185b;
}

.badge.cars {
  background-color: #f5f5f5;
  color: #666;
}

.badge.cars.has-cars {
  background-color: #e8f5e9;
  color: #2e7d32;
}

.badge.used {
  background-color: #fff3e0;
  color: #ef6c00;
}

.badge.big {
  background-color: #e8eaf6;
  color: #3f51b5;
}

/* Cars Section Styles */
.cars-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
  margin-top: 16px;
}

.car-card {
  background: white;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  transition: transform 0.2s;
}

.car-card:hover {
  transform: translateY(-2px);
}

.car-images {
  height: 200px;
  overflow: hidden;
  background: #f5f5f5;
}

.car-images img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.car-details {
  padding: 16px;
}

.car-details h4 {
  margin: 0 0 12px 0;
  font-size: 1.2em;
  color: #333;
}

.car-info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  margin-bottom: 12px;
}

.car-info-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.car-info-item label {
  font-size: 0.85em;
  color: #666;
  font-weight: 500;
}

.car-info-item span {
  font-size: 0.95em;
  color: #333;
}

.car-status {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 0.9em;
  margin-top: 8px;
}

.car-status.available {
  background-color: #e8f5e9;
  color: #2e7d32;
}

.car-status.sold {
  background-color: #fce4ec;
  color: #c2185b;
}

.car-status.pending {
  background-color: #fff3e0;
  color: #ef6c00;
}

.text-danger {
  color: #c2185b;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 40px;
  color: #666;
  text-align: center;
}

/* Mobile Responsive Styles */
@media (max-width: 768px) {
  .client-details {
    padding: 12px;
  }

  .header {
    flex-direction: column;
    align-items: flex-start;
  }

  .share-section {
    width: 100%;
    max-width: none;
  }

  .info-grid {
    grid-template-columns: 1fr;
  }

  .cars-grid {
    grid-template-columns: 1fr;
  }

  .car-info-grid {
    grid-template-columns: 1fr;
  }

  .car-images {
    height: 180px;
  }

  .car-details {
    padding: 12px;
  }

  .car-details h4 {
    font-size: 1.1em;
  }
}

/* Small Mobile Devices */
@media (max-width: 480px) {
  .client-details {
    padding: 8px;
  }

  .header-left {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .share-url {
    flex-direction: column;
    gap: 8px;
  }

  .copy-btn {
    width: 100%;
    justify-content: center;
  }

  .car-images {
    height: 160px;
  }
}

.debug-section {
  margin: 20px 0;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
  border: 1px solid #dee2e6;
}

.debug-info {
  font-family: monospace;
  white-space: pre-wrap;
  word-break: break-all;
}

.debug-info h4 {
  margin: 10px 0;
  color: #666;
}

.debug-info pre {
  background: #fff;
  padding: 10px;
  border-radius: 4px;
  border: 1px solid #dee2e6;
  overflow-x: auto;
}

.car-documents {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #eee;
}

.car-documents h5 {
  margin: 0 0 8px 0;
  font-size: 0.95em;
  color: #666;
}

.document-links {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.document-link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  background-color: #f5f5f5;
  color: #333;
  text-decoration: none;
  border-radius: 4px;
  font-size: 0.9em;
  transition: background-color 0.2s;
}

.document-link:hover {
  background-color: #e0e0e0;
}

.status-field {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-field i {
  font-size: 1.1em;
}

.status-field i.done {
  color: #4caf50;
}

.status-field i.undone {
  color: #f44336;
}
</style>
