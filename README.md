# admission-webhook-pod

This repo is used to create a Kubernetes [MutatingAdmissionWebhook](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#mutatingadmissionwebhook) that inject a busybox sidecar container into pod prior to persistence of the object.

This project comes from [kube-sidecar-injector](https://github.com/morvencao/kube-sidecar-injector) & [webhookExample](https://github.com/yuenandi/webhookExample), and combining the advantages of both, thanks a lot .

## Prerequisites

- git
- go version v1.20+
- docker version 24.0.4+
- kubectl version v1.20.15
- Access to a kubernetes v1.20+ cluster with the `admissionregistration.k8s.io/v1` API enabled. Verify that by the following command:
```bash
kubectl api-versions | grep admissionregistration.k8s.io
```
The result should be:
```
admissionregistration.k8s.io/v1
admissionregistration.k8s.io/v1beta1
```

> Note: In addition, the `MutatingAdmissionWebhook` and `ValidatingAdmissionWebhook` admission controllers should be added and listed in the correct order in the admission-control flag of kube-apiserver.

## Build and Deploy

1. Build and push docker image
```bash
make docker-build docker-push IMAGE=<your_dockerhub_username>/admission-webhook-pod:latest
```

2. Deploy the admission-webhook-pod to kubernetes cluster:

```bash
make deploy IMAGE=<your_dockerhub_username>/admission-webhook-pod:latest
```

3. Verify the admission-webhook-pod is up and running:

```bash
# kubectl get pods  -n admission-webhook-pod 
NAME                                     READY   STATUS    RESTARTS   AGE
admission-webhook-pod-574485ffdb-67d46   1/1     Running   0          26s
```

## Debug mode

1. Build birnary
```bash
make build
```

2. Run in debug mode (The server must be able to connect to the k8s cluster)

```bash
./admission-webhook-pod --isDebug=true --port=1443 --url=<your_develop_server_ip> 
```

> Note: You can use debug mode to develop and test various functions of webhook, instead of replacing the tedious packaging into k8s cluster each time, so that you can quickly develop products.

## How to use
1. Create a new namespace `webhook-test` and label it with `admission-webhook-pod=enabled`:
```bash
# kubectl create ns webhook-test
# kubectl label ns webhook-test admission-webhook-pod=enabled
# kubectl get ns webhook-test --show-labels 
NAME           STATUS   AGE   LABELS
webhook-test   Active   42s   admission-webhook-pod=enabled
```

2. Deploy an pod in Kubernetes cluster

```bash
# cd debug
# kubectl apply -f . -n webhook-test
pod/hello-pod created
deployment.apps/sleep created
```

3. Verify sidecar container is injected:

```bash
# kubectl get pods  -n webhook-test 
NAME                    READY   STATUS    RESTARTS   AGE
hello-pod               2/2     Running   0          62s
sleep-5d8896486-zp4bh   2/2     Running   0          62s
# kubectl get pod sleep-5d8896486-zp4bh hello-pod -o jsonpath="{.spec.containers[*].name}"
sleep side-car
# kubectl get pod hello-pod  -o jsonpath="{.spec.containers[*].name}"
ubuntu side-car
```

> Note: Not using sidecar---add the following label to the application:

```
labels:
  admission-webhook-pod.zeratullich.com/app: "false"
```


# Troubleshooting
Sometimes you may find that pod is injected with sidecar container as expected, check the following items:

1. The sidecar pod is in running state and no error logs.
2. The namespace in which application pod is deployed has the correct labels(`admission-webhook-pod=enabled`) as configured in `mutatingwebhookconfiguration`.
3. Check if the application pod has annotation `admission-webhook-pod.zeratullich.com/status=true` or check if the application pod has label `admission-webhook-pod.zeratullich.com/app=true`.