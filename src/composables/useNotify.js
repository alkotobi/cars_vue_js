// Composable: centralised notification/confirm system powered by MessageBox.
// The view owns a single <MessageBox> instance; this composable exposes the
// reactive state and helpers without any UI markup.
import { ref, onScopeDispose } from 'vue'

export function useNotify() {
  const msgBox = ref({
    show: false,
    type: 'info',
    title: '',
    message: '',
    details: '',
    actionText: 'OK',
    cancelText: 'Cancel',
    showCancel: false,
    onConfirm: null,
    onCancel: null
  })

  let confirmResolver = null
  let cancelResolver = null

  const settle = (result) => {
    if (confirmResolver) {
      confirmResolver(result)
      confirmResolver = null
    }
    if (cancelResolver) {
      cancelResolver(result)
      cancelResolver = null
    }
    msgBox.value.show = false
    msgBox.value.onConfirm = null
    msgBox.value.onCancel = null
  }

  const notify = async (type, title, message, details = '', actionText = 'OK') => {
    return new Promise((resolve) => {
      confirmResolver = resolve
      cancelResolver = null
      msgBox.value = {
        show: true,
        type,
        title,
        message,
        details,
        actionText,
        cancelText: 'Cancel',
        showCancel: false,
        onConfirm: () => settle(true),
        onCancel: () => settle(false)
      }
    })
  }

  const notifyError = async (title, message, details = '') => {
    return notify('error', title, message, details, 'OK')
  }

  const onConfirm = () => {
    if (msgBox.value.onConfirm) {
      msgBox.value.onConfirm()
    } else {
      settle(true)
    }
  }

  const onCancel = () => {
    if (msgBox.value.onCancel) {
      msgBox.value.onCancel()
    } else {
      settle(false)
    }
  }

  const onClose = () => {
    settle(false)
  }

  onScopeDispose(() => {
    settle(false)
  })

  return {
    msgBox,
    notify,
    notifyError,
    onConfirm,
    onCancel,
    onClose,
    settle
  }
}
