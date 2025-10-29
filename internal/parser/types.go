package parser

// ProxyType 代理协议类型
type ProxyType string

const (
	ProxyTypeVLESS      ProxyType = "vless"
	ProxyTypeVMess      ProxyType = "vmess"
	ProxyTypeShadowsocks ProxyType = "ss"
	ProxyTypeUnknown    ProxyType = "unknown"
)

// Node 代理节点信息
type Node struct {
	Type     ProxyType // 协议类型
	Name     string    // 节点名称
	Server   string    // 服务器地址
	Port     string    // 端口
	UUID     string    // UUID (VLESS/VMess)
	Password string    // 密码 (Shadowsocks)
	Method   string    // 加密方式 (Shadowsocks)
	Network  string    // 传输协议 (tcp/ws/grpc等)
	TLS      bool      // 是否启用TLS
	Raw      string    // 原始链接
}

// Address 返回完整的服务器地址
func (n *Node) Address() string {
	return n.Server + ":" + n.Port
}
