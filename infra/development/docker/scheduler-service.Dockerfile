FROM alpine:3.22

WORKDIR /app

COPY build/scheduler /app/scheduler

RUN chmod +x /app/scheduler

USER 1001

ENTRYPOINT ["/app/scheduler"]
