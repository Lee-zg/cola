<!-- 文件说明：frontend/src/pages/TaxonomyPage.vue，对应当前模块的界面或交互逻辑。 -->
<script setup lang="ts">
import { NAlert, NButton, NCard, NEmpty, NIcon, NList, NListItem, NSpace, NTag } from 'naive-ui'
import { useTaxonomy } from '../composables/useTaxonomy'
import { appIcons } from '../icons'

const taxonomy = useTaxonomy()
</script>

<template>
  <section class="page taxonomy-page split-page">
    <NCard class="taxonomy-card" :bordered="false">
      <template #header>
        <div>
          <span class="eyebrow">FOLDERS</span>
          <h2>分类</h2>
        </div>
      </template>
      <NList v-if="taxonomy.folders.value.length" hoverable clickable>
        <NListItem v-for="folder in taxonomy.folders.value" :key="folder" @click="taxonomy.selectFolder(folder)">
          <NSpace align="center">
            <NIcon :component="appIcons.folder" />
            <strong>{{ folder }}</strong>
          </NSpace>
          <template #suffix>
            <NButton size="small" secondary>筛选</NButton>
          </template>
        </NListItem>
      </NList>
      <NEmpty v-else description="暂无分类" />
    </NCard>

    <NCard class="taxonomy-card" :bordered="false">
      <template #header>
        <div>
          <span class="eyebrow">TAGS</span>
          <h2>标签</h2>
        </div>
      </template>
      <NSpace v-if="taxonomy.tags.value.length" wrap>
        <NTag v-for="tag in taxonomy.tags.value" :key="tag" round type="info" checkable @click="taxonomy.selectTag(tag)">
          {{ tag }}
        </NTag>
      </NSpace>
      <NEmpty v-else description="暂无标签" />
      <NAlert class="taxonomy-alert" type="info" :show-icon="false">
        当前后端只提供分类和标签读取能力，重命名、合并、删除可在后续加入专用 API 后接入。
      </NAlert>
    </NCard>
  </section>
</template>
