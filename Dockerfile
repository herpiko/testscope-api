FROM golang:1.18-buster AS builder

ENV GO111MODULE=on
RUN mkdir -p /app
WORKDIR /app
COPY . .
RUN rm .env
RUN rm firebase.json
RUN go vet
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api

FROM alpine
RUN apk --no-cache add ca-certificates bash
WORKDIR /app
COPY --from=builder /app/api .
COPY --from=builder /app/scripts ./scripts
COPY --from=builder /app/migrations ./migrations

EXPOSE 8000
CMD ["/app/api"]
