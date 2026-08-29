<template>
  <AppLayout v-if="isAuthenticated">
    <ModelDetailContent />
  </AppLayout>
  <div v-else class="min-h-screen bg-gray-50 dark:bg-dark-950">
    <PlazaNavBar />
    <main class="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8 lg:py-8">
      <ModelDetailContent />
    </main>
  </div>

  <PaymentQRDialog
    :show="qrDialog.show"
    :order-id="qrDialog.orderId"
    :qr-code="qrDialog.qrCode"
    :expires-at="qrDialog.expiresAt"
    payment-type="qris"
    @close="closeQRDialog"
    @success="handlePaymentSuccess"
  />
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { getModelPlaza, type ModelPlazaGroup, type PlazaModel } from '@/api/modelPlaza'
import { paymentAPI } from '@/api/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import PlazaNavBar from '@/components/modelPlaza/PlazaNavBar.vue'
import PaymentQRDialog from '@/components/payment/PaymentQRDialog.vue'
import SubscriptionPlanCard from '@/components/payment/SubscriptionPlanCard.vue'
import { buildCreateOrderPayload } from '@/components/payment/paymentFlow'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import { usePaymentStore } from '@/stores/payment'
import type { SubscriptionPlan } from '@/types/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { isMobileDevice } from '@/utils/device'
import { platformBadgeLightClass, platformLabel } from '@/utils/platformColors'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const authStore = useAuthStore()
const paymentStore = usePaymentStore()

const loading = ref(true)
const loadFailed = ref(false)
const submittingPlanId = ref<number | null>(null)
const model = ref<PlazaModel | null>(null)
const group = ref<ModelPlazaGroup | null>(null)
const plans = ref<SubscriptionPlan[]>([])
const isAuthenticated = computed(() => authStore.isAuthenticated)

const qrDialog = ref({
  show: false,
  orderId: 0,
  qrCode: '',
  expiresAt: '',
})

let pollTimer: ReturnType<typeof setInterval> | null = null

const decodedModelName = computed(() => {
  const rawName = Array.isArray(route.params.name) ? route.params.name[0] : route.params.name
  try {
    return decodeURIComponent(String(rawName || ''))
  } catch {
    return String(rawName || '')
  }
})

const planGridClass = computed(() => {
  if (plans.value.length <= 2) return 'grid grid-cols-1 gap-5 sm:grid-cols-2'
  return 'grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3'
})

function isUnauthorized(error: unknown): boolean {
  if (!error || typeof error !== 'object') return false
  const apiError = error as { status?: number; response?: { status?: number } }
  return apiError.status === 401 || apiError.response?.status === 401
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

function closeQRDialog() {
  stopPolling()
  qrDialog.value.show = false
}

async function handlePaymentSuccess() {
  closeQRDialog()
  await router.push('/subscriptions')
}

async function pollOrderStatus(orderId: number) {
  const order = await paymentStore.pollOrderStatus(orderId)
  if (!order) return

  const status = String(order.status || '').toUpperCase()
  if (status === 'COMPLETED' || status === 'PAID') {
    await handlePaymentSuccess()
  } else if (status === 'EXPIRED' || status === 'FAILED' || status === 'CANCELLED') {
    closeQRDialog()
    appStore.showError(t('payment.result.failed'))
  }
}

function startPolling(orderId: number) {
  stopPolling()
  pollTimer = setInterval(() => {
    void pollOrderStatus(orderId)
  }, 3000)
}

async function openCheckout(plan: SubscriptionPlan) {
  if (submittingPlanId.value !== null) return
  submittingPlanId.value = plan.id

  try {
    const result = await paymentStore.createOrder(buildCreateOrderPayload({
      amount: plan.price,
      paymentType: 'qris',
      orderType: 'subscription',
      planId: plan.id,
      isMobile: isMobileDevice(),
      isWechatBrowser: false,
    }))

    if (!result.qr_code) {
      appStore.showError(t('payment.result.failed'))
      return
    }

    qrDialog.value = {
      show: true,
      orderId: result.order_id,
      qrCode: result.qr_code,
      expiresAt: result.expires_at || '',
    }
    startPolling(result.order_id)
  } catch (error: unknown) {
    if (isUnauthorized(error)) {
      await router.push({ path: '/login', query: { redirect: route.fullPath } })
      return
    }
    appStore.showError(extractI18nErrorMessage(error, t, 'payment.errors', t('payment.result.failed')))
  } finally {
    submittingPlanId.value = null
  }
}

async function loadModelDetail() {
  loading.value = true
  loadFailed.value = false

  try {
    const [plazaResponse, checkoutResponse] = await Promise.all([
      getModelPlaza(),
      paymentAPI.getCheckoutInfo(),
    ])

    for (const candidateGroup of plazaResponse.groups) {
      const candidateModel = candidateGroup.models.find((item) => item.name === decodedModelName.value)
      if (candidateModel) {
        group.value = candidateGroup
        model.value = candidateModel
        break
      }
    }

    plans.value = model.value
      ? checkoutResponse.data.plans.filter((plan) => plan.for_sale && plan.product_name === model.value?.name)
      : []
  } catch (error: unknown) {
    if (isUnauthorized(error)) {
      await router.push({ path: '/login', query: { redirect: route.fullPath } })
      return
    }
    loadFailed.value = true
    appStore.showError(extractI18nErrorMessage(error, t, 'payment.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

const ModelDetailContent = defineComponent({
  name: 'ModelDetailContent',
  setup() {
    return () => h('div', { class: 'mx-auto max-w-5xl space-y-6' }, [
      loading.value
        ? h('div', { class: 'flex min-h-[240px] items-center justify-center' }, [
            h('div', { class: 'h-8 w-8 animate-spin rounded-full border-2 border-primary-600/25 border-t-primary-600 dark:border-primary-400/25 dark:border-t-primary-400' }),
          ])
        : loadFailed.value || !model.value || plans.value.length === 0
          ? h('div', { class: 'card px-6 py-16 text-center' }, [
              h(Icon, { name: 'cube', size: 'xl', class: 'mx-auto mb-4 text-gray-300 dark:text-dark-600' }),
              h('p', { class: 'text-sm text-gray-500 dark:text-dark-400' }, t('modelPlaza.noPlansForModel')),
              h('button', { class: 'btn btn-secondary mt-5', onClick: () => router.push('/model-plaza') }, t('modelPlaza.backToCatalog')),
            ])
          : h('div', { class: 'space-y-6' }, [
              h('header', { class: 'card overflow-hidden p-6 sm:p-8' }, [
                h('div', { class: 'flex flex-wrap items-center gap-2' }, [
                  h('span', { class: ['rounded-full px-2.5 py-1 text-xs font-medium', platformBadgeLightClass(model.value.platform || group.value?.platform || '')] }, platformLabel(model.value.platform || group.value?.platform || '')),
                  group.value ? h('span', { class: 'text-xs font-medium text-gray-400 dark:text-dark-500' }, group.value.name) : null,
                ]),
                h('h1', { class: 'mt-4 break-words text-2xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-3xl' }, model.value.name),
                group.value?.description ? h('p', { class: 'mt-2 max-w-3xl text-sm leading-relaxed text-gray-500 dark:text-dark-400' }, group.value.description) : null,
              ]),
              h('div', { class: planGridClass.value }, plans.value.map((plan) =>
                h(SubscriptionPlanCard, {
                  key: plan.id,
                  plan,
                  onSelect: openCheckout,
                })
              )),
            ]),
    ])
  },
})

onMounted(() => {
  void appStore.fetchPublicSettings()
  void loadModelDetail()
})

onUnmounted(stopPolling)
</script>
