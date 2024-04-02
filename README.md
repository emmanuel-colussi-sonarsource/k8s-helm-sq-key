![SonarQube](images/sonar.png)![AWS](https://img.shields.io/badge/AWS-%23FF9900.svg?style=for-the-badge&logo=amazon-aws&logoColor=white)![Amazon EKS](https://img.shields.io/static/v1?style=for-the-badge&message=Amazon+EKS&color=222222&logo=Amazon+ECS&logoColor=FF9900&label=)


# Securing SonarQube Deployments: Encrypting JDBC Passwords in Kubernetes Secrets

SonarQube is a powerful tool for continuous code quality inspection. However, ensuring the security of sensitive information such as JDBC passwords is crucial when deploying SonarQube in production environments. In this guide, we'll walk through the process of securely encrypting JDBC passwords and storing them in Kubernetes secrets to enhance the security of your SonarQube deployment.

## Use Case Scenario:
Our customer, let's call them CompanyX, has a robust infrastructure running on Kubernetes. They utilize PostgreSQL for their database needs and have implemented best practices for securing their deployments.He already has an encryption key that he uses for his passwords stored in a Kubernetes secret. Now, CompanyX aims to integrate SonarQube into their workflow to enhance code quality analysis. However, they prioritize security and want to ensure that sensitive data, such as database passwords, remains protected.

![Flow pods](images/secretkey.png)

## Prerequisites

In this tutorial, we'll leverage an existing AWS EKS server for Kubernetes deployment. Additionally, we'll utilize AWS SDK in Go to provision a basic instance of PostgreSQL. However, it's not mandatory to follow this step if you already have your Kubernetes server deployed. We'll also provide scripts for deploying your own database instance, ensuring flexibility in your setup.

Why use the AWS SDK in go ? Because i already had a PostgreSQL deployment module used in another project, so I utilized it. We could have employed it to deploy the entire database and SonarQube DCE, but I wanted to demonstrate the addition of a Helm template.

If you have the prerequisites such as a Kubernetes server, a database instance, and Helm installed, you can skip directly to the SonarQube deployment section.

Before you get started, you’ll need to have these things:

* AWS account
* [AWS Development Kit (AWS SDK) v2](https://aws.github.io/aws-sdk-go-v2/docs/getting-started/)
* [Go language installed](https://go.dev/)
* [Kubectl installed](https://docs.aws.amazon.com/eks/latest/userguide/install-kubectl.html) is a command line tool that you use to communicate with the Kubernetes API 
server.
* [HELM](https://helm.sh/docs/intro/install/) The package manager for kubernetes
* bash version 4+
* A Git Client
* AWS EKS cluster installed

---

## Steps


### ✅ Clone the repositories

```bash
:> git clone https://github.com/emmanuel-colussi-sonarsource/k8s-helm-sq-key k8s-helm-sq-key
:> cd k8s-helm-sq-key
```

### ✅ Database deployment
go to directory [db](db) (please read the README.md)

### ✅ SonarQube deployment
go to directory [sonarqube](sonarqube) (please read the README.md)

###  ✅ Generated an AES/GCM key
I have a go program which generates the AES/GCM type key with no padding and a random Initialization Vector of 12 bytes.And which encrypts my JDBC user and password.

The key value : R2FJc1h3YkQ0aktnb3NhTExVZHF1VkVrQ1ZMOVVmR201MlJ2WEpKOSt5TT0=

###  ✅ Created a k8s secret for AES/GCM key

creation of a secret k8s **sonarsecretkey** to store the key.

```yaml
apiVersion: v1
data:
  sonar-secret.txt: R2FJc1h3YkQ0aktnb3NhTExVZHF1VkVrQ1ZMOVVmR201MlJ2WEpKOSt5TT0=
kind: Secret
metadata:
  name: sonarsecretkey
  namespace: sonarqube1
type: Opaque

```
Create the secret :
```bash
:> kubectl apply -f sonarsecretkey.yaml

secret/sonarsecretkey created
:>

```
▶️ Check if Key secret is created :
```bash
:> kubectl -n sonarqube1 get secret

NAME             TYPE     DATA   AGE
sonarsecretkey   Opaque   1      87s
:>
```
▶️ Show the secret :
```bash
:> kubectl -n sonarqube1 get secret sonarsecretkey -o yaml

apiVersion: v1
data:
  sonar-secret.txt: R2FJc1h3YkQ0aktnb3NhTExVZHF1VkVrQ1ZMOVVmR201MlJ2WEpKOSt5TT0=
kind: Secret
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"v1","data":{"sonar-secret.txt":"R2FJc1h3YkQ0aktnb3NhTExVZHF1VkVrQ1ZMOVVmR201MlJ2WEpKOSt5TT0="},"kind":"Secret","metadata":{"annotations":{},"name":"sonarsecretkey","namespace":"sonarqube1"},"type":"Opaque"}
  creationTimestamp: "2024-04-02T09:39:36Z"
  name: sonarsecretkey
  namespace: sonarqube1
  resourceVersion: "9600019"
  uid: 6416a1b9-cc20-431c-9ebf-8b918681d24b
type: Opaque
:>
```

###  ✅ Created a k8s secret for JDBC User,Password and URL
The JDBC user and password are encrypted by the key and encoded in bas64 in secret.


```yaml
apiVersion: v1
data:
  SONAR_JDBC_PASSWORD: e2Flcy1nY219OG9OTFN1THRlUHlaRUVKY29Pd0c4NDdwQ3M1aTBJOTJOWmRiTVdmQ01aNU93alhX
  SONAR_JDBC_URL: amRiYzpwb3N0Z3Jlc3FsOi8vcG9zdGdyZXMtc2VydmljZS5kYXRhYmFzZXBnMS5zdmMuY2x1c3Rlci5sb2NhbDo1NDMyL3NvbmFycXViZT9jdXJyZW50U2NoZW1hPXB1YmxpYw==
  SONAR_JDBC_USERNAME: e2Flcy1nY219dVhnWlhhd2dQT1pFaW1oRVpMSjF2K2tHbXA2eXBoZUwzaWhUOFVhSVZ1MEpFMzh1MFE9PQ==
kind: Secret
metadata:
  name: sonarsecret
  namespace: sonarqube1
type: Opaque
```

The base64 decoded content of SONAR_JDBC_PASSWORD is :
```bash
:> echo "e2Flcy1nY219OG9OTFN1THRlUHlaRUVKY29Pd0c4NDdwQ3M1aTBJOTJOWmRiTVdmQ01aNU93alhX"|base64 -d
{aes-gcm}8oNLSuLtePyZEEJcoOwG847pCs5i0I92NZdbMWfCMZ5OwjXW
:>
```

▶️ Create the secret :
```bash
:> kubectl create -f sonarsecret.yaml 

secret/sonarsecret created
:>

▶️  Check if Key secret is created :
```bash
:> kubectl -n sonarqube1 get secret sonarsecret

NAME          TYPE     DATA   AGE
sonarsecret   Opaque   3      80s
:>
```
▶️ Show the secret :
```bash
:> kubectl -n sonarqube1 get secret sonarsecret -o yaml   
apiVersion: v1
data:
  SONAR_JDBC_PASSWORD: e2Flcy1nY219OG9OTFN1THRlUHlaRUVKY29Pd0c4NDdwQ3M1aTBJOTJOWmRiTVdmQ01aNU93alhX
  SONAR_JDBC_URL: amRiYzpwb3N0Z3Jlc3FsOi8vcG9zdGdyZXMtc2VydmljZS5kYXRhYmFzZXBnMS5zdmMuY2x1c3Rlci5sb2NhbDo1NDMyL3NvbmFycXViZT9jdXJyZW50U2NoZW1hPXB1YmxpYw==
  SONAR_JDBC_USERNAME: e2Flcy1nY219dVhnWlhhd2dQT1pFaW1oRVpMSjF2K2tHbXA2eXBoZUwzaWhUOFVhSVZ1MEpFMzh1MFE9PQ==
kind: Secret
metadata:
  creationTimestamp: "2024-04-02T10:15:16Z"
  name: sonarsecret
  namespace: sonarqube1
  resourceVersion: "9609149"
  uid: 5427c421-6b5c-4aa6-9f1d-13c323352e29
type: Opaque
```

###  ✅  Sonarqube deployment

The HELM value file :
```yaml
sonarqube:
 image:
  repository: sonarqube
  tag: sonarqube:10.3.0-enterprise

sonarSecretKey: "sonarsecretkey" 

postgresql:
 enabled: false
 
jdbcOverwrite:
  enable: true 
  jdbcUrl: "jdbc:postgresql://postgres-service.databasepg1.svc.cluster.local:5432/sonarqube?currentSchema=public"
  jdbcUsername: "sonarqube"
  jdbcSecretName: "sonarsecret"
  jdbcSecretPasswordKey: "SONAR_JDBC_PASSWORD"
```

▶️ Install the SonarQube EE Helm Chart with a custom values :

```bash
:> helm upgrade --install -n sonarqube1 sonarqube sonarqube/sonarqube -f values.yml
Release "sonarqube" does not exist. Installing it now.
NAME: sonarqube
LAST DEPLOYED: Tue Apr  2 12:30:41 2024
NAMESPACE: sonarqube1
STATUS: deployed
REVISION: 1
NOTES:
1. Get the application URL by running these commands:
  export POD_NAME=$(kubectl get pods --namespace sonarqube1 -l "app=sonarqube,release=sonarqube" -o jsonpath="{.items[0].metadata.name}")
  echo "Visit http://127.0.0.1:8080 to use your application"
  kubectl port-forward $POD_NAME 8080:9000 -n sonarqube1
WARNING: 
         Please note that the SonarQube image runs with a non-root user (uid=1000) belonging to the root group (guid=0). In this way, the chart can support arbitrary user ids as recommended in OpenShift.
         Please visit https://docs.openshift.com/container-platform/4.14/openshift_images/create-images.html#use-uid_create-images for more information.
```

▶️ Check if SonarQube is deployed and running

```bash
kubectl -n sonarqube1 get all                       
NAME                        READY   STATUS    RESTARTS   AGE
pod/sonarqube-sonarqube-0   1/1     Running   0          92s

NAME                          TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/sonarqube-sonarqube   ClusterIP   10.100.63.111   <none>        9000/TCP   93s

NAME                                   READY   AGE
statefulset.apps/sonarqube-sonarqube   1/1     93s
```