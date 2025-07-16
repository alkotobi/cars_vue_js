<template>
  <div v-if="isDevelopment && showPanel" class="i18n-debug-panel">
    <div class="debug-header">
      <h3>Missing Translation Keys</h3>
      <button @click="togglePanel" class="close-btn">
        <i class="fas fa-times"></i>
      </button>
    </div>

    <div class="debug-content">
      <div v-if="missingKeys.length === 0" class="no-missing">
        <i class="fas fa-check-circle"></i>
        <span>No missing translation keys!</span>
      </div>

      <div v-else class="missing-keys">
        <div v-for="key in missingKeys" :key="key" class="missing-key">
          <span class="key-name">{{ key }}</span>
          <button @click="copyKey(key)" class="copy-btn" title="Copy key">
            <i class="fas fa-copy"></i>
          </button>
        </div>
      </div>

      <div class="debug-actions">
        <button @click="exportMissingKeys" class="action-btn">
          <i class="fas fa-download"></i>
          Export Missing Keys
        </button>
        <button @click="clearMissingKeys" class="action-btn">
          <i class="fas fa-trash"></i>
          Clear Tracking
        </button>
      </div>
    </div>
  </div>

  <!-- Floating toggle button -->
  <button
    v-if="isDevelopment"
    @click="togglePanel"
    class="debug-toggle-btn"
    :class="{ active: showPanel }"
    title="Toggle I18n Debug Panel"
  >
    <i class="fas fa-language"></i>
    <span v-if="missingKeys.length > 0" class="missing-count">{{ missingKeys.length }}</span>
  </button>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useEnhancedI18n } from '../composables/useI18n'

const { getDevHelper } = useEnhancedI18n()
const devHelper = getDevHelper()

const isDevelopment = process.env.NODE_ENV === 'development'
const showPanel = ref(false)
const missingKeys = ref([])

const togglePanel = () => {
  showPanel.value = !showPanel.value
  if (showPanel.value) {
    updateMissingKeys()
  }
}

const updateMissingKeys = () => {
  missingKeys.value = devHelper.getMissingKeys()
}

const exportMissingKeys = () => {
  const missingByLocale = devHelper.exportMissingKeys()
  console.log('Missing keys by locale:', missingByLocale)

  // Create downloadable JSON
  const dataStr = JSON.stringify(missingByLocale, null, 2)
  const dataBlob = new Blob([dataStr], { type: 'application/json' })
  const url = URL.createObjectURL(dataBlob)
  const link = document.createElement('a')
  link.href = url
  link.download = 'missing-translation-keys.json'
  link.click()
  URL.revokeObjectURL(url)
}

const clearMissingKeys = () => {
  devHelper.clearMissingKeys()
  updateMissingKeys()
}

const copyKey = (key) => {
  navigator.clipboard.writeText(key).then(() => {
    // Show brief feedback
    const btn = event.target.closest('.copy-btn')
    const icon = btn.querySelector('i')
    const originalClass = icon.className
    icon.className = 'fas fa-check'
    setTimeout(() => {
      icon.className = originalClass
    }, 1000)
  })
}

onMounted(() => {
  // Update missing keys every 5 seconds in development
  if (isDevelopment) {
    setInterval(updateMissingKeys, 5000)
  }
})
</script>

<style scoped>
.i18n-debug-panel {
  position: fixed;
  top: 20px;
  right: 20px;
  width: 400px;
  max-height: 500px;
  background: white;
  border: 2px solid #e74c3c;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  z-index: 10000;
  font-family: 'Courier New', monospace;
  font-size: 12px;
}

.debug-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: #e74c3c;
  color: white;
  border-radius: 6px 6px 0 0;
}

.debug-header h3 {
  margin: 0;
  font-size: 14px;
  font-weight: bold;
}

.close-btn {
  background: none;
  border: none;
  color: white;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.2);
}

.debug-content {
  padding: 16px;
  max-height: 400px;
  overflow-y: auto;
}

.no-missing {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #27ae60;
  font-weight: bold;
}

.missing-keys {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.missing-key {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px;
  background: #f8f9fa;
  border-radius: 4px;
  border-left: 3px solid #e74c3c;
}

.key-name {
  font-family: 'Courier New', monospace;
  font-size: 11px;
  color: #2c3e50;
  word-break: break-all;
}

.copy-btn {
  background: none;
  border: none;
  color: #7f8c8d;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  font-size: 10px;
}

.copy-btn:hover {
  color: #2c3e50;
  background: #ecf0f1;
}

.debug-actions {
  display: flex;
  gap: 8px;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #ecf0f1;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: #3498db;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 11px;
  font-weight: bold;
}

.action-btn:hover {
  background: #2980b9;
}

.debug-toggle-btn {
  position: fixed;
  bottom: 20px;
  right: 20px;
  width: 50px;
  height: 50px;
  background: #e74c3c;
  color: white;
  border: none;
  border-radius: 50%;
  cursor: pointer;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  position: relative;
  z-index: 9999;
}

.debug-toggle-btn:hover {
  background: #c0392b;
  transform: scale(1.1);
}

.debug-toggle-btn.active {
  background: #27ae60;
}

.missing-count {
  position: absolute;
  top: -5px;
  right: -5px;
  background: #f39c12;
  color: white;
  border-radius: 50%;
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 10px;
  font-weight: bold;
}
</style>
