# 🎯 Variables
BINARY_NAME=xyphos
DOCKER_IMAGE=xyphos:latest
PYTHON_CLIENT_DIR=clients/py
TS_CLIENT_DIR=clients/typescript
DOCS_DIR=docs
API_DOCS_DIR=$(DOCS_DIR)/api
CLIENT_DOCS_DIR=$(DOCS_DIR)/clients
MASTER_KEY_FOLDER=/var/lib/xyphos

# 🔨 Build commands
.PHONY: build
build: build-server build-clients

.PHONY: build-server
build-server:
	@echo "🔧 Building $(BINARY_NAME) server..."
	go build -o bin/$(BINARY_NAME) ./cmd/main.go

.PHONY: build-clients
build-clients: build-py-client build-ts-client

.PHONY: build-py-client
build-py-client:
	@echo "🐍 Building Python client..."
	cd $(PYTHON_CLIENT_DIR) && python3 -m pip install -e .

.PHONY: build-ts-client
build-ts-client:
	@echo "📦 Building TypeScript client..."
	cd $(TS_CLIENT_DIR) && npm install && npm run build

.PHONY: build-docker
build-docker:
	@echo "🐳 Building Docker image..."
	docker build -t $(DOCKER_IMAGE) -f devops/docker/Dockerfile .

# 🧪 Test commands
.PHONY: test
test: test-server test-clients

.PHONY: test-server
test-server:
	@echo "🧪 Running server tests..."
	go test -v ./...

.PHONY: test-clients
test-clients: test-py-client test-ts-client

.PHONY: test-py-client
test-py-client:
	@echo "🐍 Running Python client tests..."
	cd $(PYTHON_CLIENT_DIR) && python3 -m pytest --cov=xyphos_client

.PHONY: test-ts-client
test-ts-client:
	@echo "📦 Running TypeScript client tests..."
	cd $(TS_CLIENT_DIR) && npm test

.PHONY: test-coverage
test-coverage:
	@echo "📊 Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# 🔍 Lint commands
.PHONY: lint
lint: lint-server lint-clients

.PHONY: lint-server
lint-server:
	@echo "🔍 Running Go linters..."
	golangci-lint run

.PHONY: lint-py-client
lint-py-client:
	@echo "🐍 Running Python linters..."
	cd $(PYTHON_CLIENT_DIR) && black . && isort . && ruff check . && mypy xyphos_client

.PHONY: lint-ts-client
lint-ts-client:
	@echo "📦 Running TypeScript linters..."
	cd $(TS_CLIENT_DIR) && npm run lint

.PHONY: fmt
fmt: fmt-server fmt-clients

.PHONY: fmt-server
fmt-server:
	@echo "✨ Formatting Go code..."
	go fmt ./...

.PHONY: fmt-py-client
fmt-py-client:
	@echo "🐍 Formatting Python code..."
	cd $(PYTHON_CLIENT_DIR) && black .

.PHONY: fmt-ts-client
fmt-ts-client:
	@echo "📦 Formatting TypeScript code..."
	cd $(TS_CLIENT_DIR) && npm run format

# create master key folder
.PHONY: create-master-key-folder
create-master-key-folder:
	@echo "🔑 Creating master key folder..."
	mkdir -p $(MASTER_KEY_FOLDER) 
	chmod 700 $(MASTER_KEY_FOLDER)

# 🚀 Run commands
.PHONY: run
run: create-master-key-folder
	@echo "🚀 Running $(BINARY_NAME)..."
	go run ./cmd/main.go

.PHONY: run-docker
run-docker:
	@echo "🐳 Running Docker container..."
	docker run -p 8080:8080 $(DOCKER_IMAGE)

# 🧹 Clean commands
.PHONY: clean
clean: clean-server clean-clients
	@echo "🧹 Cleaning up..."
	rm -f coverage.out

.PHONY: clean-server
clean-server:
	@echo "🧹 Cleaning server build..."
	rm -rf bin/

.PHONY: clean-py-client
clean-py-client:
	@echo "🧹 Cleaning Python client..."
	cd $(PYTHON_CLIENT_DIR) && rm -rf build/ dist/ *.egg-info/ __pycache__/ .pytest_cache/ .mypy_cache/ .coverage htmlcov/

.PHONY: clean-ts-client
clean-ts-client:
	@echo "🧹 Cleaning TypeScript client..."
	cd $(TS_CLIENT_DIR) && rm -rf dist/ node_modules/ coverage/

# 🔄 Development workflow
.PHONY: dev
dev: fmt lint test build

# 📦 Deployment commands
.PHONY: deploy
deploy: build-docker
	@echo "🚀 Deploying to Kubernetes..."
	kubectl apply -f devops/k8s/deployment.yaml

.PHONY: undeploy
undeploy:
	@echo "🔽 Removing from Kubernetes..."
	kubectl delete -f devops/k8s/deployment.yaml

# 🏗️ Setup commands
.PHONY: setup
setup: setup-server setup-clients

.PHONY: setup-server
setup-server:
	@echo "🔧 Setting up Go development environment..."
	go mod download
	go install github.com/golangci/golint/cmd/golint@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: setup-py-client
setup-py-client:
	@echo "🐍 Setting up Python development environment..."
	cd $(PYTHON_CLIENT_DIR) && python3 -m pip install -e ".[dev]"
	cd $(PYTHON_CLIENT_DIR) && python3 -m pip install -r requirements-dev.txt

.PHONY: setup-ts-client
setup-ts-client:
	@echo "📦 Setting up TypeScript development environment..."
	cd $(TS_CLIENT_DIR) && npm install

# 📝 Generate commands
.PHONY: gen-docs
gen-docs: gen-api-docs gen-client-docs

.PHONY: gen-api-docs
gen-api-docs: setup-swagger
	@echo "📚 Generating API documentation..."
	@echo "🔄 Generating Swagger/OpenAPI specs..."
	cd cmd && swag init -g main.go -o ../$(API_DOCS_DIR)/swagger --parseDependency --parseInternal --parseVendor --outputTypes json,yaml
	@echo "📖 Generating API markdown documentation..."
	swag fmt
	@echo "🌐 Generating Swagger UI..."
	@mkdir -p $(API_DOCS_DIR)/swagger-ui
	@if [ ! -d "$(API_DOCS_DIR)/swagger-ui/dist" ]; then \
		echo "Downloading Swagger UI..."; \
		curl -L https://github.com/swagger-api/swagger-ui/archive/refs/tags/v5.11.2.tar.gz | tar xz -C /tmp && \
		cp -r /tmp/swagger-ui-5.11.2/dist/* $(API_DOCS_DIR)/swagger-ui/ && \
		rm -rf /tmp/swagger-ui-5.11.2; \
	fi
	@cp $(API_DOCS_DIR)/swagger/swagger.json $(API_DOCS_DIR)/swagger-ui/
	@sed -i.bak 's|https://petstore.swagger.io/v2/swagger.json|swagger.json|g' $(API_DOCS_DIR)/swagger-ui/swagger-initializer.js
	@rm -f $(API_DOCS_DIR)/swagger-ui/swagger-initializer.js.bak
	@echo "✅ API documentation generated in $(API_DOCS_DIR)"

.PHONY: gen-client-docs
gen-client-docs:
	@echo "📚 Generating client documentation..."
	@mkdir -p $(CLIENT_DOCS_DIR)
	@echo "🐍 Generating Python client documentation..."
	cd $(PYTHON_CLIENT_DIR) && pdoc3 --html --output-dir ../../$(CLIENT_DOCS_DIR)/py xyphos_client --force
	@echo "📦 Generating TypeScript client documentation..."
	cd $(TS_CLIENT_DIR) && npm run docs -- --out ../../$(CLIENT_DOCS_DIR)/typescript

.PHONY: serve-docs
serve-docs: gen-docs
	@echo "🌐 Starting documentation server..."
	@echo "📚 API docs will be available at http://localhost:8088/api/swagger-ui/"
	@echo "📚 Python client docs will be available at http://localhost:8088/clients/python/"
	@echo "📚 TypeScript client docs will be available at http://localhost:8088/clients/typescript/"
	cd $(DOCS_DIR) && python3 -m http.server 8088

.PHONY: setup-swagger
setup-swagger:
	@echo "🔧 Setting up Swagger tools..."
	go install github.com/swaggo/swag/cmd/swag@latest
	go get -u github.com/swaggo/gin-swagger
	go get -u github.com/swaggo/files
	go get -u github.com/swaggo/swag

# Add swagger annotations example target
.PHONY: swagger-example
swagger-example:
	@echo "📝 Example Swagger annotations:"
	@echo ""
	@echo "// @title Xyphos API"
	@echo "// @version 1.0"
	@echo "// @description A secure multi-tenant key management system"
	@echo "// @contact.name API Support"
	@echo "// @contact.email support@xyphos.io"
	@echo "// @license.name MIT"
	@echo "// @BasePath /api"
	@echo ""
	@echo "Example endpoint annotation:"
	@echo ""
	@echo "// @Summary Create a new project"
	@echo "// @Description Create a new Xyphos project"
	@echo "// @Tags projects"
	@echo "// @Accept json"
	@echo "// @Produce json"
	@echo "// @Param project body CreateProjectRequest true \"Project details\""
	@echo "// @Success 200 {object} Project"
	@echo "// @Failure 400 {object} ErrorResponse"
	@echo "// @Failure 401 {object} ErrorResponse"
	@echo "// @Router /projects [post]"

.PHONY: help
help:
	@echo "🔧 Available commands:"
	@echo ""
	@echo "Build commands:"
	@echo "  build              - Build everything"
	@echo "  build-server      - Build the server binary"
	@echo "  build-clients     - Build all clients"
	@echo "  build-py-client   - Build Python client"
	@echo "  build-ts-client   - Build TypeScript client"
	@echo "  build-docker      - Build Docker image"
	@echo ""
	@echo "Test commands:"
	@echo "  test              - Run all tests"
	@echo "  test-server      - Run server tests"
	@echo "  test-clients     - Run client tests"
	@echo "  test-coverage    - Run tests with coverage"
	@echo ""
	@echo "Lint commands:"
	@echo "  lint              - Run all linters"
	@echo "  fmt               - Format all code"
	@echo ""
	@echo "Run commands:"
	@echo "  run               - Run server locally"
	@echo "  run-docker       - Run in Docker"
	@echo ""
	@echo "Clean commands:"
	@echo "  clean             - Clean everything"
	@echo "  clean-server     - Clean server build"
	@echo "  clean-clients    - Clean client builds"
	@echo ""
	@echo "Documentation commands:"
	@echo "  gen-docs         - Generate all documentation"
	@echo "  gen-api-docs     - Generate API documentation with Swagger"
	@echo "  gen-client-docs  - Generate client documentation"
	@echo "  serve-docs       - Serve documentation locally at http://localhost:8088"
	@echo "  swagger-example  - Show example Swagger annotations"
	@echo ""
	@echo "Development workflow:"
	@echo "  dev               - Full development workflow"
	@echo "  setup             - Setup development environment"
	@echo "  setup-swagger    - Install Swagger tools" 