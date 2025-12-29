# syntax=docker/dockerfile:1.20

# Build the operator manager
FROM golang:1.25 AS build

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

# Avoid invoking VCS tools (git) during build metadata stamping.
ENV GOFLAGS=-buildvcs=false

# Copy module files first for better caching
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
	--mount=type=cache,target=/root/.cache/go-build \
	go mod download

# Copy the rest of the source
COPY . ./

RUN --mount=type=cache,target=/go/pkg/mod \
	--mount=type=cache,target=/root/.cache/go-build \
	CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
	go build -trimpath -ldflags "-s -w" -o /out/manager ./cmd/manager

# Runtime image
FROM gcr.io/distroless/static:nonroot

COPY --from=build /out/manager /manager

USER 65532:65532
EXPOSE 8080 8081
ENTRYPOINT ["/manager"]
