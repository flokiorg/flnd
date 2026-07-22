# If you change this please also update GO_VERSION in Makefile (then run
# `make lint` to see where else it needs to be updated as well).
FROM golang:1.26.1-alpine AS builder

LABEL maintainer="Olaoluwa Osuntokun <laolu@lightning.engineering>"

# Force Go to use the cgo based DNS resolver. This is required to ensure DNS
# queries required to connect to linked containers succeed.
ENV GODEBUG netdns=cgo

# Install dependencies.
RUN apk add --no-cache --update alpine-sdk \
    bash \
    git \
    make 

# Copy in the local repository to build from.
COPY . /go/src/github.com/flokiorg/flnd

# Bring in local github.com/flokiorg/walletd and github.com/flokiorg/
# flokicoin-neutrino checkouts as Go workspace members, so flnd builds
# against whatever's on disk there (including uncommitted or unreleased
# committed changes) instead of the versions pinned in go.mod/go.sum. This
# mirrors what flokiorg/go.work already does for native `go build` on the
# host: this fork's flnd/walletd/flokicoin-neutrino evolve in lockstep, are
# usually checked out side by side, and — as on the host — walletd's own
# NeutrinoChainService interface only matches *neutrino.ChainService when
# built against the same local flokicoin-neutrino checkout, not the older
# version flnd/walletd's go.mod happen to have pinned. Requires passing
# --build-context walletd=../walletd
# --build-context flokicoin-neutrino=../flokicoin-neutrino
# (wired in already via the Justfile and the consuming
# docker-compose.dev.yml files).
COPY --from=walletd . /go/src/github.com/flokiorg/walletd
COPY --from=flokicoin-neutrino . /go/src/github.com/flokiorg/flokicoin-neutrino

#  Install/build flnd, using the same tag set as the release build
# (make/release_flags.mk RELEASE_TAGS) so a locally-built dev image has the
# same subservers (routerrpc has no build tag so it's always included).
RUN cd /go/src/github.com/flokiorg/flnd \
    &&  go work init && go work use . ../walletd ../flokicoin-neutrino \
    &&  make release-install

# Start a new, final image to reduce size.
FROM alpine AS final

# Add bash for quality of life, and ca-certificates since flnd makes outbound
# HTTPS calls (e.g. fee.url).
RUN apk add --no-cache \
    bash \
    ca-certificates

# Define a root volume for data persistence.
VOLUME /root/.flnd

# Copy the binaries from the builder image.
COPY --from=builder /go/bin/flncli /bin/
COPY --from=builder /go/bin/flnd /bin/

# Expose flnd ports (p2p, rpc).
EXPOSE 5521 10005

ENTRYPOINT ["flnd"]
