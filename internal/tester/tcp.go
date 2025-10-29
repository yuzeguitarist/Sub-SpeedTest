package tester

import (
	"fmt"
	"net"
	"time"
)

// tcpPing 测试TCP连接延迟（绕过系统代理）
func tcpPing(host, port string, timeout time.Duration) (int, error) {
	address := net.JoinHostPort(host, port)
	
	// 使用直连 dialer 绕过系统代理
	dialer := getDirectDialer(timeout)
	
	start := time.Now()
	conn, err := dialer.Dial("tcp", address)
	if err != nil {
		return -1, fmt.Errorf("TCP连接失败: %w", err)
	}
	defer conn.Close()
	
	latency := time.Since(start).Milliseconds()
	return int(latency), nil
}
