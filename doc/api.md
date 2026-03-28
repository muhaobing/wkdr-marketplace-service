# WKDR 商城服务 - 接口协议文档

## 通用说明

### 基础信息

| 项目 | 说明 |
|------|------|
| 基础地址 | `http://{host}:9090` |
| Content-Type | `application/json` |
| 字符编码 | UTF-8 |

### 统一响应格式

所有业务接口（支付回调除外）统一使用以下响应格式：

```json
{
  "retcode": 0,
  "message": "success",
  "data": {}
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| retcode | int | 状态码，0 表示成功，非 0 表示失败 |
| message | string | 状态描述 |
| data | object/null | 业务数据，失败时为 null |

### 鉴权方式

请求头携带登录 Token：

```
Authorization: Bearer {token}
```

### 鉴权规则

| 路径 | 鉴权要求 |
|------|----------|
| `/marketplace/login` | 免鉴权 |
| `/openapi/user/bind` | 免鉴权 |
| `/openapi/callback/*` | 免鉴权（第三方回调） |
| `/mock/*` | 免鉴权（测试接口） |
| `/ping` | 免鉴权 |
| `/marketplace/*` | 用户登录鉴权 |
| `/openapi/*`（除 callback 等） | JWT 鉴权（见部署配置） |
| `/ops/*` | 管理员权限（Admin） |

### 通用枚举

**订单状态 (status)**

| 值 | 说明 |
|----|------|
| 0 | 待支付 |
| 1 | 已支付 |
| 2 | 已履约 |
| 3 | 已取消 |
| 4 | 已退款 |

**商品上架状态 (sku_status)**

| 值 | 说明 |
|----|------|
| 0 | 未上架 |
| 1 | 已上架 |

**支付类型 (pay_type)**

| 值 | 说明 |
|----|------|
| ecoin | 积分支付 |
| money | 货币支付 |

**支付渠道 (channel)**

| 值 | 说明 |
|----|------|
| ecoin | 积分支付 |
| wechat | 微信支付 |

**支付方式 (pay_method)**

| 值 | 说明 |
|----|------|
| native | 扫码支付 |

---

## 一、商城接口（Marketplace）

### 1.1 用户登录

**POST** `/marketplace/login`

> 鉴权：免鉴权

**请求参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| tel_no | string | 否* | 手机号（与 email 二选一） |
| email | string | 否* | 邮箱（与 tel_no 二选一） |
| secret | string | 是 | 用户密钥 |

**响应 data**

| 字段 | 类型 | 说明 |
|------|------|------|
| token | string | 登录 Token |
| user_id | uint | 用户 ID |
| user | object | 用户信息 |

---

### 1.2 商品列表

**GET** `/marketplace/skus`

> 鉴权：用户登录

**查询参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| sku_name | string | 否 | 商品名称（模糊查询） |
| offset | int | 否 | 偏移量，默认 0 |
| limit | int | 否 | 每页数量，默认 20 |

> 仅返回已上架商品

**响应 data**

| 字段 | 类型 | 说明 |
|------|------|------|
| total | int64 | 总数 |
| list | array | 商品列表 |
| list[].id | uint64 | 商品 ID |
| list[].biz_code | string | 业务编码 |
| list[].sku_code | string | 商品代码 |
| list[].sku_name | string | 商品名称 |
| list[].sku_avatar | string | 商品图标 URL |
| list[].sku_desc | string | 商品描述 |
| list[].sku_status | uint8 | 上架状态 |
| list[].cost | float32 | 商品售价（元） |
| list[].multi_select | uint8 | 是否支持多选下单：0-否 1-是 |
| list[].delivery_method | string | 履约回调 URL |
| list[].ctime | uint32 | 创建时间戳 |
| list[].mtime | uint32 | 更新时间戳 |

---

### 1.3 商品详情

**GET** `/marketplace/skus/:id`

> 鉴权：用户登录

**路径参数**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint64 | 商品 ID |

**响应 data**

与商品列表中单个商品结构相同。

---

### 1.4 创建订单

**POST** `/marketplace/checkout`

> 鉴权：用户登录

**请求参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| user_id | uint64 | 是 | 用户 ID |
| sku_items | array | 是 | 商品列表 |
| sku_items[].sku_id | uint64 | 是 | 商品 ID |
| sku_items[].quantity | int | 否 | 购买数量，默认 1 |
| pay_type | string | 否 | 支付类型，默认 money |
| remark | string | 否 | 备注 |

**响应 data**

| 字段 | 类型 | 说明 |
|------|------|------|
| order | object | 订单信息 |
| order.order_no | string | 订单号 |
| order.user_id | uint64 | 用户 ID |
| order.item_count | int | 商品种类数 |
| order.total_quantity | int | 商品总数量 |
| order.original_amount | float32 | 原价（元） |
| order.pay_amount | float32 | 实付金额（元） |
| order.pay_type | string | 支付类型 |
| order.status | uint8 | 订单状态 |
| order.ctime | uint32 | 创建时间戳 |

---

### 1.5 订单列表

**GET** `/marketplace/orders`

> 鉴权：用户登录

**查询参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| user_id | uint64 | 是 | 用户 ID |
| status | uint8 | 否 | 订单状态过滤 |
| offset | int | 否 | 偏移量，默认 0 |
| limit | int | 否 | 每页数量，默认 20 |

**响应 data**

| 字段 | 类型 | 说明 |
|------|------|------|
| total | int64 | 总数 |
| list | array | 订单列表（含 items 明细） |

---

### 1.6 订单详情

**GET** `/marketplace/orders/:order_no`

> 鉴权：用户登录

**路径参数**

| 字段 | 类型 | 说明 |
|------|------|------|
| order_no | string | 订单号 |

**响应 data**

| 字段 | 类型 | 说明 |
|------|------|------|
| order_no | string | 订单号 |
| user_id | uint64 | 用户 ID |
| original_amount | float32 | 原价（元） |
| pay_amount | float32 | 实付金额（元） |
| pay_type | string | 支付类型 |
| status | uint8 | 订单状态 |
| pay_time | uint32 | 支付时间戳 |
| ctime | uint32 | 创建时间戳 |
| items | array | 订单明细 |
| items[].sku_code | string | 商品代码 |
| items[].sku_name | string | 商品名称 |
| items[].sku_avatar | string | 商品图标 |
| items[].quantity | int | 数量 |
| items[].unit_price | float32 | 单价（元） |
| items[].total_price | float32 | 小计（元） |
| items[].fulfill_status | uint8 | 履约状态 |

---

### 1.7 取消订单

**POST** `/marketplace/orders/:order_no/cancel`

> 鉴权：用户登录

**路径参数**

| 字段 | 类型 | 说明 |
|------|------|------|
| order_no | string | 订单号 |

**请求参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| reason | string | 否 | 取消原因 |

**响应 data**

null（成功即可）

---

### 1.8 支付订单

**POST** `/marketplace/orders/:order_no/pay`

> 鉴权：用户登录

**路径参数**

| 字段 | 类型 | 说明 |
|------|------|------|
| order_no | string | 订单号 |

**请求参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| channel | string | 是 | 支付渠道：ecoin / wechat |
| pay_method | string | 否 | 支付方式（积分支付无需提供，微信支付传 native） |

**响应 data**

| 字段 | 类型 | 说明 |
|------|------|------|
| order_no | string | 订单号 |
| code_url | string | 微信扫码支付二维码 URL（仅微信 native 支付） |

> 积分支付时直接完成支付，无额外返回字段。

---

### 1.9 同步订单状态

**POST** `/marketplace/orders/:order_no/sync`

> 鉴权：用户登录
> 
> 仅查询并同步支付状态，不触发履约。

**路径参数**

| 字段 | 类型 | 说明 |
|------|------|------|
| order_no | string | 订单号 |

**响应 data**

返回最新的订单信息，结构同订单详情。

---

### 1.10 获取积分余额

**GET** `/marketplace/ecoin/balance`

> 鉴权：用户登录

**查询参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| user_id | uint64 | 是 | 用户 ID |

**响应 data**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint64 | 积分账户 ID |
| user_id | uint64 | 用户 ID |
| available_stock | float64 | 可用积分余额 |
| frozen_stock | float64 | 冻结积分 |
| ctime | uint32 | 创建时间戳 |
| mtime | uint32 | 更新时间戳 |

---

### 1.11 积分流水列表

**GET** `/marketplace/ecoin/transactions`

> 鉴权：用户登录

**查询参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| user_id | uint64 | 是 | 用户 ID |
| offset | int | 否 | 偏移量 |
| limit | int | 否 | 每页数量 |

**响应 data**

| 字段 | 类型 | 说明 |
|------|------|------|
| total | int64 | 总数 |
| list | array | 流水列表 |
| list[].id | uint64 | 流水 ID |
| list[].user_id | uint64 | 用户 ID |
| list[].amount | float64 | 变动数量（正数增加，负数扣除） |
| list[].before_stock | float64 | 操作前余额 |
| list[].after_stock | float64 | 操作后余额 |
| list[].tx_type | int | 交易类型：1-增加 2-扣除 |
| list[].source_type | string | 来源类型 |
| list[].source_id | string | 来源业务 ID |
| list[].description | string | 描述 |
| list[].status | int | 状态：0-处理中 1-已完成 2-已失败 |
| list[].ctime | uint32 | 创建时间戳 |

---

### 1.12 获取充值配置

**GET** `/marketplace/ecoin/recharge_config`

> 鉴权：用户登录

**查询参数**

无

**响应 data**

| 字段 | 类型 | 说明 |
|------|------|------|
| unit_price | float32 | 积分单价（元/积分） |
| min_amount | int | 最低充值数量 |

---

### 1.13 积分充值

**POST** `/marketplace/ecoin/recharge`

> 鉴权：用户登录

**请求参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| user_id | uint64 | 是 | 用户 ID |
| amount | int | 是 | 充值积分数量 |
| pay_type | string | 是 | 支付类型（money） |
| channel | string | 否 | 支付渠道（wechat） |
| pay_method | string | 否 | 支付方式（native） |

**响应 data**

| 字段 | 类型 | 说明 |
|------|------|------|
| order | object | 订单信息 |
| payment_info | object | 支付信息（含 code_url 等） |

---

### 1.14 获取支付方式

**GET** `/marketplace/payment-methods`

> 鉴权：用户登录

**响应 data**

```json
[
  { "channel": "ecoin", "name": "积分支付", "pay_method": "ecoin" },
  { "channel": "wechat", "name": "微信扫码支付", "pay_method": "native" }
]
```

---

### 1.15 添加购物车

**POST** `/marketplace/shopping_cart/add`

> 鉴权：用户登录

**请求参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| user_id | uint64 | 是 | 用户 ID |
| sku_id | uint64 | 是 | 商品 ID |
| quantity | int | 是 | 数量 |

**响应 data**

购物车项信息。

---

### 1.16 移除购物车商品

**POST** `/marketplace/shopping_cart/remove`

> 鉴权：用户登录

**请求参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| user_id | uint64 | 是 | 用户 ID |
| sku_id | uint64 | 是 | 商品 ID |

**响应 data**

null

---

### 1.17 更新购物车数量

**POST** `/marketplace/shopping_cart/update`

> 鉴权：用户登录

**请求参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| user_id | uint64 | 是 | 用户 ID |
| sku_id | uint64 | 是 | 商品 ID |
| quantity | int | 是 | 新数量 |

**响应 data**

更新后的购物车项信息。

---

### 1.18 清空购物车

**POST** `/marketplace/shopping_cart/clear`

> 鉴权：用户登录

**请求参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| user_id | uint64 | 是 | 用户 ID |

**响应 data**

null

---

### 1.19 购物车列表

**GET** `/marketplace/shopping_cart/list`

> 鉴权：用户登录

**查询参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| user_id | uint64 | 是 | 用户 ID |

**响应 data**

| 字段 | 类型 | 说明 |
|------|------|------|
| items | array | 购物车商品列表（含 SKU 详情） |
| total_count | int | 商品总数 |
| total_cost | float32 | 总价（元） |

---

### 1.20 购物车下单

**POST** `/marketplace/shopping_cart/checkout`

> 鉴权：用户登录

**请求参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| user_id | uint64 | 是 | 用户 ID |
| sku_ids | array[uint64] | 是 | 要下单的商品 ID 列表 |
| pay_type | string | 否 | 支付类型，默认 money |
| remark | string | 否 | 备注 |

> 下单成功后自动从购物车移除对应商品。

**响应 data**

同「1.4 创建订单」响应。

---

## 二、运营接口（Ops）

> 所有运营接口需要管理员权限（Admin）。

### 2.1 创建商品

**POST** `/ops/skus/create`

**请求参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| biz_code | string | 是 | 业务编码 |
| sku_code | string | 是 | 商品代码（业务域内唯一） |
| sku_name | string | 是 | 商品名称 |
| cost | float32 | 是 | 商品售价（元） |
| sku_avatar | string | 否 | 商品图标 URL |
| sku_desc | string | 否 | 商品描述 |
| delivery_method | string | 否 | 履约回调接口 URL |
| multi_select | uint8 | 否 | 是否支持多选：0-否 1-是，默认 0 |

**响应 data**

创建成功的商品信息。

---

### 2.2 编辑商品

**POST** `/ops/skus/edit`

**请求参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint64 | 是 | 商品 ID |
| sku_name | string | 是 | 商品名称 |
| cost | float32 | 是 | 商品售价（元） |
| sku_avatar | string | 否 | 商品图标 URL |
| sku_desc | string | 否 | 商品描述 |
| delivery_method | string | 否 | 履约回调接口 URL |
| multi_select | uint8 | 否 | 是否支持多选 |

**响应 data**

编辑后的商品信息。

---

### 2.3 商品详情

**GET** `/ops/skus/:id`

**路径参数**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint64 | 商品 ID |

**响应 data**

商品完整信息。

---

### 2.4 商品列表

**GET** `/ops/skus`

**查询参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| sku_name | string | 否 | 商品名称（模糊查询） |
| biz_code | string | 否 | 业务编码过滤 |
| status | uint8 | 否 | 上架状态过滤 |
| offset | int | 否 | 偏移量 |
| limit | int | 否 | 每页数量 |

> 支持返回所有状态的商品，支持组合查询。

**响应 data**

| 字段 | 类型 | 说明 |
|------|------|------|
| total | int64 | 总数 |
| list | array | 商品列表 |

---

### 2.5 上架商品

**POST** `/ops/skus/listing`

**请求参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint64 | 是 | 商品 ID |

**响应 data**

null

---

### 2.6 下架商品

**POST** `/ops/skus/delisting`

**请求参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint64 | 是 | 商品 ID |

**响应 data**

null

---

### 2.7 删除商品

**POST** `/ops/skus/delete`

> 仅允许删除未上架的商品。

**请求参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint64 | 是 | 商品 ID |

**响应 data**

null

---

### 2.8 订单列表（运营）

**GET** `/ops/orders`

**查询参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| user_id | uint64 | 否 | 按用户 ID 过滤 |
| status | uint8 | 否 | 按状态过滤 |
| offset | int | 否 | 偏移量 |
| limit | int | 否 | 每页数量 |

**响应 data**

| 字段 | 类型 | 说明 |
|------|------|------|
| total | int64 | 总数 |
| list | array | 订单列表 |

---

### 2.9 订单详情（运营）

**GET** `/ops/orders/:order_no`

**路径参数**

| 字段 | 类型 | 说明 |
|------|------|------|
| order_no | string | 订单号 |

**响应 data**

订单完整信息（含明细）。

---

### 2.10 订单退款

**POST** `/ops/orders/refund`

**请求参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| order_no | string | 是 | 订单号 |
| reason | string | 否 | 退款原因 |

**响应 data**

null

---

### 2.11 手动履约

**POST** `/ops/orders/fulfill`

**请求参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| order_no | string | 是 | 订单号 |
| biz_user_id | string | 是 | 业务用户 ID |

**响应 data**

null

---

## 三、OpenAPI 接口

> 面向公司内部其他业务平台调用。

### 3.1 用户绑定

**POST** `/openapi/user/bind`

> 鉴权：免鉴权

**请求参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| biz_code | string | 是 | 业务平台代码 |
| biz_user_id | uint64 | 是 | 业务平台用户 ID |
| tel_no | string | 否* | 手机号（与 email 二选一） |
| email | string | 否* | 邮箱 |
| secret | string | 是 | 用户密钥 |

**响应 data**

| 字段 | 类型 | 说明 |
|------|------|------|
| user_id | uint | 商城用户 ID |
| is_new_user | bool | 是否新创建的用户 |
| secret_key | string | 用户密钥（仅新用户返回） |

---

### 3.2 用户解绑

**POST** `/openapi/user/unbind`

> 鉴权：用户登录

**请求参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| user_id | uint | 是 | 商城用户 ID |
| biz_code | string | 是 | 业务平台代码 |

**响应 data**

null

---

### 3.3 获取用户绑定列表

**GET** `/openapi/user/:user_id/bindings`

> 鉴权：用户登录

**路径参数**

| 字段 | 类型 | 说明 |
|------|------|------|
| user_id | uint | 用户 ID |

**响应 data**

绑定信息数组，每条包含 biz_code、biz_user_id 等。

---

### 3.4 增加积分

**POST** `/openapi/ecoin/add`

> 鉴权：用户登录

**请求参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| user_id | uint64 | 是 | 用户 ID |
| amount | float64 | 是 | 增加的积分数量 |
| source_type | string | 是 | 来源类型 |
| source_id | string | 是 | 来源业务 ID |
| description | string | 否 | 描述 |

**响应 data**

积分流水记录。

---

### 3.5 扣除积分

**POST** `/openapi/ecoin/deduct`

> 鉴权：用户登录

**请求参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| user_id | uint64 | 是 | 用户 ID |
| amount | float64 | 是 | 扣除的积分数量 |
| source_type | string | 是 | 来源类型 |
| source_id | string | 是 | 来源业务 ID |
| description | string | 否 | 描述 |

**响应 data**

积分流水记录。

---

### 3.6 初始化积分账户

**POST** `/openapi/ecoin/init`

> 鉴权：用户登录

**请求参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| user_id | uint64 | 是 | 用户 ID |

**响应 data**

积分账户信息。

---

### 3.7 查询积分余额

**GET** `/openapi/ecoin/:user_id`

> 鉴权：用户登录

**路径参数**

| 字段 | 类型 | 说明 |
|------|------|------|
| user_id | uint64 | 用户 ID |

**响应 data**

积分账户信息（同 1.10）。

---

### 3.8 查询积分流水

**GET** `/openapi/ecoin/:user_id/transactions`

> 鉴权：用户登录

**路径参数**

| 字段 | 类型 | 说明 |
|------|------|------|
| user_id | uint64 | 用户 ID |

**查询参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| offset | int | 否 | 偏移量 |
| limit | int | 否 | 每页数量 |

**响应 data**

同「1.11 积分流水列表」。

---

## 四、支付回调接口

> 免鉴权，由第三方支付平台调用。

### 4.1 微信支付回调

**POST** `/openapi/callback/wechat/pay`

> 鉴权：免鉴权
> 
> 由微信支付平台回调，非业务方调用。

**请求体**

微信支付 APIv3 标准回调通知格式（加密 JSON）。

**响应格式**

```json
{ "code": "SUCCESS", "message": "OK" }
```

---

### 4.2 微信退款回调

**POST** `/openapi/callback/wechat/refund`

> 鉴权：免鉴权

同微信标准退款回调格式。

---

### 4.3 支付宝支付回调

**POST** `/openapi/callback/alipay/pay`

> 鉴权：免鉴权

**响应**

纯文本 `success`

---

## 五、系统接口

### 5.1 健康检查

**GET** `/ping`

> 鉴权：免鉴权

**响应 data**

null（retcode 为 0 表示数据库和缓存连接正常）

---

## 六、Mock 接口（仅测试用）

### 6.1 模拟履约回调

**POST** `/mock/delivery_callback`

> 鉴权：免鉴权

**请求参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| sku_code | string | 是 | 商品代码 |
| biz_user_id | string | 是 | 业务用户 ID |

**响应格式**

```json
{ "retcode": 0, "message": "mock fulfill success" }
```

---

## 附录：SKU 履约回调协议

商品的 `delivery_method` 字段配置的履约回调 URL，由系统在订单履约时自动调用。

**请求方式**：POST

**请求体**

```json
{
  "sku_code": "商品代码",
  "biz_user_id": "业务用户ID"
}
```

**期望响应**

```json
{
  "retcode": 0,
  "message": ""
}
```

| retcode | 说明 |
|---------|------|
| 0 | 履约成功 |
| 非 0 | 履约失败，message 为错误信息 |
