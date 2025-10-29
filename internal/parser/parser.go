package parser

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
)

// ParseNodes 解析订阅内容中的所有节点
func ParseNodes(content string, verbose bool) ([]*Node, error) {
	lines := strings.Split(content, "\n")
	var nodes []*Node

	if verbose {
		log.Printf("📋 开始解析，共 %d 行内容\n", len(lines))
	}

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		node := parseNode(line)
		if node != nil && node.Type != ProxyTypeUnknown {
			nodes = append(nodes, node)
			if verbose {
				log.Printf("✅ [%d] 解析成功: %s (%s:%s)\n", i+1, node.Name, node.Server, node.Port)
			}
		} else if verbose {
			log.Printf("⚠️  [%d] 跳过未知格式: %s\n", i+1, line[:min(50, len(line))])
		}
	}

	if verbose {
		log.Printf("\n📊 解析完成: 成功 %d 个节点\n", len(nodes))
	}

	return nodes, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// parseNode 解析单个节点链接
func parseNode(line string) *Node {
	line = strings.TrimSpace(line)
	
	if strings.HasPrefix(line, "vless://") {
		return parseVLESS(line)
	} else if strings.HasPrefix(line, "vmess://") {
		return parseVMess(line)
	} else if strings.HasPrefix(line, "ss://") {
		return parseShadowsocks(line)
	}

	return &Node{Type: ProxyTypeUnknown, Raw: line}
}

// parseVLESS 解析VLESS链接
// 格式: vless://uuid@server:port?params#name
func parseVLESS(link string) *Node {
	node := &Node{
		Type: ProxyTypeVLESS,
		Raw:  link,
	}

	// 移除协议前缀
	link = strings.TrimPrefix(link, "vless://")

	// 分离名称
	parts := strings.SplitN(link, "#", 2)
	if len(parts) == 2 {
		name, _ := url.QueryUnescape(parts[1])
		node.Name = name
		link = parts[0]
	}

	// 分离参数
	parts = strings.SplitN(link, "?", 2)
	if len(parts) == 2 {
		params, err := url.ParseQuery(parts[1])
		if err == nil {
			node.Network = params.Get("type")
			if node.Network == "" {
				node.Network = "tcp"
			}
			security := params.Get("security")
			if security == "tls" || security == "reality" {
				node.TLS = true
			}
		}
		link = parts[0]
	}

	// 解析 uuid@server:port
	parts = strings.SplitN(link, "@", 2)
	if len(parts) != 2 {
		return node
	}

	node.UUID = parts[0]

	// 解析 server:port (处理IPv6地址)
	serverPart := parts[1]
	
	// 检查是否是IPv6地址
	if strings.HasPrefix(serverPart, "[") {
		// IPv6格式: [2606:4700:440a::601f:11eb]:443
		closeBracket := strings.Index(serverPart, "]")
		if closeBracket > 0 {
			node.Server = serverPart[:closeBracket+1]
			if len(serverPart) > closeBracket+2 && serverPart[closeBracket+1] == ':' {
				node.Port = serverPart[closeBracket+2:]
			}
		}
	} else {
		// IPv4或域名格式: server:port
		serverPort := strings.SplitN(serverPart, ":", 2)
		if len(serverPort) == 2 {
			node.Server = serverPort[0]
			node.Port = serverPort[1]
		} else if len(serverPort) == 1 {
			node.Server = serverPort[0]
			node.Port = "443" // 默认端口
		}
	}

	// 验证必要字段
	if node.Server == "" || node.Port == "" || node.UUID == "" {
		node.Type = ProxyTypeUnknown
	}

	return node
}

// parseVMess 解析VMess链接
// 格式: vmess://base64(json)
func parseVMess(link string) *Node {
	node := &Node{
		Type: ProxyTypeVMess,
		Raw:  link,
	}

	// 移除协议前缀
	link = strings.TrimPrefix(link, "vmess://")

	// Base64解码
	decoded, err := base64.StdEncoding.DecodeString(link)
	if err != nil {
		// 尝试RawStdEncoding
		decoded, err = base64.RawStdEncoding.DecodeString(link)
		if err != nil {
			return node
		}
	}

	// 解析JSON
	var config map[string]interface{}
	if err := json.Unmarshal(decoded, &config); err != nil {
		return node
	}

	// 提取字段
	if v, ok := config["ps"].(string); ok {
		node.Name = v
	}
	if v, ok := config["add"].(string); ok {
		node.Server = v
	}
	if v, ok := config["port"].(float64); ok {
		node.Port = fmt.Sprintf("%.0f", v)
	} else if v, ok := config["port"].(string); ok {
		node.Port = v
	}
	if v, ok := config["id"].(string); ok {
		node.UUID = v
	}
	if v, ok := config["net"].(string); ok {
		node.Network = v
	}
	if v, ok := config["tls"].(string); ok && v == "tls" {
		node.TLS = true
	}

	return node
}

// parseShadowsocks 解析Shadowsocks链接
// 格式: ss://base64(method:password)@server:port#name
func parseShadowsocks(link string) *Node {
	node := &Node{
		Type: ProxyTypeShadowsocks,
		Raw:  link,
	}

	// 移除协议前缀
	link = strings.TrimPrefix(link, "ss://")

	// 分离名称
	parts := strings.SplitN(link, "#", 2)
	if len(parts) == 2 {
		name, _ := url.QueryUnescape(parts[1])
		node.Name = name
		link = parts[0]
	}

	// 分离 userinfo 和 server:port
	parts = strings.SplitN(link, "@", 2)
	if len(parts) != 2 {
		return node
	}

	userInfo := parts[0]
	serverPort := parts[1]

	// 解码 userinfo (method:password)
	decoded, err := base64.StdEncoding.DecodeString(userInfo)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(userInfo)
		if err != nil {
			// 可能已经是明文
			decoded = []byte(userInfo)
		}
	}

	// 解析 method:password
	methodPass := strings.SplitN(string(decoded), ":", 2)
	if len(methodPass) == 2 {
		node.Method = methodPass[0]
		node.Password = methodPass[1]
	}

	// 解析 server:port
	serverPortParts := strings.SplitN(serverPort, ":", 2)
	if len(serverPortParts) == 2 {
		node.Server = serverPortParts[0]
		node.Port = serverPortParts[1]
	}

	return node
}
