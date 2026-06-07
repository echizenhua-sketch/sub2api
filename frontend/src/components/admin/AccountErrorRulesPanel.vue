<template>
  <div class="space-y-6">
    <div class="card">
      <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
          账号错误处理规则（全局）
        </h2>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
          配置上游错误码触发的“临时拉黑 + 轮询换号”规则。开启后，所有账号默认套用这些规则；
          单个账号可在账号编辑里覆盖。命中规则的请求会临时把该账号拉黑指定时长并切换到其它账号。
        </p>
      </div>

      <div class="p-6 space-y-6">
        <!-- 全局开关 -->
        <label class="flex items-center gap-3 cursor-pointer">
          <input v-model="enabled" type="checkbox" class="toggle-checkbox" />
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
            启用全局错误处理规则
          </span>
        </label>

        <!-- 规则列表 -->
        <div v-if="enabled" class="space-y-4">
          <div
            v-for="(rule, idx) in rules"
            :key="idx"
            class="rounded-xl border border-gray-200 p-4 dark:border-dark-700"
          >
            <div class="grid grid-cols-1 gap-4 md:grid-cols-12">
              <div class="md:col-span-2">
                <label class="input-label">错误码</label>
                <input
                  v-model.number="rule.error_code"
                  type="number"
                  min="100"
                  max="599"
                  class="input"
                  placeholder="502"
                />
              </div>
              <div class="md:col-span-2">
                <label class="input-label">拉黑时长(分钟)</label>
                <input
                  v-model.number="rule.duration_minutes"
                  type="number"
                  min="1"
                  class="input"
                  placeholder="5"
                />
              </div>
              <div class="md:col-span-5">
                <label class="input-label">关键词(可选，逗号分隔)</label>
                <input
                  v-model="rule._keywordsText"
                  type="text"
                  class="input"
                  placeholder="留空表示仅按错误码匹配"
                />
              </div>
              <div class="md:col-span-3">
                <label class="input-label">说明(可选)</label>
                <input v-model="rule.description" type="text" class="input" placeholder="备注" />
              </div>
            </div>
            <div class="mt-3 flex justify-end">
              <button type="button" class="btn btn-secondary btn-sm" @click="removeRule(idx)">
                删除
              </button>
            </div>
          </div>

          <button type="button" class="btn btn-secondary" @click="addRule">+ 添加规则</button>

          <p class="text-xs text-gray-500 dark:text-gray-400">
            说明：关键词留空时只要响应错误码匹配即触发；填写关键词时需错误码匹配且响应体包含任一关键词。
            拉黑时长至少 1 分钟；命中规则的账号会被临时拉黑该时长并触发换号。
          </p>
        </div>

        <div class="flex justify-end">
          <button type="button" class="btn btn-primary" :disabled="saving" @click="save">
            {{ saving ? '保存中...' : '保存规则' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import {
  getAccountTempUnschedulableRules,
  updateAccountTempUnschedulableRules,
} from '@/api/admin/settings'
import { useAppStore } from '@/stores/app'

interface EditableRule {
  error_code: number
  duration_minutes: number
  description: string
  _keywordsText: string
}

const appStore = useAppStore()
const enabled = ref(false)
const rules = ref<EditableRule[]>([])
const saving = ref(false)

function toEditable(r: any): EditableRule {
  return {
    error_code: Number(r.error_code) || 0,
    duration_minutes: Number(r.duration_minutes) || 0,
    description: r.description || '',
    _keywordsText: Array.isArray(r.keywords) ? r.keywords.join(', ') : '',
  }
}

function addRule(): void {
  rules.value.push({ error_code: 502, duration_minutes: 5, description: '', _keywordsText: '' })
}

function removeRule(idx: number): void {
  rules.value.splice(idx, 1)
}

async function load(): Promise<void> {
  try {
    const cfg = await getAccountTempUnschedulableRules()
    enabled.value = !!cfg.enabled
    rules.value = (cfg.rules || []).map(toEditable)
  } catch (e: any) {
    appStore.showError(e?.message || '加载规则失败')
  }
}

async function save(): Promise<void> {
  for (const r of rules.value) {
    if (r.error_code < 100 || r.error_code > 599) {
      appStore.showError('错误码必须在 100-599 之间')
      return
    }
    if (r.duration_minutes < 1) {
      appStore.showError('拉黑时长至少为 1 分钟')
      return
    }
  }
  saving.value = true
  try {
    const payload = {
      enabled: enabled.value,
      rules: rules.value.map((r) => ({
        error_code: r.error_code,
        duration_minutes: r.duration_minutes,
        description: r.description || undefined,
        keywords: r._keywordsText
          .split(',')
          .map((s) => s.trim())
          .filter(Boolean),
      })),
    }
    await updateAccountTempUnschedulableRules(payload)
    appStore.showSuccess('规则已保存')
  } catch (e: any) {
    appStore.showError(e?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>
