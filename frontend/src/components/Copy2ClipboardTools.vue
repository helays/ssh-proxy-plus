<script lang="ts">
import {ElMessage} from "element-plus";

export const TxtToClipboard=(txt:any)=>{
  navigator.clipboard.writeText(txt).then(()=>{
    ElMessage({
      message: "复制成功！",
      type: "success",
    })
  }).catch((err)=>{
    ElMessage({
      message: "复制失败！",
      type: "error",
    })
  });
}

export default {
}
</script>