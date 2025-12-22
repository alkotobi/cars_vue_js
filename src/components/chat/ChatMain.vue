<script setup>
import { ref, computed } from 'vue'
import ChatMessages from './ChatMessages.vue'
import { useEnhancedI18n } from '../../composables/useI18n'

const { t } = useEnhancedI18n()

const props = defineProps({
  welcomeMessage: {
    type: String,
    default: null,
  },
  selectedGroup: {
    type: Object,
    default: null,
  },
  onForceUpdateCounts: {
    type: Function,
    default: null,
  },
  // Client mode props (optional)
  clientId: {
    type: Number,
    default: null,
  },
  clientName: {
    type: String,
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
      <h2>{{ welcomeMessage || t('chat.welcomeToChat') }}</h2>
      <p>{{ t('chat.selectGroupToStart') }}</p>
    </div>

    <ChatMessages
      v-else
      ref="chatMessagesRef"
      :group-id="selectedGroup.id"
      :group-name="selectedGroup.name"
      :client-id="clientId"
      :client-name="clientName"
      @message-sent="handleMessageSent"
      @new-messages-received="handleNewMessagesReceived"
      @reset-triggered="handleReset"
    />
  </div>
</template>

<style scoped>
.chat-main {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
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
</style>
