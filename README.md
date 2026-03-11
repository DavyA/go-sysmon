# 🚀 Go-Sysmon: Advanced Terminal System Monitor

[![Go Report Card](https://goreportcard.com/badge/github.com/DavyA/go-sysmon)](https://goreportcard.com/report/github.com/DavyA/go-sysmon)
[![Go Version](https://img.shields.io/github/go-mod/go-version/DavyA/go-sysmon)](https://github.com/DavyA/go-sysmon/blob/main/go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A lightweight, high-performance terminal system monitor written in **Go**. Designed for developers and system administrators who want a clean, responsive, and efficient way to monitor their Linux systems without the overhead of heavy dependencies.

![Go-Sysmon Info View](assets/screenshot-info.png)

> [!NOTE] 
> This project was built from the ground up using raw ANSI escape codes and manual terminal state management, showcasing low-level system interaction and concurrent programming in Go.

---

## ✨ Key Features

- **📊 Real-time CPU Tracking**: Per-core usage monitoring with high-resolution historical graphs.
- **🌡️ Thermal Awareness**: Detailed temperature monitoring for CPU packages and cores.
- **🎮 GPU Insights**: Native support for **NVIDIA (NVML)** and open-source drivers (AMD/Intel), displaying utilization, VRAM, and power draw.
- **🧠 Memory & Swap**: Precision monitoring of RAM and Swap usage.
- **🌐 Network Traffic**: Real-time ingress/egress speed monitoring per interface.
- **🖥️ Responsive TUI**: 
  - Dynamic layout adaptation based on terminal size.
  - "Alternate Buffer" support for a clean exit without cluttering your terminal history.
  - Fast, flicker-free rendering using optimized ANSI sequences.
- **🐧 System Info**: Built-in "Neofetch" style dashboard with stylized OS logos.

---

## 🛠️ Technology Stack

- **Language**: Go (Golang)
- **Concurrency**: Heavy use of Goroutines and Channels for non-blocking data collection.
- **Low-Level**: Direct `ioctl` syscalls for terminal size detection and raw mode management.
- **Zero Dependencies**: Core logic relies minimal external libraries, ensuring a tiny binary size.

---

## 🚀 Getting Started

### Prerequisites

- **OS**: Linux (tested on Ubuntu, Arch, Fedora)
- **Go**: 1.25.6 or higher (as per `go.mod`)

### Installation

```bash
# Clone the repository
git clone https://github.com/DavyA/go-sysmon.git
cd go-sysmon

# Build and run
go build -o sysmon
./sysmon
```

### Quick Run (via Go Install)
```bash
go install github.com/DavyA/go-sysmon@latest
```

---

## ⌨️ Controls

Navigate through different views using your keyboard:

| Key | Action |
|-----|--------|
| `1` | **Dashboard**: Overview of CPU, RAM, and History |
| `2` | **CPU View**: Detailed per-core monitoring (WIP) |
| `3` | **GPU View**: Dedicated GPU statistics and graphs |
| `4` | **Process View**: Top processes and PID info (WIP) |
| `5` | **Info**: System summary and OS branding |
| `Q` | **Quit**: Cleanly restore terminal state and exit |

---

## 🏗️ Architecture Overview

The project follows a modular "Collector-Renderer" architecture:

1.  **Collectors**: Independent modules (`collector/`) gather system data using `/proc` filesystem, `sysfs`, and vendor-specific APIs (like NVML).
2.  **State Management**: A centralized, thread-safe `SystemState` (`utils/systemstate.go`) aggregates data from various concurrent sources.
3.  **Engine**: The main loop coordinates timing using tickers and triggers re-renders on data updates or terminal resize events.
4.  **TUI Engine**: Custom rendering logic in `utils/render.go` handles ASCII art, colored bars, and adaptive layouts.

---

## 🧪 Development & Contribution

I built this project to demonstrate performance-oriented Go development. Contributions, issues, and feature requests are welcome!

1. Fork the project.
2. Create your feature branch (`git checkout -b feature/AmazingFeature`).
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`).
4. Push to the branch (`git push origin feature/AmazingFeature`).
5. Open a Pull Request.

---

## 📄 License

Distributed under the MIT License. See `LICENSE` for more information.

---

**Built with ❤️ by [Davide](https://github.com/DavyA)**
