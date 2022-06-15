# k8a-admission-controller

Admission Controller playground

## Prerequisite
* Docker installed
* Running cluster (preferred a minikube cluster)
  
## Build & Run

Build admission controller executable and image
```
./build.sh
```

The admission image is not pushed to a registry, if you wish to use an external registry you should add `docker push` to the `build.sh` file and update the `imagePullPolicy` to `Always`

> If you are running a local minikube, add `eval $(minikube -p <minikube profile> docker-env)` to your `~/.zprofile`/`~/.bashrc` file

### Run admission controller

1. Startup
    ```
    ./register.sh
    ```

2. Teardown
    ```
    ./unregister
    ```