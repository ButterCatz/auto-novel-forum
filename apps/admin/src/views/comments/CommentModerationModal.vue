<script setup lang="ts">
import { NModal } from 'naive-ui';

import { useForumApi, type Comment, type CommentStatusName } from '@/api';

export interface CommentModerationTarget {
  comment: Comment;
  status: CommentStatusName;
}

const props = defineProps<{ target: CommentModerationTarget | null }>();
const emit = defineEmits<{
  close: [];
  success: [message: string];
  error: [message: string];
}>();

const api = useForumApi();
const saving = defineModel<boolean>('saving', { default: false });

function close() {
  if (!saving.value) emit('close');
}

async function confirm() {
  if (!props.target || saving.value) return;
  saving.value = true;
  try {
    await api.setCommentStatus(props.target.comment.id, props.target.status);
    emit(
      'success',
      props.target.status === 'published'
        ? '评论已恢复'
        : props.target.status === 'hidden'
          ? '评论已隐藏'
          : '评论已删除',
    );
  } catch (error) {
    emit('error', error instanceof Error ? error.message : String(error));
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <n-modal
    :show="target !== null"
    preset="dialog"
    type="warning"
    :title="
      target?.status === 'published'
        ? '确认恢复评论'
        : target?.status === 'hidden'
          ? '确认隐藏评论'
          : '确认删除评论'
    "
    positive-text="确认处理"
    negative-text="取消"
    :loading="saving"
    :mask-closable="!saving"
    :close-on-esc="!saving"
    :closable="!saving"
    @close="close"
    @positive-click="confirm"
    @negative-click="close"
    @mask-click="close"
  >
    <p>{{ target?.comment.authorUsername }} 的评论 #{{ target?.comment.id }}</p>
    {{
      target?.status === 'published'
        ? '恢复后，评论原文将重新公开展示。'
        : '处理后，公开页面将显示占位；管理员仍可查看原文并恢复。'
    }}
  </n-modal>
</template>
