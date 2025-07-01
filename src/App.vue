<script setup>
import { RouterView } from 'vue-router'
import AppHeader from './components/AppHeader.vue'
import { onMounted } from 'vue'
import { useApi } from './composables/useApi'

const { handleCookieVerification } = useApi()

onMounted(async () => {
  // Handle cookie verification on app start
  try {
    await handleCookieVerification()
  } catch (error) {
    console.error('Cookie verification failed:', error)
  }
})
</script>

<template>
  <div id="app">
    <AppHeader />
    <main class="app-main">
  <RouterView />
    </main>
  </div>
</template>

<style>
#app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.app-main {
  flex: 1;
  background-color: #f8fafc;
}

@media print {
  .app-header,
  header,
  .app-header * {
    display: none !important;
    visibility: hidden !important;
    height: 0 !important;
    overflow: hidden !important;
    margin: 0 !important;
    padding: 0 !important;
  }
}
</style>
