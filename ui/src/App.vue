<template>
  <el-container>
    <el-header style="text-align: right">
      <el-button type="success" :icon="Plus" @click="isAdd=true;current={};drawer=true">添加</el-button>
      <el-button type="success" :icon="Refresh">刷新</el-button>
    </el-header>
    <el-main>
      <el-table :data="data.tableData" style="width: 100%" stripe table-layout="auto">
        <el-table-column label="" align="right" header-align="right" width="40">
          <template #default="scope">
            <el-switch v-model="scope.row.state" inline-prompt active-text="代理" inactive-text="代理"/>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="名称" align="left" header-align="left" width="180">
          <template #default="scope">{{ scope.row.name }}</template>
        </el-table-column>
        <el-table-column prop="protocol" label="协议" align="center" header-align="center" sortable/>
        <el-table-column prop="port" label="端口" align="center" header-align="center" sortable/>
        <el-table-column prop="to" label="目标转发地址" align="center" header-align="center"/>
        <el-table-column label="Consul状态" align="center" header-align="center">
          <template #default="scope">
            <el-switch v-model="scope.row.consulRegister" inline-prompt
                       :active-text="(scope.row.ip ? scope.row.ip:'') + ':' + scope.row.port" inactive-text=""/>
          </template>
        </el-table-column>
        <el-table-column fixed="right" label="操作" align="center" header-align="center">
          <template #default="scope">
            <el-button type="warning" size="small" @click="modify(scope.row)" :icon="Edit">
              修改
            </el-button>
            <el-button type="danger" size="small" @click="deleteClick(scope.row)" :icon="Delete">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-drawer
          v-model="drawer"
          size="40%"
          :title="isAdd?'新增配置':'修改配置'"
          :direction="'rtl'"
          :before-close="handleClose">
        <el-input v-model="name" placeholder="服务名称" :disabled="!isAdd">
          <template #prepend>服务名称</template>
        </el-input>
        <el-input
            v-model="port"
            placeholder="端口"
            type="number">
          <template #prepend>
            <el-select v-model="protocol" class="m-2" placeholder="Select" size="large">
              <el-option label="TCP" value="TCP"/>
              <el-option label="UDP" value="UDP"/>
              <el-option label="HTTP(别样的实现)" value="HTTP"/>
              <template #prefix>协议</template>
            </el-select>
          </template>
        </el-input>
        <el-input
            v-model="ip"
            :disabled="!defaultIp"
            placeholder="绑定的IP地址,默认localhost">
          <template #prepend>
            <el-switch v-model="defaultIp" active-text="自定义" inactive-text="默认"/>
          </template>
        </el-input>
        <el-input v-model="to" placeholder="目标转发地址(含端口)，例如: www.baidu.com:80">
          <template #prepend>目标转发</template>
        </el-input>
        <el-button @click="saveProxy">保存</el-button>
      </el-drawer>
    </el-main>
    <el-footer></el-footer>
  </el-container>
</template>

<script>
import {ElMessageBox} from 'element-plus'
import {defineStore} from 'pinia'
import {ref} from 'vue'
import {Delete, Edit, Plus, Refresh} from "@element-plus/icons-vue";

const initData = [
  {
    name: 'db',
    protocol: 'UDP',
    ip: "localhost",
    port: 220,
    to: 'www.baidu.com:90',
    state: false,
    consulRegister: false,
  },
  {
    name: 'redis',
    protocol: 'TCP',
    ip: "localhost",
    port: 230,
    to: 'www.baidu.com:80',
    state: false,
    consulRegister: false,
  },
]
export const dataStore = defineStore('dataStore', {
  state: () => {
    return {
      editing: {},
      tableData: initData
    }
  },
  actions: {
    saveProxy: (proxy) => {
      this.state.tableData.push(proxy)
    }
  },
})
export default {
  name: 'App',
  computed: {
    Edit() {
      return Edit
    },
    Plus() {
      return Plus
    },
    Refresh() {
      return Refresh
    },
    Delete() {
      return Delete
    }
  },
  setup() {
    const data = dataStore()
    const isAdd = ref(false)
    const drawer = ref(false)
    const defaultIp = ref(false)


    const name = ref("")
    const protocol = ref("")
    const ip = ref("localhost")
    const port = ref(0)
    const to = ref("")


    const deleteClick = (row) => {
      data.tableData = data.tableData.filter(function (value) {
        return value.name !== row.name
      });
      console.log("删除" + row)
    }
    const initDrawer = (isModify, proxy) => {
      isAdd.value = !isModify
      if (isModify && proxy) {
        name.value = proxy.name
        protocol.value = proxy.protocol
        port.value = proxy.port
        to.value = proxy.to
        ip.value = proxy.ip
        defaultIp.value = proxy.ip === "localhost"
      } else {
        name.value = "unnamed"
        protocol.value = "TCP"
        ip.value = "localhost"
        port.value = 0
        to.value = ""
      }
    }
    const modify = (row) => {
      isAdd.value = false
      name.value = row.name
      protocol.value = row.protocol
      port.value = row.port
      to.value = row.to
      drawer.value = true
    }
    const handleClose = (done) => {
      ElMessageBox.confirm('未保存，确认关闭', '提示')
          .then(() => {
            initDrawer(false)
            done()
          })
          .catch(() => {
          })
    }
    const saveProxy = () => {
      try {
        if (isAdd.value) {
          const newData = {
            name: name.value,
            protocol: protocol.value,
            ip: ip.value,
            port: port.value,
            to: to.value,
          }
          data.tableData.push(newData)

        }else{
          data.tableData.forEach((value)=>{
            if(value.name===name.value){
              value.protocol = protocol.value
              value.port = port.value
              value.ip = ip.value
              value.to = to.value
            }
          })
        }
        drawer.value = false
      } catch (e) {
        console.log(e)
      }
    }
    return {
      data,

      name, protocol, ip, port, to,

      isAdd, drawer, defaultIp,

      initDrawer,

      deleteClick, modify, handleClose, saveProxy
    }
  },
}
</script>

<style>
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
  margin-top: 60px;
}
</style>
