FROM golang:1.18-buster AS builder

ENV GO111MODULE=on

ENV PROJECT_NAME ${PROJECT_NAME}
ENV APP_NAME ${APP_NAME}
ENV DB_HOST ${DB_HOST}
ENV DB_USER ${DB_USER}
ENV DB_PASS ${DB_PASS}
ENV DB_NAME ${DB_NAME}
ENV FIREBASE_ACCOUNT_KEY_PATH ${FIREBASE_ACCOUNT_KEY_PATH}
ENV FIREBASE_PROJECT_ID ${FIREBASE_PROJECT_ID}
ENV XENDIT_API_KEY ${XENDIT_API_KEY}
ENV XENDIT_API_SECRET ${XENDIT_API_SECRET}
ENV XENDIT_API_PUB_KEY ${XENDIT_API_PUB_KEY}
ENV XENDIT_CALLBACK_TOKEN ${XENDIT_CALLBACK_TOKEN}
ENV SENDINDBLUE_API_KEY=${SENDINDBLUE_API_KEY}

RUN mkdir -p /app
WORKDIR /app
COPY . .
RUN go vet
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api


FROM alpine
RUN apk --no-cache add ca-certificates bash
WORKDIR /app
COPY --from=builder /app/testscope-io-firebase.json ./testscope-io-firebase.json
COPY --from=builder /app/api .
COPY --from=builder /app/scripts ./scripts
COPY --from=builder /app/migrations ./migrations

EXPOSE 8000
CMD ["/app/api"]
