<script setup lang="ts">
import { useBookmarkStore } from '../stores/bookmarks'

const store = useBookmarkStore()
</script>

<template>
  <section class="page settings-page split-page">
    <section class="surface">
      <div class="section-head">
        <div>
          <span class="eyebrow">Data</span>
          <h2>数据概览</h2>
        </div>
      </div>
      <dl class="detail-list">
        <div>
          <dt>书签数量</dt>
          <dd>{{ store.total }}</dd>
        </div>
        <div>
          <dt>分类数量</dt>
          <dd>{{ store.folders.length }}</dd>
        </div>
        <div>
          <dt>标签数量</dt>
          <dd>{{ store.tags.length }}</dd>
        </div>
      </dl>
    </section>

    <section class="surface">
      <div class="section-head">
        <div>
          <span class="eyebrow">Shortcuts</span>
          <h2>快捷键</h2>
        </div>
      </div>
      <dl class="detail-list">
        <div>
          <dt>全局搜索</dt>
          <dd>Ctrl + K</dd>
        </div>
      </dl>
    </section>

    <section class="surface">
      <div class="section-head">
        <div>
          <span class="eyebrow">About</span>
          <h2>关于</h2>
        </div>
      </div>
      <p class="hint">Cola Bookmarks 是本地优先的桌面书签管理器。</p>
    </section>
  </section>
</template>
