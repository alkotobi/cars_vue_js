<script setup>
import { ref, watch, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useApi, loginErrorText } from '../composables/useApi'
import { useEnhancedI18n } from '../composables/useI18n'
import LanguageSwitcher from '../components/LanguageSwitcher.vue'
import { armSessionLost } from '../composables/useSessionLost'
import { resetToEnglish } from '../i18n'

const router = useRouter()
const { callApi, error: apiError, loading } = useApi()
const { t } = useEnhancedI18n()

const username = ref('')
const password = ref('')
const error = ref('')
const showChangePassword = ref(false)
const newPassword = ref('')
const confirmPassword = ref('')
const successMessage = ref('')
const messageTimeout = ref(null)
const isProcessing = ref(false)
const showPassword = ref(false)
const showNewPassword = ref(false)
const showConfirmPassword = ref(false)

// Clear messages after a delay
const clearMessages = () => {
  if (messageTimeout.value) {
    clearTimeout(messageTimeout.value)
  }
  messageTimeout.value = setTimeout(() => {
    error.value = ''
    successMessage.value = ''
  }, 3000) // Messages will disappear after 3 seconds
}

// Watch for changes in error or success messages
watch([error, successMessage], () => {
  if (error.value || successMessage.value) {
    clearMessages()
  }
})

const login = async () => {
  try {
    if (isProcessing.value) return // Prevent double-clicks

    if (!username.value || !password.value) {
      error.value = t('auth.enterUsernamePassword')
      return
    }

    isProcessing.value = true // Start processing
    error.value = '' // Clear previous errors

    // The server verifies the password and hands back a token. This used to
    // SELECT the user row (hash included) and call `verify_password` from the
    // browser, which meant the hashes were readable through the generic query
    // endpoint and every request after login was trusted on the client's word.
    // The token in the stored user is what api/lib/auth.php checks.
    const result = await callApi({
      action: 'login',
      username: username.value,
      password: password.value,
    })

    if (result.success && result.token && result.user) {
      const userInfo = {
        ...result.user,
        token: result.token,
        permissions: result.user.permissions || [],
      }

      localStorage.setItem('user', JSON.stringify(userInfo))

      // A fresh session has to be able to report a fresh loss. useSessionLost dedupes
      // repeated reports by token, and a sign-out with no token at all dedupes under a
      // sentinel - so without re-arming here, the *second* sign-out of a reader who had
      // never held a token would be swallowed as a duplicate of the first.
      armSessionLost()

      // Clear logo cache by updating assets version to force reload
      const STORAGE_KEY = 'assets_version'
      localStorage.setItem(STORAGE_KEY, Date.now().toString())

      // Dispatch custom events for header to update
      window.dispatchEvent(new CustomEvent('userLogin'))
      window.dispatchEvent(new CustomEvent('forceUpdateTasks'))

      router.push('/cars')
    } else {
      error.value = loginErrorText(t, result)
      // The raw envelope, because the copy above is a summary of it. A failed login
      // was previously indistinguishable from a dead database here, which sent the
      // reader - and whoever was helping them - after the credentials instead.
      console.error('Login rejected:', result)
    }
  } catch (err) {
    error.value = t('auth.loginError')
    console.error(err)
  } finally {
    isProcessing.value = false // End processing
  }
}

// Update the changePassword function verification check as well
const changePassword = async () => {
  try {
    if (isProcessing.value) return // Prevent double-clicks

    if (!username.value || !password.value) {
      error.value = t('auth.enterCurrentCredentials')
      return
    }

    if (!newPassword.value || !confirmPassword.value) {
      error.value = t('auth.enterNewPassword')
      return
    }

    if (newPassword.value !== confirmPassword.value) {
      error.value = t('auth.passwordsNotMatch')
      return
    }

    isProcessing.value = true // Start processing
    error.value = '' // Clear previous errors

    // One call that verifies the current password and writes the new one server-side.
    //
    // This was three anonymous calls: a SELECT that shipped the stored bcrypt hash
    // to the browser, `verify_password` to compare it over there, and
    // `UPDATE users SET password = ? WHERE username = ?` with the username read
    // straight off the form. Anyone could rewrite anyone's password, and anyone
    // could ask for a hash.
    const updateResult = await callApi({
      action: 'change_password_with_credentials',
      username: username.value,
      current_password: password.value,
      new_password: newPassword.value,
    })

    if (updateResult.success) {
      // The action rotates api_token server-side, so the token this tab is holding
      // is already dead. Logging out is best-effort and must not hide the success
      // message.
      const currentToken = (() => {
        const userStr = localStorage.getItem('user')
        return userStr ? JSON.parse(userStr)?.token : null
      })()
      if (currentToken) {
        callApi({ action: 'logout', token: currentToken }).catch(() => {})
      }

      successMessage.value = t('auth.passwordChangeSuccess')
      showChangePassword.value = false
      newPassword.value = ''
      confirmPassword.value = ''
      password.value = ''
    } else {
      error.value = t('auth.passwordChangeError')
    }
  } catch (err) {
    error.value = t('auth.passwordChangeErrorOccurred')
    console.error(err)
  } finally {
    isProcessing.value = false // End processing
  }
}

// Clean up timeout when component is unmounted
onUnmounted(() => {
  if (messageTimeout.value) {
    clearTimeout(messageTimeout.value)
  }
})
</script>

<template>
  <div class="login-container">
    <!-- Language Switcher -->
    <div class="language-switcher-container">
      <LanguageSwitcher />
    </div>

    <form @submit.prevent="showChangePassword ? changePassword() : login()" class="login-form">
      <h2>
        <i :class="showChangePassword ? 'fas fa-key' : 'fas fa-sign-in-alt'"></i>
        {{ showChangePassword ? $t('auth.changePassword') : $t('auth.login') }}
      </h2>

      <!-- Username field -->
      <div class="form-group">
        <div class="input-with-icon">
          <i class="fas fa-user"></i>
          <input
            type="text"
            v-model="username"
            :placeholder="$t('auth.username')"
            required
            :disabled="isProcessing"
            autocomplete="username"
          />
        </div>
      </div>

      <!-- Current Password field -->
      <div class="form-group">
        <div class="input-with-icon">
          <i class="fas fa-lock"></i>
          <input
            :type="showPassword ? 'text' : 'password'"
            v-model="password"
            :placeholder="$t('auth.currentPassword')"
            required
            :disabled="isProcessing"
            autocomplete="current-password"
          />
          <button
            type="button"
            class="toggle-password"
            @click="showPassword = !showPassword"
            :disabled="isProcessing"
          >
            <i :class="showPassword ? 'fas fa-eye-slash' : 'fas fa-eye'"></i>
          </button>
        </div>
      </div>

      <!-- Change Password Fields -->
      <template v-if="showChangePassword">
        <div class="form-group">
          <div class="input-with-icon">
            <i class="fas fa-key"></i>
            <input
              :type="showNewPassword ? 'text' : 'password'"
              v-model="newPassword"
              :placeholder="$t('auth.newPassword')"
              required
              :disabled="isProcessing"
              autocomplete="new-password"
            />
            <button
              type="button"
              class="toggle-password"
              @click="showNewPassword = !showNewPassword"
              :disabled="isProcessing"
            >
              <i :class="showNewPassword ? 'fas fa-eye-slash' : 'fas fa-eye'"></i>
            </button>
          </div>
        </div>
        <div class="form-group">
          <div class="input-with-icon">
            <i class="fas fa-key"></i>
            <input
              :type="showConfirmPassword ? 'text' : 'password'"
              v-model="confirmPassword"
              :placeholder="$t('auth.confirmPassword')"
              required
              :disabled="isProcessing"
              autocomplete="new-password"
            />
            <button
              type="button"
              class="toggle-password"
              @click="showConfirmPassword = !showConfirmPassword"
              :disabled="isProcessing"
            >
              <i :class="showConfirmPassword ? 'fas fa-eye-slash' : 'fas fa-eye'"></i>
            </button>
          </div>
        </div>
      </template>

      <!-- Error and Success Messages -->
      <div v-if="error" class="message error" role="alert">
        <i class="fas fa-exclamation-circle"></i>
        {{ error }}
      </div>
      <div v-if="successMessage" class="message success" role="alert">
        <i class="fas fa-check-circle"></i>
        {{ successMessage }}
      </div>

      <!-- Submit Button -->
      <button type="submit" :disabled="isProcessing" :class="{ processing: isProcessing }">
        <span class="button-content">
          <i
            :class="
              isProcessing
                ? 'fas fa-spinner fa-spin'
                : showChangePassword
                  ? 'fas fa-key'
                  : 'fas fa-sign-in-alt'
            "
          ></i>
          {{
            isProcessing
              ? $t('auth.processing')
              : showChangePassword
                ? $t('auth.changePassword')
                : $t('auth.login')
          }}
        </span>
      </button>

      <!-- Toggle Change Password Mode -->
      <button
        type="button"
        class="secondary-btn"
        @click="showChangePassword = !showChangePassword"
        :disabled="isProcessing"
      >
        <i :class="showChangePassword ? 'fas fa-arrow-left' : 'fas fa-key'"></i>
        {{ showChangePassword ? $t('auth.backToLogin') : $t('auth.changePassword') }}
      </button>
    </form>
    <div class="copyright">© Merhab Noureddine 2025</div>
  </div>
</template>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  position: relative;
}

.language-switcher-container {
  position: absolute;
  top: 20px;
  right: 20px;
  z-index: 1000;
}

/* RTL Support for login page */
[dir='rtl'] .language-switcher-container {
  right: auto;
  left: 20px;
}

.login-form {
  width: 100%;
  max-width: 400px;
  padding: 2rem;
  background: white;
  border-radius: 10px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

h2 {
  text-align: center;
  color: #2c3e50;
  margin-bottom: 1.5rem;
  font-size: 1.8rem;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
}

.form-group {
  margin-bottom: 1rem;
  position: relative;
}

.input-with-icon {
  position: relative;
  display: flex;
  align-items: center;
}

.input-with-icon i {
  position: absolute;
  left: 1rem;
  color: #7f8c8d;
  font-size: 1rem;
}

input {
  width: 100%;
  padding: 0.8rem 1rem 0.8rem 2.5rem;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  font-size: 1rem;
  transition: all 0.3s ease;
}

input:focus {
  border-color: #409eff;
  outline: none;
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
}

input:disabled {
  background-color: #f5f7fa;
  cursor: not-allowed;
}

.toggle-password {
  position: absolute;
  right: 1rem;
  background: none;
  border: none;
  color: #7f8c8d;
  cursor: pointer;
  padding: 0;
}

.toggle-password:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.message {
  padding: 0.8rem;
  margin-bottom: 1rem;
  border-radius: 4px;
  display: flex;
  align-items: center;
  font-size: 0.9rem;
}

.message i {
  margin-right: 0.5rem;
}

.error {
  background-color: #fef0f0;
  color: #f56c6c;
  border: 1px solid #fde2e2;
}

.success {
  background-color: #f0f9eb;
  color: #67c23a;
  border: 1px solid #e1f3d8;
}

button {
  width: 100%;
  padding: 0.8rem;
  border: none;
  border-radius: 4px;
  font-size: 1rem;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 1rem;
  background-color: #409eff;
  color: white;
}

button:not(:disabled):hover {
  background-color: #66b1ff;
  transform: translateY(-1px);
}

button:disabled {
  background-color: #a0cfff;
  cursor: not-allowed;
}

.processing {
  position: relative;
  pointer-events: none;
}

.button-content {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.secondary-btn {
  background-color: #909399;
}

.secondary-btn:not(:disabled):hover {
  background-color: #a6a9ad;
}

@media (max-width: 480px) {
  .login-form {
    margin: 1rem;
    padding: 1.5rem;
  }

  h2 {
    font-size: 1.5rem;
  }

  input {
    font-size: 16px; /* Prevents zoom on mobile */
  }
}

.copyright {
  text-align: center;
  color: #6b7280;
  font-size: 0.875rem;
  margin-top: 1rem;
  position: absolute;
  bottom: 1rem;
  width: 100%;
  left: 0;
}
</style>
