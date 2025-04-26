#!/bin/bash

GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Print a line of dashes
print_line() {
  echo "------------------------------------------"
}

# Navigate to the docker directory
cd docker || exit

print_line
echo "Setting up database container..."
docker compose down -v >/dev/null 2>&1
docker compose up -d >/dev/null 2>&1

# Wait for the database to be ready (with minimal output)
echo "Waiting for database to be ready..."
until docker exec capstone_db pg_isready -U dbuser -d capstone_db -q; do
  sleep 1
done
echo "Database is ready"

# Verify database tables exist (quietly)
echo "Verifying database tables..."
TABLES=$(docker exec capstone_db psql -U dbuser -d capstone_db -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('user_course_progress', 'user_lesson_progress');" 2>/dev/null)
if [ "$(echo $TABLES | tr -d ' ')" -lt 2 ]; then
  echo "Initializing database tables..."
  docker exec -i capstone_db psql -U dbuser -d capstone_db < ./init.sql >/dev/null 2>&1
fi
print_line
# Navigate back to the root directory
cd ..

# Run the Go application with exactly one line above and below
echo -e "${GREEN}Running Go application...${NC}"
print_line

go run main.go
