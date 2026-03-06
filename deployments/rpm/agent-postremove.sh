#!/bin/bash
set -e

# RPM %postun — $1 = 0 on final removal, 1 on upgrade.
# Remove user and data directories only on complete removal.
if [ "$1" -eq 0 ]; then
    if id distributed-encoder-agent &>/dev/null 2>&1; then
        userdel distributed-encoder-agent 2>/dev/null || true
    fi
    if getent group distributed-encoder-agent &>/dev/null; then
        groupdel distributed-encoder-agent 2>/dev/null || true
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
