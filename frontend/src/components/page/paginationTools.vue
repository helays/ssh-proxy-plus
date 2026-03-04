<script setup lang="ts">

import {computed, onBeforeMount, ref, watch} from "vue";
import {type LocationQuery, useRoute, useRouter} from "vue-router";
import {deepCopyJson, MapVal} from "@/components/ts/utils";

const props = defineProps({
  total: {type: Number},
  getLists: {type: Function, required: false},
  pageField: {type: String, default: 'page'},
  pageSizeField: {type: String, default: 'page_size'},
  searchModel: {type: Object, default: {}}
})

// const emits=defineEmits(['getLists'])

const router = useRouter();
const route = useRoute();

const pageSize = ref(parseInt(MapVal(route.query, props.pageSizeField, 15)))
const currentPage = ref(parseInt(MapVal(route.query, props.pageField, 1)))

const changeQuery = (): LocationQuery => {
  let query: LocationQuery = deepCopyJson(route.query);
  // 然后将search model的数据赋值上去
  for (const key in props.searchModel) {
    if ( props.searchModel[key] !== '') {
      query[key] = props.searchModel[key]
    }
  }
  return query;
}

const handleCurrentChange = (val: number) => {
  let query: LocationQuery = changeQuery();
  query[props.pageField] = val.toString()
  query[props.pageSizeField] = pageSize.value.toString()
  currentPage.value = val
  router.push({
    path: route.fullPath,
    query: query
  });
}

const handleSizeChange = (val: number) => {
  let query: LocationQuery = changeQuery();
  pageSize.value = val
  query[props.pageSizeField] = val.toString()
  router.push({
    path: route.path,
    query: query
  });
}

// 监测路径变化
watch(() => route.fullPath, (newValue, oldValue) => {
  // 将当前页、每页数量 反写到对象上面
  currentPage.value = parseInt(route.query[props.pageField]?.toString() || '1') || 1;
  pageSize.value = parseInt(route.query[props.pageSizeField]?.toString() || '15') || 15;

  if (props.getLists && typeof props.getLists==="function"){
    props.getLists();
  }
})

</script>

<template>
  <div style="margin-top: 1em;display: flex;justify-content: center" class="text-center">
    <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[15, 30, 50, 100]"
        :size="'default'"
        layout="total, prev, pager, next,sizes"
        :total="props.total"
        :hide-on-single-page="true"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"/>
  </div>
</template>

<style scoped>

</style>