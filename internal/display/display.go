package display

import (
    "fmt"
    "proxy-tester/internal/tester"
    "sort"
    "strings"

    "github.com/fatih/color"
)

// 定义颜色函数
var (
    // 标题和边框
    cyan    = color.New(color.FgCyan).SprintFunc()
    cyanB   = color.New(color.FgCyan, color.Bold).SprintFunc()
    
    // 成功状态
    green   = color.New(color.FgGreen).SprintFunc()
    greenB  = color.New(color.FgGreen, color.Bold).SprintFunc()
    
    // 警告状态
    yellow  = color.New(color.FgYellow).SprintFunc()
    yellowB = color.New(color.FgYellow, color.Bold).SprintFunc()
    
    // 错误状态
    red     = color.New(color.FgRed).SprintFunc()
    redB    = color.New(color.FgRed, color.Bold).SprintFunc()
    
    // 中性状态
    white   = color.New(color.FgWhite).SprintFunc()
    whiteB  = color.New(color.FgWhite, color.Bold).SprintFunc()
    gray    = color.New(color.FgHiBlack).SprintFunc()
    
    // 高亮
    magenta = color.New(color.FgMagenta).SprintFunc()
    magentaB = color.New(color.FgMagenta, color.Bold).SprintFunc()
)

// ShowResults 显示测试结果
func ShowResults(results []*tester.TestResult, verbose bool) {
    if len(results) == 0 {
        fmt.Println(yellow("⚠️  没有测试结果"))
        return
    }

    // 按真实延迟排序 (低到高)
    sortResults(results)

    // 统计数据
    stats := calculateStats(results)

    // 打印分隔线
    printSeparator("═")
    
    // 打印统计摘要
    printSummary(stats)
    
    printSeparator("═")
    
    // 打印表头
    printTableHeader()

    // 打印每个结果
    for i, result := range results {
        printTableRow(result, i+1)
    }

    // 打印表格底部边框
    printSeparator("─")
    fmt.Println()

    // 打印延迟分布
    if stats.Success > 0 {
        printLatencyDistribution(results)
        fmt.Println()
    }

    // 打印最快节点
    if stats.Success > 0 {
        printTopNodes(results, 5)
        fmt.Println()
    }

    // 在 verbose 模式下显示失败节点的详细错误信息
    if verbose && stats.Failed > 0 {
        printFailedNodesDetail(results)
    }
}

// calculateStats 计算统计数据
func calculateStats(results []*tester.TestResult) *Stats {
    stats := &Stats{
        Total: len(results),
    }

    var totalLatency int64
    var validLatencyCount int
    
    for _, r := range results {
        if r.IsSuccess() {
            stats.Success++
            latency := r.ProxyLatency
            if latency <= 0 {
                latency = r.TCPLatency
            }
            if latency > 0 {
                totalLatency += int64(latency)
                validLatencyCount++
                
                if stats.MinLatency == 0 || latency < stats.MinLatency {
                    stats.MinLatency = latency
                    stats.FastestNode = r
                }
                if latency > stats.MaxLatency {
                    stats.MaxLatency = latency
                    stats.SlowestNode = r
                }
            }
        } else {
            stats.Failed++
        }
    }

    if validLatencyCount > 0 {
        stats.AvgLatency = int(totalLatency / int64(validLatencyCount))
    }

    stats.SuccessRate = float64(stats.Success) / float64(stats.Total) * 100

    return stats
}

// Stats 统计信息
type Stats struct {
    Total        int
    Success      int
    Failed       int
    SuccessRate  float64
    AvgLatency   int
    MinLatency   int
    MaxLatency   int
    FastestNode  *tester.TestResult
    SlowestNode  *tester.TestResult
}

// printSummary 打印统计摘要
func printSummary(stats *Stats) {
    fmt.Printf("\n  %s\n\n", cyanB("📊 测试结果统计"))
    
    // 总数
    fmt.Printf("  总节点数: %s", whiteB(fmt.Sprintf("%d", stats.Total)))
    
    // 成功
    successStr := fmt.Sprintf("%d (%.1f%%)", stats.Success, stats.SuccessRate)
    if stats.SuccessRate >= 80 {
        fmt.Printf("  │  成功: %s", greenB(successStr))
    } else if stats.SuccessRate >= 50 {
        fmt.Printf("  │  成功: %s", yellowB(successStr))
    } else {
        fmt.Printf("  │  成功: %s", redB(successStr))
    }
    
    // 失败
    if stats.Failed > 0 {
        fmt.Printf("  │  失败: %s", red(fmt.Sprintf("%d", stats.Failed)))
    } else {
        fmt.Printf("  │  失败: %s", gray(fmt.Sprintf("%d", stats.Failed)))
    }
    fmt.Println()

    // 延迟统计
    if stats.Success > 0 {
        fmt.Printf("\n  %s  ", "⚡")
        fmt.Printf("平均延迟: %s", formatLatencyWithColor(stats.AvgLatency))
        fmt.Printf("  │  最快: %s", formatLatencyWithColor(stats.MinLatency))
        fmt.Printf("  │  最慢: %s", formatLatencyWithColor(stats.MaxLatency))
        fmt.Println()
    }
    fmt.Println()
}

// printTableHeader 打印表头
func printTableHeader() {
    printSeparator("─")
    
    header := fmt.Sprintf("│ %s │ %s │ %s │ %s │ %s │ %s │",
        centerString("序号", 4),
        padRight("节点名称", 35),
        padRight("服务器地址", 22),
        centerString("协议", 8),
        centerString("TCP延迟", 9),
        centerString("真实延迟", 10),
    )
    
    fmt.Println(cyan(header))
    printSeparator("─")
}

// printTableRow 打印单行结果
func printTableRow(result *tester.TestResult, index int) {
    // 序号
    indexStr := fmt.Sprintf("%d", index)
    
    // 节点名称
    name := result.Node.Name
    if name == "" {
        name = "未命名"
    }
    name = truncateString(name, 35)
    
    // 服务器地址
    address := truncateString(result.Node.Address(), 22)
    
    // 协议类型（带颜色）
    protocol := formatProtocol(result.Node.Type)
    
    // TCP延迟
    tcpLatency := formatLatency(result.TCPLatency)
    
    // 真实延迟（带颜色）
    proxyLatency := formatLatency(result.ProxyLatency)
    
    // 根据状态设置行颜色
    var rowColor func(a ...interface{}) string
    if result.IsSuccess() {
        rowColor = white
        // 根据延迟着色名称
        latency := result.ProxyLatency
        if latency <= 0 {
            latency = result.TCPLatency
        }
        name = colorizeByLatency(name, latency)
        proxyLatency = colorizeLatencyValue(proxyLatency, result.ProxyLatency)
        tcpLatency = colorizeLatencyValue(tcpLatency, result.TCPLatency)
    } else {
        rowColor = gray
        name = gray(name)
        address = gray(address)
    }
    
    // 状态图标
    statusIcon := formatStatusIcon(result.Status)
    
    row := fmt.Sprintf("│ %s │ %s │ %s │ %s │ %s │ %s %s │",
        rowColor(centerString(indexStr, 4)),
        name,
        rowColor(padRight(address, 22)),
        protocol,
        tcpLatency,
        proxyLatency,
        statusIcon,
    )
    
    fmt.Println(row)
}

// printSeparator 打印分隔线
func printSeparator(char string) {
    // 计算总宽度: 序号(6) + 名称(37) + 地址(24) + 协议(10) + TCP(11) + 真实(12) = 100
    width := 108
    fmt.Println(gray(strings.Repeat(char, width)))
}

// formatProtocol 格式化协议类型
func formatProtocol(proxyType interface{}) string {
    protocolStr := fmt.Sprintf("%v", proxyType)
    switch protocolStr {
    case "vless":
        return centerString(magenta("VLESS"), 8)
    case "vmess":
        return centerString(cyan("VMess"), 8)
    case "ss":
        return centerString(yellow("SS"), 8)
    default:
        return centerString(gray("Unknown"), 8)
    }
}

// formatStatusIcon 格式化状态图标
func formatStatusIcon(status string) string {
    switch status {
    case "成功":
        return greenB("✓")
    case "超时":
        return yellow("⏱")
    case "失败":
        return red("✗")
    case "端口可达但连接失败":
        return yellow("⚠")
    default:
        return gray("?")
    }
}

// formatLatency 格式化延迟显示
func formatLatency(latency int) string {
    if latency < 0 {
        return centerString("-", 9)
    }
    return centerString(fmt.Sprintf("%dms", latency), 9)
}

// formatLatencyWithColor 格式化延迟并根据值着色
func formatLatencyWithColor(latency int) string {
    latencyStr := fmt.Sprintf("%dms", latency)
    return colorizeByLatency(latencyStr, latency)
}

// colorizeByLatency 根据延迟值着色
func colorizeByLatency(text string, latency int) string {
    if latency < 0 {
        return gray(text)
    } else if latency < 100 {
        return greenB(text)
    } else if latency < 300 {
        return yellow(text)
    } else if latency < 500 {
        return yellowB(text)
    } else {
        return red(text)
    }
}

// colorizeLatencyValue 根据延迟值着色延迟数值
func colorizeLatencyValue(text string, latency int) string {
    if latency < 0 {
        return centerString(gray(text), 9)
    }
    colored := colorizeByLatency(text, latency)
    return centerString(colored, 9)
}

// printLatencyDistribution 打印延迟分布
func printLatencyDistribution(results []*tester.TestResult) {
    fmt.Printf("  %s\n\n", cyanB("📈 延迟分布"))
    
    // 分组统计
    ranges := []struct {
        min   int
        max   int
        label string
        color func(a ...interface{}) string
    }{
        {0, 100, "极快 (<100ms)", greenB},
        {100, 200, "快速 (100-200ms)", green},
        {200, 300, "良好 (200-300ms)", yellow},
        {300, 500, "较慢 (300-500ms)", yellowB},
        {500, 99999, "很慢 (>500ms)", red},
    }
    
    counts := make([]int, len(ranges))
    maxCount := 0
    successfulCount := 0
    
    for _, r := range results {
        if !r.IsSuccess() {
            continue
        }
        latency := r.ProxyLatency
        if latency <= 0 {
            latency = r.TCPLatency
        }
        if latency < 0 {
            continue
        }
        
        successfulCount++
        
        for i, rng := range ranges {
            if latency >= rng.min && latency < rng.max {
                counts[i]++
                if counts[i] > maxCount {
                    maxCount = counts[i]
                }
                break
            }
        }
    }
    
    // 打印柱状图
    barWidth := 40
    for i, rng := range ranges {
        count := counts[i]
        if count == 0 {
            continue
        }
        
        // 使用成功节点数作为百分比计算的分母，避免除零错误
        percentage := 0.0
        if successfulCount > 0 {
            percentage = float64(count) / float64(successfulCount) * 100
        }
        
        barLen := int(float64(count) / float64(maxCount) * float64(barWidth))
        if barLen == 0 && count > 0 {
            barLen = 1
        }
        
        bar := strings.Repeat("█", barLen)
        label := fmt.Sprintf("%-20s", rng.label)
        countStr := fmt.Sprintf("%2d (%.1f%%)", count, percentage)
        
        fmt.Printf("  %s %s %s\n", 
            rng.color(label), 
            rng.color(bar),
            rng.color(countStr))
    }
}

// printTopNodes 打印最快的节点
func printTopNodes(results []*tester.TestResult, topN int) {
    fmt.Printf("  %s\n\n", cyanB("🏆 最快节点 TOP 5"))
    
    successResults := make([]*tester.TestResult, 0)
    for _, r := range results {
        if r.IsSuccess() {
            successResults = append(successResults, r)
        }
    }
    
    if len(successResults) == 0 {
        return
    }
    
    if len(successResults) > topN {
        successResults = successResults[:topN]
    }
    
    for i, r := range successResults {
        latency := r.ProxyLatency
        if latency <= 0 {
            latency = r.TCPLatency
        }
        
        medal := ""
        switch i {
        case 0:
            medal = "🥇"
        case 1:
            medal = "🥈"
        case 2:
            medal = "🥉"
        default:
            medal = fmt.Sprintf("%d.", i+1)
        }
        
        name := r.Node.Name
        if name == "" {
            name = "未命名"
        }
        name = truncateString(name, 40)
        
        fmt.Printf("  %s  %-42s %s  %s\n", 
            medal,
            whiteB(name),
            formatLatencyWithColor(latency),
            gray(r.Node.Address()))
    }
}

// printFailedNodesDetail 打印失败节点详细信息
func printFailedNodesDetail(results []*tester.TestResult) {
    failedCount := 0
    for _, r := range results {
        if !r.IsSuccess() {
            failedCount++
        }
    }
    
    if failedCount == 0 {
        return
    }
    
    fmt.Printf("  %s\n\n", redB("❌ 失败节点详细信息"))
    printSeparator("─")
    
    for i, result := range results {
        if !result.IsSuccess() {
            name := result.Node.Name
            if name == "" {
                name = "未命名"
            }
            
            fmt.Printf("  %s %s\n", red("▸"), whiteB(name))
            fmt.Printf("    地址: %s\n", gray(result.Node.Address()))
            fmt.Printf("    协议: %s\n", gray(fmt.Sprintf("%v", result.Node.Type)))
            if result.Error != "" {
                fmt.Printf("    错误: %s\n", red(result.Error))
            }
            if i < len(results)-1 {
                fmt.Println()
            }
        }
    }
    fmt.Println()
}

// sortResults 按真实延迟排序
func sortResults(results []*tester.TestResult) {
    sort.SliceStable(results, func(i, j int) bool {
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

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
    runes := []rune(s)
    if len(runes) <= maxLen {
        return s
    }
    return string(runes[:maxLen-3]) + "..."
}

// padRight 右填充字符串到指定长度
func padRight(s string, length int) string {
    runes := []rune(s)
    runeLen := len(runes)
    if runeLen >= length {
        return s
    }
    return s + strings.Repeat(" ", length-runeLen)
}

// centerString 居中字符串
func centerString(s string, width int) string {
    // 移除 ANSI 颜色代码来计算实际长度
    actualLen := len([]rune(stripAnsi(s)))
    
    if actualLen >= width {
        return s
    }
    
    leftPad := (width - actualLen) / 2
    rightPad := width - actualLen - leftPad
    
    return strings.Repeat(" ", leftPad) + s + strings.Repeat(" ", rightPad)
}

// stripAnsi 移除 ANSI 颜色代码
func stripAnsi(s string) string {
    // 简单的 ANSI 代码移除（用于长度计算）
    inEscape := false
    result := ""
    for _, r := range s {
        if r == '\x1b' {
            inEscape = true
            continue
        }
        if inEscape {
            if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
                inEscape = false
            }
            continue
        }
        result += string(r)
    }
    return result
}
