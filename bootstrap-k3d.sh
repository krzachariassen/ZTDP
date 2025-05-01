#!/bin/bash

set -e

echo "Installing k3d, kubectl, Helm, and basic ZTDP dependencies..."

# Install dependencies
sudo apt-get update
sudo apt-get install -y curl wget gnupg lsb-release software-properties-common apt-transport-https

# Install k3d
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash

# Install kubectl
# - Download latest version tag
KUBECTL_VERSION=$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)

# - Download the binary
curl -LO "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl"

chmod +x kubectl
sudo mv kubectl /usr/local/bin/

# Install Helm
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Create k3d cluster
echo "Creating k3d cluster 'ztdp'..."
k3d cluster create ztdp --agents 2 --api-port 6550 -p "8080:80@loadbalancer"

# Wait for nodes to be ready
echo "Waiting for cluster to be ready..."
kubectl wait --for=condition=Ready nodes --all --timeout=120s

# Create namespace
kubectl create namespace ztdp || true

# Install Redis
echo "Installing Redis..."
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
helm install redis bitnami/redis --namespace ztdp --set architecture=standalone

# Install Postgres
echo "Installing Postgres..."
helm install postgres bitnami/postgresql --namespace ztdp --set auth.postgresPassword=ztdp123

# Install NATS
echo "Installing NATS..."
helm repo add nats https://nats-io.github.io/k8s/helm/charts/
helm repo update
helm install nats nats/nats --namespace ztdp

echo "ZTDP local environment setup complete."
echo "Cluster:      k3d ztdp"
echo "Namespace:    ztdp"
echo "Services:     Redis, Postgres, NATS"