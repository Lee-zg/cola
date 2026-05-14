<!-- 文件说明：frontend/src/components/NavRail.vue，对应当前模块的界面或交互逻辑。 -->
<script setup lang="ts">
import { computed } from 'vue'
import { NButton, NIcon, NTooltip } from 'naive-ui'
import { appIcons } from '../icons'
import { routes, type RoutePath } from '../navigation'

const props = defineProps<{
  activePath: RoutePath
}>()

const emit = defineEmits<{
  navigate: [path: RoutePath]
}>()

const groupedRoutes = computed(() => ({
  primary: routes.filter((route) => route.group === 'primary'),
  workflow: routes.filter((route) => route.group === 'workflow'),
  system: routes.filter((route) => route.group === 'system')
}))

const navigate = (path: RoutePath) => {
  emit('navigate', path)
}
</script>

<template>
  <aside class="nav-rail" aria-label="主导航">
    <div class="nav-brand" title="Cola Bookmarks">
      <NIcon :component="appIcons.bookmarks" />
    </div>

    <nav class="nav-stack">
      <div class="nav-group">
        <NTooltip v-for="route in groupedRoutes.primary" :key="route.path" placement="right" trigger="hover">
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

      <div class="nav-divider"></div>

      <div class="nav-group">
        <NTooltip v-for="route in groupedRoutes.workflow" :key="route.path" placement="right" trigger="hover">
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

      <div class="nav-divider"></div>

      <div class="nav-group nav-bottom">
        <NTooltip v-for="route in groupedRoutes.system" :key="route.path" placement="right" trigger="hover">
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
    </nav>
  </aside>
</template>
