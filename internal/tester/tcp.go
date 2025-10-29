package tester

import (
	"fmt"
	"net"
	"time"
)

// tcpPing 测试TCP连接延迟
func tcpPing(host, port string, timeout time.Duration) (int, error) {
	address := net.JoinHostPort(host, port)
	
	start := time.Now()
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return -1, fmt.Errorf("TCP连接失败: %w", err)
	}
	defer conn.Close()
	
	latency := time.Since(start).Milliseconds()
	return int(latency), nil
}
