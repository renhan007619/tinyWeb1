# 从0理解整个 tinyWeb1 项目

> **目标**：深度理解项目每一行代码，能给别人讲清楚，能自己改功能
> **周期**：6周（38天）
> **每日投入**：高强度日 2h / 低强度日 0.5-1h / 休息日不学项目
> **节奏**：连续4天高强度后需1天低强度或休息

---

## 个人画像

```
大一 / CS专业 / 每天2h高强度 / 4天一休息
LeetCode 60题(二叉树强+栈) / 操作系统学到进程线程
HTML/CSS/JS基础有 / Go和数据库零基础 / 目前全懵
目标：能讲清楚每行代码 / 能改功能 / 追求深度不追速度
```

---

## 强度节奏设计

```
第1周：█████░░░░ 高高高中低    （Day1-4强 Day5休）
第2周：░░█████░░ 低高高高中    （Day6低 Day7-10强 Day11休）
第3周：░░░░█████ 低低低高高中   （Day12-14低 Day15-18强 Day19休）
第4周：██████░░░ 高高高高中低    （Day20-24强 Day25休）
第5周：░░█████░░ 低高高高中     （Day26低 Day27-30强 Day31休）
第6周：█████░░░░ 高高高中低低    （Day32-35强 Day36-38低强度收尾）

高强度日 = 2h 项目学习 + 写代码
低强度日 = 0.5-1h 回顾/看视频/整理笔记
休息日 = 不学项目，可以刷LeetCode或学操作系统
```

---

## 项目全貌

```
tinyWeb1/
├── fronted/index.html          ← 前端（一个文件包含所有）
│   ├── 访问统计（visit）         ← 第3530-3646行
│   ├── 备忘录（todo）            ← 搜索 todo 关键字
│   ├── 留言板（guestbook）       ← 搜索 guestbook 关键字
│   ├── 主题切换（setting）       ← 搜索 theme 关键字
│   └── 页面UI/动画/交互          ← HTML/CSS部分
│
└── server(数据库代码)/           ← 后端
    ├── main.go                  ← 服务器入口 + 路由
    ├── config/config.go         ← 配置管理
    ├── db/db.go                 ← 数据库连接
    ├── model/model.go           ← 数据结构定义（5个struct）
    └── handler/
        ├── visit.go             ← 访问统计
        ├── todo.go              ← 备忘录CRUD
        ├── guestbook.go         ← 留言板
        ├── setting.go           ← 主题设置
        └── helpers.go           ← 公共工具函数
```

**关键点：5个handler的代码模式完全一样**

```go
// 所有 handler 都是这个模式：
func XxxHandler(w http.ResponseWriter, r *http.Request) {
    json.Decode(r.Body)     // 解析请求
    database.Where()        // 查库
    database.Create/Update  // 写库
    sendJSON(w, ...)        // 返回响应
}
```

把访问统计吃透 = 懂了70%的项目。剩下的就是举一反三。

---

# 第1周：前端代码精读（以访问统计为切入点）

## 周目标

> 能完整口述 `index.html` 第 3530-3646 行的每一行在干什么，为什么这么写。

---

### Day 1（高强度）— IIFE 和函数定义

**学什么：** `index.html` 第 3530-3572 行，所有函数定义部分

**时间分配：**

| 时间 | 任务 | 验收标准 |
|------|------|---------|
| 前30min | 读第 3530-3538 行 | 能说出 `(function(){` 是什么、`'use strict'` 干嘛的 |
| 中间60min | 读第 3534-3565 行，4个检测函数 | 每个函数输入是什么、输出是什么、正则怎么匹配的 |
| 后30min | 读第 3568-3572 行 animateNumber | DOM操作classList的add/remove是干嘛的 |

**今天要搞懂的5个概念：**

```javascript
// 1. IIFE — 第3530行
(function() {
    // ...
})();          // ← 为什么最后要加()？

// 2. var vs let vs const
var API_BASE = '';      // 你代码用的var，和let有什么区别？

// 3. function 声明 vs 函数表达式
function getVisitorIP() { ... }   // 第3534行，函数声明
var detectDeviceType = function() { ... } // 如果写成这样呢？

// 4. navigator 对象 — 浏览器内置API
navigator.userAgent   // 浏览器自己提供的，不是你写的
document.referrer     // 同上

// 5. 正则表达式 — 第3542行
/Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i
.test(navigator.userAgent)   // .test() 返回 true/false
```

**动手练习（必须写）：**

```javascript
// 练习1：写一个自己的IIFE，打印"hello"
(function(){
    console.log("hello");
})();

// 练习2：模仿 detectOS，写一个 detectBrowser 的简化版本
function myDetectBrowser(ua) {
    if (ua.includes('Chrome')) return 'Chrome';
    if (ua.includes('Firefox')) return 'Firefox';
    return 'Other';
}
console.log(myDetectBrowser(navigator.userAgent));  // 看看输出什么

// 练习3：用正则判断一个字符串是不是手机号
function isPhone(str) {
    return /^1[3-9]\d{9}$/.test(str);
}
console.log(isPhone("13812345678"));  // true还是false？
```

**Day 1 过关标准：**
- [ ] 能解释 IIFE 为什么定义完立刻执行
- [ ] 4个检测函数每个都能说清楚输入输出
- [ ] 上面的3个练习都跑通了

---

### Day 2（高强度）— 核心业务逻辑

**学什么：** `index.html` 第 3575-3646 行，renderStats + fetch + sessionStorage

**时间分配：**

| 时间 | 任务 |
|------|------|
| 前30min | 读第 3575-3591 行 renderStats 函数，逐行注释 |
| 中间60min | 读第 3593-3646 行，核心逻辑，画流程图 |
| 后30min | 打开浏览器 F12 → Console，手动执行代码验证 |

**今天要搞懂的5个概念：**

```javascript
// 1. DOM 操作 — 第3576行
document.getElementById('totalVisits')
// 这返回的是什么？一个HTML元素对象
// .textContent = 27  把这个元素的文字改成 "27"

// 2. sessionStorage — 第3596行
sessionStorage.getItem('visit_recorded')   // 取值，没有返回null
sessionStorage.setItem('visit_recorded', 'true')  // 存值

// 3. fetch API — 第3600行
fetch(url, options).then(fn1).then(fn2).catch(fn3)
//                         ↑        ↑       ↑
//                      成功回调  再成功   出错了

// 4. Promise / .then() — 异步
// fetch 发请求需要时间，不会卡住后面的代码
// .then() 表示"等结果回来了再执行这个函数"

// 5. JSON 序列化 — 第3603行
JSON.stringify({a:1})   →  '{"a":1}'     对象→字符串
JSON.parse('{"a":1}')   →  {a:1}         字符串→对象
```

**重点理解第 3596-3645 行的 if/else 分支：**

```
sessionStorage 有标记？
├─ 没有（首次打开）→ 存标记 → POST 记录访问 → GET 获取统计 → 显示
└─ 有（刷新页面）    → 跳过POST → 直接 GET 获取统计 → 显示
```

**动手练习（必须写）：**

```javascript
// 练习1：在浏览器 F12 Console 里执行
sessionStorage.getItem('visit_recorded')   // 应该返回 null 或 "true"
sessionStorage.setItem('test', 'hello')
sessionStorage.getItem('test')              // 应该返回 "hello"
sessionStorage.removeItem('test')           // 清除

// 练习2：手动发一个 GET 请求看返回什么
fetch('/api/visit/stats')
    .then(r => r.json())
    .then(data => console.log(data))
    .catch(err => console.error(err));
// 在F12 Console里粘贴执行，看看返回了什么

// 练习3：写一个简单的 fetch POST
fetch('/api/visit', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({visitor_ip:'', browser:'Test', os:'Linux'})
})
.then(r => r.json())
.then(data => console.log('服务器返回:', data));
```

**Day 2 过关标准：**
- [ ] 能画出 if/else 的分支流程图
- [ ] 能说清楚 fetch 的参数各是什么意思
- [ ] 3个练习都在 Console 里跑通看到结果

---

### Day 3（高强度）— HTTP 协议实战

**学什么：** 用你自己的项目当教材，看真实的 HTTP 报文

**时间分配：**

| 时间 | 任务 |
|------|------|
| 前30min | B站搜"HTTP协议详解"，看一个30min以内的视频 |
| 中间60min | 打开网站 → F12 Network → 刷新 → 逐个分析请求 |
| 后30min | 整理笔记 |

**实操任务：**

打开 `http://1.15.224.88:8080`，按 F12 → 点 Network 标签 → 刷新页面。找到这几个请求：

| 请求名 | 方法 | 它是干嘛的 |
|--------|------|-----------|
| `visit` | POST | 记录一次访问（第3600行的fetch发的） |
| `visit/stats` | GET | 获取统计数据（第3617行的fetch发的） |
| `index.html` | GET | 加载页面本身 |
| `background3.webp` | GET | 加载背景图片 |

**对 `visit` 这个请求，点开它，看这几个标签：**

```
Headers 标签：
├── Request URL: http://1.15.224.88:8080/api/visit
├── Request Method: POST                    ← 方法
├── Status Code: 200 OK                     ← 成功
├── Remote Address: 1.15.224.88:8080        ← 服务器地址
│
├── Request Headers（浏览器自动加的）：       ← 请求头
│   Accept: */*
│   Content-Type: application/json          ← 你代码里headers写的
│   User-Agent: Mozilla/5.0 ...             ← 浏览器自动带的
│
└── Response Headers（服务器返回的）：        ← 响应头
    Content-Type: application/json

Payload 标签：
{"visitor_ip":"","user_agent":"Mozilla...","browser":"Chrome","os":"Windows"...}
                                          ↑ 这就是你第3603行的body！

Response 标签：
{"code":0,"message":"success","data":{"is_first_visit":true,"visit_count":1}}
                                           ↑ 这是后端返回的数据！
```

**动手练习（必须做）：**

1. 在 F12 Network 里找到 `visit` 请求，截图保存
2. 在 F12 Network 里找到 `visit/stats` 请求，对比两者的区别（方法？有没有Payload？）
3. 手动在浏览器地址栏输入 `http://1.15.224.88:8080/api/visit/stats` 回车，看返回什么

**Day 3 过关标准：**
- [ ] 能说清楚 POST 和 GET 在 Network 面板里的视觉区别
- [ ] 能指出 Payload 里哪个字段对应代码哪一行
- [ ] 能说清 Request Headers 和 Response Headers 分别是谁发给谁的

---

### Day 4（高强度）— 前端全流程贯通

**学什么：** 把前3天学的串起来，形成完整认知

**时间分配：**

| 时间 | 任务 |
|------|------|
| 前40min | 从头到尾再读一遍 3530-3646 行，这次带着理解读 |
| 中间50min | 画一张完整的流程图（纸笔画或用工具） |
| 后30min | 给每一行代码写中文注释 |

**你要画的流程图（至少包含这些节点）：**

```
用户打开网页
    ↓
浏览器加载 index.html，解析到 <script>
    ↓
IIFE 开始执行 (第3530行)
    ↓
检查 sessionStorage (第3596行)
    ├── 没有 → 存标记 → fetch POST (第3600行)
    │                ↓
    │           浏览器发送HTTP请求
    │                ↓
    │           服务器接收处理
    │                ↓
    │           收到响应 → fetch GET stats (第3617行)
    │                ↓
    │           renderStats() 显示数据 (第3621行)
    │
    └── 有 → 跳过POST → fetch GET stats (第3640行)
                      ↓
                 renderStats() 显示数据 (第3643行)
```

**动手练习（必须写）：**

给第 3530-3646 行每一行加上中文注释。示例：

```javascript
// 第3530行：开始定义一个匿名函数并立即执行（IIFE模式）
(function() {
    // 第3531行：API基地址，同源部署时为空字符串
    var API_BASE = '';
    
    // 第3533-3538行：获取访客IP的函数
    // 前端无法直接获取真实IP，所以返回空字符串让后端从TCP连接取
    function getVisitorIP() {
        return '';  // 故意返回空
    }
    
    // （... 每一行都这样写 ...）
    
// 第3646行：IIFE结束，括号表示立即调用上面的函数
})();
```

**Day 4 过关标准：**
- [ ] 流程图覆盖了从页面加载到数据显示的全过程
- [ ] 每一行代码都有注释
- [ ] 合上文件，能凭记忆说出大致流程

---

### Day 5（高强度）— 前端总结 + 输出

**时间分配：**

| 时间 | 任务 |
|------|------|
| 前60min | 写掘金文章初稿 |
| 后60min | 文章里的每个技术点都要能在你的代码里定位到行号 |

**文章结构（边写边回看代码验证）：**

```
标题：《我的博客一刷新访问量就+1？排查与修复》

1. 问题描述
   → 截图：F12 Network 看到 visit 请求每次刷新都发
   
2. 排查过程
   → 代码第3530行 IIFE 是什么？
   → 代码第3600行 fetch 每次都执行
   
3. 发现原因
   → IIFE 页面加载就执行，刷新也执行
   
4. 解决方案
   → 代码第3596行 sessionStorage 检查
   → 首次存标记，后续跳过
   
5. 知识扩展
   → IIFE 是什么（配图）
   → sessionStorage vs localStorage（配表格）
   
6. 效果对比
   → 修改前后 F12 截图对比
```

**Day 5 过关标准：**
- [ ] 文章初稿完成（可以在掘金草稿箱保存）
- [ ] 文章里引用的每个代码片段都能定位到你项目中的行号
- [ ] 自己读一遍文章，觉得逻辑通顺

---

### Day 6（低强度休息日）

**不做新内容。可选：**

- [ ] 刷 1-2 道 LeetCode（保持手感）
- [ ] 学操作系统（进程线程相关）
- [ ] 回顾前5天的笔记
- [ ] 整理 F12 截图

---

# 第2周：Go 语言入门 + 后端代码精读

## 周目标

> 能读懂 `handler/visit.go` 的每一个函数，知道 Go 的基本语法。

---

### Day 7（低强度）— Go 环境搭建 + Hello World

**你有C语言基础，Go 会很快。先搭环境。**

| 时间 | 任务 |
|------|------|
| 30min | 安装 Go（go.dev 下载，一路下一步） |
| 30min | VS Code 装 Go 插件 |
| — | 验证：终端输入 `go version` 能显示版本号 |

**第一个 Go 程序：**

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, tinyWeb1!")
}
```

终端运行：

```bash
go run hello.go
```

**Day 7 过关标准：**
- [ ] `go run hello.go` 能输出 Hello
- [ ] VS Code 里 Go 插件装好了（有语法高亮和提示）

---

### Day 8（高强度）— Go 基础速通（对标C语言）

**你有C基础，只需要学差异点。**

**今天要掌握的（对照C来学）：**

```go
// 1. 变量声明 — C是 int a = 1; Go有多种写法
var a int = 1           // 类似C，完整写法
b := 2                  // 短声明，自动推断类型（C没有）
const PI = 3.14          // 和C一样

// 2. 函数 — C的升级版，可以多返回值
func add(a int, b int) int {    // 和C类似
    return a + b
}

func divide(a, b int) (int, error) {  // 多返回值（C没有！）
    if b == 0 {
        return 0, fmt.Errorf("不能除0")
    }
    return a / b, nil
}

// 3. struct — 和C几乎一样
type Person struct {
    Name string    // 注意：类型写在后面
    Age  int
}
p := Person{Name: "张三", Age: 20}  // 字面量初始化

// 4. 方法 — Go特有，函数前面加receiver
func (p Person) SayHello() {    // (p Person) 就是 receiver
    fmt.Println("我是", p.Name)
}
p.SayHello()   // 调用

// 5. interface{} — 空接口，可以存任何类型
var x interface{} = 42
x = "hello"       // 可以重新赋值为不同类型

// 6. error 处理 — Go特色，if err != nil 到处都是
result, err := divide(10, 3)
if err != nil {            // Go的标准错误处理模式
    fmt.Println("出错:", err)
    return                 // 出错就返回
}
fmt.Println(result)        // 没问题才继续

// 7. slice — 类似动态数组（C里要手写）
nums := []int{1, 2, 3}    // 切片
nums = append(nums, 4)     // 追加元素

// 8. map — 哈希表（你LeetCode用过）
m := map[string]int{
    "张三": 90,
    "李四": 85,
}
score := m["张三"]  // 90

// 9. 指针 — 和C一样
a := 10
p := &a     // 取地址
*p = 20      // 解引用，a变成20
```

**动手练习（必须写）：**

```go
// 练习1：写一个 VisitStats 结构体（仿照你model.go里的）
type VisitStats struct {
    IP   string
    Count int
}

v := VisitStats{IP: "127.0.0.1", Count: 5}
fmt.Println(v.IP)   // 输出什么？

// 练习2：写一个带 error 返回值的函数
func GetIP(s string) (string, error) {
    if s == "" {
        return "", fmt.Errorf("IP为空")
    }
    return s, nil
}

ip, err := GetIP("")
if err != nil {
    fmt.Println("错误:", err)
} else {
    fmt.Println("IP:", ip)
}

// 练习3：用 map 统计字符出现次数
func countChars(s string) map[rune]int {
    m := make(map[rune]int)
    for _, c := range s {
        m[c]++
    }
    return m
}
fmt.Println(countChars("hello"))  // 输出什么？
```

**Day 8 过关标准：**
- [ ] 上面9个知识点每个都能写出示例代码
- [ ] 3个练习都跑通
- [ ] 能说出 Go 和 C 在变量声明、函数、error 处理上的区别

---

### Day 9（高强度）— 读 model.go + config.go

**这两个文件最简单，适合作为读Go代码的起点。**

**读 model.go，逐个 struct 分析：**

| 行号 | 内容 | 要搞懂的 |
|------|------|---------|
| 65-85 | `VisitStats struct` | 每个字段含义、gorm tag、json tag |
| 100-107 | `VisitRecord struct` | 请求体结构体，对应前端传来的 JSON |
| 111-115 | `VisitStatsResponse struct` | 响应体结构体 |

**重点搞懂 tag：**

```go
VisitorIP string `gorm:"type:varchar(45);uniqueIndex;not null" json:"visitor_ip"`
//          ↑                                                  ↑
//     gorm tag：告诉GORM数据库里这列怎么建              json tag：JSON序列化时用什么key名
//
// gorm tag 拆解：
//   type:varchar(45)  → 数据库列类型，最长45字符（IPv6最长45位）
//   uniqueIndex      → 建唯一索引，同一个IP只能有一条记录
//   not null         → 不允许为空
//
// json tag 拆解：
//   json:"visitor_ip" → JSON里显示为 "visitor_ip"
//                        如果不写，默认用字段名 VisitorIP（大写开头）
```

**读 config.go：**

```go
type DBConfig struct {
    Host string `json:"-"`   // json:"-" 表示不参与JSON序列化
    Port string
    User string
    Pass string
    Name string
}

type AppConfig struct {
    AppEnv      string
    MainDB      DBConfig    // 嵌套结构体
    TestDB      DBConfig
    ServerPort  string      // ":8081"
    StaticDir   string
}
```

**动手练习（必须写）：**

```go
// 练习1：自己定义一个 struct，加 gorm tag 和 json tag
type Student struct {
    ID   int    `gorm:"primarykey" json:"id"`
    Name string `gorm:"type:varchar(20)" json:"name"`
    Age  int    `json:"age"`
}

// 练习2：JSON序列化看看效果
import "encoding/json"

s := Student{ID: 1, Name: "任海", Age: 19}
data, _ := json.Marshal(s)
fmt.Println(string(data))
// 输出什么？注意 json tag 怎么影响输出的key名

// 练习3：模拟 config.go 的环境变量读取
func getEnv(key, defaultValue string) string {
    if value := "fake_env"; value != "" {  // 模拟 os.Getenv
        return value
    }
    return defaultValue
}
port := getEnv("SERVER_PORT", ":8081")
fmt.Println(port)  // 输出什么？
```

**Day 9 过关标准：**
- [ ] VisitStats 的每个字段都能说出用途和对应的 gorm tag 含义
- [ ] 能解释 `json:"visitor_ip"` 和 `gorm:"uniqueIndex"` 各自的作用
- [ ] 3个练习跑通

---

### Day 10（高强度）— 读 main.go（服务器入口）

**这是整个后端的骨架。按执行顺序读：**

```
main() 第50行
    │
    ├─ 第54行 config.Load()
    │   └─ 读环境变量 → 生成 AppConfig 全局实例 appConfig
    │
    ├─ 第60行 db.Initialize()
    │   └─ 连接 MySQL 主库 → AutoMigrate 建表
    │
    ├─ 第65行 db.InitializeTestDB()
    │   └─ 连接 MySQL 测试库 → AutoMigrate
    │
    ├─ 第70行 testVisitStats()
    │   └─ 启动时测试一下数据库能不能正常读写
    │
    └─ 第75行 startServer()    ★ 重点！
        ├─ 第189行 rootDir = config.GetStaticDir()
        │   └─ 决定静态文件的根目录（你的 index.html 就在这里）
        │
        ├─ 第199行 mux := http.NewServeMux()
        │   └─ 创建路由器（空的映射表）
        │
        ├─ 第203行 mux.HandleFunc("/api/health", ...)
        │   └─ 注册路由1：健康检查
        │
        ├─ 第206行 mux.HandleFunc("/api/visit", ...)
        │   ├─ POST → handler.RecordVisit(w,r)
        │   └─ 其他方法 → 405 错误
        │
        ├─ 第213行 mux.HandleFunc("/api/visit/stats", ...)
        │   └─ GET → handler.GetVisitStats(w,r)
        │
        ├─ 第223行 mux.Handle("/", fs)
        │   └─ 兜底路由：其他路径返回静态文件（index.html等）
        │
        ├─ 第238行 http.ListenAndServe(addr, corsMiddleware(mux))
        │   └─ 启动监听，程序阻塞在这里等待请求
        │
        └─ 第308行 corsMiddleware(next Handler) Handler
            └─ CORS中间件：在每个请求前后加跨域头
```

**重点理解三个概念：**

```go
// 1. ServeMux — Go内置的路由器
mux := http.NewServeMux()
mux.HandleFunc("/api/visit", myHandler)  // 注册映射关系
// 本质就是个 map[string]Handler

// 2. http.HandlerFunc — 函数类型
// 只要是一个 func(w ResponseWriter, r *Request) 的函数
// 就可以被注册成路由处理器

// 3. ListenAndServe — 启动服务器
http.ListenAndServe(":8081", mux)
// 监听 8081 端口，收到请求后交给 mux 匹配路由
// 程序在这里阻塞，不会退出
```

**动手练习（必须写）：**

```go
// 练习1：写一个最简HTTP服务器（不用框架）
package main

import "net/http"

func helloHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello from tinyWeb1!"))
}

func main() {
    // 注册路由
    http.HandleFunc("/hello", helloHandler)
    
    // 启动服务器
    println("服务器启动在 :9999")
    http.ListenAndServe(":9999", nil)
}
// go run 之后浏览器访问 localhost:9999/hello

// 练习2：加一个 JSON 返回的接口
func apiHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(`{"code":0,"message":"success","data":{"name":"tinyWeb1"}}`))
}

// 练习3：模拟路由匹配
mux := http.NewServeMux()
mux.HandleFunc("/a", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("这是A"))
})
mux.HandleFunc("/b", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("这是B"))
})
// 访问 /a 和 /b 分别返回什么？
```

**Day 10 过关标准：**
- [ ] 能画出 main() 的调用链
- [ ] 能说清楚 HandleFunc 注册了什么、ListenAndServe 干嘛的
- [ ] 练习1的服务器跑起来并在浏览器访问

---

### Day 11（高强度）— 读 handler/visit.go（核心！）

**这是最重要的文件。前端和数据库在这里交汇。**

#### 函数 1：getClientIP（第39-54行）

```go
func getClientIP(r *http.Request) string {
    // 参数 r *http.Request = 浏览器发来的请求对象
    // * 号是指针，类似C语言的指针
    
    // 策略1：X-Forwarded-For 头
    // 当请求经过 Nginx/CDN 时，真实IP会放在这里
    if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
        parts := strings.Split(xff, ",")   // 可能有多个IP，取第一个
        return strings.TrimSpace(parts[0])
    }
    
    // 策略2：X-Real-IP 头
    // Nginx 通常会设置这个
    if xri := r.Header.Get("X-Real-IP"); xri != "" {
        return strings.TrimSpace(xri)
    }
    
    // 策略3：RemoteAddr（最终兜底）
    addr := r.RemoteAddr   // 类似 "106.86.8.116:52341"
    if lastColon := strings.LastIndex(addr, ":"); lastColon != -1 {
        return addr[:lastColon]   // 去掉端口 → "106.86.8.116"
    }
    return addr
}
```

#### 函数 2：RecordVisit（第81-162行）— 最核心

```go
func RecordVisit(w http.ResponseWriter, r *http.Request) {
    // w = 往回写给浏览器的笔
    // r = 浏览器发来的请求
    
    // 步骤1：解析 body JSON（第82-99行）
    var req model.VisitRecord
    json.NewDecoder(r.Body).Decode(&req)
    // r.Body = '{"visitor_ip":"","browser":"Chrome",...}'
    // Decode之后 req 就是个 Go struct
    
    // 步骤2：补全 IP（第91-99行）
    if req.VisitorIP == "" {
        req.VisitorIP = getClientIP(r)  // 调用上面的函数
    }
    
    // 步骤3：查数据库（第101-150行）
    database.Where("visitor_ip = ?", req.VisitorIP).First(&existing)
    // SQL: SELECT * FROM visit_stats WHERE visitor_ip='xxx' LIMIT 1
    
    if result.Error == gorm.ErrRecordNotFound {
        // 没找到 → INSERT（第111-128行）
        newRecord := model.VisitStats{
            VisitorIP:    req.VisitorIP,
            VisitCount:   1,
            FirstVisitAt: now,
            LastVisitAt:  now,
            Browser:      req.Browser,
            OS:           req.OS,
        }
        database.Create(&newRecord)
        // SQL: INSERT INTO visit_stats VALUES(...)
        
    } else {
        // 找到了 → UPDATE（第134-149行）
        database.Model(&existing).Updates(map[string]interface{}{
            "visit_count":  gorm.Expr("visit_count + 1"),
            "last_visit_at": now,
        })
        // SQL: UPDATE visit_stats SET visit_count=visit_count+1 WHERE id=?
    }
    
    // 步骤4：返回响应（第153-161行）
    sendJSON(w, http.StatusOK, model.SuccessResponse(responseData))
    // 通过 w 写回 JSON 给前端
}
```

#### 函数 3：GetVisitStats（第186-206行）

```go
func GetVisitStats(w http.ResponseWriter, r *http.Request) {
    // 查询1：总访问次数
    database.Model(&model.VisitStats{}).
        Select("COALESCE(SUM(visit_count), 0)").Scan(&stats.TotalVisits)
    // SQL: SELECT COALESCE(SUM(visit_count),0) FROM visit_stats
    
    // 查询2：独立访客数
    database.Model(&model.VisitStats{}).
        Select("COUNT(*)").Scan(&stats.UniqueVisitors)
    // SQL: SELECT COUNT(*) FROM visit_stats
    
    // 查询3：最后访问时间
    database.Model(&model.VisitStats{}).
        Select("MAX(last_visit_at)").Scan(&lastVisit)
    // SQL: SELECT MAX(last_visit_at) FROM visit_stats
    
    sendJSON(w, http.StatusOK, model.SuccessResponse(stats))
}
```

**动手练习（必须写）：**

```go
// 练习1：模拟 RecordVisit 的核心逻辑（不用数据库，用map代替）
var db = make(map[int]string)  // 模拟数据库

func record(id int, name string) string {
    for k, v := range db {
        if v == name {
            db[k] = name + "(更新)"
            return "更新成功: " + name
        }
    }
    lenDB := len(db) + 1
    db[lenDB] = name
    return "新增成功: " + name
}

fmt.Println(record(0, "Alice"))  // 新增
fmt.Println(record(0, "Bob"))    // 新增
fmt.Println(record(0, "Alice"))  // 更新？

// 练习2：把 handler/visit.go 的每个函数抄写一遍（手敲！）
// 边抄边想每一行在干什么
```

**Day 11 过关标准：**
- [ ] getClientIP 的三种策略能口述清楚
- [ ] RecordVisit 的 4 个步骤能默写出来
- [ ] 每个 GORM 操作能写出对应的 SQL
- [ ] 练习1跑通

---

### Day 12（高强度）— helpers.go + db.go

**辅助文件，快速过。**

**helpers.go：**

```go
// sendJSON — 统一的 JSON 响应函数
func sendJSON(w http.ResponseWriter, statusCode int, response model.APIResponse) {
    w.Header().Set("Content-Type", "application/json;charset=utf-8")  // 设置响应头
    w.WriteHeader(statusCode)   // 设置状态码 200/400/500 等
    json.NewEncoder(w).Encode(response)   // 把 Go struct 编码成 JSON 写入 w
}

// trimString — 去空格
func trimString(s string) string {
    return strings.TrimSpace(s)
}
```

**db.go（只看关键部分）：**

```go
// 连接池配置
sqlDB.SetMaxOpenConns(10)        // 最大同时10个连接
sqlDB.SetMaxIdleConns(5)         // 空闲保留5个连接
sqlDB.SetConnMaxLifetime(30*time.Minute)  // 连接最多活30分钟

// 为什么需要连接池？
// 每次查询都建立TCP连接太慢（三次握手）
// 提前建好一批连接放那里，用完归还，复用
```

**Day 12 过关标准：**
- [ ] 能说清 sendJSON 的三个步骤（Header → WriteHeader → Encode）
- [ ] 能解释连接池为什么存在

---

### Day 13（高强度）— 前后端联调走查

**目标：从前端第一行代码开始，跟踪到数据库最后一行SQL。**

**任务：拿一张大白纸，画完整的时序图：**

```
时间轴 →

浏览器                              Go服务器                          MySQL
  │                                   │                                │
  │ ① fetch POST /api/visit           │                                │
  │ body: {visitor_ip:"",...}         │                                │
  │==================================▶│                                │
  │                                   │ ② json.Decode 解析body        │
  │                                   │ ③ getClientIP(r) 取IP         │
  │                                   │ ④ SELECT WHERE ip='...'       │
  │                                   │===============================▶│
  │                                   │                           ⑤ 返回查询结果
  │                                   │◀==============================│
  │                                   │                                │
  │                                   │ ⑥ INSERT 或 UPDATE             │
  │                                   │===============================▶│
  │                                   │                           ⑦ 写入成功
  │                                   │◀==============================│
  │                                   │                                │
  │                                   │ ⑧ sendJSON 返回响应            │
  │◀==================================│                                │
  │ ⑨ 收到响应 {.then(res => ...)}   │                                │
  │                                    │                                │
  │ ⑩ fetch GET /api/visit/stats     │                                │
  │===================================▶│                               │
  │                                   │ ⑪ SELECT SUM/COUNT/MAX        │
  │                                   │===============================▶│
  │                                   │                          ⑫ 返回统计
  │                                   │◀==============================│
  │                                   │ ⑬ sendJSON 返回               │
  │◀==================================│                               │
  │ ⑭ renderStats() 更新DOM          │                               │
  │  用户看到 27 / 6                  │                               │
```

**每个箭头标注：对应代码哪一行。**

**Day 13 过关标准：**
- [ ] 时序图覆盖了全部 14 个步骤
- [ ] 每个步骤都能定位到具体的文件和行号
- [ ] 合上资料能口述整个过程

---

### Day 14（高强度）— 第2周总结 + 补漏

**任务：**

| 时间 | 任务 |
|------|------|
| 前60min | 回顾这一周学的所有内容，标出还不清楚的点 |
| 后60min | 针对不清楚的点，回去重读代码或查文档 |

**自查清单：**

```
□ Go 变量声明（var / := / const）懂了吗？
□ Go 函数多返回值会用了吗？
□ Go error 处理模式熟悉了吗？
□ struct 和 tag 理解了吗？
□ ServeMux 路由机制明白了吗？
□ RecordVisit 四步流程能默写吗？
□ 每个 GORM 操作对应哪条 SQL？
□ 前端 fetch 到后端 handler 到 MySQL 全程能串起来吗？
```

**有任何一项打 ❌ 的，今天就专门补这个。**

---

### Day 15（低强度休息日）

- [ ] 学操作系统
- [ ] 刷 1-2 题 LeetCode
- [ ] 整理本周笔记

---

# 第3周：数据库 + 其他模块

## 周目标

> 理解 MySQL 数据库操作，GORM 与 SQL 的对应关系，快速看懂其他模块。

---

### Day 16（高强度）— SSH 进 MySQL 实操

SSH 到服务器：

```bash
ssh user@1.15.224.88
mysql -u root tinyweb1
```

然后逐条执行，每条都要理解：

```sql
-- 1. 看有哪些表
SHOW TABLES;

-- 2. 看 visit_stats 表结构
DESCRIBE visit_stats;
-- 对照 model.go 第65-85行，一一对应

-- 3. 看所有数据
SELECT * FROM visit_stats;

-- 4. 条件查询
SELECT visitor_ip, visit_count FROM visit_stats WHERE visit_count > 5;

-- 5. 聚合函数
SELECT SUM(visit_count) as total_visits FROM visit_stats;
SELECT COUNT(*) as unique_visitors FROM visit_stats;
SELECT MAX(last_visit_at) as last_visit FROM visit_stats;
-- 这三个就是你 GetVisitStats 函数执行的SQL

-- 6. 插入一条测试数据（然后删掉）
INSERT INTO visit_stats (visitor_ip, visit_count, browser, os, first_visit_at, last_visit_at)
VALUES ('127.0.0.1', 1, 'Test', 'Linux', NOW(), NOW());

-- 7. 确认插入成功
SELECT * FROM visit_stats WHERE visitor_ip = '127.0.0.1';

-- 8. 更新这条测试数据
UPDATE visit_stats SET visit_count = visit_count + 1 WHERE visitor_ip = '127.0.0.1';

-- 9. 删除测试数据
DELETE FROM visit_stats WHERE visitor_ip = '127.0.0.1';

-- 10. 理解索引
SHOW INDEX FROM visit_stats;
-- 看 unique_index 在哪列上（应该是 visitor_ip）
```

**Day 16 过关标准：**
- [ ] 上面10条SQL全部在服务器上执行过
- [ ] 每条SQL的执行结果都理解了

---

### Day 17（高强度）— GORM 与 SQL 对应

把 Day 16 每条 SQL 和 Go 代码对应上：

```go
database.Create(&newRecord)           ↔ INSERT INTO
database.Where().First(&existing)    ↔ SELECT WHERE LIMIT 1
database.Model().Updates(...)         ↔ UPDATE SET
database.Delete(&record)             ↔ DELETE WHERE
database.Select("SUM(...)").Scan(&x)  ↔ SELECT SUM(...)
database.Select("COUNT(*)").Scan(&x)  ↔ SELECT COUNT(*)
database.AutoMigrate(&Model{})        ↔ CREATE TABLE IF NOT EXISTS
```

**动手练习（必须写）：**

```go
// 用 GORM 写出以下 SQL 对应的 Go 代码：
// 1. SELECT * FROM visit_stats WHERE browser = 'Chrome';
// 2. UPDATE visit_stats SET visit_count = 10 WHERE visitor_ip = '127.0.0.1';
// 3. DELETE FROM visit_stats WHERE id = 1;
```

**Day 17 过关标准：**
- [ ] 每个 GORM 方法都能说出对应的 SQL
- [ ] 练习跑通

---

### Day 18（高强度）— 读其他 handler

快速过 `todo.go`、`guestbook.go`、`setting.go`。
你会发现模式和 `visit.go` 完全一样：
1. json.Decode 解析请求
2. 数据库 CRUD 操作
3. sendJSON 返回响应

**每个文件花30min，重点关注：**

| 文件 | 核心函数 | 对应表 |
|------|---------|--------|
| todo.go | CreateTodo / GetTodos / UpdateTodo / DeleteTodo | todos |
| guestbook.go | CreateMessage / GetMessages | guestbook |
| setting.go | GetSettings / UpdateSettings | settings |

**Day 18 过关标准：**
- [ ] 能说出每个 handler 的 CRUD 函数名
- [ ] 能指出和 visit.go 的模式差异（如果有的话）

---

### Day 19（高强度）— 数据库设计分析 + 动手改功能

**思考几个问题：**

```
□ 为什么要用 visitor_ip 做 uniqueIndex？
  → 同一IP只存一条，多次访问更新count

□ DeletedAt 字段是干嘛的？
  → 软删除，GORM自带，DELETE时不真删而是标记时间

□ 如果要加 city 字段，需要改哪些文件？
  → model.go 加字段、handler/visit.go 写入 city、前端传 city 数据
```

**动手改功能（选做一个）：**

**选项 A（最简单）：给访问统计加城市字段**

需要改的文件：
1. `model.go` — VisitStats 加 `City string` 字段
2. `handler/visit.go` — RecordVisit 里写入 city
3. `index.html` — fetch body 里加 city 数据
4. 部署到服务器验证

**选项 B（中等）：给留言板加分页数字显示**

**选项 C（中等）：访问统计加一个"今日访问数"字段**

**Day 19 过关标准：**
- [ ] 选做的功能改动完成并部署验证
- [ ] 能说出改了哪些文件、每处改动的原因

---

### Day 20（低强度休息日）

- [ ] 操作系统学习
- [ ] 整理笔记

---

# 第4周：复习 + 深入 + 安全分析

## 周目标

> 第二轮精读代码，理解安全和性能问题，加深理解。

---

### Day 21-22（高强度）— 第二轮精读 main.go + visit.go

重新读 main.go、visit.go、index.html。
这次读应该比第一遍快很多。

**标注：**
- 第一次读不懂但现在懂了的地方 → 绿色标记
- 还是不懂的地方 → 红色标记，重点攻克

---

### Day 23（高强度）— 安全和性能分析

带着审视的眼光重新看你的代码：

**安全问题：**

```
□ SQL注入？  → GORM参数化查询，安全 ✅
□ XSS？      → 前端直接渲染数据，没转义 ⚠️
□ CSRF？     → 没有token保护 ⚠️
□ 密码明文？ → root密码为空 ⚠️
□ 限流？     → 同一IP可以无限刷 visit ⚠️
```

**性能问题：**

```
□ 每次访问都查一次DB？ → 是的，可以优化用缓存 ⚠️
□ 连接池够用吗？       → MaxOpenConns=10，小项目够了 ✅
□ 前端并发请求？       → POST完了再GET，可以优化并行 ⚠️
```

**Day 23 过关标准：**
- [ ] 能说出3个以上的安全隐患和对应的修复思路
- [ ] 能说出2个以上的性能优化方向

---

### Day 24（高强度）— 读 index.html 其他模块

之前只读了访问统计（3530-3646行），现在读备忘录和留言板的前端代码：

```
□ 搜索 index.html 里的 todo 相关 JS → 对应 todo.go
□ 搜索 index.html 里的 guestbook 相关 JS → 对应 guestbook.go
□ 找到前端 fetch 调用的 URL 和后端路由的对应关系
```

**Day 24 过关标准：**
- [ ] 能说出备忘录和留言板前端代码的位置
- [ ] 能找到每个 fetch 调用对应的后端 handler

---

### Day 25（低强度休息日）

- [ ] 学操作系统
- [ ] 刷 LeetCode
- [ ] 整理笔记

---

# 第5周：全链路贯通 + 架构理解

## 周目标

> 能不看代码画出完整架构图，能给别人讲清楚整个项目。

---

### Day 26（低强度）— 回顾前4周

重新过一遍所有笔记，标出：
- 已完全理解的知识点
- 还模糊的知识点
- 完全不懂的知识点

---

### Day 27-28（高强度）— 画完整架构图

**要求：不用看代码，凭记忆画。画完再对照代码修正。**

架构图应包含：

```
1. 前端层
   ├── HTML 结构（页面布局）
   ├── CSS 样式（主题、动画、响应式）
   └── JS 逻辑
       ├── 访问统计模块（IIFE + fetch + sessionStorage）
       ├── 备忘录模块
       ├── 留言板模块
       └── 主题设置模块

2. 后端层
   ├── main.go（入口 + 路由注册）
   ├── config.go（配置管理）
   ├── db.go（数据库连接 + 连接池）
   ├── model.go（5个struct定义）
   └── handler/
       ├── visit.go（访问统计）
       ├── todo.go（备忘录CRUD）
       ├── guestbook.go（留言板）
       ├── setting.go（设置）
       └── helpers.go（sendJSON等工具函数）

3. 数据库层
   ├── visit_stats 表
   ├── todos 表
   ├── guestbook 表
   └── settings 表

4. 前后端连接
   ├── POST /api/visit → RecordVisit
   ├── GET /api/visit/stats → GetVisitStats
   ├── CRUD /api/todos → TodoHandlers
   ├── CRUD /api/guestbook → GuestbookHandlers
   └── GET/PUT /api/settings → SettingHandlers
```

---

### Day 29-30（高强度）— 给别人讲

**找一个人（同学、室友、甚至对着手机录音），完整讲一遍：**

1. 用户打开网页后发生了什么（从前端到数据库再回来）
2. IIFE 是什么，为什么用
3. fetch 怎么发请求的
4. 后端路由怎么匹配的
5. GORM 怎么操作数据库的
6. sessionStorage 怎么防刷新重复计数的

**如果讲到一半卡住了 → 那个地方就是你还没完全理解的，回去补。**

---

### Day 31（低强度休息日）

- [ ] 学操作系统
- [ ] 刷 LeetCode

---

# 第6周：最终验收 + 收尾

## 周目标

> 完成所有验收，整理最终笔记，发布掘金文章。

---

### Day 32-33（高强度）— 最终验收

**逐项检查：**

```
□ 能画出项目的完整架构图（前端/后端/数据库三层）
□ 能说出任意一个 API 的完整调用链（前端→路由→handler→DB→返回→渲染）
□ 能读懂 main.go 每一行
□ 能读懂 handler/visit.go 每一行
□ 能读懂 index.html 第3530-3646行 每一行
□ 能说出 GORM 每个方法对应的 SQL
□ 能解释 IIFE、sessionStorage、fetch、JSON 的作用
□ 能给别人讲清楚"我的博客是怎么记录访问量的"
□ 至少做过一次功能改动并部署验证
```

**有任何一项不通过的，Day 33 专门补。**

---

### Day 34-35（高强度）— 完善掘金文章 + 整理笔记

把 Day 5 的掘金文章草稿完善，加入后端和数据库的理解。

**文章升级版结构：**

```
标题：《我的博客一刷新访问量就+1？从前端到数据库全链路排查与修复》

1. 问题描述
2. 前端排查（IIFE + fetch）
3. 后端分析（路由 + handler + GORM）
4. 数据库验证（SQL 查询验证）
5. 解决方案（sessionStorage）
6. 全链路流程图
7. 知识总结
```

---

### Day 36-38（低强度收尾）

- Day 36：整理所有笔记，按模块归档
- Day 37：回顾整个学习过程，写一篇学习心得
- Day 38：制定下一步计划（学什么？做什么项目？）

---

# 附录

## A. 知识点速查表

### 前端核心概念

| 概念 | 代码位置 | 一句话解释 |
|------|---------|-----------|
| IIFE | index.html 第3530行 | 定义完立刻执行的函数，防止全局污染 |
| sessionStorage | index.html 第3596行 | 标签页级存储，关闭标签页就清空 |
| fetch | index.html 第3600行 | 浏览器发HTTP请求的API |
| JSON.stringify | index.html 第3603行 | JS对象转字符串 |
| DOM操作 | index.html 第3576行 | 用JS修改HTML元素内容 |
| navigator.userAgent | index.html 第3605行 | 浏览器自动提供的用户信息字符串 |

### 后端核心概念

| 概念 | 代码位置 | 一句话解释 |
|------|---------|-----------|
| 路由 | main.go 第206行 | URL路径和处理函数的映射 |
| handler | visit.go 第81行 | 接收请求、处理逻辑、返回响应的函数 |
| json.Decode | visit.go 第82行 | 把JSON字符串解析成Go结构体 |
| GORM Where | visit.go 第108行 | 对应SQL的WHERE条件 |
| GORM Create | visit.go 第125行 | 对应SQL的INSERT |
| GORM Updates | visit.go 第136行 | 对应SQL的UPDATE |
| sendJSON | helpers.go | 统一把Go结构体编码成JSON返回给前端 |
| ServeMux | main.go 第199行 | Go内置路由器 |
| ListenAndServe | main.go 第238行 | 启动HTTP服务器，阻塞等待请求 |

### 数据库核心概念

| 概念 | 对应代码 | SQL |
|------|---------|-----|
| 查询 | database.Where().First() | SELECT WHERE LIMIT 1 |
| 插入 | database.Create() | INSERT INTO |
| 更新 | database.Updates() | UPDATE SET |
| 删除 | database.Delete() | DELETE WHERE |
| 求和 | Select("SUM(...)").Scan() | SELECT SUM(...) |
| 计数 | Select("COUNT(*)").Scan() | SELECT COUNT(*) |
| 建表 | AutoMigrate() | CREATE TABLE IF NOT EXISTS |
| 唯一索引 | gorm:"uniqueIndex" | CREATE UNIQUE INDEX |

---

## B. GORM ↔ SQL 速查

```go
// 查询
database.Where("visitor_ip = ?", ip).First(&record)
// → SELECT * FROM visit_stats WHERE visitor_ip = 'xxx' LIMIT 1;

database.Where("visit_count > ?", 5).Find(&records)
// → SELECT * FROM visit_stats WHERE visit_count > 5;

// 插入
database.Create(&newRecord)
// → INSERT INTO visit_stats (visitor_ip, visit_count, ...) VALUES ('xxx', 1, ...);

// 更新
database.Model(&existing).Updates(map[string]interface{}{
    "visit_count": gorm.Expr("visit_count + 1"),
})
// → UPDATE visit_stats SET visit_count = visit_count + 1 WHERE id = ?;

// 删除
database.Delete(&record)
// → UPDATE visit_stats SET deleted_at = NOW() WHERE id = ?;  (软删除)
// → 或 DELETE FROM visit_stats WHERE id = ?;  (硬删除)

// 聚合
database.Model(&model.VisitStats{}).Select("SUM(visit_count)").Scan(&total)
// → SELECT SUM(visit_count) FROM visit_stats;

database.Model(&model.VisitStats{}).Select("COUNT(*)").Scan(&count)
// → SELECT COUNT(*) FROM visit_stats;
```

---

## C. HTTP 请求速查

### fetch POST 示例（你的代码第3600行）

```javascript
fetch('/api/visit', {
    method: 'POST',                              // 请求方法
    headers: { 'Content-Type': 'application/json' }, // 声明body是JSON
    body: JSON.stringify({ visitor_ip: '', browser: 'Chrome' }) // 请求数据
})
.then(res => res.json())    // 把响应体解析成JS对象
.then(data => console.log(data))  // 使用数据
.catch(err => console.error(err)) // 错误处理
```

### fetch GET 示例（你的代码第3617行）

```javascript
fetch('/api/visit/stats')    // 默认GET，不需要body
.then(r => r.json())
.then(data => console.log(data))
```

### GET vs POST

| | GET | POST |
|---|-----|------|
| 用途 | 向服务器要数据 | 给服务器传数据 |
| body | 没有 | 有 |
| 你的代码里 | `/api/visit/stats` | `/api/visit` |
| 语义 | 读 | 写 |

---

## D. 每日时间分配建议

```
┌─────────────┬──────────┬──────────┐
│   时间       │  内容    │  方式     │
├─────────────┼──────────┼──────────┤
│ 前30min     │  读代码  │  精读+注释│
│ 中间60min   │  学概念  │  文档/视频│
│ 后30-60min  │  动手改  │  改代码验证│
└─────────────┴──────────┴──────────┘
```

**核心原则：读代码占30%，学概念占40%，动手改占30%。只看不练记不住。**

---

## E. 学习资源推荐

| 内容 | 资源 | 说明 |
|------|------|------|
| JavaScript 进阶 | MDN Web Docs | 免费，权威 |
| HTTP 协议 | 《图解HTTP》 | 薄书，一天能看完 |
| Go 语言 | tour.go-programming.cn | 官方中文交互式教程 |
| GORM | gorm.io/docs | 只看你用到的那些方法 |
| MySQL | B站"MySQL入门" | 播放量高的就行 |

---

*本指南由 CodeBuddy AI 生成，学习过程中有任何问题随时提问！*
