import {defineStore} from "pinia";
import {userService} from "@/services/userService";
import type {UserInfo} from "@/types/api";

const useGlobalStatusStore = defineStore('useGlobalStatusStore', {
    state() {
        return {
            userInfos: {} as UserInfo,
            mobileDrawerVisible: false,
            settingsDrawerVisible: false
        }
    },
    getters: {
        isLogin(state): boolean {
            return Object.keys(state.userInfos).length !== 0
        }
    },
    actions: {
        /** 初始化用户信息：通过 userService 获取当前用户数据 */
        init(callback: () => void) {
            userService.getUserInfo().then((res: any) => {
                if (res.errorNo === 0) {
                    Object.assign(this.userInfos, res.data)
                    callback()
                }
            })
        }
    }
})


export {useGlobalStatusStore};
