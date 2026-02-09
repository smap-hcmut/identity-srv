#!/bin/bash

set -e

echo "🚀 Setting up SMAP Auth Service Development Environment"
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Create secrets directory
echo "📁 Creating secrets directory..."
mkdir -p secrets

# Generate RSA keypair for JWT
if [ ! -f "secrets/jwt-private.pem" ]; then
    echo "🔐 Generating RSA keypair for JWT signing..."
    openssl genrsa -out secrets/jwt-private.pem 2048
    openssl rsa -in secrets/jwt-private.pem -pubout -out secrets/jwt-public.pem
    chmod 600 secrets/jwt-private.pem
    chmod 644 secrets/jwt-public.pem
    echo -e "${GREEN}✓ RSA keypair generated${NC}"
else
    echo -e "${YELLOW}⚠ RSA keypair already exists, skipping...${NC}"
fi

# Generate encryption key for encrypter
if [ ! -f "secrets/encrypt.key" ]; then
    echo "🔐 Generating encryption key..."
    openssl rand -base64 32 > secrets/encrypt.key
    echo -e "${GREEN}✓ Encryption key generated${NC}"
else
    echo -e "${YELLOW}⚠ Encryption key already exists, skipping...${NC}"
fi

# Create auth-config.yaml if not exists
if [ ! -f "auth-config.yaml" ]; then
    echo "📝 Creating auth-config.yaml from template..."
    cp auth-config.example.yaml auth-config.yaml
    
    # Update paths in config
    sed -i '' 's|/secrets/jwt-private.pem|./secrets/jwt-private.pem|g' auth-config.yaml
    sed -i '' 's|/secrets/jwt-public.pem|./secrets/jwt-public.pem|g' auth-config.yaml
    
    echo -e "${GREEN}✓ auth-config.yaml created${NC}"
    echo -e "${YELLOW}⚠ Please update auth-config.yaml with your Google OAuth credentials${NC}"
else
    echo -e "${YELLOW}⚠ auth-config.yaml already exists, skipping...${NC}"
fi

# Create .env file if not exists
if [ ! -f ".env" ]; then
    echo "📝 Creating .env file..."
    ENCRYPT_KEY=$(cat secrets/encrypt.key)
    cat > .env << EOF
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=smap_auth
DB_SSL_MODE=disable
DB_SCHEMA=public

# Encryption Key
ENCRYPT_KEY=${ENCRYPT_KEY}

# Google OAuth (Update these with your credentials)
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret

# Redis
REDIS_PASSWORD=

# PostgreSQL Password
POSTGRES_PASSWORD=postgres

# Discord Webhook (Optional)
DISCORD_WEBHOOK_ID=
DISCORD_WEBHOOK_TOKEN=
EOF
    echo -e "${GREEN}✓ .env file created${NC}"
    echo -e "${YELLOW}⚠ Please update .env with your Google OAuth credentials${NC}"
else
    echo -e "${YELLOW}⚠ .env already exists, skipping...${NC}"
fi

echo ""
echo "🐳 Starting Docker containers..."
docker-compose -f docker-compose.dev.yml up -d

echo ""
echo "⏳ Waiting for services to be ready..."
sleep 5

# Check if PostgreSQL is ready
echo "🔍 Checking PostgreSQL..."
until docker exec smap-postgres pg_isready -U postgres > /dev/null 2>&1; do
    echo "Waiting for PostgreSQL..."
    sleep 2
done
echo -e "${GREEN}✓ PostgreSQL is ready${NC}"

# Check if Redis is ready
echo "🔍 Checking Redis..."
until docker exec smap-redis redis-cli ping > /dev/null 2>&1; do
    echo "Waiting for Redis..."
    sleep 2
done
echo -e "${GREEN}✓ Redis is ready${NC}"

echo ""
echo "✅ Development environment setup complete!"
echo ""
echo "📋 Next steps:"
echo "1. Update auth-config.yaml with your Google OAuth credentials"
echo "2. Update .env with your Google OAuth credentials"
echo "3. Run: make models (to generate SQLBoiler models)"
echo "4. Run: make run-api (to start the API server)"
echo ""
echo "🔗 Useful commands:"
echo "  - View logs: docker-compose -f docker-compose.dev.yml logs -f"
echo "  - Stop services: docker-compose -f docker-compose.dev.yml down"
echo "  - Restart services: docker-compose -f docker-compose.dev.yml restart"
echo ""
