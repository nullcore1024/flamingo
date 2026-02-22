# FlamingoClient 实现原理详解

## 一、项目概述

FlamingoClient 是 Flamingo IM 即时通讯软件的 Windows PC 客户端，使用 C++ 和 WTL 框架开发。本文档详细介绍其实现原理、核心技术、网络协议以及开发流程。

## 二、核心技术栈

### 2.1 开发语言与框架
- **编程语言**: C++ (C++11 标准)
- **UI 框架**: WTL (Windows Template Library) 9.0
- **开发工具**: Visual Studio 2019

### 2.2 核心库
- **JSON 解析**: JsonCpp 1.9.0 (用于协议数据序列化/反序列化)
- **压缩库**: zlib 1.2.11 (用于数据包压缩)
- **数据库**: SQLite 3.7.17 (本地聊天记录存储)

### 2.3 网络与多线程
- **网络通信**: Windows Socket API
- **多线程**: C++11 std::thread, std::mutex, std::condition_variable
- **线程模型**: 生产者-消费者模式

## 三、整体架构

### 3.1 架构图
```
┌─────────────────────────────────────────────────────────────────┐
│                          UI 层                                   │
│  (LoginDlg, MainDlg, FindFriendDlg, BuddyChatDlg, ...)        │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      业务逻辑层 (CFlamingoClient)               │
│  用户管理、消息处理、状态管理、文件/图片传输                      │
└─────────────────────────────────────────────────────────────────┘
                              │
          ┌───────────────────┴───────────────────┐
          ▼                                       ▼
┌──────────────────────┐               ┌──────────────────────┐
│  发送线程            │               │  接收线程            │
│  (CSendMsgThread)    │               │  (CRecvMsgThread)    │
└──────────────────────┘               └──────────────────────┘
          │                                       │
          └───────────────────┬───────────────────┘
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                  网络层 (CIUSocket, IUSocket)                   │
│  TCP 连接管理、数据包收发、心跳保活、断线重连                      │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      服务器端                                     │
└─────────────────────────────────────────────────────────────────┘
```

### 3.2 核心模块说明

#### 3.2.1 CFlamingoClient (主控制类)
- **位置**: `Source/FlamingoClient.h/cpp`
- **职责**: 整个客户端的核心控制类，协调各个模块
- **主要功能**:
  - 用户登录/注销管理
  - 好友列表、群组列表管理
  - 消息收发调度
  - 网络状态监控

#### 3.2.2 CIUSocket (网络层)
- **位置**: `Source/net/IUSocket.h/cpp`
- **职责**: 底层 TCP 网络通信
- **主要功能**:
  - 连接服务器 (聊天服务器、文件服务器、图片服务器)
  - 数据包发送与接收
  - 心跳保活
  - 断线自动重连
  - 数据压缩/解压缩

#### 3.2.3 CSendMsgThread (发送线程)
- **位置**: `Source/SendMsgThread.h/cpp`
- **职责**: 异步发送消息队列
- **工作模式**: 生产者-消费者模式
- **主要功能**:
  - 接收业务层的发送请求
  - 序列化协议数据
  - 通过网络层发送
  - 维护序列号 (seq)

#### 3.2.4 CRecvMsgThread (接收线程)
- **位置**: `Source/RecvMsgThread.h/cpp`
- **职责**: 异步接收和处理服务器回包
- **主要功能**:
  - 解析服务器回包
  - 分发到对应的处理函数
  - 通过 Windows 消息通知 UI 层更新

## 四、网络协议详解

### 4.1 协议层次结构

```
┌─────────────────────────────────────────────────────────┐
│              应用层数据 (JSON 格式)                       │
│   例如: {"type": 1, "username": "testuser"}            │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│         协议包体 (BinaryStreamWriter 序列化)             │
│  [cmd(4B)][seq(4B)][json_data_len(4B)][json_data]    │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│          协议头 (msg 或 chat_msg_header)                │
│  [compressflag(1B)][originsize(4B)][compresssize(4B)]  │
│  [reserved(16B)]                                         │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                 TCP 传输层                               │
└─────────────────────────────────────────────────────────┘
```

### 4.2 协议头结构

#### 4.2.1 协议头定义 (`Source/net/Msg.h`)
```cpp
#pragma pack(push, 1)
struct msg  // 或 chat_msg_header
{
    char     compressflag;     // 压缩标志位: 0=未压缩, 1=已压缩
    int32_t  originsize;       // 压缩前的大小 (字节)
    int32_t  compresssize;     // 压缩后的大小 (字节)
    char     reserved[16];     // 保留字段
};
#pragma pack(pop)
```

#### 4.2.2 协议头字段说明
| 字段名 | 类型 | 长度 | 说明 |
|--------|------|------|------|
| compressflag | char | 1字节 | 压缩标志: 0=未压缩, 1=已压缩 |
| originsize | int32_t | 4字节 | 原始数据大小 (小端序) |
| compresssize | int32_t | 4字节 | 压缩后数据大小 (小端序) |
| reserved | char[16] | 16字节 | 保留字段，用于扩展 |

### 4.3 协议包体结构

使用 BinaryStreamWriter 序列化，格式为：

| 字段 | 类型 | 长度 | 说明 |
|------|------|------|------|
| cmd | int32_t | 4字节 | 命令类型 (msg_type 枚举) |
| seq | int32_t | 4字节 | 序列号，用于匹配请求和响应 |
| json_data_len | int32_t | 4字节 | JSON 数据长度 |
| json_data | char[] | 可变 | JSON 格式的业务数据 |

### 4.4 命令类型 (msg_type 枚举)

| 命令值 | 命令名 | 说明 |
|--------|--------|------|
| 1000 | msg_type_heartbeat | 心跳包 |
| 1001 | msg_type_register | 用户注册 |
| 1002 | msg_type_login | 用户登录 |
| 1003 | msg_type_getofriendlist | 获取好友列表 |
| 1004 | msg_type_finduser | 查找用户/群组 |
| 1005 | msg_type_operatefriend | 好友操作 (添加/删除/同意) |
| 1006 | msg_type_userstatuschange | 用户状态变化通知 |
| 1007 | msg_type_updateuserinfo | 更新用户信息 |
| 1008 | msg_type_modifypassword | 修改密码 |
| 1009 | msg_type_creategroup | 创建群组 |
| 1010 | msg_type_getgroupmembers | 获取群成员列表 |
| 1100 | msg_type_chat | 单聊消息 |
| 1101 | msg_type_multichat | 群聊消息 |
| 1102 | msg_type_kickuser | 踢人 |
| 1103 | msg_type_remotedesktop | 远程桌面 |
| 1104 | msg_type_updateteaminfo | 更新分组信息 |
| 1105 | msg_type_modifyfriendmarkname | 修改好友备注 |
| 1106 | msg_type_movefriendtootherteam | 移动好友分组 |

### 4.5 典型协议示例

#### 4.5.1 查找用户协议 (msg_type_finduser = 1004)

**请求包**:
```
cmd = 1004, seq = 0, {"type": 1, "username": "zhangyl"}
```

**响应包**:
```
cmd = 1004, seq = 0, {"code": 0, "msg": "ok", "userinfo": [{"userid": 2, "username": "qqq", "nickname": "qqq123", "facetype": 0}]}
```

**字段说明**:
- type: 查找类型，1=查找用户，2=查找群组
- username: 要查找的用户名
- code: 错误码，0=成功
- userinfo: 查找到的用户信息列表

#### 4.5.2 登录协议 (msg_type_login = 1002)

**请求包**:
```
cmd = 1002, seq = 0, {"username": "13917043329", "password": "123", "clienttype": 1, "status": 1}
```

**响应包**:
```
cmd = 1002, seq = 0, {
    "code": 0, 
    "msg": "ok", 
    "userid": 8, 
    "username": "13917043320", 
    "nickname": "zhangyl",
    "facetype": 0, 
    "customface": "文件md5", 
    "gender": 0, 
    "birthday": 19891208, 
    "signature": "坚持就是胜利",
    "address": "上海市浦东新区3261号", 
    "phonenumber": "021-389456", 
    "mail": "balloonwj@qq.com"
}
```

### 4.6 错误码 (error_code 枚举)

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 1 | 未知错误 |
| 2 | 用户未登录 |
| 100 | 注册失败 |
| 101 | 已注册 |
| 102 | 未注册 |
| 103 | 密码错误 |
| 104 | 更新用户信息失败 |
| 105 | 修改密码失败 |
| 106 | 创建群失败 |
| 107 | 客户端版本太旧 |

## 五、完整开发流程：新增一个界面与协议

下面以"查找好友"功能为例，详细说明从界面输入到网络通信再到数据展示的完整流程。

### 5.1 完整流程图

```
┌─────────────────────────────────────────────────────────────────────┐
│  1. 用户在界面输入用户名，点击"查找"按钮                             │
│     (FindFriendDlg::OnAddFriend)                                     │
│     - 获取输入框内容 m_edtAddId.GetWindowText()                     │
│     - 验证输入长度 1-32 字符                                         │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│  2. 调用 CFlamingoClient::FindFriend()                               │
│     - Unicode → UTF-8 编码转换 (EncodeUtil::UnicodeToUtf8)         │
│     - 创建 CFindFriendRequest 对象                                   │
│     - 保存回调窗口句柄 SetFindFriendWindow(hReflectionWnd)          │
│     - 将请求添加到发送队列 AddItem(pRequest)                         │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│  3. CSendMsgThread 从队列取出请求                                     │
│     (Run() 生产者-消费者循环)                                         │
│     - std::condition_variable 等待数据                                │
│     - m_listItems.pop_front() 取出数据                                │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│  4. HandleFindFriendMessage() 处理请求                                │
│     - 构建 JSON 数据: {"type": 1, "username": "xxx"}                │
│     - 使用 BinaryStreamWriter 序列化包体:                             │
│       [cmd=1004(4B)][seq(4B)][json_len(4B)][json_data]             │
│     - 调用 CIUSocket::Send() 发送                                     │
│     - m_seq++ 序列号递增                                               │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│  5. CIUSocket::Send() 发送数据                                        │
│     - 将数据添加到 m_strSendBuf 队列                                  │
│     - 加锁保护 m_mtSendBuf                                            │
│     - 通知发送线程 m_cvSendBuf.notify_one()                         │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│  6. CIUSocket 发送线程 (SendThreadProc) 实际发送                      │
│     - 添加协议头 (struct msg)                                         │
│       [compressflag(1B)][originsize(4B)][compresssize(4B)]          │
│       [reserved(16B)]                                                  │
│     - 可选: 使用 zlib 压缩 (compressflag=1)                           │
│     - 通过 Socket 发送到服务器                                          │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│  7. 服务器处理请求，返回响应                                          │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│  8. CIUSocket 接收线程 (RecvThreadProc) 接收数据                     │
│     - 从 Socket 接收数据到 m_strRecvBuf                               │
│     - 解析协议头 (struct msg)                                         │
│     - 解压缩 (如果 compressflag=1)                                    │
│     - 调用 DecodePackages() 分包处理粘包                              │
│     - 将完整包体传递给 CRecvMsgThread                                │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│  9. CRecvMsgThread 从队列取出包 (生产者-消费者)                      │
│     (Run() 循环)                                                      │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│  10. HandleMessage() 根据 cmd 分发                                    │
│     - 使用 BinaryStreamReader 解析包体                                │
│     - cmd=1004 → HandleFindFriendMessage(strMsg)                     │
│     - 使用 JsonCpp 解析 JSON 响应数据                                 │
│     - 创建 CFindFriendResult 对象                                    │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│  11. 通过 PostMessage 发送 Windows 消息到 UI 层                      │
│     PostMessage(m_hProxyWnd, FMG_MSG_FINDFREIND, 0, (LPARAM)pResult)│
│     (FMG_MSG_FINDFREIND = WM_USER + 405)                             │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│  12. FindFriendDlg::OnFindFriendResult() 处理消息                    │
│     - 从 lParam 获取 CFindFriendResult*                              │
│     - UTF-8 → Unicode 转换 (EncodeUtil::Utf8ToUnicode)               │
│     - 根据 m_nResultCode 判断结果:                                    │
│       - FIND_FRIEND_NOT_FOUND: 显示"没有找到..."                     │
│       - FIND_FRIEND_FOUND: 弹出用户信息对话框 CUserSnapInfoDlg        │
│     - delete pResult 释放内存                                         │
└─────────────────────────────────────────────────────────────────────┘
```

### 5.2 详细步骤实现

#### 步骤 1: 创建 UI 界面 (FindFriendDlg)

**文件**: `Source/FindFriendDlg.h`

```cpp
class CFindFriendDlg : public CDialogImpl<CFindFriendDlg>, public CMessageFilter
{
public:
    enum { IDD = IDD_FINDFRIENDDLG };

    BEGIN_MSG_MAP_EX(CFindFriendDlg)
        MSG_WM_INITDIALOG(OnInitDialog)
        COMMAND_ID_HANDLER_EX(IDC_BTN_ADD, OnAddFriend)
        MESSAGE_HANDLER(FMG_MSG_FINDFREIND, OnFindFriendResult)
    END_MSG_MAP()

    void OnAddFriend(UINT uNotifyCode, int nID, CWindow wndCtl);
    LRESULT OnFindFriendResult(UINT uMsg, WPARAM wParam, LPARAM lParam, BOOL& bHandled);

private:
    CSkinEdit    m_edtAddId;      // 输入框
    CSkinButton  m_btnAdd;         // 查找按钮
    CFlamingoClient* m_pFMGClient;
};
```

**文件**: `Source/FindFriendDlg.cpp` - OnAddFriend()

```cpp
void CFindFriendDlg::OnAddFriend(UINT uNotifyCode, int nID, CWindow wndCtl)
{
    // 1. 获取输入框数据
    CString strAccountToAdd;
    m_edtAddId.GetWindowText(strAccountToAdd);
    strAccountToAdd.Trim();

    // 2. 验证输入
    if(strAccountToAdd.GetLength() > 32 || strAccountToAdd.GetLength() <= 0)
    {
        m_staticAddInfo.SetWindowText(_T("请输入账号（长度在1-30个字符之间）"));
        return;
    }

    // 3. 调用业务层
    long nType = m_btnFindTypeSingle.GetCheck() ? 1 : 2;
    m_pFMGClient->FindFriend(strAccountToAdd, nType, m_hWnd);
}
```

#### 步骤 2: 业务层处理 (CFlamingoClient::FindFriend)

**文件**: `Source/FlamingoClient.h/cpp`

```cpp
BOOL CFlamingoClient::FindFriend(PCTSTR pszAccountName, long nType, HWND hReflectionWnd)
{
    // 1. 创建请求对象
    CFindFriendRequest* pRequest = new CFindFriendRequest();
    
    // 2. 填充数据
    char szAccountName[64] = {0};
    EncodeUtil::UnicodeToUtf8(pszAccountName, szAccountName, ARRAYSIZE(szAccountName));
    strcpy_s(pRequest->m_szAccountName, ARRAYSIZE(pRequest->m_szAccountName), szAccountName);
    pRequest->m_nType = nType;

    // 3. 保存回调窗口句柄
    SetFindFriendWindow(hReflectionWnd);

    // 4. 添加到发送队列
    m_SendMsgThread.AddItem(pRequest);

    return TRUE;
}
```

#### 步骤 3: 定义协议数据结构 (IUProtocolData.h)

**文件**: `Source/net/IUProtocolData.h`

```cpp
// 1. 在 NET_DATA_TYPE 枚举中添加新类型
enum NET_DATA_TYPE
{
    NET_DATA_UNKNOWN,
    // ... 已有类型
    NET_DATA_FIND_FRIEND,  // 查找好友
    // ...
};

// 2. 定义请求类
class CFindFriendRequest : public CNetData
{
public:
    CFindFriendRequest();
    ~CFindFriendRequest();

public:
    char m_szAccountName[64];  // 要查找的用户名 (UTF-8编码)
    long m_nType;               // 类型: 1=查找用户, 2=查找群组
};

// 3. 定义响应类
class CFindFriendResult : public CNetData
{
public:
    CFindFriendResult();
    ~CFindFriendResult();

public:
    long m_nResultCode;         // 结果码: FIND_FRIEND_NOT_FOUND, FIND_FRIEND_FOUND, FIND_FRIEND_FAILED
    UINT m_uAccountID;          // 用户/群组ID
    char m_szAccountName[64];   // 用户名 (UTF-8编码)
    char m_szNickName[64];      // 昵称 (UTF-8编码)
};

// 4. 结果码枚举
enum ADD_FRIEND_RESULT
{
    FIND_FRIEND_NOT_FOUND,  // 未找到
    FIND_FRIEND_FOUND,      // 找到了
    FIND_FRIEND_FAILED,     // 查找失败
    
    ADD_FRIEND_SUCCESS,
    ADD_FRIEND_FAILED
};
```

#### 步骤 4: 发送线程处理 (CSendMsgThread::HandleFindFriendMessage)

**文件**: `Source/SendMsgThread.cpp`

```cpp
void CSendMsgThread::HandleFindFriendMessage(const CFindFriendRequest* pFindFriendRequest)
{
    if (pFindFriendRequest == NULL)
        return;

    // 1. 构建 JSON 业务数据
    char szData[64] = { 0 };
    sprintf_s(szData, 64, "{\"type\": %d, \"username\": \"%s\"}", 
              pFindFriendRequest->m_nType, 
              pFindFriendRequest->m_szAccountName);

    // 2. 使用 BinaryStreamWriter 序列化协议包体
    std::string outbuf;
    net::BinaryStreamWriter writeStream(&outbuf);
    
    // 写入命令类型 (cmd): msg_type_finduser = 1004
    writeStream.WriteInt32(msg_type_finduser);
    
    // 写入序列号 (seq)
    writeStream.WriteInt32(m_seq);
    
    // 写入 JSON 数据
    writeStream.WriteCString(szData, strlen(szData));
    
    writeStream.Flush();

    // 3. 发送到网络层 (CIUSocket)
    LOG_INFO("Request to find friend, type=%d, accountName=%s", 
             pFindFriendRequest->m_nType, 
             pFindFriendRequest->m_szAccountName);

    CIUSocket::GetInstance().Send(outbuf);
}
```

#### 步骤 5: 接收线程处理 (CRecvMsgThread::HandleFindFriendMessage)

**文件**: `Source/RecvMsgThread.cpp`

```cpp
BOOL CRecvMsgThread::HandleFindFriendMessage(const std::string& strMsg)
{
    // 1. 使用 JsonCpp 解析 JSON 响应数据
    Json::Reader JsonReader;
    Json::Value JsonRoot;
    
    if (!JsonReader.parse(strMsg, JsonRoot))
    {
        return FALSE;
    }

    // 2. 检查响应状态: code == 0 表示成功，且 userinfo 是数组
    if (JsonRoot["code"].isNull() || JsonRoot["code"].asInt() != 0 || !JsonRoot["userinfo"].isArray())
        return FALSE;

    // 3. 创建结果对象
    CFindFriendResult* pFindFriendResult = new CFindFriendResult();
    
    // 4. 根据 userinfo 数组大小判断结果
    if (JsonRoot["userinfo"].size() == 0)
    {
        // 未找到
        pFindFriendResult->m_nResultCode = FIND_FRIEND_NOT_FOUND;
    }
    else
    {
        // 找到用户/群组
        pFindFriendResult->m_nResultCode = FIND_FRIEND_FOUND;
        
        // 提取第一个用户信息 (userinfo[0])
        pFindFriendResult->m_uAccountID = JsonRoot["userinfo"][(UINT)0]["userid"].asInt();
        strcpy_s(pFindFriendResult->m_szAccountName, ARRAYSIZE(pFindFriendResult->m_szAccountName), 
                 JsonRoot["userinfo"][(UINT)0]["username"].asCString());
        strcpy_s(pFindFriendResult->m_szNickName, ARRAYSIZE(pFindFriendResult->m_szNickName), 
                 JsonRoot["userinfo"][(UINT)0]["nickname"].asCString());
    }

    // 5. 通过 Windows 消息机制将结果发送到 UI 层
    // 注意: 这里通过 m_hProxyWnd 发送，最终会路由到 FindFriendDlg
    ::PostMessage(m_hProxyWnd, FMG_MSG_FINDFREIND, 0, (LPARAM)pFindFriendResult);
    
    return TRUE;
}
```

#### 步骤 6: 定义 Windows 消息 (UserSessionData.h)

**文件**: `Source/UserSessionData.h`

```cpp
// Windows 自定义消息定义
#define FMG_MSG_FINDFREIND   WM_USER + 405  // 查找好友结果消息
```

#### 步骤 7: UI 层响应 (FindFriendDlg::OnFindFriendResult)

**文件**: `Source/FindFriendDlg.h` 和 `Source/FindFriendDlg.cpp`

**头文件 (FindFriendDlg.h)**:
```cpp
class CFindFriendDlg : public CDialogImpl<CFindFriendDlg>, public CMessageFilter
{
public:
    enum { IDD = IDD_FINDFRIENDDLG };

    BEGIN_MSG_MAP_EX(CFindFriendDlg)
        MSG_WM_INITDIALOG(OnInitDialog)
        COMMAND_ID_HANDLER_EX(IDC_BTN_ADD, OnAddFriend)
        // 注册消息处理函数
        MESSAGE_HANDLER(FMG_MSG_FINDFREIND, OnFindFriendResult)
    END_MSG_MAP()

    // 消息处理函数声明
    LRESULT OnFindFriendResult(UINT uMsg, WPARAM wParam, LPARAM lParam, BOOL& bHandled);
    // ...
};
```

**实现文件 (FindFriendDlg.cpp)**:
```cpp
LRESULT CFindFriendDlg::OnFindFriendResult(UINT uMsg, WPARAM wParam, LPARAM lParam, BOOL& bHandled)
{
    // 1. 从 lParam 获取结果对象指针
    CFindFriendResult* pResult = (CFindFriendResult*)lParam;
    if (pResult == NULL)
        return 0;

    // 2. 根据结果码更新 UI
    long nResultCode = pResult->m_nResultCode;

    if (nResultCode == FIND_FRIEND_NOT_FOUND)
    {
        m_staticAddInfo.SetWindowText(_T("没有找到该用户或群组"));
    }
    else if (nResultCode == FIND_FRIEND_FAILED)
    {
        m_staticAddInfo.SetWindowText(_T("查找失败，请稍后再试"));
    }
    else if (nResultCode == FIND_FRIEND_FOUND)
    {
        m_staticAddInfo.SetWindowText(_T(""));
        
        // 3. 弹出用户信息对话框
        CUserSnapInfoDlg userSnapInfoDlg;
        
        // 判断是用户还是群组 (群组ID通常是特殊值)
        BOOL bGroup = IsGroupTarget(pResult->m_uAccountID);
        
        userSnapInfoDlg.SetAccountID(pResult->m_uAccountID);
        
        // 将 UTF-8 转换为 Unicode (Windows UI 使用 Unicode)
        TCHAR szAccount[32] = {0};
        EncodeUtil::Utf8ToUnicode(pResult->m_szAccountName, szAccount, ARRAYSIZE(szAccount));
        
        TCHAR szNickName[32] = {0};
        EncodeUtil::Utf8ToUnicode(pResult->m_szNickName, szNickName, ARRAYSIZE(szNickName));
        
        if (bGroup)
        {
            _stprintf_s(szInfo, ARRAYSIZE(szInfo), _T("群账号：%s"), szAccount);
            _stprintf_s(szInfo, ARRAYSIZE(szInfo), _T("群名称：%s"), szNickName);
            userSnapInfoDlg.SetOKButtonText(_T("加入群"));
        }
        else
        {
            _stprintf_s(szInfo, ARRAYSIZE(szInfo), _T("账号：%s"), szAccount);
            _stprintf_s(szInfo, ARRAYSIZE(szInfo), _T("昵称：%s"), szNickName);
            userSnapInfoDlg.SetOKButtonText(_T("加为好友"));
        }
        
        userSnapInfoDlg.SetAccountName(szAccount);
        userSnapInfoDlg.SetNickName(szNickName);
        
        // 4. 显示对话框
        if (userSnapInfoDlg.DoModal(m_hWnd, NULL) == IDOK)
        {
            // 如果用户点击"添加好友"或"加入群"
            m_staticAddInfo.SetWindowText(_T("您的请求已经发送，等待对方回应"));
            m_pFMGClient->AddFriend(pResult->m_uAccountID);
        }
    }

    // 5. 释放结果对象 (重要！避免内存泄漏)
    delete pResult;

    // 重新启用按钮
    m_btnAdd.EnableWindow(TRUE);
    m_edtAddId.EnableWindow(TRUE);
        
    return (LRESULT)1;
}
```

### 5.3 关键技术点

#### 5.3.1 BinaryStreamWriter/Reader 序列化

**文件**: `Source/net/ProtocolStream.h/cpp`

这是协议序列化的核心类，用于将数据打包成二进制流：

```cpp
// 写入示例
std::string outbuf;
net::BinaryStreamWriter writeStream(&outbuf);
writeStream.WriteInt32(cmd);      // 4字节命令
writeStream.WriteInt32(seq);      // 4字节序列号
writeStream.WriteString(json_str); // 字符串 (4字节长度 + 数据)
writeStream.Flush();

// 读取示例
net::BinaryStreamReader readStream(data_ptr, data_len);
int32_t cmd;
readStream.ReadInt32(cmd);
int32_t seq;
readStream.ReadInt32(seq);
std::string json_str;
readStream.ReadString(&json_str, max_len, out_len);
```

#### 5.3.2 线程间通信 - Windows 消息机制

FlamingoClient 使用 Windows 消息机制在工作线程和 UI 线程之间通信：

```cpp
// 1. 在 RecvMsgThread 中发送消息
::PostMessage(m_hProxyWnd, FMG_MSG_FINDFREIND, 0, (LPARAM)pResult);

// 2. 在 UI 类的消息映射中处理
BEGIN_MSG_MAP_EX(CFindFriendDlg)
    MESSAGE_HANDLER(FMG_MSG_FINDFREIND, OnFindFriendResult)
END_MSG_MAP()

// 3. 消息处理函数
LRESULT OnFindFriendResult(UINT uMsg, WPARAM wParam, LPARAM lParam, BOOL& bHandled);
```

#### 5.3.3 生产者-消费者模式

SendMsgThread 和 RecvMsgThread 都使用生产者-消费者模式：

```cpp
// 生产者 (主线程)
void AddItem(CNetData* pItem)
{
    std::lock_guard<std::mutex> guard(m_mtItems);
    m_listItems.push_back(pItem);
    m_cvItems.notify_one();  // 通知消费者
}

// 消费者 (工作线程)
void Run()
{
    while (!m_bStop)
    {
        CNetData* lpMsg;
        {
            std::unique_lock<std::mutex> guard(m_mtItems);
            while (m_listItems.empty())
            {
                if (m_bStop) return;
                m_cvItems.wait(guard);  // 等待数据
            }
            lpMsg = m_listItems.front();
            m_listItems.pop_front();
        }
        HandleItem(lpMsg);  // 处理数据
    }
}
```

## 六、目录结构

```
flamingoclient/
├── Bin/                          # 编译输出目录
│   ├── Face/                     # 表情资源
│   ├── Image/                    # 图片资源
│   ├── Skins/                    # 皮肤资源
│   └── config/                   # 配置文件
├── Source/                       # 源代码目录
│   ├── net/                      # 网络层
│   │   ├── IUSocket.h/cpp        # Socket 封装
│   │   ├── IUProtocolData.h/cpp  # 协议数据结构
│   │   ├── Msg.h                  # 协议定义
│   │   └── ProtocolStream.h/cpp  # 二进制流序列化
│   ├── SkinLib/                  # UI 皮肤库
│   ├── jsoncpp-1.9.0/            # JSON 库
│   ├── zlib1.2.11/                # 压缩库
│   ├── wtl9.0/                    # WTL 框架
│   ├── FlamingoClient.h/cpp       # 主控制类
│   ├── SendMsgThread.h/cpp        # 发送线程
│   ├── RecvMsgThread.h/cpp        # 接收线程
│   ├── FindFriendDlg.h/cpp        # 查找好友对话框
│   ├── LoginDlg.h/cpp             # 登录对话框
│   ├── MainDlg.h/cpp              # 主窗口
│   ├── BuddyChatDlg.h/cpp         # 单聊对话框
│   └── ...
└── CatchScreen/                   # 截图工具
```

## 七、开发注意事项

### 7.1 编码问题
- 网络传输使用 **UTF-8** 编码
- Windows UI 使用 **Unicode (UTF-16)**
- 使用 `EncodeUtil::UnicodeToUtf8()` 和 `EncodeUtil::Utf8ToUnicode()` 进行转换

### 7.2 线程安全
- 所有对共享数据的访问都需要加锁 (`std::mutex`)
- UI 更新必须在 UI 线程进行，通过 `PostMessage` 通知

### 7.3 内存管理
- 网络请求对象使用 `new` 创建，在发送/接收线程中 `delete`
- 注意避免内存泄漏和重复释放

### 7.4 协议扩展
添加新协议时，需要按以下步骤操作：

1. **在 `msg_type` 枚举中添加命令码** (`Source/net/Msg.h`)
   ```cpp
   enum msg_type
   {
       // ...
       msg_type_mynewcommand = 2000,  // 新增命令
   };
   ```

2. **在 `NET_DATA_TYPE` 中添加数据类型** (`Source/net/IUProtocolData.h`)
   ```cpp
   enum NET_DATA_TYPE
   {
       // ...
       NET_DATA_MY_NEW_DATA,  // 新增数据类型
   };
   ```

3. **定义请求和响应数据结构** (`Source/net/IUProtocolData.h`)
   ```cpp
   class CMyNewRequest : public CNetData
   {
   public:
       CMyNewRequest();
       ~CMyNewRequest();
   public:
       // 请求字段...
   };
   
   class CMyNewResult : public CNetData
   {
   public:
       CMyNewResult();
       ~CMyNewResult();
   public:
       // 响应字段...
   };
   ```

4. **定义 Windows 消息** (`Source/UserSessionData.h`)
   ```cpp
   #define FMG_MSG_MYNEWCOMMAND   WM_USER + 500
   ```

5. **在 `SendMsgThread` 中添加处理函数** (`Source/SendMsgThread.cpp`)
   - 在 `HandleItem()` 中添加 case 分支
   - 实现 `HandleMyNewMessage()` 函数

6. **在 `RecvMsgThread` 中添加解析函数** (`Source/RecvMsgThread.cpp`)
   - 在 `HandleMessage()` 中添加 case 分支
   - 实现 `HandleMyNewMessage()` 函数

7. **在 UI 层处理消息**
   - 在消息映射中注册 `MESSAGE_HANDLER(FMG_MSG_MYNEWCOMMAND, OnMyNewResult)`
   - 实现 `OnMyNewResult()` 函数更新 UI

### 7.5 CIUSocket 网络层详解

**文件**: `Source/net/IUSocket.h/cpp`

CIUSocket 是整个网络通信的核心，采用单例模式设计：

```cpp
class CIUSocket
{
public:
    static CIUSocket& GetInstance();  // 单例
    
    // 连接管理
    bool Connect(int timeout = 3);
    bool Reconnect(int timeout = 3);
    void Close();
    
    // 发送数据
    void Send(const std::string& strBuffer);
    
    // 三个独立的Socket连接
    SOCKET m_hSocket;        // 主聊天Socket
    SOCKET m_hFileSocket;     // 文件传输Socket
    SOCKET m_hImgSocket;      // 图片传输Socket
};
```

#### 7.5.1 发送流程 (CIUSocket::Send)

```
1. 业务层调用 CIUSocket::Send(strBuffer)
   └─> strBuffer 是 BinaryStreamWriter 序列化后的包体 (不含协议头)

2. CIUSocket::Send() 内部:
   ├─> std::lock_guard<std::mutex> guard(m_mtSendBuf);  // 加锁
   ├─> m_strSendBuf += strBuffer;  // 添加到发送缓冲区
   └─> m_cvSendBuf.notify_one();   // 唤醒发送线程

3. CIUSocket 发送线程 (SendThreadProc):
   ├─> 等待 m_cvSendBuf 信号
   ├─> 从 m_strSendBuf 取出数据
   ├─> 添加协议头 (struct msg):
   │   ├─> compressflag = 0 (不压缩)
   │   ├─> originsize = 数据长度
   │   ├─> compresssize = 数据长度 (未压缩)
   │   └─> reserved[16] = 0
   ├─> 可选: 使用 zlib 压缩 (如果数据较大)
   └─> 通过 send() 系统调用发送到服务器
```

#### 7.5.2 接收流程 (CIUSocket::RecvThreadProc)

```
1. CIUSocket 接收线程 (RecvThreadProc):
   ├─> 通过 recv() 系统调用从 Socket 接收数据
   ├─> 添加到 m_strRecvBuf 接收缓冲区
   └─> 调用 DecodePackages() 分包

2. DecodePackages() 处理粘包:
   ├─> 解析协议头 (struct msg):
   │   ├─> 读取 compressflag
   │   ├─> 读取 originsize (原始大小)
   │   ├─> 读取 compresssize (压缩后大小)
   │   └─> 跳过 reserved[16]
   ├─> 判断缓冲区是否包含完整包体
   ├─> 如果完整:
   │   ├─> 解压缩 (如果 compressflag == 1)
   │   ├─> 提取包体数据
   │   ├─> 传递给 CRecvMsgThread 处理
   │   └─> 从 m_strRecvBuf 移除已处理数据
   └─> 如果不完整: 继续等待接收更多数据

3. CRecvMsgThread 处理:
   └─> 使用 BinaryStreamReader 解析包体
       ├─> 读取 cmd (4字节)
       ├─> 读取 seq (4字节)
       ├─> 读取 json_data_len (4字节)
       └─> 读取 json_data
```

## 八、总结

FlamingoClient 是一个设计良好的即时通讯客户端，其核心特点包括：

1. **清晰的分层架构**: UI 层 → 业务层 → 发送/接收线程 → 网络层
2. **高效的线程模型**: 生产者-消费者模式实现异步消息处理
3. **灵活的协议设计**: 二进制头 + JSON 数据，兼顾效率和可读性
4. **完善的线程安全**: 使用 C++11 标准库 (std::mutex, std::condition_variable) 实现互斥和同步
5. **多连接管理**: 三个独立 Socket 连接 (聊天、文件、图片)，功能解耦
6. **完整的编解码**: Unicode ↔ UTF-8 编码转换，支持国际化

### 关键文件速查表

| 功能模块 | 关键文件 | 说明 |
|---------|---------|------|
| 协议定义 | `Source/net/Msg.h` | msg_type 枚举, struct msg 协议头 |
| 协议数据结构 | `Source/net/IUProtocolData.h` | CNetData 基类, 请求/响应类 |
| 协议序列化 | `Source/net/ProtocolStream.h` | BinaryStreamWriter/Reader |
| 网络层 | `Source/net/IUSocket.h/cpp` | CIUSocket 单例, Socket 通信 |
| 发送线程 | `Source/SendMsgThread.h/cpp` | CSendMsgThread, 生产者-消费者 |
| 接收线程 | `Source/RecvMsgThread.h/cpp` | CRecvMsgThread, 消息分发 |
| 主控制类 | `Source/FlamingoClient.h/cpp` | CFlamingoClient, 业务逻辑 |
| UI 界面 | `Source/FindFriendDlg.h/cpp` | 查找好友对话框示例 |
| Windows 消息 | `Source/UserSessionData.h` | FMG_MSG_* 自定义消息定义 |

掌握以上内容，就可以轻松地在 FlamingoClient 中添加新功能了！
