#!/bin/bash
set -e

# RPM %postun — $1 = 0 on final removal, 1 on upgrade.
# Remove user, data directories, and config only on complete removal.
if [ "$1" -eq 0 ]; then
    if id distributed-encoder &>/dev/null 2>&1; then
        userdel distributed-encoder 2>/dev/null || true
    fi
    if getent group distributed-encoder &>/dev/null; then
        groupdel distributed-encoder 2>/dev/null || true
    fi

    rm -rf \
        /var/lib/distributed-encoder \
        /var/log/distributed-encoder \
        /etc/distributed-encoder

    if [ -d /run/systemd/system ]; then
        systemctl daemon-reload >/dev/null 2>&1 || true
    fi
fi
