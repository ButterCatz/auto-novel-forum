import { webKit } from './web-kit';

export const authApi = webKit.api;
export const authUser = webKit.profile;

const client = authApi.createClient(
  new URL('/api/v1/', window.location.origin).toString(),
);

export interface Page<T> {
  total: number;
  items: T[];
}

export interface PostTag {
  id: number;
  name: string;
  color: number;
}

export interface CategoryTag extends PostTag {
  sortOrder: number;
}

export interface CategoryListItem {
  id: number;
  slug: string;
  tags: CategoryTag[];
}

export interface PostSummary {
  id: number;
  categoryId: number;
  title: string;
  authorId: number;
  authorUsername: string;
  status: number;
  viewsCount: number;
  commentsCount: number;
  commentsLocked: boolean;
  pinOrder?: number;
  favorited: boolean;
  createdAt: string;
  updatedAt: string;
  activeAt: string;
  tags: PostTag[];
}

export interface Post extends PostSummary {
  content: string;
}

export type PostSort = 'active' | 'newest' | 'views' | 'comments';

export interface PostComment {
  id: number;
  postId: number;
  rootId: number | null;
  content: string;
  authorId: number;
  authorUsername: string;
  status: number;
  createdAt: string;
  updatedAt: string;
  replyCount?: number;
}

export function getPosts(
  params: {
    page: number;
    pageSize: number;
    category?: string;
    query?: string;
    tagIds?: number[];
    sort?: PostSort;
  },
  signal?: AbortSignal,
) {
  return client
    .get('post/', {
      searchParams: {
        page: params.page,
        page_size: params.pageSize,
        category: params.category,
        q: params.query,
        tag: params.tagIds?.join(','),
        sort: params.sort,
      },
      signal,
    })
    .json<Page<PostSummary>>();
}

export function getFavoritePosts(
  params: { page: number; pageSize: number },
  signal?: AbortSignal,
) {
  return client
    .get('me/favorite', {
      searchParams: {
        page: params.page,
        page_size: params.pageSize,
      },
      signal,
    })
    .json<Page<PostSummary>>();
}

export function getMyPosts(
  params: { page: number; pageSize: number },
  signal?: AbortSignal,
) {
  return client
    .get('me/post', {
      searchParams: {
        page: params.page,
        page_size: params.pageSize,
      },
      signal,
    })
    .json<Page<PostSummary>>();
}

export function getPost(id: number, signal?: AbortSignal) {
  return client.get(`post/${id}/`, { signal }).json<Post>();
}

export function setPostFavorite(id: number, favorited: boolean) {
  const path = `post/${id}/favorite`;
  return favorited ? client.put(path) : client.delete(path);
}

export function getCategories(signal?: AbortSignal) {
  return client.get('category/', { signal }).json<CategoryListItem[]>();
}

export function createPost(input: {
  categoryId: number;
  title: string;
  content: string;
  tagIds: number[];
}) {
  return client.post('post/', { json: input }).json<Post>();
}

export function updatePost(
  id: number,
  input: {
    categoryId: number;
    title: string;
    content: string;
    tagIds: number[];
  },
) {
  return client.patch(`post/${id}/`, { json: input }).json<Post>();
}

export function deletePost(id: number) {
  return client.delete(`post/${id}/`);
}

export function setPostStatus(id: number, status: number) {
  return client.put(`admin/post/${id}/status`, { json: { status } });
}

export function lockPost(id: number) {
  return client.put(`admin/post/${id}/lock`);
}

export function unlockPost(id: number) {
  return client.delete(`admin/post/${id}/lock`);
}

export function pinPost(id: number, pinOrder: number) {
  return client.put(`admin/post/${id}/pin`, { json: { pinOrder } });
}

export function unpinPost(id: number) {
  return client.delete(`admin/post/${id}/pin`);
}

export function getPostComments(
  id: number,
  params: { page: number; pageSize: number },
  signal?: AbortSignal,
) {
  return client
    .get(`post/${id}/comment`, {
      searchParams: {
        page: params.page,
        page_size: params.pageSize,
      },
      signal,
    })
    .json<Page<PostComment>>();
}

export function getPostCommentReplies(
  postId: number,
  rootId: number,
  params: { page: number; pageSize: number },
  signal?: AbortSignal,
) {
  return client
    .get(`post/${postId}/comment/${rootId}/reply`, {
      searchParams: {
        page: params.page,
        page_size: params.pageSize,
      },
      signal,
    })
    .json<Page<PostComment>>();
}

export function createPostComment(
  id: number,
  input: { content: string; rootId?: number },
) {
  return client.post(`post/${id}/comment`, { json: input }).json<PostComment>();
}

export function updatePostComment(id: number, content: string) {
  return client
    .patch(`comment/${id}`, { json: { content } })
    .json<PostComment>();
}

export function deletePostComment(id: number) {
  return client.delete(`comment/${id}`);
}

export function setPostCommentStatus(
  id: number,
  status: 'published' | 'hidden' | 'deleted',
) {
  return client.put(`admin/comment/${id}/status`, { json: { status } });
}

export function deleteCommentsByAuthor(authorId: number) {
  return client.delete(`admin/comment/author/${authorId}`);
}
