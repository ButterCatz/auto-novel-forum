# @novelia/forum-api

论坛服务的浏览器端 TypeScript 接口包。目前提供附属到第三方资源的评论接口。调用方传入
配置好认证策略的 [ky](https://github.com/sindresorhus/ky) client 和论坛服务地址；包内会派生
一个使用 `/api/v1/` 前缀的 client。

```ts
import { createForumApi } from '@novelia/forum-api';

const forumApi = createForumApi({
  client,
  url: 'https://forum.example.com/',
  type: 'novel',
});

const page = await forumApi.getComments('chapter-123', {
  page: 1,
  pageSize: 20,
});

const replies = await forumApi.getReplies('chapter-123', page.items[0].id, {
  page: 1,
  pageSize: 20,
});
```
