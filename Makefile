.PHONY: all poligon s01e01 s01e02 s01e03 s01e05

# Default target
all:
	@echo "Available commands:"
	@echo "  make poligon	- Run poligon program"
	@echo "  make s01e01	- Run s01e01 program"
	@echo "  make s01e02	- Run s01e02 program"
	@echo "  make s01e03	- Run s01e03 program"
	@echo "  make s01e05	- Run s01e05 program"
	@echo "  make s02e01	- Run s02e01 program"
	@echo "  make s02e02	- Run s02e02 program"

# Run poligon program
poligon:
	@echo "Running poligon program..."
	@go run cmd/poligon/main.go

# Run s01e01 program
s01e01:
	@echo "Running s01e01 program..."
	@go run cmd/s01e01/main.go

# Run s01e02 program
s01e02:
	@echo "Running s01e02 program..."
	@go run cmd/s01e02/main.go

# Run s01e03 program
s01e03:
	@echo "Running s01e03 program..."
	@go run cmd/s01e03/main.go

# Run s01e05 program
s01e05:
	@echo "Running s01e05 program..."
	@go run cmd/s01e05/main.go

# Run s02e01 program
s02e01:
	@echo "Running s02e01 program..."
	@go run cmd/s02e01/main.go

# Run s02e02 program
s02e02:
	@echo "Running s02e02 program..."
	@go run cmd/s02e02/main.go
