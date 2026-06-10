FROM alpine:3.22

WORKDIR /app

COPY build/dispatcher /app/dispatcher

RUN chmod +x /app/dispatcher

USER 1001

ENTRYPOINT ["/app/dispatcher"]
