<template>
  <BaseDialog
    :show="show"
    title="批量导入 Kiro 账号"
    width="wide"
    close-on-click-outside
    @close="handleClose"
  >
    <form id="import-kiro-form" class="space-y-4" @submit.prevent="handleImport">
      <div class="text-sm text-gray-600 dark:text-dark-300">
        支持上传文件或直接粘贴。自动识别 Kiro manage 导出 JSON、Sub2API JSON 数组，或卡密格式
        <code>邮箱----密码----RefreshToken----ClientId----ClientSecret</code>（分隔符支持 ----、Tab、连续空格，每行一个账号）。
      </div>
      <div
        class="rounded-lg border border-amber-200 bg-amber-50 p-3 text-xs text-amber-600 dark:border-amber-800 dark:bg-amber-900/20 dark:text-amber-400"
      >
        导入时会用 RefreshToken 实时向 Kiro 验证凭据有效性，验证通过才创建账号。Github/Google 登录无需 ClientId/ClientSecret；Builder/IdC 登录必须提供 ClientId 和 ClientSecret。
      </div>

      <div class="flex items-center gap-4">
        <label class="input-label whitespace-nowrap">默认登录类型</label>
        <select v-model="defaultLoginType" class="input max-w-[12rem]">
          <option value="builder">builder（Builder ID / IdC）</option>
          <option value="idc">idc</option>
          <option value="github">github（社交）</option>
          <option value="google">google（社交）</option>
        </select>
        <span class="text-xs text-gray-500 dark:text-dark-400">
          条目未显式指定 login_type 时使用此默认值
        </span>
      </div>

      <div>
        <label class="input-label">上传文件</label>
        <div
          class="flex items-center justify-between gap-3 rounded-lg border border-dashed border-gray-300 bg-gray-50 px-4 py-3 dark:border-dark-600 dark:bg-dark-800"
        >
          <div class="min-w-0">
            <div class="truncate text-sm text-gray-700 dark:text-dark-200">
              {{ fileName || '可选择 .txt / .json / .csv 文件，内容将填入下方文本框' }}
            </div>
            <div class="text-xs text-gray-500 dark:text-dark-400">
              支持 Kiro manage 导出包、Sub2API JSON 数组、卡密格式
            </div>
          </div>
          <button type="button" class="btn btn-secondary shrink-0" @click="openFilePicker">
            选择文件
          </button>
        </div>
        <input
          ref="fileInput"
          type="file"
          class="hidden"
          accept=".txt,.json,.csv,text/plain,application/json,text/csv"
          @change="handleFileChange"
        />
      </div>

      <div>
        <label class="input-label">凭据数据</label>
        <textarea
          v-model="rawData"
          rows="10"
          class="input font-mono text-xs"
          placeholder="Kiro manage 导出包可直接上传：{&quot;accounts&quot;:[{&quot;email&quot;:&quot;a@b.com&quot;,&quot;machineId&quot;:&quot;...&quot;,&quot;credentials&quot;:{&quot;refreshToken&quot;:&quot;...&quot;,&quot;clientId&quot;:&quot;...&quot;,&quot;clientSecret&quot;:&quot;...&quot;,&quot;authMethod&quot;:&quot;IdC&quot;}}]}&#10;&#10;Sub2API JSON 数组：[{&quot;email&quot;:&quot;a@b.com&quot;,&quot;refresh_token&quot;:&quot;...&quot;,&quot;client_id&quot;:&quot;...&quot;,&quot;client_secret&quot;:&quot;...&quot;,&quot;login_type&quot;:&quot;builder&quot;}]&#10;&#10;卡密格式：user@example.com----password----RT_xxx----CID_xxx----CSEC_xxx"
        ></textarea>
        <p v-if="lineCount > 0" class="mt-1 text-xs text-gray-500 dark:text-dark-400">
          当前约 {{ lineCount }} 行待导入
        </p>
      </div>

      <div
        v-if="preview"
        class="space-y-3 rounded-xl border border-gray-200 p-4 text-sm dark:border-dark-700"
      >
        <div class="flex flex-wrap items-center justify-between gap-2">
          <div class="font-medium text-gray-900 dark:text-white">字段映射预览</div>
          <div class="text-xs text-gray-500 dark:text-dark-400">
            {{ preview.source }} · {{ preview.count }} 条
          </div>
        </div>
        <div class="grid gap-2 sm:grid-cols-2">
          <div
            v-for="field in preview.fields"
            :key="field.name"
            class="rounded-lg bg-gray-50 px-3 py-2 dark:bg-dark-800"
          >
            <div class="text-xs text-gray-500 dark:text-dark-400">{{ field.name }}</div>
            <div class="mt-1 truncate font-mono text-xs text-gray-800 dark:text-dark-100">
              {{ field.value }}
            </div>
          </div>
        </div>
        <div
          v-if="preview.extraFields.length"
          class="text-xs text-gray-500 dark:text-dark-400"
        >
          Kiro 独有字段将保留到 <code>extra.kiro_manage</code>：{{ preview.extraFields.join('、') }}
        </div>
      </div>

      <div
        v-if="result"
        class="space-y-2 rounded-xl border border-gray-200 p-4 dark:border-dark-700"
      >
        <div class="text-sm font-medium text-gray-900 dark:text-white">导入结果</div>
        <div class="text-sm text-gray-700 dark:text-dark-300">
          共 {{ result.total }} 条：成功 {{ result.created }}，跳过 {{ result.skipped }}，失败 {{ result.failed }}
        </div>
        <div v-if="errorItems.length" class="mt-2">
          <div class="text-sm font-medium text-red-600 dark:text-red-400">失败/跳过明细</div>
          <div
            class="mt-2 max-h-48 overflow-auto rounded-lg bg-gray-50 p-3 font-mono text-xs dark:bg-dark-800"
          >
            <div v-for="(item, idx) in errorItems" :key="idx" class="whitespace-pre-wrap">
              #{{ item.index + 1 }} {{ item.name || '-' }} — {{ item.message }}
            </div>
          </div>
        </div>
      </div>
    </form>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button class="btn btn-secondary" type="button" :disabled="importing" @click="handleClose">
          关闭
        </button>
        <button
          class="btn btn-primary"
          type="submit"
          form="import-kiro-form"
          :disabled="importing"
        >
          {{ importing ? '导入中...' : '开始导入' }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { adminAPI } from '@/api/admin'
import { useAppStore } from '@/stores/app'
import type { KiroImportResult } from '@/api/admin/accounts'

interface Props {
  show: boolean
}

interface Emits {
  (e: 'close'): void
  (e: 'imported'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const appStore = useAppStore()

const importing = ref(false)
const rawData = ref('')
const defaultLoginType = ref('builder')
const result = ref<KiroImportResult | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)
const fileName = ref('')

const errorItems = computed(() => result.value?.errors || [])

interface PreviewField {
  name: string
  value: string
}

interface ImportPreview {
  source: string
  count: number
  fields: PreviewField[]
  extraFields: string[]
}

const secretText = (value: unknown) => {
  if (value === null || value === undefined || String(value).trim() === '') return '未提供'
  return '已提供'
}

const textValue = (value: unknown) => {
  if (value === null || value === undefined || String(value).trim() === '') return '未提供'
  return String(value)
}

const normalizeLoginType = (value: unknown) => {
  const raw = String(value || '').trim().toLowerCase()
  if (!raw) return defaultLoginType.value || 'builder'
  if (raw === 'builderid' || raw === 'enterprise') return 'builder'
  if (raw === 'social') return 'github'
  return raw
}

const preview = computed<ImportPreview | null>(() => {
  const text = rawData.value.trim()
  if (!text || !(text.startsWith('{') || text.startsWith('['))) return null

  try {
    const parsed = JSON.parse(text)
    const accounts = Array.isArray(parsed)
      ? parsed
      : Array.isArray(parsed?.accounts)
        ? parsed.accounts
        : parsed && typeof parsed === 'object'
          ? [parsed]
          : []
    if (!accounts.length) return null

    const first = accounts[0] || {}
    const credentials = first.credentials || {}
    const source = !Array.isArray(parsed) && Array.isArray(parsed?.accounts)
      ? 'Kiro manage 导出包'
      : 'Sub2API JSON'
    const fields: PreviewField[] = [
      { name: 'email', value: textValue(first.email || first._email) },
      { name: 'access_token', value: secretText(first.access_token || first.accessToken || credentials.access_token || credentials.accessToken) },
      { name: 'refresh_token', value: secretText(first.refresh_token || first.refreshToken || credentials.refresh_token || credentials.refreshToken) },
      { name: 'client_id', value: secretText(first.client_id || first.clientId || credentials.client_id || credentials.clientId) },
      { name: 'client_secret', value: secretText(first.client_secret || first.clientSecret || credentials.client_secret || credentials.clientSecret) },
      { name: 'region', value: textValue(first.region || credentials.region || 'us-east-1') },
      { name: 'login_type', value: normalizeLoginType(first.login_type || first.loginType || first.authMethod || first.provider || credentials.login_type || credentials.loginType || credentials.authMethod || credentials.provider) },
      { name: 'machine_id', value: secretText(first.machine_id || first.machineId || credentials.machine_id || credentials.machineId) }
    ]
    const extraCandidates = [
      ['idp', first.idp],
      ['subscription', first.subscription],
      ['usage', first.usage],
      ['tags', first.tags],
      ['user_id', first.userId],
      ['created_at', first.createdAt],
      ['last_used_at', first.lastUsedAt],
      ['last_checked_at', first.lastCheckedAt],
      ['csrf_token', credentials.csrfToken],
      ['credentials_expires_at', credentials.expiresAt]
    ]
    const extraFields = extraCandidates
      .filter(([, value]) => value !== undefined && value !== null && String(value).trim() !== '')
      .map(([name]) => String(name))
    return {
      source,
      count: accounts.length,
      fields,
      extraFields
    }
  } catch {
    return null
  }
})

const lineCount = computed(() => {
  const text = rawData.value.trim()
  if (!text) return 0
  // JSON：尝试统计数组或 Kiro manage accounts 数量
  if (text.startsWith('[') || text.startsWith('{')) {
    try {
      const data = JSON.parse(text)
      if (Array.isArray(data)) return data.length
      if (Array.isArray(data?.accounts)) return data.accounts.length
      return 1
    } catch {
      return 0
    }
  }
  // 卡密格式：统计非空、非注释行
  return text
    .split('\n')
    .filter((line) => line.trim() && !line.trim().startsWith('#')).length
})

watch(
  () => props.show,
  (open) => {
    if (open) {
      rawData.value = ''
      result.value = null
      defaultLoginType.value = 'builder'
      fileName.value = ''
      if (fileInput.value) {
        fileInput.value.value = ''
      }
    }
  }
)

const openFilePicker = () => {
  fileInput.value?.click()
}

const readFileAsText = async (sourceFile: File): Promise<string> => {
  if (typeof sourceFile.text === 'function') {
    return sourceFile.text()
  }
  if (typeof sourceFile.arrayBuffer === 'function') {
    const buffer = await sourceFile.arrayBuffer()
    return new TextDecoder().decode(buffer)
  }
  return await new Promise<string>((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result ?? ''))
    reader.onerror = () => reject(reader.error || new Error('Failed to read file'))
    reader.readAsText(sourceFile)
  })
}

const handleFileChange = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const selected = target.files?.[0]
  if (!selected) return
  try {
    const text = await readFileAsText(selected)
    rawData.value = text
    fileName.value = selected.name
    result.value = null
  } catch (error: any) {
    appStore.showError(error?.message || '文件读取失败')
  }
}

const handleClose = () => {
  if (importing.value) return
  emit('close')
}

const handleImport = async () => {
  if (!rawData.value.trim()) {
    appStore.showError('请上传文件或粘贴 Kiro 凭据数据')
    return
  }

  importing.value = true
  try {
    const res = await adminAPI.accounts.importKiro({
      data: rawData.value,
      default_login_type: defaultLoginType.value,
      skip_default_group_bind: false
    })

    result.value = res

    if (res.failed > 0) {
      appStore.showError(`导入完成：成功 ${res.created}，跳过 ${res.skipped}，失败 ${res.failed}`)
    } else {
      appStore.showSuccess(`导入成功：新增 ${res.created} 个账号，跳过 ${res.skipped} 个`)
    }
    if (res.created > 0) {
      emit('imported')
    }
  } catch (error: any) {
    appStore.showError(error?.message || '导入失败')
  } finally {
    importing.value = false
  }
}
</script>
