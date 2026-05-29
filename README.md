## What Is Wardu?
 
Wardu is a **Kubernetes-native distributed job scheduler** where workers are just containers.
 
Unlike Inngest, Trigger.dev, or BullMQ which require you to import their SDK and write jobs their way, Wardu has no SDK and no language lock-in. Your worker is any Docker container that speaks a simple HTTP or gRPC contract. A Python ML script, a Rust binary, a Go service, a Node.js function: all first-class citizens on the same platform.
 
You push a job. Wardu routes it to the right worker pod, retries on failure, scales workers based on queue depth, traces every execution end-to-end, and stores the result. That's it.

## Core Design Principles
 
1. **Workers are containers, not SDK functions.** The only contract: accept a JSON payload, return a JSON result or error. HTTP or gRPC. Nothing else required.
2. **Kubernetes-native scaling.** Wardu watches queue depth and uses the Kubernetes API directly to scale worker deployments. No KEDA, no external autoscaler setup.
3. **Observability from day one.** Every job has a trace ID. Every retry is a span. OpenTelemetry + Jaeger are wired in at the core, not bolted on later.
4. **Simple self-hosting.** One Helm chart. Postgres for state. RabbitMQ as the queue backbone. No Redis. No proprietary state store. No 10-service deployment.
5. **Multi-tenant from the start.** Namespaces isolate teams. Each namespace gets its own queue, concurrency limits, and rate limiting.
