<script setup>
import { defineProps, defineEmits, ref } from 'vue'
import ChatMessages from './ChatMessages.vue'

const props = defineProps({
  welcomeMessage: {
    type: String,
    default: 'Welcome to Chat!',
  },
  selectedGroup: {
    type: Object,
    default: null,
  },
  onForceUpdateCounts: {
    type: Function,
    default: null,
  },
})

const emit = defineEmits(['message-sent', 'new-messages-received'])

const chatMessagesRef = ref(null)

const handleMessageSent = (message) => {
  emit('message-sent', message)
}

const handleNewMessagesReceived = (groupId) => {
  emit('new-messages-received', groupId)
}

// Expose methods from ChatMessages component
const getAllNewMessagesCounts = () => {
  return chatMessagesRef.value?.getAllNewMessagesCounts?.() || {}
}

const getNewMessagesCount = (groupId) => {
  return chatMessagesRef.value?.getNewMessagesCount?.(groupId) || 0
}

const fetchAllGroupsMessages = async () => {
  return await chatMessagesRef.value?.fetchAllGroupsMessages?.()
}

// Function to handle reset and force update counts
const handleReset = async (groupId) => {
  console.log('ChatMain: Reset triggered for group', groupId)
  // Call the parent's force update function
  if (props.onForceUpdateCounts) {
    await props.onForceUpdateCounts(groupId)
  }
}

const focusMessageInput = () => {
  console.log('ChatMain: focusMessageInput called')
  if (chatMessagesRef.value?.focusMessageInput) {
    chatMessagesRef.value.focusMessageInput()
  } else {
    console.log('focusMessageInput method not available in ChatMessages')
  }
}

const scrollToBottom = async () => {
  console.log('ChatMain: scrollToBottom called')
  if (chatMessagesRef.value?.scrollToBottom) {
    await chatMessagesRef.value.scrollToBottom()
  } else {
    console.log('scrollToBottom method not available in ChatMessages')
  }
}

defineExpose({
  getAllNewMessagesCounts,
  getNewMessagesCount,
  fetchAllGroupsMessages,
  focusMessageInput,
  scrollToBottom,
})
</script>

<template>
  <div class="chat-main">
    <div v-if="!selectedGroup" class="select-group-message">
      <i class="fas fa-hand-pointer"></i>
      <h2>{{ welcomeMessage }}</h2>
      <p>Select a chat group from the sidebar to start chatting</p>
    </div>

    <ChatMessages
      v-else
      ref="chatMessagesRef"
      :group-id="selectedGroup.id"
      :group-name="selectedGroup.name"
      @message-sent="handleMessageSent"
      @new-messages-received="handleNewMessagesReceived"
      @reset-triggered="handleReset"
    />
  </div>
</template>

<style scoped>
.chat-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  height: 100%;
}

.select-group-message {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #64748b;
  text-align: center;
  padding: 40px;
}

.select-group-message i {
  font-size: 4rem;
  color: #06b6d4;
  margin-bottom: 24px;
}

.select-group-message h2 {
  color: #06b6d4;
  margin: 0 0 16px 0;
  font-size: 1.8rem;
}

.select-group-message p {
  margin: 0;
  font-size: 1.1rem;
  max-width: 400px;
  line-height: 1.5;
}

/* Mobile responsive styles */
@media (max-width: 768px) {
  .select-group-message {
    padding: 20px;
  }

  .select-group-message i {
    font-size: 3rem;
    margin-bottom: 16px;
  }

  .select-group-message h2 {
    font-size: 1.5rem;
    margin: 0 0 12px 0;
  }

  .select-group-message p {
    font-size: 1rem;
    max-width: 300px;
  }
}

/* Landscape orientation adjustments */
@media (max-width: 768px) and (orientation: landscape) {
  .select-group-message {
    padding: 16px;
  }

  .select-group-message i {
    font-size: 2.5rem;
    margin-bottom: 12px;
  }

  .select-group-message h2 {
    font-size: 1.3rem;
    margin: 0 0 8px 0;
  }

  .select-group-message p {
    font-size: 0.9rem;
    max-width: 250px;
  }
}
</style>
