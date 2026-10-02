# @novelia/forum-api

论坛服务的浏览器端 TypeScript 接口包。目前提供附属到第三方资源的评论接口。调用方传入
配置好认证策略的 [ky](https://github.com/sindresorhus/ky) client 和论坛服务地址；包内会派生
一个使用 `/api/v1/` 前缀的 client。

```ts
import { createForumApi, subjectKeys } from '@novelia/forum-api';

const forumApi = createForumApi({
  client,
  url: 'https://forum.example.com/',
  type: 'novel',
});

const subjectKey = subjectKeys.novel.web('syosetu', 'n1234');
const page = await forumApi.getComments(subjectKey, {
  page: 1,
  pageSize: 20,
});

const replies = await forumApi.getReplies(subjectKey, page.items[0].id, {
  page: 1,
  pageSize: 20,
});
```

## 构造 subject key

构造函数按资源类型分组，返回未进行 URL 编码的 key，编码由 API 客户端处理：

```ts
subjectKeys.novel.web('syosetu', 'n1234'); // web-syosetu-n1234
subjectKeys.novel.wenku('507f1f77bcf86cd799439011'); // wenku-507f1f77bcf86cd799439011
```

新增服务时，在 `subjectKeys` 下增加对应的类型分组和构造函数。`CommentType` 从分组名称自动推导；服务端也需注册同名 kind 及其校验器。
