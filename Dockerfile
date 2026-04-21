# Stage 0: Build frontend
FROM node:22-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# Stage 1: Build Go binary
FROM golang:1.25-alpine AS backend
WORKDIR /app
RUN apk add --no-cache bash git openssh libwebp gcc musl-dev
COPY go.mod go.sum ./
RUN go mod download
COPY main.go main.go
COPY cmd/ cmd/
RUN go build -ldflags="-extldflags -static" -o main .

# Stage 2: Final image
FROM scratch
WORKDIR /
ENV CORRUGATION_AUTHENTICATION=false
ENV CORRUGATION_DATA=/data
COPY --from=backend /app/main /app/main
COPY --from=frontend /app/dist /dist
EXPOSE 8083
ENTRYPOINT ["/app/main"]
