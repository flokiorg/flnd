## Flokicoin Lightning Network Daemon (FLND)

The Flokicoin Lightning Network Daemon (FLND) is a fork of
Lightning Labs' lnd tailored for the Flokicoin blockchain. It implements a
complete [Lightning Network](https://lightning.network) node for Flokicoin with
pluggable back-end chain services, including the Flokicoin full node and a
light client:

- Full node: [`go-flokicoin`](https://github.com/flokiorg/go-flokicoin) (fork of btcd)
- Light client: [`flokicoin-neutrino`](https://github.com/flokiorg/flokicoin-neutrino)

FLND exports the familiar gRPC and REST APIs from lnd and reuses many of the
same internal libraries, adapted for Flokicoin. In its current state FLND is
capable of:
  * Creating channels.
  * Closing channels.
  * Completely managing all channel states (including the exceptional ones!).
  * Maintaining a fully authenticated+validated channel graph.
  * Performing path finding within the network, passively forwarding incoming payments.
  * Sending outgoing [onion-encrypted payments](https://github.com/lightningnetwork/lightning-onion)
through the network.
  * Updating advertised fee schedules.
  * Automatic channel management (autopilot).

## Lightning Network Specification Compliance
FLND follows the [Lightning Network specification
(BOLTs)](https://github.com/lightningnetwork/lightning-rfc). BOLT stands for:
Basis of Lightning Technology. The specifications are currently being drafted
by several groups of implementers based around the world including the
developers of lnd/flnd. The set of specification documents as well as our
implementation of the specification are still a work-in-progress. With that
said, the current status of FLND's BOLT compliance is:

  - [X] BOLT 1: Base Protocol
  - [X] BOLT 2: Peer Protocol for Channel Management
  - [X] BOLT 3: Bitcoin Transaction and Script Formats
  - [X] BOLT 4: Onion Routing Protocol
  - [X] BOLT 5: Recommendations for On-chain Transaction Handling
  - [X] BOLT 7: P2P Node and Channel Discovery
  - [X] BOLT 8: Encrypted and Authenticated Transport
  - [X] BOLT 9: Assigned Feature Flags
  - [X] BOLT 10: DNS Bootstrap and Assisted Node Location
  - [X] BOLT 11: Invoice Protocol for Lightning Payments

## Developer Resources

- APIs: FLND exposes both HTTP REST and gRPC ([grpc.io](https://grpc.io/)).
  The request/response surfaces are largely compatible with lnd. See
  `lnrpc/` and `docs/grpc/` in this repository for the latest service
  definitions and usage examples.
- Go reference: https://pkg.go.dev/github.com/flokiorg/flnd
- Sample configuration: `sample-lnd.conf`

Note: Some documentation and configuration files retain the `lnd` binary name
and paths for compatibility. In this repository the binaries are still named
`lnd` and `lncli`.

First-time contributors are encouraged to start with code review
(`docs/review.md`) before opening Pull Requests.

## Installation

For building from source and supported backends, see `docs/INSTALL.md`.
Quick start for local builds:

1) Install Go (see `docs/INSTALL.md` for the required version).
2) Build and install binaries:
   - `make install-binaries` (installs `lnd` and `lncli` to `GOPATH/bin`)
   - or `make build` for local debug builds (`lnd-debug`, `lncli-debug`).

## Docker
To run FLND with Docker, see `docs/DOCKER.md`.

## Networks
FLND supports the same set of networks as lnd (mainnet, testnet, regtest,
simnet). The active chain name is `flokicoin`.

## Safety

When operating a mainnet FLND node, please refer to the [operational safety
guidelines](docs/safety.md). It is important to note that FLND is still
**beta** software and that ignoring these operational guidelines can lead to
loss of funds.

## Security

Please report vulnerabilities privately using GitHub Security Advisories for
this repository. Avoid filing public issues for security reports.

## Further reading
* Installation: `docs/INSTALL.md`
* Docker: `docs/DOCKER.md`
* Safety: `docs/safety.md`
* Contribution guide: `docs/code_contribution_guidelines.md`

## License and Credits

- License: MIT (see `LICENSE`).
- FLND is based on and includes substantial code from
  https://github.com/lightningnetwork/lnd. We thank the original authors and
  contributors.
