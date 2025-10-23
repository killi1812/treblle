# TODO: add building for frontend

# Stage 2: Build the backend
FROM golang:1.25 AS builder
WORKDIR /app

# Install Task runner
RUN go install github.com/go-task/task/v3/cmd/task@latest

# Copy Go module files and download dependencies first for caching
COPY ./server/go.mod ./server/go.sum ./
RUN go mod download

# Copy the taskfile so we can use it
COPY taskfile.yaml .

# Copy the entire server source code
COPY . .

# Copy the built frontend assets from the 'frontend' stage
# COPY --from=frontend /app/client/dist ./server/client/dist

RUN task build

FROM gcr.io/distroless/static-debian11
WORKDIR /

# Copy only the compiled binary from the builder stage
COPY --from=builder ./app/server/build/template /treblle

# Set the entrypoint for the container
ENTRYPOINT ["/treblle"]

