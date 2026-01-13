# 📋 cp

A fast and reliable command-line file copy utility written in Go.

## 📖 Overview

**cp** is a minimal yet robust implementation of the Unix `cp` command in Go. It provides a simple and efficient way to copy files from source to destination with proper error handling and safety checks.

## ✨ Features

- 🚀 **Fast File Copying** - Leverages Go's `io.Copy` for efficient streaming
- 🛡️ **Safety Checks** - Prevents accidental overwrites by validating source and destination paths
- 🔍 **Path Normalization** - Automatically resolves relative paths to absolute paths to avoid duplicates
- ⚠️ **Robust Error Handling** - Clear error messages with wrapped error context
- 📦 **Minimal Dependencies** - Uses only Go standard library
- 🎯 **Zero Configuration** - Works out of the box with sensible defaults

## 🔧 Installation

### From Source

```bash
git clone https://github.com/yuppyweb/cp.git
cd cp
go build -o cp cp.go
```

### Using go install

```bash
go install github.com/yuppyweb/cp@latest
```

### Using go install tool (Go >= 1.24)

```bash
go get -tool github.com/yuppyweb/cp@latest
go install tool
```

## 🚀 Usage

```bash
cp <source file> <destination file>
```

### Examples

```bash
# Copy a file
cp file.txt backup.txt

# Copy with relative paths
cp ./input/data.json ./output/data.json

# Copy to different directory
cp myfile.log /var/log/myfile.log
```

### Using with go:generate

You can also use `cp` in your `go:generate` directives:

```go
//go:generate cp ./templates/config.example.yaml ./config.yaml
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

For more information about the MIT License, visit [opensource.org/licenses/MIT](https://opensource.org/licenses/MIT).