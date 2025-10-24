FROM golang:1.25 AS builder

WORKDIR /app

RUN go install github.com/go-task/task/v3/cmd/task@latest

COPY ./server/go.mod ./server/go.sum ./server/
RUN cd server && go mod download

COPY taskfile.yaml .

COPY . .

# Ensure taskfile.yaml uses CGO_ENABLED=0 GOOS=linux GOARCH=amd64 in its build command
# The build command below assumes taskfile.yaml is configured for static linking
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 task build
# --- Add a check to see if the build actually produced the file ---
RUN ls -l /app/server/build/

# --- Stage 2: Final image using Alpine ---
FROM alpine:latest

# Install ca-certificates (for HTTPS calls from Go) and curl (for healthcheck)
RUN apk add --no-cache ca-certificates curl

WORKDIR /app

COPY --from=builder /app/server/build/treblle .

RUN chmod +x ./treblle

# Expose the port the Go app listens on
EXPOSE 8090

# Set the entrypoint for the container
ENTRYPOINT ["./treblle"]


