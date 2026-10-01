<script setup lang="ts">
import { MoreHorizOutlined } from '@vicons/material';
import {
  NButton,
  NDropdown,
  NIcon,
  NTag,
  NText,
  type DropdownOption,
} from 'naive-ui';
import { computed } from 'vue';

import type { Comment, CommentStatusName } from '@/api';

const props = defineProps<{ comment: Comment; actionsDisabled: boolean }>();
const emit = defineEmits<{
  filterPost: [postId: number];
  moderate: [comment: Comment, status: CommentStatusName];
}>();

const actionOptions = computed<DropdownOption[]>(() => {
  const options: DropdownOption[] = [{ label: '仅看此帖', key: 'filter-post' }];
  if (props.comment.status !== 0)
    options.push({
      label: '恢复',
      key: 'published',
      disabled: props.actionsDisabled,
    });
  if (props.comment.status !== 1)
    options.push({
      label: '隐藏',
      key: 'hidden',
      disabled: props.actionsDisabled,
    });
  if (props.comment.status !== 2)
    options.push({
      label: '删除',
      key: 'deleted',
      disabled: props.actionsDisabled,
    });
  return options;
});

function selectAction(key: string | number) {
  if (key === 'filter-post') {
    emit('filterPost', props.comment.postId);
    return;
  }
  if (props.actionsDisabled) return;
  emit('moderate', props.comment, key as CommentStatusName);
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value));
}

function formatCommentId(comment: Comment) {
  return comment.rootId
    ? `#${comment.id} → #${comment.rootId}`
    : `#${comment.id}`;
}
</script>

<template>
  <div class="comment-row">
    <div class="comment-header">
      <div class="comment-author">
        <n-text strong>{{ comment.authorUsername }}</n-text>
        <div class="comment-meta">
          <n-text depth="3">{{ formatCommentId(comment) }} ·</n-text>
          <n-button
            text
            type="primary"
            size="tiny"
            tag="a"
            :href="`/p/${comment.postId}`"
            target="_blank"
            rel="noopener noreferrer"
          >
            帖子 #{{ comment.postId }}
          </n-button>
          <n-text depth="3">· {{ formatDate(comment.createdAt) }}</n-text>
        </div>
      </div>
      <div class="comment-header-actions">
        <n-tag
          size="small"
          :bordered="false"
          :type="
            comment.status === 0
              ? 'success'
              : comment.status === 1
                ? 'warning'
                : 'error'
          "
        >
          {{ ['正常发布', '隐藏', '删除'][comment.status] }}
        </n-tag>
        <n-dropdown
          trigger="click"
          :options="actionOptions"
          @select="selectAction"
        >
          <n-button quaternary circle size="small" aria-label="评论操作">
            <template #icon>
              <n-icon :component="MoreHorizOutlined" />
            </template>
          </n-button>
        </n-dropdown>
      </div>
    </div>
    <n-text class="comment-content">{{ comment.content }}</n-text>
  </div>
</template>

<style scoped>
.comment-row {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.comment-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.comment-header-actions {
  display: flex;
  flex: none;
  align-items: center;
  gap: 6px;
}

.comment-author {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.comment-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  column-gap: 4px;
  font-size: 12px;
}

.comment-content {
  line-height: 1.72;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}
</style>
