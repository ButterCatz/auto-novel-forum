<script setup lang="ts">
import { computed, nextTick, ref } from 'vue';

import type { PostComment } from '@/api';
import { useCommentReplyPageQuery } from '@/stores/comment';
import XButton from '@/ui/XButton.vue';
import XPagination from '@/ui/XPagination.vue';

import CommentListItem from './CommentListItem.vue';

const REPLY_PAGE_SIZE = 20;

const props = defineProps<{
  comment: PostComment;
  locked: boolean;
  postId: number;
  replyToId?: number;
}>();

const emit = defineEmits<{
  reply: [comment: PostComment];
  cancelReply: [];
  created: [comment: PostComment];
  statusChanged: [id: number, status: number];
  authorCommentsDeleted: [];
}>();

const expanded = ref(false);
const replyPage = ref(1);
const { replies, total, loading, error, refresh, retry } =
  useCommentReplyPageQuery(
    () => props.postId,
    () => props.comment.id,
    replyPage,
    REPLY_PAGE_SIZE,
    expanded,
  );
const totalPages = computed(() =>
  Math.max(1, Math.ceil(total.value / REPLY_PAGE_SIZE)),
);

function toggleReplies() {
  expanded.value = !expanded.value;
}

async function handleCreated(comment: PostComment) {
  const lastPage = Math.max(
    1,
    Math.ceil((props.comment.replyCount + 1) / REPLY_PAGE_SIZE),
  );
  emit('created', comment);
  if (comment.rootId !== props.comment.id) return;
  expanded.value = true;
  replyPage.value = lastPage;
  await nextTick();
  await refresh();
}
</script>

<template>
  <section>
    <CommentListItem
      :comment="comment"
      :locked="locked"
      :post-id="postId"
      :replying="replyToId === comment.id"
      @reply="emit('reply', $event)"
      @cancel-reply="emit('cancelReply')"
      @created="handleCreated"
      @status-changed="(id, status) => emit('statusChanged', id, status)"
      @author-comments-deleted="emit('authorCommentsDeleted')"
    />

    <div v-if="comment.replyCount > 0" class="pb-3 pl-6 sm:pl-10">
      <XButton
        variant="plain"
        size="none"
        class="text-xs"
        @click="toggleReplies"
      >
        {{ expanded ? '收起回复' : `查看 ${comment.replyCount} 条回复` }}
      </XButton>
    </div>

    <div v-if="expanded" class="border-t border-divider/60">
      <p v-if="loading" class="py-4 pl-10 text-sm text-muted">正在加载回复…</p>
      <div v-else-if="error" class="flex items-center gap-3 py-4 pl-10 text-sm">
        <span class="text-red-600">{{ error }}</span>
        <XButton variant="outline" size="xs" @click="retry">重试</XButton>
      </div>
      <template v-else>
        <CommentListItem
          v-for="reply in replies"
          :key="reply.id"
          :comment="reply"
          :locked="locked"
          :post-id="postId"
          :replying="replyToId === reply.id"
          @reply="emit('reply', $event)"
          @cancel-reply="emit('cancelReply')"
          @created="handleCreated"
          @status-changed="(id, status) => emit('statusChanged', id, status)"
          @author-comments-deleted="emit('authorCommentsDeleted')"
        />
        <XPagination
          v-if="totalPages > 1"
          :page="replyPage"
          :total-pages="totalPages"
          @change="replyPage = $event"
        />
      </template>
    </div>
  </section>
</template>
