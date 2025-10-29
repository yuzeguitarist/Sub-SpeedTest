package tester

import (
	"proxy-tester/internal/parser"
)

// TestResult 测试结果
type TestResult struct {
	Node         *parser.Node
	ICMPLatency  int    // ICMP延迟(ms), -1表示失败
	TCPLatency   int    // TCP延迟(ms), -1表示失败
	ProxyLatency int    // 真实代理连接延迟(ms), -1表示失败
	Status       string // 状态: 成功/超时/失败
	Error        string // 错误信息
}

// IsSuccess 判断测试是否成功
func (r *TestResult) IsSuccess() bool {
	return r.ProxyLatency > 0 || r.TCPLatency > 0
}
