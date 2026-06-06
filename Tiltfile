load('ext://restart_process', 'docker_build_with_restart')

k8s_yaml('./infra/development/k8s/app-config.yaml')

#############
### postgres
#############
k8s_yaml("./infra/development/k8s/postgres-deployment.yaml")
k8s_resource(
  'postgres', 
  port_forwards=['5432:5432'],
  labels=['db'],
  )

#############
### scheduler
#############
k8s_yaml("./infra/development/k8s/scheduler-deployment.yaml")
scheduler_compile_cmd = 'make build-scheduler'
local_resource(
  'scheduler_compile', 
  scheduler_compile_cmd,
  deps=[
  "./services/scheduler",
  "./shared",
  ],
  labels=["compile"]
  )
# build Docker
docker_build_with_restart(
'wardu/scheduler-service',
'.',
dockerfile='./infra/development/docker/scheduler-service.Dockerfile',
  entrypoint=['/app/scheduler'],
  only=['./build/scheduler'],
  live_update=[
      sync('./build/scheduler', '/app/scheduler'),
    ],
)

# Kubernetes Resource

k8s_resource(
    'scheduler-service',
    port_forwards=['8081:8081'],
    resource_deps=[
      'scheduler_compile',
       'postgres',
    ],
    labels=['services'],
)

########################
### API Gateway Compile
########################
k8s_yaml('./infra/development/k8s/api-gateway-deployment.yaml')
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
    port_forwards=['8080:8080'],
    resource_deps=[
      'api-gateway-compile',
       'postgres',
    ],
    labels=['services'],
)
