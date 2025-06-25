package main

import (
	"fmt"
	"log"

	"github.com/Winter0rbit/yamler"
)

func main() {
	// Kubernetes Deployment YAML
	k8sDeployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.20
        ports:
        - containerPort: 80
        env:
        - name: NGINX_PORT
          value: "80"
        resources:
          requests:
            memory: "64Mi"
            cpu: "250m"
          limits:
            memory: "128Mi"
            cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: nginx-service
spec:
  selector:
    app: nginx
  ports:
  - protocol: TCP
    port: 80
    targetPort: 80
  type: ClusterIP`

	fmt.Println("=== Original Kubernetes Manifest ===")
	fmt.Println(k8sDeployment)

	// Load the Kubernetes manifest
	doc, err := yamler.Load(k8sDeployment)
	if err != nil {
		log.Fatal("Failed to load Kubernetes manifest:", err)
	}

	// Scale the deployment
	fmt.Println("\n=== Scaling Deployment ===")
	doc.Set("spec.replicas", 5)

	// Update image version
	doc.Set("spec.template.spec.containers[0].image", "nginx:1.21")

	// Add resource limits
	doc.Set("spec.template.spec.containers[0].resources.limits.memory", "256Mi")
	doc.Set("spec.template.spec.containers[0].resources.limits.cpu", "1000m")

	// Add environment variables
	fmt.Println("\n=== Adding Configuration ===")

	// Add new environment variable
	newEnvVar := map[string]interface{}{
		"name":  "NGINX_HOST",
		"value": "0.0.0.0",
	}
	doc.AppendToArray("spec.template.spec.containers[0].env", newEnvVar)

	// Add config from ConfigMap
	configMapEnv := map[string]interface{}{
		"name": "NGINX_CONFIG",
		"valueFrom": map[string]interface{}{
			"configMapKeyRef": map[string]interface{}{
				"name": "nginx-config",
				"key":  "nginx.conf",
			},
		},
	}
	doc.AppendToArray("spec.template.spec.containers[0].env", configMapEnv)

	// Add volume mounts
	fmt.Println("\n=== Adding Volumes ===")

	volumeMount := map[string]interface{}{
		"name":      "nginx-config",
		"mountPath": "/etc/nginx/nginx.conf",
		"subPath":   "nginx.conf",
	}
	doc.Set("spec.template.spec.containers[0].volumeMounts", []interface{}{volumeMount})

	volume := map[string]interface{}{
		"name": "nginx-config",
		"configMap": map[string]interface{}{
			"name": "nginx-config",
		},
	}
	doc.Set("spec.template.spec.volumes", []interface{}{volume})

	// Add labels and annotations
	fmt.Println("\n=== Adding Metadata ===")

	doc.Set("metadata.annotations", map[string]interface{}{
		"deployment.kubernetes.io/revision":                "1",
		"kubectl.kubernetes.io/last-applied-configuration": "{}",
	})

	doc.Set("spec.template.metadata.annotations", map[string]interface{}{
		"prometheus.io/scrape": "true",
		"prometheus.io/port":   "9090",
	})

	// Working with the service section (if it exists in a multi-document YAML)
	// Note: This example assumes single document, but Yamler can handle multi-document YAMLs

	// Get current configuration
	fmt.Println("\n=== Current Configuration ===")

	replicas, _ := doc.GetInt("spec.replicas")
	fmt.Printf("Replicas: %d\n", replicas)

	image, _ := doc.GetString("spec.template.spec.containers[0].image")
	fmt.Printf("Image: %s\n", image)

	envVars, _ := doc.GetSlice("spec.template.spec.containers[0].env")
	fmt.Printf("Environment variables: %d\n", len(envVars))

	for i, env := range envVars {
		if envMap, ok := env.(map[string]interface{}); ok {
			if name, exists := envMap["name"]; exists {
				fmt.Printf("  %d. %s\n", i+1, name)
			}
		}
	}

	// Display final manifest
	fmt.Println("\n=== Updated Kubernetes Manifest ===")
	result, err := doc.String()
	if err != nil {
		log.Fatal("Failed to convert to string:", err)
	}
	fmt.Println(result)
}
