package yamler

import (
	"testing"
)

// TestCustomIndentationPreservation tests that custom indentation styles are preserved
func TestCustomIndentationPreservation(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		key            string
		newValue       interface{}
		expectedOutput string
	}{
		{
			name: "preserve_6_space_indentation",
			input: `config:
      database:
            host: localhost
            port: 5432
      app:
            name: test
            debug: true`,
			key:      "config.app.name",
			newValue: "updated",
			expectedOutput: `config:
      database:
            host: localhost
            port: 5432
      app:
            name: updated
            debug: true
`,
		},
		{
			name: "preserve_8_space_indentation",
			input: `service:
        config:
                timeout: 30
                retries: 3
        endpoints:
                api: /api/v1
                health: /health`,
			key:      "service.config.timeout",
			newValue: 60,
			expectedOutput: `service:
        config:
                timeout: 60
                retries: 3
        endpoints:
                api: /api/v1
                health: /health
`,
		},
		{
			name: "preserve_mixed_indentation_levels",
			input: `level1:
  level2:
    level3:
      value: test
    other: data
  simple: key`,
			key:      "level1.level2.level3.value",
			newValue: "updated",
			expectedOutput: `level1:
  level2:
    level3:
      value: updated
    other: data
  simple: key
`,
		},
		{
			name: "preserve_4_space_indentation",
			input: `config:
    database:
        host: localhost
        port: 5432
    app:
        name: test
        debug: true`,
			key:      "config.app.name",
			newValue: "updated",
			expectedOutput: `config:
    database:
        host: localhost
        port: 5432
    app:
        name: updated
        debug: true
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Load(tt.input)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			err = doc.Set(tt.key, tt.newValue)
			if err != nil {
				t.Fatalf("Set() error = %v", err)
			}

			result, err := doc.String()
			if err != nil {
				t.Fatalf("String() error = %v", err)
			}

			if result != tt.expectedOutput {
				t.Errorf("Custom indentation not preserved.\nGot:\n%q\nWant:\n%q", result, tt.expectedOutput)
			}
		})
	}
}

// TestInlineStructurePreservation tests that inline/flow structures are preserved
func TestInlineStructurePreservation(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		key            string
		newValue       interface{}
		expectedOutput string
	}{
		{
			name: "preserve_inline_object",
			input: `config: {host: localhost, port: 5432, ssl: true}
app: test`,
			key:      "app",
			newValue: "updated",
			expectedOutput: `config: {host: localhost, port: 5432, ssl: true}
app: updated
`,
		},
		{
			name: "preserve_nested_inline_objects",
			input: `services:
  db: {host: localhost, port: 5432}
  cache: {host: redis, port: 6379}
  api: {url: /api, version: v1}
config:
  timeout: 30`,
			key:      "config.timeout",
			newValue: 60,
			expectedOutput: `services:
  db: {host: localhost, port: 5432}
  cache: {host: redis, port: 6379}
  api: {url: /api, version: v1}
config:
  timeout: 60
`,
		},
		{
			name: "preserve_mixed_inline_and_block",
			input: `database:
  primary: {host: db1, port: 5432}
  replica: {host: db2, port: 5432}
  settings:
    pool_size: 10
    timeout: 30
app:
  name: myapp`,
			key:      "app.name",
			newValue: "updated",
			expectedOutput: `database:
  primary: {host: db1, port: 5432}
  replica: {host: db2, port: 5432}
  settings:
    pool_size: 10
    timeout: 30
app:
  name: updated
`,
		},
		{
			name: "preserve_inline_arrays_in_objects",
			input: `config: {
  hosts: [web1, web2, web3],
  ports: [80, 443, 8080],
  enabled: true
}
other: value`,
			key:      "other",
			newValue: "updated",
			expectedOutput: `config: {
  hosts: [web1, web2, web3],
  ports: [80, 443, 8080],
  enabled: true
}
other: updated
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Load(tt.input)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			err = doc.Set(tt.key, tt.newValue)
			if err != nil {
				t.Fatalf("Set() error = %v", err)
			}

			result, err := doc.String()
			if err != nil {
				t.Fatalf("String() error = %v", err)
			}

			if result != tt.expectedOutput {
				t.Errorf("Inline structure not preserved.\nGot:\n%s\nWant:\n%s", result, tt.expectedOutput)
			}
		})
	}
}

// TestComplexYamlFormatPreservation tests various complex YAML formats
func TestComplexYamlFormatPreservation(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		key            string
		newValue       interface{}
		expectedOutput string
	}{
		{
			name: "preserve_folded_scalars",
			input: `description: >
  This is a very long description
  that spans multiple lines
  and should be folded
config:
  name: test`,
			key:      "config.name",
			newValue: "updated",
			expectedOutput: `description: >
  This is a very long description
  that spans multiple lines
  and should be folded
config:
  name: updated
`,
		},
		{
			name: "preserve_literal_scalars",
			input: `script: |
  #!/bin/bash
  echo "Hello World"
  exit 0
config:
  env: production`,
			key:      "config.env",
			newValue: "staging",
			expectedOutput: `script: |
  #!/bin/bash
  echo "Hello World"
  exit 0
config:
  env: staging
`,
		},
		{
			name: "preserve_quoted_strings",
			input: `config:
  message: "Hello, World!"
  pattern: 'regex: \d+'
  path: "C:\\Program Files\\App"
debug: false`,
			key:      "debug",
			newValue: true,
			expectedOutput: `config:
  message: "Hello, World!"
  pattern: 'regex: \d+'
  path: "C:\\Program Files\\App"
debug: true
`,
		},
		{
			name: "preserve_multiline_arrays",
			input: `dependencies:
  - name: package1
    version: 1.0.0
  - name: package2
    version: 2.0.0
  - name: package3
    version: 3.0.0
env: production`,
			key:      "env",
			newValue: "staging",
			expectedOutput: `dependencies:
  - name: package1
    version: 1.0.0
  - name: package2
    version: 2.0.0
  - name: package3
    version: 3.0.0
env: staging
`,
		},
		{
			name: "preserve_complex_nested_flow",
			input: `matrix: [
  [1, 2, 3],
  [4, 5, 6],
  [7, 8, 9]
]
metadata: {
  created: 2023-01-01,
  author: user,
  tags: [yaml, test, config]
}
version: 1`,
			key:      "version",
			newValue: 2,
			expectedOutput: `matrix: [
  [1, 2, 3],
  [4, 5, 6],
  [7, 8, 9]
]
metadata: {
  created: 2023-01-01,
  author: user,
  tags: [yaml, test, config]
}
version: 2
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Load(tt.input)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			err = doc.Set(tt.key, tt.newValue)
			if err != nil {
				t.Fatalf("Set() error = %v", err)
			}

			result, err := doc.String()
			if err != nil {
				t.Fatalf("String() error = %v", err)
			}

			if result != tt.expectedOutput {
				t.Errorf("Complex YAML format not preserved.\nGot:\n%s\nWant:\n%s", result, tt.expectedOutput)
			}
		})
	}
}

// TestArrayOperationsWithCustomFormats tests array operations with various formatting styles
func TestArrayOperationsWithCustomFormats(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		operation      func(*Document) error
		expectedOutput string
	}{
		{
			name: "append_to_multiline_flow_array",
			input: `items: [
  item1,
  item2,
  item3
]
other: value`,
			operation: func(d *Document) error {
				return d.AppendToArray("items", "item4")
			},
			expectedOutput: `items: [
  item1,
  item2,
  item3,
  item4
]
other: value
`,
		},
		{
			name: "append_to_compact_flow_array",
			input: `items:[1,2,3]
other: value`,
			operation: func(d *Document) error {
				return d.AppendToArray("items", 4)
			},
			expectedOutput: `items:[1,2,3,4]
other: value
`,
		},
		{
			name: "append_to_custom_indented_array",
			input: `config:
      items:
        - first
        - second
        - third
      other: data`,
			operation: func(d *Document) error {
				return d.AppendToArray("config.items", "fourth")
			},
			expectedOutput: `config:
      items:
        - first
        - second
        - third
        - fourth
      other: data
`,
		},
		{
			name: "update_element_in_spaced_flow_array",
			input: `values: [ 1 , 2 , 3 ]
name: test`,
			operation: func(d *Document) error {
				return d.UpdateArrayElement("values", 1, 99)
			},
			expectedOutput: `values: [ 1 , 99 , 3 ]
name: test
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Load(tt.input)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			err = tt.operation(doc)
			if err != nil {
				t.Fatalf("Operation error = %v", err)
			}

			result, err := doc.String()
			if err != nil {
				t.Fatalf("String() error = %v", err)
			}

			if result != tt.expectedOutput {
				t.Errorf("Custom array format not preserved.\nGot:\n%s\nWant:\n%s", result, tt.expectedOutput)
			}
		})
	}
}

// TestDocumentLevelFormatting tests document-level formatting preservation
func TestDocumentLevelFormatting(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		key            string
		newValue       interface{}
		expectedOutput string
	}{
		{
			name: "preserve_document_separators",
			input: `---
config:
  name: test
  version: 1
...`,
			key:      "config.version",
			newValue: 2,
			expectedOutput: `---
config:
  name: test
  version: 2
...
`,
		},
		{
			name: "preserve_blank_lines",
			input: `# Configuration file

config:
  database:
    host: localhost

    port: 5432


  app:
    name: test`,
			key:      "config.app.name",
			newValue: "updated",
			expectedOutput: `# Configuration file

config:
  database:
    host: localhost

    port: 5432


  app:
    name: updated
`,
		},
		{
			name: "preserve_comment_formatting",
			input: `# Main configuration
#
# This is a multi-line comment
# that explains the configuration
#
config:
  # Database configuration
  database:
    host: localhost # Primary host
    port: 5432      # Standard PostgreSQL port
  
  # Application settings
  app:
    name: myapp     # Application name
    debug: false    # Debug mode`,
			key:      "config.app.debug",
			newValue: true,
			expectedOutput: `# Main configuration
#
# This is a multi-line comment
# that explains the configuration
#
config:
  # Database configuration
  database:
    host: localhost # Primary host
    port: 5432      # Standard PostgreSQL port
  
  # Application settings
  app:
    name: myapp     # Application name
    debug: true    # Debug mode
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Load(tt.input)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			err = doc.Set(tt.key, tt.newValue)
			if err != nil {
				t.Fatalf("Set() error = %v", err)
			}

			result, err := doc.String()
			if err != nil {
				t.Fatalf("String() error = %v", err)
			}

			if result != tt.expectedOutput {
				t.Errorf("Document formatting not preserved.\nGot:\n%s\nWant:\n%s", result, tt.expectedOutput)
			}
		})
	}
}

// TestRealWorldFormats tests commonly used YAML formats in real projects
func TestRealWorldFormats(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		key            string
		newValue       interface{}
		expectedOutput string
	}{
		{
			name: "docker_compose_style",
			input: `version: '3.8'
services:
  web:
    image: nginx:latest
    ports:
      - "80:80"
      - "443:443"
    environment:
      - NODE_ENV=production
      - DEBUG=false
  db:
    image: postgres:13
    environment:
      POSTGRES_DB: myapp
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass`,
			key:      "services.web.environment[1]",
			newValue: "DEBUG=true",
			expectedOutput: `version: '3.8'
services:
  web:
    image: nginx:latest
    ports:
      - "80:80"
      - "443:443"
    environment:
      - NODE_ENV=production
      - DEBUG=true
  db:
    image: postgres:13
    environment:
      POSTGRES_DB: myapp
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
`,
		},
		{
			name: "kubernetes_manifest_style",
			input: `apiVersion: apps/v1
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
        - containerPort: 80`,
			key:      "spec.replicas",
			newValue: 5,
			expectedOutput: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 5
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
`,
		},
		{
			name: "github_actions_style",
			input: `name: CI
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        node-version: [14.x, 16.x, 18.x]
    steps:
    - uses: actions/checkout@v3
    - name: Use Node.js ${{ matrix.node-version }}
      uses: actions/setup-node@v3
      with:
        node-version: ${{ matrix.node-version }}
    - run: npm ci
    - run: npm test`,
			key:      "jobs.test.strategy.matrix.node-version[2]",
			newValue: "20.x",
			expectedOutput: `name: CI
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        node-version: [14.x, 16.x, 20.x]
    steps:
    - uses: actions/checkout@v3
    - name: Use Node.js ${{ matrix.node-version }}
      uses: actions/setup-node@v3
      with:
        node-version: ${{ matrix.node-version }}
    - run: npm ci
    - run: npm test
`,
		},
		{
			name: "ansible_playbook_style",
			input: `---
- name: Configure web servers
  hosts: webservers
  become: yes
  vars:
    http_port: 80
    max_clients: 200
  tasks:
    - name: Install nginx
      package:
        name: nginx
        state: present
    - name: Start nginx service
      service:
        name: nginx
        state: started
        enabled: yes`,
			key:      "vars.max_clients",
			newValue: 500,
			expectedOutput: `---
- name: Configure web servers
  hosts: webservers
  become: yes
  vars:
    http_port: 80
    max_clients: 500
  tasks:
    - name: Install nginx
      package:
        name: nginx
        state: present
    - name: Start nginx service
      service:
        name: nginx
        state: started
        enabled: yes
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Load(tt.input)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			err = doc.Set(tt.key, tt.newValue)
			if err != nil {
				t.Fatalf("Set() error = %v", err)
			}

			result, err := doc.String()
			if err != nil {
				t.Fatalf("String() error = %v", err)
			}

			if result != tt.expectedOutput {
				t.Errorf("Real-world format not preserved.\nGot:\n%s\nWant:\n%s", result, tt.expectedOutput)
			}
		})
	}
}

// TestCompactFormats tests preservation of compact/concise YAML styles
func TestCompactFormats(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		key            string
		newValue       interface{}
		expectedOutput string
	}{
		{
			name: "compact_config",
			input: `db: {host: localhost, port: 5432, ssl: true}
cache: {host: redis, port: 6379}
debug: false`,
			key:      "debug",
			newValue: true,
			expectedOutput: `db: {host: localhost, port: 5432, ssl: true}
cache: {host: redis, port: 6379}
debug: true
`,
		},
		{
			name: "compact_arrays",
			input: `ports: [80, 443, 8080]
hosts: [web1, web2, web3]
env: production`,
			key:      "env",
			newValue: "staging",
			expectedOutput: `ports: [80, 443, 8080]
hosts: [web1, web2, web3]
env: staging
`,
		},
		{
			name: "mixed_compact_and_verbose",
			input: `database:
  connection: {host: localhost, port: 5432}
  pool:
    min: 5
    max: 20
  features: [ssl, compression, logging]
app:
  name: myapp
  version: 1.0.0`,
			key:      "app.version",
			newValue: "1.1.0",
			expectedOutput: `database:
  connection: {host: localhost, port: 5432}
  pool:
    min: 5
    max: 20
  features: [ssl, compression, logging]
app:
  name: myapp
  version: 1.1.0
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Load(tt.input)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			err = doc.Set(tt.key, tt.newValue)
			if err != nil {
				t.Fatalf("Set() error = %v", err)
			}

			result, err := doc.String()
			if err != nil {
				t.Fatalf("String() error = %v", err)
			}

			if result != tt.expectedOutput {
				t.Errorf("Compact format not preserved.\nGot:\n%s\nWant:\n%s", result, tt.expectedOutput)
			}
		})
	}
}

// TestStringStylePreservation tests different string quoting styles
func TestStringStylePreservation(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		key            string
		newValue       interface{}
		expectedOutput string
	}{
		{
			name: "preserve_quoted_strings",
			input: `config:
  message: "Hello, World!"
  pattern: 'regex: \d+'
  path: "C:\\Program Files\\App"
  plain: simple value
debug: false`,
			key:      "debug",
			newValue: true,
			expectedOutput: `config:
  message: "Hello, World!"
  pattern: 'regex: \d+'
  path: "C:\\Program Files\\App"
  plain: simple value
debug: true
`,
		},
		{
			name: "preserve_special_characters",
			input: `strings:
  json: '{"key": "value", "array": [1, 2, 3]}'
  yaml: 'key: value'
  multiword: "string with spaces"
  symbols: "!@#$%^&*()"
updated: false`,
			key:      "updated",
			newValue: true,
			expectedOutput: `strings:
  json: '{"key": "value", "array": [1, 2, 3]}'
  yaml: 'key: value'
  multiword: "string with spaces"
  symbols: "!@#$%^&*()"
updated: true
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Load(tt.input)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			err = doc.Set(tt.key, tt.newValue)
			if err != nil {
				t.Fatalf("Set() error = %v", err)
			}

			result, err := doc.String()
			if err != nil {
				t.Fatalf("String() error = %v", err)
			}

			if result != tt.expectedOutput {
				t.Errorf("String style not preserved.\nGot:\n%s\nWant:\n%s", result, tt.expectedOutput)
			}
		})
	}
}

// TestAdvancedArrayOperations tests array operations with different styles
func TestAdvancedArrayOperations(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		operation      func(*Document) error
		expectedOutput string
	}{
		{
			name: "append_to_inline_array",
			input: `services: [web, api, db]
version: 1`,
			operation: func(d *Document) error {
				return d.AppendToArray("services", "cache")
			},
			expectedOutput: `services: [web, api, db, cache]
version: 1
`,
		},
		{
			name: "modify_nested_array_element",
			input: `config:
  servers:
    - name: web1
      port: 80
    - name: web2
      port: 80
  env: production`,
			operation: func(d *Document) error {
				return d.Set("config.servers[1].port", 8080)
			},
			expectedOutput: `config:
  servers:
    - name: web1
      port: 80
    - name: web2
      port: 8080
  env: production
`,
		},
		{
			name: "remove_from_mixed_array",
			input: `dependencies:
  runtime: [nodejs, npm, redis]
  dev: [eslint, jest, prettier]
env: development`,
			operation: func(d *Document) error {
				return d.RemoveFromArray("dependencies.runtime", 1)
			},
			expectedOutput: `dependencies:
  runtime: [nodejs, redis]
  dev: [eslint, jest, prettier]
env: development
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Load(tt.input)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			err = tt.operation(doc)
			if err != nil {
				t.Fatalf("Operation error = %v", err)
			}

			result, err := doc.String()
			if err != nil {
				t.Fatalf("String() error = %v", err)
			}

			if result != tt.expectedOutput {
				t.Errorf("Advanced array operation failed.\nGot:\n%s\nWant:\n%s", result, tt.expectedOutput)
			}
		})
	}
}
