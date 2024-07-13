# PUCP Private Cloud Orchestrator

## Proyecto Ingenieria de Redes Cloud 24-1

## Requisitos

Tener mongodb desplegado en el puerto 27017

Tener rabbitmq desplegado en el puerto 5673

Tener openstack desplegado (horizon, nova, glance, neutron, placement, keystone)

Tener ovs desplegado

### Headnode

#### Api-gateway

Ejecutar los siguientes comandos con docker

`docker build -t api-gateway .`

`docker run --network=host -p 4444:4444 api-gateway`

#### Backend

Instalar en un virtual environment con pip los siguientes modulos de python

- fastapi==0.111.0
- pymongo==4.7.2
- paramiko==3.4.0
- celery==5.4.0
- requests==2.31.0
- flask==3.0.3

En la carpeta backend/ ejecutar

`sudo python3 template-manager`

`sudo python3 slice-manager`

`sudo celery -A slice-manager.celery worker --loglevel=info`

### Local

#### Cli

Instalar go del siguiente link

`https://go.dev/doc/install`

En la carpeta cli/ ejecutar

`go run main.go` o compilar el binario

- Linux
  `go build -o bin/cli-cloud`

- Windows
  `go build -o bin/cli-cloud.exe`
