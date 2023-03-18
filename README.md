# szakdolgozat

export KUBECONFIG=/mnt/d/szakdolgozat/kubeconfig

kubectl config set-context --current --namespace=guestbook

kubectl port-forward svc/frontend 8080:80

make deploy IMG=quay.io/molnar_liviusz/selenium-test-operator:v0.0.1


helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
kubectl create namespace monitoring
kubectl config set-context --current --namespace=monitoring
helm install prometheus prometheus-community/prometheus-operator-crds


# Verisons

minikube kubernetes 1.25
javascript k8s client 0.18.x
