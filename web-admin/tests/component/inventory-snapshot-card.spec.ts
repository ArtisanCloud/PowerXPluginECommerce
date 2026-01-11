import { describe, it, expect, vi } from 'vitest'
import { mount, type VueWrapper } from '@vue/test-utils'
import { defineComponent, h, nextTick } from 'vue'

const storeMock = {
  inventorySnapshots: {
    'sku-1': {
      warehouses: [],
      summary: {
        warehouseId: 'default',
        availableQty: 10,
        lockedQty: 0,
        inTransitQty: 0,
        safetyStock: 0,
      },
      isStale: false,
    },
  },
  fetchInventorySnapshot: vi.fn(async () => storeMock.inventorySnapshots['sku-1']),
  adjustInventorySnapshot: vi.fn(async () => storeMock.inventorySnapshots['sku-1']),
}

vi.mock('~/stores/productSku', () => ({
  useProductSkuStore: () => storeMock,
}))

import InventorySnapshotCard from '../../app/components/product/sku/InventorySnapshotCard.vue'

const cardStub = defineComponent({
  setup(_, { slots }) {
    return () => h('div', {}, slots.default?.())
  },
})

const fieldStub = defineComponent({
  setup(_, { slots }) {
    return () => h('label', {}, slots.default?.({ id: 'field' }))
  },
})

const inputStub = defineComponent({
  name: 'UInput',
  props: {
    modelValue: { type: [String, Number], default: '' },
    type: { type: String, default: 'text' },
  },
  emits: ['update:modelValue'],
  setup(props, { emit, attrs }) {
    return () =>
      h('input', {
        ...attrs,
        value: props.modelValue,
        type: props.type,
        onInput: (event: Event) => {
          const target = event.target as HTMLInputElement
          emit('update:modelValue', target.value)
        },
      })
  },
})

const buttonStub = defineComponent({
  name: 'UButton',
  props: {
    disabled: { type: Boolean, default: false },
    type: { type: String, default: 'button' },
  },
  setup(props, { slots, attrs }) {
    return () =>
      h(
        'button',
        {
          ...attrs,
          disabled: props.disabled,
          type: props.type,
        },
        slots.default?.(),
      )
  },
})

const skeletonStub = defineComponent({
  setup(_, { slots }) {
    return () => h('div', {}, slots.default?.())
  },
})

const alertStub = defineComponent({
  props: {
    title: { type: String, default: '' },
    description: { type: String, default: '' },
  },
  setup(props) {
    return () => h('div', { 'data-testid': 'alert' }, `${props.title}${props.description}`)
  },
})

const formStub = defineComponent({
  name: 'UForm',
  props: {
    id: { type: String, default: '' },
  },
  setup(props, { slots, attrs }) {
    return () => h('form', { ...attrs, id: props.id }, slots.default?.())
  },
})

const modalStub = defineComponent({
  name: 'UModal',
  props: {
    open: { type: Boolean, default: false },
  },
  setup(props, { slots }) {
    return () => (props.open ? h('div', {}, [slots.body?.(), slots.footer?.()]) : h('div'))
  },
})

const buildWrapper = (): VueWrapper =>
  mount(InventorySnapshotCard, {
    props: { skuId: 'sku-1' },
    global: {
      stubs: {
        UCard: cardStub,
        UModal: modalStub,
        UForm: formStub,
        UFormField: fieldStub,
        UInput: inputStub,
        UButton: buttonStub,
        USkeleton: skeletonStub,
        UAlert: alertStub,
        UBadge: true,
      },
      mocks: {
        $t: (key: string) => key,
      },
    },
  })

describe('InventorySnapshotCard', () => {
  it('loads snapshot on mount', async () => {
    buildWrapper()
    expect(storeMock.fetchInventorySnapshot).toHaveBeenCalledWith('sku-1')
  })

  it('submits delta adjustment', async () => {
    const wrapper = buildWrapper()

    await wrapper.find('[data-testid="inventory-open-adjust"]').trigger('click')
    await nextTick()
    wrapper.getComponent({ name: 'UInput' }).vm.$emit('update:modelValue', '10')
    await nextTick()
    await wrapper.find('[data-testid="inventory-apply"]').trigger('click')
    await nextTick()

    expect(storeMock.adjustInventorySnapshot).toHaveBeenCalledWith('sku-1', 10)
  })
})
