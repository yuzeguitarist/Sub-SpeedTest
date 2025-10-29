package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "proxy-tester",
	Short: "macOS平台代理节点批量测速工具",
	Long:  `一个用于测试代理节点连通性和延迟的命令行工具，支持VLESS、VMess和Shadowsocks协议。`,
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
