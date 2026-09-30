<script setup lang="ts">
import { DeleteOutlineOutlined, VisibilityOffOutlined } from '@vicons/material';
import { NButton, NIcon, NTag, NText } from 'naive-ui';

import type { Comment, CommentStatusName } from '@/api';

defineProps<{ comment: Comment; actionsDisabled: boolean }>();
const emit = defineEmits<{
  filterPost: [postId: number];
  moderate: [comment: Comment, status: CommentStatusName];
}>();

function formatDate(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value));
}
</script>

<template>
  <div class="comment-row">
    <div class="comment-header">
      <div class="comment-author">
        <n-text strong>{{ comment.authorUsername }}</n-text>
        <n-text depth="3" class="comment-meta">
          #{{ comment.id }} · {{ formatDate(comment.createdAt) }}
        </n-text>
      </div>
      <n-tag v-if="comment.rootId" size="small" :bordered="false">
        回复 #{{ comment.rootId }}
      </n-tag>
    </div>
    <div class="comment-context">
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
      <n-button
        text
        tag="a"
        :href="`/p/${comment.postId}`"
        target="_blank"
        rel="noopener noreferrer"
      >
        查看帖子 #{{ comment.postId }}
      </n-button>
      <n-button text type="primary" @click="emit('filterPost', comment.postId)">
        仅看此帖
      </n-button>
    </div>
    <n-text class="comment-content">{{ comment.content }}</n-text>
    <div class="comment-actions">
      <n-button
        v-if="comment.status !== 0"
        :disabled="actionsDisabled"
        size="small"
        secondary
        type="success"
        @click="emit('moderate', comment, 'published')"
      >
        恢复
      </n-button>
      <n-button
        :disabled="actionsDisabled"
        size="small"
        secondary
        v-if="comment.status !== 1"
        type="warning"
        @click="emit('moderate', comment, 'hidden')"
      >
        <template #icon>
          <n-icon :component="VisibilityOffOutlined" />
        </template>
        隐藏
      </n-button>
      <n-button
        :disabled="actionsDisabled"
        size="small"
        secondary
        v-if="comment.status !== 2"
        type="error"
        @click="emit('moderate', comment, 'deleted')"
      >
        <template #icon>
          <n-icon :component="DeleteOutlineOutlined" />
        </template>
        删除
      </n-button>
    </div>
  </div>
</template>

<style scoped>
.comment-row {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.comment-header,
.comment-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.comment-author {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.comment-meta {
  font-size: 12px;
}

.comment-content {
  line-height: 1.72;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}

.comment-context {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
}

.comment-actions {
  flex-wrap: wrap;
  justify-content: flex-end;
}
</style>
