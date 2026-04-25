<script setup lang="ts" name="KbdHint">
defineProps<{ shortcut: string; show: boolean; inline?: boolean; center?: boolean }>();
</script>

<template>
    <kbd
        v-if="show"
        :class="
            inline
                ? 'text-[9px] font-sans bg-gray-800 text-white rounded px-1 leading-[14px] pointer-events-none shadow shrink-0'
                : center
                ? 'absolute -top-2 left-1/2 -translate-x-1/2 text-[9px] font-sans bg-gray-800 text-white rounded px-1 leading-[14px] pointer-events-none shadow'
                : 'absolute -top-2 -right-1 text-[9px] font-sans bg-gray-800 text-white rounded px-1 leading-[14px] pointer-events-none shadow'
        "
        >{{ shortcut }}</kbd
    >
</template>
