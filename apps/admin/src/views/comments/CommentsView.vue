<script setup lang="ts">
import { NAlert, NButton, NSpace, NText } from 'naive-ui';
import { computed, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import {
  useForumApi,
  type Comment,
  type CommentStatusFilter,
  type CommentStatusName,
} from '@/api';

import CommentFilters from './CommentFilters.vue';
import CommentList from './CommentList.vue';
import CommentModerationModal, {
  type CommentModerationTarget,
} from './CommentModerationModal.vue';

const PAGE_SIZE = 20;
const api = useForumApi();
const route = useRoute();
const router = useRouter();
const postIdInput = ref<number | null>(null);
const activePostId = ref<number>();
const queryInput = ref('');
const authorNameInput = ref('');
const statusInput = ref<CommentStatusFilter | ''>('');
const activeQuery = ref('');
const activeAuthorName = ref('');
const activeStatus = ref<CommentStatusFilter | ''>('');
const hasFilters = computed(() =>
  Boolean(
    activePostId.value ||
    activeQuery.value ||
    activeAuthorName.value ||
    activeStatus.value,
  ),
);
const comments = ref<Comment[]>([]);
const total = ref(0);
const page = ref(1);
const loading = ref(false);
const moderationSaving = ref(false);
const errorMessage = ref('');
const successMessage = ref('');
const actionErrorMessage = ref('');
const pendingAction = ref<CommentModerationTarget | null>(null);
let requestId = 0;

async function loadComments() {
  const currentRequestId = ++requestId;
  loading.value = true;
  errorMessage.value = '';
  try {
    const result = await api.getComments({
      postId: activePostId.value,
      page: page.value,
      pageSize: PAGE_SIZE,
      query: activeQuery.value,
      authorName: activeAuthorName.value,
      status: activeStatus.value || undefined,
    });
    if (currentRequestId !== requestId) return;
    if (result.total > 0 && page.value > Math.ceil(result.total / PAGE_SIZE)) {
      page.value = Math.ceil(result.total / PAGE_SIZE);
      await loadComments();
      return;
    }
    comments.value = result.items;
    total.value = result.total;
  } catch (error) {
    if (currentRequestId !== requestId) return;
    comments.value = [];
    total.value = 0;
    errorMessage.value = error instanceof Error ? error.message : String(error);
  } finally {
    if (currentRequestId === requestId) loading.value = false;
  }
}

function search() {
  const query = {
    q: queryInput.value.trim() || undefined,
    author: authorNameInput.value.trim() || undefined,
    post: postIdInput.value == null ? undefined : String(postIdInput.value),
    status: statusInput.value || undefined,
  };
  if (
    ['q', 'author', 'post', 'status'].every(
      (key) => route.query[key] === query[key as keyof typeof query],
    )
  ) {
    page.value = 1;
    void loadComments();
  } else {
    void router.replace({ query });
  }
}

function resetFilters() {
  queryInput.value = '';
  authorNameInput.value = '';
  statusInput.value = '';
  postIdInput.value = null;
  search();
}

function filterPost(postId: number) {
  postIdInput.value = postId;
  search();
}

function changePage(nextPage: number) {
  page.value = nextPage;
  void loadComments();
}

function requestModeration(comment: Comment, status: CommentStatusName) {
  actionErrorMessage.value = '';
  successMessage.value = '';
  pendingAction.value = { comment, status };
}

async function handleModerationSuccess(message: string) {
  pendingAction.value = null;
  successMessage.value = message;
  await loadComments();
}

watch(
  () => route.query,
  (query) => {
    const text = (value: unknown) => (typeof value === 'string' ? value : '');
    const postId = Number(text(query.post));
    activePostId.value =
      Number.isSafeInteger(postId) && postId > 0 ? postId : undefined;
    postIdInput.value = activePostId.value ?? null;
    activeQuery.value = queryInput.value = text(query.q).trim();
    activeAuthorName.value = authorNameInput.value = text(query.author).trim();
    const status = text(query.status);
    activeStatus.value = statusInput.value =
      status === '0' || status === '1' || status === '2' ? status : '';
    page.value = 1;
    void loadComments();
  },
  { immediate: true },
);
</script>

<template>
  <n-space vertical :size="16" class="comments-page">
    <CommentFilters
      v-model:post-id="postIdInput"
      v-model:query="queryInput"
      v-model:author-name="authorNameInput"
      v-model:status="statusInput"
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
    <n-alert v-if="errorMessage" type="error" title="评论加载失败">
      <n-space vertical size="small">
        <n-text>{{ errorMessage }}</n-text>
        <n-button size="small" @click="loadComments">重新加载</n-button>
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

    <CommentList
      v-if="!errorMessage"
      :comments="comments"
      :loading="loading"
      :total="total"
      :page="page"
      :page-size="PAGE_SIZE"
      :post-id="activePostId"
      :has-filters="hasFilters"
      :actions-disabled="moderationSaving"
      @update-page="changePage"
      @moderate="requestModeration"
      @filter-post="filterPost"
      @reset-filters="resetFilters"
    />

    <CommentModerationModal
      v-model:saving="moderationSaving"
      :target="pendingAction"
      @close="pendingAction = null"
      @success="handleModerationSuccess"
      @error="actionErrorMessage = $event"
    />
  </n-space>
</template>

<style scoped>
.comments-page {
  max-width: 1000px;
  margin-inline: auto;
}
</style>
