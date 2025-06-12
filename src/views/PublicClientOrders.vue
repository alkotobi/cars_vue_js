<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useApi } from '../composables/useApi'

const route = useRoute()
const { callApi } = useApi()
const orders = ref([])
const client = ref(null)
const loading = ref(true)
const error = ref(null)

const fetchClientOrders = async () => {
  loading.value = true
  error.value = null

  try {
    const clientId = route.params.clientId
    if (!clientId) {
      throw new Error('Client ID is required')
    }

    // First fetch client info
    const clientResult = await callApi({
      query: 'SELECT id, name, phone, email FROM clients WHERE id = ?',
      params: [clientId],
    })

    if (!clientResult.success || !clientResult.data.length) {
      throw new Error('Client not found')
    }

    client.value = clientResult.data[0]

    // Then fetch orders
    const query = `
      SELECT 
        cs.id,
        cs.id_sell_pi,
        cs.date_sell,
        cs.price_cell,
        cs.freight,
        cs.vin,
        cs.date_loding as date_loading,
        cs.date_send_documents,
        cs.date_get_bl,
        cs.date_get_documents_from_supp,
        dp.discharge_port as port_discharge,
        cn.car_name,
        clr.color
      FROM cars_stock cs
      LEFT JOIN discharge_ports dp ON cs.id_port_discharge = dp.id
      LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
      LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
      LEFT JOIN colors clr ON bd.id_color = clr.id
      WHERE cs.id_client = ?
      AND cs.hidden = 0
      ORDER BY cs.date_sell DESC
    `

    const result = await callApi({
      query,
      params: [clientId],
    })

    if (result.success) {
      orders.value = result.data
    } else {
      throw new Error('Failed to fetch orders')
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

const formatDate = (date) => {
  if (!date) return '-'
  return new Date(date).toLocaleDateString()
}

const formatCurrency = (amount) => {
  if (!amount) return '-'
  return `$${parseFloat(amount).toLocaleString()}`
}

onMounted(() => {
  fetchClientOrders()
})
</script>

<template>
  <div class="public-orders">
    <!-- Loading State -->
    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>Loading orders...</p>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="error-state">
      <i class="fas fa-exclamation-circle"></i>
      <p>{{ error }}</p>
    </div>

    <!-- Content -->
    <div v-else class="content">
      <!-- Client Info -->
      <div class="client-info">
        <h1>{{ client.name }}</h1>
        <div class="client-details">
          <p v-if="client.phone">
            <i class="fas fa-phone"></i>
            {{ client.phone }}
          </p>
          <p v-if="client.email">
            <i class="fas fa-envelope"></i>
            {{ client.email }}
          </p>
        </div>
      </div>

      <!-- Orders -->
      <div class="orders-list">
        <div v-if="orders.length === 0" class="no-orders">
          <i class="fas fa-inbox"></i>
          <p>No orders found</p>
        </div>

        <div v-else class="order-cards">
          <div v-for="order in orders" :key="order.id" class="order-card">
            <div class="order-header">
              <span class="order-id">#{{ order.id_sell_pi || order.id }}</span>
              <span class="order-date">{{ formatDate(order.date_sell) }}</span>
            </div>

            <div class="car-info">
              <h3>{{ order.car_name }} - {{ order.color }}</h3>
              <p class="vin">VIN: {{ order.vin || 'Pending' }}</p>
            </div>

            <div class="financial-info">
              <div class="price">
                <span class="label">Price:</span>
                <span class="value">{{ formatCurrency(order.price_cell) }}</span>
              </div>
              <div class="freight">
                <span class="label">Freight:</span>
                <span class="value">{{ formatCurrency(order.freight) }}</span>
              </div>
            </div>

            <div class="shipping-info">
              <div class="port">
                <i class="fas fa-anchor"></i>
                <span>{{ order.port_discharge || 'Port pending' }}</span>
              </div>
              <div class="dates">
                <div class="date-item" :class="{ completed: order.date_loading }">
                  <i class="fas fa-truck-loading"></i>
                  <span>Loading: {{ formatDate(order.date_loading) }}</span>
                </div>
                <div class="date-item" :class="{ completed: order.date_get_bl }">
                  <i class="fas fa-file-contract"></i>
                  <span>BL Received: {{ formatDate(order.date_get_bl) }}</span>
                </div>
                <div class="date-item" :class="{ completed: order.date_get_documents_from_supp }">
                  <i class="fas fa-folder-open"></i>
                  <span>Supplier Docs: {{ formatDate(order.date_get_documents_from_supp) }}</span>
                </div>
                <div class="date-item" :class="{ completed: order.date_send_documents }">
                  <i class="fas fa-paper-plane"></i>
                  <span>Documents Sent: {{ formatDate(order.date_send_documents) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.public-orders {
  max-width: 100%;
  padding: 16px;
  min-height: 100vh;
  background-color: #f3f4f6;
}

.loading-state,
.error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 50vh;
  gap: 16px;
  color: #6b7280;
}

.spinner {
  border: 4px solid #f3f3f3;
  border-top: 4px solid #3b82f6;
  border-radius: 50%;
  width: 40px;
  height: 40px;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}

.client-info {
  background-color: white;
  padding: 24px;
  border-radius: 12px;
  margin-bottom: 24px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.client-info h1 {
  margin: 0;
  color: #1f2937;
  font-size: 1.5rem;
  margin-bottom: 12px;
}

.client-details {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  color: #4b5563;
}

.client-details p {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.order-cards {
  display: grid;
  gap: 16px;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
}

.order-card {
  background-color: white;
  border-radius: 12px;
  padding: 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.order-id {
  font-weight: 600;
  color: #2563eb;
}

.order-date {
  color: #6b7280;
  font-size: 0.875rem;
}

.car-info h3 {
  margin: 0;
  color: #1f2937;
  font-size: 1.125rem;
}

.vin {
  color: #6b7280;
  font-size: 0.875rem;
  margin: 4px 0;
}

.financial-info {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin: 16px 0;
  padding: 12px;
  background-color: #f3f4f6;
  border-radius: 8px;
}

.price,
.freight {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.label {
  color: #6b7280;
  font-size: 0.75rem;
}

.value {
  color: #1f2937;
  font-weight: 600;
}

.shipping-info {
  margin-top: 16px;
}

.port {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #4b5563;
  margin-bottom: 12px;
}

.dates {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.date-item {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #6b7280;
  font-size: 0.875rem;
}

.date-item.completed {
  color: #059669;
}

.date-item i {
  width: 16px;
}

.no-orders {
  text-align: center;
  padding: 48px;
  color: #6b7280;
}

.no-orders i {
  font-size: 3rem;
  margin-bottom: 16px;
}

@media (max-width: 640px) {
  .public-orders {
    padding: 12px;
  }

  .client-info {
    padding: 16px;
  }

  .order-cards {
    grid-template-columns: 1fr;
  }

  .financial-info {
    grid-template-columns: 1fr;
  }
}
</style>
