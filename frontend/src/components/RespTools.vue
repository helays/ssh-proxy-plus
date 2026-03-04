<script lang="ts">
import {ElNotification} from "element-plus";

export const CatchMessageNotification=(err:any)=>{
  let message=err.message;
  if (err.response && err.response && err.response.data.msg){
    message=err.response.data.msg;
  }
  ElNotification({
    title: '错误提示',
    message: message,
    type: 'error',
  })
}

export const ThenErrorMsgNotification=(data:any)=>{
  ElNotification({
    title: '错误提示',
    message: data.msg ?? '请求失败',
    type: 'error',
  })
}
export default {
}

</script>