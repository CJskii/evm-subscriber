FROM golang:1.24

WORKDIR /app

# Install CompileDaemon for auto-rebuilding on file changes.
RUN go install github.com/githubnemo/CompileDaemon@latest

# Copy module files and download dependencies.
COPY go.mod .
# If you have a go.sum file, uncomment the next line:
COPY go.sum .
RUN go mod download

# Copy the rest of the source code.
COPY . .

# Expose port.
EXPOSE 8080

# Use CompileDaemon to rebuild and run the app when files change.
CMD ["CompileDaemon", "-build=go build -o subscriber .", "-command=./subscriber"]
