# Commit Review - df0bd94b9

## Commit Info
- **Hash:** df0bd94b9141fe8976f544ed80f7c00396d56e50
- **Message:** scripts: update bw-compat test LND base version

## Rebranding Check
- [x] LND -> flnd: The variable remains `LND_LATEST_VERSION` because it pulls from the `lightninglabs/lnd` image for compatibility testing.
- [x] BITCOIND -> flokicoind: Image `lightninglabs/bitcoin-core` is still used in this specific test environment.

## Technical Integrity
- [x] Builds correctly
- [x] Scripts verified (referenced in docker-compose.yaml)

## AI Assessment
- [x] No botched search/replace
- [x] Context preserved

## Notes
- Updated base version for backward compatibility tests to v0.20.0-beta.
