<template>
  <div v-if="hasVersionMismatch" class="version-modal-overlay">
    <div class="version-modal">
      <div class="version-modal-header">
        <i class="fas fa-exclamation-triangle"></i>
        <h2>Application Update Required</h2>
      </div>

      <div class="version-modal-content">
        <p>The application has been updated and requires a refresh to work properly.</p>

        <div class="version-info">
          <div class="version-item">
            <span class="version-label">Database Version:</span>
            <span class="version-value">{{ dbVersion }}</span>
          </div>
          <div class="version-item">
            <span class="version-label">Application Version:</span>
            <span class="version-value">{{ currentAppVersion }}</span>
          </div>
        </div>

        <div class="version-warning">
          <i class="fas fa-info-circle"></i>
          <span>Please refresh the page to use the latest version.</span>
        </div>

        <div class="version-actions">
          <button @click="forceRefresh" class="btn-refresh">
            <i class="fas fa-sync-alt"></i>
            Refresh Now
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useVersionCheck } from '@/composables/useVersionCheck'

const { currentAppVersion, dbVersion, isLoading, hasVersionMismatch, checkVersion, forceRefresh } =
  useVersionCheck()

onMounted(async () => {
  await checkVersion()
})
</script>

<style scoped>
.version-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.9);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10000;
}

.version-modal {
  background: white;
  border-radius: 12px;
  padding: 32px;
  max-width: 500px;
  width: 90%;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.5);
  text-align: center;
}

.version-modal-header {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin-bottom: 24px;
  color: #dc2626;
}

.version-modal-header i {
  font-size: 28px;
}

.version-modal-header h2 {
  margin: 0;
  color: #dc2626;
  font-size: 24px;
}

.version-modal-content p {
  margin: 16px 0;
  color: #374151;
  line-height: 1.6;
  font-size: 16px;
}

.version-info {
  background: #f8fafc;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 20px;
  margin: 20px 0;
}

.version-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin: 8px 0;
  padding: 8px 0;
}

.version-label {
  font-weight: 600;
  color: #374151;
}

.version-value {
  font-weight: 700;
  color: #1f2937;
  background: #e5e7eb;
  padding: 4px 12px;
  border-radius: 6px;
  font-family: monospace;
}

.version-warning {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin: 20px 0;
  padding: 16px;
  background: #fef3c7;
  border: 1px solid #f59e0b;
  border-radius: 8px;
  color: #92400e;
}

.version-warning i {
  color: #f59e0b;
  font-size: 18px;
}

.version-actions {
  margin-top: 24px;
}

.btn-refresh {
  padding: 16px 32px;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  font-size: 16px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 12px;
  transition: all 0.2s;
  background: #dc2626;
  color: white;
  box-shadow: 0 4px 6px rgba(220, 38, 38, 0.2);
}

.btn-refresh:hover {
  background: #b91c1c;
  transform: translateY(-1px);
  box-shadow: 0 6px 8px rgba(220, 38, 38, 0.3);
}

.btn-refresh:active {
  transform: translateY(0);
}
</style>
