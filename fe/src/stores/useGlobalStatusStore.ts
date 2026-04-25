import {defineStore} from "pinia";
import {http} from "@/utils/axios";
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
        init(callback: () => void) {
            http.post("/api/user/info", {}).then((res: any) => {
                if (res.errorNo === 0) {
                    Object.assign(this.userInfos, res.data)
                    console.log("userInfos")
                    callback()
                }
            })
        }
    }
})


export {useGlobalStatusStore};
