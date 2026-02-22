FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY internal ./internal
COPY cmd ./cmd
RUN go build -o /out/app ./cmd/delivery

FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/app /app/app
ENV SERVICE_NAME=delivery
CMD ["/app/app"]
