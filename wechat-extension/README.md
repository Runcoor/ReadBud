# ReadBud · WeChat 自动填充插件

在微信公众号编辑器自动填入 ReadBud 生成的标题、作者、摘要、正文与封面。
解决「个人公众号无法获得 draft/add API 权限」的问题 —— 不依赖 WeChat 认证服务号,
任何能登录公众号后台的账号都能用。

## 功能

- 在 `mp.weixin.qq.com/cgi-bin/appmsg` 页面浮动按钮一键导入
- 从 ReadBud 拉取标题 / 作者 / 摘要 / 正文 HTML / 封面图
- 自动识别 WeChat 编辑器的输入框并填充
- 封面通过 `<input type="file">` 注入,无需手动上传

## 安装(开发者模式)

> 还未上架 Chrome Web Store。下面是开发者模式加载步骤,几分钟搞定。

1. 打开 `chrome://extensions`(或 `edge://extensions`)
2. 右上角开关打开「开发者模式」
3. 点「加载已解压的扩展程序」
4. 选择本目录(`wechat-extension/`)
5. 工具栏多出 ReadBud 图标 → 点击 → 填入:
   - **API 地址**: 默认 `http://localhost:8080/api/v1`(本地开发);生产填部署的 ReadBud 地址
   - **令牌**: 在 ReadBud 设置 → 浏览器插件 → 签发令牌(`rbex_…` 开头)
6. 点「保存配置」,可选「测试连接」

## 使用流程

1. 在 ReadBud 完成草稿后,把目标公众号的「发布方式」设为「插件填充」
2. 点 ReadBud 里的「通过插件发布」按钮
3. 浏览器自动打开 WeChat 编辑器,带 `?readbud_draft=…&readbud_job=…` 参数
4. 插件检测到参数,1~2 秒后开始自动填充(右下角浮动按钮也可手动触发)
5. 在 WeChat 编辑器里检查内容 → 点「群发」/「保存草稿」
6. 回 ReadBud 点「已发布,标记完成」

## 已知限制

- WeChat 编辑器 DOM 偶尔变动。如果某次填充某个字段失败,看右下角 toast 提示是哪个字段,
  把对应内容手动复制即可;同时把失败字段反馈给我们更新选择器
- 封面如果 WeChat 改了 file input 的位置,可能需要更新 `findCoverFileInput()` 的选择器
- 仅支持 Chromium 系浏览器(Chrome / Edge / Brave / 等);未在 Firefox 测试
- 必须用同一浏览器登录公众号后台,且本扩展已加载

## 文件结构

```
wechat-extension/
├── manifest.json       # MV3 manifest
├── background.js       # service worker (config + API fetch)
├── content_script.js   # 注入到 WeChat 编辑器的脚本
├── content_script.css  # 浮动按钮 + toast 样式
├── popup.html/js/css   # 配置页面
└── icons/              # 工具栏图标(放置 16/48/128 png)
```

## 开发

无需构建步骤 —— 修改 .js 后回 `chrome://extensions` 点扩展卡片的「重新加载」按钮即可。
content_script 的 console 日志在 WeChat 编辑器的 DevTools 里看;
background.js 的日志在扩展卡片的「Service Worker」点开看。
