<div align="center">

**[English](README.md)** | **[简体中文](README.zh-CN.md)**

</div>

# NetPilot ✈️

[![MIT License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A simple, modern Web UI for managing Linux Traffic Control (QoS) with `fq_codel`.

> The goal of this project is to provide a user-friendly interface for the powerful, but complex, Linux TC subsystem. It aims to solve the common network latency (Bufferbloat) issue by making the configuration of `fq_codel` simple and accessible.

## ✨ Core Feature

*   **🚀 QoS / Smart Queue Management**: Configure upload and download speeds to apply `fq_codel` rules and enjoy a stable, low-latency network.

## 🛠️ Tech Stack

*   **Backend**: Go (using the Netlink API)
*   **Frontend**: SvelteKit + Tailwind CSS

## 🚀 Getting Started (Development)

**Prerequisites:**
*   Go (1.25)
*   Node.js (18.x+) with pnpm

**Backend:**
```bash
cd backend
go run cmd/netpilot/main.go
```

**Frontend:**
```bash
cd frontend
pnpm install
pnpm run dev
```

## 📜 License

This project is licensed under the [MIT License](LICENSE).
