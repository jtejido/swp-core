FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY poc ./poc
COPY conformance ./conformance
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o /out/swp-server ./poc/cmd/swp-server && \
    go build -o /out/swp-client ./poc/cmd/swp-client && \
    go build -o /out/vector-runner ./poc/cmd/vector-runner && \
    go build -o /out/spec-vector-runner ./poc/cmd/spec-vector-runner && \
    go build -o /out/gen-vectors ./poc/cmd/gen-vectors && \
    go build -o /out/gen-c1-vectors ./poc/cmd/gen-c1-vectors && \
    go build -o /out/gen-remaining-vectors ./poc/cmd/gen-remaining-vectors && \
    go build -o /out/mcp-json-gateway ./poc/cmd/mcp-json-gateway

FROM alpine:3.20 AS swp-server
WORKDIR /app
COPY --from=build /out/swp-server /usr/local/bin/swp-server
EXPOSE 7777
ENTRYPOINT ["/usr/local/bin/swp-server"]
CMD ["-listen",":7777"]

FROM alpine:3.20 AS swp-client
WORKDIR /app
COPY --from=build /out/swp-client /usr/local/bin/swp-client
ENTRYPOINT ["/usr/local/bin/swp-client"]
CMD ["-addr","swp-server:7777"]

FROM alpine:3.20 AS swp-vectors
WORKDIR /app
COPY --from=build /out/vector-runner /usr/local/bin/vector-runner
COPY conformance/vectors /app/conformance/vectors
ENTRYPOINT ["/usr/local/bin/vector-runner"]
CMD ["-pattern","conformance/vectors/poc_*.json"]

FROM alpine:3.20 AS swp-spec-vectors
WORKDIR /app
COPY --from=build /out/spec-vector-runner /usr/local/bin/spec-vector-runner
COPY conformance/vectors /app/conformance/vectors
COPY docs /app/docs
ENTRYPOINT ["/usr/local/bin/spec-vector-runner"]

FROM alpine:3.20 AS mcp-json-gateway
WORKDIR /app
COPY --from=build /out/mcp-json-gateway /usr/local/bin/mcp-json-gateway
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/mcp-json-gateway"]
CMD ["-listen",":8080","-swp","swp-server:7777"]
