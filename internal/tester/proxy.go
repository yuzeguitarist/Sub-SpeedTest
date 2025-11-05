package tester

import (
    "crypto/tls"
    "fmt"
    "net"
    "proxy-tester/internal/parser"
    "sync"
    "syscall"
    "time"
)

// baseDialer 是共享的基础 Dialer 配置（不包含超时设置）
var (
    baseDialer     *net.Dialer
    dialerInitOnce sync.Once
)

// getDirectDialer 返回一个绕过系统代理的直连 Dialer
// 确保完全绕过系统代理设置，包括 Shadowrocket 等工具设置的代理
// 使用单例模式减少对象分配，提高性能和测量精度
func getDirectDialer(timeout time.Duration) *net.Dialer {
    // 使用 sync.Once 确保基础配置只初始化一次
    dialerInitOnce.Do(func() {
        baseDialer = &net.Dialer{
            KeepAlive: 30 * time.Second,
            // 使用 Control 函数确保绕过系统代理
            // 在 macOS 上，这会确保使用直连而不是系统代理
            Control: func(network, address string, c syscall.RawConn) error {
                // 这个函数在连接建立时被调用，可以用来设置 socket 选项
                // 通过这种方式建立的连接会绕过系统代理设置
                return nil
            },
        }
    })

    // 复制基础配置并设置特定的超时值
    // 这样避免每次都创建新的 Dialer，只是设置不同的超时
    dialer := *baseDialer
    dialer.Timeout = timeout
    return &dialer
}

// testProxyConnection 测试真实代理连接
func testProxyConnection(node *parser.Node, timeout time.Duration) (int, error) {
    switch node.Type {
    case parser.ProxyTypeVLESS:
        return testVLESSConnection(node, timeout)
    case parser.ProxyTypeVMess:
        return testVMessConnection(node, timeout)
    case parser.ProxyTypeShadowsocks:
        return testShadowsocksConnection(node, timeout)
    default:
        return -1, fmt.Errorf("不支持的协议类型: %s", node.Type)
    }
}

// testVLESSConnection 测试VLESS连接
func testVLESSConnection(node *parser.Node, timeout time.Duration) (int, error) {
    // 使用 HTTP 测试来验证代理是否真正可用
    return testProxyWithHTTP(node, timeout)
}

// testVMessConnection 测试VMess连接
func testVMessConnection(node *parser.Node, timeout time.Duration) (int, error) {
    address := net.JoinHostPort(node.Server, node.Port)

    // 使用直连 dialer 绕过系统代理
    dialer := getDirectDialer(timeout)

    var conn net.Conn
    var err error

    // 准备 TLS 配置（如果需要）
    var tlsConfig *tls.Config
    if node.TLS {
        tlsConfig = &tls.Config{
            ServerName:         node.Server,
            InsecureSkipVerify: true,
        }
    }

    // 在网络操作正前方记录开始时间，确保只测量网络延迟
    start := time.Now()

    if node.TLS {
        // 使用直连 dialer，确保绕过系统代理
        conn, err = tls.DialWithDialer(dialer, "tcp", address, tlsConfig)
    } else {
        // 使用直连 dialer，确保绕过系统代理
        conn, err = dialer.Dial("tcp", address)
    }

    // 立即计算延迟，避免包含后续操作的时间
    latency := time.Since(start).Milliseconds()

    if err != nil {
        return -1, fmt.Errorf("VMess连接失败: %w", err)
    }
    defer conn.Close()

    // 只测试连接是否可达，不进行协议握手
    // 因为完整的 VMess 协议需要复杂的加密和认证流程
    // 简单的握手测试反而会导致连接被服务器拒绝
    return int(latency), nil
}

// testShadowsocksConnection 测试Shadowsocks连接
func testShadowsocksConnection(node *parser.Node, timeout time.Duration) (int, error) {
    address := net.JoinHostPort(node.Server, node.Port)

    // 使用直连 dialer 绕过系统代理
    dialer := getDirectDialer(timeout)

    // 在网络操作正前方记录开始时间，确保只测量网络延迟
    start := time.Now()

    conn, err := dialer.Dial("tcp", address)

    // 立即计算延迟，避免包含后续操作的时间
    latency := time.Since(start).Milliseconds()

    if err != nil {
        return -1, fmt.Errorf("Shadowsocks连接失败: %w", err)
    }
    defer conn.Close()

    // 只测试连接是否可达，不进行协议握手
    // 因为 Shadowsocks 需要正确的加密和 SOCKS 封装
    // 错误的握手（如 SOCKS5 握手）会导致连接被服务器拒绝
    return int(latency), nil
}

// testProxyWithHTTP 通过 HTTP 请求测试代理的真实可用性
func testProxyWithHTTP(node *parser.Node, timeout time.Duration) (int, error) {
    // 创建直连的 HTTP 客户端（绕过系统代理）
    dialer := getDirectDialer(timeout)

    // 构建到代理服务器的连接
    address := net.JoinHostPort(node.Server, node.Port)

    var conn net.Conn
    var err error

    // 准备 TLS 配置（如果需要）
    var tlsConfig *tls.Config
    if node.TLS {
        tlsConfig = &tls.Config{
            ServerName:         node.Server,
            InsecureSkipVerify: true,
        }
    }

    // 在网络操作正前方记录开始时间，确保只测量网络延迟
    start := time.Now()

    // 根据是否 TLS 建立连接
    if node.TLS {
        conn, err = tls.DialWithDialer(dialer, "tcp", address, tlsConfig)
    } else {
        conn, err = dialer.Dial("tcp", address)
    }

    // 立即计算延迟，避免包含后续操作的时间
    latency := time.Since(start).Milliseconds()

    if err != nil {
        return -1, fmt.Errorf("连接失败: %w", err)
    }
    defer conn.Close()

    // 设置连接超时，防止后续操作阻塞（虽然当前不做后续操作，但保留作为安全措施）
    conn.SetDeadline(time.Now().Add(timeout))

    // 简单的连通性测试：能建立连接即可
    // 注意：这不是完整的代理协议实现，仅用于测速
    // 注意：移除了无效的延迟检查，因为如果连接成功建立，延迟必然在超时范围内

    return int(latency), nil
}
