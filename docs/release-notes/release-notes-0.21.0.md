# Release Notes
- [Bug Fixes](#bug-fixes)
- [New Features](#new-features)
    - [Functional Enhancements](#functional-enhancements)
    - [RPC Additions](#rpc-additions)
    - [lncli Additions](#lncli-additions)
- [Improvements](#improvements)
    - [Functional Updates](#functional-updates)
    - [RPC Updates](#rpc-updates)
    - [lncli Updates](#lncli-updates)
    - [Breaking Changes](#breaking-changes)
    - [Performance Improvements](#performance-improvements)
    - [Deprecations](#deprecations)
- [Technical and Architectural Updates](#technical-and-architectural-updates)
    - [BOLT Spec Updates](#bolt-spec-updates)
    - [Testing](#testing)
    - [Database](#database)
    - [Code Health](#code-health)
    - [Tooling and Documentation](#tooling-and-documentation)
- [Contributors (Alphabetical Order)](#contributors)

# Bug Fixes

* [Fixed `OpenChannel` with
  `fund_max`](https://github.com/flokiorg/flnd/pull/10488) to use the
  protocol-level maximum channel size instead of the user-configured
  `maxchansize`. The `maxchansize` config option is intended only for limiting
  incoming channel requests from peers, not outgoing ones.

- Chain notifier RPCs now [return the gRPC `Unavailable`
  status](https://github.com/flokiorg/flnd/pull/10352) while the
  sub-server is still starting. This allows clients to reliably detect the
  transient condition and retry without brittle string matching.

- [Fixed an issue](https://github.com/flokiorg/flnd/pull/10399) where the
  TLS manager would fail to start if only one of the TLS pair files (certificate
  or key) existed. The manager now correctly regenerates both files when either
  is missing, preventing "file not found" errors on startup.

- [Fixed race conditions](https://github.com/flokiorg/flnd/pull/10420) in
  the channel graph database. The `Node.PubKey()` and
  `ChannelEdgeInfo.NodeKey1/NodeKey2()` methods had check-then-act races when
  caching parsed public keys. Additionally, `DisconnectBlockAtHeight` was
  accessing the reject and channel caches without proper locking. The caching
  has been removed from the public key parsing methods, and proper mutex
  protection has been added to the cache access in `DisconnectBlockAtHeight`.

- [Fixed TLV decoders to reject malformed records with incorrect lengths](https://github.com/flokiorg/flnd/pull/10249). 
  TLV decoders now strictly enforce fixed-length requirements for Fee (8 bytes),
  Musig2Nonce (66 bytes), ShortChannelID (8 bytes), Vertex (33 bytes), and
  DBytes33 (33 bytes) records, preventing malformed TLV data from being
  accepted.

- [Fixed `MarkCoopBroadcasted` to correctly use the `local`
  parameter](https://github.com/flokiorg/flnd/pull/10532). The method was
  ignoring the `local` parameter and always marking cooperative close
  transactions as locally initiated, even when they were initiated by the remote
  peer.

- [Fixed a panic in the gossiper](https://github.com/flokiorg/flnd/pull/10463)
  when `TrickleDelay` is configured with a non-positive value. The configuration
  validation now checks `TrickleDelay` at startup and defaults it to 1
  millisecond if set to zero or a negative value, preventing `time.NewTicker`
  from panicking.

- [Fixed a shutdown
  deadlock](https://github.com/flokiorg/flnd/pull/10540) in the gossiper.
  Certain gossip messages could cause multiple error messages to be sent on a
  channel that was only expected to be used for a single message. The erring
  goroutine would block on the second send, leading to a deadlock at shutdown.

* [Fixed two follow-ups to the production taproot channels
  work](https://github.com/lightningnetwork/lnd/pull/10763). The RPC channel
  acceptor switch now maps `SIMPLE_TAPROOT_FINAL` (with every combination of
  the `scid-alias` / `zero-conf` modifiers) so final-taproot opens are
  reported to external acceptor clients with the correct commitment type
  instead of `UNKNOWN_COMMITMENT_TYPE`. The taproot RBF cooperative-close
  auto-enable is also narrowed to skip taproot-overlay channels, since the
  RBF close state machine does not yet thread through the `AuxCloser` hook
  that overlay channels rely on to build aux-aware close transactions.

* [Fixed `EstimateRouteFee`](https://github.com/lightningnetwork/lnd/pull/10771)
  to use independent probe payment hashes when probing multiple LSPs, preventing
  later probes from reusing the first probe's CLTV delta.

* [Restored insta-dispatch of `CLOSED_CHANNEL` on the first confirmation of a
  cooperative close](https://github.com/lightningnetwork/lnd/pull/10794).
  After the multi-conf reorg-aware close dispatch landed,
  `SubscribeChannelEvents` no longer emitted `CLOSED_CHANNEL` until the full
  required confirmation depth was reached. The chain watcher now fires an
  early `CLOSED_CHANNEL` event over the channel notifier as soon as the coop
  close spend lands on chain, restoring the v0.20.1 behavior, while the
  channel arbitrator suppresses the duplicate event that would otherwise be
  emitted from `MarkChannelClosed` at the final confirmation depth.

# New Features

- Basic Support for [onion messaging forwarding](https://github.com/flokiorg/flnd/pull/9868) 
  consisting of a new message type, `OnionMessage`. This includes the message's
  definition, comprising a path key and an onion blob, along with the necessary
  serialization and deserialization logic for peer-to-peer communication.

## Functional Enhancements

* [Added reorg protection for channel
  closes](https://github.com/flokiorg/flnd/pull/10331). Previously,
  channel closes were considered final immediately on spend detection with no
  confirmation waiting. Now, all channel closes require between 3 and 6
  confirmations, scaled linearly with channel capacity up to the maximum
  non-wumbo channel size (~0.168 FLC), with wumbo channels always requiring
  6 confirmations.

* [Added support for production (final) simple taproot
  channels](https://github.com/flokiorg/flnd/pull/9985) using the
  finalized taproot channel scripts with feature bits 80/81. Production taproot
  channels use optimized scripts (`OP_CHECKSIGVERIFY` instead of `OP_CHECKSIG` +
  `OP_DROP`) and a map-based nonce encoding in `channel_reestablish` and
  `revoke_and_ack` keyed by funding TXID, laying the groundwork for splice
  support. The nonce type is now auto-detected from the negotiated channel type
  rather than peer feature bits, ensuring correct behavior across all recovery
  and resynchronization paths. Taproot channels must be requested explicitly
  with `lncli openchannel --channel_type=taproot` (the bare `taproot` string
  now selects the production variant; `taproot-staging` opens the legacy
  staging variant, and `taproot-final` is kept as a deprecated alias for
  `taproot`), and must remain private until announced taproot channels are
  supported. The RPC `CommitmentType` enum gains a `TAPROOT` alias for
  `SIMPLE_TAPROOT_FINAL` so new RPC clients can use the same short name.

* [Added taproot channel support for RBF cooperative
  close](https://github.com/flokiorg/flnd/pull/10063). The new RBF-based
  cooperative close protocol (enabled with `--protocol.rbf-coop-close`) now
  fully supports simple taproot channels. This includes MuSig2 partial signature
  handling with the JIT (just-in-time) nonce pattern, where closer nonces are
  bundled with signatures in `ClosingComplete` and closee nonces are rotated via
  `NextCloseeNonce` in `ClosingSig` for each RBF iteration. The implementation
  prevents nonce reuse across RBF rounds by storing the `MusigPartialSig` in the
  protocol state machine and invalidating nonces after each signing round
  completes.
## RPC Additions

* [Added support for coordinator-based MuSig2 signing
  patterns](https://github.com/flokiorg/flnd/pull/10436) with two new
  RPCs: `MuSig2RegisterCombinedNonce` allows registering a pre-aggregated
  combined nonce for a session (useful when a coordinator aggregates all nonces
  externally), and `MuSig2GetCombinedNonce` retrieves the combined nonce after
  it becomes available. These methods provide an alternative to the standard
  `MuSig2RegisterNonces` workflow and are only supported in MuSig2 v1.0.0rc2.

* The `EstimateFee` RPC now supports [explicit input
  selection](https://github.com/flokiorg/flnd/pull/10296). Users can
  specify a list of inputs to use as transaction inputs via the new
  `inputs` field in `EstimateFeeRequest`.

## lncli Additions

* The `estimatefee` command now supports the `--utxos` flag to specify explicit
  inputs for fee estimation.

# Improvements
## Functional Updates

* [Added support](https://github.com/flokiorg/flnd/pull/9432) for the
  `upfront-shutdown-address` configuration in `flnd.conf`, allowing users to
  specify an address for cooperative channel closures where funds will be sent.
  This applies to both funders and fundees, with the ability to override the
  value during channel opening or acceptance.

* Rename [experimental endorsement signal](https://github.com/lightning/blips/blob/a833e7b49f224e1240b5d669e78fa950160f5a06/blip-0004.md)
  to [accountable](https://github.com/lightningnetwork/lnd/pull/10367) to match
  the latest [proposal](https://github.com/lightning/blips/pull/67).

## RPC Updates

* routerrpc HTLC event subscribers now receive specific failure details for
  invoice-level validation failures, avoiding ambiguous `UNKNOWN` results. [#10520](https://github.com/flokiorg/flnd/pull/10520)

## lncli Updates

## Breaking Changes

## Performance Improvements

* Let the [channel graph cache be populated
  asynchronously](https://github.com/flokiorg/flnd/pull/10065) on
  startup. While the cache is being populated, the graph is still available for
  queries, but all read queries will be served from the database until the cache
  is fully populated. This new behaviour can be opted out of via the new
  `--db.sync-graph-cache-load` option.

* [Invoice pagination queries no longer use
  `OFFSET`](https://github.com/flokiorg/flnd/pull/10700). The five
  invoice filter queries previously used `LIMIT+OFFSET` for internal batching,
  which requires the database to scan and discard all preceding rows on every
  page. All pagination is now cursor-based (`WHERE id >= cursor`), making every
  page an efficient primary-key range scan regardless of how deep into the
  result set the query is.

* [Replace the catch-all `FilterInvoices` SQL query with five focused,
  index-friendly queries](https://github.com/flokiorg/flnd/pull/10601)
  (`FetchPendingInvoices`, `FilterInvoicesBySettleIndex`,
  `FilterInvoicesByAddIndex`, `FilterInvoicesForward`,
  `FilterInvoicesReverse`). The old query used `col >= $param OR $param IS
  NULL` predicates and a `CASE`-based `ORDER BY` that prevented SQLite's query
  planner from using indexes, causing full table scans. Each new query carries
  only the parameters it actually needs and uses a direct `ORDER BY`, allowing
  the planner to perform efficient index range scans on the invoice table.

* [Fix full table scans on the HTLC settlement
    hot path](https://github.com/flokiorg/flnd/pull/10619).
    Replace the catch-all `GetInvoice` query (which used `OR $1 IS NULL`
    predicates that forced full table scans) with three dedicated queries
    targeting uniquely-constrained columns. Also drop four redundant indexes
    that duplicated UNIQUE constraints or were never used as query filters.

* [Optimize the v1 node horizon
    query](https://github.com/flokiorg/flnd/pull/10692). Split the
    `GetNodesByLastUpdateRange` query into separate all-nodes and public-only
    variants, removing a dynamic `COALESCE`/`OR` branch that defeated the query
    planner. The public-only `EXISTS` check is rewritten as two direct index
    probes instead of `node_id_1 OR node_id_2`. Supporting indexes are upgraded
    to composite keys matching the full query shapes. On SQLite, the hot
    public-only path sees a ~42% speedup; on the previous code it could stall
    for minutes.

* [Tombstone closed channels on KV-over-SQL
  backends](https://github.com/flokiorg/flnd/pull/10780). Closing a
  long-lived channel previously issued a single `DeleteNestedBucket` inside
  the close transaction. On the kvdb-on-SQL schema (sqlite, postgres) that
  delete fans out into a row-by-row `ON DELETE CASCADE` over the channel's
  revocation log and forwarding-package bucket, holding the database
  write-lock for many seconds — long enough on channels with millions of
  states to stall HTLC forwarding, time out htlcswitch retries, and trigger
  force-close cycles. `CloseChannel` now skips the cascading delete on
  these backends; the outpoint-index flip from `outpointOpen` to
  `outpointClosed` (already performed by the existing close path) is the
  authoritative closed-channel marker, and every reader of the open-channel
  bucket consults it before treating a channel as open. The bulk historical
  state — the chanBucket itself, the revocation log, and the per-channel
  forwarding-package bucket — remains on disk for the channel's lifetime in
  this database and is reclaimed wholesale by the upcoming native-SQL
  channel-state migration. bbolt and etcd retain the synchronous one-shot
  close path, where nested-bucket deletion is already cheap.

  > ⚠️ **Downgrade warning.** On sqlite/postgres, once a channel is
  > closed under this build the chanBucket and its nested state remain
  > on disk; the close is signalled only by the `outpointClosed` flip
  > in the outpoint index. Earlier `flnd` releases do not consult that
  > flip when iterating `openChannelBucket`, so downgrading to a
  > pre-0.21 binary after closing channels on these backends will
  > resurrect those channels as open in `listchannels`,
  > `pendingchannels`, and the chain-watch path. Operators who close
  > channels on sqlite/postgres after upgrading should treat the
  > upgrade as one-way for that database; bbolt and etcd users are unaffected 
  > because the close path on those backends still deletes the chanBucket.

## Deprecations

### ⚠️ **Warning:** The deprecated fee rate option `--sat_per_byte` will be removed in release version **0.22**

  The deprecated `--sat_per_byte` option will be fully removed. This flag was
  originally deprecated and hidden from the lncli commands in v0.13.0
  ([PR#4704](https://github.com/flokiorg/flnd/pull/4704)). Users should
  migrate to the `--sat_per_vbyte` option, which correctly represents fee rates
  in terms of virtual bytes (vbytes).
  
  Internally `--sat_per_byte` was treated as sat/vbyte, this meant the option
  name was misleading and could result in unintended fee calculations. To avoid 
  further confusion and to align with ecosystem terminology, the option will be
  removed.

  The following RPCs will be impacted:

  | RPC Method | Messages | Removed Option | 
  |----------------------|----------------|-------------|
| [`lnrpc.CloseChannel`](https://lightning.engineering/api-docs/api/lnd/lightning/close-channel/) | [`lnrpc.CloseChannelRequest`](https://lightning.engineering/api-docs/api/lnd/lightning/close-channel/#lnrpcclosechannelrequest) | sat_per_byte
| [`lnrpc.OpenChannelSync`](https://lightning.engineering/api-docs/api/lnd/lightning/open-channel-sync/) | [`lnrpc.OpenChannelRequest`](https://lightning.engineering/api-docs/api/lnd/lightning/open-channel-sync/#lnrpcopenchannelrequest) | sat_per_byte 
| [`lnrpc.OpenChannel`](https://lightning.engineering/api-docs/api/lnd/lightning/open-channel/) | [`lnrpc.OpenChannelRequest`](https://lightning.engineering/api-docs/api/lnd/lightning/open-channel/#lnrpcopenchannelrequest) | sat_per_byte
| [`lnrpc.SendCoins`](https://lightning.engineering/api-docs/api/lnd/lightning/send-coins/) | [`lnrpc.SendCoinsRequest`](https://lightning.engineering/api-docs/api/lnd/lightning/send-coins/#lnrpcsendcoinsrequest) | sat_per_byte
| [`lnrpc.SendMany`](https://lightning.engineering/api-docs/api/lnd/lightning/send-many/) | [`lnrpc.SendManyRequest`](https://lightning.engineering/api-docs/api/lnd/lightning/send-many/#lnrpcsendmanyrequest) | sat_per_byte
| [`walletrpc.BumpFee`](https://lightning.engineering/api-docs/api/lnd/wallet-kit/bump-fee/) | [`walletrpc.BumpFeeRequest`](walletrpc.BumpFeeRequest) | sat_per_byte

# Technical and Architectural Updates
## BOLT Spec Updates

## Testing

* [Added unit tests for TLV length validation across multiple packages](https://github.com/flokiorg/flnd/pull/10249). 
  New tests  ensure that fixed-size TLV decoders reject malformed records with
  invalid lengths, including roundtrip tests for Fee, Musig2Nonce,
  ShortChannelID and Vertex records.

## Database

* Freeze the [graph SQL migration 
  code](https://github.com/flokiorg/flnd/pull/10338) to prevent the 
  need for maintenance as the sqlc code evolves. 

## Code Health

## Tooling and Documentation

# Contributors (Alphabetical Order)

* Boris Nagaev
* Elle Mouton
* Erick Cestari
* hieblmi
* Matt Morehouse
* Mohamed Awnallah
* Nishant Bansal
* Pins
