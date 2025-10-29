# 快速开始

## 一分钟上手

```bash
# 1. 编译
go build -o proxy-tester

# 2. 测试订阅
./proxy-tester test --url "你的订阅链接"

# 3. 查看结果（按延迟排序）
```

## 常用命令

### 基础测试
```bash
./proxy-tester test --url "订阅链接"
```

### 快速扫描（50并发，2秒超时）
```bash
./proxy-tester test --url "订阅链接" -c 50 -t 2
```

### 精确测试（5并发，10秒超时）
```bash
./proxy-tester test --url "订阅链接" -c 5 -t 10
```

### 推荐配置（20并发，3秒超时）
```bash
./proxy-tester test --url "订阅链接" -c 20 -t 3
```

### 调试模式（显示详细日志）
```bash
./proxy-tester test --url "订阅链接" -v
```

## 参数说明

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--url` | `-u` | 必需 | 订阅链接 URL |
| `--concurrency` | `-c` | 10 | 并发测试数量 |
| `--timeout` | `-t` | 5 | 超时时间（秒） |
| `--verbose` | `-v` | false | 显示详细日志 |

## 结果解读

### 状态说明

- ✅ **成功**: 节点可连接（延迟显示为具体数值）
- ⏱️ **超时**: 连接超时，节点可能不可用
- ❌ **失败**: 连接失败

### 延迟说明

- **TCP延迟**: TCP 连接建立时间
- **真实延迟**: 代理协议握手时间（更准确）

### 排序规则

结果按以下优先级排序：
1. 成功的节点排在前面
2. 成功的节点按真实延迟从低到高排序
3. 失败的节点排在后面

## 常见问题

### Q: 为什么有些节点显示"端口可达但代理失败"？

A: 当前版本使用简化的连接测试，可能存在误判。建议：
1. 记录这些节点
2. 在 Shadowrocket 中手动验证
3. 等待 v1.1 版本的完整协议支持

### Q: 开启 VPN 会影响测试吗？

A: 不会。程序会自动绕过系统代理，直连测试节点。

### Q: 测试很慢怎么办？

A: 调整参数：
```bash
# 增加并发数
./proxy-tester test --url "订阅链接" -c 30

# 减少超时时间
./proxy-tester test --url "订阅链接" -t 2

# 组合使用
./proxy-tester test --url "订阅链接" -c 30 -t 2
```

### Q: 如何只看成功的节点？

A: 当前版本需要手动查看表格中状态为 ✅ 的节点。v1.1 将支持自动过滤。

## 下一步

- 查看 [README.md](README.md) 了解完整功能
- 查看 [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) 解决问题
- 查看 [docs/ROADMAP.md](docs/ROADMAP.md) 了解未来规划

## 获取帮助

```bash
# 查看帮助
./proxy-tester --help
./proxy-tester test --help
```
