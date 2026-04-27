# ReadBud

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](./LICENSE)

写公众号要先搜资料、写、排版、发,流程长,我懒,所以做了这个。

输入一个关键词,ReadBud 会去网上找资料,让 LLM 写初稿,生成封面图,套上排版,最后发到公众号。一个人在浏览器里点几下就完成。

## 关于发布

如果是已认证的服务号,直接走微信 draft API 自动发。

如果是个人号或者没认证 —— 微信压根不开放 draft API 给你。ReadBud 配套了一个 Chrome 插件,点发布会自动跳到微信编辑器,把标题、正文、封面都填好,你只需要在编辑器里点最后那个「群发」。

这是这个项目稍微有点意思的地方:大多数公众号工具都假设你有认证服务号,但其实大部分个人写作者都没有。

## 跑起来

```bash
git clone <repo>
cd ReadBud
./start.sh
```

需要 Docker。起来之后:

- 前端 http://localhost:19880
- API  http://localhost:19881

第一次用:

1. 注册账号(默认开放注册,上生产记得关)
2. 设置 → 服务配置 加 LLM / 搜索 / 图像生成的 API key
3. 设置 → 公众号管理 绑 AppID + AppSecret(没认证就选「插件填充」模式)
4. 装插件:`chrome://extensions` → 开发者模式 → 加载 `wechat-extension/` 目录,填上签发的令牌

## 排版

内置三套风格,文章和封面颜色字体一致,看着像一套设计:

- **minimal** 黑白 + 一点亮黄,瑞士风
- **magazine** 米白 + 红点缀,杂志风
- **stitch**   米黄 + 赭橙,手作复古风

封面用 AI 生成,prompt 是按风格调的,不是套模板那种。

## 技术

- 后端 Go 1.26,Gin + GORM + Asynq,Postgres + Redis
- 前端 Vue 3 + Pinia + Element Plus
- 插件 Chrome MV3,纯 JS

LLM、图像生成、搜索、抓取这些都是适配层,在设置里切换,不锁死任何一家服务商。

## 还没做的

- 微信发布成功后拿回最终文章 URL(目前需要手动粘)
- 数据看板那块只占了位置,数据接进来还没做
- 插件只测过 Chrome,Edge 应该也行,Firefox 没试

## License

[AGPL-3.0](./LICENSE)。自用、改、研究都随意。拿去做 SaaS 商用要么开源改动,要么找我谈商业授权。

## 作者

Leazoot
