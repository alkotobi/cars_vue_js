<script setup>
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

const router = useRouter()
const { t } = useI18n()

const logout = () => {
  // Clear all localStorage data
  localStorage.clear()

  // Dispatch custom event for header to update
  window.dispatchEvent(new CustomEvent('userLogout'))

  router.push('/login')
}
</script>

<template>
  <button @click="logout" class="logout-btn">{{ t('auth.logout') }}</button>
</template>

<style scoped>
.logout-btn {
  padding: 8px 16px;
  background-color: #f44336;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.logout-btn:hover {
  background-color: #d32f2f;
}
</style>
