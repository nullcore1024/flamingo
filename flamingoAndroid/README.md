# FlamingoAndroid 实现原理详解

## 一、项目概述

FlamingoAndroid 是一个即时通讯应用的 Android 客户端，采用了经典的分层架构设计，实现了用户注册、登录、好友管理、群聊、单聊等功能。

## 二、核心技术栈

### 2.1 开发语言与框架
- **语言**: Java
- **开发框架**: Android SDK (API Level 16+)

### 2.2 核心第三方库
- **Gson**: Google 提供的 JSON 序列化/反序列化库，用于 Java 对象与 JSON 数据的相互转换
- **Universal Image Loader**: 图片加载库，用于显示用户头像等
- **GreenDAO**: ORM 数据库框架，用于本地数据存储（聊天记录等）
- **Zlib**: 数据压缩库，用于网络数据的压缩传输

### 2.3 并发与线程通信
- **Java Socket**: 用于与服务器建立长连接
- **Thread**: 用于处理网络读写等耗时操作
- **Android Handler**: 用于工作线程与 UI 线程之间的消息通信
- **Message**: Android 消息对象，用于在 Handler 之间传递数据

## 三、网络协议设计

### 3.1 协议结构

FlamingoAndroid 采用自定义的二进制协议，分为**包头**和**包体**两部分：

#### 包头格式（固定 25 字节）

| 字段名       | 类型   | 字节数 | 说明                     |
|--------------|--------|--------|--------------------------|
| compressflag | byte   | 1      | 压缩标志：0=不压缩，1=压缩 |
| originsize   | int    | 4      | 原始数据大小（小端序）   |
| compresssize | int    | 4      | 压缩后数据大小（小端序） |
| reserved     | byte[] | 16     | 保留字段（暂未使用）     |

#### 包体格式

| 字段名 | 类型   | 字节数 | 说明                     |
|--------|--------|--------|--------------------------|
| cmd    | int    | 4      | 协议命令号（小端序）     |
| seq    | int    | 4      | 序列号（小端序，用于请求响应匹配） |
| json   | string | 可变   | JSON 格式的业务数据      |

**注意**：对于聊天消息（cmd=1100），包体在 json 之后还包含两个 int 字段：
- arg1: 目标用户/群 ID
- arg2: 消息发送者 ID

### 3.2 协议命令号定义（MsgType.java）

| 命令常量名           | 数值 | 说明                 |
|----------------------|------|----------------------|
| msg_type_register    | 1001 | 用户注册             |
| msg_type_login       | 1002 | 用户登录             |
| msg_type_getfriendlist | 1003 | 获取好友列表       |
| msg_type_finduser    | 1004 | 搜索用户/群          |
| msg_type_operatefriend | 1005 | 操作好友（添加/删除/同意） |
| msg_type_chat        | 1100 | 单聊消息             |
| msg_type_multichat   | 1101 | 群聊消息             |
| msg_type_creategroup | 1009 | 创建群组             |

### 3.3 字节序说明

- 所有多字节数值类型（int、short）均使用**小端序（Little-Endian）**存储
- 在 `BinaryReadStream` 和 `BinaryWriteStream` 中进行字节序转换

## 四、核心架构与流程

### 4.1 整体架构

```
┌─────────────────────────────────────────────────────────┐
│                     UI 层 (Activity)                     │
│  ┌──────────────────┐  ┌─────────────────────────────┐ │
│  │ AddFriendActivity│  │ ChattingActivity            │ │
│  └────────┬─────────┘  └──────────────┬──────────────┘ │
└───────────┼─────────────────────────────┼────────────────┘
            │                             │
            │ processMessage()            │
            ▼                             ▼
┌─────────────────────────────────────────────────────────┐
│              BaseActivity (Handler 机制)                 │
│         静态 Handler: 统一消息分发中心                   │
└─────────────────────────────────────────────────────────┘
            ▲
            │ notifyUI()
            │
┌─────────────────────────────────────────────────────────┐
│                   业务逻辑层 (NetWorker)                 │
│  ┌───────────────────────────────────────────────────┐  │
│  │  - 发送请求: addPackage()                          │  │
│  │  - 接收响应: recvPackage()                         │  │
│  │  - 解析响应: handleServerResponseMsg()             │  │
│  └───────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
            ▲
            │
┌─────────────────────────────────────────────────────────┐
│                协议层 (NetPackage)                        │
│  - 封装数据包                                             │
│  - 二进制序列化/反序列化                                  │
└─────────────────────────────────────────────────────────┘
            ▲
            │
┌─────────────────────────────────────────────────────────┐
│                网络层 (Socket)                            │
│  - TCP 长连接                                             │
│  - 数据读写                                               │
└─────────────────────────────────────────────────────────┘
```

### 4.2 网络工作线程流程

NetWorker 中有一个独立的网络线程 `startNetChatThread()`，负责持续处理数据包的发送和接收：

```
while (mNetChatThreadRunning) {
    // 1. 发送队列中的所有数据包
    while (有数据包待发送) {
        retrievePackage() → 获取数据包
        writePackage() → 写入 Socket
    }
    
    // 2. 接收服务器响应
    recvPackage() → 读取包头和包体
    handleServerResponseMsg() → 根据 cmd 分发处理
}
```

## 五、回调机制与消息分发详解

### 5.1 回调机制

**问：网络协议和回包是通过注册回调函数来实现的吗？**

**答：不是传统的回调函数注册机制，而是采用了 Android 的 Handler 消息机制。**

### 5.2 完整的消息分发流程

以**查找好友**功能为例，详细说明从输入框输入到结果展示的完整流程：

#### 阶段 1：UI 层读取输入并发送请求

**文件位置**: `AddFriendActivity.java:56-68`

```java
// 1. 用户点击搜索按钮
case R.id.tv_search:
    // 2. 读取输入框内容
    String searchAccount = et_addSearch.getText().toString().trim();
    
    // 3. 调用 NetWorker 发送网络请求
    NetWorker.searchPersonOrGroup(tabType, searchAccount);
    break;
```

#### 阶段 2：NetWorker 封装数据包

**文件位置**: `NetWorker.java:684-694`

```java
public static void searchPersonOrGroup(int type, String userName) {
    // 1. 创建请求数据模型
    SearchUser searchUser = new SearchUser(type, userName);
    
    // 2. 序列化为 JSON
    String json = new Gson().toJson(searchUser, SearchUser.class);
    
    // 3. 创建网络数据包（cmd=1004, seq自增）
    NetPackage netPackage = new NetPackage(MsgType.msg_type_finduser, mSeq, json);
    mSeq++;
    
    // 4. 加入发送队列
    addPackage(netPackage);
}
```

**NetPackage 构造过程** (`NetPackage.java:19-30`):
```java
public NetPackage(int cmd, int seq, String json){
    mWriteStream.writeInt32(cmd);      // 写入命令号
    mWriteStream.writeInt32(seq);      // 写入序列号
    byte[] bytes = json.getBytes("UTF-8");
    mWriteStream.writeBytes(bytes);     // 写入 JSON 数据
    mWriteStream.flush();
}
```

#### 阶段 3：网络线程发送数据包

**文件位置**: `NetWorker.java:732-772` (startNetChatThread)

```java
// 从队列中取出数据包
NetPackage netPackage = NetWorker.retrievePackage();

// 写入 Socket
writePackage(netPackage, mDataOutputStream);
```

**writePackage 实现** (`NetWorker.java:205-230`):
```java
public static boolean writePackage(NetPackage p, DataOutputStream stream) {
    byte[] b = p.getBytes();
    int packageSize = p.getBytesSize();
    
    // 写入包头（25字节）
    stream.writeByte(NetPackage.PACKAGE_UNCOMPRESSED_FLAG);  // 压缩标志
    stream.writeInt(BinaryWriteStream.intToLittleEndian(packageSize));  // 原始大小
    stream.writeInt(0);  // 压缩后大小（不压缩为0）
    byte[] reserved = new byte[16];
    stream.write(reserved);  // 保留字段
    
    // 写入包体
    stream.write(b, 0, b.length);
    stream.flush();
    return true;
}
```

#### 阶段 4：接收服务器响应

**文件位置**: `NetWorker.java:791-844` (recvPackage)

```java
private static boolean recvPackage(DataInputStream inputStream, NetDataParser parser) {
    // 1. 读取包头
    byte compressFlag = inputStream.readByte();
    int rawpackagelength = inputStream.readInt();
    int packagelength = BinaryReadStream.intToBigEndian(rawpackagelength);
    int compresslengthRaw = inputStream.readInt();
    int compresslength = BinaryReadStream.intToBigEndian(compresslengthRaw);
    inputStream.skipBytes(16);  // 跳过保留字段
    
    // 2. 读取包体
    byte[] bodybuf = new byte[compresslength];
    inputStream.read(bodybuf);
    
    // 3. 解压（如果需要）
    byte[] result = (compressFlag == 1) ? ZlibUtil.decompressBytes(bodybuf) : bodybuf;
    
    // 4. 解析包体
    BinaryReadStream binaryReadStream = new BinaryReadStream(result);
    int cmd = binaryReadStream.readInt32();
    int seq = binaryReadStream.readInt32();
    String retJson = binaryReadStream.readString();
    
    parser.mCmd = cmd;
    parser.mSeq = seq;
    parser.mJson = retJson;
    return true;
}
```

#### 阶段 5：根据 cmd 分发处理

**文件位置**: `NetWorker.java:846-915` (handleServerResponseMsg)

```java
private static void handleServerResponseMsg(NetDataParser parser) {
    switch (parser.mCmd) {
        case MsgType.msg_type_finduser:
            // 调用查找好友结果处理函数
            handleFindUserResult(parser.mJson);
            break;
        // ... 其他 cmd 处理
    }
}
```

#### 阶段 6：解析响应数据并通知 UI

**文件位置**: `NetWorker.java:1414-1481` (handleFindUserResult)

```java
private static void handleFindUserResult(String data) {
    // 1. 解析 JSON 响应
    List<UserInfo> matchedUsers = new ArrayList<>();
    JsonReader reader = new JsonReader(new StringReader(data));
    reader.beginObject();
    while (reader.hasNext()) {
        String name = reader.nextName();
        if (name.equals("userinfo") && reader.peek() != JsonToken.NULL) {
            reader.beginArray();
            while (reader.hasNext()) {
                // 解析每个用户信息
                reader.beginObject();
                UserInfo u = new UserInfo();
                while (reader.hasNext()) {
                    String nodename2 = reader.nextName();
                    if (nodename2.equals("userid")) {
                        u.set_userid(reader.nextInt());
                    } else if (nodename2.equals("username")) {
                        u.set_username(reader.nextString());
                    }
                    // ... 其他字段
                }
                reader.endObject();
                matchedUsers.add(u);
            }
            reader.endArray();
        } else {
            reader.skipValue();
        }
    }
    reader.endObject();
    
    // 2. 通知 UI 层（通过 Handler 消息机制）
    Message msg = new Message();
    msg.what = MsgType.msg_type_finduser;  // 消息类型
    msg.obj = matchedUsers;                  // 携带的数据
    BaseActivity.sendMessage(msg);           // 发送消息
}
```

#### 阶段 7：BaseActivity 统一分发消息

**文件位置**: `BaseActivity.java:109-125`

```java
// 静态 Handler：所有 Activity 共用
private static Handler handler = new Handler() {
    @Override
    public void handleMessage(Message msg) {
        // 获取当前前台 Activity
        BaseActivity acitivity = ((BaseActivity) application.getAppManager().currentActivity());
        
        if (acitivity != null) {
            if (msg.what == MsgType.msg_networker_disconnect) {
                // 网络断开处理
            } else {
                // 调用当前 Activity 的 processMessage()
                acitivity.processMessage(msg);
            }
        }
    }
};
```

#### 阶段 8：Activity 处理消息并更新 UI

**文件位置**: `AddFriendActivity.java:111-132`

```java
@Override
public void processMessage(Message msg) {
    super.processMessage(msg);
    
    if (msg.what == MsgType.msg_type_finduser) {
        // 1. 获取返回的数据
        findUserResult = (List<UserInfo>) msg.obj;
        
        // 2. 更新 RecyclerView 数据
        mAdapter.notifyDataSetChanged();
        
        // 3. 提示用户
        if (findUserResult.isEmpty()) {
            Toast.makeText(this, "搜索的用户不存在", Toast.LENGTH_SHORT).show();
        }
    }
}
```

## 六、新增界面与协议的完整开发流程

假设我们要新增一个**修改用户昵称**的功能，步骤如下：

### 步骤 1：定义协议命令号

在 `MsgType.java` 中新增命令号：
```java
public static final int msg_type_updatenickname = 1011;
```

### 步骤 2：创建请求和响应数据模型

**请求模型** (`UpdateNickname.java`):
```java
public class UpdateNickname {
    private String nickname;
    
    public UpdateNickname(String nickname) {
        this.nickname = nickname;
    }
    
    // getter/setter
}
```

**响应模型** (`UpdateNicknameResult.java`):
```java
public class UpdateNicknameResult {
    private int code;
    private String msg;
    
    // getter/setter
}
```

### 步骤 3：在 NetWorker 中添加发送请求方法

```java
public static void updateNickname(String nickname) {
    try {
        UpdateNickname request = new UpdateNickname(nickname);
        NetPackage netPackage = new NetPackage(
            MsgType.msg_type_updatenickname, 
            mSeq, 
            new Gson().toJson(request, UpdateNickname.class)
        );
        mSeq++;
        addPackage(netPackage);
    } catch (Exception e) {
        e.printStackTrace();
    }
}
```

### 步骤 4：在 NetWorker 中添加响应处理

在 `handleServerResponseMsg()` 的 switch 中添加：
```java
case MsgType.msg_type_updatenickname:
    handleUpdateNicknameResult(parser.mJson);
    break;
```

添加处理函数：
```java
private static void handleUpdateNicknameResult(String data) {
    int retCode = MsgType.ERROR_CODE_UNKNOWNFAILED;
    try {
        JsonReader reader = new JsonReader(new StringReader(data));
        reader.beginObject();
        while (reader.hasNext()) {
            String name = reader.nextName();
            if (name.equals("code")) {
                retCode = reader.nextInt();
            } else {
                reader.skipValue();
            }
        }
        reader.endObject();
        reader.close();
    } catch (Exception e) {
        e.printStackTrace();
    }
    
    Message msg = new Message();
    msg.what = MsgType.msg_type_updatenickname;
    msg.arg1 = retCode;
    BaseActivity.sendMessage(msg);
}
```

### 步骤 5：创建新的 Activity

```java
public class UpdateNicknameActivity extends BaseActivity {
    private EditText etNickname;
    private Button btnSubmit;
    
    @Override
    public void onClick(View v) {
        if (v.getId() == R.id.btn_submit) {
            String nickname = etNickname.getText().toString().trim();
            if (!nickname.isEmpty()) {
                // 调用 NetWorker 发送请求
                NetWorker.updateNickname(nickname);
            }
        }
    }
    
    @Override
    public void processMessage(Message msg) {
        super.processMessage(msg);
        
        if (msg.what == MsgType.msg_type_updatenickname) {
            int retCode = msg.arg1;
            if (retCode == MsgType.ERROR_CODE_SUCCESS) {
                Toast.makeText(this, "修改成功", Toast.LENGTH_SHORT).show();
                finish();
            } else {
                Toast.makeText(this, "修改失败", Toast.LENGTH_SHORT).show();
            }
        }
    }
    
    @Override
    protected int getContentView() {
        return R.layout.activity_update_nickname;
    }
    
    @Override
    protected void initData() {
        btnSubmit.setOnClickListener(this);
    }
    
    @Override
    protected void setData() {
    }
}
```

### 步骤 6：创建布局文件

在 `res/layout/activity_update_nickname.xml` 中创建布局，包含 EditText 和 Button。

## 七、关键文件说明

| 文件路径 | 说明 |
|---------|------|
| `net/NetWorker.java` | 核心网络管理类，负责 Socket 连接、数据包收发、响应处理 |
| `net/NetPackage.java` | 网络数据包封装类 |
| `net/BinaryReadStream.java` | 二进制读取流，处理字节序转换 |
| `net/BinaryWriteStream.java` | 二进制写入流，处理字节序转换 |
| `enums/MsgType.java` | 协议命令号和错误码定义 |
| `activities/BaseActivity.java` | 所有 Activity 的基类，包含统一的 Handler 消息分发机制 |
| `activities/member/AddFriendActivity.java` | 添加好友 Activity，展示完整的请求-响应流程 |
| `model/SearchUser.java` | 搜索用户请求模型 |

## 八、总结

FlamingoAndroid 的核心设计特点：

1. **自定义二进制协议**：包头 + 包体（JSON）的组合，兼顾效率和灵活性
2. **长连接模式**：使用 Socket 保持与服务器的持续连接
3. **Handler 消息机制**：不是传统的回调注册，而是通过统一的静态 Handler 进行消息分发
4. **分层架构**：UI 层 → 业务逻辑层 → 协议层 → 网络层，职责清晰
5. **队列机制**：使用发送队列（mNetPackages）缓冲待发送的数据包

整个流程的核心就是：**Activity 触发请求 → NetWorker 封装并发送 → 服务器响应 → NetWorker 解析并通过 Handler 发送消息 → Activity 接收消息并更新 UI**。
