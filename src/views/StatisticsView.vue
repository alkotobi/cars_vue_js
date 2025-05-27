<script setup>
import { ref, onMounted, computed } from 'vue'
import CarStockStatistics from '../components/car-stock/CarStockStatistics.vue'

const isLoading = ref(false)

// Check if user is admin
const currentUser = computed(() => {
  const userStr = localStorage.getItem('user')
  return userStr ? JSON.parse(userStr) : null
})

const isAdmin = computed(() => {
  return currentUser.value?.role_id === 1
})

// Redirect if not admin
onMounted(() => {
  if (!isAdmin.value) {
    return
  }
})
</script>

<template>
  <div class="statistics-view" :class="{ 'is-loading': isLoading }">
    <div class="header">
      <h1>
        <i class="fas fa-chart-bar"></i>
        Statistics Dashboard
      </h1>
    </div>

    <div class="content">
      <CarStockStatistics />
      <!-- Additional statistics sections can be added here -->
    </div>
  </div>
</template>

<style scoped>
.statistics-view {
  min-height: 100%;
  background-color: #f9fafb;
}

.statistics-view.is-loading {
  opacity: 0.7;
  pointer-events: none;
}

.header {
  display: flex;
  align-items: center;
  gap: 24px;
  margin-bottom: 32px;
}

.header h1 {
  margin: 0;
  font-size: 1.5rem;
  color: #1f2937;
  display: flex;
  align-items: center;
  gap: 12px;
}

.header h1 i {
  color: #3b82f6;
}

.content {
  background-color: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

@media (max-width: 768px) {
  .statistics-view {
    padding: 16px;
  }

  .header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }

  .content {
    padding: 16px;
  }
}
</style>
