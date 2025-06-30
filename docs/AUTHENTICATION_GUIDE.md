# NixAI Web Interface - Authentication & User Management Guide

## Overview

The NixAI web interface now supports real user authentication and management. This guide explains how to:
1. Create and manage users
2. Set up authentication for real usage
3. Configure repository access for different users
4. Manage teams and permissions

## Authentication System

### Default Setup

When you first start the web interface, it automatically creates a default admin user:
- **Username**: `admin`
- **Password**: `nixai-admin-2024`
- **Role**: `admin`

**⚠️ IMPORTANT**: Change this password immediately after first login!

### Starting the Web Interface

```bash
# Start the web interface (port 35002 by default)
nixai web start --port 35002

# Or start with a custom repository path
nixai web start --port 35002 --repo-path /path/to/your/nixos-configs
```

## User Management

### 1. Creating New Users

#### Via Web Interface (Admin Only)
1. Login as admin at `http://localhost:35002/login`
2. Navigate to user management (admin panel)
3. Click "Create User"
4. Fill in user details:
   - Username
   - Email
   - Display Name
   - Password
   - Role (admin/user)

#### Via API (Admin Only)
```bash
# Create a new user
curl -X POST http://localhost:35002/api/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -d '{
    "username": "alice",
    "email": "alice@example.com",
    "display_name": "Alice Developer",
    "password": "secure-password-123",
    "role": "user"
  }'
```

### 2. User Roles and Permissions

| Role | Permissions | Description |
|------|-------------|-------------|
| **admin** | Full access | Can manage users, teams, configurations, fleet |
| **user** | Limited access | Can create configurations, join teams, use builder |

#### Admin Permissions:
- `read`, `write`, `admin`, `fleet`, `teams`, `builder`, `user_management`

#### User Permissions:
- `read`, `write`, `builder`

### 3. Managing Users

#### List All Users (Admin Only)
```bash
curl -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  http://localhost:35002/api/users
```

#### Change Password (Any User)
```bash
curl -X POST http://localhost:35002/api/users/change-password \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "current_password": "old-password",
    "new_password": "new-secure-password"
  }'
```

## Repository Configuration

### 1. Setting Up User Repositories

Each user can have their own NixOS configuration repositories. Here's how to set them up:

#### Option A: Shared Repository Structure
```
/home/nixai-configs/
├── users/
│   ├── alice/
│   │   ├── configuration.nix
│   │   ├── hardware-configuration.nix
│   │   └── modules/
│   ├── bob/
│   │   ├── configuration.nix
│   │   └── modules/
│   └── shared/
│       ├── common.nix
│       └── modules/
└── teams/
    ├── devteam/
    │   ├── team-config.nix
    │   └── shared-modules/
    └── infrastructure/
        ├── servers.nix
        └── monitoring.nix
```

#### Option B: Individual Git Repositories
```bash
# Each user has their own git repository
/home/alice/.nixos-config/
/home/bob/.nixos-config/
/home/charlie/.nixos-config/
```

### 2. Connecting User Repositories

You can configure repository paths for different users by modifying the web server configuration:

```yaml
# web-config.yaml
repositories:
  type: "user-based"  # or "shared", "git-based"
  base_path: "/home/nixai-configs"
  user_subdirs: true
  shared_path: "/home/nixai-configs/shared"
  
users:
  alice:
    repository: "/home/alice/.nixos-config"
    branches: ["main", "testing"]
  bob:
    repository: "/home/bob/.nixos-config"
    branches: ["main", "experimental"]
```

### 3. Git Integration

For git-based repository management:

```bash
# Initialize a user's repository
mkdir -p /home/nixai-configs/users/alice
cd /home/nixai-configs/users/alice
git init
git remote add origin https://github.com/alice/nixos-config.git

# Set up automatic syncing
nixai version-control init /home/nixai-configs/users/alice
nixai version-control configure \
  --remote-url https://github.com/alice/nixos-config.git \
  --auto-sync true \
  --branch main
```

## Team Management

### 1. Creating Teams

```bash
# Create a team
curl -X POST http://localhost:35002/api/teams \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "name": "DevOps Team",
    "description": "Infrastructure and deployment team"
  }'
```

### 2. Adding Users to Teams

```bash
# Add user to team
curl -X POST http://localhost:35002/api/teams/{teamId}/members \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "user_id": "user_abc123",
    "role": "developer"
  }'
```

### 3. Team Roles

| Role | Permissions |
|------|-------------|
| **Owner** | Full team management |
| **Admin** | Manage team, deploy configs |
| **Maintainer** | Create/edit/deploy configs |
| **Developer** | Create/edit configs |
| **Viewer** | Read-only access |

## Security Best Practices

### 1. Authentication Security
- Change default admin password immediately
- Use strong passwords (minimum 12 characters)
- Enable session timeouts (24 hours by default)
- Use HTTPS in production (not implemented yet)

### 2. Repository Security
- Set proper file permissions (700 for user directories)
- Use git with SSH keys for repository access
- Implement backup strategies for user configurations
- Consider encrypted storage for sensitive configurations

### 3. Network Security
- Bind to specific interfaces in production (not 0.0.0.0)
- Use firewall rules to restrict access
- Consider VPN access for remote users

## Production Deployment

### 1. Systemd Service
```ini
[Unit]
Description=NixAI Web Interface
After=network.target

[Service]
Type=simple
User=nixai
Group=nixai
ExecStart=/usr/local/bin/nixai web start --port 8080 --config /etc/nixai/web-config.yaml
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### 2. Nginx Reverse Proxy
```nginx
server {
    listen 80;
    server_name nixai.example.com;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    # WebSocket support
    location /ws {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

## API Reference

### Authentication Endpoints
- `POST /api/auth/login` - User login
- `POST /api/auth/logout` - User logout
- `GET /api/auth/status` - Check authentication status

### User Management Endpoints (Admin Only)
- `GET /api/users` - List all users
- `POST /api/users` - Create new user
- `POST /api/users/change-password` - Change password

### Team Management Endpoints
- `GET /api/teams` - List teams
- `POST /api/teams` - Create team
- `GET /api/teams/{id}` - Get team details
- `POST /api/teams/{id}/members` - Add team member

## Troubleshooting

### 1. Login Issues
```bash
# Check auth manager status
curl http://localhost:35002/api/auth/status

# Reset admin password (requires server restart)
rm ~/.config/nixai/auth/users.json
# Restart server - new admin user will be created
```

### 2. Repository Access Issues
```bash
# Check repository permissions
ls -la /home/nixai-configs/users/

# Verify git repository status
cd /home/nixai-configs/users/alice
git status

# Check server logs
journalctl -u nixai-web -f
```

### 3. Permission Issues
```bash
# Check user permissions
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:35002/api/auth/status

# Verify file permissions
sudo chown -R nixai:nixai /home/nixai-configs/
sudo chmod -R 750 /home/nixai-configs/
```

## Next Steps

1. **SSL/TLS**: Implement HTTPS support for production
2. **LDAP Integration**: Connect to existing user directories
3. **OAuth2**: Support for GitHub, GitLab, Google authentication
4. **Audit Logging**: Track user actions and changes
5. **Backup System**: Automated backups of user configurations
6. **Role-Based Access Control**: More granular permission system

This authentication system provides a solid foundation for multi-user NixAI deployments while maintaining security and ease of use.
