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

	// 在网络操作正前方记录开始时间，确保只测量网络延迟
	start := time.Now()

	conn, err := dialer.Dial("tcp", address)

	// 立即计算延迟，避免包含后续操作的时间
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return -1, fmt.Errorf("TCP连接失败: %w", err)
	}
	defer conn.Close()

	return int(latency), nil
}
