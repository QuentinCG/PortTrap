# Multi-stage build: compile static Go binary, then produce minimal image
FROM golang:1.26.6 AS builder
WORKDIR /src

# cache deps
# go.sum may not exist for projects without external dependencies.
COPY go.mod ./
RUN go mod download

# copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags "-s -w -X main.version=v1.0.0" -o /porttrap ./cmd/porttrap

## Final minimal image
FROM scratch
COPY --from=builder /porttrap /porttrap

ENTRYPOINT ["/porttrap"]
