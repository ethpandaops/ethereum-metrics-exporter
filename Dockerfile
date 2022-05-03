# syntax=docker/dockerfile:1
FROM golang:1.18 AS builder
WORKDIR /src
COPY go.sum go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/app .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
RUN apk add jq curl
COPY --from=builder /bin/app /bin/app
ENTRYPOINT ["/bin/app"]  