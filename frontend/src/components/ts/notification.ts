import {ElNotification} from "element-plus";
import {blobToText} from "@/components/ts/utils";

export const CatchMessageNotification=async (err: any) => {
    let message = err.message;
    if (err.response && err.response.data) {
        let responseData = err.response.data;
        if (responseData.msg) {
            message = err.response.data.msg;
        } else if (responseData instanceof Blob) {
            // 将 Blob 转换为文本（使用 async/await 模拟同步）
            const text = await blobToText(responseData);
            try {
                const respInfo = JSON.parse(text.toString());
                if (respInfo.msg){
                    message=respInfo.msg
                }
            } catch (e) {
                console.log(JSON)
            }
        }
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