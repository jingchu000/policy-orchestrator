# Deploy to Azure

Log in to azure CLI.

```bash
az login
```

Create a `.env_azure.sh` file to store your azure environment variables.

```bash
export APP_NAME=<app name>
export AZ_RESOURCE_GROUP=<resource group>
export AZ_AKS_CLUSTER_NAME=<cluster name>
export AZ_ACR_NAME=<name>
```

## Build and push images

Create container registry.

```bash
  az acr create --name ${AZ_ACR_NAME} \
--resource-group ${AZ_RESOURCE_GROUP} \
--sku standard \
--admin-enabled true
```

Log in to azure registry.

```bash
az acr login --name ${AZ_ACR_NAME}
```

Tag and push demo app image.

```bash
docker tag ${APP_NAME} ${AZ_ACR_NAME}.azurecr.io/${APP_NAME}:tag1
docker push ${AZ_ACR_NAME}.azurecr.io/${APP_NAME}:tag1
```

Build and push OPA Server.

## Deploy to App Services


Create App Service Plan.

```bash
az appservice plan create --name ${APP_NAME}plan \
--resource-group ${AZ_RESOURCE_GROUP} \
--is-linux
```

Deploy Hexa Demo App.

```bash
az webapp create --name ${APP_NAME}-demo \
--resource-group ${AZ_RESOURCE_GROUP} \
--plan ${APP_NAME}plan \
--startup-file="demo" \
--deployment-container-image-name ${AZ_ACR_NAME}.azurecr.io/${APP_NAME}:tag1

az webapp config appsettings set --name ${APP_NAME}-demo \
--resource-group ${AZ_RESOURCE_GROUP} \
--settings PORT=8881

az webapp config container set --name ${APP_NAME}-demo \
--resource-group ${AZ_RESOURCE_GROUP} \
--docker-custom-image-name ${AZ_ACR_NAME}.azurecr.io/${APP_NAME}:tag1 \
--docker-registry-server-url "https://${AZ_ACR_NAME}.azurecr.io"

az webapp restart --name ${APP_NAME}-demo \
--resource-group ${AZ_RESOURCE_GROUP}

az webapp show --name ${APP_NAME}-demo \
--resource-group ${AZ_RESOURCE_GROUP} \
| jq -r '.defaultHostName'
```

## Deploy to Kubernetes - AKS

Create cluster.

```bash
az aks create \
    --resource-group ${AZ_RESOURCE_GROUP} \
    --name ${AZ_AKS_CLUSTER_NAME} \
    --node-count 2 \
    --generate-ssh-keys \
    --attach-acr ${AZ_ACR_NAME}
```

View cluster.

```bash
az aks list --resource-group ${AZ_RESOURCE_GROUP}
```

Connect to cluster.

```bash
az aks get-credentials --resource-group ${AZ_RESOURCE_GROUP} --name ${AZ_AKS_CLUSTER_NAME}
```

Create IP Address for Demo app.

```bash
az network public-ip create -g ${AZ_RESOURCE_GROUP} -n ${APP_NAME}-static-ip --allocation-method static
```

```bash
az network public-ip show -g ${AZ_RESOURCE_GROUP} -n ${APP_NAME}-static-ip
```

Deploy demo app objects.

```bash
envsubst < kubernetes/demo/deployment.yaml | kubectl apply -f -
envsubst < kubernetes/demo/service.yaml | kubectl apply -f -
```

Deploy OPA Agent objects.

```bash
envsubst < kubernetes/opa-server/deployment.yaml | kubectl apply -f -
envsubst < kubernetes/opa-server/service.yaml | kubectl apply -f -
```