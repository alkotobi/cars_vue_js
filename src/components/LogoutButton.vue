<script setup>
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useApi } from '../composables/useApi'

const router = useRouter()
const { t } = useI18n()
const { callApi } = useApi()

const logout = () => {
  // The token is a server-side credential now, so tell the server to drop it
  // before the copy in localStorage goes away. Read it first, and do not await
  // the call: an unreachable API must not stop someone from signing out.
  const userStr = localStorage.getItem('user')
  const token = userStr ? JSON.parse(userStr)?.token : null
  if (token) {
    callApi({ action: 'logout', token }).catch(() => {})
  }

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
