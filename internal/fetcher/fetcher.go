package fetcher

import (
    "compress/gzip"
    "crypto/tls"
    "encoding/base64"
    "fmt"
    "io"
    "net"
    "net/http"
    "strings"
    "syscall"
    "time"
)

// FetchSubscription 从URL下载订阅内容并解码
func FetchSubscription(url string) (string, error) {
    // 创建直连 dialer，完全绕过系统代理
    dialer := &net.Dialer{
        Timeout:   30 * time.Second,
        KeepAlive: 30 * time.Second,
        // 使用 Control 函数确保绕过系统代理
        Control: func(network, address string, c syscall.RawConn) error {
            return nil
        },
    }

    // 创建自定义的 HTTP 客户端（绕过系统代理）
    client := &http.Client{
        Timeout: 30 * time.Second,
        Transport: &http.Transport{
            // 明确禁用所有代理（包括环境变量中的代理）
            Proxy:       http.ProxyFromEnvironment, // 先设置默认值
            DialContext: dialer.DialContext,
            TLSClientConfig: &tls.Config{
                InsecureSkipVerify: true,
            },
            DisableKeepAlives:     false,
            MaxIdleConns:          10,
            IdleConnTimeout:       30 * time.Second,
            TLSHandshakeTimeout:   10 * time.Second,
            ExpectContinueTimeout: 1 * time.Second,
        },
    }

    // 覆盖 Proxy 设置，确保完全绕过代理
    client.Transport.(*http.Transport).Proxy = func(req *http.Request) (*http.URL, error) {
        // 返回 nil 表示不使用任何代理，直连
        return nil, nil
    }

    // 创建请求并设置 User-Agent
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return "", fmt.Errorf("创建请求失败: %w", err)
    }

    // 设置常见的浏览器 User-Agent 以避免被拦截
    req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
    req.Header.Set("Accept", "*/*")
    req.Header.Set("Accept-Encoding", "gzip, deflate")
    req.Header.Set("Connection", "keep-alive")

    resp, err := client.Do(req)
    if err != nil {
        return "", fmt.Errorf("下载失败: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("HTTP状态码错误: %d", resp.StatusCode)
    }

    // 处理 gzip 压缩
    var reader io.Reader = resp.Body
    if resp.Header.Get("Content-Encoding") == "gzip" {
        gzipReader, err := gzip.NewReader(resp.Body)
        if err != nil {
            return "", fmt.Errorf("gzip解压失败: %w", err)
        }
        defer gzipReader.Close()
        reader = gzipReader
    }

    body, err := io.ReadAll(reader)
    if err != nil {
        return "", fmt.Errorf("读取响应失败: %w", err)
    }

    if len(body) == 0 {
        return "", fmt.Errorf("订阅内容为空")
    }

    // 尝试Base64解码
    decoded, err := base64.StdEncoding.DecodeString(string(body))
    if err != nil {
        // 尝试 RawStdEncoding
        decoded, err = base64.RawStdEncoding.DecodeString(string(body))
        if err != nil {
            // 如果解码失败，可能内容本身就是明文
            decoded = body
        }
    }

    content := strings.TrimSpace(string(decoded))
    if content == "" {
        return "", fmt.Errorf("订阅内容为空")
    }

    return content, nil
}
