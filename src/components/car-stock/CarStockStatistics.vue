<script setup>
import { ref, onMounted, computed } from 'vue'
import { useApi } from '../../composables/useApi'

const { callApi } = useApi()

const stats = ref({
  available: 0,
  soldThisMonth: 0,
  soldThisYear: 0,
  carStats: [],
})

const loading = ref(false)
const error = ref(null)

const fetchStatistics = async () => {
  loading.value = true
  error.value = null

  try {
    const [overallStats, carNameStats] = await Promise.all([
      callApi({
        query: `
          SELECT
            (SELECT COUNT(*) FROM cars_stock WHERE date_sell IS NULL AND id_client IS NULL) as available,
            (SELECT COUNT(*) FROM cars_stock 
             WHERE (date_sell IS NOT NULL OR id_client IS NOT NULL)
             AND MONTH(date_sell) = MONTH(CURRENT_DATE)
             AND YEAR(date_sell) = YEAR(CURRENT_DATE)) as sold_this_month,
            (SELECT COUNT(*) FROM cars_stock 
             WHERE (date_sell IS NOT NULL OR id_client IS NOT NULL)
             AND YEAR(date_sell) = YEAR(CURRENT_DATE)) as sold_this_year
        `,
        params: [],
      }),
      callApi({
        query: `
          SELECT 
            cn.car_name,
            COUNT(CASE WHEN cs.date_sell IS NULL AND cs.id_client IS NULL THEN 1 END) as available,
            COUNT(CASE WHEN cs.date_sell IS NOT NULL OR cs.id_client IS NOT NULL THEN 1 END) as total_sold,
            COUNT(CASE 
              WHEN (cs.date_sell IS NOT NULL OR cs.id_client IS NOT NULL)
              AND MONTH(cs.date_sell) = MONTH(CURRENT_DATE)
              AND YEAR(cs.date_sell) = YEAR(CURRENT_DATE)
              THEN 1 
            END) as sold_this_month
          FROM cars_names cn
          LEFT JOIN buy_details bd ON cn.id = bd.id_car_name
          LEFT JOIN cars_stock cs ON bd.id = cs.id_buy_details
          GROUP BY cn.car_name
          ORDER BY cn.car_name ASC
        `,
        params: [],
      }),
    ])

    if (overallStats.success && carNameStats.success) {
      stats.value = {
        available: overallStats.data[0].available,
        soldThisMonth: overallStats.data[0].sold_this_month,
        soldThisYear: overallStats.data[0].sold_this_year,
        carStats: carNameStats.data,
      }
    } else {
      error.value = overallStats.error || carNameStats.error || 'Failed to fetch statistics'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred'
  } finally {
    loading.value = false
  }
}

// Format numbers with commas
const formatNumber = (num) => {
  return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',')
}

// Computed properties for formatted stats
const formattedStats = computed(() => ({
  available: formatNumber(stats.value.available),
  soldThisMonth: formatNumber(stats.value.soldThisMonth),
  soldThisYear: formatNumber(stats.value.soldThisYear),
}))

onMounted(() => {
  fetchStatistics()
})

// Expose the refresh method
defineExpose({
  refresh: fetchStatistics,
})
</script>

<template>
  <div class="statistics-panel" :class="{ 'is-loading': loading }">
    <!-- Overall Statistics -->
    <div class="stat-cards">
      <!-- Available Cars -->
      <div class="stat-card available">
        <div class="stat-icon">
          <i class="fas fa-warehouse"></i>
        </div>
        <div class="stat-content">
          <h3>Available Cars</h3>
          <div class="stat-value">
            <span v-if="loading">
              <i class="fas fa-spinner fa-spin"></i>
            </span>
            <span v-else>{{ formattedStats.available }}</span>
          </div>
          <div class="stat-label">In Stock</div>
        </div>
      </div>

      <!-- Sold This Month -->
      <div class="stat-card monthly">
        <div class="stat-icon">
          <i class="fas fa-calendar-alt"></i>
        </div>
        <div class="stat-content">
          <h3>Sold This Month</h3>
          <div class="stat-value">
            <span v-if="loading">
              <i class="fas fa-spinner fa-spin"></i>
            </span>
            <span v-else>{{ formattedStats.soldThisMonth }}</span>
          </div>
          <div class="stat-label">Current Month</div>
        </div>
      </div>

      <!-- Sold This Year -->
      <div class="stat-card yearly">
        <div class="stat-icon">
          <i class="fas fa-chart-line"></i>
        </div>
        <div class="stat-content">
          <h3>Sold This Year</h3>
          <div class="stat-value">
            <span v-if="loading">
              <i class="fas fa-spinner fa-spin"></i>
            </span>
            <span v-else>{{ formattedStats.soldThisYear }}</span>
          </div>
          <div class="stat-label">Year to Date</div>
        </div>
      </div>
    </div>

    <!-- Car Model Statistics -->
    <div class="car-model-stats">
      <h3 class="section-title">
        <i class="fas fa-car"></i>
        Statistics by Car Model
      </h3>

      <div class="car-stats-table-wrapper">
        <table class="car-stats-table">
          <thead>
            <tr>
              <th>Car Model</th>
              <th>Available</th>
              <th>Sold This Month</th>
              <th>Total Sold</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td colspan="4" class="loading-row">
                <i class="fas fa-spinner fa-spin"></i>
                Loading car statistics...
              </td>
            </tr>
            <tr v-else v-for="car in stats.carStats" :key="car.car_name">
              <td>{{ car.car_name }}</td>
              <td>{{ formatNumber(car.available) }}</td>
              <td>{{ formatNumber(car.sold_this_month) }}</td>
              <td>{{ formatNumber(car.total_sold) }}</td>
            </tr>
            <tr v-if="!loading && stats.carStats.length === 0">
              <td colspan="4" class="empty-row">No car models found</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div v-if="error" class="error-message">
      <i class="fas fa-exclamation-circle"></i>
      {{ error }}
    </div>
  </div>
</template>

<style scoped>
.statistics-panel {
  margin-bottom: 24px;
  position: relative;
}

.statistics-panel.is-loading {
  opacity: 0.7;
}

.stat-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
  margin-bottom: 20px;
}

.stat-card {
  background: white;
  border-radius: 12px;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  transition: all 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.stat-icon {
  background-color: #f3f4f6;
  border-radius: 50%;
  width: 50px;
  height: 50px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.stat-icon i {
  font-size: 1.5rem;
}

.available .stat-icon {
  background-color: #dcfce7;
}

.available .stat-icon i {
  color: #16a34a;
}

.monthly .stat-icon {
  background-color: #dbeafe;
}

.monthly .stat-icon i {
  color: #2563eb;
}

.yearly .stat-icon {
  background-color: #fef3c7;
}

.yearly .stat-icon i {
  color: #d97706;
}

.stat-content {
  flex: 1;
}

.stat-content h3 {
  font-size: 0.875rem;
  color: #6b7280;
  margin: 0 0 4px 0;
}

.stat-value {
  font-size: 1.5rem;
  font-weight: 600;
  color: #111827;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 0.75rem;
  color: #6b7280;
}

.error-message {
  background-color: #fee2e2;
  color: #dc2626;
  padding: 12px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 12px;
}

.error-message i {
  font-size: 1.1em;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 32px 0 16px;
  color: #1f2937;
  font-size: 1.1rem;
}

.section-title i {
  color: #4b5563;
}

.car-model-stats {
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.car-stats-table-wrapper {
  overflow-x: auto;
}

.car-stats-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  margin-top: 12px;
}

.car-stats-table th,
.car-stats-table td {
  padding: 12px 16px;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

.car-stats-table th {
  background-color: #f8fafc;
  font-weight: 600;
  color: #4b5563;
  white-space: nowrap;
}

.car-stats-table td {
  color: #1f2937;
}

.car-stats-table tbody tr:hover {
  background-color: #f9fafb;
}

.loading-row {
  text-align: center;
  color: #6b7280;
}

.loading-row i {
  margin-right: 8px;
  color: #3b82f6;
}

.empty-row {
  text-align: center;
  color: #6b7280;
  padding: 24px !important;
}

/* Responsive adjustments */
@media (max-width: 768px) {
  .stat-cards {
    grid-template-columns: 1fr;
  }
  .car-stats-table {
    min-width: 600px;
  }
}
</style>
