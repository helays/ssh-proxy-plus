import {
    createRouter,
    type RouteLocationNormalized,
    createWebHashHistory
} from 'vue-router'
import {fetchRoutes} from "@/router/routeUtils";

// 异步组件加载函数，用于按需加载路由组件
// const loadComponent = (componentName:any) => {
//   defineAsyncComponent(() => import(`@/views/${componentName}.vue`))
// };
const loadComponent = (componentName: string) => () => import(`@/views/${componentName}.vue`);


const router = createRouter({
    // history: createWebHistory(import.meta.env.BASE_URL),
    history: createWebHashHistory(), // 使用 hash方式
    routes: [
        {
            path: '/page/login',
            name: '登录',
            component: ()=>import('@/layouts/LoginLayout.vue'),
            meta: {
                title: '登录',
                require_auth: false,
                mod: 'login'
            }
        },
        {
            path: '/',
            name: 'indexpage',
            component: ()=>import('@/layouts/MainLayout.vue'),
            meta:{
                require_auth:true,
            }
        },
    ]
})

await fetchRoutes(router)



// router.beforeEach((to: RouteLocationNormalized, from: RouteLocationNormalized, next: NavigationGuardNext) => {
//     // 开启登录，并且对应的路由需要登录才能访问
//     if ((window as any)['enable_pass'] == 'on'){
//         let loginStatus=localStorage.getItem("login_status");
//         if (to.matched.some(record => record.meta.require_auth)){
//             if (loginStatus!=='login') { // 这里通过判断全局登录状态的方式，判断用户是否登录
//                 next({path: '/page/login'});
//                 return
//             }
//         }
//         if (to.matched.some(record => record.name == '登录')){
//             if ((loginStatus==='login')) { // 这里通过判断全局登录状态的方式，判断用户是否登录
//                 next({path: '/'});
//                 return
//             }
//         }
//     }
//
//     let title: string = !to.name ? '' : to.name as string;
//     if (to.meta !== undefined && to.meta.title !== undefined) {
//         title = to.meta.title as string
//     }
//
//     document.title = title + " - SSH隧道代理服务";
//     next();
// })

router.beforeEach(async (to: RouteLocationNormalized, from: RouteLocationNormalized) => {
    // 开启登录，并且对应的路由需要登录才能访问
    if ((window as any)['enable_pass'] == 'on'){
        let loginStatus=localStorage.getItem("login_status");
        if (to.matched.some(record => record.meta.require_auth)){
            if (loginStatus!=='login') { // 这里通过判断全局登录状态的方式，判断用户是否登录
                return {path: '/page/login'}
            }
        }
        if (to.matched.some(record => record.name == '登录')){
            if ((loginStatus==='login')) { // 这里通过判断全局登录状态的方式，判断用户是否登录
                return {path: '/'}
            }
        }
    }

    let title: string = !to.name ? '' : to.name as string;
    if (to.meta !== undefined && to.meta.title !== undefined) {
        title = to.meta.title as string
    }

    document.title = title + " - SSH隧道代理服务";
    return true;
})

export default router
