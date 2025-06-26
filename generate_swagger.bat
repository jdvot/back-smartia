@echo off

REM Install swag if not already installed
go install github.com/swaggo/swag/cmd/swag@latest

REM Generate Swagger documentation
swag init -g cmd/server/main.go -o docs

echo Swagger documentation generated in docs/ directory
pause 