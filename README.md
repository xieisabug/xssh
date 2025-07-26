# XSSH - SSH Connection Manager

一个基于 TUI (Terminal User Interface) 的 SSH 连接管理工具，使用 Go 语言和 Charm 的 Bubbletea 框架构建。

## 功能特性

- ✅ 读取和解析 SSH config 文件
- ✅ 面板式列表展示所有 SSH 主机配置
- ✅ 实时搜索和过滤主机列表（按 `:` 进入搜索模式）
- ✅ 连接到选定的 SSH 主机（Enter 键）
- ✅ 复制 SSH 命令到剪贴板（c 键）
- ✅ 添加新的 SSH 主机配置（a 键 + 完整设置流程）
- ✅ 编辑现有主机配置（e 键）
- ✅ 删除主机配置（d 键 + 确认）
- ✅ SSH 密钥文件选择功能
- ✅ 自动 SSH 密钥生成和配置
- ✅ 密码连接测试和密钥部署
- ✅ 完整的 ssh-copy-id 功能集成
- 🚧 文件复制操作 (SCP)

## 安装和运行

```bash
# 构建项目
go build -o xssh

# 运行
./xssh
```

## 使用方法

### 导航和操作

**普通模式:**
- `↑/k`: 上移选择
- `↓/j`: 下移选择
- `Enter`: 连接选定主机
- `c`: 复制 SSH 命令到剪贴板
- `a`: 添加新主机
- `e`: 编辑选定主机
- `d`: 删除选定主机（需确认）
- `:`: 进入搜索模式
- `ESC`: 清空过滤条件
- `q` 或 `Ctrl+C`: 退出程序

**搜索模式:**
- 直接输入字符: 实时过滤主机列表
- `Backspace`: 删除过滤字符
- `ESC`: 退出搜索模式
- `Enter`: 确认搜索并退出搜索模式
- `Ctrl+C`: 退出程序

**添加/编辑模式:**
- `Tab` 或 `↓`: 下一个字段
- `Shift+Tab` 或 `↑`: 上一个字段
- 直接输入: 编辑当前字段内容
- `Backspace`: 删除字符
- `Enter`: 进入下一步或保存
- `ESC`: 取消并返回列表

**密码输入模式:**
- 直接输入: 输入密码（显示为 *）
- `Backspace`: 删除字符
- `Enter`: 开始连接测试和设置
- `ESC`: 返回表单

**连接测试模式:**
- 程序自动测试连接并设置 SSH 密钥
- `Enter`: 完成设置并保存（测试成功后）
- `ESC`: 取消设置

**认证方式选择:**
- `1`: 选择密码认证
- `2`: 选择 SSH 密钥认证
- `ESC`: 返回表单

**SSH 密钥选择:**
- `↑/k`: 上移选择
- `↓/j`: 下移选择
- `Enter`: 选择密钥
- `ESC`: 返回认证方式选择

**删除确认:**
- `Y`: 确认删除
- `N` 或 `ESC`: 取消删除

## 项目结构

```
xssh/
├── main.go                 # 程序入口
├── internal/
│   ├── config/
│   │   └── ssh.go         # SSH config 解析
│   ├── ui/
│   │   ├── model.go       # TUI 主要逻辑和状态管理
│   │   └── views.go       # 不同界面的渲染函数
│   └── ssh/
│       └── client.go      # SSH 连接和命令处理
├── pkg/                   # 公共包 (待使用)
└── cmd/                   # 命令行工具 (待使用)
```

## 待完成功能

1. 文件传输 (SCP) 功能
2. 配置文件备份和恢复
3. 主机分组管理
4. 连接历史记录
5. 批量操作功能

## 依赖

- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI 框架
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - 样式库
- [golang.org/x/crypto/ssh](https://golang.org/x/crypto/ssh) - SSH 客户端库
- [atotto/clipboard](https://github.com/atotto/clipboard) - 剪贴板操作