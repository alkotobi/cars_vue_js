<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useApi } from '../composables/useApi'

const router = useRouter()
const { callApi, error } = useApi()
const user = ref(null)
const transfers = ref([])

const canSendTransfer = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some(p => p.permission_name === 'is_exchange_sender')
})

const canReceiveTransfer = computed(() => {
  if (!user.value) return false
  if (user.value.role_id === 1) return true
  return user.value.permissions?.some(p => p.permission_name === 'is_exchange_receiver')
})

// Statistics computations
const statistics = computed(() => {
  if (!transfers.value.length) return {
    totalSent: 0,
    totalReceived: 0,
    totalPending: 0,
    totalPendingDA: 0,
    totalPendingUSD: 0
  }

  const stats = transfers.value.reduce((acc, transfer) => {
    // Count total sent
    acc.totalSent++

    if (transfer.date_receive) {
      // Count received transfers
      acc.totalReceived++
    } else {
      // Count pending transfers
      acc.totalPending++
      // Sum pending amounts
      acc.totalPendingDA += parseFloat(transfer.amount_sending_da)
      acc.totalPendingUSD += parseFloat(transfer.amount_sending_da) / parseFloat(transfer.rate)
    }

    return acc
  }, {
    totalSent: 0,
    totalReceived: 0,
    totalPending: 0,
    totalPendingDA: 0,
    totalPendingUSD: 0
  })

  // Round USD to 2 decimal places
  stats.totalPendingUSD = Math.round(stats.totalPendingUSD * 100) / 100

  return stats
})

const fetchTransfers = async () => {
  const result = await callApi({
    query: `
      SELECT t.* 
      FROM transfers t
      WHERE t.id_user_do_transfer = ?
      ORDER BY t.date_do_transfer DESC
    `,
    params: [user.value.id]
  })
  if (result.success) {
    transfers.value = result.data
  }
}

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
    fetchTransfers()
  }
})
</script>

<template>
  <div class="transfers-page">
    <div class="sidebar">
      <router-link to="/dashboard" class="dashboard-btn">← Return to Dashboard</router-link>
      <div class="nav-buttons">
        <router-link 
          v-if="canSendTransfer" 
          to="/send" 
          class="nav-btn"
        >
          Send Transfer
        </router-link>
        <router-link 
          v-if="canReceiveTransfer" 
          to="/receive" 
          class="nav-btn"
        >
          Receive Transfer
        </router-link>
      </div>
    </div>
    <div class="main-content">
      <!-- Statistics Section -->
      <div class="statistics-container">
        <h2>Transfer Statistics</h2>
        <div class="statistics-grid">
          <div class="stat-card">
            <span class="stat-label">Total Sent</span>
            <span class="stat-value">{{ statistics.totalSent }}</span>
          </div>
          <div class="stat-card">
            <span class="stat-label">Total Received</span>
            <span class="stat-value">{{ statistics.totalReceived }}</span>
          </div>
          <div class="stat-card">
            <span class="stat-label">Pending Transfers</span>
            <span class="stat-value">{{ statistics.totalPending }}</span>
          </div>
          <div class="stat-card">
            <span class="stat-label">Total Pending (DA)</span>
            <span class="stat-value">{{ statistics.totalPendingDA.toLocaleString() }} DA</span>
          </div>
          <div class="stat-card">
            <span class="stat-label">Total Pending (USD)</span>
            <span class="stat-value">${{ statistics.totalPendingUSD.toLocaleString() }}</span>
          </div>
        </div>
      </div>
      <router-view></router-view>
    </div>
  </div>
</template>

<style scoped>
.transfers-page {
  display: flex;
  min-height: 100vh;
}

.sidebar {
  width: 250px;
  background-color: #f3f4f6;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.nav-buttons {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.nav-btn {
  padding: 10px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  text-decoration: none;
  text-align: center;
}

.nav-btn:hover {
  background-color: #2563eb;
}

.main-content {
  flex: 1;
  padding: 20px;
}

.dashboard-btn {
  padding: 8px 16px;
  background-color: #4b5563;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  text-decoration: none;
  display: flex;
  align-items: center;
  gap: 8px;
}

.dashboard-btn:hover {
  background-color: #374151;
}

.statistics-container {
  margin-bottom: 30px;
}

.statistics-container h2 {
  color: #1f2937;
  font-size: 1.5em;
  margin-bottom: 20px;
}

.statistics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
}

.stat-card {
  background-color: #ffffff;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}

.stat-label {
  color: #6b7280;
  font-size: 0.9em;
  margin-bottom: 8px;
}

.stat-value {
  font-size: 1.5em;
  font-weight: 600;
  color: #1f2937;
}
</style>