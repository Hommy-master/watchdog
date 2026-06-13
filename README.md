# watchdog

跨平台进程看门狗，兼容 Windows / Linux / macOS。通过 JSON 配置文件定时检查进程是否存活，退出后按延迟时间自动拉起。

## 配置

```json
{
	"interval": 1,
	"delay": 5,
	"apps": [
		{
			"path": "/usr/bin/demo_app",
			"workdir": "/opt/demo",
			"args": ["--mode=run", "--port=8080"]
		}
	]
}
```

| 字段 | 说明 |
|------|------|
| `interval` | 定时检查间隔（秒），默认 `1` |
| `delay` | 检测到进程退出后，延迟多少秒再拉起 |
| `apps` | 监控的进程列表 |
| `apps[].path` | 进程可执行文件绝对路径（必填） |
| `apps[].workdir` | 工作目录，可为空 |
| `apps[].args` | 启动参数，可为空 |

示例文件见 [config.json](config.json)。

## 构建与运行

```bash
go build -o watchdog .
./watchdog -config config.json
```

Windows:

```powershell
go build -o watchdog.exe .
.\watchdog.exe -config config.json
```

收到 `SIGINT` / `SIGTERM`（Windows 为 Ctrl+C）后，看门狗会停止监控并终止已拉起的子进程。

## 测试

```bash
go test ./...
```

测试会编译内置 helper 程序，覆盖配置解析、进程启停、退出检测、延迟重启等场景，在 Windows 与 Linux 上均可运行。

## 项目结构

```
.
├── main.go                      # 程序入口
├── config.json                  # 默认配置示例
├── internal/
│   ├── config/
│   │   └── config.go            # JSON 配置结构体及加载逻辑
│   └── monitor/
│       ├── monitor.go           # 核心监控逻辑
│       ├── monitor_test.go      # 监控逻辑单元测试
│       ├── alive_unix.go        # Unix 进程存活检测
│       └── alive_windows.go     # Windows 进程存活检测
└── internal/testutil/
    └── helper.go                # 测试用 helper 程序
```
