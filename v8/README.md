Build it
----------

```
go get -insecure gitlab.internal.upstreamsystems.com/devops/mandelbrot
cd $GOPATH/src/gitlab.internal.upstreamsystems.com/devops/mandelbrot
git checkout release-7.0
make tools
```

Install a local kubernetes (minikube)
-------------------------------------

Just use the following [link]: https://upstreamsystems.atlassian.net/wiki/spaces/DEVOPS/pages/142671878/MiniKube].
BUT use the following *start* command:

```
minikube start --memory=4096 --cpus=2 --insecure-registry=localhost:5000
```

Now make sure you have docker installed on your linux desktop, so we can point it to minikube docker with
the following command:

```
eval $(minikube docker-env)
```

Build and deploy our app
-------------------------
Just do the following

```
make            # Build the binaries
make docker     # Build the docker images
```

Let's start the frontend
```
export VERSION=`git describe --tags`
kubectl create -f frontend/k8s/deployment.yml
kubectl set image deployment mandelbrot-frontend frontend=mandelbrot-frontend:${VERSION}
```

But where the fuck is my service? Which IP it uses. Can I ping it?. We need to export it some how.
```
kubectl create -f frontend/k8s/service.yml
```

Now if you check the service status is not doing well
```
curl -s http://`minikube ip`:32400/status | jq .
```

We need to start redis and backend also:

```
kubectl create -f frontend/k8s/redis-service.yml
kubectl create -f frontend/k8s/redis-deployment.yml
kubectl create -f backend/k8s/
kubectl set image deployment mandelbrot-backend backend=mandelbrot-backend:${VERSION}
```

Verify the installation

```
minikube service mandelbrot-service
```

Scale the backend installation

```
kubectl scale deployment --replicas=5 mandelbrot-backend
```
