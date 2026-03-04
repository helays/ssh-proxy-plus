// axios.ts
import axios, {
    AxiosError,
    type AxiosResponse,
} from 'axios';
import type {AxiosRequestConfig} from 'axios';
import * as qs from 'qs';
import type {Ref} from "vue";
import {ThenErrorMsgNotification, CatchMessageNotification} from '@/components/ts/notification'
import {MapVal, validStatas} from "@/components/ts/utils";
import {ElLoading, ElMessage, ElNotification} from "element-plus";


// 判断响应内容 是否是未登录
export const isLogin=(err:any)=>{
    let resp=err.response;
    if (resp.data?.code===401){
        localStorage.removeItem('login_status'); // 如果是未登录，就直接删除本地登录状态，后续这里还需要优化，自动跳转或者弹窗 // todo
        return
    }
}


// 正常响应拦截器
const successResp = (response: AxiosResponse) => {
    // 访问正常，预定义的拦截器占位
    return response;
}

// 异常响应拦截器
const errorResp = (error: AxiosError) => {
    const {status} = error.response ?? {}; // 使用空对象解构避免 undefined
    // 返回错误 promise，以便在调用处可以继续处理错误
    return Promise.reject(error);
}

// query参数 编码器
const encodeParams = (params: any) => {
    return qs.stringify(params, {
        encode: false,
        // allowDots: true, // 允许使用点号表示对象的属性
        // arrayFormat: 'repeat', // 处理数组的方式
        // skipNull: true, // 跳过 null 值
        // skipEmptyString: true, // 跳过空字符串
        // encodeValuesOnly: true // 只对值进行编码，而不是键
    })
}

const axiosService = axios.create({
    paramsSerializer: encodeParams
    // baseURL: import.meta.env.VITE_API_BASE_URL,
    // timeout: 5000,
});
// 设置响应拦截器
axiosService.interceptors.response.use(successResp, errorResp);

// 通用get 获取列表的方法
const listGetHasPage = (url: string, config: Record<string, any>, listTotal: Ref<number>, objectList: Record<any, any>[], userConfig: any) => {
    if (!config.params) {
        config.params = {}
    }
    let enablePage=userConfig?.enablePagination !== false;
    if (enablePage){
        let pageField= userConfig.pageField? userConfig.pageField : 'page'
        let pageSizeField= userConfig.pageSizeField? userConfig.pageSizeField : 'page_size'
        config.params[pageField] = MapVal(config.params, pageField, 1);
        config.params[pageSizeField] = MapVal(config.params, pageSizeField, 15);
    }

    config['url']= url;
    axiosService.request(config).then((resp: any) => {
        let data = resp.data;
        listTotal.value = 0
        objectList.splice(0, objectList.length)
        if (validStatas(data)!==0){
            ThenErrorMsgNotification(data)
            return
        }
        listTotal.value = parseInt(data.data.total)
        objectList.push(...data.data.lists)
    }).catch(err => {
        CatchMessageNotification(err)
        isLogin(err)
    }).finally(() => {
        userConfig.loadding.value = false;
    })
}


interface SendRequestConfig {
    request: AxiosRequestConfig;
    loading?: boolean;
    disableErrorNotice?: boolean;
    disableSuccessNotice?: boolean;
    disableCatchNotice?: boolean;
    error?:Function;
    success?:Function;
    catch?:Function;
    finally?:Function;
    download?:Function;
}

export const sendRequest=(config:SendRequestConfig)=>{
    let loadingInstance:any
    if (config.loading){
        loadingInstance = ElLoading.service({fullscreen: true})
    }

    axiosService.request(config.request).then((resp: AxiosResponse) => {
        if (config.download && typeof config['download']==="function"){
            config['download'](resp);
            return;
        }
        let data=resp.data;
        if (validStatas(data)!==0){
            if (typeof config['disableErrorNotice']==="undefined" || config['disableErrorNotice']!==true){
            ThenErrorMsgNotification(data)
            }

            if (config['error'] && typeof config['error']==="function"){
                config['error'](data,resp)
            }
            return
        }
        if (typeof config['disableSuccessNotice']==="undefined" || config['disableSuccessNotice']!==true){
            ElMessage({
                message: data['msg'],
                type: 'success',
            })
        }
        if (config['success'] && typeof config['success']==="function"){
            config['success'](data,resp)
        }
    }).catch((err: any) => {
        if (typeof config['disableCatchNotice']==="undefined" || config['disableCatchNotice']!==true){
            CatchMessageNotification(err)
        }
        if (config['catch'] && typeof config['catch']==="function"){
            config['catch'](err)
        }
        isLogin(err)
    }).finally(() => {
        if (config.loading){
            loadingInstance.close()
        }
        if (config['finally'] && typeof config['finally']==="function"){
            config['finally']()
        }
    })
}


export {axiosService, listGetHasPage}