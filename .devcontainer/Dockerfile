FROM mcr.microsoft.com/devcontainers/go:1-1.21-bullseye

# Instalar ferramentas adicionais úteis para desenvolvimento
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
    curl \
    jq \
    make \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/*

# Instalar ferramentas Go comumente usadas
RUN go install -v \
    github.com/cweill/gotests/gotests@latest \
    github.com/fatih/gomodifytags@latest \
    github.com/josharian/impl@latest \
    github.com/haya14busa/goplay/cmd/goplay@latest \
    github.com/go-delve/delve/cmd/dlv@latest \
    golang.org/x/tools/gopls@latest \
    && rm -rf /go/pkg/mod/* /go/src/*
