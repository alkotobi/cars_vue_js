<script setup>
import { defineProps, defineEmits } from 'vue'
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
})

const emit = defineEmits(['message-sent'])

const handleMessageSent = (message) => {
  emit('message-sent', message)
}
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
      :group-id="selectedGroup.id"
      :group-name="selectedGroup.name"
      @message-sent="handleMessageSent"
    />
  </div>
</template>

<style scoped>
.chat-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  height: 100vh;
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
