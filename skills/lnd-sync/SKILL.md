---
name: lnd-sync
description: Specialized skill for syncing flnd (Flokicoin fork) with upstream lnd. Helps identify missing commits, apply batches, and handle rebranding/parameter adjustments.
---

# LND Sync Skill

This skill facilitates the ongoing synchronization between `flnd` and its upstream parent repository, `lnd`.

## Core Objectives

1.  **Identify Sync Gaps**: Detect commits in `lnd` that have not yet been ported to `flnd`.
2.  **Batch Processing**: Apply commits in manageable batches (typically 10-50 commits).
3.  **Rebranding & Parameters**: Ensure all incoming code adheres to Flokicoin branding (flnd, flokicoin, etc.) and preserves Flokicoin-specific constant overrides.
4.  **Verification**: Validate each batch with builds and tests.

## Workflow

### 1. Research & Identification
- Compare the last synced commit in `flnd` with the current `master` branch of the `lnd` clone in `forks/lnd`.
- Maintain a list of pending commits in `reports/batch_N.txt`.

### 2. Strategy
- For each batch, determine if it contains sensitive areas (e.g., hardcoded "bitcoin" strings, network parameters).
- Plan the order of application (usually sequential).

### 3. Execution (Plan -> Act -> Validate)
- **Plan**: Choose a subset of commits from the current batch.
- **Act**:
    - `git cherry-pick` the commits.
    - Run rebranding scripts or manual `replace` calls to maintain `flnd` identity.
    - Check for conflicts and resolve them using Flokicoin-first logic.
- **Validate**:
    - Run `go build ./...`.
    - Run relevant tests (especially those affected by the new commits).

## Rebranding Rules
- `lnd` -> `flnd`
- `btcsuite` -> `flokiorg` (where applicable for dependencies)
- `bitcoin` -> `flokicoin`
- Port: `9735` -> `15213` (Mainnet)
- Default directory: `.lnd` -> `.flnd`

## Progress Tracking
- Use `reports/` folder to store status files for each sync session.
- Each report should list:
    - Range of commits processed.
    - Major changes introduced.
    - Any manual adjustments or rebranding fixes applied.
