<script setup>
import { ref, onMounted, watch, nextTick } from 'vue'
import { useApi } from '../../composables/useApi'

const props = defineProps({
  groupId: {
    type: Number,
    required: true,
  },
  groupName: {
    type: String,
    required: true,
  },
})

const emit = defineEmits(['message-sent', 'new-messages-received', 'reset-triggered'])

const { callApi, uploadFile, getFileUrl, loading, error } = useApi()
const messages = ref([]) // Current group's messages
const messagesByGroup = ref({}) // Store messages for each group: { groupId: [messages] }
const newMessagesCountByGroup = ref({}) // Track new messages count per group: { groupId: count }
const newMessage = ref('')
const currentUser = ref(null)
const isSending = ref(false)
const messagesContainer = ref(null)
const messageInputRef = ref(null)
const fileInputRef = ref(null)

// File upload state
const isUploading = ref(false)
const uploadProgress = ref(0)
const selectedFile = ref(null)
const fileError = ref('')

// File size limit (10MB)
const MAX_FILE_SIZE = 10 * 1024 * 1024 // 10MB in bytes

// Debug section state
const showDebugSection = ref(false)
const debugInfo = ref('')
const currentNewMessagesCount = ref(0) // Track current newMessagesCount value

// Emoji picker state
const showEmojiPicker = ref(false)

// Common emojis for the picker
const emojis = [
  '😀',
  '😃',
  '😄',
  '😁',
  '😆',
  '😅',
  '😂',
  '🤣',
  '😊',
  '😇',
  '🙂',
  '🙃',
  '😉',
  '😌',
  '😍',
  '🥰',
  '😘',
  '😗',
  '😙',
  '😚',
  '😋',
  '😛',
  '😝',
  '😜',
  '🤪',
  '🤨',
  '🧐',
  '🤓',
  '😎',
  '🤩',
  '🥳',
  '😏',
  '😒',
  '😞',
  '😔',
  '😟',
  '😕',
  '🙁',
  '☹️',
  '😣',
  '😖',
  '😫',
  '😩',
  '🥺',
  '😢',
  '😭',
  '😤',
  '😠',
  '😡',
  '🤬',
  '🤯',
  '😳',
  '🥵',
  '🥶',
  '😱',
  '😨',
  '😰',
  '😥',
  '😓',
  '🤗',
  '🤔',
  '🤭',
  '🤫',
  '🤥',
  '😶',
  '😐',
  '😑',
  '😯',
  '😦',
  '😧',
  '😮',
  '😲',
  '🥱',
  '😴',
  '🤤',
  '😪',
  '😵',
  '🤐',
  '🥴',
  '🤢',
  '🤮',
  '🤧',
  '😷',
  '🤒',
  '🤕',
  '🤑',
  '🤠',
  '💩',
  '👻',
  '💀',
  '☠️',
  '👽',
  '👾',
  '🤖',
  '😺',
  '😸',
  '😹',
  '😻',
  '😼',
  '😽',
  '🙀',
  '😿',
  '😾',
  '🙈',
  '🙉',
  '🙊',
  '👶',
  '👧',
  '🧒',
  '👦',
  '👩',
  '🧑',
  '👨',
  '👵',
  '🧓',
  '👴',
  '👮‍♀️',
  '👮',
  '👮‍♂️',
  '🕵️‍♀️',
  '🕵️',
  '🕵️‍♂️',
  '💂‍♀️',
  '💂',
  '💂‍♂️',
  '👷‍♀️',
  '👷',
  '👷‍♂️',
  '🤴',
  '👸',
  '👳‍♀️',
  '👳',
  '👳‍♂️',
  '👲',
  '🧕',
  '🤵‍♀️',
  '🤵',
  '🤵‍♂️',
  '👰‍♀️',
  '👰',
  '👰‍♂️',
  '🤰',
  '🤱',
  '👼',
  '🎅',
  '🤶',
  '🧙‍♀️',
  '🧙',
  '🧙‍♂️',
  '🧝‍♀️',
  '🧝',
  '🧝‍♂️',
  '🧛‍♀️',
  '🧛',
  '🧛‍♂️',
  '🧟‍♀️',
  '🧟',
  '🧟‍♂️',
  '🧞‍♀️',
  '🧞',
  '🧞‍♂️',
  '🧜‍♀️',
  '🧜',
  '🧜‍♂️',
  '🧚‍♀️',
  '🧚',
  '🧚‍♂️',
  '👼',
  '🤰',
  '🤱',
  '❤️',
  '🧡',
  '💛',
  '💚',
  '💙',
  '💜',
  '🖤',
  '🤍',
  '🤎',
  '💔',
  '❣️',
  '💕',
  '💞',
  '💓',
  '💗',
  '💖',
  '💘',
  '💝',
  '💟',
  '☮️',
  '✝️',
  '☪️',
  '🕉️',
  '☸️',
  '✡️',
  '🔯',
  '🕎',
  '☯️',
  '☦️',
  '🛐',
  '⛎',
  '♈',
  '♉',
  '♊',
  '♋',
  '♌',
  '♍',
  '♎',
  '♏',
  '♐',
  '♑',
  '♒',
  '♓',
  '🆔',
  '⚛️',
  '🉑',
  '☢️',
  '☣️',
  '📴',
  '📳',
  '🈶',
  '🈚',
  '🈸',
  '🈺',
  '🈷️',
  '✴️',
  '🆚',
  '💮',
  '🉐',
  '㊙️',
  '㊗️',
  '🈴',
  '🈵',
  '🈹',
  '🈲',
  '🅰️',
  '🅱️',
  '🆎',
  '🆑',
  '🅾️',
  '🆘',
  '❌',
  '⭕',
  '🛑',
  '⛔',
  '📛',
  '🚫',
  '💯',
  '💢',
  '♨️',
  '🚷',
  '🚯',
  '🚳',
  '🚱',
  '🔞',
  '📵',
  '🚭',
  '❗',
  '❕',
  '❓',
  '❔',
  '‼️',
  '⁉️',
  '🔅',
  '🔆',
  '〽️',
  '⚠️',
  '🚸',
  '🔱',
  '⚜️',
  '🔰',
  '♻️',
  '✅',
  '🈯',
  '💹',
  '❇️',
  '✳️',
  '❎',
  '🌐',
  '💠',
  'Ⓜ️',
  '🌀',
  '💤',
  '🏧',
  '🚾',
  '♿',
  '🅿️',
  '🛗',
  '🛂',
  '🛃',
  '🛄',
  '🛅',
  '🚹',
  '🚺',
  '🚼',
  '🚻',
  '🚮',
  '🎦',
  '📶',
  '🈁',
  '🔣',
  'ℹ️',
  '🔤',
  '🔡',
  '🔠',
  '🆖',
  '🆗',
  '🆙',
  '🆒',
  '🆕',
  '🆓',
  '0️⃣',
  '1️⃣',
  '2️⃣',
  '3️⃣',
  '4️⃣',
  '5️⃣',
  '6️⃣',
  '7️⃣',
  '8️⃣',
  '9️⃣',
  '🔟',
  '🔢',
  '#️⃣',
  '*️⃣',
  '⏏️',
  '▶️',
  '⏸️',
  '⏯️',
  '⏹️',
  '⏺️',
  '⏭️',
  '⏮️',
  '⏩',
  '⏪',
  '⏫',
  '⏬',
  '◀️',
  '🔼',
  '🔽',
  '➡️',
  '⬅️',
  '⬆️',
  '⬇️',
  '↗️',
  '↘️',
  '↙️',
  '↖️',
  '↕️',
  '↔️',
  '↪️',
  '↩️',
  '⤴️',
  '⤵️',
  '🔀',
  '🔁',
  '🔂',
  '🔄',
  '🔃',
  '🎵',
  '🎶',
  '➕',
  '➖',
  '➗',
  '✖️',
  '♾️',
  '💲',
  '💱',
  '™️',
  '©️',
  '®️',
  '👁️‍🗨️',
  '🔚',
  '🔙',
  '🔛',
  '🔝',
  '🔜',
  '〰️',
  '➰',
  '➿',
  '✔️',
  '☑️',
  '🔘',
  '🔴',
  '🟠',
  '🟡',
  '🟢',
  '🔵',
  '🟣',
  '⚫',
  '⚪',
  '🟤',
  '🔺',
  '🔻',
  '🔸',
  '🔹',
  '🔶',
  '🔷',
  '🔳',
  '🔲',
  '▪️',
  '▫️',
  '◾',
  '◽',
  '◼️',
  '◻️',
  '🟥',
  '🟧',
  '🟨',
  '🟩',
  '🟦',
  '🟪',
  '⬛',
  '⬜',
  '🟫',
  '🔈',
  '🔇',
  '🔉',
  '🔊',
  '🔔',
  '🔕',
  '📣',
  '📢',
  '💬',
  '💭',
  '🗯️',
  '♠️',
  '♣️',
  '♥️',
  '♦️',
  '🃏',
  '🎴',
  '🀄',
  '🕐',
  '🕑',
  '🕒',
  '🕓',
  '🕔',
  '🕕',
  '🕖',
  '🕗',
  '🕘',
  '🕙',
  '🕚',
  '🕛',
  '🕜',
  '🕝',
  '🕞',
  '🕟',
  '🕠',
  '🕡',
  '🕢',
  '🕣',
  '🕤',
  '🕥',
  '🕦',
  '🕧',
]

const getCurrentUser = () => {
  try {
    const userStr = localStorage.getItem('user')
    if (userStr) {
      currentUser.value = JSON.parse(userStr)
      console.log('Current user loaded for chat:', currentUser.value)
    }
  } catch (err) {
    console.error('Error getting current user:', err)
  }
}

// Watch for changes in groupId to switch message lists
watch(
  [() => props.groupId],
  async ([groupId]) => {
    console.log('ChatMessages watch triggered - groupId:', groupId)
    if (groupId && groupId > 0) {
      console.log('Group changed, switching to cached messages or fetching if needed...')
      await getCurrentUser()
      await switchToGroup(groupId)
    } else if (groupId === 0) {
      console.log('Initialization mode - skipping group-specific operations')
      await getCurrentUser()
    } else {
      messages.value = []
    }
  },
  { immediate: true },
)

const switchToGroup = async (groupId) => {
  // Check if we already have messages for this group
  if (messagesByGroup.value[groupId]) {
    console.log(`Using cached messages for group ${groupId}`)
    messages.value = messagesByGroup.value[groupId]
    // Reset new message count when switching to this group (user clicked on group)
    await resetNewMessageCount(groupId)
    await scrollToBottom()
  } else {
    console.log(`No cached messages for group ${groupId}, fetching all messages...`)
    await fetchAllMessages()
  }

  // Focus the message input after switching groups
  await nextTick()
  if (messageInputRef.value) {
    messageInputRef.value.focus()
  }
}

const fetchAllMessages = async () => {
  try {
    console.log('Fetching all messages for group:', props.groupId)

    const result = await callApi({
      query: `
        SELECT 
          cm.id,
          cm.id_chat_group,
          cm.message_from_user_id,
          cm.chat_replay_to_message_id,
          cm.message,
          cm.time,
          u.username as sender_username,
          r.role_name as sender_role
        FROM chat_messages cm
        LEFT JOIN users u ON cm.message_from_user_id = u.id
        LEFT JOIN roles r ON u.role_id = r.id
        WHERE cm.id_chat_group = ?
        ORDER BY cm.id ASC
      `,
      params: [props.groupId],
      requiresAuth: true,
    })

    console.log('All messages result:', result)

    if (result.success && result.data) {
      // Store messages in the group-specific cache
      messagesByGroup.value[props.groupId] = result.data
      // Set current messages
      messages.value = result.data
      // Initialize new messages count to 0 for this group
      await resetNewMessageCount(props.groupId)
      console.log('All messages loaded and cached:', messages.value.length)

      // Scroll to bottom on initial load
      await scrollToBottom()

      // Focus the message input after loading messages
      await nextTick()
      if (messageInputRef.value) {
        messageInputRef.value.focus()
      }
    } else {
      console.log('No messages found or API failed:', result)
      messagesByGroup.value[props.groupId] = []
      messages.value = []
      await resetNewMessageCount(props.groupId)
    }
  } catch (err) {
    console.error('Error fetching all messages:', err)
    messagesByGroup.value[props.groupId] = []
    messages.value = []
  }
}

const fetchMessages = async (groupId = props.groupId) => {
  try {
    console.log('Fetching new messages for group:', groupId)

    // Get the highest message ID from current messages for this group
    const currentMessages = messagesByGroup.value[groupId] || []
    const currentHighestId =
      currentMessages.length > 0 ? Math.max(...currentMessages.map((msg) => msg.id)) : 0

    console.log('Current highest message ID:', currentHighestId)

    const result = await callApi({
      query: `
        SELECT 
          cm.id,
          cm.id_chat_group,
          cm.message_from_user_id,
          cm.chat_replay_to_message_id,
          cm.message,
          cm.time,
          u.username as sender_username,
          r.role_name as sender_role
        FROM chat_messages cm
        LEFT JOIN users u ON cm.message_from_user_id = u.id
        LEFT JOIN roles r ON u.role_id = r.id
        WHERE cm.id_chat_group = ?
          AND cm.id > ?
        ORDER BY cm.id ASC
      `,
      params: [groupId, currentHighestId],
      requiresAuth: true,
    })

    console.log('New messages result:', result)

    if (result.success && result.data) {
      const newMessagesCount = result.data.length
      // Update debug variable only if this is the current group
      if (groupId === props.groupId) {
        currentNewMessagesCount.value = newMessagesCount
      }
      console.log(`Found ${newMessagesCount} new messages`)

      if (newMessagesCount > 0) {
        // Add new messages to the group cache
        if (messagesByGroup.value[groupId]) {
          messagesByGroup.value[groupId].push(...result.data)
        } else {
          messagesByGroup.value[groupId] = [...result.data]
        }

        // Update current messages if this is the active group
        if (groupId === props.groupId) {
          messages.value = [...messagesByGroup.value[groupId]]
        }

        // Increment new messages count for this group
        if (newMessagesCountByGroup.value[groupId] !== undefined) {
          newMessagesCountByGroup.value[groupId] += newMessagesCount
        } else {
          newMessagesCountByGroup.value[groupId] = newMessagesCount
        }

        console.log(
          'New messages added to list. Total messages:',
          messagesByGroup.value[groupId]?.length || 0,
        )
        console.log(
          `New messages count for group ${groupId}:`,
          newMessagesCountByGroup.value[groupId],
        )
        console.log('All new message counts:', newMessagesCountByGroup.value)

        // Update debug info if debug section is open and this is the current group
        if (showDebugSection.value && groupId === props.groupId) {
          updateDebugInfo()
        }

        // Notify parent about new messages only if this is the current group
        if (groupId === props.groupId) {
          console.log(`Emitting new-messages-received for group ${groupId}`)
          emit('new-messages-received', groupId)

          // Scroll to bottom when new messages are added
          await scrollToBottom()
        }
      } else {
        // Update debug info even when no new messages (only for current group)
        if (showDebugSection.value && groupId === props.groupId) {
          updateDebugInfo()
        }
      }
    } else {
      console.log('No new messages found or API failed:', result)
      if (groupId === props.groupId) {
        currentNewMessagesCount.value = 0
        if (showDebugSection.value) {
          updateDebugInfo()
        }
      }
    }
  } catch (err) {
    console.error('Error fetching new messages:', err)
  }
}

const sendMessage = async () => {
  if (!newMessage.value.trim() || !currentUser.value) {
    return
  }

  try {
    isSending.value = true
    console.log('Sending message:', newMessage.value)

    const result = await callApi({
      query: `
        INSERT INTO chat_messages (id_chat_group, message_from_user_id, message, time) 
        VALUES (?, ?, ?, UTC_TIMESTAMP())
      `,
      params: [props.groupId, currentUser.value.id, newMessage.value.trim()],
      requiresAuth: true,
    })

    console.log('Send message result:', result)

    if (result.success) {
      // Add the new message to both current list and group cache
      const newMsg = {
        id: result.lastInsertId,
        id_chat_group: props.groupId,
        message_from_user_id: currentUser.value.id,
        message: newMessage.value.trim(),
        time: new Date().toISOString(), // Use UTC time
        sender_username: currentUser.value.username,
        sender_role: currentUser.value.role_name,
      }

      messages.value.push(newMsg)
      messagesByGroup.value[props.groupId] = [...messages.value]
      newMessage.value = ''

      // Scroll to bottom after message is added
      await nextTick()
      await scrollToBottom()

      // Emit event to parent
      emit('message-sent', newMsg)
    } else {
      alert('Failed to send message')
    }
  } catch (err) {
    console.error('Error sending message:', err)
    alert('Error sending message')
  } finally {
    isSending.value = false

    // Focus the message input after sending is complete and textarea is enabled
    setTimeout(() => {
      requestAnimationFrame(() => {
        if (messageInputRef.value) {
          messageInputRef.value.focus()
          console.log('Message input focused after sending')
        } else {
          console.log('Message input ref not available for focus')
        }
      })
    }, 100)
  }
}

const scrollToBottom = async () => {
  if (messagesContainer.value) {
    await nextTick()
    // Use multiple timeouts to ensure DOM is fully updated
    setTimeout(() => {
      if (messagesContainer.value) {
        messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
      }
    }, 50)

    // Double-check after a longer delay
    setTimeout(() => {
      if (messagesContainer.value) {
        messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
      }
    }, 200)
  }
}

const formatTime = (timeString) => {
  const date = new Date(timeString)
  const userTimezone = Intl.DateTimeFormat().resolvedOptions().timeZone
  const timeZoneAbbr =
    new Intl.DateTimeFormat('en', { timeZoneName: 'short' })
      .formatToParts(date)
      .find((part) => part.type === 'timeZoneName')?.value || ''

  return date.toLocaleTimeString([], {
    hour: '2-digit',
    minute: '2-digit',
    timeZoneName: 'short',
  })
}

const formatDate = (timeString) => {
  const date = new Date(timeString)
  const today = new Date()
  const yesterday = new Date(today)
  yesterday.setDate(yesterday.getDate() - 1)

  if (date.toDateString() === today.toDateString()) {
    return 'Today'
  } else if (date.toDateString() === yesterday.toDateString()) {
    return 'Yesterday'
  } else {
    return date.toLocaleDateString()
  }
}

const getRelativeTime = (timeString) => {
  const date = new Date(timeString)
  const now = new Date()
  const diffInMinutes = Math.floor((now - date) / (1000 * 60))

  if (diffInMinutes < 1) return 'Just now'
  if (diffInMinutes < 60) return `${diffInMinutes}m ago`

  const diffInHours = Math.floor(diffInMinutes / 60)
  if (diffInHours < 24) return `${diffInHours}h ago`

  const diffInDays = Math.floor(diffInHours / 24)
  if (diffInDays < 7) return `${diffInDays}d ago`

  return date.toLocaleDateString()
}

const isOwnMessage = (message) => {
  return message.message_from_user_id === currentUser.value?.id
}

const handleKeyPress = (event) => {
  if (event.key === 'Enter' && !event.shiftKey) {
    event.preventDefault()
    sendMessage()
  }
}

// Auto-refresh messages every 10 seconds
let refreshInterval = null

onMounted(async () => {
  // Fetch all groups messages on mount
  await getCurrentUser()
  await fetchAllGroupsMessages()

  // Start polling for new messages
  refreshInterval = setInterval(async () => {
    // Fetch messages for all groups that have been loaded
    const groupIds = Object.keys(messagesByGroup.value)
    for (const groupId of groupIds) {
      await fetchMessages(parseInt(groupId))
    }
  }, 10000) // 10 seconds
})

// Clean up interval on unmount
const cleanup = () => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
    refreshInterval = null
  }
}

// Function to reset new message count for a specific group
const resetNewMessageCount = async (groupId) => {
  console.log(`Resetting new message count for group ${groupId}`)
  console.log(
    `Before reset - newMessagesCountByGroup for group ${groupId}:`,
    newMessagesCountByGroup.value[groupId],
  )

  // Get the highest message ID for this group before resetting
  const currentMessages = messagesByGroup.value[groupId] || []
  const lastReadMessageId =
    currentMessages.length > 0 ? Math.max(...currentMessages.map((msg) => msg.id)) : 0

  console.log(`Last read message ID for group ${groupId}: ${lastReadMessageId}`)

  // Save the last read message ID to database
  if (currentUser.value && lastReadMessageId > 0) {
    try {
      const result = await callApi({
        query: `
          INSERT INTO chat_last_read_message (id_group, id_user, id_last_read_message) 
          VALUES (?, ?, ?) 
          ON DUPLICATE KEY UPDATE id_last_read_message = VALUES(id_last_read_message)
        `,
        params: [groupId, currentUser.value.id, lastReadMessageId],
        requiresAuth: true,
      })

      if (result.success) {
        console.log(`Saved last read message ID ${lastReadMessageId} for group ${groupId}`)
      } else {
        console.error(`Failed to save last read message ID for group ${groupId}`)
      }
    } catch (err) {
      console.error('Error saving last read message ID:', err)
    }
  }

  // Reset the count - FORCE IT TO 0
  newMessagesCountByGroup.value[groupId] = 0
  console.log(
    `After reset - newMessagesCountByGroup for group ${groupId}:`,
    newMessagesCountByGroup.value[groupId],
  )

  // Also reset current count if this is the active group
  if (groupId === props.groupId) {
    currentNewMessagesCount.value = 0
    console.log(`Reset currentNewMessagesCount to 0 for active group ${groupId}`)
  }

  // Force reactive update by triggering a change
  newMessagesCountByGroup.value = { ...newMessagesCountByGroup.value }
  console.log(`Final newMessagesCountByGroup after force update:`, newMessagesCountByGroup.value)

  // Update debug info if debug section is open and this is the current group
  if (showDebugSection.value && groupId === props.groupId) {
    updateDebugInfo()
  }

  console.log(`✓ Reset completed for group ${groupId}`)

  // Emit reset event to trigger parent update
  emit('reset-triggered', groupId)

  // Also force an immediate update of the hidden component's counts
  // by calling fetchAllGroupsMessages to recalculate unread counts
  if (props.groupId === 0) {
    // This is the hidden component, so we need to recalculate all counts
    console.log('Hidden component: Recalculating all unread counts after reset')
    await fetchAllGroupsMessages()
  }
}

// Function to fetch messages for all groups at once
const fetchAllGroupsMessages = async () => {
  try {
    console.log('Fetching messages for all groups...')

    if (!currentUser.value) {
      console.log('No current user available')
      return
    }

    // Get all groups this user is in
    const groupsResult = await callApi({
      query: `
        SELECT DISTINCT
          cg.id, 
          cg.name, 
          cg.description, 
          cg.id_user_owner, 
          cg.is_active 
        FROM chat_groups cg
        INNER JOIN chat_users cu ON cg.id = cu.id_chat_group
        WHERE cg.is_active = 1 
          AND cu.is_active = 1
          AND cu.id_user = ?
        ORDER BY cg.name ASC
      `,
      params: [currentUser.value.id],
      requiresAuth: true,
    })

    if (!groupsResult.success || !groupsResult.data) {
      console.log('No groups found or API failed:', groupsResult)
      return
    }

    console.log(`Found ${groupsResult.data.length} groups for user`)

    // Fetch messages and count unread messages for each group
    for (const group of groupsResult.data) {
      console.log(`Fetching messages for group: ${group.name} (ID: ${group.id})`)

      // Get the last read message ID for this group and user
      const lastReadResult = await callApi({
        query: `
          SELECT id_last_read_message 
          FROM chat_last_read_message 
          WHERE id_group = ? AND id_user = ?
        `,
        params: [group.id, currentUser.value.id],
        requiresAuth: true,
      })

      const lastReadMessageId =
        lastReadResult.success && lastReadResult.data && lastReadResult.data.length > 0
          ? lastReadResult.data[0].id_last_read_message
          : 0

      console.log(`Last read message ID for group ${group.name}: ${lastReadMessageId}`)

      // Count unread messages (messages with ID > last_read_message_id and not from current user)
      const unreadCountResult = await callApi({
        query: `
          SELECT COUNT(*) as unread_count
          FROM chat_messages 
          WHERE id_chat_group = ? 
            AND id > ?
            AND message_from_user_id != ?
        `,
        params: [group.id, lastReadMessageId, currentUser.value.id],
        requiresAuth: true,
      })

      const unreadCount =
        unreadCountResult.success && unreadCountResult.data && unreadCountResult.data.length > 0
          ? parseInt(unreadCountResult.data[0].unread_count)
          : 0

      console.log(`Unread messages count for group ${group.name}: ${unreadCount}`)

      // Fetch all messages for this group
      const messagesResult = await callApi({
        query: `
          SELECT 
            cm.id,
            cm.id_chat_group,
            cm.message_from_user_id,
            cm.chat_replay_to_message_id,
            cm.message,
            cm.time,
            u.username as sender_username,
            r.role_name as sender_role
          FROM chat_messages cm
          LEFT JOIN users u ON cm.message_from_user_id = u.id
          LEFT JOIN roles r ON u.role_id = r.id
          WHERE cm.id_chat_group = ?
          ORDER BY cm.id ASC
        `,
        params: [group.id],
        requiresAuth: true,
      })

      if (messagesResult.success && messagesResult.data) {
        // Store messages in group cache
        messagesByGroup.value[group.id] = messagesResult.data
        console.log(`Loaded ${messagesResult.data.length} messages for group ${group.name}`)

        // Set unread count for this group
        newMessagesCountByGroup.value[group.id] = unreadCount
        console.log(`Set unread count for group ${group.name}: ${unreadCount}`)
      } else {
        console.log(`No messages found for group ${group.name}`)
        messagesByGroup.value[group.id] = []
        newMessagesCountByGroup.value[group.id] = 0
      }
    }

    console.log('All groups messages loaded successfully')
    console.log('Groups loaded:', Object.keys(messagesByGroup.value))
    console.log('Total groups with messages:', Object.keys(messagesByGroup.value).length)
    console.log('Unread counts by group:', newMessagesCountByGroup.value)

    // If there's a selected group, set its messages as current
    if (props.groupId && messagesByGroup.value[props.groupId]) {
      messages.value = messagesByGroup.value[props.groupId]
      console.log(
        `Set current messages for group ${props.groupId}: ${messages.value.length} messages`,
      )
    }
  } catch (err) {
    console.error('Error fetching all groups messages:', err)
  }
}

// Click handlers to reset new message count
const handleMessagesContainerClick = async () => {
  console.log(`Chat area clicked for group ${props.groupId}`)
  console.log(
    `Current newMessagesCountByGroup for this group:`,
    newMessagesCountByGroup.value[props.groupId],
  )
  console.log(`All newMessagesCountByGroup:`, newMessagesCountByGroup.value)

  // Always reset the count when chat area is clicked
  console.log(`Reset new message count for group ${props.groupId} (chat area clicked)`)
  await resetNewMessageCount(props.groupId)
}

const handleInputClick = async () => {
  console.log(`Input clicked for group ${props.groupId}`)
  console.log(
    `Current newMessagesCountByGroup for this group:`,
    newMessagesCountByGroup.value[props.groupId],
  )
  console.log(`All newMessagesCountByGroup:`, newMessagesCountByGroup.value)

  // Always reset the count when input is clicked
  console.log(`Reset new message count for group ${props.groupId} (input clicked)`)
  await resetNewMessageCount(props.groupId)
}

// Debug functions
const toggleDebugSection = () => {
  showDebugSection.value = !showDebugSection.value
  if (showDebugSection.value) {
    updateDebugInfo()
  }
}

const updateDebugInfo = () => {
  let info = `=== DEBUG: NEW MESSAGES COUNT ===\n`
  info += `Current Group: ${props.groupName} (ID: ${props.groupId})\n`
  info += `Current User: ${currentUser.value?.username} (ID: ${currentUser.value?.id})\n\n`

  info += `=== CURRENT STATE ===\n`
  info += `Current newMessagesCount: ${currentNewMessagesCount.value}\n`
  info += `Total messages in current group: ${messages.value.length}\n`
  info += `Messages cached for this group: ${messagesByGroup.value[props.groupId]?.length || 0}\n\n`

  info += `=== ALL GROUPS COUNTS ===\n`
  if (Object.keys(newMessagesCountByGroup.value).length === 0) {
    info += `No new message counts tracked yet\n`
  } else {
    Object.keys(newMessagesCountByGroup.value).forEach((groupId) => {
      const count = newMessagesCountByGroup.value[groupId]
      info += `Group ${groupId}: ${count} new messages\n`
    })
  }

  info += `\n=== SYSTEM INFO ===\n`
  info += `Last Updated: ${new Date().toLocaleString()}\n`
  info += `Polling Interval: 10 seconds\n`
  info += `Debug Section Active: ${showDebugSection.value}\n`

  debugInfo.value = info
}

const testNewMessagesCount = () => {
  console.log('=== TESTING NEW MESSAGES COUNT ===')
  console.log('Current newMessagesCount:', currentNewMessagesCount.value)
  console.log('All newMessagesCountByGroup:', newMessagesCountByGroup.value)
  updateDebugInfo()
}

// Expose cleanup function and new message counts
defineExpose({
  cleanup,
  fetchAllGroupsMessages,
  getNewMessagesCount: (groupId) => {
    const count = newMessagesCountByGroup.value[groupId] || 0
    console.log(`Exposed getNewMessagesCount for group ${groupId}:`, count)
    return count
  },
  getAllNewMessagesCounts: () => {
    console.log('Exposed getAllNewMessagesCounts:', newMessagesCountByGroup.value)
    return newMessagesCountByGroup.value
  },
  resetGroupCount: (groupId) => {
    console.log(`Exposed resetGroupCount called for group ${groupId}`)
    newMessagesCountByGroup.value[groupId] = 0
    // Force reactive update
    newMessagesCountByGroup.value = { ...newMessagesCountByGroup.value }
    console.log(`Exposed resetGroupCount completed for group ${groupId}`)
  },
  focusMessageInput: () => {
    console.log('ChatMessages: focusMessageInput called')
    if (messageInputRef.value) {
      messageInputRef.value.focus()
      console.log('Message input focused')
    } else {
      console.log('Message input ref not available')
    }
  },
})

const toggleEmojiPicker = () => {
  showEmojiPicker.value = !showEmojiPicker.value
}

const addEmoji = (emoji) => {
  newMessage.value += emoji
  // Close the emoji picker after adding emoji
  showEmojiPicker.value = false
  // Focus back to the input after adding emoji
  nextTick(() => {
    if (messageInputRef.value) {
      messageInputRef.value.focus()
    }
  })
}

const closeEmojiPicker = () => {
  showEmojiPicker.value = false
}

// File upload functions
const selectFile = () => {
  fileInputRef.value?.click()
}

const handleFileSelect = (event) => {
  const file = event.target.files[0]
  if (!file) return

  // Reset error state
  fileError.value = ''

  // Check file size
  if (file.size > MAX_FILE_SIZE) {
    fileError.value = `File size (${(file.size / 1024 / 1024).toFixed(2)}MB) exceeds the 10MB limit`
    return
  }

  selectedFile.value = file
  uploadFileToChat()
}

const uploadFileToChat = async () => {
  if (!selectedFile.value) return

  isUploading.value = true
  uploadProgress.value = 0
  fileError.value = ''

  try {
    // Create folder name for this chat group
    const chatFolder = `chat_files/group_${props.groupId}`

    // Generate unique filename with timestamp
    const timestamp = Date.now()
    const fileExtension = selectedFile.value.name.split('.').pop()
    const customFilename = `${timestamp}_${selectedFile.value.name}`

    console.log('Uploading file to chat:', {
      fileName: selectedFile.value.name,
      fileSize: selectedFile.value.size,
      chatFolder,
      customFilename,
    })

    // Upload the file
    const uploadResult = await uploadFile(selectedFile.value, chatFolder, customFilename)

    if (uploadResult.success) {
      // Create a file message with special prefix
      const fileMessage = {
        id_chat_group: props.groupId,
        message_from_user_id: currentUser.value.id,
        message: `[FILE]${uploadResult.relativePath}|${selectedFile.value.name}|${selectedFile.value.size}|${selectedFile.value.type}`,
      }

      // Send the file message to the database
      const result = await callApi({
        query: `
          INSERT INTO chat_messages (id_chat_group, message_from_user_id, message) 
          VALUES (?, ?, ?)
        `,
        params: [fileMessage.id_chat_group, fileMessage.message_from_user_id, fileMessage.message],
        requiresAuth: true,
      })

      if (result.success) {
        // Add the new message to both current list and group cache
        const newMsg = {
          id: result.lastInsertId,
          id_chat_group: props.groupId,
          message_from_user_id: currentUser.value.id,
          message: fileMessage.message,
          time: new Date().toISOString(),
          sender_username: currentUser.value.username,
          sender_role: currentUser.value.role_name,
        }

        messages.value.push(newMsg)
        messagesByGroup.value[props.groupId] = [...messages.value]

        // Emit event to parent
        emit('message-sent', newMsg)

        // Scroll to bottom after message is added
        await nextTick()
        await scrollToBottom()

        console.log('File message sent successfully:', newMsg)
      } else {
        throw new Error('Failed to save file message to database')
      }
    } else {
      throw new Error(uploadResult.message || 'File upload failed')
    }
  } catch (err) {
    console.error('Error uploading file:', err)
    fileError.value = err.message || 'Failed to upload file'
  } finally {
    isUploading.value = false
    uploadProgress.value = 0
    selectedFile.value = null
    // Clear the file input
    if (fileInputRef.value) {
      fileInputRef.value.value = ''
    }
  }
}

const downloadFile = async (filePath, fileName) => {
  try {
    const fileUrl = getFileUrl(filePath)
    const response = await fetch(fileUrl)

    if (!response.ok) {
      throw new Error('Failed to download file')
    }

    const blob = await response.blob()
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = fileName
    document.body.appendChild(a)
    a.click()
    window.URL.revokeObjectURL(url)
    document.body.removeChild(a)
  } catch (err) {
    console.error('Error downloading file:', err)
    alert('Failed to download file')
  }
}

const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const getFileIcon = (fileType) => {
  if (fileType.startsWith('image/')) return '🖼️'
  if (fileType.startsWith('video/')) return '🎥'
  if (fileType.startsWith('audio/')) return '🎵'
  if (fileType.includes('pdf')) return '📄'
  if (fileType.includes('word') || fileType.includes('document')) return '📝'
  if (fileType.includes('excel') || fileType.includes('spreadsheet')) return '📊'
  if (fileType.includes('powerpoint') || fileType.includes('presentation')) return '📽️'
  if (fileType.includes('zip') || fileType.includes('rar') || fileType.includes('archive'))
    return '📦'
  return '📎'
}

// Parse file message to extract file information
const parseFileMessage = (message) => {
  if (!message.startsWith('[FILE]')) return null

  const fileData = message.substring(6) // Remove '[FILE]' prefix
  const parts = fileData.split('|')

  if (parts.length >= 4) {
    return {
      filePath: parts[0],
      fileName: parts[1],
      fileSize: parseInt(parts[2]),
      fileType: parts[3],
    }
  }

  return null
}

// Check if message is a file message
const isFileMessage = (message) => {
  return message.startsWith('[FILE]')
}
</script>

<template>
  <div class="chat-messages">
    <div class="messages-header">
      <h3><i class="fas fa-comments"></i> {{ groupName }}</h3>
      <div class="group-info">
        <button @click="toggleDebugSection" class="debug-btn" title="Debug New Messages Count">
          <i class="fas fa-bug"></i>
        </button>
        <span class="timezone-info">
          <i class="fas fa-clock"></i> {{ Intl.DateTimeFormat().resolvedOptions().timeZone }}
        </span>
        <span class="member-count">
          <i class="fas fa-users"></i> {{ messages.length }} messages
        </span>
      </div>
    </div>

    <div class="messages-container" ref="messagesContainer" @click="handleMessagesContainerClick">
      <div v-if="loading && messages.length === 0" class="loading-messages">
        <i class="fas fa-spinner fa-spin"></i> Loading messages...
      </div>

      <div v-else-if="error" class="error-messages">
        <i class="fas fa-exclamation-triangle"></i> {{ error }}
      </div>

      <div v-else-if="messages.length === 0" class="no-messages">
        <i class="fas fa-comment-slash"></i>
        <p>No messages yet</p>
        <p class="sub-text">Start the conversation!</p>
      </div>

      <div v-else class="messages-list">
        <div
          v-for="(message, index) in messages"
          :key="message.id"
          class="message-wrapper"
          :class="{ 'own-message': isOwnMessage(message) }"
        >
          <!-- Date separator -->
          <div
            v-if="index === 0 || formatDate(message.time) !== formatDate(messages[index - 1].time)"
            class="date-separator"
          >
            {{ formatDate(message.time) }}
          </div>

          <div class="message">
            <div class="message-header">
              <span class="sender-name">{{ message.sender_username }}</span>
              <div class="time-info">
                <span class="relative-time">{{ getRelativeTime(message.time) }}</span>
                <span class="absolute-time">{{ formatTime(message.time) }}</span>
              </div>
            </div>

            <div class="message-content">
              <!-- Show file attachment if it's a file message -->
              <div v-if="isFileMessage(message.message)" class="file-attachment">
                <div class="file-info">
                  <span class="file-icon">{{
                    getFileIcon(parseFileMessage(message.message)?.fileType)
                  }}</span>
                  <div class="file-details">
                    <span class="file-name">{{ parseFileMessage(message.message)?.fileName }}</span>
                    <span class="file-size">{{
                      formatFileSize(parseFileMessage(message.message)?.fileSize)
                    }}</span>
                  </div>
                </div>
                <button
                  @click="
                    downloadFile(
                      parseFileMessage(message.message)?.filePath,
                      parseFileMessage(message.message)?.fileName,
                    )
                  "
                  class="download-btn"
                  title="Download file"
                >
                  <i class="fas fa-download"></i>
                </button>
              </div>

              <!-- Show regular text message if it's not a file message -->
              <div v-else>
                {{ message.message }}
              </div>
            </div>

            <div class="message-footer">
              <span class="sender-role">{{ message.sender_role }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="message-input-container">
      <div class="input-wrapper">
        <button @click="toggleEmojiPicker" class="emoji-button" title="Add emoji" type="button">
          <i class="fas fa-smile"></i>
        </button>
        <button
          @click="selectFile"
          class="file-button"
          title="Attach file"
          type="button"
          :disabled="isUploading"
        >
          <i :class="isUploading ? 'fas fa-spinner fa-spin' : 'fas fa-paperclip'"></i>
        </button>
        <textarea
          v-model="newMessage"
          @keypress="handleKeyPress"
          @click="handleInputClick"
          placeholder="Type your message..."
          class="message-input"
          :disabled="isSending"
          rows="1"
          maxlength="1000"
          ref="messageInputRef"
        ></textarea>
        <button
          @click="
            (e) => {
              e.preventDefault()
              sendMessage()
            }
          "
          :disabled="!newMessage.trim() || isSending"
          class="send-button"
        >
          <i :class="isSending ? 'fas fa-spinner fa-spin' : 'fas fa-paper-plane'"></i>
        </button>
      </div>

      <!-- File upload progress and error -->
      <div v-if="isUploading" class="upload-progress">
        <div class="progress-bar">
          <div class="progress-fill" :style="{ width: uploadProgress + '%' }"></div>
        </div>
        <span class="progress-text">Uploading file...</span>
      </div>

      <div v-if="fileError" class="file-error">
        <i class="fas fa-exclamation-triangle"></i>
        {{ fileError }}
      </div>

      <div class="input-footer">
        <span class="char-count">{{ newMessage.length }}/1000</span>
        <span class="send-hint">Press Enter to send, Shift+Enter for new line</span>
      </div>

      <!-- Hidden file input -->
      <input
        ref="fileInputRef"
        type="file"
        @change="handleFileSelect"
        style="display: none"
        accept="*/*"
      />
    </div>

    <!-- Emoji Picker Modal -->
    <div v-if="showEmojiPicker" class="emoji-picker-modal" @click="closeEmojiPicker">
      <div class="emoji-picker-content" @click.stop>
        <div class="emoji-picker-header">
          <h4>Select Emoji</h4>
          <button @click="closeEmojiPicker" class="close-emoji-btn">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <div class="emoji-grid">
          <button
            v-for="emoji in emojis"
            :key="emoji"
            @click="addEmoji(emoji)"
            class="emoji-item"
            :title="emoji"
          >
            {{ emoji }}
          </button>
        </div>
      </div>
    </div>

    <!-- Debug Info Modal -->
    <div v-if="showDebugSection" class="debug-info-modal">
      <div class="debug-info-content">
        <h3>Debug: New Messages Count</h3>
        <pre>{{ debugInfo }}</pre>
        <div class="debug-buttons">
          <button @click="testNewMessagesCount" class="test-btn">Test Count</button>
          <button @click="resetNewMessageCount(props.groupId)" class="reset-btn">
            Reset Count
          </button>
          <button @click="toggleDebugSection">Close</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.chat-messages {
  display: flex;
  flex-direction: column;
  height: 100%;
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.messages-header {
  padding: 16px 20px;
  background-color: #06b6d4;
  color: white;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.messages-header h3 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 1.1rem;
}

.group-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.timezone-info {
  font-size: 0.9rem;
  opacity: 0.9;
  display: flex;
  align-items: center;
  gap: 4px;
}

.member-count {
  font-size: 0.9rem;
  opacity: 0.9;
  display: flex;
  align-items: center;
  gap: 4px;
}

.messages-container {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  background-color: #f8fafc;
}

.loading-messages,
.error-messages,
.no-messages {
  text-align: center;
  padding: 40px 20px;
  color: #64748b;
}

.error-messages {
  color: #ef4444;
}

.no-messages i {
  font-size: 3rem;
  color: #06b6d4;
  margin-bottom: 16px;
  display: block;
}

.sub-text {
  font-size: 0.9rem;
  color: #9ca3af;
  margin-top: 8px;
}

.messages-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.message-wrapper {
  display: flex;
  flex-direction: column;
}

.message-wrapper.own-message {
  align-items: flex-end;
}

.date-separator {
  text-align: center;
  margin: 20px 0 12px 0;
  color: #64748b;
  font-size: 0.8rem;
  font-weight: 500;
  position: relative;
}

.date-separator::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 0;
  right: 0;
  height: 1px;
  background-color: #e2e8f0;
  z-index: 1;
}

.date-separator span {
  background-color: #f8fafc;
  padding: 0 12px;
  position: relative;
  z-index: 2;
}

.message {
  max-width: 70%;
  padding: 12px 16px;
  border-radius: 12px;
  background-color: white;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
  border: 1px solid #e2e8f0;
}

.own-message .message {
  background-color: #06b6d4;
  color: white;
  border-color: #06b6d4;
}

.message-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.sender-name {
  font-weight: 600;
  font-size: 0.9rem;
}

.own-message .sender-name {
  color: white;
}

.time-info {
  font-size: 0.8rem;
  opacity: 0.7;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 2px;
}

.own-message .time-info {
  color: rgba(255, 255, 255, 0.8);
}

.relative-time {
  font-size: 0.75rem;
  font-weight: 500;
}

.absolute-time {
  font-size: 0.7rem;
  opacity: 0.8;
}

.message-content {
  line-height: 1.4;
  word-wrap: break-word;
  white-space: pre-wrap;
}

.message-footer {
  margin-top: 4px;
}

.sender-role {
  font-size: 0.7rem;
  opacity: 0.6;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.own-message .sender-role {
  color: rgba(255, 255, 255, 0.7);
}

.message-input-container {
  padding: 16px 20px;
  background-color: white;
  border-top: 1px solid #e2e8f0;
}

.input-wrapper {
  display: flex;
  align-items: flex-end;
  gap: 8px;
  padding: 12px 16px;
  background-color: white;
  border-top: 1px solid #e2e8f0;
}

.emoji-button {
  background-color: transparent;
  border: none;
  color: #64748b;
  font-size: 1.2rem;
  cursor: pointer;
  padding: 8px;
  border-radius: 4px;
  transition: background-color 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 40px;
  height: 40px;
}

.emoji-button:hover {
  background-color: #f1f5f9;
  color: #06b6d4;
}

.file-button {
  background-color: transparent;
  border: none;
  color: #64748b;
  font-size: 1.2rem;
  cursor: pointer;
  padding: 8px;
  border-radius: 4px;
  transition: background-color 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 40px;
  height: 40px;
}

.file-button:hover {
  background-color: #f1f5f9;
  color: #06b6d4;
}

.message-input {
  flex: 1;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 12px;
  font-size: 0.95rem;
  resize: none;
  outline: none;
  transition: border-color 0.2s;
  font-family: inherit;
  line-height: 1.4;
  max-height: 120px;
  min-height: 40px;
}

.message-input:focus {
  border-color: #06b6d4;
}

.message-input:disabled {
  background-color: #f8fafc;
  color: #64748b;
  cursor: not-allowed;
}

.send-button {
  background-color: #06b6d4;
  color: white;
  border: none;
  border-radius: 8px;
  padding: 12px 16px;
  cursor: pointer;
  font-size: 1rem;
  transition: background-color 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 40px;
  height: 40px;
}

.send-button:hover:not(:disabled) {
  background-color: #0891b2;
}

.send-button:disabled {
  background-color: #cbd5e1;
  cursor: not-allowed;
}

.input-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  background-color: #f8fafc;
  border-top: 1px solid #e2e8f0;
  font-size: 0.8rem;
  color: #64748b;
}

.char-count {
  font-weight: 500;
}

.send-hint {
  font-style: italic;
}

.debug-btn {
  background-color: rgba(255, 255, 255, 0.2);
  border: none;
  color: white;
  font-size: 0.8rem;
  cursor: pointer;
  padding: 6px 12px;
  border-radius: 4px;
  transition: background-color 0.2s;
  display: flex;
  align-items: center;
  gap: 4px;
}

.debug-btn:hover {
  background-color: rgba(255, 255, 255, 0.3);
}

.debug-info-modal {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.debug-info-content {
  background-color: white;
  padding: 24px;
  border-radius: 12px;
  width: 90%;
  max-width: 700px;
  max-height: 80vh;
  overflow-y: auto;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
}

.debug-info-content h3 {
  margin-top: 0;
  margin-bottom: 16px;
  color: #06b6d4;
  border-bottom: 2px solid #e2e8f0;
  padding-bottom: 8px;
}

.debug-info-content pre {
  white-space: pre-wrap;
  word-break: break-word;
  font-family: 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.4;
  background-color: #f8fafc;
  padding: 16px;
  border-radius: 6px;
  border: 1px solid #e2e8f0;
  max-height: 400px;
  overflow-y: auto;
}

.debug-buttons {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 16px;
}

.debug-info-content button {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  background-color: #06b6d4;
  color: white;
  cursor: pointer;
  font-weight: 500;
  transition: background-color 0.2s;
}

.debug-info-content button:hover {
  background-color: #0891b2;
}

.test-btn {
  background-color: #10b981 !important;
}

.test-btn:hover {
  background-color: #059669 !important;
}

.reset-btn {
  background-color: #ef4444 !important;
}

.reset-btn:hover {
  background-color: #dc2626 !important;
}

/* Responsive design */
@media (max-width: 768px) {
  .message {
    max-width: 85%;
  }

  .messages-header {
    padding: 12px 16px;
  }

  .messages-container {
    padding: 16px;
  }

  .message-input-container {
    padding: 12px 16px;
  }

  .input-footer {
    flex-direction: column;
    gap: 4px;
    align-items: flex-start;
  }
}

.emoji-picker-modal {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.emoji-picker-content {
  background-color: white;
  border-radius: 12px;
  width: 90%;
  max-width: 400px;
  max-height: 500px;
  overflow: hidden;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
  display: flex;
  flex-direction: column;
}

.emoji-picker-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background-color: #f8fafc;
  border-bottom: 1px solid #e2e8f0;
}

.emoji-picker-header h4 {
  margin: 0;
  color: #374151;
  font-size: 1rem;
}

.close-emoji-btn {
  background-color: transparent;
  border: none;
  color: #64748b;
  font-size: 1.2rem;
  cursor: pointer;
  padding: 8px;
  border-radius: 4px;
  transition: background-color 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
}

.close-emoji-btn:hover {
  background-color: #e2e8f0;
  color: #374151;
}

.emoji-grid {
  display: grid;
  grid-template-columns: repeat(8, 1fr);
  gap: 4px;
  padding: 16px;
  max-height: 400px;
  overflow-y: auto;
}

.emoji-item {
  background-color: transparent;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  padding: 8px;
  border-radius: 6px;
  transition: background-color 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
}

.emoji-item:hover {
  background-color: #f1f5f9;
  transform: scale(1.1);
}

.upload-progress {
  margin-top: 8px;
  margin-bottom: 8px;
  height: 20px;
  background-color: #f8fafc;
  border-radius: 10px;
  overflow: hidden;
}

.progress-bar {
  height: 100%;
  background-color: #06b6d4;
  border-radius: 10px;
  transition: width 0.2s;
}

.progress-fill {
  height: 100%;
  background-color: #06b6d4;
  border-radius: 10px;
  transition: width 0.2s;
}

.progress-text {
  font-size: 0.8rem;
  font-weight: 500;
  color: #374151;
}

.file-error {
  margin-top: 8px;
  margin-bottom: 8px;
  padding: 8px;
  background-color: #fef2f2;
  border: 1px solid #fef2f2;
  border-radius: 6px;
  color: #ef4444;
  font-size: 0.8rem;
  display: flex;
  align-items: center;
  gap: 6px;
}

.file-attachment {
  margin-top: 8px;
  padding: 8px;
  background-color: rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.own-message .file-attachment {
  background-color: rgba(255, 255, 255, 0.2);
}

.file-info {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}

.file-icon {
  font-size: 1.2rem;
}

.file-details {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.file-name {
  font-weight: 500;
  font-size: 0.9rem;
  word-break: break-word;
}

.file-size {
  font-size: 0.7rem;
  opacity: 0.7;
}

.download-btn {
  background-color: rgba(255, 255, 255, 0.2);
  border: none;
  color: inherit;
  padding: 6px 8px;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 32px;
  height: 32px;
}

.download-btn:hover {
  background-color: rgba(255, 255, 255, 0.3);
}
</style>
