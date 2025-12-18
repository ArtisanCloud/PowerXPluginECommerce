import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import ChannelHealthCard from '../../app/components/channels/ChannelHealthCard.vue'

const cardStub = defineComponent({
  setup(_, { slots }) {
    return () => h('div', {}, slots.default?.())
  },
})

const badgeStub = defineComponent({
  props: { color: { type: String, default: 'neutral' } },
  setup(props, { slots }) {
    return () =>
      h(
        'span',
        {
          'data-badge': props.color,
        },
        slots.default?.(),
      )
  },
})

describe('ChannelHealthCard', () => {
  it('renders score and labels', () => {
    const wrapper = mount(ChannelHealthCard, {
      props: {
        score: 82,
        labels: ['sync_unstable', 'inventory_low'],
      },
      global: {
        stubs: {
          UCard: cardStub,
          UBadge: badgeStub,
        },
      },
    })
    expect(wrapper.get('[data-testid="health-score"]').text()).toBe('82')
    const badges = wrapper.findAll('[data-testid="health-label"]')
    expect(badges).toHaveLength(2)
    expect(badges[0].text()).toContain('sync_unstable')
  })

  it('shows fallback when no labels', () => {
    const wrapper = mount(ChannelHealthCard, {
      props: { score: 30 },
      global: {
        stubs: {
          UCard: cardStub,
          UBadge: badgeStub,
        },
      },
    })
    expect(wrapper.text()).toContain('渠道运行健康')
  })
})
