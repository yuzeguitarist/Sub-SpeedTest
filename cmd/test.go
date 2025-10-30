package cmd

import (
    "fmt"
    "os"
    "proxy-tester/internal/display"
    "proxy-tester/internal/fetcher"
    "proxy-tester/internal/parser"
    "proxy-tester/internal/tester"

    "github.com/fatih/color"
    "github.com/spf13/cobra"
)

var (
    subscriptionURL string
    concurrency     int
    timeout         int
    verbose         bool
)

// 定义颜色函数
var (
    cyan    = color.New(color.FgCyan).SprintFunc()
    cyanB   = color.New(color.FgCyan, color.Bold).SprintFunc()
    green   = color.New(color.FgGreen).SprintFunc()
    greenB  = color.New(color.FgGreen, color.Bold).SprintFunc()
    yellow  = color.New(color.FgYellow).SprintFunc()
    red     = color.New(color.FgRed).SprintFunc()
    redB    = color.New(color.FgRed, color.Bold).SprintFunc()
    white   = color.New(color.FgWhite).SprintFunc()
    whiteB  = color.New(color.FgWhite, color.Bold).SprintFunc()
    gray    = color.New(color.FgHiBlack).SprintFunc()
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
    // 标准化参数：确保并发数和超时时间在有效范围内
    normalizedConcurrency := concurrency
    if normalizedConcurrency < 1 {
        normalizedConcurrency = 1
    }
    
    normalizedTimeout := timeout
    if normalizedTimeout <= 0 {
        normalizedTimeout = 30
    }
    
    // 打印欢迎横幅
    printBanner()
    
    // 显示代理绕过提示
    if verbose {
        fmt.Printf("  %s %s\n", cyan("ℹ"), gray("已启用代理绕过模式，所有连接将直连目标服务器"))
        fmt.Printf("  %s %s\n", cyan("ℹ"), gray("即使系统开启了 VPN 或代理（如 Shadowrocket），也会被绕过"))
        fmt.Println()
    }

    // 1. 下载订阅
    if verbose {
        fmt.Printf("  %s %s\n", cyanB("→"), white("正在从 URL 下载订阅..."))
        fmt.Printf("    %s\n\n", gray(subscriptionURL))
    } else {
        fmt.Printf("  %s %s\n\n", cyanB("→"), white("正在从 URL 下载订阅..."))
    }
    
    content, err := fetcher.FetchSubscription(subscriptionURL)
    if err != nil {
        fmt.Fprintf(os.Stderr, "  %s %s\n", redB("✗"), red(fmt.Sprintf("下载订阅失败: %v", err)))
        os.Exit(1)
    }

    if verbose {
        fmt.Printf("  %s %s\n", greenB("✓"), white(fmt.Sprintf("下载成功，内容长度: %d 字节", len(content))))
        preview := content
        if len(preview) > 100 {
            preview = preview[:100] + "..."
        }
        fmt.Printf("    %s\n\n", gray(fmt.Sprintf("内容预览: %s", preview)))
    } else {
        fmt.Printf("  %s %s\n\n", greenB("✓"), white("下载成功"))
    }

    // 2. 解析节点
    if verbose {
        fmt.Printf("  %s %s\n\n", cyanB("→"), white("正在解析节点..."))
    } else {
        fmt.Printf("  %s %s\n", cyanB("→"), white("正在解析节点..."))
    }
    
    nodes, err := parser.ParseNodes(content, verbose)
    if err != nil {
        fmt.Fprintf(os.Stderr, "  %s %s\n", redB("✗"), red(fmt.Sprintf("解析节点失败: %v", err)))
        os.Exit(1)
    }

    if len(nodes) == 0 {
        fmt.Printf("  %s %s\n", yellow("⚠"), yellow("未发现任何节点"))
        if !verbose {
            fmt.Printf("  %s %s\n", cyan("💡"), gray("提示: 使用 -v 参数查看详细日志"))
        }
        return
    }

    fmt.Printf("  %s %s\n", greenB("✓"), whiteB(fmt.Sprintf("发现 %d 个节点", len(nodes))))
    if verbose {
        fmt.Printf("    %s\n", gray(fmt.Sprintf("并发数: %d, 超时: %d秒", normalizedConcurrency, normalizedTimeout)))
    }
    fmt.Println()

    // 3. 并发测试
    fmt.Printf("  %s %s\n\n", cyanB("→"), white("开始并发测试..."))
    results := tester.TestNodes(nodes, normalizedConcurrency, normalizedTimeout)

    // 4. 显示结果
    display.ShowResults(results, verbose)
}

// printBanner 打印欢迎横幅
func printBanner() {
    banner := `
  ╔═══════════════════════════════════════════════════════════════╗
  ║                                                               ║
  ║    %s                                        ║
  ║    %s                       ║
  ║                                                               ║
  ╚═══════════════════════════════════════════════════════════════╝
`
    title := cyanB("🚀 代理节点测速工具")
    subtitle := gray("macOS 平台代理节点批量测速与分析工具")
    fmt.Printf(banner, title, subtitle)
    fmt.Println()
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
