<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import ChatSidebar from '../components/chat/ChatSidebar.vue'
import ChatMain from '../components/chat/ChatMain.vue'

const welcomeMessage = ref('')
const selectedGroup = ref(null)
const chatSidebarRef = ref(null)
const chatMainRef = ref(null)
const newMessagesCounts = ref({})

// Set up periodic updates for new message counts
let countsUpdateInterval = null

onMounted(() => {
  // Get welcome message from URL parameters
  const urlParams = new URLSearchParams(window.location.search)
  welcomeMessage.value = urlParams.get('welcome') || 'Welcome to Chat!'

  // Set up periodic updates for new message counts
  countsUpdateInterval = setInterval(updateNewMessageCounts, 5000) // Update every 5 seconds
})

onUnmounted(() => {
  // Clean up intervals and resources
  if (chatSidebarRef.value?.cleanup) {
    chatSidebarRef.value.cleanup()
  }
  if (chatMainRef.value?.cleanup) {
    chatMainRef.value.cleanup()
  }
  if (countsUpdateInterval) {
    clearInterval(countsUpdateInterval)
  }
})

const handleGroupSelected = (group) => {
  selectedGroup.value = group
}

const handleMessageSent = (message) => {
  console.log('Message sent:', message)
}

const handleNewMessagesReceived = (groupId) => {
  console.log('New messages received for group:', groupId)
  console.log('chatMainRef.value:', chatMainRef.value)

  // Update new message counts from ChatMessages component
  if (chatMainRef.value?.getAllNewMessagesCounts) {
    const counts = chatMainRef.value.getAllNewMessagesCounts()
    console.log('Retrieved counts from new messages event:', counts)
    newMessagesCounts.value = counts
    console.log('Updated newMessagesCounts from event:', newMessagesCounts.value)
  } else {
    console.log('getAllNewMessagesCounts method not available in event handler')
  }
}

// Function to update new message counts periodically
const updateNewMessageCounts = () => {
  console.log('Updating new message counts...')
  console.log('chatMainRef.value:', chatMainRef.value)

  if (chatMainRef.value?.getAllNewMessagesCounts) {
    const counts = chatMainRef.value.getAllNewMessagesCounts()
    console.log('Retrieved counts:', counts)
    newMessagesCounts.value = counts
    console.log('Updated newMessagesCounts:', newMessagesCounts.value)
  } else {
    console.log('getAllNewMessagesCounts method not available')
  }
}
</script>

<template>
  <div class="chat-layout">
    <ChatSidebar
      ref="chatSidebarRef"
      :new-messages-counts="newMessagesCounts"
      @group-selected="handleGroupSelected"
    />
    <ChatMain
      ref="chatMainRef"
      :welcome-message="welcomeMessage"
      :selected-group="selectedGroup"
      @message-sent="handleMessageSent"
      @new-messages-received="handleNewMessagesReceived"
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
