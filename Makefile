.PHONY: all poligon s01e01 s01e02

# Default target
all:
	@echo "Available commands:"
	@echo "  make poligon	- Run poligon program"
	@echo "  make s01e01	- Run s01e01 program"
	@echo "  make s01e02	- Run s01e02 program"

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
