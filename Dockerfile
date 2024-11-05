FROM golang:1.21.0-alpine as builder
ARG CGO_ENABLED=0
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build /app/cmd/go-api-template

FROM scratch as release
COPY --from=builder /app/go-api-template /go-api-template
ENTRYPOINT ["/go-api-template"]