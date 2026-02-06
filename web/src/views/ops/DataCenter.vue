<template>
  <div class="data-center">
    <div class="page-header">
      <h1>数据中心</h1>
    </div>

    <div class="empty-state card">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
        <line x1="18" y1="20" x2="18" y2="10"/>
        <line x1="12" y1="20" x2="12" y2="4"/>
        <line x1="6" y1="20" x2="6" y2="14"/>
      </svg>
      <h2>功能开发中</h2>
      <p>数据中心功能正在开发中，敬请期待...</p>
    </div>
  </div>
</template>

<script setup>
</script>

<style scoped>
.data-center {
  padding: 24px;
}

.page-header {
  margin-bottom: 24px;
}

.page-header h1 {
  font-size: 24px;
  font-weight: 600;
  color: var(--gray-700);
}

.empty-state {
  padding: 80px 40px;
  text-align: center;
}

.empty-state svg {
  width: 80px;
  height: 80px;
  color: var(--gray-300);
  margin-bottom: 24px;
}

.empty-state h2 {
  font-size: 20px;
  font-weight: 600;
  color: var(--gray-600);
  margin-bottom: 12px;
}

.empty-state p {
  font-size: 14px;
  color: var(--gray-400);
}
</style>
