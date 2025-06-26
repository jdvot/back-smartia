#!/bin/bash

# Install swag if not already installed
go install github.com/swaggo/swag/cmd/swag@latest

# Generate Swagger documentation
swag init -g cmd/server/main.go -o docs

echo "Swagger documentation generated in docs/ directory" 