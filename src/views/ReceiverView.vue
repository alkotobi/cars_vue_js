<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()
const goToTransfers = () => {
  router.push('/transfers')
}
</script>

<template>
  <div class="receiver-view">
    <h1>Receive Transfer</h1>
    <button @click="goToTransfers" class="back-btn">← Return to Transfers</button>
  </div>
</template>

<style scoped>
.receiver-view {
  padding: 20px;
}

.back-btn {
  padding: 8px 16px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  margin-top: 20px;
}

.back-btn:hover {
  background-color: #2563eb;
}
</style>