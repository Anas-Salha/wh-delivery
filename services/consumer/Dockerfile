FROM golang:1.24-alpine AS build
WORKDIR /src
COPY main.go ./main.go
RUN go build -o /out/app ./main.go

FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/app /app/app
ENV SERVICE_NAME=consumer
CMD ["/app/app"]
