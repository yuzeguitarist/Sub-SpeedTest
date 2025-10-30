package cmd

import (
    "fmt"
    "os"

    "github.com/fatih/color"
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "proxy-tester",
    Short: "macOS平台代理节点批量测速工具",
    Long: color.CyanString(`
  ╔═══════════════════════════════════════════════════════════════╗
  ║  🚀 代理节点测速工具 - Proxy Tester                          ║
  ╚═══════════════════════════════════════════════════════════════╝

  一个用于测试代理节点连通性和延迟的命令行工具
  
  `) + color.WhiteString(`支持的协议:`) + `
    • VLESS  - 新一代轻量级代理协议
    • VMess  - V2Ray 传统代理协议
    • Shadowsocks (SS) - 经典代理协议
  
  ` + color.WhiteString(`主要特性:`) + `
    • 🔥 并发测试，速度快
    • 📊 详细的统计分析
    • 🎨 美观的终端UI
    • 🔧 自动绕过系统代理
    • ⚡ TCP和真实代理延迟双重检测
  
  ` + color.YellowString(`使用示例:`) + `
    proxy-tester test -u <订阅URL>
    proxy-tester test -u <订阅URL> -c 20 -t 10 -v
  
  ` + color.HiBlackString(`更多信息请访问: https://github.com/proxy-node-tester/proxy-tester`),
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}

func init() {
    rootCmd.AddCommand(testCmd)
}
