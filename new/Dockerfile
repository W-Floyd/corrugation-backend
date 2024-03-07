FROM golang:1-alpine

RUN apk add --no-cache bash git openssh

LABEL maintainer="William Floyd <docker@notmy.space>"

WORKDIR /app

COPY backend/ .

RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build go build -tags mysql,postgres -o main .

FROM alpine:latest

WORKDIR /app

COPY --from=0 /app/main /app/main

EXPOSE 8083

# Run the executable
ENTRYPOINT ["/app/main" ]

COPY frontend/ frontend/