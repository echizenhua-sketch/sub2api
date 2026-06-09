import { describe, it, expect, vi, beforeEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import ImportDataModal from '@/components/admin/account/ImportDataModal.vue'
import { adminAPI } from '@/api/admin'

const showError = vi.fn()
const showSuccess = vi.fn()

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess
  })
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      importData: vi.fn(),
      importCodexSession: vi.fn()
    }
  }
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
  })
}))

describe('ImportDataModal', () => {
  beforeEach(() => {
    showError.mockReset()
    showSuccess.mockReset()
    vi.mocked(adminAPI.accounts.importData).mockReset()
    vi.mocked(adminAPI.accounts.importCodexSession).mockReset()
  })

  it('未选择文件时提示错误', async () => {
    const wrapper = mount(ImportDataModal, {
      props: { show: true },
      global: {
        stubs: {
          BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' }
        }
      }
    })

    await wrapper.find('form').trigger('submit')
    expect(showError).toHaveBeenCalledWith('admin.accounts.dataImportSelectFile')
  })

  it('无效 JSON 时提示解析失败', async () => {
    const wrapper = mount(ImportDataModal, {
      props: { show: true },
      global: {
        stubs: {
          BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' }
        }
      }
    })

    const input = wrapper.find('input[type="file"]')
    const file = new File(['invalid json'], 'data.json', { type: 'application/json' })
    Object.defineProperty(file, 'text', {
      value: () => Promise.resolve('invalid json')
    })
    Object.defineProperty(input.element, 'files', {
      value: [file]
    })

    await input.trigger('change')
    await wrapper.find('form').trigger('submit')
    await Promise.resolve()

    expect(showError).toHaveBeenCalledWith('admin.accounts.dataImportParseFailed')
  })

  it('Codex token JSON 文件走 Codex session 导入接口', async () => {
    vi.mocked(adminAPI.accounts.importCodexSession).mockResolvedValue({
      total: 1,
      created: 1,
      updated: 0,
      skipped: 0,
      failed: 0,
      items: [{ index: 1, action: 'created', account_id: 1734, name: 'codex-account' }]
    })

    const wrapper = mount(ImportDataModal, {
      props: { show: true },
      global: {
        stubs: {
          BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' }
        }
      }
    })

    const content = JSON.stringify({
      type: 'codex',
      email: 'codex@example.com',
      token_source: 'codex_tokens',
      access_token: 'access-token',
      refresh_token: 'refresh-token',
      id_token: 'id-token',
      saved_at: '2026-06-09T13:50:37Z'
    })
    const input = wrapper.find('input[type="file"]')
    const file = new File([content], 'codex.json', { type: 'application/json' })
    Object.defineProperty(file, 'text', {
      value: () => Promise.resolve(content)
    })
    Object.defineProperty(input.element, 'files', {
      value: [file]
    })

    await input.trigger('change')
    await wrapper.find('form').trigger('submit')
    await Promise.resolve()

    expect(adminAPI.accounts.importCodexSession).toHaveBeenCalledWith({
      content,
      update_existing: true,
      skip_default_group_bind: true
    })
    expect(adminAPI.accounts.importData).not.toHaveBeenCalled()
    expect(showSuccess).toHaveBeenCalledWith('Codex session 导入完成：创建 1，更新 0，跳过 0，失败 0')
  })

  it('选择多个 JSON 文件时逐个解析并导入', async () => {
    vi.mocked(adminAPI.accounts.importData)
      .mockResolvedValueOnce({
        proxy_created: 0,
        proxy_reused: 0,
        proxy_failed: 0,
        account_created: 1,
        account_failed: 0
      })
      .mockResolvedValueOnce({
        proxy_created: 0,
        proxy_reused: 0,
        proxy_failed: 0,
        account_created: 2,
        account_failed: 0
      })

    const wrapper = mount(ImportDataModal, {
      props: { show: true },
      global: {
        stubs: {
          BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' }
        }
      }
    })

    const firstPayload = { accounts: [{ name: 'account-a' }], proxies: [] }
    const secondPayload = { accounts: [{ name: 'account-b' }, { name: 'account-c' }], proxies: [] }
    const firstContent = JSON.stringify(firstPayload)
    const secondContent = JSON.stringify(secondPayload)
    const firstFile = new File([firstContent], 'accounts-a.json', { type: 'application/json' })
    const secondFile = new File([secondContent], 'accounts-b.json', { type: 'application/json' })
    Object.defineProperty(firstFile, 'text', {
      value: () => Promise.resolve(firstContent)
    })
    Object.defineProperty(secondFile, 'text', {
      value: () => Promise.resolve(secondContent)
    })

    const input = wrapper.find('input[type="file"]')
    expect((input.element as HTMLInputElement).multiple).toBe(true)
    Object.defineProperty(input.element, 'files', {
      value: [firstFile, secondFile]
    })

    await input.trigger('change')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(adminAPI.accounts.importData).toHaveBeenCalledTimes(2)
    expect(adminAPI.accounts.importData).toHaveBeenNthCalledWith(1, {
      data: firstPayload,
      skip_default_group_bind: true
    })
    expect(adminAPI.accounts.importData).toHaveBeenNthCalledWith(2, {
      data: secondPayload,
      skip_default_group_bind: true
    })
    expect(showSuccess).toHaveBeenCalledWith('admin.accounts.dataImportSuccess')
  })

  it('选择多个 Codex token JSON 文件时逐个调用 Codex session 导入接口', async () => {
    vi.mocked(adminAPI.accounts.importCodexSession)
      .mockResolvedValueOnce({
        total: 1,
        created: 1,
        updated: 0,
        skipped: 0,
        failed: 0,
        items: [{ index: 1, action: 'created', account_id: 1734, name: 'codex-account-a' }]
      })
      .mockResolvedValueOnce({
        total: 1,
        created: 1,
        updated: 0,
        skipped: 0,
        failed: 0,
        items: [{ index: 1, action: 'created', account_id: 1735, name: 'codex-account-b' }]
      })

    const wrapper = mount(ImportDataModal, {
      props: { show: true },
      global: {
        stubs: {
          BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' }
        }
      }
    })

    const firstContent = JSON.stringify({
      type: 'codex',
      email: 'codex-a@example.com',
      token_source: 'codex_tokens',
      access_token: 'access-token-a',
      refresh_token: 'refresh-token-a',
      id_token: 'id-token-a'
    })
    const secondContent = JSON.stringify({
      type: 'codex',
      email: 'codex-b@example.com',
      token_source: 'codex_tokens',
      access_token: 'access-token-b',
      refresh_token: 'refresh-token-b',
      id_token: 'id-token-b'
    })
    const firstFile = new File([firstContent], 'codex-a.json', { type: 'application/json' })
    const secondFile = new File([secondContent], 'codex-b.json', { type: 'application/json' })
    Object.defineProperty(firstFile, 'text', {
      value: () => Promise.resolve(firstContent)
    })
    Object.defineProperty(secondFile, 'text', {
      value: () => Promise.resolve(secondContent)
    })

    const input = wrapper.find('input[type="file"]')
    expect((input.element as HTMLInputElement).multiple).toBe(true)
    Object.defineProperty(input.element, 'files', {
      value: [firstFile, secondFile]
    })

    await input.trigger('change')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(adminAPI.accounts.importCodexSession).toHaveBeenCalledTimes(2)
    expect(adminAPI.accounts.importCodexSession).toHaveBeenNthCalledWith(1, {
      content: firstContent,
      update_existing: true,
      skip_default_group_bind: true
    })
    expect(adminAPI.accounts.importCodexSession).toHaveBeenNthCalledWith(2, {
      content: secondContent,
      update_existing: true,
      skip_default_group_bind: true
    })
    expect(adminAPI.accounts.importData).not.toHaveBeenCalled()
    expect(showSuccess).toHaveBeenCalledWith(
      'codex-a.json: Codex session 导入完成：创建 1，更新 0，跳过 0，失败 0'
    )
    expect(showSuccess).toHaveBeenCalledWith(
      'codex-b.json: Codex session 导入完成：创建 1，更新 0，跳过 0，失败 0'
    )
  })
})
