package main

import (
    "os"
    "proxy-tester/cmd"
)

func main() {
    // 清除所有代理相关的环境变量
    // 确保程序完全绕过系统代理设置（如 Shadowrocket 等）
    clearProxyEnv()

    cmd.Execute()
}

// clearProxyEnv 清除所有代理相关的环境变量
func clearProxyEnv() {
    proxyEnvVars := []string{
        "HTTP_PROXY",
        "http_proxy",
        "HTTPS_PROXY",
        "https_proxy",
        "ALL_PROXY",
        "all_proxy",
        "NO_PROXY",
        "no_proxy",
    }

    for _, env := range proxyEnvVars {
        os.Unsetenv(env)
    }
}
