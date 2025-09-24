<div align="center">

**[English](README.md)** | **[简体中文](README.zh-CN.md)**

</div>

---

# NetPilot ✈️

[![MIT License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

一个用于管理Linux流量控制（QoS）与 **CAKE** 算法的、简洁的现代化Web UI。

> 本项目的目标是为强大而复杂的Linux TC子系统提供一个用户友好的界面。它致力于通过简化 **CAKE** 的配置，来解决常见的网络延迟（Bufferbloat）问题。

## ✨ 核心功能

*   **🚀 智能队列管理**: 配置您的带宽，以应用 **CAKE** 的智能队列管理，享受稳定、低延迟的网络。

## 🛠️ 技术栈

*   **后端**: Go (使用Netlink API)
*   **Frontend**: SvelteKit + Tailwind CSS

## 🚀 快速开始 (开发环境)

**前提条件:**
*   Go (1.25+)
*   Node.js (24+) with pnpm

**后端:**
```bash
cd backend
go run cmd/netpilot/main.go
```

**前端:**
```bash
cd frontend
pnpm install
pnpm run dev
````
  
