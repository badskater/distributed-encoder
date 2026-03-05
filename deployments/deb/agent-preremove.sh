#!/bin/bash
set -e

if [ -d /run/systemd/system ]; then
    systemctl stop    distributed-encoder-agent 2>/dev/null || true
    systemctl disable distributed-encoder-agent 2>/dev/null || true
fi
