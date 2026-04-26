# Sync Review Summary - Session 1

## Overview
Syncing `flnd` with `lnd` upstream from commit `4aef4d00a` up to approximately `e532d4e49`.

## Functional Areas

### 1. Core Protocol & Parameters
- **Commits:** `4bce4ebe0`, `3fb59b1b6`
- **Changes:** Increased `MinCLTVDelta` from 18 to 24.
- **Review:** Rebranding applied to comments and itest parameters. `bitcoin.timelockdelta` rebranded to `flokicoin.timelockdelta`.
- **Status:** Success.

### 2. Contractcourt & Re-org Logic
- **Commits:** `50a6c0b8d`, `5e014b7d5`, `2e83fb179`, `064a14894`, `399da8f72`, `e5ac1a063`, `5af593fdd`
- **Changes:** Massive overhaul of re-org aware close logic. Added `ChannelCloseConfs` and scaled confirmation logic.
- **Review:** Critical AI issue found where `contractcourt/chain_watcher_test.go` had committed conflict markers. Fixed manually. Multiple `btcsuite` imports were not rebranded in new test files.
- **Status:** Fixed.

### 3. RPC & CLI
- **Commits:** `05eed5cdb`
- **Changes:** Added `FailureDetail` enums for invoice/AMP failures.
- **Review:** Applied to `.proto` and rebranded generated `.pb.go` files.
- **Status:** Success.

### 4. Test Infrastructure & itest
- **Changes:** Recovery of `lntest` package.
- **Review:** Major issue where `lntest/harness_assertion.go` was truncated in a previous session, causing dozens of undefined method errors. Restored from upstream and rebranded.
- **Status:** Recovered.

### 5. Documentation & Release Notes
- **Review:** Release notes for `0.21.0` were applied but contained many "LND" and "bitcoin" references.
- **Status:** Partially rebranded.

## Detected Regressions / AI Issues
1. **Truncated Files:** `lntest/harness_assertion.go` was severely truncated, likely due to a botched replacement or copy-paste error in a previous session.
2. **Conflict Markers:** `contractcourt/chain_watcher_test.go` contained committed conflict markers.
3. **Inconsistent Rebranding:** New files from upstream (e.g., `lnwallet/confscale_test.go`) often use `btcutil` and `Satoshi` which need manual adjustment to `chainutil` and `Loki`.
4. **Logging Directives:** Introduced `%w` in `log.Errorf` which is not supported by the local logger (requires `%v`).

## Conclusion
The build is currently **green**. Test suites for `lnwallet` and `contractcourt` have been verified.
