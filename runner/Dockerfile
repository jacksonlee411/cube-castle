FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates curl jq git sudo bash libicu70 libssl3 tar \
  && rm -rf /var/lib/apt/lists/*

RUN useradd -m runner && echo "runner ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/runner
WORKDIR /home/runner

# Install GitHub Actions Runner (version pinned)
ARG RUNNER_VERSION=2.315.0
ENV RUNNER_VERSION=${RUNNER_VERSION}
# Resilient download with retries and archive validation (network may be flaky)
RUN bash -lc 'set -e; \
  url=https://github.com/actions/runner/releases/download/v${RUNNER_VERSION}/actions-runner-linux-x64-${RUNNER_VERSION}.tar.gz; \
  for i in $(seq 1 6); do \
    echo \"📥 Download attempt $i: $url\"; \
    curl -fL --retry 5 --retry-delay 3 --retry-max-time 600 -o actions-runner.tar.gz \"$url\" || true; \
    if tar tzf actions-runner.tar.gz >/dev/null 2>&1; then \
      echo \"✅ archive ok\"; break; \
    else \
      echo \"⚠️  invalid or partial archive, retrying...\"; rm -f actions-runner.tar.gz; sleep 5; \
    fi; \
  done; \
  [ -f actions-runner.tar.gz ] || (echo \"❌ failed to download runner archive\" && exit 2); \
  tar xzf actions-runner.tar.gz && rm -f actions-runner.tar.gz && chown -R runner:runner /home/runner'

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

USER runner
ENV RUNNER_WORKDIR=/home/runner/_work
ENV EPHEMERAL=true
ENV DISABLE_AUTO_UPDATE=true

VOLUME ["/home/runner/_work"]
ENTRYPOINT ["/entrypoint.sh"]
