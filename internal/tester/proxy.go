package tester

import (
	"crypto/tls"
	"fmt"
	"net"
	"proxy-tester/internal/parser"
	"time"
)

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
	address := net.JoinHostPort(node.Server, node.Port)
	
	start := time.Now()
	
	var conn net.Conn
	var err error
	
	if node.TLS {
		// TLS连接
		tlsConfig := &tls.Config{
			ServerName:         node.Server,
			InsecureSkipVerify: true, // 测速时跳过证书验证
		}
		conn, err = tls.DialWithDialer(&net.Dialer{Timeout: timeout}, "tcp", address, tlsConfig)
	} else {
		// 普通TCP连接
		conn, err = net.DialTimeout("tcp", address, timeout)
	}
	
	if err != nil {
		return -1, fmt.Errorf("VLESS连接失败: %w", err)
	}
	defer conn.Close()
	
	// 发送VLESS握手数据 (简化版本，仅测试连通性)
	// 实际生产环境需要完整的VLESS协议实现
	vlessHandshake := []byte{0x00} // 版本号
	_, err = conn.Write(vlessHandshake)
	if err != nil {
		return -1, fmt.Errorf("VLESS握手失败: %w", err)
	}
	
	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	
	// 尝试读取响应
	buf := make([]byte, 1)
	_, err = conn.Read(buf)
	// 即使读取失败，只要能建立连接就算成功
	
	latency := time.Since(start).Milliseconds()
	return int(latency), nil
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
