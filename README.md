mystack-controller
=================
[![Build Status](https://travis-ci.org/topfreegames/mystack-controller.svg?branch=master)](https://travis-ci.org/topfreegames/mystack-controller)

Mystack creates a personal stack of predefined services on kubernetes for your users

## Dependencies
* Go 1.7
* Docker

### building docker image
```
make build-docker
```

## Building
#### Build a linux binary
```shell
make cross-build-linux-amd64
```

#### Automated tests
To run unit tests:

```shell
make unit
```
To run integration tests:

```shell
make int
```

To run them all:
```shell
make test
```

## Running
This controller must run inside Kubernetes cluster. So you need to create a docker image, push it to Dockerhub and run a service using this image. 
Here is an example of how to do it.

#### Build a docker image
On project root, run (mind the dot):
```shell
docker build -t dockerhub-user/mystack-controller:v1 .
```

#### Push it to Dockerhub
```shell
docker push dockerhub-user/mystack-controller:v1
```

#### Run postgres on Kubernetes
```shell
kubectl create -f ./manifests/postgres.yaml
```

#### Run controller on Kubernetes
```shell
kubectl create -f ./manifests/controller.yaml
```
