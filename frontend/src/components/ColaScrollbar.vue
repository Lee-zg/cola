<!-- 文件说明：frontend/src/components/ColaScrollbar.vue，提供应用内可复用的可乐主题滚动容器。 -->
<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'

const props = withDefaults(
  defineProps<{
    ariaLabel?: string
    compact?: boolean
    shadow?: boolean
  }>(),
  {
    ariaLabel: '内容滚动区',
    compact: false,
    shadow: true
  }
)

const emit = defineEmits<{
  'reach-bottom': []
  scroll: [payload: { clientHeight: number; progress: number; scrollHeight: number; scrollTop: number }]
}>()

const viewportRef = ref<HTMLElement | null>(null)
const trackRef = ref<HTMLElement | null>(null)
const thumbHeight = ref(40)
const thumbOffset = ref(0)
const scrollProgress = ref(0)
const isScrollable = ref(false)
const isScrolling = ref(false)
const isDragging = ref(false)
const hasTopShadow = ref(false)
const hasBottomShadow = ref(false)

let resizeObserver: ResizeObserver | null = null
let mutationObserver: MutationObserver | null = null
let observedContent: Element | null = null
let animationFrameId = 0
let scrollStateTimer: ReturnType<typeof window.setTimeout> | null = null
let dragStartY = 0
let dragStartScrollTop = 0

const rootClasses = computed(() => ({
  'is-ready': isScrollable.value,
  'is-scrollable': isScrollable.value,
  'is-scrolling': isScrolling.value,
  'is-dragging': isDragging.value,
  'is-compact': props.compact,
  'has-top-shadow': props.shadow && hasTopShadow.value,
  'has-bottom-shadow': props.shadow && hasBottomShadow.value
}))

const rootStyle = computed(() => ({
  '--cola-scroll-progress': `${scrollProgress.value}`
}))

const thumbStyle = computed(() => ({
  height: `${thumbHeight.value}px`,
  transform: `translate3d(0, ${thumbOffset.value}px, 0)`
}))

const markScrolling = () => {
  isScrolling.value = true
  if (scrollStateTimer) window.clearTimeout(scrollStateTimer)
  scrollStateTimer = window.setTimeout(() => {
    isScrolling.value = false
  }, 520)
}

const updateMetricsNow = () => {
  const viewport = viewportRef.value
  const track = trackRef.value
  if (!viewport || !track) return

  const scrollRange = Math.max(0, viewport.scrollHeight - viewport.clientHeight)
  isScrollable.value = scrollRange > 1
  hasTopShadow.value = viewport.scrollTop > 2
  hasBottomShadow.value = viewport.scrollTop < scrollRange - 2

  if (!isScrollable.value) {
    thumbOffset.value = 0
    scrollProgress.value = 0
    return
  }

  const trackHeight = track.clientHeight
  const ratio = viewport.clientHeight / viewport.scrollHeight
  const minThumbHeight = props.compact ? 28 : 40
  thumbHeight.value = Math.max(minThumbHeight, Math.round(trackHeight * ratio))

  const maxThumbOffset = Math.max(0, trackHeight - thumbHeight.value)
  scrollProgress.value = scrollRange ? viewport.scrollTop / scrollRange : 0
  thumbOffset.value = Math.round(maxThumbOffset * scrollProgress.value)
}

const updateMetrics = () => {
  if (animationFrameId) window.cancelAnimationFrame(animationFrameId)
  animationFrameId = window.requestAnimationFrame(updateMetricsNow)
}

const observeContent = () => {
  if (!resizeObserver || !viewportRef.value) return
  if (observedContent) resizeObserver.unobserve(observedContent)

  // 监听插槽内容尺寸变化，路由切换或列表更新后自动刷新滑块位置。
  observedContent = viewportRef.value.firstElementChild
  if (observedContent) resizeObserver.observe(observedContent)
}

const handleScroll = () => {
  const viewport = viewportRef.value
  markScrolling()
  updateMetrics()
  if (!viewport) return

  const scrollRange = Math.max(0, viewport.scrollHeight - viewport.clientHeight)
  const progress = scrollRange ? viewport.scrollTop / scrollRange : 0
  emit('scroll', {
    clientHeight: viewport.clientHeight,
    progress,
    scrollHeight: viewport.scrollHeight,
    scrollTop: viewport.scrollTop
  })
  if (scrollRange > 0 && viewport.scrollTop + viewport.clientHeight >= viewport.scrollHeight - 96) {
    emit('reach-bottom')
  }
}

const handlePointerMove = (event: PointerEvent) => {
  const viewport = viewportRef.value
  const track = trackRef.value
  if (!viewport || !track) return

  const scrollRange = Math.max(0, viewport.scrollHeight - viewport.clientHeight)
  const maxThumbOffset = Math.max(1, track.clientHeight - thumbHeight.value)
  const dragDelta = event.clientY - dragStartY
  viewport.scrollTop = dragStartScrollTop + (dragDelta / maxThumbOffset) * scrollRange
}

const stopDrag = () => {
  isDragging.value = false
  window.removeEventListener('pointermove', handlePointerMove)
  window.removeEventListener('pointerup', stopDrag)
  markScrolling()
}

const startDrag = (event: PointerEvent) => {
  const viewport = viewportRef.value
  if (!viewport || !isScrollable.value) return

  event.preventDefault()
  isDragging.value = true
  dragStartY = event.clientY
  dragStartScrollTop = viewport.scrollTop
  markScrolling()
  window.addEventListener('pointermove', handlePointerMove)
  window.addEventListener('pointerup', stopDrag)
}

const handleTrackPointerDown = (event: PointerEvent) => {
  const viewport = viewportRef.value
  const track = trackRef.value
  if (!viewport || !track || event.target !== event.currentTarget) return

  const scrollRange = Math.max(0, viewport.scrollHeight - viewport.clientHeight)
  const maxThumbOffset = Math.max(1, track.clientHeight - thumbHeight.value)
  const trackRect = track.getBoundingClientRect()
  const nextOffset = Math.min(Math.max(event.clientY - trackRect.top - thumbHeight.value / 2, 0), maxThumbOffset)

  markScrolling()
  viewport.scrollTo({
    top: (nextOffset / maxThumbOffset) * scrollRange,
    behavior: 'smooth'
  })
}

onMounted(() => {
  resizeObserver = new ResizeObserver(() => {
    updateMetrics()
  })

  if (viewportRef.value) {
    resizeObserver.observe(viewportRef.value)
    mutationObserver = new MutationObserver(() => {
      observeContent()
      updateMetrics()
    })
    mutationObserver.observe(viewportRef.value, { childList: true, subtree: true })
  }

  observeContent()
  window.addEventListener('resize', updateMetrics)
  void nextTick(updateMetrics)
})

onBeforeUnmount(() => {
  if (animationFrameId) window.cancelAnimationFrame(animationFrameId)
  if (scrollStateTimer) window.clearTimeout(scrollStateTimer)
  resizeObserver?.disconnect()
  mutationObserver?.disconnect()
  window.removeEventListener('resize', updateMetrics)
  window.removeEventListener('pointermove', handlePointerMove)
  window.removeEventListener('pointerup', stopDrag)
})
</script>

<template>
  <div class="cola-scrollbar" :class="rootClasses" :style="rootStyle">
    <div ref="viewportRef" class="cola-scrollbar__viewport" role="region" :aria-label="props.ariaLabel" @scroll="handleScroll">
      <slot />
    </div>

    <div ref="trackRef" class="cola-scrollbar__track" aria-hidden="true" @pointerdown="handleTrackPointerDown">
      <div class="cola-scrollbar__thumb" :style="thumbStyle" @pointerdown.stop="startDrag">
        <span class="cola-scrollbar__label"></span>
        <span class="cola-scrollbar__bubble bubble-one"></span>
        <span class="cola-scrollbar__bubble bubble-two"></span>
        <span class="cola-scrollbar__bubble bubble-three"></span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.cola-scrollbar {
  position: relative;
  min-width: 0;
  min-height: 0;
  height: 100%;
  overflow: hidden;
}

.cola-scrollbar__viewport {
  width: 100%;
  height: 100%;
  min-width: 0;
  min-height: 0;
  overflow: auto;
  overscroll-behavior: contain;
  scroll-behavior: smooth;
  scrollbar-width: none;
}

.cola-scrollbar__viewport::-webkit-scrollbar {
  width: 0;
  height: 0;
}

.cola-scrollbar__track {
  position: absolute;
  top: 12px;
  right: 5px;
  bottom: 12px;
  z-index: 4;
  width: 14px;
  border-radius: 999px;
  background:
    linear-gradient(180deg, rgb(255 255 255 / 18%), transparent 30%, rgb(255 255 255 / 14%)),
    color-mix(in srgb, var(--panel) 54%, transparent);
  box-shadow:
    inset 0 0 0 1px color-mix(in srgb, var(--accent) 18%, transparent),
    0 12px 26px color-mix(in srgb, var(--accent) 12%, transparent);
  opacity: 0;
  pointer-events: none;
  transition:
    opacity 180ms ease,
    box-shadow 180ms ease;
}

.cola-scrollbar:hover .cola-scrollbar__track,
.cola-scrollbar.is-scrolling .cola-scrollbar__track,
.cola-scrollbar.is-dragging .cola-scrollbar__track {
  opacity: 1;
  pointer-events: auto;
}

.cola-scrollbar:not(.is-ready) .cola-scrollbar__track {
  opacity: 0;
  pointer-events: none;
}

.cola-scrollbar.is-scrolling .cola-scrollbar__track,
.cola-scrollbar.is-dragging .cola-scrollbar__track {
  box-shadow:
    inset 0 0 0 1px color-mix(in srgb, var(--accent) 34%, transparent),
    0 16px 34px color-mix(in srgb, var(--accent) 22%, transparent);
}

.cola-scrollbar__track::before {
  content: '';
  position: absolute;
  inset: 4px 5px;
  border-radius: inherit;
  background:
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--accent) 80%, #f8f5f0 20%) calc(var(--cola-scroll-progress) * 100%),
      transparent 0
    );
  opacity: 0.32;
}

.cola-scrollbar__thumb {
  position: absolute;
  left: 2px;
  width: 10px;
  min-height: 28px;
  cursor: grab;
  border-radius: 999px;
  background:
    radial-gradient(circle at 50% 18%, rgb(255 255 255 / 86%) 0 2px, transparent 3px),
    linear-gradient(90deg, rgb(255 255 255 / 34%) 0 2px, transparent 2px 8px, rgb(71 30 18 / 24%) 8px 100%),
    linear-gradient(180deg, #f4c06f 0%, #b65035 30%, #6e281c 72%, #381711 100%);
  box-shadow:
    inset 0 0 0 1px rgb(255 255 255 / 22%),
    inset 0 -10px 18px rgb(48 19 13 / 34%),
    0 0 0 2px color-mix(in srgb, var(--accent-soft) 62%, transparent),
    0 10px 22px color-mix(in srgb, var(--accent) 30%, transparent);
  overflow: hidden;
  transition:
    filter 180ms ease,
    transform 80ms linear,
    width 180ms ease;
}

.cola-scrollbar__thumb::before {
  content: '';
  position: absolute;
  inset: 8px 3px;
  border-radius: 999px;
  background:
    linear-gradient(180deg, rgb(255 255 255 / 50%), transparent 22%),
    repeating-linear-gradient(180deg, transparent 0 8px, rgb(255 255 255 / 20%) 8px 10px);
  mix-blend-mode: screen;
  opacity: 0.74;
}

.cola-scrollbar__label {
  position: absolute;
  left: 2px;
  right: 2px;
  top: 45%;
  height: 12px;
  border-radius: 4px;
  background:
    linear-gradient(180deg, rgb(255 255 255 / 94%), rgb(255 247 235 / 84%)),
    linear-gradient(90deg, #fff, #f2d5bd);
  box-shadow: 0 0 12px rgb(255 255 255 / 28%);
  transform: translateY(-50%);
}

.cola-scrollbar__bubble {
  position: absolute;
  display: block;
  width: 3px;
  height: 3px;
  border-radius: 999px;
  background: rgb(255 255 255 / 82%);
  box-shadow: 0 0 8px rgb(255 255 255 / 56%);
  opacity: 0.78;
}

.bubble-one {
  top: 18%;
  left: 3px;
}

.bubble-two {
  top: 62%;
  right: 2px;
}

.bubble-three {
  top: 78%;
  left: 4px;
}

.cola-scrollbar.is-scrolling .cola-scrollbar__thumb,
.cola-scrollbar.is-dragging .cola-scrollbar__thumb {
  filter: saturate(1.18) brightness(1.08);
  width: 12px;
}

.cola-scrollbar.is-dragging .cola-scrollbar__thumb {
  cursor: grabbing;
}

.cola-scrollbar.is-scrolling .cola-scrollbar__bubble {
  animation: cola-bubble-rise 760ms ease-in-out infinite alternate;
}

.cola-scrollbar.has-top-shadow::before,
.cola-scrollbar.has-bottom-shadow::after {
  content: '';
  position: absolute;
  left: 0;
  right: 18px;
  z-index: 3;
  height: 22px;
  pointer-events: none;
}

.cola-scrollbar.has-top-shadow::before {
  top: 0;
  background: linear-gradient(180deg, color-mix(in srgb, var(--bg) 62%, transparent), transparent);
}

.cola-scrollbar.has-bottom-shadow::after {
  bottom: 0;
  background: linear-gradient(0deg, color-mix(in srgb, var(--bg) 62%, transparent), transparent);
}

.cola-scrollbar.is-compact .cola-scrollbar__track {
  top: 4px;
  right: 1px;
  bottom: 4px;
  width: 10px;
}

.cola-scrollbar.is-compact .cola-scrollbar__thumb {
  left: 1px;
  width: 8px;
}

@keyframes cola-bubble-rise {
  from {
    transform: translateY(3px) scale(0.9);
  }

  to {
    transform: translateY(-5px) scale(1.08);
  }
}

@media (prefers-reduced-motion: reduce) {
  .cola-scrollbar__viewport {
    scroll-behavior: auto;
  }

  .cola-scrollbar__track,
  .cola-scrollbar__thumb,
  .cola-scrollbar__bubble {
    animation: none !important;
    transition-duration: 1ms !important;
  }
}
</style>
