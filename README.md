# go-supervisor

A tiny, production-ready supervisor library for Go that automatically restarts goroutines when they crash.

This package provides a robust supervisor pattern similar to Erlang/Elixir OTP.  
It catches panics, restarts workers with exponential backoff, and supports clean shutdown using contexts.

---

## ✨ Features

- 🛡️ Auto-restart goroutines after panic  
- ⚡ Exponential backoff (configurable)  
- 🧹 Graceful shutdown using `context.Context`  
- 🧱 Panic isolation (crashing worker does not kill the app)  
- 📦 Very small API (easy to use)  
- 🔧 Configurable logger & backoff settings  

---

## 🚀 Installation

```bash
go get github.com/emcodest/go-supervisor
