# ibapi-go

A Go client library for the Interactive Brokers Web API.

## Installation

```bash
go get github.com/bishop-bot/ibapi-go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    ibapi "github.com/bishop-bot/ibapi-go"
)

func main() {
    // Create client with default settings
    client, err := ibapi.NewClient(
        ibapi.WithBaseURL("https://localhost:5000"),
        ibapi.WithInsecureSkipVerify(true),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx := context.Background()

    // Check authentication status
    status, err := client.Session().AuthStatus(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Authenticated: %v\n", status.Authenticated)

    // Search for a contract
    results, err := client.Contract().SecdefSearch(ctx, "AAPL")
    if err != nil {
        log.Fatal(err)
    }
    for _, r := range results {
        fmt.Printf("Found: %s (%s)\n", r.Symbol, r.CompanyName)
    }
}
```

## Features

### Session Management
- `Session().AuthStatus()` - Check authentication status
- `Session().InitSession()` - Initialize a new session
- `Session().Ping()` - Keep session alive
- `Session().ValidateSSO()` - Validate SSO session
- `Session().Logout()` - Log out

### Contract & Security Definitions
- `Contract().SecdefSearch()` - Search contracts by symbol
- `Contract().SecdefInfo()` - Get contract details
- `Contract().ContractInfo()` - Get contract information by conid
- `Contract().SecdefStrikes()` - Get strike prices for options
- `Contract().SecdefByConids()` - Get security definitions by conids
- `Contract().AllConidsByExchange()` - Get all conids on an exchange
- `Contract().Stocks()` - Get stock contracts by symbols
- `Contract().Futures()` - Get futures contracts
- `Contract().TradingSchedule()` - Get trading schedule

### Market Data
- `MarketData().HistoricalData()` - Get historical OHLCV data
- `MarketData().Snapshot()` - Get market data snapshot
- `MarketData().Unsubscribe()` - Unsubscribe from market data

### Market Scanner
- `Scanner().Params()` - Get available scanner parameters
- `Scanner().Run()` - Run a market scan

## Configuration

### Client Options

```go
client, err := ibapi.NewClient(
    ibapi.WithBaseURL("https://localhost:5000"),    // Default API base URL
    ibapi.WithInsecureSkipVerify(true),              // Skip TLS verification (for localhost)
    ibapi.WithTimeout(30 * time.Second),            // Request timeout
    ibapi.WithHTTPClient(customClient),              // Custom HTTP client
)
```

## API Endpoints

### Session Management
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/iserver/auth/status` | Check authentication status |
| POST | `/iserver/auth/ssodh/init` | Initialize session |
| POST | `/tickle` | Ping to prevent timeout |
| GET | `/sso/validate` | Validate SSO |
| POST | `/logout` | Log out |

### Contract Definitions
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/trsrv/secdef` | Search security definitions |
| GET | `/trsrv/all-conids` | Get all conids by exchange |
| GET | `/iserver/contract/{conid}/info` | Get contract info |
| GET | `/iserver/secdef/search` | Search contracts |
| GET | `/iserver/secdef/info` | Get secdef info |
| GET | `/trsrv/futures` | Get futures contracts |
| GET | `/trsrv/stocks` | Get stock contracts |
| GET | `/trsrv/secdef/schedule` | Get trading schedule |
| GET | `/contract/trading-schedule` | Get trading schedule |

### Market Data
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/iserver/marketdata/history` | Historical data |
| GET | `/iserver/marketdata/snapshot` | Market data snapshot |
| POST | `/iserver/marketdata/unsubscribe` | Unsubscribe |
| GET | `/iserver/marketdata/unsubscribeall` | Unsubscribe all |
| GET | `/md/regsnapshot` | Regulatory snapshot |

### Market Scanner
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/iserver/scanner/params` | Get scanner parameters |
| POST | `/iserver/scanner/run` | Run market scan |

## License

MIT