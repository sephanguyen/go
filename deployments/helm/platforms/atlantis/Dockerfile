FROM ghcr.io/runatlantis/atlantis:v0.24.2

RUN wget -q https://github.com/gruntwork-io/terragrunt/releases/download/v0.45.2/terragrunt_linux_amd64 && \
  mv terragrunt_linux_amd64 terragrunt && \
  chmod +x terragrunt && \
  mv terragrunt /usr/local/bin/ && \
  apk add postgresql-client
