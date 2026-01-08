<template>
  <div class="upgrades-page">
    <div class="header">
      <button @click="goToDashboard" class="back-btn">
        <i class="fas fa-arrow-left"></i>
        Return to Dashboard
      </button>
      <h1>
        <i class="fas fa-wrench"></i>
        {{ t('upgrades.title') || 'Upgrades Management' }}
      </h1>
    </div>
    <div class="crud-section">
      <UpgradesCrud />
    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import UpgradesCrud from '../components/UpgradesCrud.vue'
import { useEnhancedI18n } from '@/composables/useI18n'

const { t } = useEnhancedI18n()
const router = useRouter()

const goToDashboard = () => {
  router.push('/dashboard')
}
</script>

<style scoped>
.upgrades-page {
  padding: 20px;
  max-width: 1400px;
  margin: 0 auto;
}

.header {
  display: flex;
  align-items: center;
  gap: 24px;
  margin-bottom: 24px;
  background: white;
  padding: 16px;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.header h1 {
  margin: 0;
  font-size: 24px;
  color: #1e293b;
  display: flex;
  align-items: center;
  gap: 12px;
}

.header h1 i {
  color: #f59e0b;
  font-size: 28px;
}

.back-btn {
  padding: 10px 16px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  gap: 8px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}

.back-btn i {
  font-size: 14px;
  transition: transform 0.2s ease;
}

.back-btn:hover {
  background-color: #2563eb;
  transform: translateY(-1px);
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.back-btn:hover i {
  transform: translateX(-2px);
}

.back-btn:active {
  transform: translateY(0);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}

.crud-section {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
</style>

