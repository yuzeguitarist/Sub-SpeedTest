package cmd

import (
	"fmt"
	"os"
	"proxy-tester/internal/fetcher"
	"proxy-tester/internal/parser"
	"proxy-tester/internal/tester"
	"proxy-tester/internal/display"

	"github.com/spf13/cobra"
)

var (
	subscriptionURL string
	concurrency     int
	timeout         int
	verbose         bool
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "测试订阅链接中的所有节点",
	Long:  `从订阅链接下载节点信息，解析并并发测试所有节点的连通性和延迟。`,
	Run:   runTest,
}

func init() {
	testCmd.Flags().StringVarP(&subscriptionURL, "url", "u", "", "订阅链接URL (必需)")
	testCmd.Flags().IntVarP(&concurrency, "concurrency", "c", 10, "并发测试数量")
	testCmd.Flags().IntVarP(&timeout, "timeout", "t", 5, "超时时间(秒)")
	testCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "显示详细日志")
	testCmd.MarkFlagRequired("url")
}

func runTest(cmd *cobra.Command, args []string) {
	// 1. 下载订阅
	if verbose {
		fmt.Printf("🔄 正在从 URL 下载订阅: %s\n", subscriptionURL)
	} else {
		fmt.Printf("🔄 正在从 URL 下载订阅...\n")
	}
	
	content, err := fetcher.FetchSubscription(subscriptionURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ 下载订阅失败: %v\n", err)
		os.Exit(1)
	}

	if verbose {
		fmt.Printf("📦 下载成功，内容长度: %d 字节\n", len(content))
		fmt.Printf("📝 前100字符: %s\n\n", content[:min(100, len(content))])
	}

	// 2. 解析节点
	fmt.Printf("🔍 解码成功，正在解析节点...\n")
	if verbose {
		fmt.Println()
	}
	
	nodes, err := parser.ParseNodes(content, verbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ 解析节点失败: %v\n", err)
		os.Exit(1)
	}

	if len(nodes) == 0 {
		fmt.Println("⚠️  未发现任何节点")
		if !verbose {
			fmt.Println("💡 提示: 使用 -v 参数查看详细日志")
		}
		return
	}

	fmt.Printf("✅ 发现 %d 个节点，开始并发测试...\n", len(nodes))
	if verbose {
		fmt.Printf("⚙️  并发数: %d, 超时: %d秒\n\n", concurrency, timeout)
	} else {
		fmt.Println()
	}

	// 3. 并发测试
	results := tester.TestNodes(nodes, concurrency, timeout)

	// 4. 显示结果
	display.ShowResults(results)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
