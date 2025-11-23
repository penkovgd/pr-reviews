FROM golang:1.25 AS build

ENV PATH="$PATH:$(go env GOPATH)/bin"

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

ENV CGO_ENABLED=0
RUN go build -o /api ./cmd/api

FROM alpine:3.20

COPY --from=build /api /api

ENTRYPOINT [ "/api" ]