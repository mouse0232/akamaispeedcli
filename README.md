# Atlanta CLI Speed Test

[![Test](https://github.com/{your-username}/atlanta/actions/workflows/test.yml/badge.svg)](https://github.com/{your-username}/atlanta/actions/workflows/test.yml)
[![Release](https://github.com/{your-username}/atlanta/actions/workflows/release.yml/badge.svg)](https://github.com/{your-username}/atlanta/actions/workflows/release.yml)

Atlanta是一个基于Go语言开发的命令行网络测速工具，灵感来源于经典的网页版测速工具。它能够测试网络的下载速度、上传速度、延迟和抖动等指标。

## 功能特点

- 测试下载速度
- 测试上传速度
- 测试网络延迟(Ping)
- 测试延迟抖动(Jitter)
- 获取客户端IP和ISP信息
- 支持多种输出格式（标准、JSON、简洁模式）
- 可配置的测试参数
- 多服务器支持，自动选择最近服务器
- 可自定义并发线程数

## 安装

确保你的系统已安装Go语言环境（1.16或更高版本）。

```bash
# 克隆项目
git clone <repository-url>
cd atlanta

# 下载依赖
go mod tidy

# 编译
go build -o atlanta .

# 或直接运行
go run .
```

## 使用方法

### 基本测速（自动选择最近服务器）

```bash
./atlanta
```

### 查看帮助信息

```bash
./atlanta --help
```

### 列出可用服务器

```bash
./atlanta --list
```

### 指定服务器测速

```bash
./atlanta --server 5
```

### 自动选择最近服务器测速

```bash
./atlanta --server -1
```

### JSON格式输出

```bash
./atlanta --json
```

### 简洁模式输出

```bash
./atlanta --simple
```

### 设置并发线程数

```bash
# 设置下载和上传测试均为10个并发线程
./atlanta -c 10

# 单线程模式
./atlanta -c 1
```

## 命令行参数

- `-help`: 显示帮助信息
- `-list`: 列出所有可用的测速服务器
- `-json`: 以JSON格式输出结果
- `-simple`: 仅输出结果，不显示详细信息
- `-c int`: 设置并发线程数（0为默认值，1为单线程模式）
- `-server int`: 指定服务器ID（-1表示自动选择最近服务器，默认为自动选择）

## 配置选项

Atlanta支持多种配置选项，可以根据需要调整测速行为：

- `TestOrder`: 测试项目顺序，默认为"IP_D_U"（获取IP、下载测速、上传测速）
- `TimeDLMax`: 最大下载测试时间（秒），默认15秒
- `TimeULMax`: 最大上传测试时间（秒），默认15秒
- `CountPing`: Ping测试次数，默认10次
- `XhrDLMultistream`: 下载并发连接数，默认6个
- `XhrULMultistream`: 上传并发连接数，默认3个

## 技术实现

本工具参考了经典的网页版测速工具的核心算法，使用以下技术实现：

1. 使用HTTP协议与测速服务端通信
2. 多协程并发测试提高准确性
3. 支持自动补偿网络开销
4. 提供丰富的命令行接口
5. 自动选择延迟最低的服务器进行测速
6. 为所有网络请求设置超时，防止程序卡死

## 注意事项

1. 需要可访问互联网才能进行测速
2. 测速过程中会消耗一定的网络流量
3. 测速结果可能受网络环境、硬件性能等因素影响
4. 程序会自动为所有HTTP请求设置超时，防止因网络问题导致程序无响应

## 开源许可

本项目仅供学习和参考使用。