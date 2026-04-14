# 从0理解整个 tinyWeb1 项目

> **目标**：深度理解项目每一行代码，能给别人讲清楚，能自己改功能
> **周期**：6周（41天）
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
第1周：████░░░░░ 高高高中低      （Day1-3强 Day4低 Day5休）← Day1-2已完成！
第2周：░░█████░░ 低高高高中      （Day6低 Day7-11强 Day12休）
第3周：░░█████░░ 低高高高高中    （Day13低 Day14-18强 Day19休）
第4周：█████░░░░ 高高高中高中低   （Day20-24强 Day25-26休 Day27-28中低）
第5周：░░█████░░ 低高高高中      （Day29低 Day30-33强 Day34休）
第6周：████░░░░░ 高高中低低低     （Day35-37强 Day38-41低强度收尾）

高强度日 = 2h 项目学习 + 写代码
低强度日 = 0.5-1h 回顾/看视频/整理笔记
休息日 = 不学项目，可以刷LeetCode或学操作系统
```

---

## 项目全貌

```
tinyWeb1/
├── fronted/index.html          ← 前端（一个文件包含所有，共4186行）
│   ├── 访问统计（visit）         ← JS: 第3787-3903行（IIFE）
│   ├── 用户认证（auth）          ← CSS: 第2040-2264行 / HTML: 第2320-2360行 / JS: 第3905-4183行（AuthManager）
│   ├── 备忘录（todo）            ← 搜索 todo 关键字
│   ├── 留言板（guestbook）       ← 搜索 guestbook 关键字
│   ├── 主题切换（setting）       ← 搜索 theme 关键字
│   └── 页面UI/动画/交互          ← HTML/CSS部分
│
└── server(数据库代码)/           ← 后端
    ├── main.go                  ← 服务器入口 + 路由（含认证路由 第229-254行）
    ├── config/config.go         ← 配置管理
    ├── db/db.go                 ← 数据库连接
    ├── model/model.go           ← 数据结构定义（含User模型 第59-96行）
    ├── handler/
    │   ├── visit.go             ← 访问统计
    │   ├── auth.go              ← 用户认证（Register/Login/GetCurrentUser）
    │   ├── todo.go              ← 备忘录CRUD
    │   ├── guestbook.go         ← 留言板
    │   ├── setting.go           ← 主题设置
    │   └── helpers.go           ← 公共工具函数
    ├── middleware/
    │   └── auth.go              ← JWT认证中间件（拦截请求验token）
    ├── session/
    │   └── memory.go            ← Session管理（内存存储）
    └── utils/
        ├── jwt.go               ← JWT token生成与验证
        └── hash.go              ← bcrypt密码加密
```

**关键点：6个handler的代码模式基本一样**（auth多了一步密码加密和token生成）

```go
// visit/todo/guestbook/setting 的 handler 都是这个模式：
func XxxHandler(w http.ResponseWriter, r *http.Request) {
    json.Decode(r.Body)     // 解析请求
    database.Where()        // 查库
    database.Create/Update  // 写库
    sendJSON(w, ...)        // 返回响应
}

// auth handler 多了两步：
func AuthHandler(w http.ResponseWriter, r *http.Request) {
    json.Decode(r.Body)     // 解析请求
    utils.HashPassword(...) // 密码加密（注册时）
    utils.CheckPassword()   // 密码验证（登录时）
    utils.GenerateToken()   // 生成JWT（登录时）
    database.Create/Where   // 写库/查库
    sendJSON(w, ...)        // 返回响应
}
```

把访问统计吃透 = 懂了60%的项目。auth模块多了JWT和bcrypt，单独学。

---

# 第1周：前端代码精读（以访问统计为切入点）

## 周目标

> 能完整口述 `index.html` 第 3787-3903 行的每一行在干什么，为什么这么写。

---

### Day 1（高强度）— IIFE 和函数定义 ✅ 已完成

**学什么：** `index.html` 第 3787-3829 行，所有函数定义部分

**时间分配：**

| 时间 | 任务 | 验收标准 |
|------|------|---------|
| 前30min | 读第 3787-3795 行 | 能说出 `(function(){` 是什么、`'use strict'` 干嘛的 |
| 中间60min | 读第 3797-3822 行，4个检测函数 | 每个函数输入是什么、输出是什么、正则怎么匹配的 |
| 后30min | 读第 3824-3829 行 animateNumber | DOM操作classList的add/remove是干嘛的 |

**今天要搞懂的5个概念：**

```javascript
// 1. IIFE — 第3787行
(function() {
    // ...
})();          // ← 为什么最后要加()？

// 2. var vs let vs const
var API_BASE = '';      // 你代码用的var，和let有什么区别？

// 3. function 声明 vs 函数表达式
function getVisitorIP() { ... }   // 第3791行，函数声明
var detectDeviceType = function() { ... } // 如果写成这样呢？

// 4. navigator 对象 — 浏览器内置API
navigator.userAgent   // 浏览器自己提供的，不是你写的
document.referrer     // 同上

// 5. 正则表达式 — 第3799行
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

### Day 2（高强度）— 核心业务逻辑 ✅ 已完成

**学什么：** `index.html` 第 3830-3903 行，renderStats + fetch + sessionStorage

**时间分配：**

| 时间 | 任务 |
|------|------|
| 前30min | 读第 3830-3845 行 renderStats 函数，逐行注释 |
| 中间60min | 读第 3847-3903 行，核心逻辑，画流程图 |
| 后30min | 打开浏览器 F12 → Console，手动执行代码验证 |

**今天要搞懂的5个概念：**

```javascript
// 1. DOM 操作 — 第3831行
document.getElementById('totalVisits')
// 这返回的是什么？一个HTML元素对象
// .textContent = 27  把这个元素的文字改成 "27"

// 2. sessionStorage — 第3848行
sessionStorage.getItem('visit_recorded')   // 取值，没有返回null
sessionStorage.setItem('visit_recorded', 'true')  // 存值

// 3. fetch API — 第3853行
fetch(url, options).then(fn1).then(fn2).catch(fn3)
//                         ↑        ↑       ↑
//                      成功回调  再成功   出错了

// 4. Promise / .then() — 异步
// fetch 发请求需要时间，不会卡住后面的代码
// .then() 表示"等结果回来了再执行这个函数"

// 5. JSON 序列化 — 第3856行
JSON.stringify({a:1})   →  '{"a":1}'     对象→字符串
JSON.parse('{"a":1}')   →  {a:1}         字符串→对象
```

**重点理解第 3848-3902 行的 if/else 分支：**

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
| `visit` | POST | 记录一次访问（第3853行的fetch发的） |
| `visit/stats` | GET | 获取统计数据（第3870行的fetch发的） |
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
                                          ↑ 这就是你第3856行的body！

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

### Day 4（低强度）— 前端全流程贯通 + 总结

**学什么：** 把前3天学的串起来，形成完整认知，然后本周收工

**时间分配：**

| 时间 | 任务 |
|------|------|
| 前40min | 从头到尾再读一遍 3787-3903 行，这次带着理解读 |
| 中间50min | 画一张完整的流程图（纸笔画或用工具） |
| 后30min | 写博客/掘金文章初稿（Day5内容合并到今天） |

**你要画的流程图（至少包含这些节点）：**

```
用户打开网页
    ↓
浏览器加载 index.html，解析到 <script>
    ↓
IIFE 开始执行 (第3787行)
    ↓
检查 sessionStorage (第3848行)
    ├── 没有 → 存标记 → fetch POST (第3853行)
    │                ↓
    │           浏览器发送HTTP请求
    │                ↓
    │           服务器接收处理
    │                ↓
    │           收到响应 → fetch GET stats (第3870行)
    │                ↓
    │           renderStats() 显示数据 (第3875行)
    │
    └── 有 → 跳过POST → fetch GET stats (第3895行)
                      ↓
                 renderStats() 显示数据 (第3900行)
```

**动手练习（选做）：**

给第 3787-3903 行每一行加上中文注释（如果你Day1-2已经写过博客了，可以跳过）。

**Day 4 过关标准：**
- [ ] 流程图覆盖了从页面加载到数据显示的全过程
- [ ] 合上文件，能凭记忆说出大致流程
- [ ] 博客/掘金文章初稿完成（Day5内容合并）

---

### Day 5（低强度休息日）

**不做新内容。可选：**

- [ ] 刷 1-2 道 LeetCode（保持手感）
- [ ] 学操作系统（进程线程相关）
- [ ] 回顾前4天的笔记
- [ ] 整理 F12 截图

---

### Day 6（低强度）— 休息 / 整理

**不做新内容。可选：**

- [ ] 整理博客
- [ ] 学操作系统
- [ ] 刷 LeetCode

---

# 第2周：Go 语言入门 + 后端代码精读

## 周目标

> 能读懂 `handler/visit.go` 的每一个函数，知道 Go 的基本语法。
> 能读懂 `handler/auth.go` 的三个认证函数。

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
| 59-69 | `User struct` | 用户模型：用户名唯一索引、密码哈希json:"-"不返回前端、角色默认user |
| 71-96 | `RegisterRequest` / `LoginRequest` / `LoginResponse` / `UserInfo` | 认证相关的请求/响应结构体 |
| 127-154 | `VisitStats struct` | 每个字段含义、gorm tag、json tag |
| 162-177 | `VisitRecord` / `VisitStatsResponse` | 请求体结构体、响应体结构体 |

**重点搞懂 tag：**

```go
// User struct 的关键字段（第59-64行）
Username     string `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
PasswordHash string `gorm:"type:varchar(255);not null" json:"-"`
//                                              ↑
//                                   json:"-" 表示这个字段永远不出现在JSON响应中！
//                                   因为密码哈希是敏感信息，绝对不能返回给前端
Role         string `gorm:"type:varchar(20);default:user;not null" json:"role"`
//                                     ↑
//                           default:user 新注册用户默认是普通用户

// VisitStats 的 tag（第127-147行）
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
// 练习1：自己定义一个 User struct，加 gorm tag 和 json tag
type User struct {
    ID           uint   `gorm:"primarykey" json:"id"`
    Username     string `gorm:"type:varchar(50);uniqueIndex" json:"username"`
    PasswordHash string `gorm:"type:varchar(255)" json:"-"`        // 注意 json:"-"
    Role         string `gorm:"type:varchar(20);default:user" json:"role"`
}

// 练习2：JSON序列化看看 json:"-" 的效果
type Student struct {
    ID   int    `gorm:"primarykey" json:"id"`
    Name string `gorm:"type:varchar(20)" json:"name"`
    Age  int    `json:"age"`
}

// 练习2：JSON序列化看看效果
import "encoding/json"

u := User{ID: 1, Username: "zhangsan", PasswordHash: "$2a$10$xxx", Role: "user"}
data, _ := json.Marshal(u)
fmt.Println(string(data))
// 输出什么？注意 PasswordHash 字段会不会出现！

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
- [ ] User 和 VisitStats 的每个字段都能说出用途和对应的 gorm tag 含义
- [ ] 能解释 `json:"-"` 为什么重要（密码不能返回前端）
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
    └─ 第83行 startServer()    ★ 重点！
        ├─ 第197行 rootDir = config.GetStaticDir()
        │   └─ 决定静态文件的根目录（你的 index.html 就在这里）
        │
        ├─ 第207行 mux := http.NewServeMux()
        │   └─ 创建路由器（空的映射表）
        │
        ├─ 第211行 mux.HandleFunc("/api/health", ...)
        │   └─ 注册路由1：健康检查
        │
        ├─ 第214行 mux.HandleFunc("/api/visit", ...)
        │   ├─ POST → handler.RecordVisit(w,r)
        │   └─ 其他方法 → 405 错误
        │
        ├─ 第221行 mux.HandleFunc("/api/visit/stats", ...)
        │   └─ GET → handler.GetVisitStats(w,r)
        │
        ├─ 第230行 mux.HandleFunc("/api/auth/register", ...)  ★ Day1新增
        │   └─ POST → handler.Register(w,r)
        │
        ├─ 第239行 mux.HandleFunc("/api/auth/login", ...)     ★ Day2新增
        │   └─ POST → handler.Login(w,r)
        │
        ├─ 第248行 mux.HandleFunc("/api/auth/me", ...)        ★ Day2新增
        │   └─ GET → middleware.AuthMiddleware(handler.GetCurrentUser)
        │       ↑ 注意：这个路由套了中间件！必须带token才能访问
        │
        ├─ 第258行 mux.Handle("/", fs)
        │   └─ 兜底路由：其他路径返回静态文件（index.html等）
        │
        ├─ 第276行 http.ListenAndServe(addr, corsMiddleware(mux))
        │   └─ 启动监听，程序阻塞在这里等待请求
        │
        └─ 第346行 corsMiddleware(next Handler) Handler
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

// 4. 中间件 — 请求的"安检门"
// 普通路由：请求 → handler
// 带中间件：请求 → 中间件(验token) → handler
// 第248行：middleware.AuthMiddleware(handler.GetCurrentUser)
//          ↑ 先验token，通过后才执行GetCurrentUser
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

### Day 13（高强度）— 读 handler/auth.go（认证核心！）

**这是新增的认证模块，比 visit.go 多了密码加密和 JWT 两个概念。**

#### 函数 1：Register（第61-125行）— 用户注册

```go
func Register(w http.ResponseWriter, r *http.Request) {
    // 步骤和 visit.go 类似，但多了密码加密
    
    // 1. 解析请求体（第63-67行）
    var req model.RegisterRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    // 2. 参数校验（第70-84行）
    // 用户名不能空、至少3字符、密码至少6位
    
    // 3. 检查用户名是否已存在（第87-99行）
    database.Where("username = ?", req.Username).First(&existingUser)
    if result.Error == nil {
        // 找到了 → 409 用户名已存在
    }
    
    // 4. bcrypt 加密密码（第101-106行）★ 新概念
    hashedPassword, err := utils.HashPassword(req.Password)
    // 数据库里存的是哈希值，不是明文！
    // 同一个密码每次加密结果都不同（因为bcrypt自带盐值）
    
    // 5. 创建用户（第108-117行）
    user := model.User{
        Username:     req.Username,
        PasswordHash: hashedPassword,  // 存哈希，不存明文
        Role:         "user",           // 默认普通用户
    }
    database.Create(&user)
    
    // 6. 返回用户信息（第119-124行）
    // 注意：返回的是 UserInfo{ID, Username, Role}
    // PasswordHash 不会返回（json:"-" 的作用）
}
```

#### 函数 2：Login（第156-208行）— 用户登录

```go
func Login(w http.ResponseWriter, r *http.Request) {
    // 1-3. 解析+校验+查用户（和Register类似）
    
    // 4. 验证密码（第183-187行）★ 新概念
    if !utils.CheckPassword(req.Password, user.PasswordHash) {
        // 密码不匹配 → 401
    }
    // bcrypt.CompareHashAndPassword 内部会：
    //   从 hash 中提取盐值 → 用同样的参数对输入密码哈希 → 对比
    
    // 5. 生成 JWT token（第189-194行）★ 新概念
    token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
    // JWT = Header.Payload.Signature 三段式
    // Payload 里有 user_id, username, role, exp(过期时间)
    
    // 6. 返回 token + 用户信息（第196-204行）
    sendJSON(w, http.StatusOK, model.SuccessResponse(model.LoginResponse{
        Token: token,
        User:  model.UserInfo{...},
    }))
    
    // 7. 创建 Session（第206-207行）
    session.Create(user.ID, user.Username, user.Role, token)
}
```

#### 函数 3：GetCurrentUser（第228-245行）— 获取当前用户

```go
func GetCurrentUser(w http.ResponseWriter, r *http.Request) {
    // 这个函数前面套了 AuthMiddleware 中间件！
    // 中间件已经验过 token 了，用户信息在 context 里
    
    // 1. 从 context 取用户信息（第230-237行）
    userID, ok := middleware.GetUserID(r.Context())
    username, _ := middleware.GetUsername(r.Context())
    role, _ := middleware.GetRole(r.Context())
    
    // 2. 直接返回，不需要查数据库
    sendJSON(w, http.StatusOK, model.SuccessResponse(model.UserInfo{...}))
}
```

**和 visit.go 的对比：**

| 对比项 | visit.go | auth.go |
|--------|----------|---------|
| 请求解析 | VisitRecord | RegisterRequest / LoginRequest |
| 数据库操作 | INSERT 或 UPDATE | INSERT（注册）或 SELECT（登录） |
| 特殊操作 | 无 | bcrypt加密/验密 + JWT生成 |
| 响应格式 | 访问统计数据 | token + 用户信息 |
| 中间件 | 无 | /api/auth/me 套了 AuthMiddleware |

**动手练习（必须写）：**

```go
// 练习1：模拟 Register 的核心逻辑（用map代替数据库）
var users = make(map[string]string)  // username → hashedPassword

func register(username, password string) string {
    if _, exists := users[username]; exists {
        return "用户名已存在"
    }
    users[username] = "hashed_" + password  // 模拟bcrypt
    return "注册成功: " + username
}

func login(username, password string) string {
    hash, exists := users[username]
    if !exists || hash != "hashed_"+password {  // 模拟CheckPassword
        return "用户名或密码错误"
    }
    return "登录成功，token: jwt_" + username  // 模拟GenerateToken
}

fmt.Println(register("zhangsan", "123456"))
fmt.Println(register("zhangsan", "123456"))  // 重复注册？
fmt.Println(login("zhangsan", "123456"))
fmt.Println(login("zhangsan", "wrong"))      // 密码错？

// 练习2：读 utils/hash.go 和 utils/jwt.go，每个函数写一行注释
```

**Day 13 过关标准：**
- [ ] Register 的 6 个步骤能口述
- [ ] Login 的 7 个步骤能口述
- [ ] 能说清 bcrypt 和 JWT 各自的作用
- [ ] 能解释为什么 PasswordHash 用 json:"-"
- [ ] 练习1跑通

---

### Day 14（高强度）— 读 utils/ + middleware/ + session/

**认证模块的三个辅助包，每个都很短。**

#### utils/hash.go（48行）— 密码加密

```go
// HashPassword（第31-34行）
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    // bcrypt.DefaultCost = 10，表示 2^10 = 1024 轮迭代
    // 每次加密约 100ms，安全性和性能的平衡点
    return string(bytes), err)
}

// CheckPassword（第44-47行）
func CheckPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
    // bcrypt 从 hash 中提取盐值，用同样的参数对输入密码哈希，对比结果
}
```

#### utils/jwt.go（105行）— Token 生成与验证

```go
// JwtSecretKey（第30行）— 签名密钥，必须保密！
var JwtSecretKey = []byte("tinyweb1-secret-key-2026")

// CustomClaims（第34-39行）— 自定义载荷
type CustomClaims struct {
    UserID   uint   `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims  // 包含 exp(过期时间)、iat(签发时间) 等
}

// GenerateToken（第46-73行）— 生成 JWT
func GenerateToken(userID uint, username, role string) (string, error) {
    // 1. 设置过期时间 = 当前 + 24小时
    // 2. 创建 CustomClaims
    // 3. jwt.NewWithClaims(HS256算法, claims)
    // 4. token.SignedString(密钥) → 签名并返回字符串
}

// ValidateToken（第83-104行）— 验证 JWT
func ValidateToken(tokenString string) (*CustomClaims, error) {
    // 1. jwt.ParseWithClaims 解析 + 验证签名
    // 2. 检查签名算法是否为 HS256（防算法混淆攻击）
    // 3. 检查是否过期
    // 4. 提取 CustomClaims 返回
}
```

#### middleware/auth.go（110行）— JWT 中间件

```go
// AuthMiddleware（第41-79行）— 请求的"安检门"
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. 从请求头取 Authorization（第44行）
        authHeader := r.Header.Get("Authorization")
        
        // 2. 提取 Bearer token（第51-56行）
        // 格式："Bearer eyJhbGciOi..."
        parts := strings.Split(authHeader, " ")
        
        // 3. 验证 token（第59-68行）
        claims, err := utils.ValidateToken(tokenString)
        // 失败 → 401 未授权
        
        // 4. 注入用户信息到 context（第71-74行）
        ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
        ctx = context.WithValue(ctx, UsernameKey, claims.Username)
        ctx = context.WithValue(ctx, RoleKey, claims.Role)
        
        // 5. 调用下一个处理器（第77行）
        next.ServeHTTP(w, r.WithContext(ctx))
    }
}

// GetUserID / GetUsername / GetRole（第82-109行）
// 从 context 中提取中间件注入的用户信息
```

#### session/memory.go（88行）— Session 管理

```go
// SessionInfo（第28-34行）
type SessionInfo struct {
    UserID    uint      // 用户ID
    Username  string    // 用户名
    Role      string    // 角色
    LoginTime time.Time // 登录时间
    Token     string    // 关联的JWT token
}

// 用 sync.Map 存储（并发安全的map）
// Create / Get / Delete 三个方法
// 注意：当前是内存存储，重启服务器会丢失
```

**动手练习（必须写）：**

```go
// 练习1：模拟中间件模式
func authMiddleware(next func(string)) func(string) {
    return func(token string) {
        if token == "" {
            fmt.Println("❌ 未登录，拒绝访问")
            return
        }
        fmt.Println("✅ token验证通过")
        next(token)  // 调用下一个处理器
    }
}

func getUserInfo(token string) {
    fmt.Println("返回用户信息, token:", token)
}

// 包装后使用
protectedHandler := authMiddleware(getUserInfo)
protectedHandler("")           // ❌ 未登录
protectedHandler("jwt_token")  // ✅ 通过

// 练习2：画 auth 请求的完整流程图
// 注册：前端 → POST /api/auth/register → Register → bcrypt → INSERT → 返回
// 登录：前端 → POST /api/auth/login → Login → bcrypt验密 → JWT → 返回token
// 获取用户：前端 → GET /api/auth/me (带token) → AuthMiddleware → GetCurrentUser → 返回
```

**Day 14 过关标准：**
- [ ] 能说清 bcrypt 为什么比 MD5 更适合存密码
- [ ] 能画出 JWT 的三段结构（Header.Payload.Signature）
- [ ] 能解释中间件的工作流程（请求 → 中间件 → handler）
- [ ] 能解释 context.WithValue 的作用（中间件给handler传数据）
- [ ] 练习1跑通

---

### Day 15（高强度）— 前后端联调走查

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

### Day 16（高强度）— 第2周总结 + 补漏

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
□ struct 和 tag 理解了吗？（特别是 json:"-" ）
□ ServeMux 路由机制明白了吗？
□ RecordVisit 四步流程能默写吗？
□ Register / Login 的步骤能口述吗？
□ bcrypt 和 JWT 各自的作用能说清吗？
□ 中间件的工作流程明白了吗？
□ 每个 GORM 操作对应哪条 SQL？
□ 前端 fetch 到后端 handler 到 MySQL 全程能串起来吗？
```

**有任何一项打 ❌ 的，今天就专门补这个。**

---

### Day 17（低强度休息日）

- [ ] 学操作系统
- [ ] 刷 1-2 题 LeetCode
- [ ] 整理本周笔记

---

# 第3周：数据库 + 认证精读 + 其他模块

## 周目标

> 理解 MySQL 数据库操作，GORM 与 SQL 的对应关系，精读认证模块，快速看懂其他模块。

---

### Day 18（高强度）— SSH 进 MySQL 实操

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

-- 11. 看认证功能的 users 表
DESCRIBE users;
-- 对照 model.go 第59-64行

-- 12. 查看注册的用户
SELECT id, username, role, created_at FROM users;
-- 看看密码哈希存在哪个字段？password_hash 列

-- 13. 看看密码哈希长什么样（选一条记录）
SELECT username, password_hash FROM users LIMIT 1;
-- 应该类似 $2a$10$xxxxx... 这样的格式，60字符
-- 这就是 bcrypt 哈希！同一个密码每次注册结果都不同
```

**Day 18 过关标准：**
- [ ] 上面13条SQL全部在服务器上执行过
- [ ] 每条SQL的执行结果都理解了
- [ ] 能说出 users 表和 visit_stats 表的字段差异

---

### Day 19（高强度）— GORM 与 SQL 对应

把 Day 18 每条 SQL 和 Go 代码对应上：

```go
database.Create(&newRecord)           ↔ INSERT INTO
database.Where().First(&existing)    ↔ SELECT WHERE LIMIT 1
database.Model().Updates(...)         ↔ UPDATE SET
database.Delete(&record)             ↔ DELETE WHERE
database.Select("SUM(...)").Scan(&x)  ↔ SELECT SUM(...)
database.Select("COUNT(*)").Scan(&x)  ↔ SELECT COUNT(*)
database.AutoMigrate(&Model{})        ↔ CREATE TABLE IF NOT EXISTS

// 认证模块的 GORM ↔ SQL（新增）
database.Where("username = ?", name).First(&user)  ↔ SELECT * FROM users WHERE username='xxx' LIMIT 1
database.Create(&user)                               ↔ INSERT INTO users (username, password_hash, role) VALUES (...)
```

**动手练习（必须写）：**

```go
// 用 GORM 写出以下 SQL 对应的 Go 代码：
// 1. SELECT * FROM visit_stats WHERE browser = 'Chrome';
// 2. UPDATE visit_stats SET visit_count = 10 WHERE visitor_ip = '127.0.0.1';
// 3. DELETE FROM visit_stats WHERE id = 1;
```

**Day 19 过关标准：**
- [ ] 每个 GORM 方法都能说出对应的 SQL
- [ ] 练习跑通

---

### Day 20（高强度）— 读其他 handler

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

**Day 20 过关标准：**
- [ ] 能说出每个 handler 的 CRUD 函数名
- [ ] 能指出和 visit.go 的模式差异（如果有的话）

---

### Day 21（高强度）— 数据库设计分析 + 动手改功能

**思考几个问题：**

```
□ 为什么要用 visitor_ip 做 uniqueIndex？
  → 同一IP只存一条，多次访问更新count

□ 为什么 username 也要 uniqueIndex？
  → 防止两个用户注册同一个用户名

□ DeletedAt 字段是干嘛的？
  → 软删除，GORM自带，DELETE时不真删而是标记时间

□ password_hash 为什么用 json:"-"？
  → 防止密码哈希通过API返回前端，安全！

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

**Day 21 过关标准：**
- [ ] 选做的功能改动完成并部署验证
- [ ] 能说出改了哪些文件、每处改动的原因

---

### Day 22（高强度）— 认证全链路精读

**目标：从"点击登录按钮"到"看到用户名"，每一步都能说清。**

**任务：画认证流程时序图**

```
注册流程：
浏览器                              Go服务器                          MySQL
  │                                   │                                │
  │ ① 点击"注册" → AuthManager.handleSubmit()                        │
  │ ② fetch POST /api/auth/register   │                                │
  │ body: {username, password}         │                                │
  │==================================▶│                                │
  │                                   │ ③ json.Decode 解析body        │
  │                                   │ ④ 参数校验（3字符/6位）         │
  │                                   │ ⑤ SELECT WHERE username='...' │
  │                                   │===============================▶│
  │                                   │                          ⑥ 没找到=可以注册
  │                                   │◀==============================│
  │                                   │ ⑦ utils.HashPassword(password)│
  │                                   │    bcrypt加密，约100ms          │
  │                                   │ ⑧ INSERT INTO users           │
  │                                   │===============================▶│
  │                                   │                          ⑨ 写入成功
  │                                   │◀==============================│
  │                                   │ ⑩ sendJSON 201 Created       │
  │◀==================================│                                │
  │ ⑪ 注册成功 → 自动切换到登录弹窗    │                                │

登录流程：
浏览器                              Go服务器                          MySQL
  │                                   │                                │
  │ ① 点击"登录" → AuthManager.handleSubmit()                        │
  │ ② fetch POST /api/auth/login      │                                │
  │ body: {username, password}         │                                │
  │==================================▶│                                │
  │                                   │ ③ json.Decode 解析body        │
  │                                   │ ④ SELECT WHERE username='...' │
  │                                   │===============================▶│
  │                                   │                          ⑤ 找到用户
  │                                   │◀==============================│
  │                                   │ ⑥ utils.CheckPassword(明文, 哈希)│
  │                                   │    bcrypt验证，约100ms          │
  │                                   │ ⑦ utils.GenerateToken(id,name,role)│
  │                                   │    生成JWT，有效期24h           │
  │                                   │ ⑧ session.Create() 记录会话    │
  │                                   │ ⑨ sendJSON 200 + token        │
  │◀==================================│                                │
  │ ⑩ localStorage.setItem(token)     │                                │
  │ ⑪ AuthManager.showLoggedIn(user)  │                                │
  │ ⑫ 页面右上角显示"用户名 + 退出"    │                                │

获取当前用户：
浏览器                              Go服务器                          MySQL
  │                                   │                                │
  │ ① 页面加载 → AuthManager.checkAuthStatus()                       │
  │ ② fetchWithAuth GET /api/auth/me  │                                │
  │ Header: Authorization: Bearer jwt │                                │
  │==================================▶│                                │
  │                                   │ ③ AuthMiddleware 拦截           │
  │                                   │    提取 Bearer token            │
  │                                   │ ④ utils.ValidateToken(token)   │
  │                                   │    验签名+过期时间               │
  │                                   │ ⑤ context注入 userID/username   │
  │                                   │ ⑥ GetCurrentUser()             │
  │                                   │    从context取信息，不查DB！     │
  │                                   │ ⑦ sendJSON 200                 │
  │◀==================================│                                │
  │ ⑧ token有效 → showLoggedIn()      │                                │
  │    token无效 → clearToken + showLoggedOut()                        │
```

**每个步骤标注：对应代码哪个文件哪一行。**

**Day 22 过关标准：**
- [ ] 注册的 11 个步骤能口述
- [ ] 登录的 12 个步骤能口述
- [ ] 能解释为什么 GetCurrentUser 不需要查数据库（因为中间件已经从token解析出来了）
- [ ] 能说出 localStorage 和 sessionStorage 的区别

---

### Day 23（低强度休息日）

- [ ] 操作系统学习
- [ ] 整理笔记

---

# 第4周：复习 + 深入 + 安全分析

## 周目标

> 第二轮精读代码，理解安全和性能问题，加深理解。

---

### Day 24-25（高强度）— 第二轮精读 main.go + visit.go + auth.go

重新读 main.go、visit.go、auth.go、index.html（3787-3903行 + 3905-4183行）。
这次读应该比第一遍快很多。

**标注：**
- 第一次读不懂但现在懂了的地方 → 绿色标记
- 还是不懂的地方 → 红色标记，重点攻克

---

### Day 26（高强度）— 安全和性能分析

带着审视的眼光重新看你的代码：

**安全问题：**

```
□ SQL注入？  → GORM参数化查询，安全 ✅
□ XSS？      → AuthManager有escapeHtml()，部分防护 ✅ / 其他模块没转义 ⚠️
□ CSRF？     → 没有token保护 ⚠️
□ 密码明文？ → 用bcrypt加密存储 ✅ / 但root密码为空 ⚠️
□ JWT密钥？  → 硬编码在代码里，生产环境应从环境变量读 ⚠️
□ 限流？     → 同一IP可以无限刷 visit / 无限注册 ⚠️
□ token过期？ → 24小时过期，但没有刷新token机制 ⚠️
```

**性能问题：**

```
□ 每次访问都查一次DB？ → 是的，可以优化用缓存 ⚠️
□ 连接池够用吗？       → MaxOpenConns=10，小项目够了 ✅
□ 前端并发请求？       → POST完了再GET，可以优化并行 ⚠️
```

**Day 26 过关标准：**
- [ ] 能说出3个以上的安全隐患和对应的修复思路
- [ ] 能说出2个以上的性能优化方向

---

### Day 27（高强度）— 读 index.html 其他模块

之前只读了访问统计（3787-3903行）和认证模块（3905-4183行），现在读备忘录和留言板的前端代码：

```
□ 搜索 index.html 里的 todo 相关 JS → 对应 todo.go
□ 搜索 index.html 里的 guestbook 相关 JS → 对应 guestbook.go
□ 找到前端 fetch 调用的 URL 和后端路由的对应关系
```

**Day 27 过关标准：**
- [ ] 能说出备忘录和留言板前端代码的位置
- [ ] 能找到每个 fetch 调用对应的后端 handler
- [ ] 能说出 AuthManager（3905-4183行）和访问统计IIFE（3787-3903行）的结构区别

---

### Day 28（低强度休息日）

- [ ] 学操作系统
- [ ] 刷 LeetCode
- [ ] 整理笔记

---

# 第5周：全链路贯通 + 架构理解

## 周目标

> 能不看代码画出完整架构图，能给别人讲清楚整个项目。

---

### Day 29（低强度）— 回顾前4周

重新过一遍所有笔记，标出：
- 已完全理解的知识点
- 还模糊的知识点
- 完全不懂的知识点

---

### Day 30-31（高强度）— 画完整架构图

**要求：不用看代码，凭记忆画。画完再对照代码修正。**

架构图应包含：

```
1. 前端层
   ├── HTML 结构（页面布局 + 认证弹窗）
   ├── CSS 样式（主题、动画、响应式、认证弹窗样式）
   └── JS 逻辑
       ├── 访问统计模块（IIFE + fetch + sessionStorage）第3787-3903行
       ├── 用户认证模块（AuthManager + localStorage + JWT）第3905-4183行
       ├── 备忘录模块
       ├── 留言板模块
       └── 主题设置模块

2. 后端层
   ├── main.go（入口 + 路由注册 + 认证路由）
   ├── config.go（配置管理）
   ├── db.go（数据库连接 + 连接池）
   ├── model.go（6个核心struct + 认证请求/响应struct）
   ├── handler/
   │   ├── visit.go（访问统计）
   │   ├── auth.go（注册 + 登录 + 获取当前用户）★
   │   ├── todo.go（备忘录CRUD）
   │   ├── guestbook.go（留言板）
   │   ├── setting.go（设置）
   │   └── helpers.go（sendJSON等工具函数）
   ├── middleware/
   │   └── auth.go（JWT认证中间件）★
   ├── session/
   │   └── memory.go（Session管理）★
   └── utils/
       ├── jwt.go（JWT生成与验证）★
       └── hash.go（bcrypt密码加密）★

3. 数据库层
   ├── users 表 ★（认证功能新增）
   ├── visit_stats 表
   ├── todos 表
   ├── guestbook 表
   └── settings 表

4. 前后端连接
   ├── POST /api/auth/register → Register → bcrypt → INSERT users ★
   ├── POST /api/auth/login → Login → bcrypt验密 → JWT → 返回token ★
   ├── GET /api/auth/me → AuthMiddleware(验token) → GetCurrentUser ★
   ├── POST /api/visit → RecordVisit
   ├── GET /api/visit/stats → GetVisitStats
   ├── CRUD /api/todos → TodoHandlers
   ├── CRUD /api/guestbook → GuestbookHandlers
   └── GET/PUT /api/settings → SettingHandlers
```

---

### Day 32-33（高强度）— 给别人讲

**找一个人（同学、室友、甚至对着手机录音），完整讲一遍：**

1. 用户打开网页后发生了什么（从前端到数据库再回来）
2. IIFE 是什么，为什么用
3. fetch 怎么发请求的
4. 后端路由怎么匹配的
5. GORM 怎么操作数据库的
6. sessionStorage 怎么防刷新重复计数的
7. 用户注册时密码是怎么加密存储的（bcrypt流程）
8. 用户登录时token是怎么生成和验证的（JWT流程）
9. 中间件是怎么拦截请求验token的

**如果讲到一半卡住了 → 那个地方就是你还没完全理解的，回去补。**

---

### Day 34（低强度休息日）

- [ ] 学操作系统
- [ ] 刷 LeetCode

---

# 第6周：最终验收 + 收尾

## 周目标

> 完成所有验收，整理最终笔记，发布掘金文章。

---

### Day 35-36（高强度）— 最终验收

**逐项检查：**

```
□ 能画出项目的完整架构图（前端/后端/数据库三层）
□ 能说出任意一个 API 的完整调用链（前端→路由→handler→DB→返回→渲染）
□ 能读懂 main.go 每一行
□ 能读懂 handler/visit.go 每一行
□ 能读懂 handler/auth.go 每一行（Register/Login/GetCurrentUser）
□ 能读懂 index.html 第3787-3903行 每一行
□ 能读懂 index.html 第3905-4183行 AuthManager 每个方法
□ 能说出 GORM 每个方法对应的 SQL
□ 能解释 IIFE、sessionStorage、fetch、JSON 的作用
□ 能解释 bcrypt、JWT、中间件、context 的作用
□ 能给别人讲清楚"我的博客是怎么记录访问量的"
□ 能给别人讲清楚"用户注册登录的完整流程"
□ 至少做过一次功能改动并部署验证
```

**有任何一项不通过的，Day 36 专门补。**

---

### Day 37-38（高强度）— 完善掘金文章 + 整理笔记

把博客文章完善，加入后端、数据库和认证模块的理解。

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
8. 附录：用户认证系统实现（bcrypt + JWT + 中间件）
```

---

### Day 39-41（低强度收尾）

- Day 39：整理所有笔记，按模块归档
- Day 40：回顾整个学习过程，写一篇学习心得
- Day 41：制定下一步计划（学什么？做什么项目？）

---

# 附录

## A. 知识点速查表

### 前端核心概念

| 概念 | 代码位置 | 一句话解释 |
|------|---------|-----------|
| IIFE | index.html 第3787行 | 定义完立刻执行的函数，防止全局污染 |
| sessionStorage | index.html 第3848行 | 标签页级存储，关闭标签页就清空 |
| fetch | index.html 第3853行 | 浏览器发HTTP请求的API |
| JSON.stringify | index.html 第3856行 | JS对象转字符串 |
| DOM操作 | index.html 第3831行 | 用JS修改HTML元素内容 |
| navigator.userAgent | index.html 第3805行 | 浏览器自动提供的用户信息字符串 |
| localStorage | index.html 第3962行 | 浏览器持久化存储，关浏览器不丢（vs sessionStorage） |
| AuthManager | index.html 第3916行 | 用户认证管理对象，封装登录/注册/token管理 |
| fetchWithAuth | index.html 第3975行 | 自动带Authorization头的fetch |
| escapeHtml | index.html 第4175行 | 防XSS：转义HTML特殊字符 |

### 后端核心概念

| 概念 | 代码位置 | 一句话解释 |
|------|---------|-----------|
| 路由 | main.go 第214行 | URL路径和处理函数的映射 |
| 认证路由 | main.go 第230-254行 | /api/auth/* 三个路由的注册 |
| handler | visit.go 第81行 | 接收请求、处理逻辑、返回响应的函数 |
| json.Decode | visit.go 第82行 | 把JSON字符串解析成Go结构体 |
| GORM Where | visit.go 第108行 | 对应SQL的WHERE条件 |
| GORM Create | visit.go 第125行 | 对应SQL的INSERT |
| GORM Updates | visit.go 第136行 | 对应SQL的UPDATE |
| sendJSON | helpers.go | 统一把Go结构体编码成JSON返回给前端 |
| ServeMux | main.go 第207行 | Go内置路由器 |
| ListenAndServe | main.go 第276行 | 启动HTTP服务器，阻塞等待请求 |
| bcrypt | utils/hash.go 第31行 | 密码哈希加密，自带盐值，每次加密结果不同 |
| JWT | utils/jwt.go 第46行 | 无状态token认证，三段式 Header.Payload.Signature |
| 中间件 | middleware/auth.go 第41行 | 请求的安检门，验token通过后才执行handler |
| context | middleware/auth.go 第72行 | Go的上下文传递机制，中间件给handler传数据 |
| Session | session/memory.go 第45行 | 服务端存储登录状态，当前用内存（重启丢失） |

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

### fetch POST 示例（你的代码第3853行）

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

### fetch GET 示例（你的代码第3870行）

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

### 认证 API 速查

```javascript
// 1. 注册（前端第4081行）
fetch('/api/auth/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: 'zhangsan', password: '123456' })
})
// 成功 → {code:0, data:{id:1, username:"zhangsan", role:"user"}}

// 2. 登录（前端第4081行）
fetch('/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: 'zhangsan', password: '123456' })
})
// 成功 → {code:0, data:{token:"eyJ...", user:{id:1, username:"zhangsan", role:"user"}}}

// 3. 获取当前用户（前端第3996行，需要token）
fetch('/api/auth/me', {
    headers: { 'Authorization': 'Bearer eyJhbGciOi...' }
})
// 成功 → {code:0, data:{id:1, username:"zhangsan", role:"user"}}
// 失败 → {code:401, message:"缺少认证token"}
```

### localStorage vs sessionStorage

| | localStorage | sessionStorage |
|---|---|---|
| 生命周期 | 永久，除非手动删除 | 标签页关闭就清空 |
| 你的代码用在哪 | 存JWT token（第3966行） | 存visit_recorded标记（第3849行） |
| 为什么不同 | token要持久保持登录 | 访问标记只需本次会话 |

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
