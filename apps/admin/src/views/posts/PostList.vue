<script setup lang="ts">
import {
  NButton,
  NCard,
  NEmpty,
  NList,
  NListItem,
  NPagination,
  NSkeleton,
  NSpin,
  NText,
} from 'naive-ui';
import { computed } from 'vue';

import type { PostStatus, PostSummary } from '@/api';

import PostListItem from './PostListItem.vue';

const props = defineProps<{
  posts: PostSummary[];
  loading: boolean;
  total: number;
  page: number;
  pageSize: number;
  hasFilters: boolean;
  categoryNames: Map<number, string>;
  busyPostId: number | null;
}>();

const emit = defineEmits<{
  updatePage: [page: number];
  resetFilters: [];
  setStatus: [post: PostSummary, status: PostStatus];
  setCommentsLocked: [post: PostSummary, locked: boolean];
  setPinOrder: [post: PostSummary, pinOrder: number | null];
  reviewComments: [post: PostSummary];
}>();

const pageCount = computed(() =>
  Math.max(1, Math.ceil(props.total / props.pageSize)),
);
const rangeStart = computed(() =>
  props.total === 0 ? 0 : (props.page - 1) * props.pageSize + 1,
);
const rangeEnd = computed(() =>
  Math.min(props.page * props.pageSize, props.total),
);
</script>

<template>
  <div class="post-list">
    <n-text v-if="!loading || posts.length" depth="3" class="list-summary">
      <template v-if="total">
        显示 {{ rangeStart }}–{{ rangeEnd }}，共 {{ total }} 篇帖子
      </template>
      <template v-else>共 0 篇帖子</template>
    </n-text>

    <n-card class="list-card" content-style="padding: 0;">
      <div v-if="loading && !posts.length" class="skeleton-list">
        <div v-for="index in 5" :key="index" class="skeleton-row">
          <n-skeleton text :repeat="3" />
        </div>
      </div>
      <n-spin v-else :show="loading">
        <n-list v-if="posts.length">
          <n-list-item v-for="post in posts" :key="post.id">
            <PostListItem
              :post="post"
              :category-name="
                categoryNames.get(post.categoryId) ?? `分类 ${post.categoryId}`
              "
              :actions-disabled="busyPostId !== null || loading"
              :saving="busyPostId === post.id"
              @set-status="(item, status) => emit('setStatus', item, status)"
              @set-comments-locked="
                (item, locked) => emit('setCommentsLocked', item, locked)
              "
              @set-pin-order="
                (item, pinOrder) => emit('setPinOrder', item, pinOrder)
              "
              @review-comments="emit('reviewComments', $event)"
            />
          </n-list-item>
        </n-list>
        <n-empty
          v-else
          class="empty"
          :description="hasFilters ? '没有符合条件的帖子' : '暂无帖子'"
        >
          <template v-if="hasFilters" #extra>
            <n-button size="small" @click="emit('resetFilters')">
              清除筛选
            </n-button>
          </template>
        </n-empty>
      </n-spin>
    </n-card>

    <div v-if="total > pageSize" class="pagination">
      <n-pagination
        :page="page"
        :page-count="pageCount"
        :page-slot="5"
        show-quick-jumper
        @update:page="emit('updatePage', $event)"
      >
        <template #goto>跳至</template>
      </n-pagination>
    </div>
  </div>
</template>

<style scoped>
.post-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.list-summary {
  min-height: 28px;
  display: flex;
  align-items: center;
}

.list-card {
  border-right: 0;
  border-left: 0;
}

.skeleton-list {
  padding: 0 16px;
}

.skeleton-row {
  padding: 18px 0;
}

.skeleton-row + .skeleton-row {
  border-top: 1px solid var(--n-border-color);
}

:deep(.n-list-item) {
  padding: 14px 16px;
}

.empty {
  padding: 64px 16px;
}

.pagination {
  display: flex;
  overflow-x: auto;
}

@media (max-width: 767px) {
  :deep(.n-list-item) {
    padding: 12px 10px;
  }

  .skeleton-list {
    padding-inline: 10px;
  }
}
</style>
