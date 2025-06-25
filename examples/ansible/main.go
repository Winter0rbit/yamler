package main

import (
	"fmt"
	"log"

	"github.com/Winter0rbit/yamler"
)

func main() {
	// Ansible Playbook YAML (array root style)
	ansiblePlaybook := `- name: Deploy web application
  hosts: webservers
  become: yes
  vars:
    app_name: myapp
    app_version: 1.0.0
    app_port: 3000
    db_host: localhost
    db_port: 5432
  
  tasks:
  - name: Update package cache
    apt:
      update_cache: yes
      cache_valid_time: 3600
    
  - name: Install required packages
    apt:
      name:
        - nginx
        - nodejs
        - npm
        - postgresql-client
      state: present
    
  - name: Create application directory
    file:
      path: /opt/{{ app_name }}
      state: directory
      owner: www-data
      group: www-data
      mode: '0755'
    
  - name: Copy application files
    copy:
      src: ./app/
      dest: /opt/{{ app_name }}/
      owner: www-data
      group: www-data
      mode: '0644'
    
  - name: Install npm dependencies
    npm:
      path: /opt/{{ app_name }}
      state: present
    
  - name: Configure nginx
    template:
      src: nginx.conf.j2
      dest: /etc/nginx/sites-available/{{ app_name }}
    notify: restart nginx
    
  - name: Enable nginx site
    file:
      src: /etc/nginx/sites-available/{{ app_name }}
      dest: /etc/nginx/sites-enabled/{{ app_name }}
      state: link
    notify: restart nginx
  
  handlers:
  - name: restart nginx
    service:
      name: nginx
      state: restarted`

	fmt.Println("=== Original Ansible Playbook ===")
	fmt.Println(ansiblePlaybook)

	// Load the Ansible playbook (array root document)
	doc, err := yamler.Load(ansiblePlaybook)
	if err != nil {
		log.Fatal("Failed to load Ansible playbook:", err)
	}

	// Update application configuration
	fmt.Println("\n=== Updating Application Configuration ===")

	// Update app version
	doc.Set("[0].vars.app_version", "2.0.0")
	doc.Set("[0].vars.app_port", 8080)

	// Add new variables
	doc.Set("[0].vars.environment", "production")
	doc.Set("[0].vars.log_level", "info")
	doc.Set("[0].vars.ssl_enabled", true)

	// Add new tasks
	fmt.Println("\n=== Adding New Tasks ===")

	// Add SSL certificate task
	sslTask := map[string]interface{}{
		"name": "Install SSL certificate",
		"copy": map[string]interface{}{
			"src":  "./ssl/cert.pem",
			"dest": "/etc/ssl/certs/{{ app_name }}.pem",
			"mode": "0644",
		},
	}
	doc.AppendToArray("[0].tasks", sslTask)

	// Add SSL key task
	sslKeyTask := map[string]interface{}{
		"name": "Install SSL private key",
		"copy": map[string]interface{}{
			"src":  "./ssl/key.pem",
			"dest": "/etc/ssl/private/{{ app_name }}.key",
			"mode": "0600",
		},
	}
	doc.AppendToArray("[0].tasks", sslKeyTask)

	// Add firewall configuration
	firewallTask := map[string]interface{}{
		"name": "Configure firewall",
		"ufw": map[string]interface{}{
			"rule":  "allow",
			"port":  "{{ app_port }}",
			"proto": "tcp",
		},
	}
	doc.AppendToArray("[0].tasks", firewallTask)

	// Add monitoring setup
	fmt.Println("\n=== Adding Monitoring ===")

	// Add monitoring variables
	doc.Set("[0].vars.monitoring_enabled", true)
	doc.Set("[0].vars.metrics_port", 9090)

	// Add monitoring task
	monitoringTask := map[string]interface{}{
		"name": "Setup application monitoring",
		"template": map[string]interface{}{
			"src":  "monitoring.conf.j2",
			"dest": "/opt/{{ app_name }}/monitoring.conf",
		},
		"when": "monitoring_enabled",
	}
	doc.AppendToArray("[0].tasks", monitoringTask)

	// Add new handler for application restart
	appRestartHandler := map[string]interface{}{
		"name": "restart application",
		"systemd": map[string]interface{}{
			"name":  "{{ app_name }}",
			"state": "restarted",
		},
	}
	doc.AppendToArray("[0].handlers", appRestartHandler)

	// Working with existing tasks - get task information
	fmt.Println("\n=== Task Information ===")

	tasks, _ := doc.GetSlice("[0].tasks")
	fmt.Printf("Total tasks: %d\n", len(tasks))

	for i, task := range tasks {
		if taskMap, ok := task.(map[string]interface{}); ok {
			if name, exists := taskMap["name"]; exists {
				fmt.Printf("  %d. %s\n", i+1, name)
			}
		}
	}

	// Get variables
	fmt.Println("\n=== Variables ===")
	vars, _ := doc.GetMap("[0].vars")
	for key, value := range vars {
		fmt.Printf("  %s: %v\n", key, value)
	}

	// Display final playbook
	fmt.Println("\n=== Updated Ansible Playbook ===")
	result, err := doc.String()
	if err != nil {
		log.Fatal("Failed to convert to string:", err)
	}
	fmt.Println(result)
}
