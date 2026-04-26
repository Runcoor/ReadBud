<template>
  <div class="settings-page">
    <!-- Header -->
    <header class="settings-header">
      <div class="header-brand">
        <h1 class="header-title">阅芽</h1>
        <span class="header-divider">|</span>
        <span class="header-desc">系统设置</span>
      </div>
      <div class="header-actions">
        <button class="back-btn" @click="router.push({ name: 'Workbench' })">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="15 18 9 12 15 6" />
          </svg>
          返回工作台
        </button>
      </div>
    </header>

    <!-- Tab switcher -->
    <div class="tab-bar">
      <button
        class="tab-item"
        :class="{ active: activeTab === 'providers' }"
        @click="activeTab = 'providers'"
      >
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <rect x="2" y="3" width="20" height="18" rx="2" />
          <path d="M8 7h8M8 12h5" />
        </svg>
        服务配置
      </button>
      <button
        class="tab-item"
        :class="{ active: activeTab === 'wechat' }"
        @click="activeTab = 'wechat'"
      >
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z" />
        </svg>
        公众号管理
      </button>
    </div>

    <!-- Provider Tab — split layout -->
    <main v-if="activeTab === 'providers'" class="settings-body">
      <div class="split-layout">
        <!-- Left: Provider list -->
        <aside class="provider-sidebar">
          <div class="sidebar-header">
            <span class="sidebar-title">服务列表</span>
            <span v-if="providers.length" class="sidebar-count">{{ providers.length }}</span>
          </div>

          <div v-if="providerLoading" class="sidebar-loading">
            <span class="mini-spinner"></span>
            <span>加载中...</span>
          </div>

          <div v-else-if="providerError" class="sidebar-error">
            <p>{{ providerError }}</p>
            <button class="text-btn" @click="loadProviders">重试</button>
          </div>

          <div v-else class="provider-list">
            <button
              v-for="p in providers"
              :key="p.id"
              class="provider-item"
              :class="{ active: selectedProviderId === p.id }"
              @click="selectProvider(p)"
            >
              <div class="provider-icon">
                {{ getProviderIcon(p.provider_type) }}
              </div>
              <div class="provider-info">
                <span class="provider-name">{{ p.provider_name }}</span>
                <span class="provider-type-label">{{ getProviderLabel(p.provider_type) }}</span>
              </div>
              <span v-if="p.is_default" class="badge badge-default" title="默认">默认</span>
              <span
                class="status-dot"
                :class="{ online: p.status === 1 }"
                :title="p.status === 1 ? '启用' : '停用'"
              ></span>
            </button>

            <div v-if="providers.length === 0" class="sidebar-empty">
              <p>暂无服务配置</p>
            </div>
          </div>

          <button class="add-provider-btn" @click="startAddProvider">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="12" y1="5" x2="12" y2="19" />
              <line x1="5" y1="12" x2="19" y2="12" />
            </svg>
            添加服务
          </button>
        </aside>

        <!-- Right: Provider detail / form -->
        <section class="provider-detail">
          <!-- Empty state -->
          <div v-if="!editingProvider && !isAddingNew" class="detail-empty">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" opacity="0.25">
              <rect x="2" y="3" width="20" height="18" rx="2" />
              <path d="M8 7h8M8 12h5M8 17h3" />
            </svg>
            <p>选择左侧服务查看配置</p>
          </div>

          <!-- Add new provider form -->
          <div v-else-if="isAddingNew" class="detail-form">
            <div class="detail-header">
              <h3>添加新服务</h3>
            </div>

            <div class="form-scroll">
              <div class="form-group">
                <label class="form-label">服务类型</label>
                <div class="select-wrapper">
                  <select v-model="newProviderType" class="mono-select" @change="onNewTypeChange">
                    <option value="" disabled>请选择服务类型</option>
                    <option v-for="(label, key) in PROVIDER_TYPE_LABELS" :key="key" :value="key">
                      {{ label }}
                    </option>
                  </select>
                </div>
              </div>

              <template v-if="newProviderType">
                <div class="form-group">
                  <label class="form-label">服务名称</label>
                  <input v-model="formFields.name" class="mono-input" placeholder="例如：OpenAI GPT-4" />
                </div>

                <!-- LLM fields -->
                <template v-if="newProviderType === 'llm'">
                  <div class="form-group">
                    <label class="form-label">API 格式</label>
                    <div class="select-wrapper">
                      <select v-model="formFields.api_format" class="mono-select">
                        <option value="openai">OpenAI 兼容</option>
                        <option value="anthropic">Anthropic</option>
                      </select>
                    </div>
                  </div>
                  <div class="form-group">
                    <label class="form-label">API 地址</label>
                    <input v-model="formFields.base_url" class="mono-input" placeholder="https://api.openai.com/v1" />
                  </div>
                  <div class="form-group">
                    <label class="form-label">模型名称</label>
                    <input v-model="formFields.model" class="mono-input" placeholder="gpt-4 / claude-3-opus / deepseek-chat" />
                  </div>
                  <div class="form-group">
                    <label class="form-label">API Key</label>
                    <div class="input-with-toggle">
                      <input
                        v-model="formFields.api_key"
                        :type="showApiKey ? 'text' : 'password'"
                        class="mono-input mono-key"
                        placeholder="sk-..."
                      />
                      <button type="button" class="toggle-btn" @click="showApiKey = !showApiKey">
                        <svg v-if="!showApiKey" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                          <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" /><circle cx="12" cy="12" r="3" />
                        </svg>
                        <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                          <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24" />
                          <line x1="1" y1="1" x2="23" y2="23" />
                        </svg>
                      </button>
                    </div>
                  </div>
                  <div class="form-row">
                    <div class="form-group flex-1">
                      <label class="form-label">温度 <span class="label-value">{{ formFields.temperature }}</span></label>
                      <input
                        v-model.number="formFields.temperature"
                        type="range"
                        min="0"
                        max="2"
                        step="0.1"
                        class="mono-slider"
                      />
                    </div>
                    <div class="form-group flex-1">
                      <label class="form-label">最大 Token 数</label>
                      <input v-model.number="formFields.max_tokens" type="number" class="mono-input" placeholder="4096" />
                    </div>
                  </div>
                </template>

                <!-- Image Gen fields -->
                <template v-else-if="newProviderType === 'image_gen'">
                  <div class="form-group">
                    <label class="form-label">API 格式</label>
                    <div class="select-wrapper">
                      <select v-model="formFields.api_format" class="mono-select">
                        <option value="openai">OpenAI 兼容</option>
                        <option value="vertex_imagen">Vertex AI Imagen</option>
                      </select>
                    </div>
                  </div>
                  <template v-if="formFields.api_format === 'openai'">
                    <div class="form-group">
                      <label class="form-label">API 地址</label>
                      <input v-model="formFields.base_url" class="mono-input" placeholder="https://api.openai.com/v1" />
                    </div>
                    <div class="form-group">
                      <label class="form-label">模型名称</label>
                      <input v-model="formFields.model" class="mono-input" placeholder="dall-e-3" />
                    </div>
                    <div class="form-group">
                      <label class="form-label">API Key</label>
                      <div class="input-with-toggle">
                        <input
                          v-model="formFields.api_key"
                          :type="showApiKey ? 'text' : 'password'"
                          class="mono-input mono-key"
                          placeholder="sk-..."
                        />
                        <button type="button" class="toggle-btn" @click="showApiKey = !showApiKey">
                          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" /><circle cx="12" cy="12" r="3" />
                          </svg>
                        </button>
                      </div>
                    </div>
                  </template>
                  <template v-else-if="formFields.api_format === 'vertex_imagen'">
                    <div class="form-group">
                      <label class="form-label">项目 ID</label>
                      <input v-model="formFields.project_id" class="mono-input" placeholder="my-gcp-project" />
                    </div>
                    <div class="form-group">
                      <label class="form-label">区域</label>
                      <input v-model="formFields.region" class="mono-input" placeholder="us-central1" />
                    </div>
                    <div class="form-group">
                      <label class="form-label">模型名称</label>
                      <input v-model="formFields.model" class="mono-input" placeholder="imagegeneration@006" />
                    </div>
                    <div class="form-group">
                      <label class="form-label">API Key / 凭证</label>
                      <div class="input-with-toggle">
                        <input
                          v-model="formFields.api_key"
                          :type="showApiKey ? 'text' : 'password'"
                          class="mono-input mono-key"
                          placeholder="凭证 JSON 或 API Key"
                        />
                        <button type="button" class="toggle-btn" @click="showApiKey = !showApiKey">
                          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" /><circle cx="12" cy="12" r="3" />
                          </svg>
                        </button>
                      </div>
                    </div>
                  </template>
                  <div class="form-row">
                    <div class="form-group flex-1">
                      <label class="form-label">图片宽度</label>
                      <input v-model.number="formFields.image_width" type="number" class="mono-input" placeholder="1024" />
                    </div>
                    <div class="form-group flex-1">
                      <label class="form-label">图片高度</label>
                      <input v-model.number="formFields.image_height" type="number" class="mono-input" placeholder="1024" />
                    </div>
                  </div>
                </template>

                <!-- Search fields -->
                <template v-else-if="newProviderType === 'search'">
                  <div class="form-group">
                    <label class="form-label">搜索服务商</label>
                    <div class="select-wrapper">
                      <select v-model="formFields.search_provider" class="mono-select">
                        <option value="google_custom">Google Custom Search</option>
                        <option value="tavily">Tavily Search</option>
                        <option value="serpapi">SerpAPI</option>
                        <option value="bing">Bing Search</option>
                      </select>
                    </div>
                  </div>
                  <template v-if="formFields.search_provider === 'google_custom'">
                    <div class="form-group">
                      <label class="form-label">搜索引擎 ID</label>
                      <input v-model="formFields.search_engine_id" class="mono-input" placeholder="Google CSE 的搜索引擎 ID" />
                      <span class="form-hint">在 https://programmablesearchengine.google.com 创建获取</span>
                    </div>
                    <div class="form-group">
                      <label class="form-label">API Key</label>
                      <div class="input-with-toggle">
                        <input v-model="formFields.api_key" :type="showApiKey ? 'text' : 'password'" class="mono-input mono-key" placeholder="Google API Key" />
                        <button type="button" class="toggle-btn" @click="showApiKey = !showApiKey">
                          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" /><circle cx="12" cy="12" r="3" /></svg>
                        </button>
                      </div>
                      <span class="form-hint">在 Google Cloud Console → APIs & Services → Credentials 获取</span>
                    </div>
                  </template>
                  <template v-else-if="formFields.search_provider === 'tavily'">
                    <div class="form-group">
                      <label class="form-label">API Key</label>
                      <div class="input-with-toggle">
                        <input v-model="formFields.api_key" :type="showApiKey ? 'text' : 'password'" class="mono-input mono-key" placeholder="tvly-xxxx" />
                        <button type="button" class="toggle-btn" @click="showApiKey = !showApiKey">
                          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" /><circle cx="12" cy="12" r="3" /></svg>
                        </button>
                      </div>
                      <span class="form-hint">在 https://app.tavily.com 注册获取，免费 1000 次/月</span>
                    </div>
                  </template>
                  <template v-else-if="formFields.search_provider === 'serpapi'">
                    <div class="form-group">
                      <label class="form-label">API Key</label>
                      <div class="input-with-toggle">
                        <input v-model="formFields.api_key" :type="showApiKey ? 'text' : 'password'" class="mono-input mono-key" placeholder="SerpAPI Key" />
                        <button type="button" class="toggle-btn" @click="showApiKey = !showApiKey">
                          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" /><circle cx="12" cy="12" r="3" /></svg>
                        </button>
                      </div>
                      <span class="form-hint">在 https://serpapi.com/dashboard 获取，免费 100 次/月</span>
                    </div>
                  </template>
                  <template v-else-if="formFields.search_provider === 'bing'">
                    <div class="form-group">
                      <label class="form-label">API Key</label>
                      <div class="input-with-toggle">
                        <input v-model="formFields.api_key" :type="showApiKey ? 'text' : 'password'" class="mono-input mono-key" placeholder="Bing Search API Key" />
                        <button type="button" class="toggle-btn" @click="showApiKey = !showApiKey">
                          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" /><circle cx="12" cy="12" r="3" /></svg>
                        </button>
                      </div>
                      <span class="form-hint">在 Azure Portal → Cognitive Services → Bing Search 获取</span>
                    </div>
                  </template>
                </template>

                <!-- Image search fields -->
                <template v-else-if="newProviderType === 'image_search'">
                  <div class="form-group">
                    <label class="form-label">图片搜索服务</label>
                    <div class="select-wrapper">
                      <select v-model="formFields.image_search_provider" class="mono-select">
                        <option value="unsplash">Unsplash（免费）</option>
                        <option value="pexels">Pexels（免费）</option>
                        <option value="bing">Bing Image Search</option>
                      </select>
                    </div>
                  </div>
                  <div class="form-group">
                    <label class="form-label">API Key</label>
                    <div class="input-with-toggle">
                      <input v-model="formFields.api_key" :type="showApiKey ? 'text' : 'password'" class="mono-input mono-key" :placeholder="formFields.image_search_provider === 'unsplash' ? 'Access Key' : 'API Key'" />
                      <button type="button" class="toggle-btn" @click="showApiKey = !showApiKey">
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" /><circle cx="12" cy="12" r="3" /></svg>
                      </button>
                    </div>
                    <span v-if="formFields.image_search_provider === 'unsplash'" class="form-hint">在 https://unsplash.com/developers 注册应用获取</span>
                    <span v-else-if="formFields.image_search_provider === 'pexels'" class="form-hint">在 https://www.pexels.com/api/ 注册获取</span>
                    <span v-else class="form-hint">在 Azure Portal → Bing Image Search 获取</span>
                  </div>
                </template>

                <!-- Crawler fields -->
                <template v-else-if="newProviderType === 'crawler'">
                  <div class="form-group">
                    <label class="form-label">抓取服务</label>
                    <div class="select-wrapper">
                      <select v-model="formFields.crawler_provider" class="mono-select">
                        <option value="jina">Jina Reader（推荐，简单好用）</option>
                        <option value="firecrawl">Firecrawl</option>
                        <option value="custom">自定义接口</option>
                      </select>
                    </div>
                  </div>
                  <template v-if="formFields.crawler_provider === 'jina'">
                    <div class="form-group">
                      <label class="form-label">API Key</label>
                      <div class="input-with-toggle">
                        <input v-model="formFields.api_key" :type="showApiKey ? 'text' : 'password'" class="mono-input mono-key" placeholder="jina_xxxx" />
                        <button type="button" class="toggle-btn" @click="showApiKey = !showApiKey">
                          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" /><circle cx="12" cy="12" r="3" /></svg>
                        </button>
                      </div>
                      <span class="form-hint">在 https://jina.ai/reader 获取，有免费额度</span>
                    </div>
                  </template>
                  <template v-else-if="formFields.crawler_provider === 'firecrawl'">
                    <div class="form-group">
                      <label class="form-label">API Key</label>
                      <div class="input-with-toggle">
                        <input v-model="formFields.api_key" :type="showApiKey ? 'text' : 'password'" class="mono-input mono-key" placeholder="fc-xxxx" />
                        <button type="button" class="toggle-btn" @click="showApiKey = !showApiKey">
                          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" /><circle cx="12" cy="12" r="3" /></svg>
                        </button>
                      </div>
                      <span class="form-hint">在 https://firecrawl.dev 注册获取</span>
                    </div>
                  </template>
                  <template v-else>
                    <div class="form-group">
                      <label class="form-label">API 地址</label>
                      <input v-model="formFields.base_url" class="mono-input" placeholder="https://your-crawler-api.com" />
                    </div>
                    <div class="form-group">
                      <label class="form-label">API Key（可选）</label>
                      <div class="input-with-toggle">
                        <input v-model="formFields.api_key" :type="showApiKey ? 'text' : 'password'" class="mono-input mono-key" placeholder="留空则不传认证" />
                        <button type="button" class="toggle-btn" @click="showApiKey = !showApiKey">
                          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" /><circle cx="12" cy="12" r="3" /></svg>
                        </button>
                      </div>
                    </div>
                  </template>
                  <div class="form-group">
                    <label class="form-label">超时时间（秒）</label>
                    <input v-model.number="formFields.crawler_timeout" type="number" class="mono-input" placeholder="30" />
                  </div>
                </template>
              </template>
            </div>

            <div class="detail-actions">
              <button class="btn-secondary" @click="cancelAdd">取消</button>
              <button class="btn-primary" :disabled="providerSaving" @click="handleCreateProvider">
                <span v-if="providerSaving" class="mini-spinner"></span>
                保存
              </button>
            </div>
          </div>

          <!-- Edit existing provider form -->
          <div v-else-if="editingProvider" class="detail-form">
            <div class="detail-header">
              <h3>{{ editingProvider.provider_name }}</h3>
              <span class="type-badge">
                {{ getProviderLabel(editingProvider.provider_type) }}
              </span>
            </div>

            <div class="form-scroll">
              <div class="form-group">
                <label class="form-label">服务名称</label>
                <input v-model="formFields.name" class="mono-input" />
              </div>

              <!-- LLM fields -->
              <template v-if="editingProvider.provider_type === 'llm'">
                <div class="form-group">
                  <label class="form-label">API 格式</label>
                  <div class="select-wrapper">
                    <select v-model="formFields.api_format" class="mono-select">
                      <option value="openai">OpenAI 兼容</option>
                      <option value="anthropic">Anthropic</option>
                    </select>
                  </div>
                </div>
                <div class="form-group">
                  <label class="form-label">API 地址</label>
                  <input v-model="formFields.base_url" class="mono-input" placeholder="https://api.openai.com/v1" />
                </div>
                <div class="form-group">
                  <label class="form-label">模型名称</label>
                  <input v-model="formFields.model" class="mono-input" placeholder="gpt-4" />
                </div>
                <div class="form-group">
                  <label class="form-label">API Key</label>
                  <div class="input-with-toggle">
                    <input
                      v-model="formFields.api_key"
                      :type="showApiKey ? 'text' : 'password'"
                      class="mono-input mono-key"
                      :placeholder="editingProvider.has_secret ? '已配置（留空保持不变）' : 'sk-...'"
                    />
                    <button type="button" class="toggle-btn" @click="showApiKey = !showApiKey">
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" /><circle cx="12" cy="12" r="3" />
                      </svg>
                    </button>
                  </div>
                </div>
                <div class="form-row">
                  <div class="form-group flex-1">
                    <label class="form-label">温度 <span class="label-value">{{ formFields.temperature }}</span></label>
                    <input v-model.number="formFields.temperature" type="range" min="0" max="2" step="0.1" class="mono-slider" />
                  </div>
                  <div class="form-group flex-1">
                    <label class="form-label">最大 Token 数</label>
                    <input v-model.number="formFields.max_tokens" type="number" class="mono-input" placeholder="4096" />
                  </div>
                </div>
              </template>

              <!-- Image Gen fields -->
              <template v-else-if="editingProvider.provider_type === 'image_gen'">
                <div class="form-group">
                  <label class="form-label">API 格式</label>
                  <div class="select-wrapper">
                    <select v-model="formFields.api_format" class="mono-select">
                      <option value="openai">OpenAI 兼容</option>
                      <option value="vertex_imagen">Vertex AI Imagen</option>
                    </select>
                  </div>
                </div>
                <template v-if="formFields.api_format === 'openai'">
                  <div class="form-group">
                    <label class="form-label">API 地址</label>
                    <input v-model="formFields.base_url" class="mono-input" placeholder="https://api.openai.com/v1" />
                  </div>
                  <div class="form-group">
                    <label class="form-label">模型名称</label>
                    <input v-model="formFields.model" class="mono-input" placeholder="dall-e-3" />
                  </div>
                </template>
                <template v-else-if="formFields.api_format === 'vertex_imagen'">
                  <div class="form-group">
                    <label class="form-label">项目 ID</label>
                    <input v-model="formFields.project_id" class="mono-input" />
                  </div>
                  <div class="form-group">
                    <label class="form-label">区域</label>
                    <input v-model="formFields.region" class="mono-input" />
                  </div>
                  <div class="form-group">
                    <label class="form-label">模型名称</label>
                    <input v-model="formFields.model" class="mono-input" />
                  </div>
                </template>
                <div class="form-group">
                  <label class="form-label">API Key</label>
                  <div class="input-with-toggle">
                    <input
                      v-model="formFields.api_key"
                      :type="showApiKey ? 'text' : 'password'"
                      class="mono-input mono-key"
                      :placeholder="editingProvider.has_secret ? '已配置（留空保持不变）' : 'API Key'"
                    />
                    <button type="button" class="toggle-btn" @click="showApiKey = !showApiKey">
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" /><circle cx="12" cy="12" r="3" />
                      </svg>
                    </button>
                  </div>
                </div>
                <div class="form-row">
                  <div class="form-group flex-1">
                    <label class="form-label">图片宽度</label>
                    <input v-model.number="formFields.image_width" type="number" class="mono-input" placeholder="1024" />
                  </div>
                  <div class="form-group flex-1">
                    <label class="form-label">图片高度</label>
                    <input v-model.number="formFields.image_height" type="number" class="mono-input" placeholder="1024" />
                  </div>
                </div>
              </template>

              <!-- Search fields (edit) -->
              <template v-else-if="editingProvider.provider_type === 'search'">
                <div class="form-group">
                  <label class="form-label">搜索服务商</label>
                  <div class="select-wrapper">
                    <select v-model="formFields.search_provider" class="mono-select">
                      <option value="google_custom">Google Custom Search</option>
                      <option value="serpapi">SerpAPI</option>
                      <option value="bing">Bing Search</option>
                    </select>
                  </div>
                </div>
                <template v-if="formFields.search_provider === 'google_custom'">
                  <div class="form-group">
                    <label class="form-label">搜索引擎 ID</label>
                    <input v-model="formFields.search_engine_id" class="mono-input" placeholder="Google CSE 搜索引擎 ID" />
                    <span class="form-hint">在 https://programmablesearchengine.google.com 创建获取</span>
                  </div>
                </template>
                <div class="form-group">
                  <label class="form-label">API Key</label>
                  <div class="input-with-toggle">
                    <input v-model="formFields.api_key" :type="showApiKey ? 'text' : 'password'" class="mono-input mono-key" :placeholder="editingProvider.has_secret ? '已配置（留空保持不变）' : 'API Key'" />
                    <button type="button" class="toggle-btn" @click="showApiKey = !showApiKey">
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" /><circle cx="12" cy="12" r="3" /></svg>
                    </button>
                  </div>
                </div>
              </template>

              <!-- Image search fields (edit) -->
              <template v-else-if="editingProvider.provider_type === 'image_search'">
                <div class="form-group">
                  <label class="form-label">图片搜索服务</label>
                  <div class="select-wrapper">
                    <select v-model="formFields.image_search_provider" class="mono-select">
                      <option value="unsplash">Unsplash（免费）</option>
                      <option value="pexels">Pexels（免费）</option>
                      <option value="bing">Bing Image Search</option>
                    </select>
                  </div>
                </div>
                <div class="form-group">
                  <label class="form-label">API Key</label>
                  <div class="input-with-toggle">
                    <input v-model="formFields.api_key" :type="showApiKey ? 'text' : 'password'" class="mono-input mono-key" :placeholder="editingProvider.has_secret ? '已配置（留空保持不变）' : 'Access Key'" />
                    <button type="button" class="toggle-btn" @click="showApiKey = !showApiKey">
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" /><circle cx="12" cy="12" r="3" /></svg>
                    </button>
                  </div>
                </div>
              </template>

              <!-- Crawler fields (edit) -->
              <template v-else-if="editingProvider.provider_type === 'crawler'">
                <div class="form-group">
                  <label class="form-label">抓取服务</label>
                  <div class="select-wrapper">
                    <select v-model="formFields.crawler_provider" class="mono-select">
                      <option value="jina">Jina Reader</option>
                      <option value="firecrawl">Firecrawl</option>
                      <option value="custom">自定义接口</option>
                    </select>
                  </div>
                </div>
                <template v-if="formFields.crawler_provider === 'custom'">
                  <div class="form-group">
                    <label class="form-label">API 地址</label>
                    <input v-model="formFields.base_url" class="mono-input" placeholder="https://..." />
                  </div>
                </template>
                <div class="form-group">
                  <label class="form-label">API Key</label>
                  <div class="input-with-toggle">
                    <input v-model="formFields.api_key" :type="showApiKey ? 'text' : 'password'" class="mono-input mono-key" :placeholder="editingProvider.has_secret ? '已配置（留空保持不变）' : 'API Key'" />
                    <button type="button" class="toggle-btn" @click="showApiKey = !showApiKey">
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" /><circle cx="12" cy="12" r="3" /></svg>
                    </button>
                  </div>
                </div>
                <div class="form-group">
                  <label class="form-label">超时时间（秒）</label>
                  <input v-model.number="formFields.crawler_timeout" type="number" class="mono-input" placeholder="30" />
                </div>
              </template>
            </div>

            <div class="detail-actions">
              <button class="btn-danger" @click="handleDeleteProvider">删除</button>
              <div class="actions-right">
                <button
                  v-if="editingProvider && !editingProvider.is_default"
                  class="btn-secondary"
                  :disabled="settingDefault"
                  @click="handleSetDefault"
                >
                  <span v-if="settingDefault" class="mini-spinner"></span>
                  设为默认
                </button>
                <button class="btn-secondary" :disabled="testingConnection" @click="handleTestConnection">
                  <span v-if="testingConnection" class="mini-spinner"></span>
                  测试连接
                </button>
                <button class="btn-primary" :disabled="providerSaving" @click="handleUpdateProvider">
                  <span v-if="providerSaving" class="mini-spinner"></span>
                  保存
                </button>
              </div>
            </div>
          </div>
        </section>
      </div>
    </main>

    <!-- WeChat Tab -->
    <main v-else-if="activeTab === 'wechat'" class="settings-body">
      <div class="wechat-panel">
        <div class="panel-header">
          <h3 class="panel-title">公众号账号管理</h3>
          <button class="btn-primary small" @click="showWechatDialog = true">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="12" y1="5" x2="12" y2="19" />
              <line x1="5" y1="12" x2="19" y2="12" />
            </svg>
            添加账号
          </button>
        </div>

        <div v-if="wechatLoading" class="panel-loading">
          <span class="mini-spinner"></span>
          加载中...
        </div>

        <div v-else-if="wechatError" class="panel-error">
          <p>{{ wechatError }}</p>
          <button class="text-btn" @click="loadWechatAccounts">重试</button>
        </div>

        <div v-else-if="wechatAccounts.length === 0" class="panel-empty">
          <p>暂无公众号配置</p>
        </div>

        <div v-else class="wechat-grid">
          <div v-for="acct in wechatAccounts" :key="acct.id" class="wechat-card">
            <div class="wechat-card-header">
              <span class="wechat-name">{{ acct.name }}</span>
              <div class="wechat-badges">
                <span v-if="acct.is_default" class="badge badge-warning">默认</span>
                <span class="badge" :class="acct.status === 1 ? 'badge-success' : 'badge-muted'">
                  {{ acct.status === 1 ? '启用' : '停用' }}
                </span>
              </div>
            </div>
            <div class="wechat-card-body">
              <div class="wechat-field">
                <span class="field-label">AppID</span>
                <span class="field-value">{{ acct.app_id }}</span>
              </div>
              <div class="wechat-field">
                <span class="field-label">令牌模式</span>
                <span class="field-value">{{ getTokenModeLabel(acct.token_mode) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- WeChat Dialog -->
    <teleport to="body">
      <div v-if="showWechatDialog" class="dialog-overlay" @click.self="showWechatDialog = false">
        <div class="dialog-panel">
          <div class="dialog-header">
            <h3>添加公众号</h3>
            <button class="dialog-close" @click="showWechatDialog = false">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="18" y1="6" x2="6" y2="18" /><line x1="6" y1="6" x2="18" y2="18" />
              </svg>
            </button>
          </div>
          <div class="dialog-body">
            <div class="form-group">
              <label class="form-label">名称 <span class="required">*</span></label>
              <input v-model="wechatForm.name" class="mono-input" placeholder="公众号名称" />
            </div>
            <div class="form-group">
              <label class="form-label">AppID <span class="required">*</span></label>
              <input v-model="wechatForm.app_id" class="mono-input" placeholder="公众号 AppID" />
            </div>
            <div class="form-group">
              <label class="form-label">AppSecret</label>
              <div class="input-with-toggle">
                <input
                  v-model="wechatForm.app_secret"
                  :type="showWechatSecret ? 'text' : 'password'"
                  class="mono-input mono-key"
                  placeholder="公众号 AppSecret"
                />
                <button type="button" class="toggle-btn" @click="showWechatSecret = !showWechatSecret">
                  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" /><circle cx="12" cy="12" r="3" />
                  </svg>
                </button>
              </div>
            </div>
            <div class="form-group">
              <label class="form-label">令牌模式 <span class="required">*</span></label>
              <div class="select-wrapper">
                <select v-model="wechatForm.token_mode" class="mono-select">
                  <option v-for="(label, key) in TOKEN_MODE_LABELS" :key="key" :value="key">{{ label }}</option>
                </select>
              </div>
            </div>
            <div class="form-group form-row-inline">
              <label class="toggle-label">
                <input type="checkbox" v-model="wechatForm.is_default" class="mono-checkbox" />
                <span>设为默认</span>
              </label>
            </div>
            <div class="form-group">
              <label class="form-label">备注</label>
              <input v-model="wechatForm.remark" class="mono-input" placeholder="备注信息" />
            </div>
          </div>
          <div class="dialog-footer">
            <button class="btn-secondary" @click="showWechatDialog = false">取消</button>
            <button class="btn-primary" :disabled="wechatSaving" @click="handleCreateWechat">
              <span v-if="wechatSaving" class="mini-spinner"></span>
              保存
            </button>
          </div>
        </div>
      </div>
    </teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listProviders, createProvider, listWechatAccounts, createWechatAccount } from '@/api/provider'
import { patch, del, post } from '@/api/request'
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
function buildSecretJson(providerType: string): string | undefined {
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

function getProviderIcon(type_: string): string {
  const map: Record<string, string> = {
    llm: 'AI',
    image_search: 'IS',
    image_gen: 'IG',
    search: 'SE',
    storage: 'ST',
    crawler: 'CR',
  }
  return map[type_] || '??'
}

function getTokenModeLabel(mode: string): string {
  return TOKEN_MODE_LABELS[mode as TokenMode] || mode
}

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
const showWechatDialog = ref(false)
const showWechatSecret = ref(false)
const wechatForm = reactive({
  name: '',
  app_id: '',
  app_secret: '',
  token_mode: 'direct' as TokenMode,
  is_default: false,
  remark: '',
})

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
@use '@/styles/tokens' as *;

// ============================================================
// Settings page — Monochrome
// ============================================================

.settings-page {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background: var(--surface-bg);
}

// --- Header ---
.settings-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 60px;
  padding: 0 $spacing-xl;
  background: var(--surface-bg);
  border-bottom: 1px solid var(--border-light);
  position: sticky;
  top: 0;
  z-index: $z-sticky;
}

.header-brand {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
}

.header-title {
  font-size: $font-size-lg;
  font-weight: $font-weight-bold;
  color: var(--text-primary);
}

.header-divider {
  color: var(--border-medium);
}

.header-desc {
  font-size: $font-size-base;
  color: $text-secondary;
}

.back-btn {
  display: flex;
  align-items: center;
  gap: $spacing-xs;
  background: var(--surface-bg);
  border: 1px solid var(--border-light);
  border-radius: $radius-md;
  color: $text-secondary;
  padding: $spacing-sm $spacing-md;
  font-size: $font-size-sm;
  cursor: pointer;
  transition: all $transition-base;

  &:hover {
    border-color: var(--border-medium);
    color: var(--text-primary);
    background: var(--surface-tertiary);
  }
}

// --- Tab bar ---
.tab-bar {
  display: flex;
  gap: $spacing-xs;
  padding: $spacing-base $spacing-xl;
  border-bottom: 1px solid var(--border-light);
}

.tab-item {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
  padding: $spacing-sm $spacing-base;
  background: transparent;
  border: 1px solid transparent;
  border-radius: $radius-md;
  color: $text-tertiary;
  font-size: $font-size-sm;
  cursor: pointer;
  transition: all $transition-base;

  &:hover {
    color: $text-secondary;
    background: var(--surface-tertiary);
  }

  &.active {
    color: var(--text-primary);
    background: var(--surface-tertiary);
    border-color: var(--border-light);
    font-weight: $font-weight-medium;
  }
}

// --- Body ---
.settings-body {
  flex: 1;
  padding: $spacing-xl;
  max-width: 1200px;
  width: 100%;
  margin: 0 auto;
}

// --- Split layout ---
.split-layout {
  display: flex;
  gap: $spacing-xl;
  height: calc(100vh - 60px - 56px - 48px);
  min-height: 500px;
}

// --- Sidebar ---
.provider-sidebar {
  width: 280px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  background: var(--surface-bg);
  border: 1px solid var(--border-light);
  border-radius: $radius-lg;
  overflow: hidden;
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: $spacing-base;
  border-bottom: 1px solid var(--border-light);
}

.sidebar-title {
  font-size: $font-size-sm;
  font-weight: $font-weight-semibold;
  color: $text-secondary;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.sidebar-count {
  font-size: $font-size-xs;
  color: $text-tertiary;
  background: var(--surface-tertiary);
  padding: 2px 8px;
  border-radius: $radius-full;
}

.sidebar-loading,
.sidebar-error {
  padding: $spacing-xl;
  text-align: center;
  color: $text-tertiary;
  font-size: $font-size-sm;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: $spacing-sm;
}

.sidebar-empty {
  padding: $spacing-2xl $spacing-base;
  text-align: center;
  color: $text-tertiary;
  font-size: $font-size-sm;
}

.provider-list {
  flex: 1;
  overflow-y: auto;
  padding: $spacing-sm;
}

.provider-item {
  display: flex;
  align-items: center;
  gap: $spacing-md;
  width: 100%;
  padding: $spacing-md;
  background: transparent;
  border: 1px solid transparent;
  border-left: 4px solid transparent;
  border-radius: $radius-md;
  cursor: pointer;
  transition: all $transition-base;
  text-align: left;

  &:hover {
    background: var(--surface-tertiary);
  }

  &.active {
    background: var(--surface-tertiary);
    border-left-color: var(--text-primary);
  }

  & + & {
    margin-top: 2px;
  }
}

.provider-icon {
  width: 36px;
  height: 36px;
  border-radius: $radius-md;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: $font-size-xs;
  font-weight: $font-weight-bold;
  flex-shrink: 0;
  background: var(--surface-tertiary);
  color: var(--text-secondary);
}

.provider-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.provider-name {
  font-size: $font-size-sm;
  font-weight: $font-weight-medium;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.provider-type-label {
  font-size: $font-size-xs;
  color: $text-tertiary;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
  background: var(--border-medium);

  &.online {
    background: $status-success;
  }
}

.add-provider-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: $spacing-sm;
  padding: $spacing-md;
  border-top: 1px solid var(--border-light);
  background: transparent;
  border-left: none;
  border-right: none;
  border-bottom: none;
  color: $text-secondary;
  font-size: $font-size-sm;
  cursor: pointer;
  transition: all $transition-base;

  &:hover {
    background: var(--surface-tertiary);
    color: var(--text-primary);
  }
}

// --- Detail panel ---
.provider-detail {
  flex: 1;
  min-width: 0;
  background: var(--surface-bg);
  border: 1px solid var(--border-light);
  border-radius: $radius-lg;
  padding: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.detail-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: $spacing-base;
  color: $text-tertiary;
  font-size: $font-size-sm;
}

.detail-form {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.detail-header {
  display: flex;
  align-items: center;
  gap: $spacing-md;
  padding: $spacing-lg $spacing-xl;
  border-bottom: 1px solid var(--border-light);

  h3 {
    font-size: $font-size-md;
    font-weight: $font-weight-bold;
    color: var(--text-primary);
  }
}

.type-badge {
  font-size: $font-size-xs;
  padding: 2px $spacing-sm;
  border-radius: $radius-full;
  font-weight: $font-weight-medium;
  background: var(--surface-tertiary);
  color: var(--text-secondary);
}

.form-scroll {
  flex: 1;
  overflow-y: auto;
  padding: $spacing-xl;
  display: flex;
  flex-direction: column;
  gap: $spacing-lg;
}

.detail-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: $spacing-base $spacing-xl;
  border-top: 1px solid var(--border-light);
}

.actions-right {
  display: flex;
  gap: $spacing-sm;
}

// --- Form elements ---
.form-group {
  display: flex;
  flex-direction: column;
  gap: $spacing-sm;
}

.form-label {
  font-size: $font-size-sm;
  font-weight: $font-weight-medium;
  color: $text-secondary;
  display: flex;
  align-items: center;
  gap: $spacing-xs;
}

.label-value {
  color: var(--text-primary);
  font-weight: $font-weight-normal;
}

.required {
  color: $status-danger;
}

.form-hint {
  display: block;
  margin-top: 4px;
  font-size: 11px;
  color: var(--text-tertiary);
  line-height: 1.4;
}

.mono-checkbox {
  margin-right: 6px;
  accent-color: var(--text-primary);
}

.form-row {
  display: flex;
  gap: $spacing-base;
}

.flex-1 {
  flex: 1;
}

.mono-input {
  background: var(--surface-bg);
  border: 1px solid var(--border-light);
  border-radius: $radius-md;
  color: var(--text-primary);
  padding: $spacing-sm $spacing-md;
  font-size: $font-size-base;
  width: 100%;
  transition: all $transition-base;

  &::placeholder {
    color: $text-placeholder;
  }

  &:hover {
    border-color: var(--border-medium);
  }

  &:focus {
    border-color: var(--text-primary);
    box-shadow: 0 0 0 2px rgba(0, 0, 0, 0.06);
    outline: none;
  }
}

.mono-key {
  font-family: $font-mono;
}

.select-wrapper {
  position: relative;

  &::after {
    content: '';
    position: absolute;
    right: $spacing-md;
    top: 50%;
    transform: translateY(-50%);
    width: 0;
    height: 0;
    border-left: 5px solid transparent;
    border-right: 5px solid transparent;
    border-top: 5px solid $text-tertiary;
    pointer-events: none;
  }
}

.mono-select {
  background: var(--surface-bg);
  border: 1px solid var(--border-light);
  border-radius: $radius-md;
  color: var(--text-primary);
  padding: $spacing-sm $spacing-2xl $spacing-sm $spacing-md;
  font-size: $font-size-base;
  width: 100%;
  cursor: pointer;
  appearance: none;
  -webkit-appearance: none;
  transition: all $transition-base;

  &:hover {
    border-color: var(--border-medium);
  }

  &:focus {
    border-color: var(--text-primary);
    box-shadow: 0 0 0 2px rgba(0, 0, 0, 0.06);
    outline: none;
  }

  option {
    background: var(--surface-bg);
    color: var(--text-primary);
  }
}

.mono-slider {
  width: 100%;
  -webkit-appearance: none;
  appearance: none;
  height: 4px;
  border-radius: 2px;
  background: var(--border-light);
  outline: none;
  margin-top: $spacing-sm;

  &::-webkit-slider-thumb {
    -webkit-appearance: none;
    appearance: none;
    width: 18px;
    height: 18px;
    border-radius: 50%;
    background: var(--text-primary);
    border: 2px solid #fff;
    cursor: pointer;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.15);
  }

  &::-moz-range-thumb {
    width: 18px;
    height: 18px;
    border-radius: 50%;
    background: var(--text-primary);
    border: 2px solid #fff;
    cursor: pointer;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.15);
  }
}

.input-with-toggle {
  position: relative;
  display: flex;
  align-items: stretch;

  .mono-input {
    padding-right: 40px;
  }

  .toggle-btn {
    position: absolute;
    right: 0;
    top: 0;
    bottom: 0;
    width: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: transparent;
    border: none;
    color: $text-tertiary;
    cursor: pointer;
    transition: color $transition-fast;

    &:hover {
      color: $text-secondary;
    }
  }
}

.mono-checkbox {
  appearance: none;
  -webkit-appearance: none;
  width: 18px;
  height: 18px;
  border: 1px solid var(--border-light);
  border-radius: $radius-sm;
  background: var(--surface-bg);
  cursor: pointer;
  position: relative;
  transition: all $transition-base;

  &:checked {
    background: var(--text-primary);
    border-color: var(--text-primary);

    &::after {
      content: '';
      position: absolute;
      left: 5px;
      top: 2px;
      width: 6px;
      height: 10px;
      border: solid #fff;
      border-width: 0 2px 2px 0;
      transform: rotate(45deg);
    }
  }
}

.toggle-label {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
  color: $text-secondary;
  font-size: $font-size-sm;
  cursor: pointer;
}

.form-row-inline {
  flex-direction: row;
  align-items: center;
}

// --- Buttons ---
.btn-primary {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: $spacing-sm;
  padding: $spacing-sm $spacing-lg;
  font-size: $font-size-sm;
  font-weight: $font-weight-medium;
  border-radius: $radius-md;
  cursor: pointer;
  transition: all $transition-base;
  background: var(--text-primary);
  border: 1px solid var(--text-primary);
  color: var(--text-inverse);

  &:hover {
    background: var(--surface-inverse);
    border-color: var(--surface-inverse);
  }

  &:active {
    transform: scale(0.98);
  }

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    transform: none !important;
  }

  &.small {
    padding: $spacing-xs $spacing-md;
    font-size: $font-size-xs;
  }
}

.btn-secondary {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: $spacing-sm;
  padding: $spacing-sm $spacing-lg;
  font-size: $font-size-sm;
  font-weight: $font-weight-medium;
  border-radius: $radius-md;
  cursor: pointer;
  transition: all $transition-base;
  background: var(--surface-bg);
  border: 1px solid var(--border-light);
  color: var(--text-primary);

  &:hover {
    background: var(--surface-tertiary);
    border-color: var(--border-medium);
  }

  &:active {
    transform: scale(0.98);
  }

  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    transform: none !important;
  }
}

.btn-danger {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: $spacing-sm;
  padding: $spacing-sm $spacing-lg;
  font-size: $font-size-sm;
  font-weight: $font-weight-medium;
  border-radius: $radius-md;
  cursor: pointer;
  transition: all $transition-base;
  background: transparent;
  border: 1px solid transparent;
  color: $status-danger;

  &:hover {
    background: #fef2f2;
  }

  &:active {
    transform: scale(0.98);
  }
}

.text-btn {
  background: transparent;
  border: none;
  color: $text-secondary;
  font-size: $font-size-sm;
  cursor: pointer;
  transition: color $transition-fast;
  text-decoration: underline;

  &:hover {
    color: var(--text-primary);
  }
}

// --- WeChat panel ---
.wechat-panel {
  background: var(--surface-bg);
  border: 1px solid var(--border-light);
  border-radius: $radius-lg;
  padding: $spacing-xl;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: $spacing-xl;
}

.panel-title {
  font-size: $font-size-md;
  font-weight: $font-weight-bold;
  color: var(--text-primary);
}

.panel-loading,
.panel-error,
.panel-empty {
  text-align: center;
  padding: $spacing-3xl;
  color: $text-tertiary;
  font-size: $font-size-sm;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: $spacing-sm;
}

.wechat-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: $spacing-base;
}

.wechat-card {
  background: var(--surface-bg);
  border: 1px solid var(--border-light);
  border-radius: $radius-lg;
  padding: $spacing-base;
  transition: all $transition-base;

  &:hover {
    border-color: var(--border-medium);
    box-shadow: $shadow-sm;
  }
}

.wechat-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: $spacing-md;
}

.wechat-name {
  font-size: $font-size-base;
  font-weight: $font-weight-semibold;
  color: var(--text-primary);
}

.wechat-badges {
  display: flex;
  gap: $spacing-xs;
}

.badge {
  font-size: $font-size-xs;
  padding: 2px $spacing-sm;
  border-radius: $radius-full;
  font-weight: $font-weight-medium;

  &.badge-default { background: var(--border-light); color: var(--text-secondary); }
  &.badge-success { background: #f0fdf4; color: #16a34a; }
  &.badge-warning { background: #fefce8; color: #ca8a04; }
  &.badge-muted { background: var(--surface-tertiary); color: $text-tertiary; }
}

.wechat-card-body {
  display: flex;
  flex-direction: column;
  gap: $spacing-sm;
}

.wechat-field {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
}

.field-label {
  font-size: $font-size-xs;
  color: $text-tertiary;
  width: 60px;
  flex-shrink: 0;
}

.field-value {
  font-size: $font-size-sm;
  color: $text-secondary;
  font-family: $font-mono;
}

// --- Dialog ---
.dialog-overlay {
  position: fixed;
  inset: 0;
  z-index: $z-modal;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.45);
}

.dialog-panel {
  width: 480px;
  max-width: 92vw;
  max-height: 90vh;
  background: var(--surface-bg);
  border: 1px solid var(--border-light);
  border-radius: $radius-xl;
  box-shadow: $shadow-lg;
  display: flex;
  flex-direction: column;
}

.dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: $spacing-lg $spacing-xl;
  border-bottom: 1px solid var(--border-light);

  h3 {
    font-size: $font-size-md;
    font-weight: $font-weight-bold;
    color: var(--text-primary);
  }
}

.dialog-close {
  background: transparent;
  border: none;
  color: $text-tertiary;
  cursor: pointer;
  padding: $spacing-xs;
  border-radius: $radius-sm;
  transition: all $transition-fast;

  &:hover {
    color: var(--text-primary);
    background: var(--surface-tertiary);
  }
}

.dialog-body {
  padding: $spacing-xl;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: $spacing-lg;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: $spacing-sm;
  padding: $spacing-base $spacing-xl;
  border-top: 1px solid var(--border-light);
}

// --- Utilities ---
.mini-spinner {
  display: inline-block;
  width: 14px;
  height: 14px;
  border: 2px solid var(--border-medium);
  border-top-color: var(--text-primary);
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

// --- Responsive ---
@media (max-width: $breakpoint-md) {
  .split-layout {
    flex-direction: column;
    height: auto;
  }

  .provider-sidebar {
    width: 100%;
    max-height: 300px;
  }

  .provider-detail {
    min-height: 400px;
  }

  .settings-body {
    padding: $spacing-base;
  }
}

@media (max-width: $breakpoint-sm) {
  .settings-header {
    height: 52px;
    padding: 0 $spacing-base;
  }

  .header-desc,
  .header-divider {
    display: none;
  }

  .tab-bar {
    padding: $spacing-sm $spacing-base;
  }

  .settings-body {
    padding: $spacing-sm;
  }

  .form-row {
    flex-direction: column;
  }

  .detail-actions {
    flex-direction: column;
    gap: $spacing-sm;
  }

  .actions-right {
    width: 100%;
    justify-content: flex-end;
  }

  .wechat-grid {
    grid-template-columns: 1fr;
  }
}
</style>
