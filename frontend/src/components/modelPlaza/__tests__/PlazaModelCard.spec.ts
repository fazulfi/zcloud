import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import PlazaModelCard from '../PlazaModelCard.vue'
import type { PlazaModel } from '@/api/modelPlaza'

const routerPush = vi.fn()
vi.mock('vue-router', () => ({ useRouter: () => ({ push: routerPush }) }))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return { ...actual, useI18n: () => ({ t: (key: string) => key }) }
})

function tokenModel(overrides: Partial<PlazaModel> = {}): PlazaModel {
  return {
    name: '', platform: 'anthropic',
    pricing: { billing_mode: 'token', input_price: 3e-6, output_price: 1.5e-5, cache_write_price: 3.75e-6, cache_read_price: 3e-7, image_input_price: null, image_output_price: null, per_request_price: null, intervals: [] },
    official_pricing: { input_price: 3e-6, output_price: 1.5e-5, cache_write_price: 3.75e-6, cache_write_1h_price: 6e-6, cache_read_price: 3e-7 },
    ...overrides
  }
}

function mountCard(model = tokenModel(), props: Record<string, unknown> = {}) {
  return mount(PlazaModelCard, { props: { model, rateMultiplier: 1, ...props } })
}

describe('PlazaModelCard', () => {
  it('renders model name and buy button navigation', async () => {
    const wrapper = mountCard()
    expect(wrapper.text()).toContain('')
    await wrapper.find('button:last-child').trigger('click')
    expect(routerPush).toHaveBeenCalledWith('/model-plaza/model/')
  })

  it('renders official price and cache formatting (no paid-price panel)', () => {
    const text = mountCard().text()
    expect(text).toContain('$3.00')
    expect(text).toContain('$15.00')
    expect(text).toContain('$3.75')
    expect(text).toContain('$0.30')
    expect(text).toContain('$6.00')
    expect(text).not.toContain('modelPlaza.table.paidPrice')
    expect(text).not.toContain('$2.40')
  })

  it('applies custom rates in footer while keeping official prices unchanged', () => {
    const text = mountCard(tokenModel(), { rateMultiplier: 1, userRateMultiplier: 0.8 }).text()
    expect(text).toContain('$3.00')
    expect(text).toContain('$15.00')
    expect(text).toContain('0.8x')
    expect(mountCard(tokenModel(), { userRateMultiplier: 0.8 }).find('.line-through').text()).toBe('1x')
  })

  it('renders per-request and per-image pricing chips from official data', () => {
    const model = tokenModel({ pricing: { billing_mode: 'image', input_price: null, output_price: null, cache_write_price: null, cache_read_price: null, image_input_price: null, image_output_price: 3e-5, per_request_price: null, intervals: [] }, official_pricing: { input_price: null, output_price: null, cache_write_price: null, cache_read_price: null, cache_write_1h_price: null, intervals: [{ min_tokens: 0, max_tokens: null, tier_label: '1K', input_price: null, output_price: null, cache_write_price: null, cache_read_price: null, per_request_price: 0.01 }] } })
    const text = mountCard(model, { rateMultiplier: 0.1 }).text()
    expect(text).toContain('1K')
    expect(text).not.toContain('$0.000003')
  })

  it('renders tier labels and time-period row pricing from official data', () => {
    const model = tokenModel({ official_pricing: { ...tokenModel().official_pricing!, intervals: [{ min_tokens: 0, max_tokens: 200000, tier_label: '', input_price: 3e-6, output_price: 1.5e-5, cache_write_price: null, cache_read_price: null, per_request_price: null }, { min_tokens: 200000, max_tokens: null, tier_label: '', input_price: 6e-6, output_price: 3e-5, cache_write_price: null, cache_read_price: null, per_request_price: null }] }, time_pricing: { timezone: 'Asia/Shanghai', periods: [{ start_time: '00:30', end_time: '08:30:00', multiplier: 0.5 }] } })
    const base = mountCard(model, { rateMultiplier: 0.8 })
    expect(base.text()).toContain('≤200K')
    expect(base.text()).toContain('>200K')
    const period = mountCard(model, { rateMultiplier: 0.8, period: model.time_pricing!.periods[0] })
    expect(period.text()).toContain('00:30–08:30')
    expect(period.text()).toContain('0.4x')
  })
})
