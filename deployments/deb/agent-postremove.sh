#!/bin/bash
set -e

# Only remove user and data directories when purging (apt purge / dpkg --purge).
# On a normal removal (apt remove) configuration files and the user are kept.
if [ "$1" = "purge" ]; then
    if id distributed-encoder-agent &>/dev/null 2>&1; then
        deluser --remove-home distributed-encoder-agent 2>/dev/null || true
    fi

    rm -rf \
        /var/lib/distributed-encoder-agent \
        /var/log/distributed-encoder-agent

    # Remove agent-specific config files but leave /etc/distributed-encoder/
    # in case the controller package is also installed on this host.
    rm -f \
        /etc/distributed-encoder/agent.yaml \
        /etc/distributed-encoder/agent-environment

    if [ -d /run/systemd/system ]; then
        systemctl daemon-reload >/dev/null 2>&1 || true
    fi
fi
