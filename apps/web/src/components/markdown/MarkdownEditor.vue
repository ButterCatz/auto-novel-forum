<script setup lang="ts">
import {
  LinkOutlined,
  StarBorderOutlined,
  UnfoldMoreOutlined,
  VisibilityOffOutlined,
} from '@vicons/material';
import { nextTick, onMounted, ref, useTemplateRef, watch } from 'vue';

import { XButton } from '@novelia/web-kit';

import type { MarkdownMode } from '@novelia/forum-api';
import MarkdownContent from './MarkdownContent.vue';

const props = withDefaults(
  defineProps<{
    mode: MarkdownMode;
    placeholder?: string;
    disabled?: boolean;
    describedBy?: string;
    invalid?: boolean;
    maxlength?: number;
    rows?: number;
  }>(),
  {
    placeholder: '使用 Markdown 输入内容…',
    disabled: false,
    maxlength: 100000,
    rows: 7,
  },
);

const value = defineModel<string>({ required: true });
const textarea = useTemplateRef<HTMLTextAreaElement>('textarea');
const activeTab = ref<'edit' | 'preview'>('edit');
const editorTabClass =
  'relative min-w-16 border-r border-border px-[0.9rem] py-[0.55rem] text-[0.8125rem] font-semibold';

function resizeCommentEditor() {
  const element = textarea.value;
  if (!element || props.mode !== 'comment') return;
  element.style.height = 'auto';
  element.style.height = `${element.scrollHeight}px`;
}

onMounted(async () => {
  await nextTick();
  resizeCommentEditor();
});

watch(value, () => void nextTick(resizeCommentEditor));
watch(activeTab, (tab) => {
  if (tab === 'edit') void nextTick(resizeCommentEditor);
});

async function restoreSelection(start: number, end: number) {
  await nextTick();
  textarea.value?.focus();
  textarea.value?.setSelectionRange(start, end);
}

function wrapSelection(prefix: string, suffix: string, placeholder: string) {
  const element = textarea.value;
  if (!element || props.disabled) return;
  const start = element.selectionStart;
  const end = element.selectionEnd;
  const selected = value.value.slice(start, end) || placeholder;
  value.value = `${value.value.slice(0, start)}${prefix}${selected}${suffix}${value.value.slice(end)}`;
  void restoreSelection(
    start + prefix.length,
    start + prefix.length + selected.length,
  );
}

function insertBlock(prefix: string, suffix: string, placeholder: string) {
  const element = textarea.value;
  if (!element || props.disabled) return;
  const start = element.selectionStart;
  const end = element.selectionEnd;
  const before = value.value.slice(0, start);
  const after = value.value.slice(end);
  const selected = value.value.slice(start, end) || placeholder;
  const leadingBreak = before && !before.endsWith('\n') ? '\n' : '';
  const trailingBreak = after && !after.startsWith('\n') ? '\n' : '';
  const insertion = `${leadingBreak}${prefix}${selected}${suffix}${trailingBreak}`;
  value.value = `${before}${insertion}${after}`;
  const selectionStart = start + leadingBreak.length + prefix.length;
  void restoreSelection(selectionStart, selectionStart + selected.length);
}

function focus() {
  activeTab.value = 'edit';
  void nextTick(() => textarea.value?.focus());
}

defineExpose({ focus });
</script>

<template>
  <div class="overflow-hidden rounded-md border border-border">
    <div class="flex flex-wrap items-stretch border-b border-border">
      <div class="flex flex-none" role="tablist" aria-label="Markdown 编辑模式">
        <XButton
          variant="plain"
          size="none"
          :class="[
            editorTabClass,
            activeTab === 'edit' ? '-mb-px text-ink' : 'text-muted',
          ]"
          role="tab"
          :aria-selected="activeTab === 'edit'"
          @click="activeTab = 'edit'"
        >
          编辑
        </XButton>
        <XButton
          variant="plain"
          size="none"
          :class="[
            editorTabClass,
            activeTab === 'preview' ? '-mb-px text-ink' : 'text-muted',
          ]"
          role="tab"
          :aria-selected="activeTab === 'preview'"
          @click="activeTab = 'preview'"
        >
          预览
        </XButton>
      </div>

      <div
        v-if="activeTab === 'edit'"
        class="order-last flex w-full items-center gap-0.5 overflow-x-auto border-t border-border p-1 sm:order-none sm:ml-auto sm:w-auto sm:border-t-0"
        aria-label="Markdown 格式工具"
      >
        <XButton
          variant="toolbar"
          size="icon-sm"
          class="font-bold"
          title="粗体"
          aria-label="粗体"
          @mousedown.prevent="wrapSelection('**', '**', '粗体')"
        >
          B
        </XButton>
        <XButton
          variant="toolbar"
          size="icon-sm"
          class="italic"
          title="斜体"
          aria-label="斜体"
          @mousedown.prevent="wrapSelection('*', '*', '斜体')"
        >
          I
        </XButton>
        <XButton
          variant="toolbar"
          size="icon-sm"
          class="line-through"
          title="删除线"
          aria-label="删除线"
          @mousedown.prevent="wrapSelection('~~', '~~', '删除线')"
        >
          S
        </XButton>
        <XButton
          variant="toolbar"
          size="icon-sm"
          title="链接"
          aria-label="链接"
          @mousedown.prevent="wrapSelection('[', '](https://)', '链接文字')"
        >
          <LinkOutlined class="mx-auto size-4" aria-hidden="true" />
        </XButton>
        <XButton
          variant="toolbar"
          size="icon-sm"
          title="剧透"
          aria-label="剧透"
          @mousedown.prevent="wrapSelection('!!', '!!', '剧透内容')"
        >
          <VisibilityOffOutlined class="mx-auto size-4" aria-hidden="true" />
        </XButton>
        <XButton
          variant="toolbar"
          size="icon-sm"
          title="评分"
          aria-label="评分"
          @mousedown.prevent="insertBlock('::: star 5', '', '')"
        >
          <StarBorderOutlined class="mx-auto size-4" aria-hidden="true" />
        </XButton>
        <XButton
          variant="toolbar"
          size="icon-sm"
          title="折叠内容"
          aria-label="折叠内容"
          @mousedown.prevent="
            insertBlock('::: details 点击展开\n', '\n:::', '折叠内容')
          "
        >
          <UnfoldMoreOutlined class="mx-auto size-4" aria-hidden="true" />
        </XButton>
      </div>
    </div>

    <div v-show="activeTab === 'edit'" role="tabpanel">
      <textarea
        ref="textarea"
        v-model="value"
        class="block min-h-24 w-full border-0 bg-transparent px-3 py-3 text-sm leading-6 text-ink outline-none placeholder:text-muted/70 disabled:cursor-not-allowed disabled:opacity-50"
        :class="mode === 'comment' ? 'resize-none overflow-hidden' : 'resize-y'"
        :rows="rows"
        :maxlength="maxlength"
        :placeholder="placeholder"
        :disabled="disabled"
        :aria-describedby="describedBy"
        :aria-invalid="invalid || undefined"
        spellcheck="false"
      />
    </div>

    <div v-if="activeTab === 'preview'" class="min-h-48 p-4" role="tabpanel">
      <MarkdownContent v-if="value.trim()" :mode="mode" :source="value" />
      <p v-else class="text-sm text-muted">没有可预览的内容</p>
    </div>
  </div>
</template>
