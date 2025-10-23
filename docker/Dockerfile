# Dockerfile for Ansible testing - Debian with SSH
FROM debian:12

# Install SSH server and required packages
RUN apt-get update && \
    apt-get install -y \
    openssh-server \
    sudo \
    python3 \
    python3-pip \
    systemd \
    systemd-sysv \
    && rm -rf /var/lib/apt/lists/*

# Create SSH directory
RUN mkdir -p /var/run/sshd

# Configure SSH
RUN sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config && \
    sed -i 's/#PasswordAuthentication yes/PasswordAuthentication yes/' /etc/ssh/sshd_config && \
    sed -i 's/#PubkeyAuthentication yes/PubkeyAuthentication yes/' /etc/ssh/sshd_config

# Create debian user with sudo access (matching Scaleway default)
RUN useradd -m -s /bin/bash debian && \
    echo "debian:debian" | chpasswd && \
    usermod -aG sudo debian && \
    echo "debian ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/debian

# Setup SSH keys directory for debian user
RUN mkdir -p /home/debian/.ssh && \
    chown -R debian:debian /home/debian/.ssh && \
    chmod 700 /home/debian/.ssh

# Expose SSH port
EXPOSE 22

# Copy entrypoint script
COPY docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

# Use systemd as init system
CMD ["/usr/local/bin/docker-entrypoint.sh"]
