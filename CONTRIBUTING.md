# Contributing to flnd

## Building and testing

```sh
BUILD_TAGS="dev autopilotrpc chainrpc invoicesrpc neutrinorpc peersrpc routerrpc signrpc verrpc walletrpc watchtowerrpc wtclientrpc"
go build -tags="$BUILD_TAGS" ./...
go vet -tags="$BUILD_TAGS" ./...
go test -tags="$BUILD_TAGS" -timeout 20m ./...
```

These are the commands CI runs, so run them before opening a pull request. CI
excludes a handful of packages that need external infrastructure (a bitcoind or
btcd binary, a real neutrino node to sync against, or a Postgres instance) and
do not respect `go test -short`; see `.github/workflows/ci.yaml` for the list.

`make unit` and `make itest` also exist, but they need a `btcd` binary and a
database instance respectively, which is why CI uses plain `go test` with the
exclusions above.

## Pull requests

- Keep each change focused; split unrelated work into separate pull requests.
- Add a `CHANGELOG.md` entry for anything that changes behaviour, under the
  topmost `## [X.Y.Z]` heading in the matching `### Added` / `### Changed` /
  `### Fixed` subsection. If the last release just shipped and no heading is
  open yet, add one with the version the change warrants.
- Once your pull request has a number, append `(#N)` to the changelog bullets it
  introduces. The release notes are generated from that text, so a bullet
  without its reference loses the link back to the discussion.

Note `docs/release-notes/` is inherited from upstream lnd and describes
upstream's releases, not this fork's. `CHANGELOG.md` is where this project's
own releases are recorded.

## Versioning

There is no `VERSION` file. `CHANGELOG.md` is the only place the version is
recorded, and it is injected into the binaries at build time via
`build.AppVersion` — from the git tag for a release, and from the most recent
tag for a local `make build`.

Nothing else should hold a copy of the version. The components served
individually by `verrpc`, `GetInfo` and `flncli version` are parsed from that
one value; previously they came from hand-edited constants, which is how the
binaries ended up reporting 0.1.21-beta three releases later.

A plain `go build` injects nothing and reports `0.0.0-dev`. That is deliberate:
a development build should never be mistaken for a release. Builds through the
Makefile get `APP_VERSION` injected, so they report the most recent tag.

## How releases are cut

Releases are manual. `.github/workflows/release.yml` is `workflow_dispatch`-only
and does the tagging itself:

```sh
gh workflow run release.yml --repo flokiorg/flnd
```

It resolves the version from the topmost `## [X.Y.Z]` heading in `CHANGELOG.md`
(or from the optional `version` input, given as a bare number with no `v`),
re-runs the build/vet/test gate, extracts that changelog section as the release
notes, creates and pushes the annotated `vX.Y.Z` tag, then publishes the `flnd`
and `flncli` binaries with GoReleaser.

Do not create the tag by hand — the workflow creates it, and a manual tag would
collide with the one it makes.
