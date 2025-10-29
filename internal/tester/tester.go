package tester

import (
    "fmt"
    "proxy-tester/internal/parser"
    "sync"
    "time"
)

// TestNodes 并发测试所有节点
func TestNodes(nodes []*parser.Node, concurrency int, timeoutSec int) []*TestResult {
    results := make([]*TestResult, 0, len(nodes))
    var resultsMutex sync.Mutex
    
    // 创建工作池
    var wg sync.WaitGroup
    semaphore := make(chan struct{}, concurrency)

    // 进度计数
    var completed int
    var completedMutex sync.Mutex

    // 并发测试
    for _, node := range nodes {
        wg.Add(1)
        go func(n *parser.Node) {
            defer wg.Done()
            
            // 获取信号量
            semaphore <- struct{}{}
            defer func() { <-semaphore }()

            result := testNode(n, timeoutSec)
            
            // 保存结果
            resultsMutex.Lock()
            results = append(results, result)
            resultsMutex.Unlock()
            
            // 更新进度
            completedMutex.Lock()
            completed++
            current := completed
            completedMutex.Unlock()
            
            fmt.Printf("\r⏳ 测试进度: %d/%d", current, len(nodes))
        }(node)
    }

    // 等待所有测试完成
    wg.Wait()
    fmt.Println() // 换行

    return results
}

// testNode 测试单个节点
func testNode(node *parser.Node, timeoutSec int) *TestResult {
    result := &TestResult{
        Node:         node,
        ICMPLatency:  -1,
        TCPLatency:   -1,
        ProxyLatency: -1,
        Status:       "失败",
    }

    timeout := time.Duration(timeoutSec) * time.Second

    // 1. TCP Ping测试（快速测试端口是否可达）
    tcpLatency, tcpErr := tcpPing(node.Server, node.Port, timeout)
    if tcpErr == nil {
        result.TCPLatency = tcpLatency
    }

    // 2. 真实代理连接测试（包含 TLS 握手等）
    proxyLatency, proxyErr := testProxyConnection(node, timeout)
    if proxyErr == nil {
        result.ProxyLatency = proxyLatency
        result.Status = "成功"
    } else {
        // 记录详细错误信息
        result.Error = proxyErr.Error()
        
        // 如果 TCP 可达但代理测试失败，可能是 TLS 或其他问题
        if result.TCPLatency > 0 {
            result.Status = "端口可达但连接失败"
        } else {
            result.Status = "超时"
        }
    }

    return result
}
