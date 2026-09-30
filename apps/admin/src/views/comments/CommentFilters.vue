<script setup lang="ts">
import { NButton, NInput, NInputNumber } from 'naive-ui';
import FilterRow from '@/components/FilterRow.vue';
import FilterChoiceGroup from '@/components/FilterChoiceGroup.vue';
import type { CommentStatusFilter } from '@/api';

const emit = defineEmits<{ search: []; reset: [] }>();
const query = defineModel<string>('query', { required: true });
const authorName = defineModel<string>('authorName', { required: true });
const postId = defineModel<number | null>('postId', { required: true });
const status = defineModel<CommentStatusFilter | ''>('status', {
  required: true,
});
const statusOptions = [
  { label: '全部', value: '' },
  { label: '正常发布', value: '0' },
  { label: '隐藏', value: '1' },
  { label: '删除', value: '2' },
];
function changeStatus(value: string) {
  status.value = value as CommentStatusFilter | '';
  emit('search');
}
</script>

<template>
  <div class="filters">
    <FilterRow label="搜索">
      <n-input
        v-model:value="query"
        class="field-input"
        clearable
        placeholder="搜索评论内容"
        @change="emit('search')"
      />
    </FilterRow>
    <FilterRow label="作者">
      <n-input
        v-model:value="authorName"
        class="field-input"
        clearable
        placeholder="搜索作者名字"
        @change="emit('search')"
      />
    </FilterRow>
    <FilterRow label="帖子">
      <n-input-number
        v-model:value="postId"
        class="field-input"
        :min="1"
        :max="Number.MAX_SAFE_INTEGER"
        :precision="0"
        :show-button="false"
        clearable
        placeholder="全部帖子，可输入帖子 ID"
        @change="emit('search')"
      />
    </FilterRow>
    <FilterRow label="状态">
      <FilterChoiceGroup
        :value="status"
        :options="statusOptions"
        @update:value="changeStatus"
      />
    </FilterRow>
    <div>
      <n-button size="small" quaternary @click="emit('reset')">
        重置筛选
      </n-button>
    </div>
  </div>
</template>

<style scoped>
.filters {
  display: grid;
  gap: 12px;
}
.field-input {
  width: min(400px, 100%);
}
</style>
