# Wisp CLI

<div align="center">

**Fast, deterministic backtesting and live trading for algorithmic strategies.**

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Documentation](https://img.shields.io/badge/docs-usewisp.dev-blue)](https://usewisp.dev/docs)

*Build, backtest, and deploy trading strategies with a simple TUI interface.*

[Quick Start](https://usewisp.dev/docs/getting-started) • [Features](#features) • [Documentation](https://usewisp.dev/docs) • [Examples](https://usewisp.dev/docs/examples)

</div>

---

[![Go Report Card](https://goreportcard.com/badge/github.com/wisp-trading/wisp)](https://goreportcard.com/report/github.com/wisp-trading/wisp)

---

## 🚀 What is Wisp?

Wisp is a **low-code algorithmic trading framework** that lets you write strategies in Go and deploy them to live markets. Strategies run directly in the framework with minimal setup and maximum control.

### What Wisp Does For You

✅ **Event-Driven Execution** - Write Go code that owns its own run loop, no polling or framework callbacks
✅ **Interactive TUI** - Beautiful terminal interface for managing strategies and monitoring live trades
✅ **Multi-Exchange Support** - Unified API across Hyperliquid, Bybit, Paradex, Polymarket, and Gate.io
✅ **Real-Time Monitoring** - Live orderbook, P&L, positions, and trade data via Unix sockets
✅ **Graceful Lifecycle Management** - HTTP-based process control for reliable starts and stops
✅ **Production-Ready** - Deploy strategies to live markets with confidence

---

## 📦 Installation

### Install via Go

```bash
go install github.com/wisp-trading/wisp@latest
```

### Build from Source

```bash
git clone https://github.com/wisp-trading/wisp
cd wisp
go build -o wisp
sudo mv wisp /usr/local/bin/
```

### Verify Installation

```bash
wisp version
```

---

## ⚡ Quick Start

### 1. Initialize a New Project

```bash
mkdir my-trading-bot && cd my-trading-bot
wisp
# Navigate to: 🆕 Create New Project
```

This creates:
```
my-trading-bot/
├── config.yml              # Framework configuration
├── exchanges.yml           # Exchange credentials & settings
└── strategies/
    └── momentum/
        ├── config.yml      # Strategy metadata
        └── main.go         # Strategy implementation
```

### 2. Write Your Strategy

Create your strategy in `strategies/momentum/main.go`:

```go
package main

import (
    "context"
    "time"

    "github.com/wisp-trading/sdk/pkg/types/connector"
    "github.com/wisp-trading/sdk/pkg/types/strategy"
    "github.com/wisp-trading/sdk/pkg/types/wisp"
)

type MyStrategy struct {
    strategy.BaseStrategy
    k wisp.Wisp
}

func NewStrategy(k wisp.Wisp) strategy.Strategy {
    s := &MyStrategy{k: k}
    s.BaseStrategy = *strategy.NewBaseStrategy(strategy.BaseStrategyConfig{Name: "my-strategy"})
    return s
}

func (s *MyStrategy) Start(ctx context.Context) error {
    return s.StartWithRunner(ctx, s.run)
}

func (s *MyStrategy) run(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    pair := s.k.Pair("BTC", "USDT")

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            // Fetch klines
            klines := s.k.Spot().Klines(connector.Hyperliquid, pair, "1h", 14)

            // Calculate RSI
            rsi, _ := s.k.Indicators().RSI(klines, 14)
            if len(rsi) == 0 {
                continue
            }

            currentRSI := rsi[len(rsi)-1]

            // Oversold: buy
            if currentRSI < 30 {
                signal, _ := s.k.Spot().Signal(s.GetName()).
                    Buy(pair, connector.Hyperliquid, s.k.Asset("BTC").Qty(0.01)).
                    Build()
                s.Emit(signal)
            }

            // Overbought: sell
            if currentRSI > 70 {
                signal, _ := s.k.Spot().Signal(s.GetName()).
                    Sell(pair, connector.Hyperliquid, s.k.Asset("BTC").Qty(0.01)).
                    Build()
                s.Emit(signal)
            }
        }
    }
}
```

### 3. Configure Your Strategy

Edit `strategies/momentum/config.yml`:

```yaml
name: momentum
display_name: "Momentum Strategy"
description: "RSI-based momentum trading"
type: momentum

exchanges:
  - hyperliquid

assets:
  hyperliquid:
    - BTC/USDT
    - ETH/USDT

indicators:
  rsi:
    period: 14
    oversold: 30
    overbought: 70

parameters:
  position_size: 0.1
```

### 4. Deploy to Live Trading

```bash
wisp
# Navigate to: Strategies → momentum → Start Live
```

Your strategy runs as a detached process, continuing even after you close the CLI.

### 5. Monitor Live Strategies

```bash
wisp
# Navigate to: Monitor
```

Real-time monitoring dashboard shows:
- **Overview**: Strategy status, uptime, health
- **Positions**: Active positions across exchanges
- **Orderbook**: Live orderbook depth
- **Trades**: Recent trade history
- **PnL**: Realized/unrealized profit & loss

### 6. Stop a Running Strategy

```bash
# From Monitor view:
# 1. Select running instance
# 2. Press [S]
# 3. Confirm "Yes, Stop"
```

Graceful HTTP-based shutdown ensures clean process termination.

---

## 🎯 Features

### Strategy Development

- **Event-Driven Architecture** - Own your run loop with `Start()` and `run()` methods, no polling framework
- **Goroutine-Native** - Leverage Go's concurrency without fighting the runtime
- **Type-Safe API** - Full IDE support with autocomplete
- **Rich Indicators** - RSI, MACD, Bollinger Bands, EMA, SMA, ATR, Stochastic and more
- **Multi-Asset** - Trade multiple assets simultaneously
- **Multi-Exchange** - Execute across multiple exchanges in one strategy

### Live Trading

- **Process Isolation** - Each strategy runs in its own process
- **Detached Execution** - Strategies continue after CLI closes
- **State Persistence** - Instance state survives CLI restarts
- **Real-Time Data** - WebSocket + REST hybrid ingestion
- **Position Tracking** - Automatic position reconciliation
- **Trade Backfill** - Recovers trades on restart

### Monitoring

- **Unix Socket Communication** - Fast, local IPC
- **HTTP API** - RESTful access to strategy data
- **Live Orderbook** - Real-time order book updates
- **PnL Tracking** - Realized and unrealized profit/loss
- **Health Checks** - System health and error reporting
- **Multi-Instance** - Monitor multiple strategies at once

### Exchange Support
| Exchange | Spot | Perpetual | Prediction |
|----------|------|-----------|-----------|
| Hyperliquid | - | ✅ | - |
| Bybit | ✅ | ✅ | - |
| Paradex | ✅ | ✅ | - |
| Polymarket | - | - | ✅ |
| Gate.io | ✅ | - | - |

### User Interface

- **Beautiful TUI** - Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Keyboard Navigation** - Vim-style keybindings (hjkl)
- **Responsive Design** - Adapts to terminal size
- **Color Coding** - Visual status indicators
- **Progress Tracking** - Real-time backtest progress

---

## 📚 Documentation

**Full Documentation**: [https://usewisp.dev/docs](https://usewisp.dev/docs)

### Key Resources

- [Getting Started](https://usewisp.dev/docs/getting-started) - Installation and first steps
- [Writing Strategies](https://usewisp.dev/docs/getting-started/writing-strategies) - 13 strategy patterns with examples
- [Strategy Examples](https://usewisp.dev/docs/examples) - Real strategies from basic to advanced
- [API Reference](https://usewisp.dev/docs/api/indicators/rsi) - Complete API documentation
- [Architecture](https://usewisp.dev/docs/intro) - How Wisp works

---

## 🎨 Screenshots

### Main Menu
```
╭─────────────────────────────────────────────╮
│            WISP CLI v0.1.0                │
│                                             │
│  What would you like to do?                 │
│                                             │
│  ▶ 📂 Strategies                            │
│    📊 Monitor                               │
│    ⚙️  Settings                             │
│    ℹ️  Help                                  │
│    🆕 Create New Project                    │
│                                             │
│  ↑↓/jk Navigate  ↵ Select  q Quit           │
╰─────────────────────────────────────────────╯
```

### Live Monitoring
```
╭─────────────────────────────────────────────────────────────────────╮
│ MONITOR                                                             │
│                                                                     │
│  STATUS  STRATEGY           PID     UPTIME    PNL        HEALTH     │
│  ────────────────────────────────────────────────────────────────   │
│  ✓ RUN   momentum           86697   2h 30m    +$125.50  ███████     │
│    STP   arbitrage          -       -         -$43.20   ─────       │
│                                                                     │
│ [↑↓] Navigate • [Enter] Details • [S] Stop • [R] Refresh • [Q] Back │
╰─────────────────────────────────────────────────────────────────────╯
```

### Orderbook View
```
╭─────────────────────────────────────────────╮
│ ORDERBOOK - BTC/USDT (hyperliquid)         │
│                                             │
│ ASKS                                        │
│ 43,251.50  ████████░░  0.5420  $23,456    │
│ 43,250.00  ██████░░░░  0.3210  $13,888    │
│ 43,249.50  ████░░░░░░  0.1890  $8,174     │
│                                             │
│ BIDS                                        │
│ 43,248.00  ██████████  0.8920  $38,577    │
│ 43,247.50  ████████░░  0.6540  $28,284    │
│ 43,247.00  ██████░░░░  0.4320  $18,683    │
│                                             │
│ Spread: $3.50 (0.008%)  Last: 43,248.75   │
╰─────────────────────────────────────────────╯
```

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      WISP CLI (TUI)                         │
│  • Strategy Browser  • Monitor  • Live Trading              │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│              INSTANCE MANAGER (Process Control)             │
│  • Load strategies  • Start/Stop  • State Persistence       │
└────────────────────┬────────────────────────────────────────┘
                     │
        ┌────────────┴────────────┐
        ▼                         ▼
┌──────────────────┐    ┌──────────────────┐
│  Strategy Process│    │  Strategy Process│
│  ┌────────────┐ │    │  ┌────────────┐ │
│  │Strategy Go │ │    │  │Strategy Go │ │
│  │   Code     │ │    │  │   Code     │ │
│  └────────────┘ │    │  └────────────┘ │
│  ┌────────────┐ │    │  ┌────────────┐ │
│  │  SDK Core  │ │    │  │  SDK Core  │ │
│  └────────────┘ │    │  └────────────┘ │
│  ┌────────────┐ │    │  ┌────────────┐ │
│  │ Monitoring │ │    │  │ Monitoring │ │
│  │   Server   │ │    │  │   Server   │ │
│  └────────────┘ │    │  └────────────┘ │
└────────┬─────────┘    └────────┬─────────┘
         │                       │
         └───────────┬───────────┘
                     ▼
         ┌─────────────────────┐
         │   EXCHANGES         │
         │  • Hyperliquid      │
         │  • Bybit            │
         │  • Paradex          │
         └─────────────────────┘
```

### Key Components

1. **CLI** - Interactive terminal interface for managing strategies
2. **Instance Manager** - Controls strategy lifecycle (start/stop/monitor)
3. **Strategy Runtime** - Loads and executes Go strategies
4. **SDK Runtime** - Core execution engine with data ingestion
5. **Monitoring Server** - HTTP API exposed via Unix sockets
6. **Exchange Connectors** - Unified interface to multiple exchanges

---

## 🔧 Commands

### Interactive Mode (Default)

```bash
wisp
```

Launches the TUI interface with full navigation.

### CLI Mode

```bash
wisp --cli          # Show command help
wisp version        # Show version info
```

---

## 📊 Example Strategies

### Momentum (RSI-Based)

```go
func (s *momentumStrategy) run(ctx context.Context) {
ticker := time.NewTicker(5 * time.Minute)
defer ticker.Stop()
pair := s.k.Pair("BTC", "USDT")

for {
select {
case <-ctx.Done():
return
case <-ticker.C:
klines := s.k.Spot().Klines(connector.Hyperliquid, pair, "1h", 14)
rsi, _ := s.k.Indicators().RSI(klines, 14)

if len(rsi) == 0 {
continue
}

current := rsi[len(rsi)-1]

if current < 30 {
signal, _ := s.k.Spot().Signal(s.GetName()).
Buy(pair, connector.Hyperliquid, s.k.Asset("BTC").Qty(0.1)).
Build()
s.Emit(signal)
}

if current > 70 {
signal, _ := s.k.Spot().Signal(s.GetName()).
Sell(pair, connector.Hyperliquid, s.k.Asset("BTC").Qty(0.1)).
Build()
s.Emit(signal)
}
}
}
}
```

### Cross-Exchange Trading

```go
func (s *crossExchangeStrategy) run(ctx context.Context) {
ticker := time.NewTicker(1 * time.Minute)
defer ticker.Stop()
pair := s.k.Pair("BTC", "USDT")

for {
select {
case <-ctx.Done():
return
case <-ticker.C:
// Fetch price data from both exchanges
bybitPrice := s.k.Spot().Klines(connector.Bybit, pair, "1m", 1)
hyperliquidPrice := s.k.Spot().Klines(connector.Hyperliquid, pair, "1m", 1)

if len(bybitPrice) == 0 || len(hyperliquidPrice) == 0 {
continue
}

// Compare prices and execute on spread
spread := bybitPrice[0].Close.Sub(hyperliquidPrice[0].Close)

// If profitable spread detected, trade both sides
if spread.GreaterThan(numerical.Zero) {
buy, _ := s.k.Spot().Signal(s.GetName()).
Buy(pair, connector.Hyperliquid, s.k.Asset("BTC").Qty(0.5)).
Build()
s.Emit(buy)

sell, _ := s.k.Spot().Signal(s.GetName()).
Sell(pair, connector.Bybit, s.k.Asset("BTC").Qty(0.5)).
Build()
s.Emit(sell)
}
}
}
}
```

---

## 🤝 Contributing

We welcome contributions! Here's how you can help:

### Development Setup

```bash
# Clone the repo
git clone https://github.com/wisp-trading/wisp
cd wisp

# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build -o wisp
```

### Running Tests

```bash
# All tests
go test ./...

# Specific package
go test ./internal/services/monitoring/...

# With coverage
go test -cover ./...

# Watch mode
ginkgo watch -r
```

### Code Style

- Use `gofmt` for formatting
- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Write tests with [Ginkgo](https://github.com/onsi/ginkgo) and [Gomega](https://github.com/onsi/gomega)
- Use [mockery](https://github.com/vektra/mockery) for mocks

### Areas We Need Help

- 🌐 Additional exchange connectors
- 📊 More technical indicators
- 📚 Documentation improvements
- 🐛 Bug fixes and testing
- 🎨 UI/UX enhancements

---

## 🗺️ Roadmap

### v0.2.0 (Q1 2026)

- [ ] WebSocket-based live monitoring dashboard
- [ ] Strategy performance comparison tool
- [ ] Paper trading mode
- [ ] Discord/Telegram notifications
- [ ] Portfolio optimization tools

### v0.3.0 (Q2 2026)

- [ ] Cloud deployment support
- [ ] Strategy marketplace
- [ ] Multi-user support
- [ ] Advanced risk management
- [ ] Machine learning integration

### v1.0.0 (Q3 2026)

- [ ] Enterprise features
- [ ] Professional support
- [ ] Advanced analytics suite

---

## ❓ FAQ

**Q: Is Wisp suitable for production trading?**
A: Yes, but use appropriate risk management. Start with small positions and paper trading.

**Q: What exchanges are supported?**
A: Currently only Hyperliquid perpetuals are stable in production. Bybit and Paradex are in active development.

**Q: Can I run multiple strategies simultaneously?**
A: Yes! Each strategy runs in its own isolated process.

**Q: How do I handle API keys securely?**
A: Store them in `exchanges.yml` with proper file permissions (chmod 600).

**Q: Can I write strategies in languages other than Go?**
A: Wisp strategies must be written in Go to run in the framework. However, you can integrate machine learning models from any language using:
- **gRPC** - Call ML inference services in Python, R, or any language
- **ONNX Runtime** - Load pre-trained models directly in Go
- **HTTP APIs** - Connect to external prediction services

---

## 🐛 Troubleshooting

### Strategy Won't Start

```bash
# Ensure Go version matches
go version  # Should be 1.24+

# Check strategy code for syntax errors
# The CLI will show runtime errors during startup

# Verify imports are correct
grep -r "github.com/wisp-trading/sdk" strategies/momentum/main.go

# Check file permissions
chmod 644 strategies/momentum/main.go
```

If you see errors when starting a strategy, check your Go source code for syntax or import errors. The CLI will display the exact error message.

### Process Won't Stop

```bash
# Find the process
ps aux | grep momentum

# Force kill
kill -9 <PID>

# Clean up socket files
rm ~/.wisp/sockets/*.sock
```

### "Already Running" Error

```bash
# Check for orphaned processes
ps aux | grep wisp

# Clean up state files
rm ~/.wisp/instances/*/state.json
```

---

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

---

## 🙏 Acknowledgments

Built with:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions
- [Cobra](https://github.com/spf13/cobra) - CLI commands
- [Fx](https://github.com/uber-go/fx) - Dependency injection
- [Ginkgo](https://github.com/onsi/ginkgo) - Testing framework

---

## 📞 Support

- 📖 [Documentation](https://usewisp.dev/docs)
- 💬 [Discord Community](#) *(coming soon)*
- 🐛 [Issue Tracker](https://github.com/wisp-trading/wisp/issues)
- ✉️ [Email Support](#) *(for enterprise)*

---

<div align="center">

**⭐ If you find Wisp useful, please consider starring the repo! ⭐**

Made with ❤️ by the Wisp Team

</div>
