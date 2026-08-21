# Multi-stage build: compile static Go binary, then produce minimal image
FROM golang:1.26.6 AS builder
WORKDIR /src

# cache deps
# go.sum may not exist for projects without external dependencies.
COPY go.mod ./
RUN go mod download

# copy only required source for a deterministic build context
COPY cmd ./cmd
COPY internal ./internal

# build for the target platform selected by buildx
ARG TARGETOS
ARG TARGETARCH
# version is embedded from internal/version/VERSION (single source of truth)
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags "-s -w" -o /porttrap ./cmd/porttrap

## Final minimal image
FROM scratch
COPY --from=builder /porttrap /porttrap

# scratch has no shell/curl, so the binary probes itself via its -healthcheck flag
HEALTHCHECK --interval=5m --timeout=5s --start-period=5s --retries=3 \
    CMD ["/porttrap", "-healthcheck"]

ENTRYPOINT ["/porttrap"]
