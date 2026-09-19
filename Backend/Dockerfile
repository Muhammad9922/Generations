# Builds the API as a static binary and ships only that binary.
FROM golang:1.27-alpine AS build

WORKDIR /src

# Dependencies first, so this layer is only rebuilt when the manifests change.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO off makes the binary static, which is what lets it run on a bare alpine
# rather than the build image that carries a libc. -trimpath keeps the build
# path out of the binary; -s -w drops the symbol table and DWARF data.
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.21

# ca-certificates is for a TLS bolt:// URI; the rest is a non-root user to run
# as, since the API needs no privileges at all.
RUN apk add --no-cache ca-certificates \
    && adduser -D -H -u 10001 app

COPY --from=build /out/server /usr/local/bin/server

# The graph lives outside this image: set NEO4J_URI to reach yours. Left unset,
# the binary falls back to the development instance baked into main.go.
ENV API_ADDR=":8080"

USER app
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/health || exit 1

ENTRYPOINT ["server"]
