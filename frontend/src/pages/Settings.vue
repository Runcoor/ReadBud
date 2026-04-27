<!--
  Copyright (C) 2026 Leazoot
  SPDX-License-Identifier: AGPL-3.0-or-later
  This file is part of ReadBud, licensed under the GNU AGPL v3.
  See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.
-->
<template>
  <div class="settings-page">
    <AppTopBar crumb="系统设置">
      <template #right>
        <button class="rail-btn" @click="router.push({ name: 'Workbench' })">← 返回工作台</button>
      </template>
    </AppTopBar>

    <div class="settings-shell">
      <!-- COL 1: Sidebar nav (220px) -->
      <aside class="nav-rail">
        <div class="nav-rail__label">SETTINGS</div>
        <nav class="nav-rail__list">
          <button
            v-for="item in navItems"
            :key="item.key"
            class="nav-item"
            :class="{
              'nav-item--active': !item.disabled && activeTab === item.tab,
              'nav-item--disabled': item.disabled,
            }"
            @click="onNavClick(item)"
          >
            <span class="nav-item__indicator" />
            <span class="nav-item__label">{{ item.label }}</span>
            <span v-if="item.disabled" class="nav-item__chip">SOON</span>
          </button>
        </nav>
      </aside>

      <!-- COL 2: List column (320px) -->
      <aside class="list-col">
        <!-- Provider list -->
        <template v-if="activeTab === 'providers'">
          <header class="list-col__header">
            <div class="list-col__head-row">
              <SectionLabel
                title="服务列表"
                :hint="providerSubtitle"
              />
              <span class="list-col__count">{{ providerCountLabel }}</span>
            </div>
          </header>

          <div v-if="providerLoading" class="list-col__state">
            <span class="mini-spinner"></span>
            <span>加载中…</span>
          </div>

          <div v-else-if="providerError" class="list-col__state list-col__state--error">
            <p>{{ providerError }}</p>
            <button class="text-link" @click="loadProviders">重试</button>
          </div>

          <div v-else class="list-col__items">
            <button
              v-for="p in providers"
              :key="p.id"
              class="row-item"
              :class="{ 'row-item--active': selectedProviderId === p.id }"
              @click="selectProvider(p)"
            >
              <div class="row-item__icon">{{ getProviderIcon(p.provider_type) }}</div>
              <div class="row-item__body">
                <div class="row-item__title-row">
                  <span class="row-item__name">{{ p.provider_name }}</span>
                  <MonoChip v-if="p.is_default" kind="sprout">DEFAULT</MonoChip>
                </div>
                <span class="row-item__sub">{{ getProviderLabel(p.provider_type) }}</span>
              </div>
              <StatusDot
                :kind="p.status === 1 ? 'sprout' : 'mute'"
                :size="6"
              />
            </button>

            <p v-if="providers.length === 0" class="list-col__empty-text">暂无服务配置</p>
          </div>

          <button class="add-cta" @click="startAddProvider">+ 添加服务</button>
        </template>

        <!-- WeChat account list -->
        <template v-else-if="activeTab === 'wechat'">
          <header class="list-col__header">
            <div class="list-col__head-row">
              <SectionLabel
                title="公众号列表"
                :hint="wechatSubtitle"
              />
              <span class="list-col__count">{{ wechatCountLabel }}</span>
            </div>
          </header>

          <div v-if="wechatLoading" class="list-col__state">
            <span class="mini-spinner"></span>
            <span>加载中…</span>
          </div>

          <div v-else-if="wechatError" class="list-col__state list-col__state--error">
            <p>{{ wechatError }}</p>
            <button class="text-link" @click="loadWechatAccounts">重试</button>
          </div>

          <div v-else class="list-col__items">
            <button
              v-for="acct in wechatAccounts"
              :key="acct.id"
              class="row-item"
              :class="{ 'row-item--active': selectedWechatId === acct.id }"
              @click="selectWechatAccount(acct)"
            >
              <div class="row-item__icon">WX</div>
              <div class="row-item__body">
                <div class="row-item__title-row">
                  <span class="row-item__name">{{ acct.name }}</span>
                  <MonoChip v-if="acct.is_default" kind="sprout">DEFAULT</MonoChip>
                </div>
                <span class="row-item__sub">{{ getTokenModeLabel(acct.token_mode) }}</span>
              </div>
              <StatusDot
                :kind="acct.status === 1 ? 'sprout' : 'mute'"
                :size="6"
              />
            </button>

            <p v-if="wechatAccounts.length === 0" class="list-col__empty-text">暂无公众号配置</p>
          </div>

          <button class="add-cta" @click="startAddWechat">+ 添加公众号</button>
        </template>
      </aside>

      <!-- COL 3: Detail / edit panel (flex) -->
      <section class="edit-col">
        <!-- ============== PROVIDER FLOW ============== -->
        <template v-if="activeTab === 'providers'">
          <!-- Empty state -->
          <div v-if="!editingProvider && !isAddingNew" class="empty-card">
            <div class="empty-card__code">EMPTY · 选择左侧服务查看配置</div>
          </div>

          <!-- Add new provider -->
          <template v-else-if="isAddingNew">
            <div class="edit-header">
              <div class="edit-header__title-row">
                <h2 class="edit-header__title">添加服务</h2>
              </div>
              <p class="edit-header__desc">
                选择服务类型并填入连接信息。完成后可设为默认或测试连接。
              </p>
            </div>

            <div class="form-card">
              <div class="form-grid">
                <div class="field">
                  <label class="field__label">服务类型</label>
                  <el-select v-model="newProviderType" placeholder="请选择服务类型" @change="onNewTypeChange">
                    <el-option
                      v-for="(label, key) in PROVIDER_TYPE_LABELS"
                      :key="key"
                      :value="key"
                      :label="label"
                    />
                  </el-select>
                </div>

                <template v-if="newProviderType">
                  <div class="field">
                    <label class="field__label">服务名称</label>
                    <el-input v-model="formFields.name" placeholder="例如：OpenAI GPT-4" />
                  </div>

                  <!-- LLM -->
                  <template v-if="newProviderType === 'llm'">
                    <div class="field">
                      <label class="field__label">API 格式</label>
                      <el-select v-model="formFields.api_format">
                        <el-option value="openai" label="OpenAI 兼容" />
                        <el-option value="anthropic" label="Anthropic" />
                      </el-select>
                    </div>
                    <div class="field">
                      <label class="field__label">API 地址</label>
                      <el-input v-model="formFields.base_url" placeholder="https://api.openai.com/v1" />
                    </div>
                    <div class="field">
                      <label class="field__label">模型名称</label>
                      <el-input v-model="formFields.model" placeholder="gpt-4 / claude-3-opus / deepseek-chat" />
                    </div>
                    <div class="field">
                      <label class="field__label">API Key</label>
                      <el-input
                        v-model="formFields.api_key"
                        :type="showApiKey ? 'text' : 'password'"
                        placeholder="sk-..."
                        class="field__mono"
                        show-password
                      />
                    </div>
                    <div class="field-row">
                      <div class="field">
                        <label class="field__label">温度 <span class="field__hint">{{ formFields.temperature }}</span></label>
                        <el-slider v-model="formFields.temperature" :min="0" :max="2" :step="0.1" />
                      </div>
                      <div class="field">
                        <label class="field__label">最大 Token 数</label>
                        <el-input-number v-model="formFields.max_tokens" :min="0" :controls="false" placeholder="4096" />
                      </div>
                    </div>
                  </template>

                  <!-- Image Gen -->
                  <template v-else-if="newProviderType === 'image_gen'">
                    <div class="field">
                      <label class="field__label">API 格式</label>
                      <el-select v-model="formFields.api_format">
                        <el-option value="openai" label="OpenAI 兼容" />
                        <el-option value="vertex_imagen" label="Vertex AI Imagen" />
                      </el-select>
                    </div>
                    <template v-if="formFields.api_format === 'openai'">
                      <div class="field">
                        <label class="field__label">API 地址</label>
                        <el-input v-model="formFields.base_url" placeholder="https://api.openai.com/v1" />
                      </div>
                      <div class="field">
                        <label class="field__label">模型名称</label>
                        <el-input v-model="formFields.model" placeholder="dall-e-3" />
                      </div>
                    </template>
                    <template v-else-if="formFields.api_format === 'vertex_imagen'">
                      <div class="field">
                        <label class="field__label">项目 ID</label>
                        <el-input v-model="formFields.project_id" placeholder="my-gcp-project" />
                      </div>
                      <div class="field">
                        <label class="field__label">区域</label>
                        <el-input v-model="formFields.region" placeholder="us-central1" />
                      </div>
                      <div class="field">
                        <label class="field__label">模型名称</label>
                        <el-input v-model="formFields.model" placeholder="imagegeneration@006" />
                      </div>
                    </template>
                    <div class="field">
                      <label class="field__label">API Key / 凭证</label>
                      <el-input
                        v-model="formFields.api_key"
                        :type="showApiKey ? 'text' : 'password'"
                        placeholder="API Key 或凭证 JSON"
                        class="field__mono"
                        show-password
                      />
                    </div>
                    <div class="field-row">
                      <div class="field">
                        <label class="field__label">图片宽度</label>
                        <el-input-number v-model="formFields.image_width" :min="0" :controls="false" placeholder="1024" />
                      </div>
                      <div class="field">
                        <label class="field__label">图片高度</label>
                        <el-input-number v-model="formFields.image_height" :min="0" :controls="false" placeholder="1024" />
                      </div>
                    </div>
                  </template>

                  <!-- Search -->
                  <template v-else-if="newProviderType === 'search'">
                    <div class="field">
                      <label class="field__label">搜索服务商</label>
                      <el-select v-model="formFields.search_provider">
                        <el-option value="google_custom" label="Google Custom Search" />
                        <el-option value="tavily" label="Tavily Search" />
                        <el-option value="serpapi" label="SerpAPI" />
                        <el-option value="bing" label="Bing Search" />
                      </el-select>
                    </div>
                    <template v-if="formFields.search_provider === 'google_custom'">
                      <div class="field">
                        <label class="field__label">搜索引擎 ID</label>
                        <el-input v-model="formFields.search_engine_id" placeholder="Google CSE 的搜索引擎 ID" />
                        <p class="field__hint-line">在 https://programmablesearchengine.google.com 创建获取</p>
                      </div>
                      <div class="field">
                        <label class="field__label">API Key</label>
                        <el-input
                          v-model="formFields.api_key"
                          :type="showApiKey ? 'text' : 'password'"
                          placeholder="Google API Key"
                          class="field__mono"
                          show-password
                        />
                        <p class="field__hint-line">在 Google Cloud Console → APIs & Services → Credentials 获取</p>
                      </div>
                    </template>
                    <template v-else>
                      <div class="field">
                        <label class="field__label">API Key</label>
                        <el-input
                          v-model="formFields.api_key"
                          :type="showApiKey ? 'text' : 'password'"
                          :placeholder="searchKeyPlaceholder"
                          class="field__mono"
                          show-password
                        />
                        <p class="field__hint-line">{{ searchKeyHint }}</p>
                      </div>
                    </template>
                  </template>

                  <!-- Image search -->
                  <template v-else-if="newProviderType === 'image_search'">
                    <div class="field">
                      <label class="field__label">图片搜索服务</label>
                      <el-select v-model="formFields.image_search_provider">
                        <el-option value="unsplash" label="Unsplash（免费）" />
                        <el-option value="pexels" label="Pexels（免费）" />
                        <el-option value="bing" label="Bing Image Search" />
                      </el-select>
                    </div>
                    <div class="field">
                      <label class="field__label">API Key</label>
                      <el-input
                        v-model="formFields.api_key"
                        :type="showApiKey ? 'text' : 'password'"
                        :placeholder="formFields.image_search_provider === 'unsplash' ? 'Access Key' : 'API Key'"
                        class="field__mono"
                        show-password
                      />
                      <p class="field__hint-line">{{ imageSearchHint }}</p>
                    </div>
                  </template>

                  <!-- Crawler -->
                  <template v-else-if="newProviderType === 'crawler'">
                    <div class="field">
                      <label class="field__label">抓取服务</label>
                      <el-select v-model="formFields.crawler_provider">
                        <el-option value="jina" label="Jina Reader（推荐，简单好用）" />
                        <el-option value="firecrawl" label="Firecrawl" />
                        <el-option value="custom" label="自定义接口" />
                      </el-select>
                    </div>
                    <template v-if="formFields.crawler_provider === 'custom'">
                      <div class="field">
                        <label class="field__label">API 地址</label>
                        <el-input v-model="formFields.base_url" placeholder="https://your-crawler-api.com" />
                      </div>
                    </template>
                    <div class="field">
                      <label class="field__label">{{ formFields.crawler_provider === 'custom' ? 'API Key（可选）' : 'API Key' }}</label>
                      <el-input
                        v-model="formFields.api_key"
                        :type="showApiKey ? 'text' : 'password'"
                        :placeholder="crawlerKeyPlaceholder"
                        class="field__mono"
                        show-password
                      />
                      <p v-if="crawlerKeyHint" class="field__hint-line">{{ crawlerKeyHint }}</p>
                    </div>
                    <div class="field">
                      <label class="field__label">超时时间（秒）</label>
                      <el-input-number v-model="formFields.crawler_timeout" :min="0" :controls="false" placeholder="30" />
                    </div>
                  </template>
                </template>
              </div>
            </div>

            <div class="action-bar">
              <button class="btn-ghost" @click="cancelAdd">取消</button>
              <div class="action-bar__right">
                <button class="btn-primary" :disabled="providerSaving" @click="handleCreateProvider">
                  <span v-if="providerSaving" class="mini-spinner mini-spinner--inverse"></span>
                  保存
                </button>
              </div>
            </div>
          </template>

          <!-- Edit existing provider -->
          <template v-else-if="editingProvider">
            <div class="edit-header">
              <div class="edit-header__title-row">
                <h2 class="edit-header__title">{{ editingProvider.provider_name }}</h2>
                <MonoChip>{{ getProviderLabel(editingProvider.provider_type) }}</MonoChip>
                <div class="edit-header__spacer" />
                <div class="edit-header__status">
                  <StatusDot :kind="editingProvider.status === 1 ? 'sprout' : 'mute'" :size="6" />
                  <span>{{ editingProvider.status === 1 ? '已连通' : '未启用' }}</span>
                </div>
              </div>
              <p class="edit-header__desc">
                配置 {{ getProviderLabel(editingProvider.provider_type) }} 服务的连接信息。可作为默认服务被流水线优先调用。
              </p>
            </div>

            <div class="form-card">
              <div class="form-grid">
                <div class="field">
                  <label class="field__label">服务名称</label>
                  <el-input v-model="formFields.name" />
                </div>

                <!-- LLM -->
                <template v-if="editingProvider.provider_type === 'llm'">
                  <div class="field">
                    <label class="field__label">API 格式</label>
                    <el-select v-model="formFields.api_format">
                      <el-option value="openai" label="OpenAI 兼容" />
                      <el-option value="anthropic" label="Anthropic" />
                    </el-select>
                  </div>
                  <div class="field">
                    <label class="field__label">API 地址</label>
                    <el-input v-model="formFields.base_url" placeholder="https://api.openai.com/v1" />
                  </div>
                  <div class="field">
                    <label class="field__label">模型名称</label>
                    <el-input v-model="formFields.model" placeholder="gpt-4" />
                  </div>
                  <div class="field">
                    <label class="field__label">API Key</label>
                    <el-input
                      v-model="formFields.api_key"
                      :type="showApiKey ? 'text' : 'password'"
                      :placeholder="editingProvider.has_secret ? '•••••• 已配置（留空保持不变）' : 'sk-...'"
                      class="field__mono"
                      show-password
                    />
                  </div>
                  <div class="field-row">
                    <div class="field">
                      <label class="field__label">温度 <span class="field__hint">{{ formFields.temperature }}</span></label>
                      <el-slider v-model="formFields.temperature" :min="0" :max="2" :step="0.1" />
                    </div>
                    <div class="field">
                      <label class="field__label">最大 Token 数</label>
                      <el-input-number v-model="formFields.max_tokens" :min="0" :controls="false" placeholder="4096" />
                    </div>
                  </div>
                </template>

                <!-- Image Gen -->
                <template v-else-if="editingProvider.provider_type === 'image_gen'">
                  <div class="field">
                    <label class="field__label">API 格式</label>
                    <el-select v-model="formFields.api_format">
                      <el-option value="openai" label="OpenAI 兼容" />
                      <el-option value="vertex_imagen" label="Vertex AI Imagen" />
                    </el-select>
                  </div>
                  <template v-if="formFields.api_format === 'openai'">
                    <div class="field">
                      <label class="field__label">API 地址</label>
                      <el-input v-model="formFields.base_url" placeholder="https://api.openai.com/v1" />
                    </div>
                    <div class="field">
                      <label class="field__label">模型名称</label>
                      <el-input v-model="formFields.model" placeholder="dall-e-3" />
                    </div>
                  </template>
                  <template v-else-if="formFields.api_format === 'vertex_imagen'">
                    <div class="field">
                      <label class="field__label">项目 ID</label>
                      <el-input v-model="formFields.project_id" />
                    </div>
                    <div class="field">
                      <label class="field__label">区域</label>
                      <el-input v-model="formFields.region" />
                    </div>
                    <div class="field">
                      <label class="field__label">模型名称</label>
                      <el-input v-model="formFields.model" />
                    </div>
                  </template>
                  <div class="field">
                    <label class="field__label">API Key</label>
                    <el-input
                      v-model="formFields.api_key"
                      :type="showApiKey ? 'text' : 'password'"
                      :placeholder="editingProvider.has_secret ? '•••••• 已配置（留空保持不变）' : 'API Key'"
                      class="field__mono"
                      show-password
                    />
                  </div>
                  <div class="field-row">
                    <div class="field">
                      <label class="field__label">图片宽度</label>
                      <el-input-number v-model="formFields.image_width" :min="0" :controls="false" placeholder="1024" />
                    </div>
                    <div class="field">
                      <label class="field__label">图片高度</label>
                      <el-input-number v-model="formFields.image_height" :min="0" :controls="false" placeholder="1024" />
                    </div>
                  </div>
                </template>

                <!-- Search -->
                <template v-else-if="editingProvider.provider_type === 'search'">
                  <div class="field">
                    <label class="field__label">搜索服务商</label>
                    <el-select v-model="formFields.search_provider">
                      <el-option value="google_custom" label="Google Custom Search" />
                      <el-option value="serpapi" label="SerpAPI" />
                      <el-option value="bing" label="Bing Search" />
                    </el-select>
                  </div>
                  <template v-if="formFields.search_provider === 'google_custom'">
                    <div class="field">
                      <label class="field__label">搜索引擎 ID</label>
                      <el-input v-model="formFields.search_engine_id" placeholder="Google CSE 搜索引擎 ID" />
                      <p class="field__hint-line">在 https://programmablesearchengine.google.com 创建获取</p>
                    </div>
                  </template>
                  <div class="field">
                    <label class="field__label">API Key</label>
                    <el-input
                      v-model="formFields.api_key"
                      :type="showApiKey ? 'text' : 'password'"
                      :placeholder="editingProvider.has_secret ? '•••••• 已配置（留空保持不变）' : 'API Key'"
                      class="field__mono"
                      show-password
                    />
                  </div>
                </template>

                <!-- Image search -->
                <template v-else-if="editingProvider.provider_type === 'image_search'">
                  <div class="field">
                    <label class="field__label">图片搜索服务</label>
                    <el-select v-model="formFields.image_search_provider">
                      <el-option value="unsplash" label="Unsplash（免费）" />
                      <el-option value="pexels" label="Pexels（免费）" />
                      <el-option value="bing" label="Bing Image Search" />
                    </el-select>
                  </div>
                  <div class="field">
                    <label class="field__label">API Key</label>
                    <el-input
                      v-model="formFields.api_key"
                      :type="showApiKey ? 'text' : 'password'"
                      :placeholder="editingProvider.has_secret ? '•••••• 已配置（留空保持不变）' : 'Access Key'"
                      class="field__mono"
                      show-password
                    />
                  </div>
                </template>

                <!-- Crawler -->
                <template v-else-if="editingProvider.provider_type === 'crawler'">
                  <div class="field">
                    <label class="field__label">抓取服务</label>
                    <el-select v-model="formFields.crawler_provider">
                      <el-option value="jina" label="Jina Reader" />
                      <el-option value="firecrawl" label="Firecrawl" />
                      <el-option value="custom" label="自定义接口" />
                    </el-select>
                  </div>
                  <template v-if="formFields.crawler_provider === 'custom'">
                    <div class="field">
                      <label class="field__label">API 地址</label>
                      <el-input v-model="formFields.base_url" placeholder="https://..." />
                    </div>
                  </template>
                  <div class="field">
                    <label class="field__label">API Key</label>
                    <el-input
                      v-model="formFields.api_key"
                      :type="showApiKey ? 'text' : 'password'"
                      :placeholder="editingProvider.has_secret ? '•••••• 已配置（留空保持不变）' : 'API Key'"
                      class="field__mono"
                      show-password
                    />
                  </div>
                  <div class="field">
                    <label class="field__label">超时时间（秒）</label>
                    <el-input-number v-model="formFields.crawler_timeout" :min="0" :controls="false" placeholder="30" />
                  </div>
                </template>

                <!-- Default switch -->
                <div class="field-switch">
                  <el-switch
                    :model-value="editingProvider.is_default"
                    :disabled="settingDefault || editingProvider.is_default"
                    @change="handleSetDefault"
                  />
                  <span class="field-switch__label">
                    设为「{{ getProviderLabel(editingProvider.provider_type) }}」类别的默认服务
                  </span>
                </div>
              </div>
            </div>

            <div class="action-bar">
              <button class="btn-danger-ghost" @click="handleDeleteProvider">删除服务</button>
              <div class="action-bar__right">
                <button class="btn-ghost" :disabled="testingConnection" @click="handleTestConnection">
                  <span v-if="testingConnection" class="mini-spinner"></span>
                  测试连接
                </button>
                <button class="btn-primary" :disabled="providerSaving" @click="handleUpdateProvider">
                  <span v-if="providerSaving" class="mini-spinner mini-spinner--inverse"></span>
                  保存
                </button>
              </div>
            </div>
          </template>
        </template>

        <!-- ============== EXTENSION FLOW ============== -->
        <template v-else-if="activeTab === 'extension'">
          <div class="edit-header">
            <div class="edit-header__title-row">
              <h2 class="edit-header__title">浏览器插件</h2>
              <MonoChip>EXTENSION</MonoChip>
            </div>
            <p class="edit-header__desc">
              个人公众号无法获得 draft/add API 权限。安装 ReadBud 浏览器插件后，
              点击「发布」会自动跳转到 mp.weixin.qq.com 编辑器并填好标题、正文、封面，
              你只需在 WeChat 编辑器里点「群发」即可。
            </p>
          </div>

          <div class="form-card form-card--guide">
            <h3 class="guide-h3">1 · 安装插件</h3>
            <ol class="guide-list">
              <li>下载 <code>wechat-extension</code> 目录（在 ReadBud 仓库根目录下）。</li>
              <li>打开 Chrome / Edge，进入 <code>chrome://extensions</code>。</li>
              <li>开启「开发者模式」（页面右上角开关）。</li>
              <li>点击「加载已解压的扩展程序」，选择 <code>wechat-extension</code> 文件夹。</li>
              <li>插件出现在工具栏，点击图标 → 粘贴下方签发的令牌 → 完成。</li>
            </ol>

            <h3 class="guide-h3">2 · 签发插件令牌</h3>
            <p class="guide-p">
              令牌是插件访问 ReadBud 数据的凭据。<strong>仅在签发那一刻显示一次，请立刻复制保存。</strong>
              丢失后只能重新签发。
            </p>

            <div class="token-issue-row">
              <el-input
                v-model="extensionTokenName"
                placeholder="备注名称（如：Chrome · 工作机）"
                style="max-width: 280px"
              />
              <button class="btn-primary" :disabled="extensionTokenIssuing" @click="handleIssueExtensionToken">
                <span v-if="extensionTokenIssuing" class="mini-spinner mini-spinner--inverse"></span>
                签发新令牌
              </button>
            </div>

            <div v-if="extensionTokenJustIssued" class="token-issued">
              <div class="token-issued__head">
                <span class="token-issued__label">新令牌（仅显示一次）</span>
                <button class="text-link" @click="copyExtensionToken">复制</button>
              </div>
              <code class="token-issued__value">{{ extensionTokenJustIssued }}</code>
            </div>

            <h3 class="guide-h3">3 · 已签发的令牌</h3>
            <div v-if="extensionTokensLoading" class="list-col__state">
              <span class="mini-spinner"></span>
              <span>加载中…</span>
            </div>
            <div v-else-if="extensionTokens.length === 0" class="guide-empty">
              尚未签发任何令牌
            </div>
            <ul v-else class="token-list">
              <li v-for="t in extensionTokens" :key="t.id" class="token-list__item">
                <div class="token-list__main">
                  <span class="token-list__name">{{ t.name }}</span>
                  <code class="token-list__prefix">{{ t.token_prefix }}…</code>
                </div>
                <div class="token-list__meta">
                  <span v-if="t.revoked_at" class="token-list__state token-list__state--revoked">已撤销</span>
                  <span v-else-if="t.last_used_at">最近使用 {{ formatTokenDate(t.last_used_at) }}</span>
                  <span v-else>从未使用</span>
                </div>
                <button
                  v-if="!t.revoked_at"
                  class="btn-danger-ghost"
                  @click="handleRevokeExtensionToken(t)"
                >撤销</button>
              </li>
            </ul>
          </div>
        </template>

        <!-- ============== WECHAT FLOW ============== -->
        <template v-else-if="activeTab === 'wechat'">
          <div v-if="!editingWechat && !isAddingWechat" class="empty-card">
            <div class="empty-card__code">EMPTY · 选择左侧公众号查看配置</div>
          </div>

          <template v-else-if="isAddingWechat || editingWechat">
            <div class="edit-header">
              <div class="edit-header__title-row">
                <h2 class="edit-header__title">
                  {{ isAddingWechat ? '添加公众号' : (editingWechat?.name || '') }}
                </h2>
                <MonoChip v-if="editingWechat">{{ getTokenModeLabel(editingWechat.token_mode) }}</MonoChip>
                <div class="edit-header__spacer" />
                <div v-if="editingWechat" class="edit-header__status">
                  <StatusDot :kind="editingWechat.status === 1 ? 'sprout' : 'mute'" :size="6" />
                  <span>{{ editingWechat.status === 1 ? '已启用' : '未启用' }}</span>
                </div>
              </div>
              <p class="edit-header__desc">
                配置微信公众号的 AppID 与 AppSecret，用于发布与素材上传。
              </p>
            </div>

            <div class="form-card">
              <div class="form-grid">
                <div class="field">
                  <label class="field__label">名称 <span class="field__required">*</span></label>
                  <el-input v-model="wechatForm.name" placeholder="公众号名称" />
                </div>
                <div class="field">
                  <label class="field__label">AppID <span class="field__required">*</span></label>
                  <el-input v-model="wechatForm.app_id" placeholder="公众号 AppID" class="field__mono" />
                </div>
                <div class="field">
                  <label class="field__label">AppSecret</label>
                  <el-input
                    v-model="wechatForm.app_secret"
                    :type="showWechatSecret ? 'text' : 'password'"
                    :placeholder="editingWechat ? '•••••• 已配置（留空保持不变）' : '公众号 AppSecret'"
                    class="field__mono"
                    show-password
                  />
                </div>
                <div class="field">
                  <label class="field__label">令牌模式 <span class="field__required">*</span></label>
                  <el-select v-model="wechatForm.token_mode">
                    <el-option
                      v-for="(label, key) in TOKEN_MODE_LABELS"
                      :key="key"
                      :value="key"
                      :label="label"
                    />
                  </el-select>
                </div>
                <div class="field">
                  <label class="field__label">发布方式 <span class="field__required">*</span></label>
                  <div class="delivery-grid">
                    <button
                      v-for="(label, key) in DELIVERY_MODE_LABELS"
                      :key="key"
                      type="button"
                      class="delivery-card"
                      :class="{ 'delivery-card--active': wechatForm.delivery_mode === key }"
                      @click="wechatForm.delivery_mode = key as DeliveryMode"
                    >
                      <span class="delivery-card__title">{{ label }}</span>
                      <span class="delivery-card__hint">{{ DELIVERY_MODE_HINTS[key as DeliveryMode] }}</span>
                    </button>
                  </div>
                  <p v-if="wechatForm.delivery_mode === 'api'" class="field__hint-line">
                    需要服务号 + 微信认证。个人订阅号会被 WeChat 拒绝（48001 api unauthorized）。
                  </p>
                  <p v-else-if="wechatForm.delivery_mode === 'extension'" class="field__hint-line">
                    点击「浏览器插件」面板查看安装步骤并签发插件令牌。
                  </p>
                </div>
                <div class="field">
                  <label class="field__label">备注</label>
                  <el-input
                    v-model="wechatForm.remark"
                    type="textarea"
                    :rows="3"
                    placeholder="备注信息"
                  />
                </div>
                <div class="field-switch">
                  <el-switch v-model="wechatForm.is_default" />
                  <span class="field-switch__label">设为默认公众号</span>
                </div>
              </div>
            </div>

            <div class="action-bar">
              <button v-if="editingWechat" class="btn-danger-ghost" @click="handleDeleteWechat">删除公众号</button>
              <button v-else class="btn-ghost" @click="cancelWechatEdit">取消</button>
              <div class="action-bar__right">
                <button class="btn-primary" :disabled="wechatSaving" @click="handleSaveWechat">
                  <span v-if="wechatSaving" class="mini-spinner mini-spinner--inverse"></span>
                  保存
                </button>
              </div>
            </div>
          </template>
        </template>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listProviders, createProvider, listWechatAccounts, createWechatAccount } from '@/api/provider'
import { patch, del, post } from '@/api/request'
import {
  PROVIDER_TYPE_LABELS,
  TOKEN_MODE_LABELS,
  DELIVERY_MODE_LABELS,
  DELIVERY_MODE_HINTS,
} from '@/types/provider'
import type {
  ProviderConfigVO,
  ProviderType,
  WechatAccountVO,
  TokenMode,
  DeliveryMode,
  ExtensionTokenVO,
} from '@/types/provider'
import {
  listExtensionTokens,
  issueExtensionToken,
  revokeExtensionToken,
} from '@/api/extension'
import AppTopBar from '@/components/common/AppTopBar.vue'
import StatusDot from '@/components/common/StatusDot.vue'
import MonoChip from '@/components/common/MonoChip.vue'
import SectionLabel from '@/components/common/SectionLabel.vue'

const router = useRouter()
const activeTab = ref<'providers' | 'wechat' | 'extension'>('providers')

// ---- Sidebar nav items ----
interface NavItem {
  key: string
  label: string
  tab?: 'providers' | 'wechat' | 'extension'
  disabled?: boolean
}

const navItems: NavItem[] = [
  { key: 'providers', label: '服务配置', tab: 'providers' },
  { key: 'wechat', label: '公众号管理', tab: 'wechat' },
  { key: 'extension', label: '浏览器插件', tab: 'extension' },
  { key: 'templates', label: '模板与风格', disabled: true },
  { key: 'team', label: '账户与团队', disabled: true },
  { key: 'billing', label: '计费 · 用量', disabled: true },
  { key: 'notify', label: '通知偏好', disabled: true },
]

function onNavClick(item: NavItem) {
  if (item.disabled) {
    ElMessage.info('即将上线')
    return
  }
  if (item.tab) {
    activeTab.value = item.tab
  }
}

// ---- Provider state ----
const providers = ref<ProviderConfigVO[]>([])
const providerLoading = ref(false)
const providerSaving = ref(false)
const providerError = ref<string | null>(null)
const selectedProviderId = ref<string | null>(null)
const editingProvider = ref<ProviderConfigVO | null>(null)
const isAddingNew = ref(false)
const newProviderType = ref<ProviderType | ''>('')
const showApiKey = ref(false)
const testingConnection = ref(false)
const settingDefault = ref(false)

interface FormFields {
  name: string
  api_format: string
  base_url: string
  model: string
  api_key: string
  temperature: number
  max_tokens: number | null
  project_id: string
  region: string
  image_width: number | null
  image_height: number | null
  // Search (Google Custom Search, SerpAPI, etc.)
  search_engine_id: string
  search_provider: string  // 'google_custom' | 'serpapi' | 'bing'
  // Image search
  image_search_provider: string  // 'unsplash' | 'pexels' | 'bing'
  // Crawler
  crawler_provider: string  // 'jina' | 'firecrawl' | 'custom'
  crawler_timeout: number | null
}

const formFields = reactive<FormFields>({
  name: '',
  api_format: 'openai',
  base_url: '',
  model: '',
  api_key: '',
  search_engine_id: '',
  search_provider: 'google_custom',
  image_search_provider: 'unsplash',
  crawler_provider: 'jina',
  crawler_timeout: null,
  temperature: 0.7,
  max_tokens: null,
  project_id: '',
  region: '',
  image_width: null,
  image_height: null,
})

function resetFormFields() {
  formFields.name = ''
  formFields.api_format = 'openai'
  formFields.base_url = ''
  formFields.model = ''
  formFields.api_key = ''
  formFields.temperature = 0.7
  formFields.max_tokens = null
  formFields.project_id = ''
  formFields.region = ''
  formFields.image_width = null
  formFields.image_height = null
  formFields.search_engine_id = ''
  formFields.search_provider = 'google_custom'
  formFields.image_search_provider = 'unsplash'
  formFields.crawler_provider = 'jina'
  formFields.crawler_timeout = null
  showApiKey.value = false
}

function populateFormFromProvider(p: ProviderConfigVO) {
  resetFormFields()
  formFields.name = p.provider_name
  const cfg = p.config_json || {}
  formFields.api_format = (cfg.api_format as string) || (cfg.format as string) || 'openai'
  formFields.base_url = (cfg.base_url as string) || ''
  formFields.model = (cfg.model as string) || ''
  formFields.temperature = (cfg.temperature as number) ?? 0.7
  formFields.max_tokens = (cfg.max_tokens as number) || null
  formFields.project_id = (cfg.project_id as string) || ''
  formFields.region = (cfg.region as string) || ''
  formFields.image_width = (cfg.image_width as number) || null
  formFields.image_height = (cfg.image_height as number) || null
  // Search
  formFields.search_provider = (cfg.search_provider as string) || 'google_custom'
  formFields.search_engine_id = (cfg.search_engine_id as string) || ''
  // Image search
  formFields.image_search_provider = (cfg.image_search_provider as string) || 'unsplash'
  // Crawler
  formFields.crawler_provider = (cfg.crawler_provider as string) || 'jina'
  formFields.crawler_timeout = (cfg.timeout as number) || null
}

function buildConfigJson(providerType: string): Record<string, unknown> {
  const cfg: Record<string, unknown> = {}

  if (providerType === 'llm') {
    cfg.api_format = formFields.api_format
    if (formFields.base_url) cfg.base_url = formFields.base_url
    if (formFields.model) cfg.model = formFields.model
    cfg.temperature = formFields.temperature
    if (formFields.max_tokens) cfg.max_tokens = formFields.max_tokens
  } else if (providerType === 'image_gen') {
    cfg.api_format = formFields.api_format
    if (formFields.api_format === 'vertex_imagen') {
      if (formFields.project_id) cfg.project_id = formFields.project_id
      if (formFields.region) cfg.region = formFields.region
    } else {
      if (formFields.base_url) cfg.base_url = formFields.base_url
    }
    if (formFields.model) cfg.model = formFields.model
    if (formFields.image_width) cfg.image_width = formFields.image_width
    if (formFields.image_height) cfg.image_height = formFields.image_height
  } else if (providerType === 'search') {
    cfg.search_provider = formFields.search_provider
    if (formFields.search_provider === 'google_custom') {
      if (formFields.search_engine_id) cfg.search_engine_id = formFields.search_engine_id
      if (formFields.project_id) cfg.project_id = formFields.project_id
    }
    if (formFields.base_url) cfg.base_url = formFields.base_url
  } else if (providerType === 'image_search') {
    cfg.image_search_provider = formFields.image_search_provider
    if (formFields.base_url) cfg.base_url = formFields.base_url
  } else if (providerType === 'crawler') {
    cfg.crawler_provider = formFields.crawler_provider
    if (formFields.base_url) cfg.base_url = formFields.base_url
    if (formFields.crawler_timeout) cfg.timeout = formFields.crawler_timeout
  }

  return cfg
}

// Build secret_json based on provider type
function buildSecretJson(_providerType: string): string | undefined {
  return formFields.api_key || undefined
}

function selectProvider(p: ProviderConfigVO) {
  isAddingNew.value = false
  selectedProviderId.value = p.id
  editingProvider.value = p
  populateFormFromProvider(p)
}

function startAddProvider() {
  editingProvider.value = null
  selectedProviderId.value = null
  isAddingNew.value = true
  newProviderType.value = ''
  resetFormFields()
}

function cancelAdd() {
  isAddingNew.value = false
}

function onNewTypeChange() {
  resetFormFields()
}

function getProviderLabel(type_: string): string {
  return PROVIDER_TYPE_LABELS[type_ as ProviderType] || type_
}

const PROVIDER_TYPE_CODE: Record<string, string> = {
  llm: 'AI',
  image_search: 'IS',
  image_gen: 'IG',
  search: 'SE',
  storage: 'ST',
  crawler: 'CR',
}

function getProviderIcon(type_: string): string {
  return PROVIDER_TYPE_CODE[type_] || '??'
}

function getTokenModeLabel(mode: string): string {
  return TOKEN_MODE_LABELS[mode as TokenMode] || mode
}

// ---- Computed list metrics ----
const providerCountLabel = computed(() => {
  const n = providers.value.length
  return n.toString().padStart(2, '0')
})

const providerSubtitle = computed(() => {
  const total = providers.value.length
  const defaults = providers.value.filter(p => p.is_default).length
  if (total === 0) return '尚未接入任何服务'
  return `${total} 个已接入 · ${defaults} 个为默认`
})

const wechatCountLabel = computed(() => {
  const n = wechatAccounts.value.length
  return n.toString().padStart(2, '0')
})

const wechatSubtitle = computed(() => {
  const total = wechatAccounts.value.length
  const defaults = wechatAccounts.value.filter(a => a.is_default).length
  if (total === 0) return '尚未添加公众号'
  return `${total} 个已绑定 · ${defaults} 个为默认`
})

// ---- Search hint helpers ----
const searchKeyPlaceholder = computed(() => {
  switch (formFields.search_provider) {
    case 'tavily': return 'tvly-xxxx'
    case 'serpapi': return 'SerpAPI Key'
    case 'bing': return 'Bing Search API Key'
    default: return 'API Key'
  }
})

const searchKeyHint = computed(() => {
  switch (formFields.search_provider) {
    case 'tavily': return '在 https://app.tavily.com 注册获取，免费 1000 次/月'
    case 'serpapi': return '在 https://serpapi.com/dashboard 获取，免费 100 次/月'
    case 'bing': return '在 Azure Portal → Cognitive Services → Bing Search 获取'
    default: return ''
  }
})

const imageSearchHint = computed(() => {
  switch (formFields.image_search_provider) {
    case 'unsplash': return '在 https://unsplash.com/developers 注册应用获取'
    case 'pexels': return '在 https://www.pexels.com/api/ 注册获取'
    default: return '在 Azure Portal → Bing Image Search 获取'
  }
})

const crawlerKeyPlaceholder = computed(() => {
  switch (formFields.crawler_provider) {
    case 'jina': return 'jina_xxxx'
    case 'firecrawl': return 'fc-xxxx'
    default: return '留空则不传认证'
  }
})

const crawlerKeyHint = computed(() => {
  switch (formFields.crawler_provider) {
    case 'jina': return '在 https://jina.ai/reader 获取，有免费额度'
    case 'firecrawl': return '在 https://firecrawl.dev 注册获取'
    default: return ''
  }
})

// ---- API calls ----
async function loadProviders() {
  providerLoading.value = true
  providerError.value = null
  try {
    const resp = await listProviders()
    if (resp.code === 0) {
      providers.value = resp.data || []
    }
  } catch (e: unknown) {
    providerError.value = e instanceof Error ? e.message : '加载服务配置失败'
  } finally {
    providerLoading.value = false
  }
}

async function handleCreateProvider() {
  if (!newProviderType.value) {
    ElMessage.warning('请选择服务类型')
    return
  }
  if (!formFields.name) {
    ElMessage.warning('请输入服务名称')
    return
  }

  providerSaving.value = true
  try {
    const cfg = buildConfigJson(newProviderType.value)
    const resp = await createProvider({
      provider_type: newProviderType.value as ProviderType,
      provider_name: formFields.name,
      config_json: cfg,
      secret_json: buildSecretJson(newProviderType.value),
    })
    if (resp.code === 0) {
      ElMessage.success('服务已创建')
      isAddingNew.value = false
      resetFormFields()
      await loadProviders()
      // Auto-select newly created provider
      if (resp.data) {
        selectProvider(resp.data)
      }
    }
  } catch {
    ElMessage.error('创建失败')
  } finally {
    providerSaving.value = false
  }
}

async function handleUpdateProvider() {
  if (!editingProvider.value) return
  if (!formFields.name) {
    ElMessage.warning('请输入服务名称')
    return
  }

  providerSaving.value = true
  try {
    const cfg = buildConfigJson(editingProvider.value.provider_type)
    const payload: Record<string, unknown> = {
      provider_type: editingProvider.value.provider_type,
      provider_name: formFields.name,
      config_json: cfg,
    }
    const secretVal = buildSecretJson(editingProvider.value.provider_type)
    if (secretVal) {
      payload.secret_json = secretVal
    }
    await patch(`/providers/${editingProvider.value.id}`, payload)
    ElMessage.success('已保存')
    await loadProviders()
    // Re-select to refresh
    const updated = providers.value.find(p => p.id === editingProvider.value?.id)
    if (updated) selectProvider(updated)
  } catch {
    ElMessage.error('保存失败')
  } finally {
    providerSaving.value = false
  }
}

async function handleDeleteProvider() {
  if (!editingProvider.value) return

  try {
    await ElMessageBox.confirm(
      `确定要删除「${editingProvider.value.provider_name}」吗？此操作不可恢复。`,
      '确认删除',
      { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' },
    )
  } catch {
    return // cancelled
  }

  try {
    await del(`/providers/${editingProvider.value.id}`)
    ElMessage.success('已删除')
    editingProvider.value = null
    selectedProviderId.value = null
    await loadProviders()
  } catch {
    ElMessage.error('删除失败')
  }
}

async function handleTestConnection() {
  if (!editingProvider.value) return
  testingConnection.value = true
  try {
    const resp = await post(`/providers/${editingProvider.value.id}/test`) as { code: number; message?: string }
    if (resp.code === 0) {
      ElMessage.success('连接成功')
    } else {
      ElMessage.warning(resp.message || '连接测试未通过')
    }
  } catch {
    ElMessage.error('连接测试失败')
  } finally {
    testingConnection.value = false
  }
}

async function handleSetDefault() {
  if (!editingProvider.value) return
  if (editingProvider.value.is_default) return
  settingDefault.value = true
  try {
    await post(`/providers/${editingProvider.value.id}/set-default`)
    ElMessage.success('已设为默认')
    await loadProviders()
    const updated = providers.value.find(p => p.id === editingProvider.value?.id)
    if (updated) selectProvider(updated)
  } catch {
    ElMessage.error('设置失败')
  } finally {
    settingDefault.value = false
  }
}

// ---- WeChat state ----
const wechatAccounts = ref<WechatAccountVO[]>([])
const wechatLoading = ref(false)
const wechatSaving = ref(false)
const wechatError = ref<string | null>(null)
const showWechatSecret = ref(false)
const selectedWechatId = ref<string | null>(null)
const editingWechat = ref<WechatAccountVO | null>(null)
const isAddingWechat = ref(false)
const wechatForm = reactive({
  name: '',
  app_id: '',
  app_secret: '',
  token_mode: 'direct' as TokenMode,
  delivery_mode: 'extension' as DeliveryMode,
  is_default: false,
  remark: '',
})

function resetWechatForm() {
  wechatForm.name = ''
  wechatForm.app_id = ''
  wechatForm.app_secret = ''
  wechatForm.token_mode = 'direct'
  wechatForm.delivery_mode = 'extension'
  wechatForm.is_default = false
  wechatForm.remark = ''
  showWechatSecret.value = false
}

function selectWechatAccount(acct: WechatAccountVO) {
  isAddingWechat.value = false
  selectedWechatId.value = acct.id
  editingWechat.value = acct
  wechatForm.name = acct.name
  wechatForm.app_id = acct.app_id
  wechatForm.app_secret = ''
  wechatForm.token_mode = acct.token_mode
  wechatForm.delivery_mode = acct.delivery_mode || 'extension'
  wechatForm.is_default = acct.is_default
  wechatForm.remark = acct.remark || ''
}

function startAddWechat() {
  editingWechat.value = null
  selectedWechatId.value = null
  isAddingWechat.value = true
  resetWechatForm()
}

function cancelWechatEdit() {
  isAddingWechat.value = false
  editingWechat.value = null
  selectedWechatId.value = null
  resetWechatForm()
}

async function loadWechatAccounts() {
  wechatLoading.value = true
  wechatError.value = null
  try {
    const resp = await listWechatAccounts()
    if (resp.code === 0) {
      wechatAccounts.value = resp.data || []
    }
  } catch (e: unknown) {
    wechatError.value = e instanceof Error ? e.message : '加载公众号配置失败'
  } finally {
    wechatLoading.value = false
  }
}

async function handleSaveWechat() {
  if (!wechatForm.name || !wechatForm.app_id) {
    ElMessage.warning('请填写必要信息')
    return
  }

  wechatSaving.value = true
  try {
    if (editingWechat.value) {
      // Update existing
      const payload: Record<string, unknown> = {
        name: wechatForm.name,
        app_id: wechatForm.app_id,
        token_mode: wechatForm.token_mode,
        delivery_mode: wechatForm.delivery_mode,
        is_default: wechatForm.is_default,
        remark: wechatForm.remark,
      }
      if (wechatForm.app_secret) {
        payload.app_secret = wechatForm.app_secret
      }
      await patch(`/wechat-accounts/${editingWechat.value.id}`, payload)
      ElMessage.success('已保存')
      await loadWechatAccounts()
      const updated = wechatAccounts.value.find(a => a.id === editingWechat.value?.id)
      if (updated) selectWechatAccount(updated)
    } else {
      // Create new
      const resp = await createWechatAccount({
        name: wechatForm.name,
        app_id: wechatForm.app_id,
        app_secret: wechatForm.app_secret || undefined,
        token_mode: wechatForm.token_mode,
        delivery_mode: wechatForm.delivery_mode,
        is_default: wechatForm.is_default,
        remark: wechatForm.remark,
      })
      if (resp.code === 0) {
        ElMessage.success('公众号已保存')
        isAddingWechat.value = false
        resetWechatForm()
        await loadWechatAccounts()
        if (resp.data) {
          selectWechatAccount(resp.data)
        }
      }
    }
  } catch {
    ElMessage.error('保存失败')
  } finally {
    wechatSaving.value = false
  }
}

async function handleDeleteWechat() {
  if (!editingWechat.value) return

  try {
    await ElMessageBox.confirm(
      `确定要删除「${editingWechat.value.name}」吗？此操作不可恢复。`,
      '确认删除',
      { confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning' },
    )
  } catch {
    return
  }

  try {
    await del(`/wechat-accounts/${editingWechat.value.id}`)
    ElMessage.success('已删除')
    editingWechat.value = null
    selectedWechatId.value = null
    resetWechatForm()
    await loadWechatAccounts()
  } catch {
    ElMessage.error('删除失败')
  }
}

// ---- Extension token state ----
const extensionTokens = ref<ExtensionTokenVO[]>([])
const extensionTokensLoading = ref(false)
const extensionTokenName = ref('')
const extensionTokenIssuing = ref(false)
const extensionTokenJustIssued = ref<string | null>(null)

async function loadExtensionTokens() {
  extensionTokensLoading.value = true
  try {
    const resp = await listExtensionTokens()
    if (resp.code === 0) {
      extensionTokens.value = resp.data || []
    }
  } catch {
    ElMessage.error('加载插件令牌失败')
  } finally {
    extensionTokensLoading.value = false
  }
}

async function handleIssueExtensionToken() {
  extensionTokenIssuing.value = true
  try {
    const resp = await issueExtensionToken({
      name: extensionTokenName.value || undefined,
    })
    if (resp.code === 0 && resp.data) {
      extensionTokenJustIssued.value = resp.data.token
      extensionTokenName.value = ''
      ElMessage.success('令牌已签发,请立刻复制保存')
      await loadExtensionTokens()
    }
  } catch {
    ElMessage.error('签发失败')
  } finally {
    extensionTokenIssuing.value = false
  }
}

async function copyExtensionToken() {
  if (!extensionTokenJustIssued.value) return
  try {
    await navigator.clipboard.writeText(extensionTokenJustIssued.value)
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.warning('复制失败,请手动复制')
  }
}

async function handleRevokeExtensionToken(t: ExtensionTokenVO) {
  try {
    await ElMessageBox.confirm(
      `撤销「${t.name}」后，使用该令牌的浏览器插件将立即失效。继续吗？`,
      '撤销令牌',
      { confirmButtonText: '撤销', cancelButtonText: '取消', type: 'warning' },
    )
  } catch {
    return
  }
  try {
    await revokeExtensionToken(t.id)
    ElMessage.success('已撤销')
    await loadExtensionTokens()
  } catch {
    ElMessage.error('撤销失败')
  }
}

function formatTokenDate(s?: string): string {
  if (!s) return ''
  try {
    return new Date(s).toLocaleString('zh-CN', {
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    })
  } catch {
    return s
  }
}

onMounted(() => {
  loadProviders()
  loadWechatAccounts()
  loadExtensionTokens()
})
</script>

<style lang="scss" scoped>
@use '@/styles/tokens' as *;

// ============================================================
// Settings page — 阅芽 优化方案
// ============================================================

.settings-page {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background: var(--brand-paper);
  font-family: var(--font-sans);
  color: var(--text-primary);
}

.rail-btn {
  background: transparent;
  border: none;
  padding: 0 4px;
  font-family: var(--font-mono);
  font-size: 11px;
  letter-spacing: 0.06em;
  color: var(--text-tertiary);
  cursor: pointer;
  transition: color 120ms ease;

  &:hover {
    color: var(--text-primary);
  }
}

// ===== Three-column shell =====
.settings-shell {
  flex: 1;
  display: flex;
  min-height: 0;
}

// ===== COL 1: Sidebar nav =====
.nav-rail {
  width: 220px;
  flex-shrink: 0;
  border-right: 1px solid var(--border-hair);
  background: var(--surface-card);
  padding: 24px 16px;
  display: flex;
  flex-direction: column;

  &__label {
    font-size: 10px;
    color: var(--text-tertiary);
    font-family: var(--font-mono);
    letter-spacing: 0.15em;
    padding: 0 8px;
    margin-bottom: 12px;
  }

  &__list {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
}

.nav-item {
  position: relative;
  height: 34px;
  padding: 0 12px 0 14px;
  background: transparent;
  border: none;
  border-radius: 4px;
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text-body);
  font-size: 13px;
  cursor: pointer;
  transition: background 120ms ease, color 120ms ease;
  text-align: left;

  &__indicator {
    position: absolute;
    left: 0;
    top: 50%;
    transform: translateY(-50%);
    width: 2px;
    height: 16px;
    background: transparent;
    border-radius: 1px;
  }

  &__label {
    flex: 1;
    line-height: 1;
  }

  &__chip {
    font-family: var(--font-mono);
    font-size: 9px;
    letter-spacing: 0.1em;
    color: var(--text-faint);
    padding: 2px 5px;
    border: 1px solid var(--border-hair-soft);
    border-radius: 3px;
  }

  &:hover:not(&--disabled):not(&--active) {
    background: var(--border-hair-soft);
    color: var(--text-primary);
  }

  &--active {
    background: var(--border-hair-soft);
    color: var(--text-primary);
    font-weight: 500;

    .nav-item__indicator {
      background: var(--brand-ink);
    }
  }

  &--disabled {
    color: var(--text-faint);
    cursor: default;

    &:hover {
      background: transparent;
    }
  }
}

// ===== COL 2: List column =====
.list-col {
  width: 320px;
  flex-shrink: 0;
  border-right: 1px solid var(--border-hair);
  background: var(--brand-paper);
  display: flex;
  flex-direction: column;
  overflow: hidden;

  &__header {
    padding: 22px 22px 4px;
  }

  &__head-row {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;

    :deep(.section-label) {
      margin-bottom: 0;
      flex: 1;
    }
  }

  &__count {
    font-family: var(--font-mono);
    font-size: 13px;
    color: var(--text-tertiary);
    letter-spacing: 0.06em;
    padding-top: 1px;
  }

  &__items {
    flex: 1;
    overflow-y: auto;
    padding: 8px 12px 16px;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  &__state {
    padding: 24px 22px;
    color: var(--text-tertiary);
    font-size: 12px;
    display: flex;
    align-items: center;
    gap: 8px;

    &--error {
      flex-direction: column;
      align-items: flex-start;
      gap: 8px;

      p {
        margin: 0;
        color: var(--brand-danger);
      }
    }
  }

  &__empty-text {
    margin: 24px 0;
    text-align: center;
    color: var(--text-faint);
    font-size: 12px;
  }
}

.row-item {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  background: transparent;
  border: 1px solid transparent;
  border-radius: 4px;
  padding: 12px 14px 12px 12px;
  cursor: pointer;
  text-align: left;
  position: relative;
  transition: background 120ms ease, border-color 120ms ease;

  &__icon {
    width: 30px;
    height: 30px;
    flex-shrink: 0;
    border: 1px solid var(--border-hair);
    border-radius: 4px;
    display: grid;
    place-items: center;
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.06em;
    color: var(--text-tertiary);
    background: var(--surface-card);
  }

  &__body {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 3px;
  }

  &__title-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  &__name {
    font-size: 13px;
    font-weight: 500;
    color: var(--text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  &__sub {
    font-size: 11px;
    color: var(--text-tertiary);
  }

  &:hover:not(&--active) {
    background: var(--surface-card);
  }

  &--active {
    background: var(--surface-card);
    border-color: var(--border-hair);
    box-shadow: -2px 0 0 0 var(--brand-ink);

    .row-item__icon {
      border-color: var(--border-medium);
      color: var(--text-primary);
    }
  }
}

.add-cta {
  margin: 0 12px 16px;
  padding: 10px 14px;
  background: transparent;
  border: 1px dashed var(--border-medium);
  border-radius: 4px;
  font-size: 12px;
  color: var(--text-tertiary);
  cursor: pointer;
  transition: color 120ms ease, border-color 120ms ease;
  font-family: var(--font-sans);

  &:hover {
    border-color: var(--brand-ink);
    color: var(--text-primary);
  }
}

.text-link {
  background: transparent;
  border: none;
  color: var(--brand-ink);
  font-size: 12px;
  cursor: pointer;
  padding: 0;
  text-decoration: underline;
}

// ===== COL 3: Edit column =====
.edit-col {
  flex: 1;
  background: var(--brand-paper);
  padding: 32px 40px 40px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 24px;
  min-width: 0;
}

.empty-card {
  flex: 1;
  display: grid;
  place-items: center;
  border: 1px dashed var(--border-medium);
  border-radius: 8px;
  background: transparent;
  min-height: 240px;

  &__code {
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.12em;
    color: var(--text-tertiary);
    text-transform: uppercase;
  }
}

.edit-header {
  display: flex;
  flex-direction: column;
  gap: 10px;
  max-width: 720px;

  &__title-row {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  &__title {
    font-family: var(--font-serif);
    font-size: 26px;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0;
    line-height: 1.2;
  }

  &__spacer {
    flex: 1;
  }

  &__status {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-family: var(--font-mono);
    font-size: 11px;
    letter-spacing: 0.06em;
    color: var(--text-tertiary);
  }

  &__desc {
    font-size: 13px;
    line-height: 1.6;
    color: var(--text-body);
    margin: 0;
  }
}

// ===== Form card =====
.form-card {
  background: var(--surface-card);
  border: 1px solid var(--border-hair);
  border-radius: 8px;
  padding: 28px;
  max-width: 600px;
}

.form-grid {
  display: flex;
  flex-direction: column;
  gap: 22px;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 6px;

  &__label {
    font-size: 12px;
    font-weight: 500;
    color: var(--text-body);
    display: flex;
    align-items: center;
    gap: 6px;
  }

  &__hint {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-tertiary);
    letter-spacing: 0.04em;
  }

  &__hint-line {
    margin: 4px 0 0;
    font-size: 11px;
    color: var(--text-tertiary);
    line-height: 1.5;
  }

  &__required {
    color: var(--brand-danger);
  }

  // mono input flavor (api keys, app ids)
  &__mono :deep(.el-input__inner),
  &__mono :deep(input) {
    font-family: var(--font-mono);
    letter-spacing: 0.04em;
  }

  :deep(.el-input-number) {
    width: 100%;
  }

  :deep(.el-input-number .el-input__inner) {
    text-align: left;
  }

  :deep(.el-select) {
    width: 100%;
  }
}

.field-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 18px;
}

.field-switch {
  display: flex;
  align-items: center;
  gap: 12px;
  padding-top: 4px;

  &__label {
    font-size: 12px;
    color: var(--text-body);
  }
}

// ===== Bottom action bar =====
.action-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  max-width: 600px;
  padding-top: 4px;

  &__right {
    display: flex;
    align-items: center;
    gap: 8px;
  }
}

%btn-base {
  height: 32px;
  padding: 0 16px;
  border-radius: 4px;
  font-size: 12px;
  font-family: var(--font-sans);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transition: background 120ms ease, color 120ms ease, border-color 120ms ease;
  white-space: nowrap;

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
}

.btn-primary {
  @extend %btn-base;
  background: var(--brand-ink);
  color: var(--text-inverse);
  border: 1px solid var(--brand-ink);

  &:hover:not(:disabled) {
    background: var(--brand-ink-soft);
  }
}

.btn-ghost {
  @extend %btn-base;
  background: transparent;
  color: var(--text-body);
  border: 1px solid var(--border-medium);

  &:hover:not(:disabled) {
    border-color: var(--brand-ink);
    color: var(--text-primary);
  }
}

.btn-danger-ghost {
  @extend %btn-base;
  background: transparent;
  color: var(--brand-danger);
  border: 1px solid transparent;

  &:hover:not(:disabled) {
    border-color: var(--brand-danger);
    background: var(--brand-danger-soft);
  }
}

// ===== Spinner =====
.mini-spinner {
  width: 12px;
  height: 12px;
  border: 1.5px solid var(--border-medium);
  border-top-color: var(--brand-ink);
  border-radius: 50%;
  display: inline-block;
  animation: spin 0.8s linear infinite;

  &--inverse {
    border-color: rgba(255, 255, 255, 0.35);
    border-top-color: #fff;
  }
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

// ===== Delivery mode card grid =====
.delivery-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 8px;
  margin-top: 4px;
}

.delivery-card {
  display: flex;
  flex-direction: column;
  gap: 4px;
  text-align: left;
  padding: 12px 14px;
  background: var(--brand-paper);
  border: 1px solid var(--border-hair);
  border-radius: 6px;
  cursor: pointer;
  transition: border-color 120ms ease, background 120ms ease;

  &__title {
    font-size: 13px;
    font-weight: 500;
    color: var(--text-primary);
  }

  &__hint {
    font-size: 11px;
    color: var(--text-tertiary);
    line-height: 1.5;
  }

  &:hover {
    border-color: var(--border-medium);
  }

  &--active {
    border-color: var(--brand-ink);
    background: var(--surface-card);
    box-shadow: inset 2px 0 0 0 var(--brand-ink);

    .delivery-card__title {
      color: var(--brand-ink);
    }
  }
}

// ===== Extension panel =====
.form-card--guide {
  max-width: 720px;
}

.guide-h3 {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  letter-spacing: 0.04em;
  margin: 24px 0 12px;

  &:first-child {
    margin-top: 0;
  }
}

.guide-list,
.guide-p {
  font-size: 13px;
  line-height: 1.8;
  color: var(--text-body);
}

.guide-list {
  padding-left: 20px;
  margin: 0 0 8px;

  code {
    font-family: var(--font-mono);
    font-size: 12px;
    background: var(--brand-paper);
    border: 1px solid var(--border-hair);
    padding: 1px 5px;
    border-radius: 3px;
  }
}

.guide-p {
  margin: 0 0 12px;

  strong {
    color: var(--brand-danger);
    font-weight: 500;
  }
}

.guide-empty {
  font-size: 12px;
  color: var(--text-faint);
  padding: 16px 0;
}

.token-issue-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.token-issued {
  background: var(--brand-paper);
  border: 1px dashed var(--brand-ink);
  border-radius: 6px;
  padding: 12px 14px;
  margin-bottom: 18px;

  &__head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 6px;
  }

  &__label {
    font-size: 11px;
    font-family: var(--font-mono);
    letter-spacing: 0.06em;
    color: var(--text-tertiary);
  }

  &__value {
    display: block;
    font-family: var(--font-mono);
    font-size: 13px;
    word-break: break-all;
    color: var(--text-primary);
    user-select: all;
  }
}

.token-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;

  &__item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 10px 14px;
    border: 1px solid var(--border-hair);
    border-radius: 6px;
    background: var(--brand-paper);
  }

  &__main {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 4px;
    min-width: 0;
  }

  &__name {
    font-size: 13px;
    font-weight: 500;
    color: var(--text-primary);
  }

  &__prefix {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-tertiary);
  }

  &__meta {
    font-size: 11px;
    color: var(--text-tertiary);
  }

  &__state--revoked {
    color: var(--brand-danger);
  }
}

// ===== Element Plus token overrides (scoped) =====
:deep(.el-input__wrapper),
:deep(.el-textarea__inner) {
  background: var(--surface-card);
  border-radius: 4px;
  box-shadow: 0 0 0 1px var(--border-hair) inset;
  transition: box-shadow 120ms ease;

  &:hover {
    box-shadow: 0 0 0 1px var(--border-medium) inset;
  }

  &.is-focus {
    box-shadow: 0 0 0 1px var(--brand-ink) inset;
  }
}

:deep(.el-input__inner) {
  font-size: 13px;
  color: var(--text-primary);

  &::placeholder {
    color: var(--text-placeholder);
  }
}

:deep(.el-select .el-input__wrapper) {
  background: var(--surface-card);
}

:deep(.el-switch.is-checked .el-switch__core) {
  background-color: var(--brand-ink) !important;
  border-color: var(--brand-ink) !important;
}

:deep(.el-slider__bar) {
  background-color: var(--brand-ink);
}

:deep(.el-slider__button) {
  border-color: var(--brand-ink);
}
</style>
