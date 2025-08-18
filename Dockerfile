# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: © 2025 Cory Parent <goedelsoup+waveform@goedelsoup.io>

# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o waveform ./cmd/waveform

# Final stage
FROM scratch

# Import from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /app/waveform /waveform

# Copy documentation and examples
COPY --from=builder /app/README.md /README.md
COPY --from=builder /app/LICENSE /LICENSE
COPY --from=builder /app/examples /examples

EXPOSE 8080

ENTRYPOINT ["/waveform"]
