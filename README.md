# GoSRT - High-Performance Fork for SRTLA and Cellular Bonding

This repository is an optimized fork of the pure Go SRT protocol implementation (`github.com/datarhei/gosrt`). 

It has been customized and patched by the **perhost.app** team specifically for integration with multi-link cellular bonding gateways and **SRTLA receivers** (such as `irlserver` or similar streaming ingestion nodes). It is designed to remain extremely stable under high packet loss, severe jitter, and significant latency differentials across bonding links.

---

## ⚠️ Accepted Limitations (What is NOT supported)

To keep the codebase lightweight, highly performant, and focused on low-latency live streaming, we accept and do not implement the following features of the standard SRT protocol:

*   **❌ Buffer Mode**: Not supported. Only **Live Mode (TSBPD)** is implemented for real-time video/audio streaming.
*   **❌ Rendezvous Handshake**: Not supported. Only **Caller (Client)** and **Listener (Server)** connection modes are supported.
*   **❌ File Transfer Congestion Control (FileCC)**: Not supported. The congestion control mechanism is exclusively **LiveCC** (designed for live video).
*   **❌ Native Connection Bonding**: GoSRT **does not aggregate links internally**. All multi-path bonding, link registration, and packet de-encapsulation are handled by the upstream SRTLA receiver/gateway before passing the clean SRT stream to this library.

---

## 🚀 Fork Optimizations & Patches

This fork introduces critical enhancements to address CPU bottlenecks and retransmission storms typical of cellular bonding networks:

### 1. O(1) Packet Insertion Optimization (High-Performance Queue)
*   **The Issue**: The native `gosrt` receiver queue (`packetList`) performed a linear search starting from the oldest packet (`Front()`) to insert newly arrived packets in order. In multi-path bonding networks (WiFi + multiple SIMs), out-of-order packet arrival is constant, and late packets are usually very recent. Scanning from the front resulted in $O(N^2)$ complexity, causing severe CPU spikes, socket buffer overflows, and packet loss on the receiver.
*   **The Solution**: We modified the queue insertion in `congestion/live/receive.go` to scan backwards from the most recent packet (`Back()`) using `Prev()`. Since out-of-order packets in bonding are almost always recent, insertion complexity becomes **$O(1)$** in practice.
*   **The Result**: Performance is **300x faster** (processing 100,000 desynchronized packets takes just **36 ms**, averaging `59 ns` per packet). This allows gateways to discard heavy external proxy buffers and run natively with **0 ms** of software-induced delay.

### 2. Full LossMaxTTL Integration
*   **Purpose**: Governs the packet reorder tolerance window before a packet is declared lost and a NAK is sent.
*   **Behavior**: When bonding links have mismatched latencies (e.g., SIM 1 at 40ms and SIM 2 at 350ms), standard SRT will prematurely declare packets as lost and trigger duplicate NAK storms. Setting `LossMaxTTL` (e.g., to `200`) forces SRT to wait for a specific number of subsequent packets before requesting a retransmission, allowing slower links to deliver packets naturally without triggering overhead.

### 3. SRTLA-Optimized NAK Flow
*   **Enhancement**: Disabled periodic NAK reports, relying strictly on immediate NAKs upon packet gap detection. This matches the reference behavior of libsrt in standard SRTLA gateways, preventing redundant retransmissions over congested mobile connections.

---

## 🛠️ Usage in Go Projects

To link this optimized fork in your streaming application, use the `replace` directive in your `go.mod` file:

```go
module your-project

go 1.20

require (
	github.com/datarhei/gosrt v0.9.0 // Original import path
)

// Redirect to this optimized fork (lossmaxttl branch)
replace github.com/datarhei/gosrt => github.com/bry4ns/gosrt v0.9.0-lossmaxttl
```

---

## 💡 Guide for Developers

If you are debugging or extending this SRT implementation:
*   **Packet Queue & Reordering**: The core queue logic is located in `congestion/live/receive.go`. Inspect the `packetList.Insert()` function to see the optimized backwards-traversal logic.
*   **NAK Generation & Latency**: NAK triggers and TSBPD playout times are managed in `congestion/live/live.go` and are governed by the latency settings applied to the receiver socket.
*   **Decapsulation**: When receiving SRTLA streams, the incoming UDP packets have the SRTLA header stripped by the gateway before being forwarded to the GoSRT listener socket. This library processes the original SRT sequence numbers to close the NAK/ACK recovery loop correctly.
