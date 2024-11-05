FROM golang:1.21.0-alpine as builder
ARG CGO_ENABLED=0
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build /app/cmd/portfolio-api

FROM scratch as release
COPY --from=builder /app/portfolio-api /portfolio-api
ENTRYPOINT ["/portfolio-api"]