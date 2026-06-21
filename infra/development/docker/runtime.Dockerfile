FROM alpine:3.22

WORKDIR /app

COPY build/runtime /app/runtime

RUN chmod +x /app/runtime

USER 1001

ENTRYPOINT ["/app/runtime"]
