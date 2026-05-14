<script setup lang="ts">
import { computed } from 'vue'
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
    <div class="nav-brand" title="Cola Bookmarks">CB</div>

    <nav class="nav-stack">
      <div class="nav-group">
        <button
          v-for="route in groupedRoutes.primary"
          :key="route.path"
          class="nav-item"
          :class="{ active: route.path === props.activePath }"
          :title="route.label"
          type="button"
          @click="navigate(route.path)"
        >
          <span class="nav-icon">{{ route.icon }}</span>
          <span class="nav-text">{{ route.label }}</span>
        </button>
      </div>

      <div class="nav-divider"></div>

      <div class="nav-group">
        <button
          v-for="route in groupedRoutes.workflow"
          :key="route.path"
          class="nav-item"
          :class="{ active: route.path === props.activePath }"
          :title="route.label"
          type="button"
          @click="navigate(route.path)"
        >
          <span class="nav-icon">{{ route.icon }}</span>
          <span class="nav-text">{{ route.label }}</span>
        </button>
      </div>

      <div class="nav-divider"></div>

      <div class="nav-group nav-bottom">
        <button
          v-for="route in groupedRoutes.system"
          :key="route.path"
          class="nav-item"
          :class="{ active: route.path === props.activePath }"
          :title="route.label"
          type="button"
          @click="navigate(route.path)"
        >
          <span class="nav-icon">{{ route.icon }}</span>
          <span class="nav-text">{{ route.label }}</span>
        </button>
      </div>
    </nav>
  </aside>
</template>
