#!/bin/bash
set -e

# Create the system user and group (no login shell, no home directory)
if ! id distributed-encoder-agent &>/dev/null 2>&1; then
    adduser --system --group --no-create-home \
        --home /var/lib/distributed-encoder-agent \
        --shell /usr/sbin/nologin \
        distributed-encoder-agent
fi

# Create runtime directories with correct ownership
install -d -o distributed-encoder-agent -g distributed-encoder-agent -m 750 \
    /var/lib/distributed-encoder-agent \
    /var/lib/distributed-encoder-agent/work \
    /var/log/distributed-encoder-agent

# Create optional environment file if it doesn't exist yet
if [ ! -f /etc/distributed-encoder/agent-environment ]; then
    install -o root -g distributed-encoder-agent -m 640 \
        /dev/null /etc/distributed-encoder/agent-environment
fi

# Fix ownership of the certs directory (created by nFPM as root)
chown distributed-encoder-agent:distributed-encoder-agent \
    /etc/distributed-encoder/certs 2>/dev/null || true

# Reload systemd and enable the service
if [ -d /run/systemd/system ]; then
    systemctl daemon-reload >/dev/null 2>&1 || true
    systemctl enable distributed-encoder-agent >/dev/null 2>&1 || true

    # Do NOT auto-start: the agent requires TLS certs and a configured
    # controller address before it can connect.  The operator must start
    # the service manually after completing configuration.
fi

echo ""
echo "================================================================"
echo "  Distributed Encoder Agent installed"
echo "================================================================"
echo ""
echo "  Before starting the service, complete these steps:"
echo ""
echo "  1. Edit /etc/distributed-encoder/agent.yaml"
echo "     Required settings:"
echo "       controller.address      Controller hostname:port (gRPC)"
echo "       controller.tls.*        mTLS certificate paths"
echo ""
echo "  2. Place TLS certificates in /etc/distributed-encoder/certs/"
echo "     Required files: ca.crt  agent.crt  agent.key"
echo "     See: https://github.com/badskater/distributed-encoder/blob/main/DEPLOYMENT.md"
echo ""
echo "  3. Start the service:"
echo "     systemctl start distributed-encoder-agent"
echo "     systemctl status distributed-encoder-agent"
echo ""
echo "  Logs:  journalctl -u distributed-encoder-agent -f"
echo "================================================================"
echo ""
