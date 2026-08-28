<template>
  <div class="relative flex min-h-screen flex-col bg-gray-50 dark:bg-dark-950">
    <header class="relative z-20 border-b border-gray-200/70 px-6 py-4 dark:border-dark-800">
      <nav class="mx-auto flex max-w-6xl items-center justify-between">
        <router-link to="/home" class="flex items-center gap-3">
          <div class="h-10 w-10 overflow-hidden rounded-xl shadow-md">
            <img :src="siteLogo" alt="Logo" class="h-full w-full object-contain" />
          </div>
          <span class="text-lg font-semibold tracking-tight text-gray-900 dark:text-white">{{ siteName }}</span>
        </router-link>
        <div class="flex items-center gap-3">
          <LocaleSwitcher />
          <router-link to="/home" class="text-sm font-medium text-gray-500 hover:text-gray-900 dark:text-dark-400 dark:hover:text-white">Home</router-link>
        </div>
      </nav>
    </header>

    <main class="flex flex-1 items-center justify-center px-6 py-16">
      <section class="w-full max-w-2xl">
        <div class="mb-8 text-center">
          <p class="text-sm font-semibold uppercase tracking-[0.16em] text-primary-600 dark:text-primary-400">Support</p>
          <h1 class="mt-3 text-3xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-4xl">Contact support</h1>
          <p class="mt-3 text-gray-500 dark:text-dark-400">Tell us what you need help with and our team will get back to you.</p>
        </div>

        <div v-if="isSubmitted" class="rounded-2xl border border-green-200 bg-green-50 p-8 text-center dark:border-green-800/50 dark:bg-green-900/20">
          <Icon name="checkCircle" size="lg" class="mx-auto text-green-600 dark:text-green-400" />
          <h2 class="mt-4 text-xl font-semibold text-green-800 dark:text-green-200">Message sent</h2>
          <p class="mt-2 text-sm text-green-700 dark:text-green-300">Thanks for reaching out. We will respond as soon as possible.</p>
          <button type="button" class="btn btn-primary mt-6" @click="resetForm">Send another message</button>
        </div>

        <form v-else class="space-y-5 rounded-2xl border border-gray-200 bg-white p-6 shadow-sm dark:border-dark-800 dark:bg-dark-900 sm:p-8" @submit.prevent="handleSubmit">
          <div class="grid gap-5 sm:grid-cols-2">
            <div>
              <label for="contact-name" class="input-label">Name</label>
              <input id="contact-name" v-model.trim="form.name" class="input" required maxlength="200" autocomplete="name" :disabled="isLoading" />
            </div>
            <div>
              <label for="contact-email" class="input-label">Email</label>
              <input id="contact-email" v-model.trim="form.email" class="input" required type="email" autocomplete="email" :disabled="isLoading" />
            </div>
          </div>
          <div>
            <label for="contact-subject" class="input-label">Subject</label>
            <input id="contact-subject" v-model.trim="form.subject" class="input" required maxlength="200" :disabled="isLoading" />
          </div>
          <div>
            <label for="contact-message" class="input-label">Message</label>
            <textarea id="contact-message" v-model.trim="form.message" class="input min-h-40 resize-y" required minlength="10" maxlength="5000" :disabled="isLoading"></textarea>
          </div>
          <p v-if="errorMessage" role="alert" class="rounded-lg bg-red-50 px-4 py-3 text-sm text-red-700 dark:bg-red-900/20 dark:text-red-300">{{ errorMessage }}</p>
          <button type="submit" class="btn btn-primary w-full" :disabled="isLoading">
            {{ isLoading ? 'Sending…' : 'Send message' }}
          </button>
        </form>
      </section>
    </main>

    <footer class="border-t border-gray-200/70 px-6 py-6 text-center text-sm text-gray-500 dark:border-dark-800 dark:text-dark-400">
      <router-link to="/contact" class="font-medium text-primary-600 hover:text-primary-500 dark:text-primary-400">Contact support</router-link>
      <span class="mx-2">·</span>
      <span>&copy; {{ currentYear }} {{ siteName }}</span>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import { sendSupportContact } from '@/api'
import { useAppStore } from '@/stores'

const appStore = useAppStore()
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '/logo.svg')
const currentYear = new Date().getFullYear()
const isLoading = ref(false)
const isSubmitted = ref(false)
const errorMessage = ref('')
const form = reactive({ name: '', email: '', subject: '', message: '' })

async function handleSubmit(): Promise<void> {
  errorMessage.value = ''
  isLoading.value = true
  try {
    await sendSupportContact(form)
    isSubmitted.value = true
  } catch (error: unknown) {
    const candidate = error as { message?: unknown }
    errorMessage.value = typeof candidate.message === 'string' ? candidate.message : 'Unable to send your message. Please try again.'
  } finally {
    isLoading.value = false
  }
}

function resetForm(): void {
  form.name = ''
  form.email = ''
  form.subject = ''
  form.message = ''
  errorMessage.value = ''
  isSubmitted.value = false
}
</script>
