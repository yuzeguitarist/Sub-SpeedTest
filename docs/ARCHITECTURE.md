# 架构说明

## 当前实现

### 1. 代理绕过机制

程序已实现自动绕过系统代理，确保测试时直连目标节点：

**实现位置**:
- `internal/fetcher/fetcher.go` - HTTP 客户端设置 `Proxy: nil`
- `internal/tester/tcp.go` - TCP 连接使用直连 Dialer
- `internal/tester/proxy.go` - 代理测试使用直连 Dialer

**工作原理**:
```go
// 创建绕过系统代理的 Dialer
dialer := &net.Dialer{
    Timeout:   timeout,
    KeepAlive: 30 * time.Second,
}

// HTTP Transport 禁用代理
transport := &http.Transport{
    Proxy: nil,  // 关键：设置为 nil 禁用系统代理
    DialContext: dialer.DialContext,
}
```

即使用户开启了 Shadowrocket 全局代理，程序也会直接连接到测试节点，不会通过代理转发。

### 2. 测试流程

```
订阅 URL
    ↓
下载并解码 (绕过代理)
    ↓
解析节点信息
    ↓
并发测试 (绕过代理)
    ├─ TCP Ping
    └─ 代理连接测试
        ↓
结果排序和展示
```

### 3. 当前测试方法的局限性

**问题**: 简化的连接测试不够准确

当前实现只是建立 TCP/TLS 连接，没有完整实现代理协议握手：

```go
// 当前实现（简化版）
func testProxyWithHTTP(node *parser.Node, timeout time.Duration) (int, error) {
    // 只建立连接，不发送实际的代理协议数据
    conn, err := dialer.Dial("tcp", address)
    if err != nil {
        return -1, err
    }
    defer conn.Close()
    
    latency := time.Since(start).Milliseconds()
    return int(latency), nil
}
```

**为什么会出现误判**:
1. **TCP 连接成功 ≠ 代理可用**: 端口开放不代表代理协议正确
2. **缺少协议握手**: 没有验证 VLESS/VMess 的实际认证流程
3. **没有真实请求**: 没有通过代理发送 HTTP 请求验证端到端连通性

### 4. 改进方向

#### 方案 A: 完整协议实现（复杂但准确）

实现完整的 VLESS/VMess/Shadowsocks 协议：

```go
// VLESS 完整握手
func testVLESSConnection(node *parser.Node, timeout time.Duration) (int, error) {
    // 1. 建立连接
    conn := establishConnection(node, timeout)
    
    // 2. 发送 VLESS 请求头
    vlessHeader := buildVLESSHeader(node.UUID, "www.google.com", 443)
    conn.Write(vlessHeader)
    
    // 3. 发送 HTTP CONNECT 请求
    httpRequest := "CONNECT www.google.com:443 HTTP/1.1\r\n\r\n"
    conn.Write([]byte(httpRequest))
    
    // 4. 读取响应验证
    response := readResponse(conn)
    if !isValidResponse(response) {
        return -1, errors.New("代理握手失败")
    }
    
    return latency, nil
}
```

**优点**: 最准确，真实模拟代理客户端
**缺点**: 实现复杂，需要深入理解各协议细节

#### 方案 B: 使用现有库（推荐）

集成成熟的代理库：

```go
import (
    "github.com/v2ray/v2ray-core/proxy/vless"
    "github.com/v2ray/v2ray-core/proxy/vmess"
)

func testVLESSConnection(node *parser.Node, timeout time.Duration) (int, error) {
    // 使用 v2ray-core 库建立真实的 VLESS 连接
    client := vless.NewClient(node.UUID, node.Server, node.Port)
    
    // 通过代理发送测试请求
    err := client.Dial("www.google.com:443", timeout)
    if err != nil {
        return -1, err
    }
    
    return latency, nil
}
```

**优点**: 准确且维护成本低
**缺点**: 增加依赖，二进制体积增大

#### 方案 C: HTTP 探测（平衡方案）

通过代理发送真实的 HTTP 请求：

```go
func testProxyWithHTTP(node *parser.Node, timeout time.Duration) (int, error) {
    // 1. 建立到代理的连接
    conn := establishConnection(node, timeout)
    
    // 2. 发送 CONNECT 请求（通过代理访问测试站点）
    connectReq := "CONNECT www.google.com:443 HTTP/1.1\r\n" +
                  "Host: www.google.com:443\r\n\r\n"
    conn.Write([]byte(connectReq))
    
    // 3. 读取代理响应
    response := readHTTPResponse(conn)
    if !strings.Contains(response, "200") {
        return -1, errors.New("代理连接失败")
    }
    
    // 4. 可选：发送实际 HTTPS 握手验证端到端连通性
    tlsConn := tls.Client(conn, &tls.Config{ServerName: "www.google.com"})
    err := tlsConn.Handshake()
    
    return latency, nil
}
```

**优点**: 相对简单，准确性较高
**缺点**: 仍需处理不同协议的差异

## 下一步计划

### 短期（v1.1）
1. ✅ 实现代理绕过机制
2. ⏳ 改进测试准确性（方案 C）
3. ⏳ 添加订阅过滤和导出功能

### 中期（v1.2）
1. 完整协议支持（方案 B）
2. 多订阅源合并
3. HTTP 服务器托管

### 长期（v2.0）
1. Web 管理界面
2. 历史记录和趋势分析
3. 自动化部署方案

## 技术债务

1. **测试准确性**: 当前简化实现可能误判，需要完整协议支持
2. **错误处理**: 部分错误信息不够详细
3. **性能优化**: 高并发时可能有资源竞争
4. **配置管理**: 缺少配置文件支持

## 贡献建议

如果你想改进测试准确性，建议：

1. **了解协议**: 阅读 VLESS/VMess/Shadowsocks 协议文档
2. **参考实现**: 查看 v2ray-core、clash 等项目的实现
3. **逐步改进**: 先支持一个协议，再扩展到其他协议
4. **充分测试**: 使用真实订阅链接验证改进效果

## 相关资源

- [VLESS 协议文档](https://github.com/XTLS/Xray-core)
- [VMess 协议文档](https://www.v2ray.com/developer/protocols/vmess.html)
- [Shadowsocks 协议](https://shadowsocks.org/en/spec/Protocol.html)
- [v2ray-core 源码](https://github.com/v2fly/v2ray-core)
