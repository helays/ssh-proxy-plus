import {h} from "vue";
import {ElMessage, ElNotification} from "element-plus";
// export function deepCopyJson<T>(jsonObj:T):T{
//     return JSON.parse(JSON.stringify(jsonObj))
// }

// utils.ts
export const deepCopyJson = <T>(jsonObj: T): T => JSON.parse(JSON.stringify(jsonObj));

export const handleCopy = (text: string) => {
    if (typeof navigator.clipboard !== 'undefined'){
        try {
            navigator.clipboard.writeText(text);
            ElMessage({
                message: "复制成功！",
                type: "success",
            })
        }catch (error){
            ElMessage({
                message: "复制失败！",
                type: "error",
            })
            console.error(error);
        }
        return
    }
    const tempInput = document.createElement('input');
    tempInput.value = text;
    document.body.appendChild(tempInput);
    tempInput.select();
    document.execCommand('copy');
    document.body.removeChild(tempInput);
    ElMessage({
        message: "复制成功！",
        type: "success",
    })
};

export const MapVal=(params: Record<string, any>, key: string, val: any): any=>{
    return (params && params[key]) ? params[key] : val
}

// 辅助函数：将 Blob 转换为文本
export const blobToText = (blob: Blob): Promise<string> => {
    return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onloadend = () => {
            if (typeof reader.result === 'string') {
                resolve(reader.result);
            } else {
                reject(new Error('Unexpected result type'));
            }
        };
        reader.onerror = () => {
            reject(reader.error);
        };
        reader.readAsText(blob);
    });
}

// 辅助函数：验证状态
// 返回值：0 表示成功，-999 表示 data 不存在，-998 表示 data[field]无效 不存在，其他值表示状态码
// @param data: any
// @param field: string
export const validStatas = (data:any,field:string='code'): number => {
    // 如果 data 或者 data[field] 不存在，返回 false
    if (!data || data[field]===undefined){
        return -999;
    }

    const code = Number(data[field]);
    if (isNaN(code)) {
        return -998;
    }
    // 返回状态码是否为 0 或 2xx
    return (code === 0 || (code >= 200 && code < 300)) ? 0 :code ;
}

export const decodeUnicodeEscapeSequences=(str: string): string => {
    return str.replace(/\\u([0-9a-fA-F]{4})/g, function(match, grp) {
        return String.fromCharCode(parseInt(grp, 16));
    });
}


export const DownloadFile = (response: any, filename: string) => {
    // 从响应头中读取文件名
    const contentDisposition = response.headers['content-disposition'];
    if (contentDisposition) {
        // 尝试匹配 RFC 5987 格式的文件名
        const matchesRFC5987 = contentDisposition.match(/filename\*=(?:UTF-8''|)([^;]*)/);
        if (matchesRFC5987 && matchesRFC5987[1]) {
            // 解码文件名
            filename = decodeURIComponent(matchesRFC5987[1]);
        } else {
            // 尝试匹配普通的文件名格式
            const matches = contentDisposition.match(/filename=["']?([^"']+)["']?/);
            if (matches && matches[1]) {
                // 解码文件名并去除前导下划线
                filename = decodeURIComponent(matches[1]).replace(/^_/, '');
            }
        }
    }
    // 创建一个 Blob 对象
    const blob = new Blob([response.data], {type: 'application/octet-stream'});
    // 创建一个隐藏的 <a> 元素
    const link = document.createElement('a');
    link.href = URL.createObjectURL(blob);
    link.download = filename; // 使用从响应头中读取的文件名
    // 触发点击事件
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
}

interface ClickSelectFileConfig {
    accept:string;
    success: (files: FileList)=>void
}

export const ClickSelectFile=(config:ClickSelectFileConfig)=>{
    let input = document.createElement('input');
    input.setAttribute('type', 'file');
    input.setAttribute('accept', config.accept);
    input.value = '';
    input.click();
    input.onerror =  (msg)=> {
        ElNotification({
            title: '异常',
            message: h('i', {style: 'color: error'}, '文件打开失败'),
        })
        console.error(msg)
    }
    input.onchange = function () {
        if (!input.files || input.files.length < 1) {
            return
        }
        if (config.success && typeof config.success==="function"){
            config.success(input.files);
        }
    };
}

export const StringToArray=(str:any, delimiter:string):any=>{
    if (!str){
        return []
    }
    return (str as String).split(delimiter)
}

export const ArrayToString=(arr:any, delimiter:string):any=>{
    if (!arr){
        return ''
    }
    if (!Array.isArray(arr)){
        return arr
    }
    return arr.join(delimiter)
}