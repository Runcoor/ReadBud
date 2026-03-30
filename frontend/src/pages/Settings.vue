<template>
  <div class="settings-page">
    <header class="settings-header">
      <div class="header-brand">
        <h1 class="header-title">阅芽</h1>
        <span class="header-divider">|</span>
        <span class="header-desc">系统设置</span>
      </div>
      <div class="header-actions">
        <el-button text @click="router.push({ name: 'Workbench' })">返回工作台</el-button>
      </div>
    </header>

    <main class="settings-main">
      <el-tabs v-model="activeTab" class="settings-tabs">
        <!-- Provider Config Tab -->
        <el-tab-pane label="服务配置" name="providers">
          <div class="tab-toolbar">
            <h3 class="tab-title">外部服务配置</h3>
            <el-button type="primary" size="small" @click="showProviderDialog = true">
              添加配置
            </el-button>
          </div>

          <el-table
            v-loading="providerLoading"
            :data="providers"
            stripe
            class="settings-table"
          >
            <el-table-column prop="provider_type" label="类型" width="140">
              <template #default="{ row }">
                <el-tag size="small" :type="getProviderTagType(row.provider_type)">
                  {{ getProviderLabel(row.provider_type) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="provider_name" label="名称" />
            <el-table-column label="密钥" width="100">
              <template #default="{ row }">
                <el-tag v-if="row.has_secret" size="small" type="success">已配置</el-tag>
                <el-tag v-else size="small" type="info">未配置</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
                  {{ row.status === 1 ? '启用' : '停用' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>

          <el-empty v-if="!providerLoading && providers.length === 0" description="暂无服务配置" />
        </el-tab-pane>

        <!-- WeChat Account Tab -->
        <el-tab-pane label="公众号管理" name="wechat">
          <div class="tab-toolbar">
            <h3 class="tab-title">公众号账号管理</h3>
            <el-button type="primary" size="small" @click="showWechatDialog = true">
              添加账号
            </el-button>
          </div>

          <el-table
            v-loading="wechatLoading"
            :data="wechatAccounts"
            stripe
            class="settings-table"
          >
            <el-table-column prop="name" label="名称" />
            <el-table-column prop="app_id" label="AppID" />
            <el-table-column label="令牌模式" width="140">
              <template #default="{ row }">
                {{ getTokenModeLabel(row.token_mode) }}
              </template>
            </el-table-column>
            <el-table-column label="默认" width="80">
              <template #default="{ row }">
                <el-tag v-if="row.is_default" size="small" type="warning">默认</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
                  {{ row.status === 1 ? '启用' : '停用' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>

          <el-empty v-if="!wechatLoading && wechatAccounts.length === 0" description="暂无公众号配置" />
        </el-tab-pane>
      </el-tabs>
    </main>

    <!-- Provider Dialog -->
    <el-dialog v-model="showProviderDialog" title="添加服务配置" width="500">
      <el-form :model="providerForm" label-position="top" @submit.prevent="handleCreateProvider">
        <el-form-item label="服务类型" required>
          <el-select v-model="providerForm.provider_type" placeholder="选择服务类型" style="width: 100%">
            <el-option
              v-for="(label, key) in PROVIDER_TYPE_LABELS"
              :key="key"
              :label="label"
              :value="key"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="名称" required>
          <el-input v-model="providerForm.provider_name" placeholder="例如：OpenAI GPT-4" />
        </el-form-item>
        <el-form-item label="配置 (JSON)">
          <el-input
            v-model="providerConfigStr"
            type="textarea"
            :rows="4"
            placeholder='{"model": "gpt-4", "base_url": "..."}'
          />
        </el-form-item>
        <el-form-item label="密钥">
          <el-input
            v-model="providerForm.secret_json"
            type="password"
            placeholder="API Key 或密钥 JSON"
            show-password
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showProviderDialog = false">取消</el-button>
        <el-button type="primary" :loading="providerSaving" @click="handleCreateProvider">保存</el-button>
      </template>
    </el-dialog>

    <!-- WeChat Dialog -->
    <el-dialog v-model="showWechatDialog" title="添加公众号" width="500">
      <el-form :model="wechatForm" label-position="top" @submit.prevent="handleCreateWechat">
        <el-form-item label="名称" required>
          <el-input v-model="wechatForm.name" placeholder="公众号名称" />
        </el-form-item>
        <el-form-item label="AppID" required>
          <el-input v-model="wechatForm.app_id" placeholder="公众号 AppID" />
        </el-form-item>
        <el-form-item label="AppSecret">
          <el-input v-model="wechatForm.app_secret" type="password" placeholder="公众号 AppSecret" show-password />
        </el-form-item>
        <el-form-item label="令牌模式" required>
          <el-select v-model="wechatForm.token_mode" style="width: 100%">
            <el-option
              v-for="(label, key) in TOKEN_MODE_LABELS"
              :key="key"
              :label="label"
              :value="key"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="设为默认">
          <el-switch v-model="wechatForm.is_default" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="wechatForm.remark" placeholder="备注信息" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showWechatDialog = false">取消</el-button>
        <el-button type="primary" :loading="wechatSaving" @click="handleCreateWechat">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { listProviders, createProvider, listWechatAccounts, createWechatAccount } from '@/api/provider'
import {
  PROVIDER_TYPE_LABELS,
  TOKEN_MODE_LABELS,
} from '@/types/provider'
import type {
  ProviderConfigVO,
  ProviderType,
  WechatAccountVO,
  TokenMode,
} from '@/types/provider'

const router = useRouter()
const activeTab = ref('providers')

// Provider state
const providers = ref<ProviderConfigVO[]>([])
const providerLoading = ref(false)
const providerSaving = ref(false)
const showProviderDialog = ref(false)
const providerConfigStr = ref('')
const providerForm = reactive({
  provider_type: 'llm' as ProviderType,
  provider_name: '',
  secret_json: '',
})

// WeChat state
const wechatAccounts = ref<WechatAccountVO[]>([])
const wechatLoading = ref(false)
const wechatSaving = ref(false)
const showWechatDialog = ref(false)
const wechatForm = reactive({
  name: '',
  app_id: '',
  app_secret: '',
  token_mode: 'direct' as TokenMode,
  is_default: false,
  remark: '',
})

function getProviderLabel(type_: string): string {
  return PROVIDER_TYPE_LABELS[type_ as ProviderType] || type_
}

function getProviderTagType(type_: string): '' | 'success' | 'warning' | 'info' | 'danger' {
  const map: Record<string, '' | 'success' | 'warning' | 'info' | 'danger'> = {
    llm: '', image_search: 'success', image_gen: 'warning',
    search: 'info', storage: 'danger', crawler: 'info',
  }
  return map[type_] || 'info'
}

function getTokenModeLabel(mode: string): string {
  return TOKEN_MODE_LABELS[mode as TokenMode] || mode
}

async function loadProviders() {
  providerLoading.value = true
  try {
    const resp = await listProviders()
    if (resp.code === 0) {
      providers.value = resp.data || []
    }
  } catch {
    // Error handled by interceptor
  } finally {
    providerLoading.value = false
  }
}

async function loadWechatAccounts() {
  wechatLoading.value = true
  try {
    const resp = await listWechatAccounts()
    if (resp.code === 0) {
      wechatAccounts.value = resp.data || []
    }
  } catch {
    // Error handled by interceptor
  } finally {
    wechatLoading.value = false
  }
}

async function handleCreateProvider() {
  if (!providerForm.provider_name) {
    ElMessage.warning('请输入名称')
    return
  }

  providerSaving.value = true
  try {
    let configJSON: Record<string, unknown> = {}
    if (providerConfigStr.value) {
      configJSON = JSON.parse(providerConfigStr.value) as Record<string, unknown>
    }
    const resp = await createProvider({
      provider_type: providerForm.provider_type,
      provider_name: providerForm.provider_name,
      config_json: configJSON,
      secret_json: providerForm.secret_json || undefined,
    })
    if (resp.code === 0) {
      ElMessage.success('配置已保存')
      showProviderDialog.value = false
      providerForm.provider_name = ''
      providerForm.secret_json = ''
      providerConfigStr.value = ''
      await loadProviders()
    }
  } catch {
    ElMessage.error('保存失败')
  } finally {
    providerSaving.value = false
  }
}

async function handleCreateWechat() {
  if (!wechatForm.name || !wechatForm.app_id) {
    ElMessage.warning('请填写必要信息')
    return
  }

  wechatSaving.value = true
  try {
    const resp = await createWechatAccount({
      name: wechatForm.name,
      app_id: wechatForm.app_id,
      app_secret: wechatForm.app_secret || undefined,
      token_mode: wechatForm.token_mode,
      is_default: wechatForm.is_default,
      remark: wechatForm.remark,
    })
    if (resp.code === 0) {
      ElMessage.success('公众号已保存')
      showWechatDialog.value = false
      wechatForm.name = ''
      wechatForm.app_id = ''
      wechatForm.app_secret = ''
      wechatForm.remark = ''
      await loadWechatAccounts()
    }
  } catch {
    ElMessage.error('保存失败')
  } finally {
    wechatSaving.value = false
  }
}

onMounted(() => {
  loadProviders()
  loadWechatAccounts()
})
</script>

<style lang="scss" scoped>
.settings-page {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background-color: $color-bg;
}

.settings-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 56px;
  padding: 0 $spacing-xl;
  background-color: $color-card-bg;
  border-bottom: 1px solid $color-border;
  box-shadow: $shadow-card;
}

.header-brand {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
}

.header-title {
  font-size: $font-size-lg;
  font-weight: $font-weight-bold;
  color: $color-primary;
}

.header-divider {
  color: $color-border;
}

.header-desc {
  font-size: $font-size-base;
  color: $color-text-secondary;
}

.settings-main {
  max-width: 960px;
  width: 100%;
  margin: 0 auto;
  padding: $spacing-xl;
}

.settings-tabs {
  background-color: $color-card-bg;
  border-radius: $radius-lg;
  border: 1px solid $color-border;
  padding: $spacing-lg;
}

.tab-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: $spacing-lg;
}

.tab-title {
  font-size: $font-size-md;
  font-weight: $font-weight-semibold;
  color: $color-text-primary;
}

.settings-table {
  width: 100%;
}

@media (max-width: $breakpoint-sm) {
  .settings-main {
    padding: $spacing-base;
  }
}
</style>
