# ReadBud

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](./LICENSE)

中文 | [English](./README.md)

ReadBud 是面向微信公众号的端到端内容自动化工具。给定一个关键词或选题,它会自动联网检索资料、调用 LLM 撰写初稿、生成封面图、套用排版风格,并将文章投递到公众号草稿箱 —— 可通过微信官方 Draft API 直接投递,也可通过配套浏览器插件自动填充公众号网页编辑器。

## 功能特性

- **检索 → 撰写 → 封面 → 发布** 全流程串联
- **三种发布路径**,适配不同账号类型:
  - **API 模式** — 通过微信 Draft API 直接投递(需已认证的服务号)
  - **插件模式** — Chrome 插件自动填充公众号网页编辑器(个人号亦可)
  - **手动模式** — 输出整理好的内容包,人工复制粘贴
- **可插拔的服务商接入** — LLM、图像生成、网页搜索、内容抓取均为适配层,在设置中切换
- **三套内置排版风格**(minimal、magazine、stitch),封面提示词与正文风格保持一致
- **品牌资料与文风画像** — 多账号可共用或独立配置语气和外观
- **草稿版本历史** 支持回滚
- **流水线可观测** — 每个阶段为一个独立任务,带重试和耗时记录

## 技术栈

- **后端** — Go 1.26、Gin、GORM、Asynq、PostgreSQL、Redis
- **前端** — Vue 3、Pinia、Element Plus、Vite
- **浏览器插件** — Chrome Manifest V3(原生 JS)
- **服务商适配** — LLM / 图像 / 搜索 / 抓取适配层,可按工作区独立配置

## 快速开始

依赖 Docker 与 Docker Compose。

```bash
git clone https://github.com/Runcoor/ReadBud.git
cd ReadBud
cp backend/configs/config.example.yaml backend/configs/config.yaml
# 按需编辑 backend/configs/config.yaml
./start.sh
```

启动后访问:

- 前端:<http://localhost:19880>
- API:<http://localhost:19881>

### 首次配置

1. 注册账号(默认开放注册,生产环境请关闭)
2. **设置 → 服务配置** — 配置 LLM、搜索、图像生成的 API Key
3. **设置 → 公众号管理** — 绑定 AppID、AppSecret,并选择投递方式(API / 插件 / 手动)
4. **设置 → 扩展插件** — 签发令牌,通过 `chrome://extensions` 开启开发者模式后加载 `wechat-extension/` 目录

## 排版风格

| 风格     | 配色                     | 风味               |
| -------- | ------------------------ | ------------------ |
| minimal  | 黑白主调,少量亮黄点缀   | 瑞士 / 编辑设计    |
| magazine | 米白底色,红色点缀       | 印刷杂志           |
| stitch   | 奶油底色 + 赭橙          | 手作复古           |

封面图按风格生成,提示词与配色与正文保持一致。

## 后续计划

- 微信发布成功后回写文章 URL
- 数据分析看板的数据接入(UI 已就位)
- 在 Edge 和 Firefox 上验证插件兼容性(目前仅在 Chrome 测试)

## License

基于 [GNU AGPL-3.0](./LICENSE) 开源。
