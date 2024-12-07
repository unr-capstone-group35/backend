#!/bin/bash

# Navigate to the docker directory
cd docker || exit

# Check if the container exists
if docker ps -a | grep -q "capstone_db"; then
    # Stop and remove containers, volumes, networks
    echo "Container exists. Stopping and removing..."
    docker-compose down -v
fi

# Start the containers in detached mode
echo "Starting containers..."
docker-compose up -d

# Navigate back to the root directory
cd ..

# Run the Go application
echo "Running Go application..."
go run main.go
