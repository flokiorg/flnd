#!/usr/bin/env bash

# exit from script if error was raised.
set -e

# error function is used within a bash function in order to send the error
# message directly to the stderr output and exit.
error() {
    echo "$1" > /dev/stderr
    exit 0
}

# return is used within bash function in order to return the value.
return() {
    echo "$1"
}

# set_default function gives the ability to move the setting of default
# env variable from docker file to the script thereby giving the ability to the
# user override it during container start.
set_default() {
    # docker initialized env variables with blank string and we can't just
    # use -z flag as usually.
    BLANK_STRING='""'

    VARIABLE="$1"
    DEFAULT="$2"

    if [[ -z "$VARIABLE" || "$VARIABLE" == "$BLANK_STRING" ]]; then

        if [ -z "$DEFAULT" ]; then
            error "You should specify default variable"
        else
            VARIABLE="$DEFAULT"
        fi
    fi

   return "$VARIABLE"
}


# Set the default network and default RPC path (if any).
DEFAULT_NETWORK="regtest"
if [ "$BACKEND" == "btcd" ]; then
    DEFAULT_NETWORK="simnet"
    DEFAULT_RPCCRTPATH="/rpc/rpc.cert"
fi

# Set default variables if needed.
NETWORK=$(set_default "$NETWORK" "$DEFAULT_NETWORK")
RPCCRTPATH=$(set_default "$RPCCRTPATH" "$DEFAULT_RPCCRTPATH")
RPCHOST=$(set_default "$RPCHOST" "blockchain")
RPCUSER=$(set_default "$RPCUSER" "devuser")
RPCPASS=$(set_default "$RPCPASS" "devpass")
DEBUG=$(set_default "$LND_DEBUG" "debug")
CHAIN=$(set_default "$CHAIN" "flokicoin")
HOSTNAME=$(hostname)

# CAUTION: DO NOT use the --noseedback for production/mainnet setups, ever!
# Also, setting --rpclisten to $HOSTNAME will cause it to listen on an IP
# address that is reachable on the internal network. If you do this outside of
# docker, this might be a security concern!

if [ "$BACKEND" == "flokicoind" ] || [ "$BACKEND" == "bitcoind" ]; then
    # Standardize BACKEND name for flags
    REAL_BACKEND="flokicoind"

    exec flnd \
        --noseedbackup \
        "--$CHAIN.active" \
        "--$CHAIN.$NETWORK" \
        "--$CHAIN.node"="$REAL_BACKEND" \
        "--$REAL_BACKEND.rpchost"="$RPCHOST" \
        "--$REAL_BACKEND.rpcuser"="$RPCUSER" \
        "--$REAL_BACKEND.rpcpass"="$RPCPASS" \
        "--$REAL_BACKEND.zmqpubrawblock"="tcp://$RPCHOST:28332" \
        "--$REAL_BACKEND.zmqpubrawtx"="tcp://$RPCHOST:28333" \
        "--rpclisten=$HOSTNAME:10005" \
        "--rpclisten=localhost:10005" \
        --debuglevel="$DEBUG" \
        "$@"
elif [ "$BACKEND" == "btcd" ]; then
    # Standardize BACKEND name for flags
    REAL_BACKEND="btcd"

    exec flnd \
        --noseedbackup \
        "--$CHAIN.active" \
        "--$CHAIN.$NETWORK" \
        "--$CHAIN.node"="$REAL_BACKEND" \
        "--$REAL_BACKEND.rpccert"="$RPCCRTPATH" \
        "--$REAL_BACKEND.rpchost"="$RPCHOST" \
        "--$REAL_BACKEND.rpcuser"="$RPCUSER" \
        "--$REAL_BACKEND.rpcpass"="$RPCPASS" \
        "--rpclisten=$HOSTNAME:10005" \
        "--rpclisten=localhost:10005" \
        --debuglevel="$DEBUG" \
        "$@"
else
    echo "Unknown backend: $BACKEND"
    exit 1
fi
