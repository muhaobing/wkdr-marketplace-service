# 韦克德瑞官方商城 - 前端

基于 Vue 3 + Vite + Pinia 构建的商城前端应用。

## 技术栈

- Vue 3 - 渐进式 JavaScript 框架
- Vue Router - 路由管理
- Pinia - 状态管理
- Axios - HTTP 客户端
- Vite - 构建工具

## 功能特性

- **首页商城** - 商品列表展示，支持搜索
- **商品详情** - 查看商品信息，支持加入购物车和直接下单
- **购物车** - 本地缓存，支持选择商品批量下单
- **订单中心** - 查看订单列表和详情
- **订单操作** - 支持取消订单和支付

## 快速开始

```bash
# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 构建生产版本
npm run build
```

## 目录结构

```
web/
├── public/              # 静态资源
├── src/
│   ├── api/            # API 接口封装
│   ├── components/     # 公共组件
│   │   └── Navbar.vue  # 导航栏
│   ├── router/         # 路由配置
│   ├── stores/         # 状态管理
│   │   ├── cart.js     # 购物车状态
│   │   └── user.js     # 用户状态
│   ├── views/          # 页面组件
│   │   ├── Home.vue         # 首页
│   │   ├── ProductDetail.vue # 商品详情
│   │   ├── Cart.vue         # 购物车
│   │   ├── Orders.vue       # 订单列表
│   │   └── OrderDetail.vue  # 订单详情
│   ├── App.vue         # 根组件
│   ├── main.js         # 入口文件
│   └── style.css       # 全局样式
├── index.html          # HTML 模板
├── package.json        # 项目配置
└── vite.config.js      # Vite 配置
```

## API 接口

前端调用的后端接口（/marketplace 前缀）：

| 接口 | 方法 | 说明 |
|------|------|------|
| /skus | GET | 获取商品列表 |
| /skus/:id | GET | 获取商品详情 |
| /orders | POST | 创建订单 |
| /orders | GET | 获取订单列表 |
| /orders/:order_no | GET | 获取订单详情 |
| /orders/:order_no/cancel | POST | 取消订单 |
| /orders/:order_no/pay | POST | 支付订单 |
| /ecoin | GET | 获取用户积分 |
| /payment-methods | GET | 获取支付方式 |

## 设计规范

- **主色调**: 深蓝色 (#1a365d)
- **背景色**: 白色/浅灰 (#ffffff / #f8fafc)
- **风格**: 简约现代，卡片式布局

## Mock 数据

开发模式下，当后端接口不可用时，会自动使用 Mock 数据。Mock 用户 ID 为 10001。
