import axios from 'axios'
import lang from '../i18n/i18n';
import {useGlobalStatusStore} from "@/stores/useGlobalStatusStore";
import {router} from "@/router";
import type {ApiResponse} from "@/types/api";

//创建axios的一个实例
const http = axios.create({
    baseURL: import.meta.env.VITE_APP_URL, //接口统一域名
    timeout: 60000, //设置超时
    headers: {
        'Content-Type': 'application/json;charset=UTF-8;',
        'Lang': lang.lang
    }
});

//请求拦截器
http.interceptors.request.use((config) => {
    //若请求方式为post，则将data参数转为JSON字符串
    if (config.method === 'POST') {
        config.data = JSON.stringify(config.data);
    }
    return config;
}, (error) =>
    // 对请求错误做些什么
    Promise.reject(error));

//响应拦截器
http.interceptors.response.use(async (response): Promise<any> => {
    //响应成功
    if (response.data.errorNo === 403) {

        await router.replace({
            path: '/login',
            query: {
                redirect: router.currentRoute.value.fullPath
            }
        });
    }
    //响应成功
    // 修改日期: 20260504，增加当前路由判断：
    // 当已在 /setup 页面时跳过重定向，防止 redirect 参数无限嵌套导致死循环。
    // 场景：Setup 页面的 API 调用返回 402 时，不应再次重定向到 /setup。
    if (response.data.errorNo === 402 && router.currentRoute.value.path !== '/setup') {
        await router.replace({
            path: '/setup',
            query: {
                redirect: router.currentRoute.value.fullPath
            }
        });
    }
    return response.data as ApiResponse;
}, async (error) => {
    //响应错误：仅处理 403 登录态失效重定向，其余错误统一透传给调用方（错误提示由后端 errorMsg 承担）
    if (error.response && error.response.status === 403) {
        await router.replace({
            path: '/login',
            query: {
                redirect: router.currentRoute.value.fullPath
            }
        });
    }
    return Promise.reject(error);
});

export {http};
