# 🕸️ Website Kubernetes Operator

A lightweight **Kubernetes Operator** written in **Go** using **Kubebuilder** that automates the deployment of static websites with **NGINX**.

---

## 🚀 Overview

This operator introduces a custom Kubernetes resource called **`Website`**.  
When you apply a `Website` YAML, the operator automatically:

- Creates a **ConfigMap** containing your HTML.  
- Spins up an **NGINX Deployment** serving that content.  
- Exposes it via a **Service**.  

Everything is managed automatically — if you update or delete the `Website` resource, the operator reconciles the changes.

---

## 🧩 Example

```yaml
apiVersion: apps.google.com/v1alpha1
kind: Website
metadata:
  name: website-sample
  namespace: default
spec:
  replicas: 2
  indexHTML: |
    <html><body><h1>Hello from the Website Operator 👋</h1></body></html>
  serviceType: ClusterIP
```
Apply and access it:
kubectl apply -f config/samples/apps_v1alpha1_website.yaml
kubectl port-forward svc/website-sample-svc 8080:80


Then open: http://localhost:8080 🎉

---

## 🛠️ Tech Stack

- Language: Go (1.21+)
- Framework: Kubebuilder
- Core Libraries: controller-runtime, client-go
- Container: NGINX

---
## 🧠 What It Does (in plain English)
“I built a Kubernetes Operator that watches for a custom resource called Website.
When a user creates one, the operator automatically provisions a Deployment, Service, and ConfigMap to host that site using NGINX.
It’s like teaching Kubernetes how to host a website on its own.”

---

## 🧑‍💻 Commands


#Install CRD
make install

#Run operator locally
make run

#Apply sample
kubectl apply -f config/samples/apps_v1alpha1_website.yaml


---

## 🧠 Key Learnings

- How to create and register Custom Resource Definitions (CRDs).
- How Kubernetes controllers reconcile desired vs actual state.
- How operators automate cluster workflows using controller-runtime.

---

## 💡 Future Ideas

- Add support for LoadBalancer or Ingress for external access.
- Add health/status conditions in the Website CRD.
- Package and deploy the operator as a container inside the cluster.

---

## ✨ Author
👤 Tobi Ajetomobi (Tobi-Ajet)
💼 Cloud & Platform Engineer

---

## 📜 License

Licensed under the Apache 2.0 License.