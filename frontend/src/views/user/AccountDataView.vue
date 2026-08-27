<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="card p-6">
        <h1 class="text-xl font-semibold">Account data</h1>
        <p class="mt-2 text-sm text-gray-500">Download your customer-safe usage and invoice records.</p>
        <div class="mt-4 flex flex-wrap gap-3">
          <button class="btn-secondary" :disabled="loading" @click="download('usage', 'json')">Export usage JSON</button>
          <button class="btn-secondary" :disabled="loading" @click="download('usage', 'csv')">Export usage CSV</button>
          <button class="btn-secondary" :disabled="loading" @click="download('invoices', 'json')">Export invoices JSON</button>
          <button class="btn-secondary" :disabled="loading" @click="download('invoices', 'csv')">Export invoices CSV</button>
        </div>
      </div>
      <div class="card border-red-200 p-6 dark:border-red-900">
        <h2 class="text-lg font-semibold text-red-700">Delete account</h2>
        <p class="mt-2 text-sm text-gray-500">This revokes all API keys, stops subscriptions, and anonymizes your account. This cannot be undone.</p>
        <button class="btn-danger mt-4" :disabled="deleting" @click="removeAccount">{{ deleting ? 'Deleting…' : 'Delete my account' }}</button>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import { exportUsage, exportInvoices } from '@/api/usage'
import { deleteAccount } from '@/api/account'
import { useAuthStore } from '@/stores/auth'

const router = useRouter(); const authStore = useAuthStore(); const loading = ref(false); const deleting = ref(false)
const save = (blob: Blob, filename: string) => { const url = URL.createObjectURL(blob); const link = document.createElement('a'); link.href = url; link.download = filename; link.click(); URL.revokeObjectURL(url) }
const download = async (kind: 'usage' | 'invoices', format: 'csv' | 'json') => { loading.value = true; try { const blob = kind === 'usage' ? await exportUsage(format) : await exportInvoices(format); save(blob, `${kind}.${format}`) } finally { loading.value = false } }
const removeAccount = async () => { if (!window.confirm('Delete your account permanently?')) return; deleting.value = true; try { await deleteAccount(); await authStore.logout(); await router.push('/login') } finally { deleting.value = false } }
</script>
