#!/bin/bash
set -e

echo "=========================================="
echo "  ReadBud - Starting all services"
echo "=========================================="

# Check prerequisites
command -v docker >/dev/null 2>&1 || { echo "Error: Docker is required but not installed."; exit 1; }
docker compose version >/dev/null 2>&1 || { echo "Error: Docker Compose V2 is required."; exit 1; }

# Build and start
echo ""
echo "Building and starting services..."
docker compose up -d --build

# Wait for API
echo ""
echo "Waiting for services to be ready..."
for i in $(seq 1 30); do
  if curl -s http://localhost:19881/health > /dev/null 2>&1; then
    break
  fi
  echo "  Waiting for API... ($i)"
  sleep 2
done

if ! curl -s http://localhost:19881/health > /dev/null 2>&1; then
  echo "Error: API failed to start. Check logs: docker compose logs api"
  exit 1
fi

echo "API is ready."

# Create default admin user (idempotent)
echo "Ensuring default admin user exists..."
docker compose exec -T postgres psql -U readbud -d readbud -c "
INSERT INTO users (public_id, username, password_hash, nickname, role, status, created_at, updated_at)
VALUES (
  '01READBUDADMIN0000000000',
  'admin@readbud.com',
  '\$2b\$10\$iLpux6uSii6Llh3BmzdOGeRBOCzhSYqs5BhLHZVCAJpSqoGNaKC3C',
  '管理员', 'admin', 1, NOW(), NOW()
) ON CONFLICT (username) DO NOTHING;
" 2>/dev/null || true

echo ""
echo "=========================================="
echo "  ReadBud is running!"
echo "=========================================="
echo ""
echo "  Frontend:  http://localhost:19880"
echo "  API:       http://localhost:19881"
echo ""
echo "  Default login:"
echo "    Username: admin@readbud.com"
echo "    Password: admin198808"
echo ""
echo "  Useful commands:"
echo "    docker compose logs -f api      # API logs"
echo "    docker compose logs -f worker   # Worker logs"
echo "    docker compose ps               # Service status"
echo "    ./stop.sh                        # Stop all"
echo ""
echo "=========================================="
