# cdf - Directory Fuzzy Finder

**cdf** (cd fuzzy) is a lightning-fast directory-only fuzzy finder that integrates with autocd-go for seamless navigation. Finally, a fuzzy finder that's faster because it ignores files and focuses on what you actually want to navigate to: directories.

> No file noise. Just directories. Just speed.

---

## ✨ Features

- **Directory-only scanning** - faster than file-based finders
- **Real-time fuzzy search** with match scoring
- **Interactive TUI** with scrolling support for large directory trees  
- **Smart ignore patterns** - skips `.git`, `node_modules`, and other dev artifacts
- **Seamless shell integration** - inherits your final directory via autocd-go
- **Cross-platform** - Linux, macOS, Windows support

---

## 📦 Installation

```bash
go install https://github.com/codinganovel/cdf@latest
```

Or build from source:
```bash
git clone https://github.com/codinganovel/cdf
cd cdf
go build -o cdf
```

---

## 🚀 Quick Start

```bash
# Launch from current directory
cdf

# Launch from specific directory  
cdf /path/to/start

# Limit scan depth
cdf --depth 3

# Disable ignore patterns (scan all directories)
cdf --no-ignore

# Enable debug output
cdf --debug
```

---

## ⌨️ Keyboard Shortcuts

| Key | Action |
|-----|--------|
| **Type** | Filter results with fuzzy search |
| **↑/↓** | Navigate through results |
| **Enter** | Select directory and inherit to shell |
| **Esc** or **Ctrl+Q** | Cancel and exit |

---

## 🎯 The Problem It Solves

You're working in your terminal and need to quickly jump to a deeply nested directory. With traditional tools:

1. **File-based finders are slower** - they scan thousands of files you don't care about
2. **No directory inheritance** - you `cd` manually after finding the path
3. **Too much noise** - files clutter your search results

With `cdf`:
1. **Fast** - directories only, smart filtering
2. **Seamless navigation** - automatically `cd` to your selection  
3. **Clean interface** - just directories, ranked by relevance

---

## 📊 Interface Example

```
cdf > proj/api

  📁 ~/projects/myapp/api/                    [95%]
  📁 ~/personal/project/api/                  [89%]  
  📁 ~/work/backend/api-docs/                 [75%]

[3 matches] • ↑↓ navigate • Enter select • Esc/Ctrl+Q cancel
```

The percentage shows fuzzy match relevance - higher is better!

---

## ⚙️ Command Line Options

| Option | Description | Default |
|--------|-------------|---------|
| `[path]` | Starting directory | Current directory |
| `--depth <n>` | Maximum scan depth | 5 |
| `--no-ignore` | Disable ignore patterns | false |
| `--debug` | Enable debug output | false |
| `--help` | Show help message | |
| `--version` | Show version | |

---

## 🚫 Smart Ignore Patterns

By default, `cdf` skips common development directories:

- `.git`, `.cache`, `.vscode`, `.idea`
- `node_modules`, `target`, `dist`, `build`  
- `__pycache__`, `.pytest_cache`, `vendor`
- `.terraform`

Use `--no-ignore` to scan all directories.

---

## 🔧 How It Works

1. **Fast directory scanning** - Uses `filepath.WalkDir()` with depth limiting
2. **Real-time fuzzy matching** - Powered by [sahilm/fuzzy](https://github.com/sahilm/fuzzy)
3. **Interactive TUI** - Built with [tcell](https://github.com/gdamore/tcell) 
4. **Directory inheritance** - Uses [autocd-go](https://github.com/codinganovel/autocd-go) for seamless shell integration

When you select a directory, `cdf` uses process replacement to spawn a new shell in that location. From your perspective, you just navigate and end up where you wanted to be.

---

## 🎛️ Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Successful directory selection |
| `1` | Error in scanning or autocd failure |
| `2` | User cancelled (Escape/Ctrl+Q pressed) |

---

## 🖥️ Platform Support

- **Linux** - bash, zsh, fish, dash, sh
- **macOS** - bash, zsh, fish, dash, sh
- **Windows** - cmd.exe, PowerShell, PowerShell Core

Shell detection is automatic. No configuration needed.

---


## 🔒 Security

- **Path validation** prevents directory traversal
- **Smart depth limiting** prevents runaway scans
- **Permission handling** gracefully handles access denied errors

---

---

## 🤝 Integration Examples

### Shell Alias
```bash
# Add to your .bashrc/.zshrc
alias cf='cdf'
alias nav='cdf'
```

### With Depth Limiting
```bash
# Quick local navigation (depth 2)
alias cdf2='cdf --depth 2'

# Deep project exploration (depth 8)  
alias cdfp='cdf ~/projects --depth 8'
```

---

## 🔍 Real-World Usage

### Project Navigation
```bash
# Navigate to any project subdirectory
cd ~/projects && cdf
# Type "api" → jump to ~/projects/myapp/src/api/
```

### Configuration Management
```bash  
# Find config directories quickly
cdf /etc --depth 3
# Type "nginx" → jump to /etc/nginx/
```

### Log Investigation
```bash
# Navigate log directories
cdf /var/log --depth 4  
# Type "app" → jump to /var/log/myapp/
```

---

## 🧩 Dependencies

- [sahilm/fuzzy](https://github.com/sahilm/fuzzy) v0.1.1 - Fuzzy matching
- [gdamore/tcell](https://github.com/gdamore/tcell) v2.7.0 - Terminal UI
- [codinganovel/autocd-go](https://github.com/codinganovel/autocd-go) - Directory inheritance

---

## 📝 License

under ☕️, check out [the-coffee-license](https://github.com/codinganovel/The-Coffee-License)

I've included both licenses with the repo, do what you know is right. The licensing works by assuming you're operating under good faith.

---

## ✍️ Created by Sam

Because life's too short to `cd` manually to deeply nested directories.

---

*"Finally, a fuzzy finder that gets it." - Every developer who navigates directories*