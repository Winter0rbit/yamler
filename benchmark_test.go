package yamler

import (
	"fmt"
	"strings"
	"testing"
)

// generateLargeYAML creates a large YAML document for testing
func generateLargeYAML(size int) string {
	var builder strings.Builder
	builder.WriteString("app:\n")
	builder.WriteString("  name: large-app\n")
	builder.WriteString("  version: 1.0.0\n")
	builder.WriteString("  servers:\n")

	for i := 0; i < size; i++ {
		builder.WriteString(fmt.Sprintf("    - name: server%d\n", i))
		builder.WriteString(fmt.Sprintf("      host: host%d.example.com\n", i))
		builder.WriteString(fmt.Sprintf("      port: %d\n", 8000+i))
		builder.WriteString("      config:\n")
		builder.WriteString(fmt.Sprintf("        timeout: %d\n", 30+i%60))
		builder.WriteString(fmt.Sprintf("        retries: %d\n", 3+i%5))
		builder.WriteString("        features:\n")
		builder.WriteString("          - logging\n")
		builder.WriteString("          - monitoring\n")
		builder.WriteString("          - metrics\n")
	}

	builder.WriteString("database:\n")
	builder.WriteString("  connections:\n")
	for i := 0; i < size/10; i++ {
		builder.WriteString(fmt.Sprintf("    - name: db%d\n", i))
		builder.WriteString(fmt.Sprintf("      host: db%d.example.com\n", i))
		builder.WriteString(fmt.Sprintf("      port: %d\n", 5432+i))
		builder.WriteString("      pool:\n")
		builder.WriteString(fmt.Sprintf("        min: %d\n", 5+i%10))
		builder.WriteString(fmt.Sprintf("        max: %d\n", 20+i%30))
	}

	return builder.String()
}

// BenchmarkLoad tests document loading performance
func BenchmarkLoad(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		yamlContent := generateLargeYAML(size)
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := Load(yamlContent)
				if err != nil {
					b.Fatal(err)
				}
			}
			b.SetBytes(int64(len(yamlContent)))
		})
	}
}

// BenchmarkToBytes tests serialization performance
func BenchmarkToBytes(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		yamlContent := generateLargeYAML(size)
		doc, err := Load(yamlContent)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result, err := doc.ToBytes()
				if err != nil {
					b.Fatal(err)
				}
				b.SetBytes(int64(len(result)))
			}
		})
	}
}

// BenchmarkSet tests setter performance
func BenchmarkSet(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		yamlContent := generateLargeYAML(size)

		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				doc, err := Load(yamlContent)
				if err != nil {
					b.Fatal(err)
				}

				err = doc.Set("app.version", fmt.Sprintf("1.0.%d", i))
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkGet tests getter performance
func BenchmarkGet(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		yamlContent := generateLargeYAML(size)
		doc, err := Load(yamlContent)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := doc.Get("app.name")
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkArrayOperations tests array manipulation performance
func BenchmarkArrayOperations(b *testing.B) {
	yamlContent := `
app:
  servers: []
  features: [logging, monitoring]
`

	b.Run("AppendToArray", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			doc, err := Load(yamlContent)
			if err != nil {
				b.Fatal(err)
			}

			err = doc.AppendToArray("app.servers", fmt.Sprintf("server%d", i))
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("UpdateArrayElement", func(b *testing.B) {
		doc, err := Load(yamlContent)
		if err != nil {
			b.Fatal(err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err = doc.UpdateArrayElement("app.features", 0, fmt.Sprintf("feature%d", i))
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkWildcardOperations tests wildcard pattern performance
func BenchmarkWildcardOperations(b *testing.B) {
	yamlContent := `
config:
  development:
    debug: true
    timeout: 30
  production:
    debug: false
    timeout: 60
  staging:
    debug: true
    timeout: 45
`

	doc, err := Load(yamlContent)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("GetAll", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := doc.GetAll("config.*.debug")
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("SetAll", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := doc.SetAll("config.*.timeout", 120)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkFormattingDetection tests formatting analysis performance
func BenchmarkFormattingDetection(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		yamlContent := generateLargeYAML(size)

		b.Run(fmt.Sprintf("detectFormattingInfo_size_%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = detectFormattingInfoOptimized(yamlContent)
			}
		})
	}
}

// BenchmarkStringOperations tests string parsing performance
func BenchmarkStringOperations(b *testing.B) {
	paths := []string{
		"app.name",
		"app.servers[0].host",
		"database.connections[5].pool.max",
		"very.deep.nested.path.with.many.levels.value",
	}

	for _, path := range paths {
		b.Run(fmt.Sprintf("parsePath_%s", strings.ReplaceAll(path, ".", "_")), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				parts := strings.Split(path, ".")
				_ = parts // Use the result to prevent optimization
			}
		})
	}
}

// BenchmarkMemoryAllocation tests memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	yamlContent := generateLargeYAML(100)

	b.Run("LoadAndModify", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			doc, err := Load(yamlContent)
			if err != nil {
				b.Fatal(err)
			}

			err = doc.Set("app.version", fmt.Sprintf("1.0.%d", i))
			if err != nil {
				b.Fatal(err)
			}

			_, err = doc.ToBytes()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkComparison compares yamler vs standard go-yaml
func BenchmarkComparison(b *testing.B) {
	yamlContent := generateLargeYAML(100)

	b.Run("Yamler_LoadAndSet", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			doc, err := Load(yamlContent)
			if err != nil {
				b.Fatal(err)
			}

			err = doc.Set("app.version", "2.0.0")
			if err != nil {
				b.Fatal(err)
			}

			_, err = doc.ToBytes()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// Note: This would require importing gopkg.in/yaml.v3 for comparison
	// b.Run("StandardYAML_LoadAndSet", func(b *testing.B) {
	//     b.ResetTimer()
	//     for i := 0; i < b.N; i++ {
	//         var data map[string]interface{}
	//         err := yaml.Unmarshal([]byte(yamlContent), &data)
	//         if err != nil {
	//             b.Fatal(err)
	//         }
	//
	//         // Modify data (complex path traversal needed)
	//         if app, ok := data["app"].(map[string]interface{}); ok {
	//             app["version"] = "2.0.0"
	//         }
	//
	//         _, err = yaml.Marshal(data)
	//         if err != nil {
	//             b.Fatal(err)
	//         }
	//     }
	// })
}

// BenchmarkRealWorldScenarios tests realistic usage patterns
func BenchmarkRealWorldScenarios(b *testing.B) {
	dockerCompose := `
version: '3.8'
services:
  web:
    image: nginx:1.21
    ports:
      - "80:80"
      - "443:443"
    environment:
      - ENV=production
      - DEBUG=false
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
  db:
    image: postgres:13
    environment:
      - POSTGRES_DB=myapp
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=password
    volumes:
      - db_data:/var/lib/postgresql/data
volumes:
  db_data:
`

	b.Run("DockerCompose_UpdateImage", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			doc, err := Load(dockerCompose)
			if err != nil {
				b.Fatal(err)
			}

			err = doc.Set("services.web.image", "nginx:1.22")
			if err != nil {
				b.Fatal(err)
			}

			_, err = doc.ToBytes()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	kubernetesManifest := `
apiVersion: apps/v1
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
        image: nginx:1.21
        ports:
        - containerPort: 80
        resources:
          requests:
            memory: "64Mi"
            cpu: "250m"
          limits:
            memory: "128Mi"
            cpu: "500m"
`

	b.Run("Kubernetes_UpdateReplicas", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			doc, err := Load(kubernetesManifest)
			if err != nil {
				b.Fatal(err)
			}

			err = doc.Set("spec.replicas", 5)
			if err != nil {
				b.Fatal(err)
			}

			_, err = doc.ToBytes()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkPathParsing tests path parsing performance
func BenchmarkPathParsing(b *testing.B) {
	paths := []string{
		"app.name",
		"app.servers[0].host",
		"database.connections[5].pool.max",
		"very.deep.nested.path.with.many.levels.value",
		"config.development.debug",
		"services.web.image",
		"spec.template.spec.containers[0].env",
	}

	b.Run("CachedParsing", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			path := paths[i%len(paths)]
			_ = parsePath(path)
		}
	})

	b.Run("DirectParsing", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			path := paths[i%len(paths)]
			_ = strings.Split(path, ".")
		}
	})
}

// BenchmarkRepeatedOperations tests performance with repeated operations on same document
func BenchmarkRepeatedOperations(b *testing.B) {
	yamlContent := generateLargeYAML(100)
	doc, err := Load(yamlContent)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("RepeatedSet", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := doc.Set("app.version", fmt.Sprintf("1.0.%d", i))
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("RepeatedGet", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := doc.Get("app.name")
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("RepeatedToBytes", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := doc.ToBytes()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
