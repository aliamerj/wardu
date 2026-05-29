FROM alpine:3.22

WORKDIR /app

COPY build/api-gateway /app/api-gateway

RUN chmod +x /app/api-gateway

USER 1001

ENTRYPOINT ["/app/api-gateway"]
