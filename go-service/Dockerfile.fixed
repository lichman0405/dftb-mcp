# Use Go 1.21 as base image for building
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server/main.go

# ===== DFTB+ Installation Stage =====
FROM continuumio/miniconda3:latest AS dftb-builder

# Install DFTB+ using conda
RUN conda update -n base -c defaults conda && \
    conda install -c conda-forge --yes dftbplus=24.1=nompi_h5d91ca9_100 && \
    conda clean -afy

# ===== Final Image Stage =====
FROM alpine:latest

# Install ca-certificates for HTTPS, timezone data, and required libraries
RUN apk --no-cache add ca-certificates tzdata bash wget

# Create non-root user
RUN addgroup -g 1001 -S appuser && \
    adduser -S appuser -u 1001

# Create working directory
WORKDIR /app

# Copy binary from Go builder
COPY --from=builder /app/main .

# Copy DFTB+ from conda environment
COPY --from=dftb-builder /opt/conda/bin/dftb+ /usr/local/bin/dftb+
COPY --from=dftb-builder /opt/conda/lib /opt/conda/lib

# Set library path for DFTB+
ENV LD_LIBRARY_PATH=/opt/conda/lib:$LD_LIBRARY_PATH

# Create work directory for calculations
RUN mkdir -p /app/work && \
    chown -R appuser:appuser /app

# Make DFTB+ executable
RUN chmod +x /usr/local/bin/dftb+

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Install wget for health check (as root)
USER root
RUN apk add --no-cache wget
USER appuser

# Verify DFTB+ installation
RUN /usr/local/bin/dftb+ --version

# Start the application
CMD ["./main"]
