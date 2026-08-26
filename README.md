# zal-blog

个人博客（界面英文、内容中文）：终端复古 TUI 风格，Go 后端提供 API，前端静态页面 + JS 渲染，文章以 org-mode 格式撰写。

## 运行

```sh
# 在仓库根目录（或 backend/ 目录下）执行均可
cd backend && go run .
# 打开 http://localhost:54345
```

端口默认 `54345`，可用环境变量 `PORT` 覆盖。`BLOG_DIR` / `FRONTEND_DIR` 可覆盖内容与前端目录（默认 `blog` / `frontend`）。

无任何缓存：修改 org 文件或前端文件后刷新即可看到效果。

## 写文章

- 文章放在 `blog/YYYY/MM/slug.org`（按年月分目录），例如 `blog/2025/08/hello-zal-blog.org`
- 头部元信息（org 关键字）：
  - `#+TITLE: 标题`（必填，缺省取第一个 `*` 标题）
  - `#+DATE: 2025-08-25`（缺省取路径中的年月）
  - `#+TAGS: :tag1:tag2:`（可省略）
  - `#+DRAFT: true`（草稿：不出现在列表/标签/RSS 中，可直接访问预览）
- 文章按日期倒序排列

## 图片

图片放在 `blog/assets/` 下，目录层级与文章镜像对应：
文章 `blog/2025/08/x.org` 的图片放 `blog/assets/2025/08/pic.png`，
文中引用 `[[file:pic.png]]` 或 `[[file:pic.png][alt 文字]]`。

## 支持的 org 语法

标题 `*`~`*****`（映射为 h2~h6）、段落、粗体/斜体/下划线/删除线、
行内代码 `~x~` / `=x=`、代码块 `#+begin_src lang`、引用块 `#+begin_quote`、
链接 `[[url][text]]` 与裸 URL、图片 `[[file:...]]`、无序/有序/嵌套列表、
表格、分隔线 `-----`。

## 路由

| 路径 | 说明 |
|------|------|
| `/` | 首页 |
| `/blog.html` `/blog` | 博客列表 |
| `/post.html?p=2025/08/slug` 或 `/p/2025/08/slug` | 文章详情 |
| `/tags.html` `/tags`（`?t=tag` 过滤） | 标签 |
| `/about.html` `/about` | 关于（渲染 `blog/about.org`） |
| `/rss.xml` | RSS 订阅 |
| `/assets/...` | 图片资源 |

API：`/api/posts`、`/api/post?p=...`、`/api/tags`、`/api/about`。

## 部署

规划中：完善后部署到 GitHub Pages（CI 生成静态站点）或自建服务器（如腾讯云）。
