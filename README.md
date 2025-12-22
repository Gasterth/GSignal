# GSignal

![Platform](https://img.shields.io/badge/Platform-Go-00ADD8.svg?logo=go&logoColor=white)
![Go Version](https://img.shields.io/badge/Go-1.20%2B-blue.svg?logo=go)
![Latest Release](https://img.shields.io/github/v/release/Merluz/GSignal)
![License](https://img.shields.io/badge/License-MIT-blue.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/Merluz/GSignal)](https://goreportcard.com/report/github.com/Merluz/GSignal)
[![Build Status](https://github.com/Merluz/GSignal/actions/workflows/go.yml/badge.svg)](https://github.com/Merluz/GSignal/actions/workflows/go.yml)


**GSignal** is a thread-safe **Event Hub** for Go, designed to decouple components through efficient synchronous and asynchronous messaging.

It provides a robust `Hub` interface that handles event dispatching with advanced features like **worker pools**, **smart buffering**, and **context-aware subscriptions**.

## Features

- **Dual Dispatch Mode**:
  - **Synchronous**: Direct blocking calls for immediate consistency.
  - **Asynchronous**: Non-blocking dispatch supported by a dedicated worker pool.
- **Smart Buffering**: Drop-on-full policy for async consumers to prevent slow subscribers from blocking the entire system.
- **Flexible Subscriptions**:
  - **Type-based**: Subscribe to specific event types (e.g., `USER.CREATED`).
  - **Global (Firehose)**: Listen to *all* events flowing through the hub.
- **Thread-Safety**: Fully protected by `sync.RWMutex` for concurrent access.
- **Graceful Shutdown**: Ensures all workers and subscriptions are closed cleanly.

## Why GSignal?

While Go channels are powerful, managing multiple fan-out patterns and safe shutdowns can become complex.

**GSignal solves this by providing:**

- **Managed Worker Pool**: The `AsyncDispatcher` uses a configurable pool of goroutines (default 4) to handle heavy event loads without spawning a goroutine per event.
- **Safe Resource Management**: Subscriptions are automatically closed when the context is cancelled or `Unsubscribe` is called, preventing goroutine leaks.
- **Low-Allocation Dispatching**: Optimized hot path with minimal allocations.

## Installation

```bash
go get github.com/Merluz/GSignal
```

## Usage

### 1. Initialization

Create a new Hub instance. This spins up the async worker pool immediately.

```go
package main

import (
    "github.com/Merluz/GSignal"
)

func main() {
    // Create the hub
    hub := gsignal.New()

    // Ensure clean shutdown
    defer hub.Shutdown()
}
```

### 2. Subscribing

You can subscribe to specific events or listen to everything.

```go
// Create a context for the subscription
ctx := context.Background()

// Subscribe to specific events
sub, err := hub.Subscribe(ctx, "ORDER_PLACED", "PAYMENT_RECEIVED")
if err != nil {
    log.Fatal(err)
}

// Consume events in a separate goroutine
go func() {
    for evt := range sub.Events() {
        log.Printf("Received event: %s | Payload: %v", evt.Type, evt.Payload)
    }
}()
```

### 3. Publishing

#### Synchronous (Blocking)
Use this when you need to ensure all subscribers have processed the event before continuing.

```go
evt, _ := gsignal.NewEvent("ORDER_PLACED", order)
err := hub.Publish(evt) // Blocks until all subscribers receive it
```

#### Asynchronous (Non-Blocking)
Use this for fire-and-forget scenarios. The event is queued and processed by the worker pool.

```go
evt, _ := gsignal.NewEvent("LOG_ENTRY", "System started")
err := hub.PublishAsync(evt) // Returns immediately
```

## Project Structure

- `hub.go`: Main entry point and `Hub` interface implementation.
- `dispatcher.go`: Manages synchronous dispatch logic and subscription lists.
- `async_dispatcher.go`: Implements the worker pool and buffered queue for async events.
- `subscription.go`: Thread-safe channel wrapper for subscribers.
- `event.go`: ULID-based immutable event model.

## Roadmap

### Short Term
- [ ] **Options Pattern**: Configure workers, queue size, and ID generation.
- [ ] **Middleware Support**: Interceptors for logging, metrics, or validation.

### Mid Term
- [ ] **Generics Support**: Typed payloads in Go 1.18+.

### Long Term / Experimental
- [ ] **Persistent Store**: Pluggable storage (Redis/SQL) for event durability.
- [ ] **Distributed Adapter**: Bridge multiple GSignal instances via NATS or Kafka.


## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.