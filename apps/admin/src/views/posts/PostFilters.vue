<script setup lang="ts">
import { SearchOutlined } from '@vicons/material';
import { NIcon, NInput, NSelect } from 'naive-ui';
import { computed } from 'vue';

import FilterChoiceGroup from '@/components/FilterChoiceGroup.vue';
import FilterRow from '@/components/FilterRow.vue';
import type { CategoryListItem, PostSort, PostStatusFilter } from '@/api';
import { categoryOrder, categoryTitle } from '@/category';

const props = defineProps<{ categories: CategoryListItem[] }>();
const emit = defineEmits<{ search: [] }>();

const query = defineModel<string>('query', { required: true });
const category = defineModel<string>('category', { required: true });
const status = defineModel<PostStatusFilter | ''>('status', { required: true });
const tagId = defineModel<number | null>('tagId', { required: true });
const authorName = defineModel<string>('authorName', { required: true });
const sort = defineModel<PostSort>('sort', { required: true });

const statusOptions = [
  { label: '全部', value: '' },
  { label: '正常发布', value: '0' },
  { label: '隐藏', value: '1' },
  { label: '删除', value: '2' },
];

const categoryOptions = computed(() => [
  { label: '全部', value: '' },
  ...[...props.categories]
    .sort((left, right) => categoryOrder(left.slug) - categoryOrder(right.slug))
    .map((item) => ({
      label: categoryTitle(item.slug),
      value: item.slug,
    })),
]);
const tagOptions = computed(() =>
  [...props.categories]
    .sort((left, right) => categoryOrder(left.slug) - categoryOrder(right.slug))
    .filter((item) => !category.value || item.slug === category.value)
    .flatMap((item) =>
      item.tags.map((tag) => ({
        label: `${tag.name} · ${categoryTitle(item.slug)}`,
        value: tag.id,
      })),
    ),
);
const sortOptions: { label: string; value: PostSort }[] = [
  { label: '最近活跃', value: 'active' },
  { label: '最新发布', value: 'newest' },
  { label: '浏览最多', value: 'views' },
  { label: '评论最多', value: 'comments' },
];

function changeCategory(value: string) {
  category.value = value;
  if (
    tagId.value != null &&
    !props.categories.some(
      (item) =>
        (value === '' || item.slug === value) &&
        item.tags.some((tag) => tag.id === tagId.value),
    )
  )
    tagId.value = null;
  emit('search');
}

function changeStatus(value: string) {
  status.value = value as PostStatusFilter | '';
  emit('search');
}

function changeSort(value: string) {
  sort.value = value as PostSort;
  emit('search');
}
</script>

<template>
  <div class="filters">
    <FilterRow label="搜索">
      <n-input
        v-model:value="query"
        class="query-input"
        clearable
        placeholder="搜索标题或正文"
        @change="emit('search')"
      >
        <template #suffix>
          <n-icon :component="SearchOutlined" />
        </template>
      </n-input>
    </FilterRow>

    <FilterRow label="作者">
      <n-input
        v-model:value="authorName"
        class="query-input"
        clearable
        placeholder="搜索作者名字"
        @change="emit('search')"
      />
    </FilterRow>

    <FilterRow label="标签">
      <n-select
        v-model:value="tagId"
        class="query-input"
        :options="tagOptions"
        clearable
        filterable
        placeholder="全部标签"
        @update:value="emit('search')"
      />
    </FilterRow>
    <FilterRow label="分类">
      <FilterChoiceGroup
        :value="category"
        :options="categoryOptions"
        @update:value="changeCategory"
      />
    </FilterRow>
    <FilterRow label="状态">
      <FilterChoiceGroup
        :value="status"
        :options="statusOptions"
        @update:value="changeStatus"
      />
    </FilterRow>
    <FilterRow label="排序">
      <FilterChoiceGroup
        :value="sort"
        :options="sortOptions"
        @update:value="changeSort"
      />
    </FilterRow>
  </div>
</template>

<style scoped>
.filters {
  display: grid;
  width: 100%;
  gap: 12px;
}

.query-input {
  width: min(400px, 100%);
}
</style>
