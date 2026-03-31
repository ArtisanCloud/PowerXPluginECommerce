import { describe, it, expect } from 'vitest'
import { mount, type VueWrapper } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import ChannelForm from '../../app/components/channels/ChannelForm.vue'
import { createEmptyChannelPayload } from '../../app/types/channels'

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
  props: {
    label: { type: String, default: '' },
    error: { type: String, default: '' },
  },
  setup(props, { slots }) {
    return () =>
      h('div', { 'data-field-label': props.label }, [
        slots.default?.({ id: `field-${String(props.label || '').replace(/\s+/g, '-')}` }),
        props.error ? h('p', { class: 'field-error' }, props.error) : null,
      ])
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

const selectLikeStub = defineComponent({
  props: {
    modelValue: { type: [String, Object], default: '' },
    items: { type: Array, default: () => [] },
    valueAttribute: { type: String, default: '' },
    optionAttribute: { type: String, default: '' },
    valueKey: { type: String, default: '' },
    labelKey: { type: String, default: '' },
  },
  emits: ['update:modelValue'],
  setup(props, { emit, attrs }) {
    const readValue = (option: any) => {
      if (props.valueKey && option?.[props.valueKey] != null) return option[props.valueKey]
      if (props.valueAttribute && option?.[props.valueAttribute] != null) return option[props.valueAttribute]
      if (option?.value != null) return option.value
      return option
    }
    const readLabel = (option: any) => {
      if (props.labelKey && option?.[props.labelKey] != null) return option[props.labelKey]
      if (props.optionAttribute && option?.[props.optionAttribute] != null) return option[props.optionAttribute]
      if (option?.label != null) return option.label
      return option
    }
    return () =>
      h(
        'select',
        {
          ...attrs,
          value: typeof props.modelValue === 'object' ? readValue(props.modelValue) : (props.modelValue as any),
          onChange: (event: Event) => {
            const target = event.target as HTMLSelectElement
            emit('update:modelValue', target.value)
          },
        },
        (props.items as any[]).map((option) => h('option', { value: readValue(option) }, readLabel(option))),
      )
  },
})

const buildWrapper = () =>
  mount(ChannelForm, {
    props: {
      modelValue: createEmptyChannelPayload(),
      mode: 'create',
    },
    global: {
      stubs: {
        UForm: formStub,
        UFormField: fieldStub,
        UInput: inputStub,
        USelect: selectLikeStub,
        USelectMenu: selectLikeStub,
        UButton: buttonStub,
      },
    },
  })

const submitForm = async (wrapper: VueWrapper) => {
  await wrapper.get('form').trigger('submit')
}

describe('ChannelForm', () => {
  it('shows required validation errors when submitting empty form', async () => {
    const wrapper = buildWrapper()
    await submitForm(wrapper)

    expect(wrapper.emitted('submit')).toBeUndefined()
    expect(wrapper.text()).toContain('请填写渠道名称')
    expect(wrapper.text()).toContain('请选择平台')
    expect(wrapper.text()).toContain('请填写店铺 ID')
  })

  it('toggles offline evidence block when channel type switches to offline', async () => {
    const wrapper = buildWrapper()
    expect(wrapper.text()).not.toContain('将于 US2 阶段支持上传 PDF/图片')

    await wrapper.get('select[placeholder="选择类型"]').setValue('offline')
    expect(wrapper.text()).toContain('将于 US2 阶段支持上传 PDF/图片')
  })

  it('emits submit payload when required fields are completed', async () => {
    const wrapper = buildWrapper()
    await wrapper.get('input[placeholder="天猫旗舰店"]').setValue('官方旗舰店')
    await wrapper.get('select[placeholder="选择平台"]').setValue('tmall')
    await wrapper.get('input[placeholder="tmall-001"]').setValue('tmall-001')
    await wrapper.get('select[placeholder="选择类型"]').setValue('platform_oauth')
    await wrapper.get('select[placeholder="选择国家"]').setValue('CN')
    await wrapper.get('select[placeholder="选择城市"]').setValue('cn-beijing')
    await wrapper.get('select[placeholder="选择负责人"]').setValue('user:ops-01')
    await wrapper.get('input[placeholder="张敏"]').setValue('张敏')
    await wrapper.get('input[placeholder="+86 138****7788"]').setValue('+86 13812345678')
    await wrapper.get('input[placeholder="ops@example.com"]').setValue('ops@example.com')

    await submitForm(wrapper)

    const emitted = wrapper.emitted('submit')
    expect(emitted).toBeTruthy()
    const payload = emitted![0][0]
    expect(payload.name).toBe('官方旗舰店')
    expect(payload.platform).toBe('tmall')
    expect(payload.storeId).toBe('tmall-001')
    expect(payload.channelType).toBe('platform_oauth')
    expect(payload.region).toBe('cn-beijing')
    expect(payload.ownerUuid).toBe('user:ops-01')
    expect(payload.contact?.name).toBe('张敏')
  })
})
