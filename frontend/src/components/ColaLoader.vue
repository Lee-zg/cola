<!-- 文件说明：frontend/src/components/ColaLoader.vue，可复用的可乐主题加载动效组件。 -->
<script setup lang="ts">
withDefaults(
  defineProps<{
    label?: string
  }>(),
  {
    label: '正在加载书签'
  }
)
</script>

<template>
  <div class="cola-loader" role="status" aria-live="polite">
    <div class="cola-loader__bottle" aria-hidden="true">
      <span class="cola-loader__cap"></span>
      <span class="cola-loader__label-band"></span>
      <!-- 瓶身内部气泡用于表达可乐加载中的流动感。 -->
      <span class="cola-loader__bubble bubble-one"></span>
      <span class="cola-loader__bubble bubble-two"></span>
      <span class="cola-loader__bubble bubble-three"></span>
      <span class="cola-loader__spark spark-one"></span>
      <span class="cola-loader__spark spark-two"></span>
      <span class="cola-loader__spark spark-three"></span>
    </div>
    <span class="cola-loader__label">{{ label }}</span>
  </div>
</template>

<style scoped>
.cola-loader {
  min-width: 180px;
  display: grid;
  place-items: center;
  gap: 12px;
  color: var(--text);
  text-align: center;
}

/* 瓶身整体负责经典玻璃瓶轮廓、深色可乐液体和轻微浮动动效。 */
.cola-loader__bottle {
  position: relative;
  width: 58px;
  height: 110px;
  overflow: hidden;
  border: 2px solid color-mix(in srgb, var(--accent) 28%, #fff 72%);
  border-radius: 14px 14px 21px 21px / 28px 28px 18px 18px;
  background:
    linear-gradient(90deg, transparent 0 12px, rgb(255 255 255 / 42%) 13px 18px, transparent 20px 100%),
    radial-gradient(ellipse at 50% 7%, rgb(255 236 186 / 72%) 0 10px, transparent 11px),
    linear-gradient(180deg, #42170f 0%, #692316 38%, #31100b 100%);
  box-shadow:
    inset 0 0 0 1px rgb(255 255 255 / 22%),
    inset -12px 0 22px rgb(22 8 4 / 32%),
    0 18px 42px color-mix(in srgb, var(--accent) 24%, transparent);
  animation: cola-bottle-float 1500ms ease-in-out infinite alternate;
}

.cola-loader__bottle::before {
  content: '';
  position: absolute;
  left: 13px;
  right: 13px;
  top: -2px;
  height: 36px;
  border: inherit;
  border-bottom: 0;
  border-radius: 9px 9px 11px 11px;
  background:
    linear-gradient(90deg, transparent 0 7px, rgb(255 255 255 / 38%) 8px 11px, transparent 12px 100%),
    linear-gradient(180deg, #3d140d 0%, #5d2015 100%);
  box-shadow: inset 0 0 0 1px rgb(255 255 255 / 16%);
}

.cola-loader__bottle::after {
  content: '';
  position: absolute;
  inset: 38px 12px 12px 16px;
  border-radius: 999px 999px 18px 18px;
  background:
    linear-gradient(90deg, rgb(255 255 255 / 36%) 0 5px, transparent 6px 100%),
    repeating-linear-gradient(180deg, transparent 0 12px, rgb(255 226 171 / 14%) 12px 15px);
  mix-blend-mode: screen;
  animation: cola-bottle-light-flow 1100ms linear infinite;
}

.cola-loader__cap {
  position: absolute;
  left: 20px;
  top: 0;
  z-index: 2;
  width: 18px;
  height: 8px;
  border-radius: 5px 5px 3px 3px;
  background:
    repeating-linear-gradient(90deg, rgb(255 255 255 / 18%) 0 2px, transparent 2px 4px),
    #8f2418;
  box-shadow: 0 2px 8px rgb(44 12 7 / 28%);
}

.cola-loader__label-band {
  position: absolute;
  left: 5px;
  right: 5px;
  top: 48px;
  z-index: 2;
  height: 24px;
  border-radius: 9px;
  background:
    linear-gradient(135deg, rgb(255 255 255 / 72%) 0 8px, transparent 9px 100%),
    linear-gradient(180deg, #d53825 0%, #9e2018 100%);
  box-shadow:
    inset 0 0 0 1px rgb(255 255 255 / 28%),
    0 5px 10px rgb(42 12 8 / 26%);
}

.cola-loader__label-band::after {
  content: '';
  position: absolute;
  left: 50%;
  top: 50%;
  width: 21px;
  height: 7px;
  border-radius: 999px;
  background: rgb(255 246 228 / 92%);
  transform: translate(-50%, -50%) rotate(-8deg);
}

.cola-loader__bubble,
.cola-loader__spark {
  position: absolute;
  display: block;
  border-radius: 999px;
  background: rgb(255 255 255 / 86%);
  box-shadow: 0 0 12px rgb(255 255 255 / 46%);
}

.cola-loader__bubble {
  width: 8px;
  height: 8px;
  animation: cola-bubble-rise 920ms ease-in-out infinite;
}

.bubble-one {
  left: 16px;
  bottom: 18px;
}

.bubble-two {
  right: 15px;
  bottom: 34px;
  animation-delay: 160ms;
}

.bubble-three {
  left: 30px;
  bottom: 11px;
  animation-delay: 320ms;
}

.cola-loader__spark {
  width: 4px;
  height: 4px;
  animation: cola-spark-rise 1100ms ease-out infinite;
}

.spark-one {
  left: 14px;
  bottom: 26px;
}

.spark-two {
  right: 18px;
  bottom: 22px;
  animation-delay: 220ms;
}

.spark-three {
  left: 38px;
  bottom: 44px;
  animation-delay: 420ms;
}

.cola-loader__label {
  color: var(--muted);
  font-size: 13px;
  font-weight: 700;
}

@keyframes cola-bottle-float {
  from {
    transform: translateY(2px) rotate(-2deg);
  }

  to {
    transform: translateY(-4px) rotate(2deg);
  }
}

@keyframes cola-bottle-light-flow {
  from {
    transform: translateY(8px);
  }

  to {
    transform: translateY(-8px);
  }
}

@keyframes cola-bubble-rise {
  0% {
    transform: translateY(8px) scale(0.6);
    opacity: 0.2;
  }

  55% {
    opacity: 0.95;
  }

  100% {
    transform: translateY(-18px) scale(1.15);
    opacity: 0;
  }
}

@keyframes cola-spark-rise {
  from {
    transform: translateY(12px);
    opacity: 0.2;
  }

  to {
    transform: translateY(-30px);
    opacity: 0;
  }
}

@media (prefers-reduced-motion: reduce) {
  .cola-loader__bottle,
  .cola-loader__bottle::after,
  .cola-loader__bubble,
  .cola-loader__spark {
    animation: none !important;
  }
}
</style>
