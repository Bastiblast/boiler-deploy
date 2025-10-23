#!/bin/bash
set -e

# Generate SSH host keys if they don't exist
ssh-keygen -A

# Start SSH service
/usr/sbin/sshd -D
