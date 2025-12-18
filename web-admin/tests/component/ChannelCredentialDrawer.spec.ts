import { describe, it, expect } from 'vitest'
import { mount, type VueWrapper } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import ChannelCredentialDrawer from '../../app/components/channels/ChannelCredentialDrawer.vue'
import { createEmptyCredentialPayload } from '../../app/types/channels'

const formStub = defineComponent({
  emits: ['submit'],
  setup(_, { slots, emit, attrs }) {
    return () =>
      h(
        'form',
        {
          ...attrs,
          onSubmit: (event: Event) => {
            event.preventDefault()
            emit('submit', event)
          },
        },
        slots.default?.(),
      )
  },
})

const fieldStub = defineComponent({
  setup(_, { slots }) {
    return () => h('label', {}, slots.default?.({ id: 'field' }))
  },
})

const inputStub = defineComponent({
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

const textareaStub = defineComponent({
  props: { modelValue: { type: String, default: '' } },
  emits: ['update:modelValue'],
  setup(props, { emit, attrs }) {
    return () =>
      h(
        'textarea',
        {
          ...attrs,
          value: props.modelValue,
          onInput: (event: Event) => {
            const target = event.target as HTMLTextAreaElement
            emit('update:modelValue', target.value)
          },
        },
        props.modelValue,
      )
  },
})

const selectStub = defineComponent({
  props: {
    modelValue: { type: String, default: '' },
    options: { type: Array, default: () => [] },
    valueAttribute: { type: String, default: 'value' },
    optionAttribute: { type: String, default: 'label' },
  },
  emits: ['update:modelValue'],
  setup(props, { emit, attrs }) {
    return () =>
      h(
        'select',
        {
          ...attrs,
          value: props.modelValue,
          onChange: (event: Event) => {
            const target = event.target as HTMLSelectElement
            emit('update:modelValue', target.value)
          },
        },
        (props.options as any[]).map((option) =>
          h(
            'option',
            {
              value: option[props.valueAttribute] ?? option,
            },
            option[props.optionAttribute] ?? option,
          ),
        ),
      )
  },
})

const buttonStub = defineComponent({
  props: {
    type: { type: String, default: 'button' },
  },
  setup(props, { slots, attrs }) {
    return () =>
      h(
        'button',
        {
          ...attrs,
          type: props.type,
        },
        slots.default?.(),
      )
  },
})

const buildWrapper = () =>
  mount(ChannelCredentialDrawer, {
    props: {
      modelValue: createEmptyCredentialPayload(),
      loading: false,
    },
    global: {
      stubs: {
        UForm: formStub,
        UFormField: fieldStub,
        UInput: inputStub,
        UTextarea: textareaStub,
        USelectMenu: selectStub,
        UButton: buttonStub,
      },
    },
  })

const submitForm = async (wrapper: VueWrapper) => {
  await wrapper.get('[data-testid="credential-form"]').trigger('submit')
}

describe('ChannelCredentialDrawer', () => {
  it('emits normalized scope and metadata when form submits successfully', async () => {
    const wrapper = buildWrapper()
    await wrapper.find('[data-testid="credential-scope"]').setValue('orders.read, inventory.write ')
    await wrapper.find('[data-testid="credential-payload"]').setValue('{"token":"abc"}')
    await wrapper.find('[data-testid="credential-metadata"]').setValue('{"note":"ok"}')

    await submitForm(wrapper)

    const emitted = wrapper.emitted('submit')
    expect(emitted).toBeTruthy()
    const payload = emitted![0][0]
    expect(payload.scope).toEqual(['orders.read', 'inventory.write'])
    expect(payload.metadata).toEqual({ note: 'ok' })
    expect(payload.payload).toEqual({ token: 'abc' })
  })

  it('requires attachment url when offline credential selected', async () => {
    const wrapper = buildWrapper()
    await wrapper.find('[data-testid="credential-type"]').setValue('offline')
    await wrapper.find('[data-testid="credential-payload"]').setValue('{"token":"offline"}')

    await submitForm(wrapper)
    expect(wrapper.emitted('submit')).toBeUndefined()

    await wrapper.find('[data-testid="credential-attachment"]').setValue('https://files.test/offline.pdf')
    await submitForm(wrapper)

    const emitted = wrapper.emitted('submit')
    expect(emitted).toBeTruthy()
    const payload = emitted![0][0]
    expect(payload.attachmentUrl).toEqual('https://files.test/offline.pdf')
  })
})
