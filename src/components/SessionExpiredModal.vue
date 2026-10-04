<script setup>
// Blocking "your session is gone" modal.
//
// Mounted once in App.vue and driven entirely by useSessionLost.js, so any request
// anywhere in the app that comes back not_authenticated opens this - no call site
// has to notice the code, and no caller can render it as its own string.
//
// It is deliberately not dismissable. A dead token means every subsequent read and
// every subsequent save fails the same way, so "dismiss" is not an option the user
// actually has; the single action takes them to /login, which is the only state
// from which the app works again. The credential has already been dropped from
// localStorage by reportSessionLost(), so the router guard lets that navigation
// through instead of bouncing them straight back.
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useEnhancedI18n } from '../composables/useI18n'
import {
  armSessionAlertGate,
  onSessionLost,
  releaseSessionAlertGate,
} from '../composables/useSessionLost'

const { t } = useEnhancedI18n()
const router = useRouter()
const route = useRoute()

const show = ref(false)
const loginButton = ref(null)

// A share link (/clients) is reachable without a session and its action is on the
// server's public allowlist, so nothing should raise this there - but if a request
// ever did, the user is already looking at a page with no session and a modal
// telling them to sign in would be noise. Same for /login itself: arriving there is
// the outcome of this modal, and re-arming on that route would trap the user in a
// loop where dismissing is impossible.
const exemptRoute = computed(() => route.path === '/login' || route.name === 'client-details')
const visible = computed(() => show.value && !exemptRoute.value)

let stopListening = null
let previousBodyOverflow = ''

function lockScroll() {
  previousBodyOverflow = document.body.style.overflow
  document.body.style.overflow = 'hidden'
}

function unlockScroll() {
  document.body.style.overflow = previousBodyOverflow
}

// Send the user where the app still works. localStorage.user is already gone, so
// the beforeEach guard's `!user` branch is what lets /login render rather than
// redirect. replace, not push: the screen they were on is not reachable any more,
// and Back into it would just fail again.
function goToLogin() {
  // Releasing the alert gate and unlocking the scroll are both handled by the
  // watcher on `visible` below, so there is one teardown path rather than two that
  // have to be kept in step.
  show.value = false
  router.replace('/login')
}

// Undo everything this modal took from the page, whenever it stops being visible -
// whether that is because the reader signed in or because they navigated somewhere
// it does not apply. Keyed on `visible` and not `show` so an exempt route releases
// the gate too: on /login the modal never appears, and an armed gate there would
// silence alerts for a session that is not the problem.
watch(visible, (isVisible) => {
  if (isVisible) return
  releaseSessionAlertGate()
  unlockScroll()
})

// Keep Tab inside the dialog. There is one focusable control, so this is the
// difference between a keyboard user reaching the login form and getting dropped
// into the page behind, where every button is a no-op against a dead token.
function onKeydown(event) {
  if (!visible.value) return
  if (event.key === 'Tab') {
    event.preventDefault()
    loginButton.value?.focus()
  }
  // Escape and backdrop clicks are intentionally ignored - see the note at the
  // top of the file. Swallowing Escape also stops it reaching the view-level
  // keydown handlers, which would otherwise close dialogs behind this one.
}

onMounted(() => {
  stopListening = onSessionLost(() => {
    show.value = true
    // Armed here rather than inside reportSessionLost() because the listener loop
    // there is synchronous: this still runs before the failing request's caller can
    // reach its own alert(), which is the case the gate exists for. Doing it in this
    // component also means the exempt-route check can run first.
    if (!exemptRoute.value) armSessionAlertGate()
    lockScroll()
    nextTick(() => loginButton.value?.focus())
  })
  document.addEventListener('keydown', onKeydown, true)
})

onUnmounted(() => {
  if (stopListening) {
    stopListening()
    stopListening = null
  }
  document.removeEventListener('keydown', onKeydown, true)
  // Watchers are torn down before onUnmounted, so this path needs its own release -
  // without it a modal removed by a render crash would leave alerts dead app-wide.
  releaseSessionAlertGate()
  if (visible.value) unlockScroll()
})
</script>

<template>
  <teleport to="body">
    <div v-if="visible" class="session-expired-overlay">
      <div
        class="session-expired-dialog"
        role="alertdialog"
        aria-modal="true"
        aria-labelledby="session-expired-title"
        aria-describedby="session-expired-message"
      >
        <div class="session-expired-header">
          <i class="fas fa-user-lock" aria-hidden="true"></i>
          <h3 id="session-expired-title">{{ t('sessionExpired.title') }}</h3>
        </div>

        <div class="session-expired-body">
          <p id="session-expired-message" class="session-expired-lead">
            {{ t('sessionExpired.message') }}
          </p>
          <p class="session-expired-hint">
            <i class="fas fa-circle-info" aria-hidden="true"></i>
            <span>{{ t('sessionExpired.hint') }}</span>
          </p>
        </div>

        <div class="session-expired-footer">
          <button ref="loginButton" type="button" class="login-btn" @click="goToLogin">
            <i class="fas fa-right-to-bracket" aria-hidden="true"></i>
            {{ t('sessionExpired.action') }}
          </button>
        </div>
      </div>
    </div>
  </teleport>
</template>

<style scoped>
.session-expired-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(15, 23, 42, 0.72);
  backdrop-filter: blur(6px);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 10005;
  padding: 20px;
  animation: sessionExpiredFadeIn 0.25s ease-out;
}

/* Above DatabaseVersionCheck's 10000: that one is a login-page gate, and if both
   ever render at once the dead session is the one the user has to act on now. */
.session-expired-dialog {
  background: linear-gradient(145deg, #ffffff, #f8fafc);
  border-radius: 16px;
  box-shadow:
    0 24px 48px rgba(15, 23, 42, 0.35),
    0 0 0 1px rgba(15, 23, 42, 0.08);
  width: 100%;
  max-width: 460px;
  overflow: hidden;
  animation: sessionExpiredSlideIn 0.3s ease-out;
}

.session-expired-header {
  background: linear-gradient(135deg, #f59e0b, #ea580c);
  padding: 26px 24px;
  text-align: center;
}

.session-expired-header i {
  display: block;
  font-size: 42px;
  color: #ffffff;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  margin-bottom: 12px;
  animation: sessionExpiredPulse 2.4s infinite;
}

.session-expired-header h3 {
  margin: 0;
  color: #ffffff;
  font-size: 22px;
  font-weight: 600;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
}

.session-expired-body {
  padding: 28px 24px 8px;
}

.session-expired-lead {
  margin: 0 0 18px;
  font-size: 17px;
  line-height: 1.6;
  color: #1f2937;
  text-align: center;
  font-weight: 500;
}

.session-expired-hint {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin: 0;
  padding: 14px 16px;
  border-radius: 10px;
  background: #fffbeb;
  border-inline-start: 4px solid #f59e0b;
  color: #78350f;
  font-size: 14px;
  line-height: 1.5;
}

.session-expired-hint i {
  margin-top: 2px;
  font-size: 15px;
  flex-shrink: 0;
}

.session-expired-footer {
  padding: 20px 24px 26px;
  display: flex;
  justify-content: center;
}

.login-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  width: 100%;
  padding: 13px 24px;
  border: none;
  border-radius: 10px;
  background: linear-gradient(135deg, #2563eb, #1d4ed8);
  color: #ffffff;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition:
    transform 0.2s ease,
    box-shadow 0.2s ease,
    background 0.2s ease;
}

.login-btn:hover {
  background: linear-gradient(135deg, #1d4ed8, #1e40af);
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(37, 99, 235, 0.35);
}

.login-btn:focus-visible {
  outline: 3px solid rgba(37, 99, 235, 0.45);
  outline-offset: 2px;
}

@keyframes sessionExpiredFadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

@keyframes sessionExpiredSlideIn {
  from {
    opacity: 0;
    transform: translateY(-30px) scale(0.95);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

@keyframes sessionExpiredPulse {
  0%,
  100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.08);
  }
}
</style>
