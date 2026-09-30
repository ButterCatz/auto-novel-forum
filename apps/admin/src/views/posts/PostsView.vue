<script setup lang="ts">
import { NAlert, NButton, NSpace, NText } from 'naive-ui';
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';

import {
  useForumApi,
  type CategoryListItem,
  type PostSort,
  type PostStatus,
  type PostSummary,
} from '@/api';
import { categoryTitle } from '@/category';

import PostFilters from './PostFilters.vue';
import PostList from './PostList.vue';

const PAGE_SIZE = 20;
const api = useForumApi();
const router = useRouter();
const categories = ref<CategoryListItem[]>([]);
const posts = ref<PostSummary[]>([]);
const loading = ref(true);
const busyPostId = ref<number | null>(null);
const errorMessage = ref('');
const actionErrorMessage = ref('');
const successMessage = ref('');
const page = ref(1);
const total = ref(0);
const queryInput = ref('');
const categoryInput = ref('');
const statusInput = ref('');
const tagIdInput = ref<number | null>(null);
const authorNameInput = ref('');
const sortInput = ref<PostSort>('active');
const query = ref('');
const category = ref('');
const status = ref('');
const tagId = ref<number | null>(null);
const authorName = ref('');
const sort = ref<PostSort>('active');
let requestId = 0;

const categoryMap = computed(
  () =>
    new Map(
      categories.value.map((item) => [item.id, categoryTitle(item.slug)]),
    ),
);
const hasFilters = computed(() =>
  Boolean(
    query.value ||
    category.value ||
    status.value ||
    tagId.value ||
    authorName.value,
  ),
);

async function loadPosts() {
  const currentRequestId = ++requestId;
  loading.value = true;
  errorMessage.value = '';
  try {
    const result = await api.getPosts({
      page: page.value,
      pageSize: PAGE_SIZE,
      query: query.value,
      category: category.value,
      status: status.value,
      tagId: tagId.value,
      authorName: authorName.value,
      sort: sort.value,
    });
    if (currentRequestId !== requestId) return;
    posts.value = result.items;
    total.value = result.total;
  } catch (error) {
    if (currentRequestId !== requestId) return;
    posts.value = [];
    total.value = 0;
    errorMessage.value = error instanceof Error ? error.message : String(error);
  } finally {
    if (currentRequestId === requestId) loading.value = false;
  }
}

function search() {
  query.value = queryInput.value.trim();
  category.value = categoryInput.value;
  status.value = statusInput.value;
  tagId.value = tagIdInput.value;
  authorName.value = authorNameInput.value.trim();
  sort.value = sortInput.value;
  page.value = 1;
  void loadPosts();
}

function resetFilters() {
  queryInput.value = '';
  categoryInput.value = '';
  statusInput.value = '';
  tagIdInput.value = null;
  authorNameInput.value = '';
  sortInput.value = 'active';
  search();
}

function changePage(nextPage: number) {
  page.value = nextPage;
  void loadPosts();
}

function reviewComments(post: PostSummary) {
  void router.push({ name: 'comments', query: { post: post.id } });
}

async function runAction(
  post: PostSummary,
  action: () => Promise<string>,
  message: string,
) {
  if (busyPostId.value !== null) return;
  busyPostId.value = post.id;
  actionErrorMessage.value = '';
  successMessage.value = '';
  try {
    await action();
    successMessage.value = `帖子「${post.title}」${message}`;
    await loadPosts();
  } catch (error) {
    actionErrorMessage.value =
      error instanceof Error ? error.message : String(error);
  } finally {
    busyPostId.value = null;
  }
}

function setStatus(post: PostSummary, status: PostStatus) {
  void runAction(post, () => api.setPostStatus(post.id, status), '状态已更新');
}

function setCommentsLocked(post: PostSummary, locked: boolean) {
  void runAction(
    post,
    () => (locked ? api.lockPost(post.id) : api.unlockPost(post.id)),
    locked ? '评论已锁定' : '评论已解锁',
  );
}

function setPinOrder(post: PostSummary, pinOrder: number | null) {
  void runAction(
    post,
    () =>
      pinOrder == null
        ? api.unpinPost(post.id)
        : api.pinPost(post.id, pinOrder),
    pinOrder == null ? '已取消置顶' : '置顶顺序已更新',
  );
}

async function initialize() {
  try {
    categories.value = await api.getCategories();
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : String(error);
  }
  await loadPosts();
}

onMounted(initialize);
</script>

<template>
  <n-space vertical :size="16" class="posts-page">
    <PostFilters
      v-model:query="queryInput"
      v-model:category="categoryInput"
      v-model:status="statusInput"
      v-model:tag-id="tagIdInput"
      v-model:author-name="authorNameInput"
      v-model:sort="sortInput"
      :categories="categories"
      @search="search"
    />

    <n-alert
      v-if="successMessage"
      type="success"
      closable
      @close="successMessage = ''"
    >
      {{ successMessage }}
    </n-alert>
    <n-alert v-if="errorMessage" type="error" title="帖子列表加载失败">
      <n-space vertical size="small">
        <n-text>{{ errorMessage }}</n-text>
        <n-button size="small" @click="loadPosts">重新加载</n-button>
      </n-space>
    </n-alert>
    <n-alert
      v-if="actionErrorMessage"
      type="error"
      closable
      @close="actionErrorMessage = ''"
    >
      {{ actionErrorMessage }}
    </n-alert>

    <PostList
      v-if="!errorMessage"
      :posts="posts"
      :loading="loading"
      :total="total"
      :page="page"
      :page-size="PAGE_SIZE"
      :has-filters="hasFilters"
      :category-names="categoryMap"
      :busy-post-id="busyPostId"
      @update-page="changePage"
      @reset-filters="resetFilters"
      @set-status="setStatus"
      @set-comments-locked="setCommentsLocked"
      @set-pin-order="setPinOrder"
      @review-comments="reviewComments"
    />
  </n-space>
</template>

<style scoped>
.posts-page {
  max-width: 1000px;
  margin-inline: auto;
}
</style>
