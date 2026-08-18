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
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags "-s -w -X main.version=v1.0.0" -o /porttrap ./cmd/porttrap

## Final minimal image
FROM scratch
COPY --from=builder /porttrap /porttrap

ENTRYPOINT ["/porttrap"]
