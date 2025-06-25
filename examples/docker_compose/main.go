package main

import (
	"fmt"
	"log"

	"github.com/Winter0rbit/yamler"
)

func main() {
	// Docker Compose YAML example
	dockerCompose := `version: '3.8'

services:
  web:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    environment:
      - NGINX_HOST=localhost
      - NGINX_PORT=80
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./html:/usr/share/nginx/html:ro
    depends_on:
      - api
    networks:
      - frontend

  api:
    build: 
      context: ./api
      dockerfile: Dockerfile
    image: myapp/api:latest
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=production
      - DB_HOST=database
      - DB_PORT=5432
      - DB_NAME=myapp
    depends_on:
      - database
    networks:
      - frontend
      - backend

  database:
    image: postgres:13
    environment:
      - POSTGRES_DB=myapp
      - POSTGRES_USER=admin
      - POSTGRES_PASSWORD=secret123
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql:ro
    networks:
      - backend

volumes:
  postgres_data:

networks:
  frontend:
  backend:`

	fmt.Println("=== Original Docker Compose ===")
	fmt.Println(dockerCompose)

	// Load the Docker Compose file
	doc, err := yamler.Load(dockerCompose)
	if err != nil {
		log.Fatal("Failed to load Docker Compose:", err)
	}

	// Update configuration for development
	fmt.Println("\n=== Configuring for Development ===")

	// Change environment to development
	doc.Set("services.api.environment[0]", "NODE_ENV=development")

	// Add development database credentials
	doc.Set("services.database.environment[2]", "POSTGRES_PASSWORD=dev123")

	// Add volume mapping for hot reload
	doc.AppendToArray("services.api.volumes", "./api:/app:rw")

	// Add debug port for API
	doc.AppendToArray("services.api.ports", "9229:9229")

	// Add Redis service
	redisService := map[string]interface{}{
		"image":    "redis:alpine",
		"ports":    []string{"6379:6379"},
		"networks": []string{"backend"},
	}
	doc.Set("services.redis", redisService)

	// Update API to depend on Redis
	doc.AppendToArray("services.api.depends_on", "redis")

	// Add environment variables for Redis
	doc.AppendToArray("services.api.environment", "REDIS_HOST=redis")
	doc.AppendToArray("services.api.environment", "REDIS_PORT=6379")

	// Scale web service
	fmt.Println("\n=== Adding Load Balancer ===")

	// Add load balancer
	loadBalancer := map[string]interface{}{
		"image":      "nginx:alpine",
		"ports":      []string{"8080:80"},
		"volumes":    []string{"./lb.conf:/etc/nginx/nginx.conf:ro"},
		"depends_on": []string{"web"},
		"networks":   []string{"frontend"},
	}
	doc.Set("services.loadbalancer", loadBalancer)

	// Working with arrays - get service names
	fmt.Println("\n=== Service Information ===")

	services, _ := doc.GetMap("services")
	fmt.Printf("Total services: %d\n", len(services))

	for serviceName := range services {
		fmt.Printf("- %s\n", serviceName)

		// Get ports if they exist
		ports, err := doc.GetSlice(fmt.Sprintf("services.%s.ports", serviceName))
		if err == nil && len(ports) > 0 {
			fmt.Printf("  Ports: %v\n", ports)
		}
	}

	// Display final Docker Compose
	fmt.Println("\n=== Updated Docker Compose ===")
	result, err := doc.String()
	if err != nil {
		log.Fatal("Failed to convert to string:", err)
	}
	fmt.Println(result)
}
