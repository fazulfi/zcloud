import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import ModelDetailView from './ModelDetailView.vue'
import PaymentQRDialog from '@/components/payment/PaymentQRDialog.vue'
import SubscriptionPlanCard from '@/components/payment/SubscriptionPlanCard.vue'
import type { ModelPlazaResponse } from '@/api/modelPlaza'
import type { CheckoutInfoResponse, SubscriptionPlan } from '@/types/payment'

const routeState = vi.hoisted(() => ({
  params: { name: 'gpt-4o' } as Record<string, string>,
  fullPath: '/model-plaza/model/gpt-4o',
}))
const routerPush = vi.hoisted(() => vi.fn())
const getModelPlaza = vi.hoisted(() => vi.fn())
const getCheckoutInfo = vi.hoisted(() => vi.fn())
const createOrder = vi.hoisted(() => vi.fn())
const pollOrderStatus = vi.hoisted(() => vi.fn())
const showError = vi.hoisted(() => vi.fn())
const fetchPublicSettings = vi.hoisted(() => vi.fn())

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRoute: () => routeState,
    useRouter: () => ({ push: routerPush }),
  }
})

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

vi.mock('@/api/modelPlaza', () => ({ getModelPlaza }))
vi.mock('@/api/payment', () => ({ paymentAPI: { getCheckoutInfo } }))
vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({ isAuthenticated: true }),
}))
vi.mock('@/stores/payment', () => ({
  usePaymentStore: () => ({ createOrder, pollOrderStatus }),
}))
vi.mock('@/stores', () => ({
  useAppStore: () => ({ showError, fetchPublicSettings }),
}))
vi.mock('@/utils/device', () => ({ isMobileDevice: () => false }))
vi.mock('@/utils/platformColors', () => ({
  platformBadgeLightClass: () => 'platform-badge',
  platformLabel: (platform: string) => platform,
}))

const matchingPlan: SubscriptionPlan = {
  id: 7,
  group_id: 3,
  group_platform: 'openai',
  product_name: 'gpt-4o',
  name: 'GPT-4o Starter',
  description: 'Starter plan',
  price: 12,
  validity_days: 30,
  validity_unit: 'day',
  features: [],
  for_sale: true,
  sort_order: 1,
}

function plazaFixture(): ModelPlazaResponse {
  return {
    description: '',
    groups: [{
      id: 3,
      name: 'OpenAI',
      description: 'Fast general-purpose model',
      platform: 'openai',
      subscription_type: 'subscription',
      rate_multiplier: 1,
      peak_rate_enabled: false,
      peak_start: '',
      peak_end: '',
      peak_rate_multiplier: 1,
      is_exclusive: false,
      image_rate_independent: false,
      image_rate_multiplier: 1,
      long_context_pricing_enabled: false,
      models: [{
        name: 'gpt-4o',
        platform: 'openai',
        pricing: null,
        official_pricing: null,
      }],
    }],
  }
}

function checkoutFixture(plans: SubscriptionPlan[]): { data: CheckoutInfoResponse } {
  return {
    data: {
      methods: {},
      global_min: 0,
      global_max: 0,
      plans,
      balance_disabled: false,
      balance_recharge_multiplier: 1,
      subscription_usd_to_cny_rate: 0,
      recharge_fee_rate: 0,
      help_text: '',
      help_image_url: '',
      stripe_publishable_key: '',
    },
  }
}

async function mountView(plans: SubscriptionPlan[] = [matchingPlan]) {
  getModelPlaza.mockResolvedValue(plazaFixture())
  getCheckoutInfo.mockResolvedValue(checkoutFixture(plans))

  const wrapper = mount(ModelDetailView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        PlazaNavBar: true,
        PaymentQRDialog: true,
        SubscriptionPlanCard: true,
      },
    },
  })
  await flushPromises()
  await flushPromises()
  return wrapper
}

describe('ModelDetailView', () => {
  beforeEach(() => {
    vi.useRealTimers()
    routeState.params = { name: 'gpt-4o' }
    routeState.fullPath = '/model-plaza/model/gpt-4o'
    routerPush.mockReset().mockResolvedValue(undefined)
    getModelPlaza.mockReset()
    getCheckoutInfo.mockReset()
    createOrder.mockReset()
    pollOrderStatus.mockReset().mockResolvedValue(null)
    showError.mockReset()
    fetchPublicSettings.mockReset().mockResolvedValue(undefined)
  })

  it('filters sale plans by product_name matching the plaza model name', async () => {
    const hiddenPlan = { ...matchingPlan, id: 8, product_name: 'claude-3-5-sonnet' }
    const notForSale = { ...matchingPlan, id: 9, for_sale: false }
    const wrapper = await mountView([matchingPlan, hiddenPlan, notForSale])

    const cards = wrapper.findAllComponents(SubscriptionPlanCard)
    expect(cards).toHaveLength(1)
    expect(cards[0].props('plan')).toMatchObject({ id: 7, product_name: 'gpt-4o' })
  })

  it('creates a QRIS subscription order with the selected plan', async () => {
    createOrder.mockResolvedValue({
      order_id: 101,
      amount: matchingPlan.price,
      pay_amount: matchingPlan.price,
      fee_rate: 0,
      qr_code: 'qris://checkout/101',
      expires_at: '2099-01-01T00:10:00.000Z',
    })
    const wrapper = await mountView()

    await wrapper.getComponent(SubscriptionPlanCard).vm.$emit('select', matchingPlan)
    await flushPromises()

    expect(createOrder).toHaveBeenCalledWith({
      amount: matchingPlan.price,
      payment_type: 'qris',
      order_type: 'subscription',
      plan_id: matchingPlan.id,
      is_mobile: false,
      payment_source: 'hosted_redirect',
    })
  })

  it('opens PaymentQRDialog when createOrder returns a QR code', async () => {
    createOrder.mockResolvedValue({
      order_id: 101,
      amount: matchingPlan.price,
      pay_amount: matchingPlan.price,
      fee_rate: 0,
      qr_code: 'qris://checkout/101',
      expires_at: '2099-01-01T00:10:00.000Z',
    })
    const wrapper = await mountView()

    await wrapper.getComponent(SubscriptionPlanCard).vm.$emit('select', matchingPlan)
    await flushPromises()

    expect(wrapper.getComponent(PaymentQRDialog).props()).toMatchObject({
      show: true,
      orderId: 101,
      qrCode: 'qris://checkout/101',
      paymentType: 'qris',
    })
  })

  it('shows the empty state when no plans match the model', async () => {
    const wrapper = await mountView([{ ...matchingPlan, product_name: 'other-model' }])

    expect(wrapper.text()).toContain('modelPlaza.noPlansForModel')
    expect(wrapper.text()).toContain('modelPlaza.backToCatalog')
    expect(wrapper.findAllComponents(SubscriptionPlanCard)).toHaveLength(0)
  })

  it('redirects to login with the detail URL when checkout returns 401', async () => {
    getModelPlaza.mockResolvedValue(plazaFixture())
    getCheckoutInfo.mockRejectedValue({ status: 401 })

    mount(ModelDetailView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          PlazaNavBar: true,
          PaymentQRDialog: true,
        },
      },
    })
    await flushPromises()
    await flushPromises()

    expect(routerPush).toHaveBeenCalledWith({
      path: '/login',
      query: { redirect: '/model-plaza/model/gpt-4o' },
    })
  })
})
