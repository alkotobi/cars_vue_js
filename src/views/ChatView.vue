<script setup>
import { ref, onMounted } from 'vue'
import ChatSidebar from '../components/chat/ChatSidebar.vue'
import ChatMain from '../components/chat/ChatMain.vue'

const welcomeMessage = ref('')
const selectedGroup = ref(null)

onMounted(() => {
  // Get welcome message from URL parameters
  const urlParams = new URLSearchParams(window.location.search)
  welcomeMessage.value = urlParams.get('welcome') || 'Welcome to Chat!'
})

const handleGroupSelected = (group) => {
  selectedGroup.value = group
}
</script>

<template>
  <div class="chat-layout">
    <ChatSidebar @group-selected="handleGroupSelected" />
    <ChatMain :welcome-message="welcomeMessage" :selected-group="selectedGroup" />
  </div>
</template>

<style scoped>
.chat-layout {
  display: flex;
  height: 100vh;
  background-color: #f8fafc;
}

/* Responsive design */
@media (max-width: 768px) {
  .chat-layout {
    flex-direction: column;
  }
}
</style>
