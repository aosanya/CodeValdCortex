# Deployment

This document covers deployment architecture, infrastructure requirements, and operational procedures for the work item system.

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Docker Host                          │
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌─────────────────┐  │
│  │   Gitea      │  │  PostgreSQL  │  │   ArangoDB      │  │
│  │  (Git+UI)    │  │  (Gitea DB)  │  │  (Graph+Docs)   │  │
│  │              │  │              │  │                 │  │
│  │  Port: 3000  │  │  Port: 5432  │  │  Port: 8529     │  │
│  │  RAM: 1GB    │  │  RAM: 512MB  │  │  RAM: 2-4GB     │  │
│  └──────────────┘  └──────────────┘  └─────────────────┘  │
│         │                 │                    │           │
│         └─────────────────┴────────────────────┘           │
│                           │                                │
│                  ┌────────────────┐                        │
│                  │ CodeValdCortex │                        │
│                  │   (Workflow    │                        │
│                  │  Orchestration)│                        │
│                  │                │                        │
│                  │  Port: 8080    │                        │
│                  │  RAM: 2GB      │                        │
│                  └────────────────┘                        │
│                           │                                │
└───────────────────────────┼────────────────────────────────┘
                            │
                    ┌───────┴────────┐
                    │  OpenAI API    │
                    │  (External)    │
                    └────────────────┘
```

## Docker Compose Configuration

### Complete docker-compose.yml

```yaml
version: "3.8"

services:
  # PostgreSQL - Gitea database
  postgres:
    image: postgres:15-alpine
    container_name: gitea-db
    environment:
      - POSTGRES_USER=gitea
      - POSTGRES_PASSWORD=gitea
      - POSTGRES_DB=gitea
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - codevald-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U gitea"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Gitea - Git server and issue tracking
  gitea:
    image: gitea/gitea:latest
    container_name: gitea
    environment:
      - USER_UID=1000
      - USER_GID=1000
      - GITEA__database__DB_TYPE=postgres
      - GITEA__database__HOST=postgres:5432
      - GITEA__database__NAME=gitea
      - GITEA__database__USER=gitea
      - GITEA__database__PASSWD=gitea
      - GITEA__webhook__ALLOWED_HOST_LIST=*
      - GITEA__server__DOMAIN=localhost
      - GITEA__server__SSH_DOMAIN=localhost
      - GITEA__server__ROOT_URL=http://localhost:3000/
    volumes:
      - gitea_data:/data
      - /etc/timezone:/etc/timezone:ro
      - /etc/localtime:/etc/localtime:ro
    ports:
      - "3000:3000"
      - "2222:22"
    networks:
      - codevald-network
    depends_on:
      postgres:
        condition: service_healthy
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:3000/api/v1/version"]
      interval: 30s
      timeout: 10s
      retries: 3

  # ArangoDB - Graph database
  arangodb:
    image: arangodb:latest
    container_name: arangodb
    environment:
      - ARANGO_ROOT_PASSWORD=openSesame
      - ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY=4G
    volumes:
      - arangodb_data:/var/lib/arangodb3
      - arangodb_apps:/var/lib/arangodb3-apps
    ports:
      - "8529:8529"
    networks:
      - codevald-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8529/_api/version"]
      interval: 30s
      timeout: 10s
      retries: 3

  # CodeValdCortex - Workflow orchestration
  codevaldcortex:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: codevaldcortex
    environment:
      # Gitea
      - GITEA_URL=http://gitea:3000
      - GITEA_TOKEN=${GITEA_TOKEN}
      - GITEA_WEBHOOK_SECRET=${GITEA_WEBHOOK_SECRET}
      
      # ArangoDB
      - ARANGODB_URL=http://arangodb:8529
      - ARANGODB_USER=root
      - ARANGODB_PASSWORD=openSesame
      - ARANGODB_DATABASE=codevaldcortex
      
      # OpenAI
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - OPENAI_MODEL=gpt-4
      
      # Application
      - APP_PORT=8080
      - LOG_LEVEL=info
      - WORKER_CONCURRENCY=10
    ports:
      - "8080:8080"
    networks:
      - codevald-network
    depends_on:
      gitea:
        condition: service_healthy
      arangodb:
        condition: service_healthy
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

networks:
  codevald-network:
    driver: bridge

volumes:
  postgres_data:
  gitea_data:
  arangodb_data:
  arangodb_apps:
```

### Environment Variables

Create `.env` file:

```bash
# Gitea
GITEA_TOKEN=your_gitea_personal_access_token
GITEA_WEBHOOK_SECRET=your_webhook_secret

# OpenAI
OPENAI_API_KEY=sk-your-openai-api-key
```

## Resource Requirements

### Minimum Configuration

For development/testing:

| Service         | CPU   | RAM    | Disk  |
|----------------|-------|--------|-------|
| Gitea          | 0.5   | 1 GB   | 10 GB |
| PostgreSQL     | 0.25  | 512 MB | 5 GB  |
| ArangoDB       | 0.5   | 2 GB   | 20 GB |
| CodeValdCortex | 0.5   | 2 GB   | 5 GB  |
| **Total**      | **2** | **6 GB** | **40 GB** |

**Recommended VPS**: $40-60/month (e.g., DigitalOcean 6GB droplet)

### Production Configuration

For production workloads:

| Service         | CPU  | RAM   | Disk   |
|----------------|------|-------|--------|
| Gitea          | 2    | 2 GB  | 50 GB  |
| PostgreSQL     | 1    | 2 GB  | 20 GB  |
| ArangoDB       | 2    | 8 GB  | 100 GB |
| CodeValdCortex | 2    | 4 GB  | 10 GB  |
| **Total**      | **7** | **16 GB** | **180 GB** |

**Recommended**: Dedicated server or cloud instance with 16GB+ RAM

## Deployment Steps

### 1. Initial Setup

```bash
# Clone repository
git clone https://github.com/your-org/codevaldcortex.git
cd codevaldcortex

# Create environment file
cp .env.example .env
nano .env  # Edit with your credentials

# Build and start services
docker-compose up -d

# Check service health
docker-compose ps
```

### 2. Configure Gitea

```bash
# Access Gitea UI
open http://localhost:3000

# Initial setup wizard:
# - Database: PostgreSQL (host: postgres)
# - Admin user: create account
# - SSH port: 2222
# - HTTP port: 3000
```

**Create API Token**:
1. Login to Gitea
2. Settings → Applications → Generate New Token
3. Name: "CodeValdCortex"
4. Permissions: Full access
5. Copy token to `.env` file as `GITEA_TOKEN`

### 3. Initialize ArangoDB

```bash
# Access ArangoDB UI
open http://localhost:8529

# Login with root/openSesame

# Create database
curl -X POST http://localhost:8529/_api/database \
  -u root:openSesame \
  -H "Content-Type: application/json" \
  -d '{"name": "codevaldcortex"}'
```

**Or use initialization script**:

```go
// internal/database/init.go
func InitializeDatabase(client arangodb.Client) error {
    ctx := context.Background()
    
    // Create database
    db, err := client.CreateDatabase(ctx, "codevaldcortex", nil)
    if err != nil {
        return err
    }
    
    // Create collections
    collections := []string{
        "work_items",
        "git_objects",
        "git_commits",
        "git_refs",
        "agencies",
        "agents",
        "workflow_executions",
        "llm_usage",
        "mutex_locks",
    }
    
    for _, name := range collections {
        _, err := db.CreateCollection(ctx, name, nil)
        if err != nil {
            return err
        }
    }
    
    // Create edge collections
    edgeCollections := []string{
        "commit_parents",
        "code_dependencies",
        "work_item_commits",
        "agent_work_items",
        "agent_code_expertise",
    }
    
    for _, name := range edgeCollections {
        _, err := db.CreateCollection(ctx, name, &arangodb.CreateCollectionOptions{
            Type: arangodb.CollectionTypeEdge,
        })
        if err != nil {
            return err
        }
    }
    
    // Create graphs
    graphs := []arangodb.GraphDefinition{
        {
            Name: "commit_graph",
            EdgeDefinitions: []arangodb.EdgeDefinition{
                {
                    Collection: "commit_parents",
                    From:       []string{"git_commits"},
                    To:         []string{"git_commits"},
                },
            },
        },
        {
            Name: "code_graph",
            EdgeDefinitions: []arangodb.EdgeDefinition{
                {
                    Collection: "code_dependencies",
                    From:       []string{"git_objects"},
                    To:         []string{"git_objects"},
                },
            },
        },
        {
            Name: "workflow_graph",
            EdgeDefinitions: []arangodb.EdgeDefinition{
                {
                    Collection: "work_item_commits",
                    From:       []string{"work_items"},
                    To:         []string{"git_commits"},
                },
                {
                    Collection: "agent_work_items",
                    From:       []string{"agents"},
                    To:         []string{"work_items"},
                },
            },
        },
        {
            Name: "knowledge_graph",
            EdgeDefinitions: []arangodb.EdgeDefinition{
                {
                    Collection: "agent_code_expertise",
                    From:       []string{"agents"},
                    To:         []string{"git_objects"},
                },
            },
        },
    }
    
    for _, graphDef := range graphs {
        _, err := db.CreateGraph(ctx, graphDef.Name, &arangodb.CreateGraphOptions{
            EdgeDefinitions: graphDef.EdgeDefinitions,
        })
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

### 4. Configure Webhooks

```bash
# Run setup script
go run scripts/setup-webhooks.go

# Or manually via Gitea UI:
# Repository → Settings → Webhooks → Add Webhook
# URL: http://codevaldcortex:8080/api/v1/webhooks/gitea/issues
# Events: Issues, Pull Requests
# Secret: (from .env)
```

### 5. Verify Installation

```bash
# Check all services healthy
docker-compose ps

# Test webhook
curl -X POST http://localhost:8080/api/v1/webhooks/gitea/issues \
  -H "Content-Type: application/json" \
  -d @test/webhook-payload.json

# Check logs
docker-compose logs -f codevaldcortex
```

## Operational Procedures

### Backup

```bash
# Backup script
#!/bin/bash

BACKUP_DIR=/backups/$(date +%Y%m%d)
mkdir -p $BACKUP_DIR

# Backup Gitea data
docker exec gitea tar czf - /data > $BACKUP_DIR/gitea-data.tar.gz

# Backup PostgreSQL
docker exec postgres pg_dump -U gitea gitea > $BACKUP_DIR/postgres.sql

# Backup ArangoDB
docker exec arangodb arangodump \
  --server.endpoint tcp://127.0.0.1:8529 \
  --server.password openSesame \
  --output-directory /var/lib/arangodb3/dump

docker exec arangodb tar czf - /var/lib/arangodb3/dump > $BACKUP_DIR/arangodb.tar.gz

echo "Backup completed: $BACKUP_DIR"
```

### Restore

```bash
# Restore Gitea
docker exec -i gitea tar xzf - -C / < gitea-data.tar.gz

# Restore PostgreSQL
docker exec -i postgres psql -U gitea gitea < postgres.sql

# Restore ArangoDB
docker exec -i arangodb tar xzf - -C / < arangodb.tar.gz
docker exec arangodb arangorestore \
  --server.endpoint tcp://127.0.0.1:8529 \
  --server.password openSesame \
  --input-directory /var/lib/arangodb3/dump
```

### Monitoring

```bash
# Resource usage
docker stats

# Logs
docker-compose logs -f --tail=100 codevaldcortex

# Health checks
curl http://localhost:8080/health
curl http://localhost:3000/api/v1/version
curl http://localhost:8529/_api/version
```

### Scaling

**Horizontal Scaling** (multiple CodeValdCortex instances):

```yaml
# docker-compose.yml
services:
  codevaldcortex:
    deploy:
      replicas: 3
    # ... rest of configuration
```

**Load Balancer** (nginx):

```nginx
upstream codevaldcortex {
    server localhost:8080;
    server localhost:8081;
    server localhost:8082;
}

server {
    listen 80;
    
    location / {
        proxy_pass http://codevaldcortex;
    }
}
```

## Security

### SSL/TLS

Use reverse proxy (nginx, Caddy) for HTTPS:

```nginx
server {
    listen 443 ssl http2;
    server_name gitea.example.com;
    
    ssl_certificate /etc/letsencrypt/live/gitea.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/gitea.example.com/privkey.pem;
    
    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### Firewall

```bash
# Allow only necessary ports
ufw allow 22/tcp    # SSH
ufw allow 80/tcp    # HTTP
ufw allow 443/tcp   # HTTPS
ufw enable
```

### Secrets Management

Use Docker secrets or external vault:

```yaml
# docker-compose.yml
secrets:
  gitea_token:
    file: ./secrets/gitea_token.txt
  openai_key:
    file: ./secrets/openai_key.txt

services:
  codevaldcortex:
    secrets:
      - gitea_token
      - openai_key
```

## Troubleshooting

### Common Issues

**Gitea can't connect to PostgreSQL**:
```bash
# Check network
docker-compose exec gitea ping postgres

# Check PostgreSQL logs
docker-compose logs postgres
```

**CodeValdCortex can't connect to Gitea**:
```bash
# Check Gitea is healthy
curl http://localhost:3000/api/v1/version

# Check token is valid
curl -H "Authorization: token $GITEA_TOKEN" \
  http://localhost:3000/api/v1/user
```

**ArangoDB out of memory**:
```bash
# Increase memory limit
# Edit docker-compose.yml:
environment:
  - ARANGODB_OVERRIDE_DETECTED_TOTAL_MEMORY=8G
```

---

**See Also**:
- [Observability](./observability.md) - Monitoring and metrics
- [README](./README.md) - Architecture overview
