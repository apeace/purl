<template>
  <div>
    <h1>Customer App</h1>
    <p>{{ message }}</p>
  </div>
</template>

<script setup>
import { hello } from '@pipeline/lib';

const message = hello();
</script>
