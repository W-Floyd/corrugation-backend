# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from golang:1.12-alpine base image
FROM golang:alpine

# The latest alpine images don't have some tools like (`git` and `bash`).
# Adding git, bash and openssh to the image
RUN apk update && apk upgrade && apk add --no-cache bash git openssh

# Add Maintainer Info
LABEL maintainer="William Floyd <william.png2000@gmail.com>"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependancies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

RUN apk add --no-cache libwebp
RUN apk add --no-cache gcc
RUN apk add --no-cache musl-dev

# Copy the source from the current directory to the Working Directory inside the container
COPY main.go main.go
COPY cmd/ cmd/

# Build the Go app
RUN go build -ldflags="-extldflags -static" -o main .

FROM scratch

WORKDIR /

ENV CORRUGATION_AUTHENTICATION=false
ENV CORRUGATION_DATA=/data
ENV CORRUGATION_ASSETS=/assets

COPY --from=0 /app/main /app/main

# Expose port 8083 to the outside world
EXPOSE 8083

# Run the executable
ENTRYPOINT ["/app/main" ]

COPY assets/ /assets
