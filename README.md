# Proxy Tester - macOS 代理节点批量测速工具

一个用于测试代理节点连通性和延迟的命令行工具，支持 VLESS、VMess 和 Shadowsocks 协议。

## 功能特性

- ✅ 支持订阅链接自动下载和解析
- ✅ 支持 VLESS、VMess、Shadowsocks (SS) 协议
- ✅ 并发测试，可自定义并发数
- ✅ 多种测速模式：TCP Ping、真实代理连接测试
- ✅ 结果按延迟自动排序
- ✅ 清晰的表格化结果展示
- ✅ 支持 IPv4 和 IPv6 地址
- ✅ **自动绕过系统代理** - 即使开启 VPN/代理工具（如 Shadowrocket）也能直连测试节点

## 安装

### 前置要求

- macOS 操作系统
- Go 1.21 或更高版本

### 编译

```bash
# 克隆或进入项目目录
cd /Users/jerry/Desktop/开发/节点测速

# 下载依赖
go mod download

# 编译
go build -o proxy-tester

# 或者直接运行
go run main.go
```

## 使用方法

### 基本用法

```bash
./proxy-tester test --url "https://example.com/subscription"
```

### 参数说明

- `-u, --url`: 订阅链接 URL（必需）
- `-c, --concurrency`: 并发测试数量（默认：10）
- `-t, --timeout`: 超时时间，单位秒（默认：5）
- `-v, --verbose`: 显示详细日志，包括解析过程和错误信息

### 使用示例

```bash
# 使用默认并发数测试
./proxy-tester test --url "https://example.com/sub"

# 自定义并发数为 20
./proxy-tester test --url "https://example.com/sub" -c 20

# 设置超时时间为 3 秒（推荐用于快速测试）
./proxy-tester test --url "https://example.com/sub" -t 3

# 显示详细日志（用于调试）
./proxy-tester test --url "https://example.com/sub" -v

# 组合使用：高并发 + 短超时 + 详细日志
./proxy-tester test -u "https://example.com/sub" -c 20 -t 3 -v
```

## 文档

- [更新日志](docs/CHANGELOG.md) - 版本历史和更新内容
- [故障排除](docs/TROUBLESHOOTING.md) - 常见问题和解决方案
- [功能路线图](docs/ROADMAP.md) - 计划中的新功能
- [架构说明](docs/ARCHITECTURE.md) - 技术实现和改进方向

## 输出示例

```
🔄 正在从 URL 下载订阅...
🔍 解码成功，正在解析节点...
✅ 发现 50 个节点，开始并发测试...

⏳ 测试进度: 50/50

✅ 测试完成，共 50 个节点，成功 45 个。结果按真实延迟排序：

────────────────────────────────────────────────────────────────────────────────────────────────────────
| 节点名称                         | 服务器地址             | TCP延迟  | 真实延迟 | 状态       |
────────────────────────────────────────────────────────────────────────────────────────────────────────
| HK-香港-01                      | 1.2.3.4:443          | 85ms    | 120ms    | ✅ 成功    |
| JP-东京-BGP                     | 5.6.7.8:443          | 110ms   | 155ms    | ✅ 成功    |
| SG-新加坡-05                    | 9.1.2.3:2053         | 150ms   | 210ms    | ✅ 成功    |
| US-洛杉矶-GIA                   | 4.5.6.7:80           | 250ms   | -        | ⏱️  超时   |
| DE-德国-02                      | 8.9.1.2:443          | -       | -        | ❌ 失败    |
```

## 项目结构

```
.
├── main.go                 # 程序入口
├── go.mod                  # Go 模块定义
├── cmd/                    # 命令行接口
│   ├── root.go            # 根命令
│   └── test.go            # test 子命令
├── internal/
│   ├── fetcher/           # 订阅下载和解码
│   │   └── fetcher.go
│   ├── parser/            # 节点解析
│   │   ├── types.go       # 数据类型定义
│   │   └── parser.go      # 解析器实现
│   ├── tester/            # 测速引擎
│   │   ├── types.go       # 测试结果类型
│   │   ├── tester.go      # 并发测试控制
│   │   ├── tcp.go         # TCP Ping
│   │   └── proxy.go       # 代理连接测试
│   └── display/           # 结果展示
│       └── display.go
└── README.md              # 项目说明
```

## 技术实现

### 支持的协议

1. **VLESS**: 解析格式 `vless://uuid@server:port?params#name`
2. **VMess**: 解析 Base64 编码的 JSON 配置
3. **Shadowsocks**: 解析格式 `ss://base64(method:password)@server:port#name`

### 测试模式

1. **TCP Ping**: 测试与服务器端口的 TCP 连接延迟
2. **真实连接测试**: 模拟真实代理握手，测试代理服务可用性

### 并发控制

使用 Goroutine 和信号量实现并发控制，避免过多并发导致系统资源耗尽。

## 注意事项

- 本工具仅用于测试节点连通性，不包含完整的代理协议实现
- 测试方式：建立 TCP/TLS 连接来验证节点是否可达（不进行完整的协议握手）
- 建议根据网络环境调整并发数和超时时间
- 测试时会跳过 TLS 证书验证以提高兼容性
- **代理绕过**：程序会自动绕过系统代理设置（包括 Shadowrocket 等工具），使用直连方式测试节点

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
