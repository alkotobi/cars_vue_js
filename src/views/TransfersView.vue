<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()
const goToDashboard = () => {
  router.push('/')
}
</script>

<template>
  <div class="transfers-page">
    <div class="header">
      <button @click="goToDashboard" class="dashboard-btn">← Return to Dashboard</button>
      <h1>Transfers</h1>
    </div>
  </div>
</template>

<style scoped>
.transfers-page {
  padding: 20px;
}

.header {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-bottom: 30px;
}

.dashboard-btn {
  padding: 8px 16px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: background-color 0.2s;
  display: flex;
  align-items: center;
  gap: 8px;
}

.dashboard-btn:hover {
  background-color: #2563eb;
}
</style>