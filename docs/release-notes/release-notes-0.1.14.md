# Release Notes 0.1.14-beta

- [Improvements](#improvements)
- [Technical and Architectural Updates](#technical-and-architectural-updates)

# Improvements

## Functional Updates

- **Flokicoin Network Parameter Tuning**: Comprehensive update of default network parameters to align with Flokicoin's 1-minute block time (10x faster than Bitcoin).
    -   **Fee Estimation**: `MaxBlockTarget` increased to 10080 blocks (1 week) to ensure accurate fee estimates for longer time horizons.
    -   **Safety Margins**: `RemoteDelay` and `LocalCSVDelay` defaults adjusted to 10080 blocks (1 week) and 1440 blocks (1 day) respectively, ensuring adequate time for breach detection and recovery on the faster chain.
    -   **Routing**: `TimeLockDelta` increased to 400 blocks to provide a ~6.5 hour safety buffer for HTLC forwarding.
    -   **Gossip**: `TrickleDelay` reduced to 9s (from 90s) to match the faster block rate and ensure timely network propagation.
    -   **Usability**: Default **Max Channel Size** increased to 5 FLC (previously 0.16 FLC), allowing for larger payment channels by default.

# Technical and Architectural Updates

## Code Health

- **Refactoring**: Renamed internal funding constants from `Btc` prefix to `Flc` (e.g., `MaxBtcFundingAmount` -> `MaxFlcFundingAmount`) to accurately reflect the codebase's focus on Flokicoin and avoid confusion.
