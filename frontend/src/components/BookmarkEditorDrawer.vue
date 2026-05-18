<!-- 文件说明：frontend/src/components/BookmarkEditorDrawer.vue，对应当前模块的界面或交互逻辑。 -->
<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import {
  NAlert,
  NButton,
  NCheckbox,
  NDrawer,
  NDrawerContent,
  NDynamicTags,
  NForm,
  NFormItem,
  NIcon,
  NImage,
  NInput,
  NPopconfirm,
  NSpace,
  NTag,
  NTreeSelect,
  NUpload
} from 'naive-ui'
import type { TreeSelectOption, UploadCustomRequestOptions, UploadFileInfo, UploadSettledFileInfo } from 'naive-ui'
import { appIcons } from '../icons'
import type { Bookmark, BookmarkInput, CategoryNode, ThumbnailUploadInput } from '../types'

const props = defineProps<{
  open: boolean
  selected: Bookmark | null
  draft: BookmarkInput
  status: string
  categories: CategoryNode[]
}>()

const emit = defineEmits<{
  analyze: []
  close: []
  remove: []
  save: []
  'refresh-thumbnail': []
  'save-thumbnail': [input: ThumbnailUploadInput]
  'save-thumbnail-url': [url: string]
  'set-thumbnail-auto': [useAuto: boolean]
  'clear-custom-thumbnail': []
  'update:draft': [patch: Partial<BookmarkInput>]
}>()

const thumbnailUrl = ref('')
const thumbnailUploading = ref(false)
const thumbnailFileList = ref<UploadFileInfo[]>([])
const thumbnailImageNamePattern = /\.(gif|ico|jpe?g|png|svg|webp)$/i

const selectedThumbnail = computed(() => props.selected?.thumbnail)
const thumbnailDisplayPath = computed(() => selectedThumbnail.value?.displayPath || '')
const autoThumbnailPath = computed(() => selectedThumbnail.value?.autoThumbPath || selectedThumbnail.value?.autoFilePath || '')
const customThumbnailPath = computed(() => selectedThumbnail.value?.customThumbPath || selectedThumbnail.value?.customFilePath || '')
const thumbnailSourceLabel = computed(() => {
  const source = selectedThumbnail.value?.displaySource || ''
  if (source === 'og') return 'OpenGraph'
  if (source === 'favicon') return 'Favicon'
  if (source === 'screenshot') return '本地截图'
  if (source === 'upload') return '本地上传'
  if (source === 'remote') return '图片地址'
  return '尚未生成'
})

const updateTextField = (field: keyof Pick<BookmarkInput, 'title' | 'url' | 'description' | 'folder' | 'categoryId'>, value: string) => {
  emit('update:draft', { [field]: value })
}

const updateListField = (field: keyof Pick<BookmarkInput, 'tags' | 'keywords' | 'aliases'>, value: string[]) => {
  emit('update:draft', { [field]: value })
}

const toCategoryTreeOption = (category: CategoryNode): TreeSelectOption => ({
  key: category.id,
  label: category.name,
  disabled: category.id === 'category_all',
  children: category.children.map(toCategoryTreeOption)
})

const categoryTreeOptions = computed(() => props.categories.map(toCategoryTreeOption))

const handleCategoryChange = (value: string | number | Array<string | number> | null) => {
  const nextValue = Array.isArray(value) ? value[0] : value
  // 清空分类时回到系统“未分类”，避免把虚拟根写入书签。
  updateTextField('categoryId', nextValue ? String(nextValue) : 'category_uncategorized')
}

const handleDrawerUpdate = (value: boolean) => {
  if (!value) emit('close')
}

const encodeFileAsDataURL = (file: File) =>
  new Promise<string>((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result || ''))
    reader.onerror = () => reject(reader.error)
    reader.readAsDataURL(file)
  })

const handleThumbnailUpload = async ({ file, onFinish, onError }: UploadCustomRequestOptions) => {
  if (!file.file) {
    onError()
    return
  }
  thumbnailUploading.value = true
  try {
    const data = await encodeFileAsDataURL(file.file)
    emit('save-thumbnail', {
      fileName: file.name,
      mime: file.type || file.file.type || '',
      data
    })
    thumbnailFileList.value = [
      {
        id: file.id,
        name: file.name,
        status: 'finished',
        url: data,
        thumbnailUrl: data
      }
    ]
    onFinish()
  } catch {
    onError()
  } finally {
    thumbnailUploading.value = false
  }
}

const beforeThumbnailUpload = ({ file }: { file: UploadSettledFileInfo }) => {
  return Boolean(file.type?.startsWith('image/') || thumbnailImageNamePattern.test(file.name))
}

const removeCustomThumbnail = () => {
  thumbnailFileList.value = []
  emit('clear-custom-thumbnail')
  return true
}

const saveThumbnailUrl = () => {
  const url = thumbnailUrl.value.trim()
  if (!url) return
  emit('save-thumbnail-url', url)
  thumbnailUrl.value = ''
}

const syncThumbnailFileList = () => {
  const path = customThumbnailPath.value
  if (!props.selected || !path) {
    thumbnailFileList.value = []
    return
  }
  // 照片墙展示当前已缓存的自定义缩略图，便于用户删除或替换。
  thumbnailFileList.value = [
    {
      id: `custom-${props.selected.id}`,
      name: '自定义缩略图',
      status: 'finished',
      url: path,
      thumbnailUrl: path
    }
  ]
}

watch([() => props.selected?.id, customThumbnailPath], syncThumbnailFileList, { immediate: true })
</script>

<template>
  <NDrawer
    :show="props.open"
    :width="520"
    placement="right"
    :trap-focus="false"
    class="bookmark-editor-drawer"
    @update:show="handleDrawerUpdate"
  >
    <NDrawerContent closable>
      <template #header>
        <div>
          <span class="eyebrow">{{ props.selected ? 'EDIT' : 'CREATE' }}</span>
          <h2 class="drawer-title">{{ props.selected ? props.draft.title || '编辑书签' : '新增书签' }}</h2>
        </div>
      </template>

      <NForm class="editor-form" label-placement="top" @submit.prevent="emit('save')">
        <NFormItem label="标题" required>
          <NInput
            :value="props.draft.title"
            placeholder="书签标题"
            @update:value="updateTextField('title', $event)"
          />
        </NFormItem>

        <NFormItem label="网址" required>
          <NInput :value="props.draft.url" placeholder="https://example.com" @update:value="updateTextField('url', $event)">
            <template #prefix>
              <NIcon :component="appIcons.link" />
            </template>
          </NInput>
        </NFormItem>

        <NFormItem label="描述">
          <NInput
            :value="props.draft.description"
            type="textarea"
            :autosize="{ minRows: 4, maxRows: 8 }"
            placeholder="用于搜索和 AI 复审的简短描述"
            @update:value="updateTextField('description', $event)"
          />
        </NFormItem>

        <NFormItem label="分类">
          <NTreeSelect
            clearable
            default-expand-all
            filterable
            show-line
            show-path
            class="editor-category-tree-select"
            :value="props.draft.categoryId || 'category_uncategorized'"
            :options="categoryTreeOptions"
            placeholder="选择书签分类"
            separator=" / "
            @update:value="handleCategoryChange"
          />
        </NFormItem>

        <NFormItem label="标签">
          <NDynamicTags :value="props.draft.tags" @update:value="updateListField('tags', $event)" />
        </NFormItem>

        <NFormItem label="关键词">
          <NDynamicTags :value="props.draft.keywords" @update:value="updateListField('keywords', $event)" />
        </NFormItem>

        <NFormItem label="别名">
          <NDynamicTags :value="props.draft.aliases" @update:value="updateListField('aliases', $event)" />
        </NFormItem>

        <NFormItem label="缩略图">
          <NSpace vertical :size="10" class="editor-thumbnail-tools">
            <div class="thumbnail-preview-panel">
              <NImage
                v-if="thumbnailDisplayPath"
                class="thumbnail-preview-image"
                :src="thumbnailDisplayPath"
                :preview-src="thumbnailDisplayPath"
                object-fit="cover"
              />
              <div v-else class="thumbnail-preview-empty">
                <NIcon :component="appIcons.library" />
                <span>暂无缩略图</span>
              </div>
              <div class="thumbnail-preview-copy">
                <strong>{{ thumbnailSourceLabel }}</strong>
                <span v-if="selectedThumbnail?.autoStatus === 'error'">{{ selectedThumbnail.autoError }}</span>
                <span v-else>{{ selectedThumbnail?.useAuto === false ? '正在使用自定义缩略图' : '正在使用自动缩略图' }}</span>
              </div>
            </div>

            <NSpace align="center" justify="space-between" class="thumbnail-mode-row">
              <NCheckbox
                :checked="selectedThumbnail?.useAuto !== false"
                :disabled="!props.selected"
                @update:checked="emit('set-thumbnail-auto', Boolean($event))"
              >
                使用自动缩略图
              </NCheckbox>
              <NButton size="small" secondary :disabled="!props.selected" @click="emit('refresh-thumbnail')">刷新自动缩略图</NButton>
            </NSpace>

            <div class="thumbnail-source-grid">
              <div class="thumbnail-source-card">
                <span>自动缓存</span>
                <NTag size="small" round>{{ autoThumbnailPath ? selectedThumbnail?.autoSource || '已缓存' : '待获取' }}</NTag>
              </div>
              <div class="thumbnail-source-card">
                <span>自定义缓存</span>
                <NTag size="small" round>{{ customThumbnailPath ? selectedThumbnail?.customSource || '已缓存' : '未设置' }}</NTag>
              </div>
            </div>

            <NUpload
              v-model:file-list="thumbnailFileList"
              accept="image/*"
              :custom-request="handleThumbnailUpload"
              :default-upload="true"
              :disabled="!props.selected || thumbnailUploading"
              :max="1"
              :on-remove="removeCustomThumbnail"
              :on-before-upload="beforeThumbnailUpload"
              list-type="image-card"
            />

            <NSpace :size="8" align="center" class="thumbnail-url-tools">
              <NInput v-model:value="thumbnailUrl" placeholder="粘贴图片地址，例如 https://example.com/cover.jpg" />
              <NButton secondary :disabled="!props.selected || !thumbnailUrl.trim()" @click="saveThumbnailUrl">缓存地址图片</NButton>
            </NSpace>
            <NAlert type="info" :show-icon="false">自动缩略图按 OpenGraph、Favicon、本地截图顺序获取；自定义缩略图会缓存到本地。</NAlert>
          </NSpace>
        </NFormItem>

        <div v-if="props.selected" class="editor-meta">
          <NTag size="small" round>{{ props.selected.domain || '本地书签' }}</NTag>
          <span>更新于 {{ props.selected.updatedAt || '-' }}</span>
        </div>
      </NForm>

      <template #footer>
        <NSpace justify="space-between" align="center" class="drawer-footer">
          <span class="drawer-status">{{ props.status || '等待编辑' }}</span>
          <NSpace :size="8">
            <NButton :disabled="!props.selected" secondary @click="emit('analyze')">
              <template #icon>
                <NIcon :component="appIcons.ai" />
              </template>
              AI
            </NButton>
            <NPopconfirm :disabled="!props.selected" positive-text="删除" negative-text="取消" @positive-click="emit('remove')">
              <template #trigger>
                <NButton :disabled="!props.selected" type="error" secondary aria-label="删除当前书签">
                  <template #icon>
                    <NIcon :component="appIcons.trash" />
                  </template>
                </NButton>
              </template>
              删除当前书签？
            </NPopconfirm>
            <NButton type="primary" @click="emit('save')">
              <template #icon>
                <NIcon :component="appIcons.save" />
              </template>
              保存
            </NButton>
          </NSpace>
        </NSpace>
      </template>
    </NDrawerContent>
  </NDrawer>
</template>
