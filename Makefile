build-api-gateway:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build -o build/api-gateway ./services/api-gateway

build-scheduler:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build -o build/scheduler ./services/scheduler

build-dispatcher:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build -o build/dispatcher ./services/dispatcher
