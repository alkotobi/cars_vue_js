<template>
  <div class="loading-view">
    <div class="header">
      <h2>
        <i class="fas fa-truck-loading"></i>
        {{ t('loading.loading_management') }}
      </h2>
    </div>

    <div class="content">
      <LoadingTable />
    </div>
  </div>
</template>

<script setup>
import { useI18n } from 'vue-i18n'
import LoadingTable from '../components/loading/LoadingTable.vue'

const { t, locale, messages } = useI18n()

// Debug: Log locale and translation attempt
console.log('LoadingView - Current locale:', locale.value)
console.log('LoadingView - Available messages:', Object.keys(messages.value || {}))
console.log('LoadingView - English messages:', Object.keys(messages.value?.en || {}))
console.log('LoadingView - Loading section:', Object.keys(messages.value?.en?.loading || {}))
console.log(
  'LoadingView - Translation attempt for loading.loading_management:',
  t('loading.loading_management'),
)
</script>

<style scoped>
.loading-view {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header h2 {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0;
  color: #2c3e50;
}

.header h2 i {
  color: #3498db;
}

.content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
</style>
