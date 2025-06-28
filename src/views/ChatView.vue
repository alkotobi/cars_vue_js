<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import ChatSidebar from '../components/chat/ChatSidebar.vue'
import ChatMain from '../components/chat/ChatMain.vue'

const welcomeMessage = ref('')
const selectedGroup = ref(null)
const chatSidebarRef = ref(null)
const chatMainRef = ref(null)

onMounted(() => {
  // Get welcome message from URL parameters
  const urlParams = new URLSearchParams(window.location.search)
  welcomeMessage.value = urlParams.get('welcome') || 'Welcome to Chat!'
})

onUnmounted(() => {
  // Clean up intervals and resources
  if (chatSidebarRef.value?.cleanup) {
    chatSidebarRef.value.cleanup()
  }
  if (chatMainRef.value?.cleanup) {
    chatMainRef.value.cleanup()
  }
})

const handleGroupSelected = (group) => {
  selectedGroup.value = group
}

const handleMessageSent = (message) => {
  console.log('Message sent:', message)
  // Update unread count in sidebar for the group where message was sent
  if (chatSidebarRef.value?.updateUnreadCount) {
    chatSidebarRef.value.updateUnreadCount(message.id_chat_group)
  }
}
</script>

<template>
  <div class="chat-layout">
    <ChatSidebar ref="chatSidebarRef" @group-selected="handleGroupSelected" />
    <ChatMain
      ref="chatMainRef"
      :welcome-message="welcomeMessage"
      :selected-group="selectedGroup"
      @message-sent="handleMessageSent"
    />
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
