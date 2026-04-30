# 从0理解整个 tinyWeb1 项目

> **目标**：深度理解项目每一行代码，能自己改功能，能完整复述全流程。
> **周期**：约25天（3周半）
> **每日投入**：高强度日 2h / 中强度日 1-1.5h
> **节奏**：没有固定休息日，累了自己休息

---

## 个人画像

```
大一 / CS专业 / 每天2h高强度 / 4天一休息
LeetCode 60题(二叉树强+栈) / 操作系统学到进程线程
HTML/CSS/JS基础有 / Go和数据库零基础 / 目前全懵
目标：能讲清楚每行代码 / 能改功能 / 追求深度不追速度
```

---

## 强度节奏

```
第1周：███░░░░░░░░░░░░░ 高高高              Day1-3（✅已完成）
第2周：█████████░░░░░░░ 高高高高中高中       Day4-12（Go+后端全读完）
第3周：░░░░░█████░░░░░░ 高高高高中           Day13-17（数据库+认证+改功能）
第4周：░░░░░░░░████░░░░ 高高高高             Day18-21（二轮精读+安全+前端）
第5周：░░░░░░░░░░████░░ 高中中               Day22-25（架构图+验收+AI复述）

高强度 = 2h 项目学习 + 写代码
中强度 = 1-1.5h（看视频+实操+整理，类似Day3的强度）
不设休息日 = 自己决定哪天休息
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

**关键点：6个handler的代码模式基本一样**（auth多了密码加密和token生成）

```go
// 通用模式：json.Decode → database操作 → sendJSON
// auth多两步：utils.HashPassword / utils.GenerateToken
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
(function() { ... })();          // ← 为什么最后要加()？

// 2. var vs let vs const
var API_BASE = '';               // 你代码用的var，和let有什么区别？

// 3. function 声明 vs 函数表达式
function getVisitorIP() { ... }  // 第3791行，函数声明

// 4. navigator 对象 — 浏览器内置API
navigator.userAgent / document.referrer  // 浏览器提供的，不是你写的

// 5. 正则表达式 — 第3799行
/Android|iPhone/i.test(navigator.userAgent)  // .test() 返回 true/false
```

**动手练习（必须写）：**

```javascript
// 练习1：写一个自己的IIFE，打印"hello"
(function(){ console.log("hello"); })();

// 练习2：模仿 detectOS，写一个 detectBrowser 的简化版本
function myDetectBrowser(ua) {
    if (ua.includes('Chrome')) return 'Chrome';
    if (ua.includes('Firefox')) return 'Firefox';
    return 'Other';
}
console.log(myDetectBrowser(navigator.userAgent));

// 练习3：用正则判断一个字符串是不是手机号
function isPhone(str) { return /^1[3-9]\d{9}$/.test(str); }
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
// 1. DOM 操作 — document.getElementById('totalVisits').textContent = 27
// 2. sessionStorage — getItem/setItem，标签页级存储，关闭就清空
// 3. fetch API — fetch(url, options).then(fn1).then(fn2).catch(fn3)
// 4. Promise / .then() — 异步，"等结果回来了再执行"
// 5. JSON 序列化 — JSON.stringify({a:1})→'{"a":1}' / JSON.parse反方向
```

**重点理解第 3848-3902 行的 if/else 分支：**

```
sessionStorage 有标记？
├─ 没有（首次打开）→ 存标记 → POST 记录访问 → GET 获取统计 → 显示
└─ 有（刷新页面）    → 跳过POST → 直接 GET 获取统计 → 显示
```

**动手练习（必须写）：**

```javascript
// 练习1：在浏览器 F12 Console 里操作 sessionStorage
sessionStorage.setItem('test', 'hello')
sessionStorage.getItem('test')      // "hello"
sessionStorage.removeItem('test')

// 练习2：手动发 GET 请求
fetch('/api/visit/stats').then(r => r.json()).then(data => console.log(data))

// 练习3：写一个 fetch POST
fetch('/api/visit', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({visitor_ip:'', browser:'Test', os:'Linux'})
}).then(r => r.json()).then(data => console.log('服务器返回:', data))
```

**Day 2 过关标准：**
- [ ] 能画出 if/else 的分支流程图
- [ ] 能说清楚 fetch 的参数各是什么意思
- [ ] 3个练习都在 Console 里跑通看到结果

---

### Day 3（高强度）— HTTP 协议实战 + 前端全流程贯通 ✅ 已完成

**学什么：** 用你自己的项目当教材，看真实的 HTTP 报文，然后把前3天内容串起来

**时间分配：**

| 时间 | 任务 |
|------|------|
| 前30min | B站搜"HTTP协议详解"，看一个30min以内的视频 |
| 中间60min | 打开网站 → F12 Network → 刷新 → 逐个分析请求 |
| 后30min | 画完整流程图 + 整理笔记 |

**实操任务：**

打开 `http://1.15.224.88:8080`，按 F12 → 点 Network 标签 → 刷新页面。找到这几个请求：

| 请求名 | 方法 | 它是干嘛的 |
|--------|------|-----------|
| `visit` | POST | 记录一次访问（第3853行的fetch发的） |
| `visit/stats` | GET | 获取统计数据（第3870行的fetch发的） |
| `index.html` | GET | 加载页面本身 |

**对 `visit` 这个请求，点开看 Headers/Payload/Response 三个标签：**

- **Headers**：Request Method=POST, Content-Type=application/json
- **Payload**：`{visitor_ip:"", browser:"Chrome", os:"Windows"...}` — 对应第3856行的body
- **Response**：`{code:0, message:"success", data:{...}}` — 后端返回的数据

**动手练习：**

1. 在 F12 Network 里找到 `visit` 请求，截图保存
2. 对比 `visit` 和 `visit/stats` 的区别（方法？有没有Payload？）
3. 手动在浏览器地址栏输入 `http://1.15.224.88:8080/api/visit/stats` 回车

**前端全流程贯通（后30min）— 画完整流程图：**

```
用户打开网页 → 浏览器加载 index.html → IIFE 开始执行(第3787行)
→ 检查 sessionStorage(第3848行)
    ├── 没有 → 存标记 → fetch POST(第3853行) → 收到响应 → fetch GET stats(第3870行) → renderStats()
    └── 有 → 跳过POST → fetch GET stats(第3895行) → renderStats()
```

**Day 3 过关标准：**
- [ ] 能说清楚 POST 和 GET 在 Network 面板里的视觉区别
- [ ] 能指出 Payload 里哪个字段对应代码哪一行
- [ ] 流程图覆盖了从页面加载到数据显示的全过程
- [ ] 合上文件，能凭记忆说出大致流程

---

# 第2周：Go 语言入门 + 后端代码精读

## 周目标

> 能读懂 `handler/visit.go` 的每一个函数，知道 Go 的基本语法。
> 能读懂 `handler/auth.go` 的三个认证函数。

---

### Day 4（高强度）— Go 基础速通（环境搭建 + 语法对照C）

**你有C语言基础，Go 会很快。环境你之前搭过了，10min带过验证。**

| 时间 | 任务 |
|------|------|
| 10min | 验证环境：终端 `go version` 能显示版本号、VS Code Go 插件装好了 |
| 80min | Go 基础语法速通（对照C来学） |
| 30min | 动手练习 |

**今天要掌握的（对照C来学）：**

```go
// 1. 变量声明 — C是 int a = 1; Go有多种写法
var a int = 1           // 类似C，完整写法
b := 2                  // 短声明，自动推断类型（C没有）
const PI = 3.14          // 和C一样

// 2. 函数 — 可以多返回值（C没有！）
func divide(a, b int) (int, error) {
    if b == 0 { return 0, fmt.Errorf("不能除0") }
    return a / b, nil
}

// 3. struct — 和C几乎一样，类型写在后面
type Person struct { Name string; Age int }
p := Person{Name: "张三", Age: 20}

// 4. 方法 — Go特有，函数前面加receiver
func (p Person) SayHello() { fmt.Println("我是", p.Name) }

// 5. error 处理 — if err != nil 到处都是
result, err := divide(10, 3)
if err != nil { fmt.Println("出错:", err); return }

// 6. slice — 类似动态数组
nums := []int{1, 2, 3}
nums = append(nums, 4)

// 7. map — 哈希表
m := map[string]int{"张三": 90, "李四": 85}

// 8. 指针 — 和C一样：&取地址，*解引用
```

**动手练习（必须写）：**

```go
// 练习1：写一个 VisitStats 结构体
type VisitStats struct { IP string; Count int }
v := VisitStats{IP: "127.0.0.1", Count: 5}
fmt.Println(v.IP)   // 输出什么？

// 练习2：写一个带 error 返回值的函数
func GetIP(s string) (string, error) {
    if s == "" { return "", fmt.Errorf("IP为空") }
    return s, nil
}

// 练习3：用 map 统计字符出现次数
func countChars(s string) map[rune]int {
    m := make(map[rune]int) 
    for _, c := range s { m[c]++ }
    return m
}
fmt.Println(countChars("hello"))
```

**Day 4 过关标准：**
- [ ] 上面8个知识点每个都能写出示例代码
- [ ] 3个练习都跑通
- [ ] 能说出 Go 和 C 在变量声明、函数、error 处理上的区别

---

### Day 5（高强度）— 读 model.go + config.go

**这两个文件最简单，适合作为读Go代码的起点。**

**读 model.go，逐个 struct 分析：**

| 行号 | 内容 | 要搞懂的 |
|------|------|---------|
| 59-69 | `User struct` | 用户名唯一索引、密码哈希json:"-"不返回前端、角色默认user |
| 71-96 | 认证相关 Request/Response | RegisterRequest / LoginRequest / LoginResponse / UserInfo |
| 127-154 | `VisitStats struct` | 每个字段含义、gorm tag、json tag |
| 162-177 | 请求/响应结构体 | VisitRecord / VisitStatsResponse |

**重点搞懂 tag：**

```go
Username     string `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
PasswordHash string `gorm:"type:varchar(255);not null" json:"-"`
//                                              ↑ json:"-" 永远不出现在JSON响应中！密码是敏感信息
Role         string `gorm:"type:varchar(20);default:user;not null" json:"role"`

// gorm tag：告诉GORM数据库里这列怎么建（类型、索引、默认值等）
// json tag：JSON序列化时用什么key名（不写就用字段名大写开头）
```

**动手练习（必须写）：**

```go
// 练习1：自己定义一个带 gorm tag 和 json tag 的 struct
type Student struct {
    ID   int    `gorm:"primarykey" json:"id"`
    Name string `gorm:"type:varchar(20)" json:"name"`
    Age  int    `json:"age"`
}

// 练习2：JSON序列化看看 json:"-" 的效果
u := User{ID: 1, Username: "zhangsan", PasswordHash: "$2a$10$xxx", Role: "user"}
data, _ := json.Marshal(u)
fmt.Println(string(data))  // PasswordHash 字段会不会出现？

// 练习3：模拟环境变量读取
port := os.Getenv("SERVER_PORT")  // 空字符串则用默认值
if port == "" { port = ":8081" }
```

**Day 5 过关标准：**
- [ ] 能说出 gorm tag 和 json tag 各自的作用
- [ ] 能解释 `json:"-"` 为什么重要
- [ ] 3个练习跑通

---

### Day 6（高强度）— 读 main.go（服务器入口）

**这是整个后端的骨架。按执行顺序读：**

```
main()
 ├─ config.Load()        → 读环境变量 → AppConfig
 ├─ db.Initialize()      → 连MySQL → AutoMigrate建表
 ├─ db.InitializeTestDB()
 ├─ testVisitStats()
 └─ startServer() ★ 重点！
     ├─ http.NewServeMux()         → 创建路由器
     ├─ mux.HandleFunc("/api/health", ...)
     ├─ mux.HandleFunc("/api/visit", ...)          → POST→RecordVisit
     ├─ mux.HandleFunc("/api/visit/stats", ...)     → GET→GetVisitStats
     ├─ mux.HandleFunc("/api/auth/register", ...)   → POST→Register
     ├─ mux.HandleFunc("/api/auth/login", ...)      → POST→Login
     ├─ mux.HandleFunc("/api/auth/me", ...)         → GET→AuthMiddleware(GetCurrentUser)
     ├─ mux.Handle("/", fs)                         → 静态文件兜底
     ├─ http.ListenAndServe(addr, corsMiddleware(mux))
     └─ corsMiddleware()                            → 加跨域头
```

**重点理解：**

```go
// 1. ServeMux — 路由器，本质是 map[string]Handler
// 2. http.HandlerFunc — func(w ResponseWriter, r *Request) 的函数就能当处理器
// 3. ListenAndServe — 启动服务器，阻塞等待请求
// 4. 中间件 — 请求→中间件(验token)→handler，如 AuthMiddleware(GetCurrentUser)
```

**动手练习（必须写）：**

```go
// 练习1：写一个最简HTTP服务器
package main
import "net/http"
func helloHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello from tinyWeb1!"))
}
func main() {
    http.HandleFunc("/hello", helloHandler)
    http.ListenAndServe(":9999", nil)
}
// go run 后浏览器访问 localhost:9999/hello

// 练习2：加一个 JSON 返回的接口
func apiHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(`{"code":0,"message":"success"}`))
}
```

**Day 6 过关标准：**
- [ ] 能画出 main() 的调用链
- [ ] 能说清楚 HandleFunc 和 ListenAndServe 干嘛的
- [ ] 练习1的服务器跑起来并在浏览器访问

---

### Day 7（高强度）— 读 handler/visit.go（核心！）

**这是最重要的文件。前端和数据库在这里交汇。**

#### getClientIP（第39-54行）— 三种策略取真实IP

1. X-Forwarded-For 头（经过Nginx/CDN时）
2. X-Real-IP 头
3. RemoteAddr 兜底（去掉端口）

#### RecordVisit（第81-162行）— 最核心

```
步骤1：json.Decode 解析 body → VisitRecord struct
步骤2：补全 IP（req.VisitorIP == "" 时调用 getClientIP）
步骤3：查数据库
  ├─ ErrRecordNotFound → INSERT 新记录
  └─ 找到了 → UPDATE visit_count+1, last_visit_at
步骤4：sendJSON 返回响应
```

**每个 GORM 操作对应一条 SQL：**
```go
database.Where("visitor_ip = ?", ip).First(&existing)
// → SELECT * FROM visit_stats WHERE visitor_ip='xxx' LIMIT 1
database.Create(&newRecord)    // → INSERT INTO visit_stats VALUES(...)
database.Model(&existing).Updates(map[...]{...})
// → UPDATE visit_stats SET visit_count=visit_count+1 WHERE id=?
```

#### GetVisitStats（第186-206行）

```go
// 三条聚合查询
Select("COALESCE(SUM(visit_count), 0)")  // 总访问次数
Select("COUNT(*)")                        // 独立访客数
Select("MAX(last_visit_at)")              // 最后访问时间
```

**动手练习（必须写）：**

```go
// 练习1：模拟 RecordVisit 核心逻辑（用map代替数据库）
var db = make(map[int]string)
func record(id int, name string) string {
    for k, v := range db {
        if v == name { db[k] = name + "(更新)"; return "更新成功" }
    }
    db[len(db)+1] = name; return "新增成功"
}

// 练习2：手敲一遍 visit.go 每个函数，边抄边想
```

**Day 7 过关标准：**
- [ ] RecordVisit 的 4 个步骤能默写出来
- [ ] 每个 GORM 操作能写出对应的 SQL
- [ ] 练习1跑通

---

### Day 8（中强度）— helpers.go + db.go

**辅助文件，内容不多，快速过。**

```go
// helpers.go — sendJSON 三步：SetHeader → WriteHeader → Encode
func sendJSON(w http.ResponseWriter, statusCode int, response model.APIResponse) {
    w.Header().Set("Content-Type", "application/json;charset=utf-8")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(response)
}

// db.go — 连接池：MaxOpenConns=10, MaxIdleConns=5, ConnMaxLifetime=30min
// 为什么需要？每次查询都建TCP连接太慢，提前建好一批复用
```

**Day 8 过关标准：**
- [ ] 能说清 sendJSON 的三个步骤
- [ ] 能解释连接池为什么存在

---

### Day 9（高强度）— 读 handler/auth.go（认证核心！）

**比 visit.go 多了密码加密和 JWT。**

#### Register（第61-125行）— 6步

```
1. json.Decode 解析body → RegisterRequest
2. 参数校验（用户名≥3字符，密码≥6位）
3. 查数据库检查用户名是否已存在 → 已存在返回409
4. bcrypt加密密码 ★ — 同一密码每次结果不同（自带盐值）
5. 创建用户 INSERT — 存哈希不存明文
6. 返回 UserInfo（PasswordHash不会返回，json:"-"的作用）
```

#### Login（第156-208行）— 7步

```
1-3. 解析+校验+查用户
4. 验证密码 ★ — bcrypt.CompareHashAndPassword 从hash提取盐值重新哈希后对比
5. 生成 JWT ★ — Header.Payload.Signature 三段式，有效期24h
6. 返回 token + UserInfo
7. 创建 Session
```

#### GetCurrentUser（第228-245行）

```
套了 AuthMiddleware 中间件！中间件验完token，用户信息在context里
→ 直接从context取信息返回，不需要查数据库
```

**和 visit.go 的对比：**

| | visit.go | auth.go |
|---|----------|---------|
| 特殊操作 | 无 | bcrypt加密/验密 + JWT生成 |
| 中间件 | 无 | /api/auth/me 套了 AuthMiddleware |

**动手练习（必须写）：**

```go
// 练习1：模拟 Register/Login 核心逻辑（用map代替数据库）
var users = make(map[string]string)
func register(username, password string) string {
    if _, exists := users[username]; exists { return "用户名已存在" }
    users[username] = "hashed_" + password; return "注册成功"
}
func login(username, password string) string {
    hash, exists := users[username]
    if !exists || hash != "hashed_"+password { return "用户名或密码错误" }
    return "登录成功，token: jwt_" + username
}
```

**Day 9 过关标准：**
- [ ] Register 6步、Login 7步能口述
- [ ] 能说清 bcrypt 和 JWT 各自的作用
- [ ] 能解释为什么 PasswordHash 用 json:"-"

---

### Day 10（高强度）— 读 handler/focus.go（专注时间核心！）

**比 auth.go 更复杂的业务逻辑，涉及统计查询和数据聚合。**

#### 文件结构概览（330行，6个API）

| API | 方法 | 功能 |
|-----|------|------|
| `/api/focus/session` | POST | 创建专注记录 |
| `/api/focus/today` | GET | 获取今日统计（含标签占比） |
| `/api/focus/summary` | GET | 获取历史总览（每日统计） |
| `/api/focus/history` | GET | 获取某日详细记录 |
| `/api/focus/tags` | POST/GET | 标签管理 |

#### CreateFocusSession（第65-113行）— 创建专注记录

```
1. json.Decode 解析 body → CreateFocusSessionRequest
2. 参数校验（duration 60-14400秒，即1分钟到4小时）
3. 设置默认标签（"未分类"、默认颜色#6C5CE7）
4. 组装 StudySession 结构体（UserID从JWT context取）
5. INSERT 插入数据库
6. 返回创建的记录
```

#### GetTodayFocus（第119-167行）— 今日统计 ★复杂

```
1. 查询今日所有记录（Where user_id=? AND date=?）
2. 遍历计算：总秒数、按标签分组累加
3. 计算每个标签的百分比（Seconds/TotalSeconds*100）
4. 组装 TodayFocusResponse 返回
```

**关键代码模式：**
```go
tagMap := make(map[string]*model.TagSummary)
for _, s := range sessions {
    totalSeconds += int64(s.Duration)
    if ts, ok := tagMap[s.Tag]; ok {
        ts.Seconds += int64(s.Duration)  // 累加已有标签
    } else {
        tagMap[s.Tag] = &model.TagSummary{Tag: s.Tag, Color: s.TagColor}
        tagMap[s.Tag].Seconds += int64(s.Duration)  // 新建标签
    }
}
```

#### GetFocusSummary（第173-225行）— 历史总览 ★复杂

```
1. 查询历史总计：SUM(duration)、COUNT(*)
2. 查询每日统计（GROUP BY date）
3. 使用原生 SQL Rows 扫描，组装 DailyStatItem 列表
```

**重点：复杂 SQL 查询**
```go
database.Model(&model.StudySession{}).
    Select("date, SUM(duration) as total_seconds, COUNT(*) as session_count").
    Where("user_id = ? AND date >= DATE_SUB(CURDATE(), INTERVAL ? DAY)", userID, days).
    Group("date").
    Order("date DESC").
    Rows()
```

#### GetFocusHistory / CreateTag / GetTags（第231-330行）

- GetFocusHistory：简单查询 + 格式转换（time.Format("15:04:05")）
- CreateTag / GetTags：标准 CRUD，类似 todo.go

**动手练习（必须写）：**

```go
// 练习1：模拟标签占比计算
sessions := []struct{Tag string; Duration int}{
    {"Go开发", 3600}, {"算法", 1800}, {"Go开发", 1200},
}
// 计算：Go开发=4800秒(66.7%)，算法=1800秒(33.3%)

// 练习2：手写 GROUP BY 查询的 GORM 代码
// 查询每个标签的总时长，按时长降序

// 练习3：实现 formatDuration 函数（秒→"X小时Y分钟"）
```

**Day 10 过关标准：**
- [ ] 能说出 focus.go 6个API的作用
- [ ] 能解释 GetTodayFocus 的标签分组累加逻辑
- [ ] 能写出 GROUP BY + SUM + COUNT 的 GORM 查询
- [ ] 能解释 DATE_SUB(CURDATE(), INTERVAL ? DAY) 的作用

---

### Day 12（高强度）— 读 utils/ + middleware/ + session/

**认证模块的三个辅助包，每个都很短。**

**utils/hash.go（48行）：**
```go
HashPassword(password) → bcrypt.GenerateFromPassword，Cost=10，约100ms
CheckPassword(password, hash) → bcrypt.CompareHashAndPassword，从hash提取盐值再哈希对比
```

**utils/jwt.go（105行）：**
```go
JwtSecretKey = "tinyweb1-secret-key-2026"  // 硬编码，生产应从环境变量读
CustomClaims { UserID, Username, Role, jwt.RegisteredClaims }
GenerateToken(userID, username, role) → 设置过期24h → HS256签名 → 返回字符串
ValidateToken(tokenString) → 解析+验签名+检查过期 → 返回 CustomClaims
```

**middleware/auth.go（110行）：**
```go
AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    1. 从Header取 Authorization
    2. 提取 "Bearer xxx" 中的 token
    3. ValidateToken 验证 → 失败返回401
    4. context.WithValue 注入 userID/username/role
    5. next.ServeHTTP(w, r.WithContext(ctx))  // 放行
}
```

**session/memory.go（88行）：**
```go
// 用 sync.Map 存 SessionInfo{UserID, Username, Role, LoginTime, Token}
// Create / Get / Delete 三个方法
// 注意：内存存储，重启服务器会丢失
```

**动手练习（必须写）：**

```go
// 练习1：模拟中间件模式
func authMiddleware(next func(string)) func(string) {
    return func(token string) {
        if token == "" { fmt.Println("拒绝访问"); return }
        fmt.Println("验证通过"); next(token)
    }
}
protectedHandler := authMiddleware(getUserInfo)
protectedHandler("")           // 拒绝
protectedHandler("jwt_token")  // 通过
```

**Day 12 过关标准：**
- [ ] 能画出 JWT 的三段结构（Header.Payload.Signature）
- [ ] 能解释中间件的工作流程
- [ ] 能解释 context.WithValue 的作用

---

### Day 13（高强度）— 前后端联调走查

**目标：从前端第一行代码跟踪到数据库最后一行SQL。**

**画完整时序图：**

```
浏览器                    Go服务器                    MySQL
  │                         │                          │
  │ ① fetch POST /api/visit │                          │
  │========================▶│                          │
  │                         │ ② json.Decode            │
  │                         │ ③ getClientIP            │
  │                         │ ④ SELECT WHERE ip        │
  │                         │=========================▶│
  │                         │                  ⑤ 返回   │
  │                         │◀=========================│
  │                         │ ⑥ INSERT 或 UPDATE       │
  │                         │=========================▶│
  │                         │                  ⑦ 成功   │
  │                         │◀=========================│
  │                  ⑧ sendJSON 返回                   │
  │◀========================│                          │
  │ ⑨ .then() 收到响应      │                          │
  │                         │                          │
  │ ⑩ fetch GET /api/visit/stats                      │
  │========================▶│                          │
  │                         │ ⑪ SELECT SUM/COUNT/MAX   │
  │                         │=========================▶│
  │                         │                  ⑫ 返回   │
  │                         │◀=========================│
  │                  ⑬ sendJSON                        │
  │◀========================│                          │
  │ ⑭ renderStats() 更新DOM │                          │
```

**每个箭头标注：对应代码哪一行。**

**Day 13 过关标准：**
- [ ] 时序图覆盖全部14个步骤
- [ ] 每个步骤能定位到具体文件和行号
- [ ] 合上资料能口述整个过程

---

### Day 14（中强度）— 第2周总结 + 补漏

**自查清单：**

```
□ Go 变量声明（var / := / const）
□ Go 函数多返回值 + error 处理
□ struct 和 tag（特别是 json:"-" ）
□ ServeMux 路由机制
□ RecordVisit 四步流程
□ Register / Login 的步骤
□ Focus 6个API的作用
□ bcrypt 和 JWT 各自的作用
□ 中间件的工作流程
□ 每个 GORM 操作对应哪条 SQL
□ 前端 fetch → 后端 handler → MySQL 全程串起来
```

**有任何一项打 ❌ 的，今天专门补。**

---

# 第3周：数据库 + 认证精读 + 其他模块

## 周目标

> 理解 MySQL 实操，GORM 与 SQL 对应，快速看懂其他模块。

---

### Day 15（高强度）— SSH 进 MySQL 实操

```bash
ssh user@1.15.224.88
mysql -u root tinyweb1
```

逐条执行：

```sql
SHOW TABLES;
DESCRIBE visit_stats;           -- 对照 model.go
SELECT * FROM visit_stats;
SELECT visitor_ip, visit_count FROM visit_stats WHERE visit_count > 5;
SELECT SUM(visit_count) FROM visit_stats;    -- 对应 GetVisitStats
SELECT COUNT(*) FROM visit_stats;            -- 对应 GetVisitStats
SELECT MAX(last_visit_at) FROM visit_stats;  -- 对应 GetVisitStats
INSERT INTO visit_stats (...) VALUES ('127.0.0.1', 1, 'Test', 'Linux', NOW(), NOW());
UPDATE visit_stats SET visit_count = visit_count + 1 WHERE visitor_ip = '127.0.0.1';
DELETE FROM visit_stats WHERE visitor_ip = '127.0.0.1';
SHOW INDEX FROM visit_stats;     -- 看 unique_index 在哪列
DESCRIBE users;                  -- 对照 model.go User struct
SELECT id, username, role FROM users;
SELECT username, password_hash FROM users LIMIT 1;  -- 看 bcrypt 哈希长什么样
```

**Day 15 过关标准：**
- [ ] 上面SQL全部在服务器上执行过且理解结果

---

### Day 16（高强度）— GORM 与 SQL 对应

```go
database.Create(&newRecord)           ↔ INSERT INTO
database.Where().First(&existing)     ↔ SELECT WHERE LIMIT 1
database.Model().Updates(...)         ↔ UPDATE SET
database.Delete(&record)             ↔ DELETE WHERE
database.Select("SUM(...)").Scan(&x)  ↔ SELECT SUM(...)
database.Select("COUNT(*)").Scan(&x)  ↔ SELECT COUNT(*)
database.AutoMigrate(&Model{})        ↔ CREATE TABLE IF NOT EXISTS
```

**动手练习：**

```go
// 用 GORM 写出以下 SQL 对应的 Go 代码：
// 1. SELECT * FROM visit_stats WHERE browser = 'Chrome';
// 2. UPDATE visit_stats SET visit_count = 10 WHERE visitor_ip = '127.0.0.1';
// 3. DELETE FROM visit_stats WHERE id = 1;
```

**Day 16 过关标准：**
- [ ] 每个 GORM 方法都能说出对应的 SQL
- [ ] 练习跑通

---

### Day 17（高强度）— 读其他 handler

快速过 `todo.go`、`guestbook.go`、`setting.go`。模式和 `visit.go` 完全一样。

| 文件 | 核心函数 | 对应表 |
|------|---------|--------|
| todo.go | CreateTodo / GetTodos / UpdateTodo / DeleteTodo | todos |
| guestbook.go | CreateMessage / GetMessages | guestbook |
| setting.go | GetSettings / UpdateSettings | settings |

**Day 17 过关标准：**
- [ ] 能说出每个 handler 的 CRUD 函数名
- [ ] 能指出和 visit.go 的模式差异（如果有的话）

---

### Day 18（高强度）— 数据库设计分析 + 动手改功能

**思考：**

```
□ 为什么 visitor_ip 做 uniqueIndex？ → 同IP一条，多次更新count
□ 为什么 username 也要 uniqueIndex？ → 防止重复注册
□ DeletedAt 字段干嘛的？ → 软删除，标记时间不真删
□ 如果要加 city 字段，改哪些文件？ → model.go + handler + 前端
```

**动手改功能（选做一个）：**

- **选项 A（简单）**：给访问统计加城市字段（改 model.go + visit.go + index.html）
- **选项 B（中等）**：给留言板加分页数字显示
- **选项 C（中等）**：加"今日访问数"字段

**Day 18 过关标准：**
- [ ] 功能改动完成并部署验证
- [ ] 能说出改了哪些文件、每处改动的原因

---

### Day 19（高强度）— 认证全链路精读

**画认证流程时序图（三个流程）：**

**注册：** 前端 → POST /api/auth/register → Register → bcrypt加密 → INSERT → 返回

**登录：** 前端 → POST /api/auth/login → Login → bcrypt验密 → 生成JWT → 返回token → localStorage存储

**获取用户：** 前端 → GET /api/auth/me(Bearer token) → AuthMiddleware验token → GetCurrentUser(从context取) → 返回

**Day 19 过关标准：**
- [ ] 三个流程的每个步骤能口述
- [ ] 能解释 GetCurrentUser 不需要查数据库
- [ ] 能说出 localStorage 和 sessionStorage 的区别

---

### Day 20（中强度）— 读 fronted/flappy-bird.js（Flappy Bird游戏）

**纯前端游戏，理解游戏循环和Canvas绘图。**

#### 文件结构概览（386行）

```
核心机制：
├── requestAnimationFrame 游戏循环
├── Canvas 绘图（bird, pipe, background）
├── 物理系统：重力、跳跃速度
├── 碰撞检测：bird与pipe的矩形碰撞
└── 计分系统
```

**重点理解代码模式：**

```javascript
// 游戏状态机
const GameState = { START: 0, PLAYING: 1, GAME_OVER: 2 };

// 游戏循环
function gameLoop() {
    if (state === GameState.PLAYING) {
        update();  // 更新位置、检测碰撞
        draw();    // 重绘画布
    }
    requestAnimationFrame(gameLoop);
}

// Canvas 绘图示例
ctx.drawImage(birdImg, bird.x, bird.y, bird.width, bird.height);
```

**动手练习：**

```javascript
// 练习1：理解碰撞检测逻辑
function checkCollision(bird, pipe) {
    return bird.x < pipe.x + pipe.width &&
           bird.x + bird.width > pipe.x &&
           bird.y < pipe.y + pipe.height &&
           bird.y + bird.height > pipe.y;
}

// 练习2：实现简单的计时器（类似专注时间的倒计时）
let seconds = 25 * 60;  // 25分钟
timer = setInterval(() => {
    seconds--;
    if (seconds <= 0) clearInterval(timer);
}, 1000);
```

**Day 20 过关标准：**
- [ ] 能说出游戏循环的工作原理
- [ ] 能理解 Canvas 的基本绘图操作
- [ ] 能解释碰撞检测的矩形相交判断

---

# 第4周：复习深入 + 安全 + 前端其他模块

## 周目标

> 第二轮精读代码，理解安全和性能问题，读懂前端其他模块。

---

### Day 23（高强度）— 第二轮精读 main.go + visit.go

重新读，标注：
- 懂了的地方 → 绿色
- 还不懂 → 红色，重点攻克

**Day 23 过关标准：**
- [ ] 没有红色标记
- [ ] 能不看代码默写 RecordVisit 四步

---

### Day 24（高强度）— 第二轮精读 auth.go + index.html

重新读 auth.go 和 index.html（3787-3903行 + 3905-4183行）。

**Day 24 过关标准：**
- [ ] auth.go 没有红色标记
- [ ] 能不看代码口述 Login 七步
- [ ] 能说出 AuthManager 的主要方法名和作用

---

### Day 25（中强度）— 安全和性能分析

**安全问题：**

```
✅ SQL注入 → GORM参数化查询，安全
✅ 密码存储 → bcrypt加密
⚠️ XSS → AuthManager有escapeHtml()，其他模块没转义
⚠️ CSRF → 没有token保护
⚠️ JWT密钥 → 硬编码，应从环境变量读
⚠️ 限流 → 同一IP可无限刷visit/无限注册
⚠️ token刷新 → 24h过期但无刷新机制
```

**性能问题：**

```
⚠️ 每次访问都查DB → 可用缓存
⚠️ 前端POST完再GET → 可并行
✅ 连接池 → MaxOpenConns=10，小项目够了
```

**Day 21 过关标准：**
- [ ] 能说出3个以上安全隐患和修复思路
- [ ] 能说出2个以上性能优化方向

---

### Day 22（高强度）— 读 index.html 其他模块

```
□ 搜索 todo 相关 JS → 对应 todo.go
□ 搜索 guestbook 相关 JS → 对应 guestbook.go
□ 找到前端 fetch URL 和后端路由的对应关系
```

**Day 22 过关标准：**
- [ ] 能说出备忘录和留言板前端代码的位置
- [ ] 能找到每个 fetch 对应的后端 handler
- [ ] 能说出 AuthManager 和 IIFE 的结构区别

---

# 第5周：架构图 + 最终验收 + AI复述

## 周目标

> 不看代码画出完整架构图，通过AI复述验收学习成果。

---

### Day 22（高强度）— 画完整架构图

**不用看代码，凭记忆画。画完再对照代码修正。**

架构图应包含：

```
1. 前端层
   ├── HTML结构 / CSS样式 / JS逻辑
   └── JS模块：访问统计(IIFE) / 认证(AuthManager) / 备忘录 / 留言板 / 设置

2. 后端层
   ├── main.go（入口+路由）
   ├── handler/（visit / auth / todo / guestbook / setting / helpers）
   ├── middleware/（auth.go JWT中间件）
   ├── session/（memory.go 内存Session）
   └── utils/（jwt.go / hash.go）

3. 数据库层
   ├── users / visit_stats / todos / guestbook / settings

4. 前后端连接（每个API的完整路径）
   ├── POST /api/auth/register → Register
   ├── POST /api/auth/login → Login
   ├── GET /api/auth/me → AuthMiddleware(GetCurrentUser)
   ├── POST/GET /api/visit → RecordVisit / GetVisitStats
   └── CRUD /api/todos → TodoHandlers 等
```

**Day 22 过关标准：**
- [ ] 不看代码画出来，对照修正不超过3处错误

---

### Day 23-24（高强度）— 最终验收

**逐项检查：**

```
□ 能画出项目完整架构图（前端/后端/数据库三层）
□ 能说出任意一个 API 的完整调用链（前端→路由→handler→DB→返回→渲染）
□ 能读懂 main.go 每一行
□ 能读懂 handler/visit.go 每一行
□ 能读懂 handler/auth.go 每一行
□ 能读懂 index.html 第3787-3903行 每一行
□ 能读懂 index.html 第3905-4183行 AuthManager
□ 能说出 GORM 每个方法对应的 SQL
□ 能解释 IIFE、sessionStorage、fetch、JSON 的作用
□ 能解释 bcrypt、JWT、中间件、context 的作用
□ 至少做过一次功能改动并部署验证
```

**有不通过的项，Day 24 专门补。**

---

### Day 25（中强度）— AI 复述验收

**给AI文字复述全流程，让AI评分。**

复述内容：

```
1. 用户打开网页后发生了什么（从前端到数据库再回来）
2. IIFE 是什么，为什么用
3. fetch 怎么发请求的
4. 后端路由怎么匹配的
5. GORM 怎么操作数据库的（每个方法对应什么SQL）
6. sessionStorage 怎么防刷新重复计数的
7. 用户注册时密码怎么加密存储的（bcrypt流程）
8. 用户登录时token怎么生成和验证的（JWT流程）
9. 中间件怎么拦截请求验token的
10. 给AI复述后让AI逐项评分（满分10分），低于8分的回去补
```

**Day 25 过关标准：**
- [ ] AI评分每项 ≥ 8分
- [ ] 低于8分的已经补完

---

# 附录（精简版）

## GORM ↔ SQL 速查

```go
database.Where("x = ?", v).First(&r)   // SELECT * FROM t WHERE x=v LIMIT 1
database.Create(&r)                     // INSERT INTO
database.Model(&r).Updates(map)         // UPDATE SET
database.Delete(&r)                     // DELETE WHERE (软删除则UPDATE deleted_at)
database.Select("SUM(x)").Scan(&v)      // SELECT SUM(x)
database.Select("COUNT(*)").Scan(&v)    // SELECT COUNT(*)
database.AutoMigrate(&M{})              // CREATE TABLE IF NOT EXISTS
```

## GET vs POST

| | GET | POST |
|---|-----|------|
| 用途 | 要数据 | 传数据 |
| body | 没有 | 有 |
| 你的代码 | /api/visit/stats | /api/visit |

## localStorage vs sessionStorage

| | localStorage | sessionStorage |
|---|---|---|
| 生命周期 | 永久 | 标签页关闭清空 |
| 你的代码用在哪 | 存JWT token | 存visit_recorded标记 |

---

*本指南由 CodeBuddy AI 生成，学习过程中有任何问题随时提问！*