.PHONY: all clean

# The name of the Go executable
BINARY_NAME := mywebservice

all: clean build run

build:
	@echo "Compiling..."
	@go build -buildvcs=false -o $(BINARY_NAME) .

run:
	@echo "Running..."
	@./$(BINARY_NAME)

clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)

test:
	@echo "Testing..."
	@go test
	
curl:
	curl -X POST http://82.27.116.39 -d @payload.json