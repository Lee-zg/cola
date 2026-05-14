<!-- 文件说明：frontend/src/components/NavRail.vue，对应当前模块的界面或交互逻辑。 -->
<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { NButton, NIcon, NTooltip } from 'naive-ui'
import { appIcons } from '../icons'
import { routes, type NavItem, type RoutePath } from '../navigation'

const props = defineProps<{
  activePath: RoutePath
}>()

const emit = defineEmits<{
  navigate: [path: RoutePath]
}>()

const navRailRef = ref<HTMLElement | null>(null)
const navStackRef = ref<HTMLElement | null>(null)
const overflowOpen = ref(false)
const visibleCount = ref(routes.length)

let resizeObserver: ResizeObserver | null = null

const itemHeight = 52
const itemGap = 7
const dividerHeight = 7
const groupGap = 10
const overflowButtonSpace = 62

const visibleRoutes = computed(() => routes.slice(0, visibleCount.value))
const overflowRoutes = computed(() => routes.slice(visibleCount.value))
const hasOverflow = computed(() => overflowRoutes.value.length > 0)

const groupedVisibleRoutes = computed(() => ({
  primary: visibleRoutes.value.filter((route) => route.group === 'primary'),
  workflow: visibleRoutes.value.filter((route) => route.group === 'workflow'),
  system: visibleRoutes.value.filter((route) => route.group === 'system')
}))

const renderedGroups = computed(() =>
  [
    { key: 'primary', routes: groupedVisibleRoutes.value.primary },
    { key: 'workflow', routes: groupedVisibleRoutes.value.workflow },
    { key: 'system', routes: groupedVisibleRoutes.value.system }
  ].filter((group) => group.routes.length)
)

const navigate = (path: RoutePath) => {
  emit('navigate', path)
  overflowOpen.value = false
}

const getRoutesHeight = (count: number) => {
  const selectedRoutes = routes.slice(0, count)
  const groups = new Set<NavItem['group']>(selectedRoutes.map((route) => route.group))
  const groupCount = groups.size
  const dividerCount = Math.max(0, groupCount - 1)
  const gapCount = Math.max(0, count - groupCount)
  const renderedBlockCount = groupCount + dividerCount
  const blockGapCount = Math.max(0, renderedBlockCount - 1)

  return count * itemHeight + gapCount * itemGap + dividerCount * dividerHeight + blockGapCount * groupGap
}

const calculateVisibleCount = () => {
  const stack = navStackRef.value
  if (!stack) return

  const availableHeight = stack.clientHeight
  let nextCount = 1

  // 为底部溢出按钮预留空间，窗口过矮时优先显示前面的高频导航项。
  for (let count = routes.length; count >= 1; count -= 1) {
    const reservedHeight = count < routes.length ? overflowButtonSpace : 0
    if (getRoutesHeight(count) + reservedHeight <= availableHeight) {
      nextCount = count
      break
    }
  }

  visibleCount.value = Math.max(1, nextCount)
}

const closeOverflow = () => {
  overflowOpen.value = false
}

const openOverflow = () => {
  if (hasOverflow.value) overflowOpen.value = true
}

const toggleOverflow = () => {
  overflowOpen.value = !overflowOpen.value
}

const handleDocumentPointerDown = (event: PointerEvent) => {
  if (!navRailRef.value?.contains(event.target as Node)) {
    closeOverflow()
  }
}

onMounted(() => {
  resizeObserver = new ResizeObserver(calculateVisibleCount)
  if (navStackRef.value) resizeObserver.observe(navStackRef.value)
  document.addEventListener('pointerdown', handleDocumentPointerDown)
  void nextTick(calculateVisibleCount)
})

onBeforeUnmount(() => {
  resizeObserver?.disconnect()
  document.removeEventListener('pointerdown', handleDocumentPointerDown)
})

watch(hasOverflow, (nextHasOverflow) => {
  if (!nextHasOverflow) closeOverflow()
})
</script>

<template>
  <aside ref="navRailRef" class="nav-rail" aria-label="主导航">
    <div class="nav-brand" title="Cola Bookmarks">
      <NIcon :component="appIcons.bookmarks" />
    </div>

    <nav ref="navStackRef" class="nav-stack">
      <div class="nav-visible-groups">
        <template v-for="(group, groupIndex) in renderedGroups" :key="group.key">
          <div v-if="groupIndex > 0" class="nav-divider"></div>
          <div class="nav-group">
            <NTooltip v-for="route in group.routes" :key="route.path" placement="right" trigger="hover">
              <template #trigger>
                <NButton
                  class="nav-item"
                  :class="{ active: route.path === props.activePath }"
                  quaternary
                  :aria-label="route.label"
                  @click="navigate(route.path)"
                >
                  <NIcon class="nav-icon" :component="appIcons[route.icon]" />
                  <span class="nav-text">{{ route.label }}</span>
                </NButton>
              </template>
              {{ route.label }}
            </NTooltip>
          </div>
        </template>
      </div>

      <div v-if="hasOverflow" class="nav-overflow" @mouseleave="closeOverflow">
        <NButton
          class="nav-item nav-overflow-trigger"
          :class="{ active: overflowOpen || overflowRoutes.some((route) => route.path === props.activePath) }"
          quaternary
          aria-label="展开更多导航"
          @click="toggleOverflow"
          @focus="openOverflow"
          @mouseenter="openOverflow"
        >
          <NIcon class="nav-icon nav-overflow-icon" :component="appIcons.overflow" />
          <span class="nav-text">更多</span>
        </NButton>

        <Transition name="nav-overflow-pop">
          <div v-if="overflowOpen" class="nav-overflow-panel" role="menu" @mouseenter="openOverflow">
            <button
              v-for="route in overflowRoutes"
              :key="route.path"
              class="nav-overflow-item"
              :class="{ active: route.path === props.activePath }"
              type="button"
              role="menuitem"
              @click="navigate(route.path)"
            >
              <NIcon :component="appIcons[route.icon]" />
              <span>{{ route.label }}</span>
            </button>
          </div>
        </Transition>
      </div>
    </nav>
  </aside>
</template>
