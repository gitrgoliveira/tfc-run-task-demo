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

buildDocker:
	@docker build -t $(BINARY_NAME) .
	
runDocker: buildDocker
	@docker run -p 80:80 $(BINARY_NAME)