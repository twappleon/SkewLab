# SkewLab v0.3.0 — release notes (2026-11-18)

This is a real source-code release, not an on-chain or live trading release. It does NOT claim any profitable real-market result.

### Added
- Public Polymarket CLOB WS client with frame validation, masked writes, ping/pong, reconnect/backoff and market rollover subscription management.
- Bid/ask L2 snapshot + price-level delta book state; stale/out-of-order/crossed book rejection, REST fallback.
- Optional Chainlink 60s TWAP via authenticated PolyBolt (first auth includes Chainlink key), verified subscription/sequence, recorded real samples and exact-boundary audit.
- Trade-event-based passive maker queue *research only*; zero maker PnL / zero maker simulated fills.
- Read-only feed diagnostics and endpoint, Dashboard status panels, RWD improvements.
- Durable synced ledger, JSONL log rotation, restart recovery and corrupted-ledger fail-closed behavior.
- Independent Chainlink-to-Gamma mismatch halt (only when exact TWAP edges are actually observed).
- Expanded unit, integration, race, browser and failure-path tests.

### Verified
- Standard Go compile, Go test and race detector, static vet, synthetic demo REST API, local responsive desktop/mobile browser.
- Public interface at local http://localhost:8193 only during testing, not remotely accessible.

### Not verified
- Public production DNS/API connectivity, PolyBolt credential authentication, sustained real WebSocket traffic and feed lag.
- Actual maker fills or net profitability.
- New GitHub commit / CI confirmation or public Railway hosting.
- 7–14 day live-feed forward test.
