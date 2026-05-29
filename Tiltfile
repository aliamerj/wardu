load('ext://restart_process', 'docker_build_with_restart')

# Kubernetes Config
k8s_yaml('./infra/development/k8s/app-config.yaml')
k8s_yaml('./infra/development/k8s/api-gateway-deployment.yaml')


# API Gateway Compile

gateway_compile_cmd = 'make build-api-gateway'

local_resource(
  'api-gateway-compile', 
   gateway_compile_cmd,
      deps=[
        './services/api-gateway',
        './shared',
    ],
    labels=['compile'],
  )

# Docker Build

docker_build_with_restart(
  'wardu/api-gateway', 
  '.',
  dockerfile='./infra/development/docker/api-gateway.Dockerfile',
  entrypoint=['/app/api-gateway'],
  only=['./build/api-gateway'],
  live_update=[
      sync('./build/api-gateway', '/app/api-gateway'),
    ],
  )

# Kubernetes Resource

k8s_resource(
    'api-gateway',
    port_forwards=['8081:8081'],
    resource_deps=[
      'api-gateway-compile',
    ],
    labels=['services'],
)
