# README.md

## full-webapp-project

This repository contains a complete example for building, containerizing, and deploying a Go-based web application to Google Kubernetes Engine (GKE) using Helm.

### Features

* **Cloud SQL**: Connects to a PostgreSQL instance via Cloud SQL Proxy
* **Filestore**: Uses a GKE Filestore share for file storage (ReadWriteMany)
* **GCS Bucket**: Reads and writes objects to a Google Cloud Storage bucket
* **Helm Chart**: Fully parameterized for deploying the app on GKE

---

## Prerequisites

1. **Google Cloud SDK**
   Install and authenticate:

   ```bash
   curl https://sdk.cloud.google.com | bash
   gcloud init
   ```

2. **kubectl** (included with `gcloud`)

3. **Helm 3**

   ```bash
   curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
   ```

4. **GKE Cluster**

   * Ensure you have a running GKE cluster and appropriate IAM permissions

5. **Filestore & Cloud SQL**

   * Create a Filestore instance (NFS) and note its network details
   * Create a Cloud SQL (PostgreSQL) instance and enable the Cloud SQL Admin API

6. **Service Account & Secrets**

   * Create a GCP service account with roles: `roles/cloudsql.client`, `roles/storage.objectAdmin`.
   * Download the service-account key JSON for Kubernetes Secrets.

---

## 1. Build & Push Docker Image

```bash
cd demo-app
# Build locally
docker build -t gcr.io/$GCP_PROJECT/webapp:latest .
# Push to Google Container Registry
docker push gcr.io/$GCP_PROJECT/webapp:latest
```

## 2. Configure kubectl for GKE

```bash
gcloud config set project $GCP_PROJECT
gcloud config set compute/zone $GCP_ZONE
gcloud container clusters get-credentials $CLUSTER_NAME
kubectl config current-context
```

## 3. Create Kubernetes Secrets

```bash
# Cloud SQL credentials (username/password)
kubectl create secret generic cloudsql-instance-credentials \
  --from-literal=username=$DB_USER \
  --from-literal=password=$DB_PASSWORD

# GCS bucket/service-account key
kubectl create secret generic gcp-bucket-credentials \
  --from-file=credentials.json=path/to/sa-key.json
```

## 4. Deploy the Helm Chart

```bash
cd webapp-gke-chart
helm install webapp ./ \
  --set image.repository=gcr.io/$GCP_PROJECT/webapp \
  --set cloudsql.instanceConnectionName=$GCP_PROJECT:$GCP_REGION:$CLOUDSQL_NAME \
  --set bucket.name=$GCS_BUCKET_NAME \
  --set filestore.server=$FILESTORE_IP \
  --set filestore.path=/mnt/webfilestore
```

## 5. Verify Deployment

```bash
kubectl get pods,svc,pvc
kubectl logs deployment/webapp
kubectl port-forward svc/webapp 8080:80
# Visit http://localhost:8080 in browser
```

## 6. Cleanup

```bash
helm uninstall webapp
kubectl delete secret cloudsql-instance-credentials gcp-bucket-credentials
```

---

```
full-webapp-project/
├── webapp/
│   ├── Dockerfile
│   ├── go.mod
│   └── main.go
└── webapp-gke-chart/
    ├── Chart.yaml
    ├── values.yaml
    └── templates/
        ├── deployment.yaml
        ├── service.yaml
        ├── cloudsql-proxy.yaml
        ├── pvc-filestore.yaml
        └── bucket-init-job.yaml
```

---


