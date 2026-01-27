// routeUtils.ts 示例
import {type Router, type RouteRecordRaw, useRouter} from 'vue-router';
import {CatchMessageNotification, ThenErrorMsgNotification} from '@/components/RespTools.vue'
import { axiosService } from "@/components/ts/axios";


// 获取路由列表
export async function fetchRoutes(router:Router|undefined|null) {
    try {
        const result = await axiosService.get('/api/v1/run.menu.lists');
        let response=result.data;
        if (response.code !== 0 && response.code !== 200) {
            ThenErrorMsgNotification(response.data.msg)
            return [];
        }
        let enable_pass=response.data.enable_pass;
        (window as any)['enable_pass'] = enable_pass;
        router=router??useRouter();
        if (!response.data.islogin){
            localStorage.removeItem('login_status');
        }

        let indexRoute = router.options.routes.find(route => route.name === 'indexpage');
        if (indexRoute && indexRoute.children){
            indexRoute.children = []; // 清空子路由数组
        }

        if (response.data && response.data.menu){
            localStorage.setItem("menu",JSON.stringify(response.data.menu))
            if (enable_pass=='on'){
                let loginStatus=localStorage.getItem("login_status");
                if (loginStatus==='login') { // 这里通过判断全局登录状态的方式，判断用户是否登录
                    await addDynamicRoutes(router,response.data.menu,"indexpage");
                }

            }else{
                await addDynamicRoutes(router,response.data.menu,"indexpage");
            }

        }
    } catch (error) {
        CatchMessageNotification(error)
        return [];
    }
}


 async function addDynamicRoutes(router:Router,dynamicRoutes: Record<any, any>[],name:string): Promise<void> {
    return new Promise((resolve) => {
        dynamicRoutes.forEach(route => {
            // 根据实际情况构造路由对象
            const routeObject:RouteRecordRaw = {
                path: route.path,
                name: route.name || route.path,
                component: () => import(`../views/${route.component}.vue`),
                meta: route.meta || {}, // 可以用来存储权限信息等
            };

            // 如果有子路由，递归处理
            if (route.children && route.children.length > 0) {
                addDynamicRoutes(router,route.children,route.name);
            }
            let parentRoute = router.getRoutes().find(r => r.name === name);

            if (parentRoute) {
                router.addRoute(name, routeObject); // 'main' 是一个假设的父路由名称，根据你的路由结构可能需要调整
            }

        });
        resolve();
    });
}