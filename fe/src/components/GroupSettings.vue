<template>
  <SettingsCard :title="lang.email_folders" :description="lang.email_folders_desc">

    <div class="tree-container">
      <el-tree
        :expand-on-click-node="false"
        :data="data"
        :defaultExpandAll="true"
        class="custom-tree"
      >
        <template #default="{ node, data }">
          <div class="tree-node-content">
            <div class="node-label">
              <el-icon class="folder-icon"><Folder /></el-icon>
              <span v-if="data.id !== -1">{{ data.label }}</span>
              <el-input
                v-if="data.id === -1"
                v-model="data.label"
                size="small"
                :placeholder="lang.folder_name"
                class="node-input"
                @blur="onInputBlur(data)"
                @keyup.enter="onInputBlur(data)"
                ref="newInput"
                autofocus
              ></el-input>
            </div>
            
            <div class="node-actions" v-if="data.id !== 0">
              <el-button @click.stop="add(data)" size="small" type="primary" text bg class="action-btn">
                <el-icon><Plus /></el-icon>
              </el-button>
              <el-button @click.stop="del(node, data)" size="small" type="danger" text bg class="action-btn">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
          </div>
        </template>
      </el-tree>
    </div>

    <div class="settings__form-actions">
      <el-button type="primary" @click="addRoot" class="add-root-btn" plain>
        <el-icon><Plus /></el-icon> {{ lang.add_group }}
      </el-button>
    </div>
  </SettingsCard>
</template>

<script setup lang="ts">
import { reactive } from "vue";
import lang from "../i18n/i18n";
import { groupService } from "@/services/groupService";
import { ElMessage } from "element-plus";
import { Folder, Delete, Plus } from "@element-plus/icons-vue";
import SettingsCard from "@/components/settings/SettingsCard.vue";

interface TreeNode {
  children?: TreeNode[];
  label: string;
  id: number;
  parent_id: number;
}

const data = reactive<TreeNode[]>([]);

/** 通过 groupService 获取分组树 */
groupService.getGroupTree().then((res: any) => {
  data.push(...res.data);
});

const del = function (node: any, dataObj: TreeNode) {
  if (dataObj.id !== -1) {
    /** 通过 groupService 删除分组 */
    groupService.deleteGroup(dataObj.id).then((res: any) => {
      if (res.errorNo !== 0) {
        ElMessage({ message: res.errorMsg, type: "error" });
      } else {
        const pc = node.parent.childNodes;
        for (let i = 0; i < pc.length; i++) {
          if (pc[i].id === node.id) {
            pc.splice(i, 1);
            return;
          }
        }
      }
    });
  } else {
    const pc = node.parent.childNodes;
    for (let i = 0; i < pc.length; i++) {
      if (pc[i].id === node.id) {
        pc.splice(i, 1);
        return;
      }
    }
  }
};

const add = function (item: TreeNode) {
  if (item.children == null) {
    item.children = [];
  }
  item.children.push({
    children: [],
    label: "",
    id: -1,
    parent_id: item.id,
  });
};

const addRoot = function () {
  data.push({
    children: [],
    label: "",
    id: -1,
    parent_id: 0,
  });
};

const onInputBlur = function (item: TreeNode) {
  if (item.label !== "") {
    /** 通过 groupService 添加分组 */
    groupService.addGroup(item.label, item.parent_id)
      .then((res: any) => {
        if (res.errorNo !== 0) {
          ElMessage({ message: res.errorMsg, type: "error" });
        } else {
          groupService.getGroupTree().then((res: any) => {
            data.splice(0, data.length);
            data.push(...res.data);
          });
        }
      });
  }
};
</script>

<style scoped>
/* 引入设置组件公共样式（settings__form-actions 等） */
@import '@/assets/settings-common.css';

.tree-container {
  background: var(--ifm-background-surface-color);
  border: 1px solid var(--ifm-border-color);
  border-radius: var(--ifm-global-radius);
  padding: 16px;
  margin-bottom: 24px;
}

.custom-tree {
  background: transparent;
}

.tree-node-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  padding-right: 16px;
}

.node-label {
  display: flex;
  align-items: center;
  font-size: 14px;
  color: var(--ifm-color-content);
}

.folder-icon {
  margin-right: 8px;
  color: var(--ifm-color-primary);
  font-size: 16px;
}

.node-input {
  width: 200px;
  margin-left: 8px;
}

.node-actions {
  display: flex;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.2s;
}

.tree-node-content:hover .node-actions {
  opacity: 1;
}

.action-btn {
  padding: 4px 8px;
  height: 24px;
}

.add-root-btn {
  border-radius: var(--ifm-global-radius);
}

/* Override element-plus tree hover styles */
:deep(.el-tree-node__content) {
  height: 40px;
  border-radius: var(--ifm-global-radius);
  margin-bottom: 4px;
}

:deep(.el-tree-node__content:hover) {
  background-color: var(--ifm-background-hover-color);
}
</style>
