<script setup lang="ts">
import { useTaxonomy } from '../composables/useTaxonomy'

const taxonomy = useTaxonomy()
</script>

<template>
  <section class="page taxonomy-page split-page">
    <section class="surface">
      <div class="section-head">
        <div>
          <span class="eyebrow">Folders</span>
          <h2>分类</h2>
        </div>
      </div>
      <div class="taxonomy-list">
        <button
          v-for="folder in taxonomy.folders.value"
          :key="folder"
          class="taxonomy-row"
          type="button"
          @click="taxonomy.selectFolder(folder)"
        >
          <span>{{ folder }}</span>
          <small>筛选</small>
        </button>
        <div v-if="!taxonomy.folders.value.length" class="empty">暂无分类</div>
      </div>
    </section>

    <section class="surface">
      <div class="section-head">
        <div>
          <span class="eyebrow">Tags</span>
          <h2>标签</h2>
        </div>
      </div>
      <div class="tag-cloud roomy">
        <button v-for="tag in taxonomy.tags.value" :key="tag" class="tag large" type="button" @click="taxonomy.selectTag(tag)">
          {{ tag }}
        </button>
        <div v-if="!taxonomy.tags.value.length" class="empty inline-empty">暂无标签</div>
      </div>
      <p class="hint">当前后端只提供分类和标签读取能力，重命名、合并、删除可在后续加入专用 API 后接入。</p>
    </section>
  </section>
</template>
