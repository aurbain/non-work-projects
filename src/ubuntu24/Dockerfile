# Dockerfile for Ubuntu 24.04 LTS (Noble)
FROM ubuntu:noble

# Update package list and install any required packages
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
        # Add your packages here
    && rm -rf /var/lib/apt/lists/*

# Set default shell to bash
SHELL ["/bin/bash", "-c"]

# Optional: set working directory
WORKDIR /app

# Copy application files if needed
# COPY . /app

# Default command
CMD ["bash"]
