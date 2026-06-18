# Go Channels and Concurrency - Primer

**Purpose**: Understand Go's concurrency primitives for decoupling business logic from display logic  
**Audience**: Developers familiar with Go basics, new to channels/goroutines  
**Date**: 2026-06-18

---

## Table of Contents

1. [Goroutines](#goroutines)
2. [Channels](#channels)
3. [Channel Patterns](#channel-patterns)
4. [Select Statement](#select-statement)
5. [Common Pitfalls](#common-pitfalls)
6. [Best Practices](#best-practices)
7. [Real-World Examples](#real-world-examples)

---

## Goroutines

### What is a Goroutine?

A **goroutine** is a lightweight thread managed by the Go runtime. Unlike OS threads, goroutines are cheap to create and the Go scheduler can multiplex thousands of goroutines onto a smaller number of OS threads.

### Basic Usage

```go
// Sequential execution
func main() {
    doWork()        // Blocks until complete
    doMoreWork()    // Runs after doWork() finishes
}

// Concurrent execution
func main() {
    go doWork()     // Runs in background
    doMoreWork()    // Runs immediately
}
```

### Important Characteristics

1. **Lightweight**: ~2KB stack size (vs 1-2MB for OS threads)
2. **Fast to create**: Millions can be created
3. **Non-blocking**: `go` keyword returns immediately
4. **No return value**: Can't return values directly (use channels)

### Example: Basic Goroutine

```go
package main

import (
    "fmt"
    "time"
)

func sayHello(name string) {
    time.Sleep(100 * time.Millisecond)
    fmt.Printf("Hello, %s!\n", name)
}

func main() {
    // Without goroutine - sequential
    sayHello("Alice")  // Takes 100ms
    sayHello("Bob")    // Takes 100ms
    // Total: 200ms

    // With goroutines - concurrent
    go sayHello("Alice")  // Starts immediately
    go sayHello("Bob")    // Starts immediately
    time.Sleep(150 * time.Millisecond)  // Wait for goroutines
    // Total: ~100ms
}
```

**Problem**: How do we wait for goroutines properly? → **Channels**

---

## Channels

### What is a Channel?

A **channel** is a typed conduit through which you can send and receive values. Channels connect concurrent goroutines, allowing them to communicate and synchronize.

### Creating Channels

```go
// Unbuffered channel (blocks until received)
ch := make(chan int)

// Buffered channel (holds N values before blocking)
ch := make(chan int, 10)

// Receive-only channel (type declaration)
var recv <-chan int

// Send-only channel (type declaration)
var send chan<- int
```

### Channel Operations

```go
ch := make(chan string)

// Send (blocks until received)
ch <- "hello"

// Receive (blocks until sent)
msg := <-ch

// Receive and check if channel is closed
msg, ok := <-ch
if !ok {
    // Channel closed
}

// Close channel (sender's responsibility)
close(ch)
```

### Unbuffered vs Buffered

**Unbuffered Channel** (synchronous):
```go
ch := make(chan int)  // No buffer

// This will DEADLOCK if there's no receiver
ch <- 42  // Blocks until someone receives

// Correct usage:
go func() {
    ch <- 42  // Send in goroutine
}()
value := <-ch  // Receive in main
```

**Buffered Channel** (asynchronous up to capacity):
```go
ch := make(chan int, 2)  // Buffer of 2

ch <- 1  // Doesn't block (buffer has space)
ch <- 2  // Doesn't block (buffer full after this)
ch <- 3  // BLOCKS (buffer full, no receiver)

// Correct usage:
ch := make(chan int, 2)
ch <- 1
ch <- 2
fmt.Println(<-ch)  // 1
fmt.Println(<-ch)  // 2
```

### Direction Annotations

Channels can be restricted to send-only or receive-only:

```go
// Function that only sends
func producer(ch chan<- int) {
    ch <- 42
    // value := <-ch  // ERROR: receive from send-only channel
}

// Function that only receives
func consumer(ch <-chan int) {
    value := <-ch
    // ch <- 42  // ERROR: send to receive-only channel
}

func main() {
    ch := make(chan int)  // Bidirectional
    go producer(ch)       // Passed as send-only
    consumer(ch)          // Passed as receive-only
}
```

---

## Channel Patterns

### Pattern 1: Simple Communication

**Use Case**: Send result from goroutine to main thread

```go
func fetchData(url string) string {
    // Simulate network request
    time.Sleep(100 * time.Millisecond)
    return "data from " + url
}

func main() {
    resultCh := make(chan string)

    go func() {
        data := fetchData("https://api.example.com")
        resultCh <- data  // Send result
    }()

    result := <-resultCh  // Wait for result
    fmt.Println(result)
}
```

### Pattern 2: Progress Updates

**Use Case**: Send multiple progress updates during long operation

```go
type ProgressUpdate struct {
    Step    string
    Percent int
}

func processData(progressCh chan<- ProgressUpdate) {
    steps := []string{"Loading", "Processing", "Saving"}
    
    for i, step := range steps {
        progressCh <- ProgressUpdate{
            Step:    step,
            Percent: (i + 1) * 33,
        }
        time.Sleep(100 * time.Millisecond)
    }
    
    close(progressCh)  // Signal completion
}

func main() {
    progressCh := make(chan ProgressUpdate)
    
    go processData(progressCh)
    
    for update := range progressCh {  // Loops until channel closed
        fmt.Printf("%s: %d%%\n", update.Step, update.Percent)
    }
}
```

### Pattern 3: Error Handling

**Use Case**: Return both result and error from goroutine

```go
type Result struct {
    Data  string
    Error error
}

func fetchWithError(url string) Result {
    if url == "" {
        return Result{Error: fmt.Errorf("empty URL")}
    }
    
    time.Sleep(100 * time.Millisecond)
    return Result{Data: "data from " + url}
}

func main() {
    resultCh := make(chan Result)
    
    go func() {
        result := fetchWithError("")
        resultCh <- result
    }()
    
    result := <-resultCh
    if result.Error != nil {
        fmt.Printf("Error: %v\n", result.Error)
        return
    }
    
    fmt.Println(result.Data)
}
```

### Pattern 4: Fan-Out (Multiple Workers)

**Use Case**: Distribute work across multiple goroutines

```go
func worker(id int, jobs <-chan int, results chan<- int) {
    for job := range jobs {
        fmt.Printf("Worker %d processing job %d\n", id, job)
        time.Sleep(100 * time.Millisecond)
        results <- job * 2
    }
}

func main() {
    jobs := make(chan int, 10)
    results := make(chan int, 10)
    
    // Start 3 workers
    for w := 1; w <= 3; w++ {
        go worker(w, jobs, results)
    }
    
    // Send 5 jobs
    for j := 1; j <= 5; j++ {
        jobs <- j
    }
    close(jobs)
    
    // Collect 5 results
    for a := 1; a <= 5; a++ {
        <-results
    }
}
```

### Pattern 5: Done Channel

**Use Case**: Signal completion without sending data

```go
func doWork(done chan<- struct{}) {
    fmt.Println("Working...")
    time.Sleep(200 * time.Millisecond)
    fmt.Println("Done!")
    close(done)  // Signal completion
}

func main() {
    done := make(chan struct{})
    
    go doWork(done)
    
    <-done  // Wait for completion
    fmt.Println("All work complete")
}
```

**Why `struct{}`?** It's a zero-byte type - memory efficient for signaling.

---

## Select Statement

### What is Select?

`select` lets a goroutine wait on multiple channel operations. It's like a `switch` statement for channels.

### Basic Usage

```go
select {
case msg := <-ch1:
    fmt.Println("Received from ch1:", msg)
case msg := <-ch2:
    fmt.Println("Received from ch2:", msg)
case ch3 <- "hello":
    fmt.Println("Sent to ch3")
default:
    fmt.Println("No channel ready")
}
```

### Example: Timeout

```go
func fetchWithTimeout(url string) (string, error) {
    resultCh := make(chan string)
    
    go func() {
        time.Sleep(2 * time.Second)  // Simulate slow request
        resultCh <- "data"
    }()
    
    select {
    case result := <-resultCh:
        return result, nil
    case <-time.After(1 * time.Second):
        return "", fmt.Errorf("timeout")
    }
}
```

### Example: Multiple Channels

```go
func monitor(dataCh <-chan string, errCh <-chan error, done <-chan struct{}) {
    for {
        select {
        case data := <-dataCh:
            fmt.Println("Data:", data)
        case err := <-errCh:
            fmt.Println("Error:", err)
        case <-done:
            fmt.Println("Shutting down")
            return
        }
    }
}
```

### Example: Non-Blocking Send/Receive

```go
// Non-blocking receive
select {
case msg := <-ch:
    fmt.Println("Received:", msg)
default:
    fmt.Println("No message available")
}

// Non-blocking send
select {
case ch <- "hello":
    fmt.Println("Sent message")
default:
    fmt.Println("Channel full or no receiver")
}
```

---

## Common Pitfalls

### 1. Deadlock - No Receiver

```go
// WRONG - will deadlock
func main() {
    ch := make(chan int)
    ch <- 42  // Blocks forever (no receiver)
}

// CORRECT - send in goroutine
func main() {
    ch := make(chan int)
    go func() {
        ch <- 42
    }()
    fmt.Println(<-ch)
}
```

### 2. Goroutine Leak - Abandoned Goroutine

```go
// WRONG - goroutine never exits
func leak() {
    ch := make(chan int)
    go func() {
        value := <-ch  // Blocks forever if nothing sent
        fmt.Println(value)
    }()
    // Function returns, goroutine still waiting
}

// CORRECT - always provide exit path
func noLeak(ctx context.Context) {
    ch := make(chan int)
    go func() {
        select {
        case value := <-ch:
            fmt.Println(value)
        case <-ctx.Done():
            return  // Exit when context cancelled
        }
    }()
}
```

### 3. Closing Already-Closed Channel

```go
// WRONG - panic
ch := make(chan int)
close(ch)
close(ch)  // PANIC: close of closed channel

// CORRECT - use sync.Once or check
var once sync.Once
once.Do(func() {
    close(ch)
})
```

### 4. Sending on Closed Channel

```go
// WRONG - panic
ch := make(chan int)
close(ch)
ch <- 42  // PANIC: send on closed channel

// CORRECT - don't send after close, or use defer/recover
```

### 5. Range Without Close

```go
// WRONG - loops forever
ch := make(chan int)
go func() {
    ch <- 1
    ch <- 2
    // Forgot to close(ch)
}()

for value := range ch {  // Never exits
    fmt.Println(value)
}

// CORRECT - always close when done sending
go func() {
    ch <- 1
    ch <- 2
    close(ch)  // Signal no more values
}()
```

---

## Best Practices

### 1. Close Channels from Sender

**Rule**: Only the sender should close a channel, never the receiver.

```go
// GOOD
func producer(ch chan<- int) {
    for i := 0; i < 5; i++ {
        ch <- i
    }
    close(ch)  // Producer closes
}

func consumer(ch <-chan int) {
    for value := range ch {
        fmt.Println(value)
    }
    // Don't close here - consumer doesn't close
}
```

### 2. Use Context for Cancellation

**Rule**: Use `context.Context` for cancellation and timeouts.

```go
func worker(ctx context.Context, jobs <-chan int) {
    for {
        select {
        case job := <-jobs:
            // Process job
        case <-ctx.Done():
            fmt.Println("Cancelled:", ctx.Err())
            return
        }
    }
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    jobs := make(chan int)
    go worker(ctx, jobs)
    
    // Work happens...
}
```

### 3. Buffered Channels for Known Capacity

**Rule**: Use buffered channels when you know the number of items.

```go
// If you know you'll send 5 items
results := make(chan Result, 5)

go func() {
    for i := 0; i < 5; i++ {
        results <- doWork(i)
    }
    close(results)
}()

for result := range results {
    handleResult(result)
}
```

### 4. Use WaitGroup for Synchronization

**Rule**: Use `sync.WaitGroup` to wait for multiple goroutines.

```go
var wg sync.WaitGroup

for i := 0; i < 5; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        doWork(id)
    }(i)
}

wg.Wait()  // Wait for all goroutines
```

### 5. Channel Direction Annotations

**Rule**: Restrict channel directions in function signatures.

```go
// Makes intent clear and prevents mistakes
func send(ch chan<- int) {  // Can only send
    ch <- 42
}

func receive(ch <-chan int) {  // Can only receive
    value := <-ch
}
```

---

## Real-World Examples

### Example 1: HTTP Request with Timeout

```go
type HTTPResult struct {
    Body  string
    Error error
}

func fetchURL(url string, timeout time.Duration) HTTPResult {
    resultCh := make(chan HTTPResult, 1)
    
    go func() {
        resp, err := http.Get(url)
        if err != nil {
            resultCh <- HTTPResult{Error: err}
            return
        }
        defer resp.Body.Close()
        
        body, err := io.ReadAll(resp.Body)
        resultCh <- HTTPResult{Body: string(body), Error: err}
    }()
    
    select {
    case result := <-resultCh:
        return result
    case <-time.After(timeout):
        return HTTPResult{Error: fmt.Errorf("timeout after %v", timeout)}
    }
}

func main() {
    result := fetchURL("https://example.com", 5*time.Second)
    if result.Error != nil {
        fmt.Println("Error:", result.Error)
        return
    }
    fmt.Println("Body:", result.Body)
}
```

### Example 2: Pipeline Pattern

```go
// Stage 1: Generate numbers
func generate(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        for _, n := range nums {
            out <- n
        }
        close(out)
    }()
    return out
}

// Stage 2: Square numbers
func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        for n := range in {
            out <- n * n
        }
        close(out)
    }()
    return out
}

// Stage 3: Sum numbers
func sum(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        total := 0
        for n := range in {
            total += n
        }
        out <- total
        close(out)
    }()
    return out
}

func main() {
    // Pipeline: generate -> square -> sum
    c := generate(1, 2, 3, 4, 5)
    c = square(c)
    c = sum(c)
    
    result := <-c
    fmt.Println("Sum of squares:", result)  // 55
}
```

### Example 3: Worker Pool

```go
type Job struct {
    ID   int
    Data string
}

type Result struct {
    Job    Job
    Output string
}

func worker(id int, jobs <-chan Job, results chan<- Result) {
    for job := range jobs {
        fmt.Printf("Worker %d started job %d\n", id, job.ID)
        time.Sleep(time.Second)
        results <- Result{
            Job:    job,
            Output: fmt.Sprintf("processed by worker %d", id),
        }
    }
}

func main() {
    numWorkers := 3
    numJobs := 5
    
    jobs := make(chan Job, numJobs)
    results := make(chan Result, numJobs)
    
    // Start workers
    for w := 1; w <= numWorkers; w++ {
        go worker(w, jobs, results)
    }
    
    // Send jobs
    for j := 1; j <= numJobs; j++ {
        jobs <- Job{ID: j, Data: fmt.Sprintf("data-%d", j)}
    }
    close(jobs)
    
    // Collect results
    for a := 1; a <= numJobs; a++ {
        result := <-results
        fmt.Printf("Job %d: %s\n", result.Job.ID, result.Output)
    }
}
```

---

## Summary

### Key Concepts

| Concept | Purpose | When to Use |
|---------|---------|-------------|
| **Goroutine** | Lightweight concurrent execution | Any background work |
| **Channel** | Communication between goroutines | Passing data between concurrent code |
| **Buffered Channel** | Async communication | Known capacity, reduce blocking |
| **Select** | Multiplex channel operations | Timeouts, multiple sources, cancellation |
| **Close** | Signal no more values | When sender is done |
| **Context** | Cancellation/timeout | Distributed cancellation |
| **WaitGroup** | Wait for goroutines | Synchronize multiple goroutines |

### Channel Rules

1. ✅ **Sender closes** - Never close from receiver
2. ✅ **Don't send on closed** - Will panic
3. ✅ **Range until close** - `for v := range ch` needs `close(ch)`
4. ✅ **Buffered for known size** - Prevents blocking when size known
5. ✅ **Direction annotations** - Makes intent clear (`chan<-`, `<-chan`)

### Common Patterns

- **Result with error**: `struct { Data T; Error error }`
- **Progress updates**: Send multiple updates, close when done
- **Done signal**: `chan struct{}` for zero-byte signaling
- **Timeout**: `select` with `time.After()`
- **Worker pool**: Fixed number of workers, job channel
- **Pipeline**: Chain of channels, each stage processes

---

## Additional Resources

- [Effective Go - Concurrency](https://go.dev/doc/effective_go#concurrency)
- [Go Blog - Pipelines and Cancellation](https://go.dev/blog/pipelines)
- [Go Blog - Context](https://go.dev/blog/context)
- [Go by Example - Channels](https://gobyexample.com/channels)

---

**Next**: See how to apply these patterns to obstool in `business-logic-decoupling-proposal.md`
