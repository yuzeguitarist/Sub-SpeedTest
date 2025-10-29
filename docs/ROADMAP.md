# 功能路线图

## 已完成功能 ✅

### v1.0.0 - 基础测速功能
- [x] 订阅链接下载和解析
- [x] 支持 VLESS、VMess、Shadowsocks 协议
- [x] 并发测速引擎
- [x] TCP Ping + 代理连接测试
- [x] 结果排序和表格化输出
- [x] 详细日志模式 (`-v`)
- [x] 绕过系统代理进行测试

## 计划中功能 📋

### v1.1.0 - 订阅过滤和生成

#### 1. 自动过滤功能
**目标**: 自动排除超时和失败的节点，只保留可用节点

**实现方案**:
- 在测试完成后，自动过滤出状态为"成功"的节点
- 按延迟排序，可选择保留前 N 个最快的节点
- 支持配置最小延迟阈值（如只保留延迟 < 500ms 的节点）

**CLI 参数**:
```bash
--filter-success     # 只保留测试成功的节点
--max-latency 500    # 只保留延迟小于 500ms 的节点
--top N              # 只保留延迟最低的前 N 个节点
```

#### 2. 订阅生成功能
**目标**: 将过滤后的节点重新编码为订阅格式

**实现方案**:
- 将过滤后的节点列表转换回原始协议链接格式
- Base64 编码生成新的订阅内容
- 支持本地文件输出和 HTTP 服务器托管

**CLI 命令**:
```bash
# 导出到本地文件
proxy-tester export --url "订阅链接" --output filtered.txt --filter-success

# 启动 HTTP 服务器托管订阅
proxy-tester serve --url "订阅链接" --port 8080 --filter-success --auto-refresh 1h
```

**输出示例**:
```
✅ 测试完成，94 个节点中 42 个可用
📝 已过滤并生成新订阅

🌐 订阅地址: http://localhost:8080/sub
📋 直接复制此链接到代理软件使用

💡 提示: 
   - 订阅每小时自动更新一次
   - 只包含延迟 < 500ms 的节点
   - 按延迟从低到高排序
```

#### 3. 永久域名托管
**目标**: 支持将订阅托管到永久域名

**实现方案**:
- 集成云存储服务（如 Cloudflare R2、AWS S3）
- 支持自定义域名绑定
- 自动定期更新订阅内容

**配置文件** (`config.yaml`):
```yaml
server:
  port: 8080
  domain: "sub.example.com"
  
storage:
  type: "cloudflare-r2"  # 或 "local", "s3"
  bucket: "my-subscriptions"
  
filter:
  success_only: true
  max_latency: 500
  top_n: 50
  
refresh:
  interval: "1h"
  sources:
    - "https://example.com/sub1"
    - "https://example.com/sub2"
```

**启动服务**:
```bash
proxy-tester serve --config config.yaml
```

### v1.2.0 - 高级功能

#### 1. 多订阅源合并
- 支持同时测试多个订阅链接
- 自动去重相同的节点
- 合并生成统一订阅

#### 2. 节点分组
- 按地理位置分组（如香港、日本、美国）
- 按运营商分组（如移动、联通、电信）
- 生成分组订阅链接

#### 3. 历史记录和趋势分析
- 记录每次测试的结果
- 生成节点可用性趋势图
- 识别稳定性高的节点

#### 4. 通知功能
- 订阅内容更新时发送通知
- 节点大面积失效时告警
- 支持 Telegram、邮件、Webhook

### v2.0.0 - Web 界面

#### 1. Web 管理面板
- 可视化测速结果
- 在线配置过滤规则
- 订阅管理和监控

#### 2. API 接口
- RESTful API 提供订阅服务
- 支持第三方集成
- API 密钥认证

## 使用场景

### 场景 1: 个人使用
```bash
# 测试并过滤订阅
proxy-tester test --url "订阅链接" --filter-success

# 导出到本地
proxy-tester export --url "订阅链接" --output my-sub.txt --top 20
```

### 场景 2: 团队共享
```bash
# 启动服务器，团队成员共享订阅链接
proxy-tester serve --url "订阅链接" --port 8080 --filter-success --auto-refresh 30m

# 团队成员使用: http://your-server:8080/sub
```

### 场景 3: 自动化部署
```bash
# Docker 部署
docker run -d -p 8080:8080 \
  -v ./config.yaml:/app/config.yaml \
  proxy-tester serve --config /app/config.yaml

# 配合 cron 定时更新
0 */1 * * * proxy-tester export --url "订阅链接" --output /var/www/sub.txt --filter-success
```

## 技术实现

### 订阅生成模块
```go
package generator

// FilterOptions 过滤选项
type FilterOptions struct {
    SuccessOnly bool
    MaxLatency  int
    TopN        int
}

// GenerateSubscription 生成订阅内容
func GenerateSubscription(results []*tester.TestResult, opts FilterOptions) (string, error) {
    // 1. 过滤节点
    filtered := filterNodes(results, opts)
    
    // 2. 转换为原始链接
    links := nodesToLinks(filtered)
    
    // 3. Base64 编码
    content := strings.Join(links, "\n")
    encoded := base64.StdEncoding.EncodeToString([]byte(content))
    
    return encoded, nil
}
```

### HTTP 服务器模块
```go
package server

// Server 订阅服务器
type Server struct {
    Port         int
    Sources      []string
    FilterOpts   FilterOptions
    RefreshInterval time.Duration
}

// Start 启动服务器
func (s *Server) Start() error {
    // 定期刷新订阅
    go s.autoRefresh()
    
    // HTTP 路由
    http.HandleFunc("/sub", s.handleSubscription)
    http.HandleFunc("/health", s.handleHealth)
    
    return http.ListenAndServe(fmt.Sprintf(":%d", s.Port), nil)
}
```

## 贡献指南

欢迎贡献代码和想法！请查看 [CONTRIBUTING.md](CONTRIBUTING.md)

## 反馈

如有建议或需求，请提交 Issue 或 Pull Request。
