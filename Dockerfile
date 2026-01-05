# If you change this please also update GO_VERSION in Makefile (then run
# `make lint` to see where else it needs to be updated as well).
FROM golang:1.25.3-alpine as builder

# Force Go to use the cgo based DNS resolver. This is required to ensure DNS
# queries required to connect to linked containers succeed.
ENV GODEBUG netdns=cgo

# Pass a tag, branch or a commit using build-arg.  This allows a docker
# image to be built from a specified Git state.  The default image
# will use the Git tip of master by default.
ARG checkout="master"
ARG git_url="https://github.com/flokiorg/flnd"

# Install dependencies and build the binaries.
RUN apk add --no-cache --update alpine-sdk \
    git \
    make \
    gcc \
&&  git clone $git_url /go/src/github.com/flokiorg/flnd \
&&  cd /go/src/github.com/flokiorg/flnd \
&&  git checkout $checkout \
&&  make release-install

# Start a new, final image.
FROM alpine as final

# Define a root volume for data persistence.
VOLUME /root/.flnd

# Add utilities for quality of life and SSL-related reasons. We also require
# curl and gpg for the signature verification script.
RUN apk --no-cache add \
    bash \
    jq \
    ca-certificates \
    gnupg \
    curl

# Copy the binaries from the builder image.
COPY --from=builder /go/bin/flncli /bin/
COPY --from=builder /go/bin/flnd /bin/
COPY --from=builder /go/src/github.com/flokiorg/flnd/scripts/verify-install.sh /
COPY --from=builder /go/src/github.com/flokiorg/flnd/scripts/keys/* /keys/

# Store the SHA256 hash of the binaries that were just produced for later
# verification.
RUN sha256sum /bin/flnd /bin/flncli > /shasums.txt \
  && cat /shasums.txt

# Expose flnd ports (p2p, rpc).
EXPOSE 5521 10005

# Specify the start command and entrypoint as the flnd daemon.
ENTRYPOINT ["flnd"]
