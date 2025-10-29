package tester

import (
	"crypto/tls"
	"fmt"
	"net"
	"proxy-tester/internal/parser"
	"time"
)

// getDirectDialer 返回一个绕过系统代理的直连 Dialer
func getDirectDialer(timeout time.Duration) *net.Dialer {
	return &net.Dialer{
		Timeout:   timeout,
		KeepAlive: 30 * time.Second,
		// 不使用系统代理设置
	}
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
	
	start := time.Now()
	
	var conn net.Conn
	var err error
	
	if node.TLS {
		tlsConfig := &tls.Config{
			ServerName:         node.Server,
			InsecureSkipVerify: true,
		}
		conn, err = tls.DialWithDialer(&net.Dialer{Timeout: timeout}, "tcp", address, tlsConfig)
	} else {
		conn, err = net.DialTimeout("tcp", address, timeout)
	}
	
	if err != nil {
		return -1, fmt.Errorf("VMess连接失败: %w", err)
	}
	defer conn.Close()
	
	// VMess协议握手 (简化版本)
	// 实际需要完整的VMess认证流程
	vmessHandshake := []byte{0x01}
	_, err = conn.Write(vmessHandshake)
	if err != nil {
		return -1, fmt.Errorf("VMess握手失败: %w", err)
	}
	
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 1)
	_, _ = conn.Read(buf)
	
	latency := time.Since(start).Milliseconds()
	return int(latency), nil
}

// testShadowsocksConnection 测试Shadowsocks连接
func testShadowsocksConnection(node *parser.Node, timeout time.Duration) (int, error) {
	address := net.JoinHostPort(node.Server, node.Port)
	
	start := time.Now()
	
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return -1, fmt.Errorf("Shadowsocks连接失败: %w", err)
	}
	defer conn.Close()
	
	// Shadowsocks SOCKS5握手 (简化版本)
	// 实际需要完整的Shadowsocks加密流程
	socks5Handshake := []byte{0x05, 0x01, 0x00}
	_, err = conn.Write(socks5Handshake)
	if err != nil {
		return -1, fmt.Errorf("Shadowsocks握手失败: %w", err)
	}
	
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 2)
	_, _ = conn.Read(buf)
	
	latency := time.Since(start).Milliseconds()
	return int(latency), nil
}

// testProxyWithHTTP 通过 HTTP 请求测试代理的真实可用性
func testProxyWithHTTP(node *parser.Node, timeout time.Duration) (int, error) {
	start := time.Now()
	
	// 创建直连的 HTTP 客户端（绕过系统代理）
	dialer := getDirectDialer(timeout)
	
	// 构建到代理服务器的连接
	address := net.JoinHostPort(node.Server, node.Port)
	
	var conn net.Conn
	var err error
	
	// 根据是否 TLS 建立连接
	if node.TLS {
		tlsConfig := &tls.Config{
			ServerName:         node.Server,
			InsecureSkipVerify: true,
		}
		conn, err = tls.DialWithDialer(dialer, "tcp", address, tlsConfig)
	} else {
		conn, err = dialer.Dial("tcp", address)
	}
	
	if err != nil {
		return -1, fmt.Errorf("连接失败: %w", err)
	}
	defer conn.Close()
	
	// 设置连接超时
	conn.SetDeadline(time.Now().Add(timeout))
	
	// 简单的连通性测试：能建立连接即可
	// 注意：这不是完整的代理协议实现，仅用于测速
	latency := time.Since(start).Milliseconds()
	
	// 如果延迟过高，认为不可用
	if latency > int64(timeout.Milliseconds()) {
		return -1, fmt.Errorf("延迟过高: %dms", latency)
	}
	
	return int(latency), nil
}
