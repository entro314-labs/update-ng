# Multi-stage Dockerfile for update-ng
# Produces a minimal container with the update-ng binary

FROM scratch

# Copy the binary from the build stage
COPY update-ng /usr/local/bin/update-ng

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/update-ng"]

# Default command
CMD ["--help"]

# Labels
LABEL org.opencontainers.image.title="update-ng"
LABEL org.opencontainers.image.description="Modern system updater with beautiful TUI"
LABEL org.opencontainers.image.vendor="entro314-labs"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.source="https://github.com/entro314-labs/update-ng"
