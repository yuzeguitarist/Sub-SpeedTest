package display

import (
	"fmt"
	"proxy-tester/internal/tester"
	"sort"
	"strings"
)

// ShowResults 显示测试结果
func ShowResults(results []*tester.TestResult) {
	if len(results) == 0 {
		fmt.Println("⚠️  没有测试结果")
		return
	}

	// 按真实延迟排序 (低到高)
	sortResults(results)

	// 统计
	total := len(results)
	success := 0
	for _, r := range results {
		if r.IsSuccess() {
			success++
		}
	}

	// 打印标题
	fmt.Printf("\n✅ 测试完成，共 %d 个节点，成功 %d 个。结果按真实延迟排序：\n\n", total, success)

	// 打印表头
	printTableHeader()

	// 打印每个结果
	for _, result := range results {
		printTableRow(result)
	}

	fmt.Println()
}

// sortResults 按真实延迟排序
func sortResults(results []*tester.TestResult) {
	sort.Slice(results, func(i, j int) bool {
		// 成功的排在前面
		if results[i].IsSuccess() && !results[j].IsSuccess() {
			return true
		}
		if !results[i].IsSuccess() && results[j].IsSuccess() {
			return false
		}

		// 都成功时，按真实延迟排序
		if results[i].IsSuccess() && results[j].IsSuccess() {
			// 优先使用ProxyLatency
			latencyI := results[i].ProxyLatency
			if latencyI <= 0 {
				latencyI = results[i].TCPLatency
			}
			latencyJ := results[j].ProxyLatency
			if latencyJ <= 0 {
				latencyJ = results[j].TCPLatency
			}
			return latencyI < latencyJ
		}

		// 都失败时保持原顺序
		return false
	})
}

// printTableHeader 打印表头
func printTableHeader() {
	fmt.Println(strings.Repeat("─", 100))
	fmt.Printf("| %-30s | %-20s | %-8s | %-8s | %-10s |\n",
		"节点名称", "服务器地址", "TCP延迟", "真实延迟", "状态")
	fmt.Println(strings.Repeat("─", 100))
}

// printTableRow 打印单行结果
func printTableRow(result *tester.TestResult) {
	name := truncateString(result.Node.Name, 30)
	if name == "" {
		name = "未命名"
	}

	address := truncateString(result.Node.Address(), 20)

	tcpLatency := formatLatency(result.TCPLatency)
	proxyLatency := formatLatency(result.ProxyLatency)
	status := formatStatus(result.Status)

	fmt.Printf("| %-30s | %-20s | %-8s | %-8s | %-10s |\n",
		name, address, tcpLatency, proxyLatency, status)
}

// formatLatency 格式化延迟显示
func formatLatency(latency int) string {
	if latency < 0 {
		return "-"
	}
	return fmt.Sprintf("%dms", latency)
}

// formatStatus 格式化状态显示
func formatStatus(status string) string {
	switch status {
	case "成功":
		return "✅ 成功"
	case "超时":
		return "⏱️  超时"
	case "失败":
		return "❌ 失败"
	default:
		return status
	}
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	// 处理中文字符
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-3]) + "..."
}
