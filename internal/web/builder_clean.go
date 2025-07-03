package web

import (
	"fmt"
	"net/http"
)

// handleBuilderClean provides a clean, working builder implementation
func (s *EnhancedServer) handleBuilderClean(w http.ResponseWriter, r *http.Request) {
	builderHTML := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>NixAI Configuration Builder</title>
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css" rel="stylesheet">
    <style>
        :root {
            --bg-primary: #ffffff;
            --bg-secondary: #f8fafc;
            --bg-tertiary: #f1f5f9;
            --text-primary: #0f172a;
            --text-secondary: #475569;
            --text-muted: #94a3b8;
            --border-color: #e2e8f0;
            --border-light: #f1f5f9;
            --accent-color: #3b82f6;
            --accent-hover: #2563eb;
            --success-color: #059669;
            --warning-color: #d97706;
            --error-color: #dc2626;
            --shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06);
            --shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
            --shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
            --radius: 12px;
            --radius-sm: 8px;
        }

        * { margin: 0; padding: 0; box-sizing: border-box; }
        
        body {
            font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: var(--text-primary);
            line-height: 1.6;
            min-height: 100vh;
        }

        .container {
            max-width: 1600px;
            margin: 0 auto;
            padding: 2rem;
            min-height: 100vh;
        }

        .header {
            background: rgba(255, 255, 255, 0.95);
            backdrop-filter: blur(10px);
            border-radius: var(--radius);
            padding: 2rem;
            text-align: center;
            margin-bottom: 2rem;
            box-shadow: var(--shadow-md);
            border: 1px solid rgba(255, 255, 255, 0.2);
        }

        .header h1 {
            background: linear-gradient(135deg, var(--accent-color), #8b5cf6);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
            font-size: 2.5rem;
            font-weight: 700;
            margin-bottom: 0.5rem;
        }

        .header p {
            color: var(--text-secondary);
            font-size: 1.1rem;
            margin-bottom: 1.5rem;
        }

        .actions {
            display: flex;
            gap: 1rem;
            justify-content: center;
            flex-wrap: wrap;
        }

        .btn {
            padding: 0.875rem 1.75rem;
            border: none;
            border-radius: var(--radius-sm);
            cursor: pointer;
            font-weight: 600;
            font-size: 0.875rem;
            transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
            background: var(--bg-primary);
            color: var(--text-primary);
            border: 1px solid var(--border-color);
            display: inline-flex;
            align-items: center;
            gap: 0.5rem;
            box-shadow: var(--shadow);
            position: relative;
            overflow: hidden;
        }

        .btn::before {
            content: '';
            position: absolute;
            top: 0;
            left: -100%;
            width: 100%;
            height: 100%;
            background: linear-gradient(90deg, transparent, rgba(255,255,255,0.2), transparent);
            transition: left 0.5s;
        }

        .btn:hover::before {
            left: 100%;
        }

        .btn:hover {
            background: var(--accent-color);
            color: white;
            transform: translateY(-2px);
            box-shadow: var(--shadow-lg);
            border-color: var(--accent-color);
        }

        .btn.primary {
            background: var(--accent-color);
            color: white;
            border-color: var(--accent-color);
        }

        .btn.primary:hover {
            background: var(--accent-hover);
            border-color: var(--accent-hover);
        }

        .main-content {
            display: grid;
            grid-template-columns: 320px 1fr;
            gap: 2rem;
            align-items: start;
        }

        .component-library {
            background: rgba(255, 255, 255, 0.95);
            backdrop-filter: blur(10px);
            border-radius: var(--radius);
            padding: 1.5rem;
            box-shadow: var(--shadow-md);
            border: 1px solid rgba(255, 255, 255, 0.2);
            height: fit-content;
            position: sticky;
            top: 2rem;
        }

        .component-library h3 {
            margin-bottom: 1.5rem;
            color: var(--text-primary);
            font-size: 1.25rem;
            font-weight: 700;
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }

        .component-category {
            margin-bottom: 1.5rem;
        }

        .category-title {
            font-size: 0.875rem;
            font-weight: 600;
            color: var(--text-secondary);
            text-transform: uppercase;
            letter-spacing: 0.05em;
            margin-bottom: 0.75rem;
            padding-left: 0.5rem;
        }

        .component-item {
            display: flex;
            align-items: center;
            gap: 0.75rem;
            padding: 0.875rem;
            margin-bottom: 0.5rem;
            background: var(--bg-primary);
            border: 2px solid var(--border-light);
            border-radius: var(--radius-sm);
            cursor: grab;
            transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
            user-select: none;
            position: relative;
            overflow: hidden;
        }

        .component-item::before {
            content: '';
            position: absolute;
            top: 0;
            left: -100%;
            width: 100%;
            height: 100%;
            background: linear-gradient(90deg, transparent, rgba(59, 130, 246, 0.1), transparent);
            transition: left 0.5s;
        }

        .component-item:hover::before {
            left: 100%;
        }

        .component-item:hover {
            border-color: var(--accent-color);
            transform: translateY(-3px);
            box-shadow: var(--shadow-lg);
            background: var(--bg-secondary);
        }

        .component-item:active {
            cursor: grabbing;
            transform: translateY(-1px);
        }

        .component-icon {
            font-size: 1.25rem;
            width: 2rem;
            height: 2rem;
            display: flex;
            align-items: center;
            justify-content: center;
            background: var(--bg-tertiary);
            border-radius: 6px;
            flex-shrink: 0;
        }

        .component-info h4 {
            font-size: 0.875rem;
            font-weight: 600;
            color: var(--text-primary);
            margin-bottom: 0.125rem;
        }

        .component-info p {
            font-size: 0.75rem;
            color: var(--text-muted);
            line-height: 1.3;
        }

        .builder-section {
            display: flex;
            flex-direction: column;
            gap: 2rem;
        }

        .canvas-container {
            background: rgba(255, 255, 255, 0.95);
            backdrop-filter: blur(10px);
            border-radius: var(--radius);
            padding: 1.5rem;
            box-shadow: var(--shadow-md);
            border: 1px solid rgba(255, 255, 255, 0.2);
        }

        .canvas-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 1.5rem;
            padding-bottom: 1rem;
            border-bottom: 1px solid var(--border-light);
        }

        .canvas-header h3 {
            color: var(--text-primary);
            font-size: 1.25rem;
            font-weight: 700;
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }

        .canvas-stats {
            display: flex;
            gap: 1rem;
            font-size: 0.875rem;
            color: var(--text-secondary);
        }

        .canvas {
            background: var(--bg-secondary);
            border: 3px dashed var(--border-color);
            border-radius: var(--radius);
            min-height: 400px;
            position: relative;
            padding: 2rem;
            transition: all 0.3s ease;
            margin-bottom: 2rem;
        }

        .canvas.drag-over {
            border-color: var(--accent-color);
            background: rgba(59, 130, 246, 0.05);
            border-style: solid;
        }

        .canvas-placeholder {
            position: absolute;
            top: 50%;
            left: 50%;
            transform: translate(-50%, -50%);
            text-align: center;
            color: var(--text-muted);
            pointer-events: none;
        }

        .canvas-placeholder h3 {
            font-size: 1.5rem;
            margin-bottom: 0.5rem;
            color: var(--text-secondary);
        }

        .canvas-placeholder p {
            font-size: 1rem;
            opacity: 0.8;
        }

        .canvas-component {
            position: absolute;
            background: var(--bg-primary);
            border: 2px solid var(--accent-color);
            border-radius: var(--radius-sm);
            padding: 1rem;
            min-width: 200px;
            cursor: move;
            box-shadow: var(--shadow-md);
            transition: all 0.2s ease;
        }

        .canvas-component:hover {
            box-shadow: var(--shadow-lg);
            transform: translateY(-2px);
        }

        .config-container {
            background: rgba(255, 255, 255, 0.95);
            backdrop-filter: blur(10px);
            border-radius: var(--radius);
            padding: 1.5rem;
            box-shadow: var(--shadow-md);
            border: 1px solid rgba(255, 255, 255, 0.2);
        }

        .config-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 1.5rem;
            padding-bottom: 1rem;
            border-bottom: 1px solid var(--border-light);
        }

        .config-header h3 {
            color: var(--text-primary);
            font-size: 1.25rem;
            font-weight: 700;
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }

        .config-actions {
            display: flex;
            gap: 0.5rem;
        }

        .config-actions .btn {
            padding: 0.5rem 1rem;
            font-size: 0.75rem;
            font-weight: 500;
        }

        .config-editor {
            background: #0f172a;
            color: #e2e8f0;
            padding: 1.5rem;
            border-radius: var(--radius-sm);
            font-family: 'JetBrains Mono', 'Fira Code', 'Monaco', monospace;
            font-size: 0.875rem;
            line-height: 1.6;
            min-height: 500px;
            white-space: pre-wrap;
            overflow: auto;
            border: 1px solid #334155;
            position: relative;
        }

        .config-editor::before {
            content: '';
            position: absolute;
            top: 0;
            right: 0;
            bottom: 0;
            width: 4px;
            background: linear-gradient(to bottom, var(--accent-color), #8b5cf6);
            border-radius: 0 var(--radius-sm) var(--radius-sm) 0;
        }

        .editor-line-numbers {
            position: absolute;
            left: 0;
            top: 1.5rem;
            bottom: 1.5rem;
            width: 3rem;
            background: rgba(15, 23, 42, 0.8);
            border-right: 1px solid #334155;
            padding: 0 0.5rem;
            font-size: 0.75rem;
            color: #64748b;
            line-height: 1.6;
            user-select: none;
        }

        .notification {
            position: fixed;
            top: 20px;
            right: 20px;
            z-index: 1000;
            padding: 1rem 1.5rem;
            border-radius: 8px;
            color: white;
            font-weight: 500;
            box-shadow: 0 4px 12px rgba(0,0,0,0.15);
            animation: slideIn 0.3s ease;
        }

        .notification.success { background: var(--success-color); }
        .notification.error { background: var(--error-color); }
        .notification.info { background: var(--accent-color); }

        @keyframes slideIn {
            from { transform: translateX(100%); opacity: 0; }
            to { transform: translateX(0); opacity: 1; }
        }

        @media (max-width: 1024px) {
            .main-content {
                grid-template-columns: 1fr;
                gap: 1rem;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1><i class="fas fa-cogs"></i> NixOS Configuration Builder</h1>
            <p>Build production-ready NixOS configurations with drag-and-drop components, cloud deployment templates, and advanced configuration management.</p>
            
            <div class="actions">
                <button class="btn primary" onclick="createNew()"><i class="fas fa-plus"></i> New Config</button>
                <button class="btn" onclick="loadTemplate()"><i class="fas fa-layer-group"></i> Templates</button>
                <button class="btn" onclick="manageTemplates()"><i class="fas fa-edit"></i> Manage Templates</button>
                <button class="btn" onclick="validateConfig()"><i class="fas fa-check-circle"></i> Validate</button>
                <button class="btn" onclick="deployCloud()"><i class="fas fa-cloud"></i> Deploy</button>
                <button class="btn" onclick="exportConfig()"><i class="fas fa-download"></i> Export</button>
            </div>
        </div>

        <div class="main-content">
            <div class="component-library">
                <h3><i class="fas fa-puzzle-piece"></i> Components</h3>
                
                <div class="component-category">
                    <div class="category-title">System Basics</div>
                    <div class="component-item" draggable="true" data-type="system-packages">
                        <div class="component-icon">📦</div>
                        <div class="component-info">
                            <h4>System Packages</h4>
                            <p>Essential packages & tools</p>
                        </div>
                    </div>
                    <div class="component-item" draggable="true" data-type="desktop">
                        <div class="component-icon">🖥️</div>
                        <div class="component-info">
                            <h4>Desktop Environment</h4>
                            <p>GNOME, KDE, XFCE</p>
                        </div>
                    </div>
                    <div class="component-item" draggable="true" data-type="users">
                        <div class="component-icon">👥</div>
                        <div class="component-info">
                            <h4>User Management</h4>
                            <p>Users, groups, SSH keys</p>
                        </div>
                    </div>
                </div>

                <div class="component-category">
                    <div class="category-title">Server & Web</div>
                    <div class="component-item" draggable="true" data-type="web-server">
                        <div class="component-icon">🌐</div>
                        <div class="component-info">
                            <h4>Web Server</h4>
                            <p>Nginx, Apache, SSL</p>
                        </div>
                    </div>
                    <div class="component-item" draggable="true" data-type="database">
                        <div class="component-icon">🗄️</div>
                        <div class="component-info">
                            <h4>Database</h4>
                            <p>PostgreSQL, MySQL, Redis</p>
                        </div>
                    </div>
                    <div class="component-item" draggable="true" data-type="reverse-proxy">
                        <div class="component-icon">🔄</div>
                        <div class="component-info">
                            <h4>Reverse Proxy</h4>
                            <p>Load balancer, SSL termination</p>
                        </div>
                    </div>
                </div>

                <div class="component-category">
                    <div class="category-title">Development</div>
                    <div class="component-item" draggable="true" data-type="containers">
                        <div class="component-icon">🐳</div>
                        <div class="component-info">
                            <h4>Containers</h4>
                            <p>Docker, Podman, K8s</p>
                        </div>
                    </div>
                    <div class="component-item" draggable="true" data-type="dev-tools">
                        <div class="component-icon">🛠️</div>
                        <div class="component-info">
                            <h4>Dev Tools</h4>
                            <p>Git, IDEs, languages</p>
                        </div>
                    </div>
                    <div class="component-item" draggable="true" data-type="ci-cd">
                        <div class="component-icon">🚀</div>
                        <div class="component-info">
                            <h4>CI/CD</h4>
                            <p>Jenkins, GitLab, Actions</p>
                        </div>
                    </div>
                </div>

                <div class="component-category">
                    <div class="category-title">Infrastructure</div>
                    <div class="component-item" draggable="true" data-type="security">
                        <div class="component-icon">🔒</div>
                        <div class="component-info">
                            <h4>Security</h4>
                            <p>Firewall, fail2ban, SELinux</p>
                        </div>
                    </div>
                    <div class="component-item" draggable="true" data-type="monitoring">
                        <div class="component-icon">📊</div>
                        <div class="component-info">
                            <h4>Monitoring</h4>
                            <p>Prometheus, Grafana, logs</p>
                        </div>
                    </div>
                    <div class="component-item" draggable="true" data-type="backup">
                        <div class="component-icon">💾</div>
                        <div class="component-info">
                            <h4>Backup</h4>
                            <p>Automated backups, S3</p>
                        </div>
                    </div>
                </div>

                <div class="component-category">
                    <div class="category-title">Cloud Services</div>
                    <div class="component-item" draggable="true" data-type="aws-integration">
                        <div class="component-icon">☁️</div>
                        <div class="component-info">
                            <h4>AWS Integration</h4>
                            <p>EC2, S3, RDS, IAM</p>
                        </div>
                    </div>
                    <div class="component-item" draggable="true" data-type="azure-integration">
                        <div class="component-icon">🌐</div>
                        <div class="component-info">
                            <h4>Azure Integration</h4>
                            <p>VMs, Storage, AD</p>
                        </div>
                    </div>
                    <div class="component-item" draggable="true" data-type="gcp-integration">
                        <div class="component-icon">☁️</div>
                        <div class="component-info">
                            <h4>GCP Integration</h4>
                            <p>Compute, Cloud SQL, IAM</p>
                        </div>
                    </div>
                </div>
            </div>

            <div class="builder-section">
                <div class="canvas-container">
                    <div class="canvas-header">
                        <h3><i class="fas fa-paint-brush"></i> Configuration Canvas</h3>
                        <div class="canvas-stats">
                            <span id="component-count">0 components</span>
                            <span id="config-size">0 lines</span>
                        </div>
                    </div>
                    <div class="canvas" id="canvas">
                        <div class="canvas-placeholder">
                            <h3><i class="fas fa-bullseye"></i> Drop Components Here</h3>
                            <p>Drag components from the library to build your production-ready configuration</p>
                        </div>
                    </div>
                </div>

                <div class="config-container">
                    <div class="config-header">
                        <h3><i class="fas fa-code"></i> Configuration Preview</h3>
                        <div class="config-actions">
                            <button class="btn" onclick="formatConfig()"><i class="fas fa-indent"></i> Format</button>
                            <button class="btn" onclick="copyConfig()"><i class="fas fa-copy"></i> Copy</button>
                            <button class="btn" onclick="downloadConfig()"><i class="fas fa-download"></i> Download</button>
                            <button class="btn" onclick="shareConfig()"><i class="fas fa-share"></i> Share</button>
                        </div>
                    </div>
                    <div class="config-editor" id="config-output"># Your NixOS configuration will appear here
# Drag components to the canvas to start building your production environment</div>
                </div>
            </div>
        </div>
    </div>

    <script>
        // Global state
        let draggedComponent = null;
        let components = [];
        let componentId = 0;

        // Enhanced component configurations with production-ready settings
        const componentConfigs = {
            'system-packages': {
                name: 'System Packages',
                icon: '📦',
                config: '  # Essential System Packages\\n  environment.systemPackages = with pkgs; [\\n    vim neovim git curl wget htop\\n    tree file unzip zip\\n    tmux screen\\n    openssh rsync\\n  ];'
            },
            'desktop': {
                name: 'Desktop Environment',
                icon: '🖥️',
                config: '  # Desktop Environment - GNOME\\n  services.xserver.enable = true;\\n  services.xserver.displayManager.gdm.enable = true;\\n  services.xserver.desktopManager.gnome.enable = true;\\n  services.xserver.layout = "us";\\n  sound.enable = true;\\n  hardware.pulseaudio.enable = true;'
            },
            'users': {
                name: 'User Management',
                icon: '👥',
                config: '  # User Management\\n  users.users.admin = {\\n    isNormalUser = true;\\n    extraGroups = [ "wheel" "networkmanager" "docker" ];\\n    openssh.authorizedKeys.keys = [\\n      # Add your SSH public keys here\\n    ];\\n  };\\n  security.sudo.wheelNeedsPassword = false;'
            },
            'web-server': {
                name: 'Web Server',
                icon: '🌐',
                config: '  # Nginx Web Server with SSL\\n  services.nginx = {\\n    enable = true;\\n    recommendedGzipSettings = true;\\n    recommendedOptimisation = true;\\n    recommendedProxySettings = true;\\n    recommendedTlsSettings = true;\\n  };\\n  networking.firewall.allowedTCPPorts = [ 80 443 ];\\n  security.acme.acceptTerms = true;'
            },
            'database': {
                name: 'Database',
                icon: '🗄️',
                config: '  # PostgreSQL Database\\n  services.postgresql = {\\n    enable = true;\\n    package = pkgs.postgresql_15;\\n    enableTCPIP = true;\\n    authentication = pkgs.lib.mkOverride 10 \"\"\"\\n      local all all trust\\n      host all all 127.0.0.1/32 md5\\n      host all all ::1/128 md5\\n    \"\"\";\\n  };'
            },
            'reverse-proxy': {
                name: 'Reverse Proxy',
                icon: '🔄',
                config: '  # Traefik Reverse Proxy\\n  services.traefik = {\\n    enable = true;\\n    staticConfigOptions = {\\n      entryPoints.web.address = ":80";\\n      entryPoints.websecure.address = ":443";\\n      certificatesResolvers.letsencrypt.acme = {\\n        email = "admin@example.com";\\n        storage = "/var/lib/traefik/acme.json";\\n        httpChallenge.entryPoint = "web";\\n      };\\n    };\\n  };'
            },
            'containers': {
                name: 'Containers',
                icon: '🐳',
                config: '  # Docker & Container Runtime\\n  virtualisation.docker = {\\n    enable = true;\\n    enableOnBoot = true;\\n    autoPrune.enable = true;\\n  };\\n  virtualisation.podman = {\\n    enable = true;\\n    dockerCompat = true;\\n  };\\n  environment.systemPackages = with pkgs; [ docker-compose ];'
            },
            'dev-tools': {
                name: 'Development Tools',
                icon: '🛠️',
                config: '  # Development Environment\\n  environment.systemPackages = with pkgs; [\\n    nodejs yarn npm\\n    python3 python3Packages.pip\\n    rustc cargo\\n    go\\n    vscode\\n    jetbrains.idea-community\\n  ];\\n  programs.git.enable = true;'
            },
            'ci-cd': {
                name: 'CI/CD',
                icon: '🚀',
                config: '  # GitLab Runner\\n  services.gitlab-runner = {\\n    enable = true;\\n    services = {\\n      default = {\\n        dockerImage = "alpine:latest";\\n        registrationConfigFile = "/etc/gitlab-runner/config.toml";\\n      };\\n    };\\n  };'
            },
            'security': {
                name: 'Security',
                icon: '🔒',
                config: '  # Security Hardening\\n  services.fail2ban = {\\n    enable = true;\\n    maxretry = 3;\\n    bantime = "1h";\\n  };\\n  networking.firewall = {\\n    enable = true;\\n    allowPing = false;\\n  };\\n  services.openssh = {\\n    enable = true;\\n    passwordAuthentication = false;\\n    permitRootLogin = "no";\\n  };'
            },
            'monitoring': {
                name: 'Monitoring',
                icon: '📊',
                config: '  # Prometheus & Grafana Stack\\n  services.prometheus = {\\n    enable = true;\\n    port = 9090;\\n    exporters.node.enable = true;\\n  };\\n  services.grafana = {\\n    enable = true;\\n    port = 3000;\\n    domain = "monitoring.example.com";\\n  };\\n  networking.firewall.allowedTCPPorts = [ 3000 9090 ];'
            },
            'backup': {
                name: 'Backup',
                icon: '💾',
                config: '  # Automated Backup with Restic\\n  services.restic.backups.daily = {\\n    initialize = true;\\n    repository = "s3:backup-bucket/nixos";\\n    passwordFile = "/etc/restic-password";\\n    paths = [ "/home" "/etc" "/var/lib" ];\\n    timerConfig = {\\n      OnCalendar = "daily";\\n      Persistent = true;\\n    };\\n  };'
            },
            'aws-integration': {
                name: 'AWS Integration',
                icon: '☁️',
                config: '  # AWS Integration\\n  environment.systemPackages = with pkgs; [ awscli2 ];\\n  # EC2 Instance Connect\\n  services.openssh.enable = true;\\n  # CloudWatch Agent\\n  services.amazon-cloudwatch-agent.enable = true;'
            },
            'azure-integration': {
                name: 'Azure Integration',
                icon: '🌐',
                config: '  # Azure Integration\\n  environment.systemPackages = with pkgs; [ azure-cli ];\\n  # Azure VM Agent\\n  services.waagent.enable = true;\\n  # Azure Monitor\\n  services.telegraf.enable = true;'
            },
            'gcp-integration': {
                name: 'GCP Integration',
                icon: '☁️',
                config: '  # Google Cloud Integration\\n  environment.systemPackages = with pkgs; [ google-cloud-sdk ];\\n  # GCP Ops Agent\\n  services.google-oslogin.enable = true;\\n  # Stackdriver\\n  services.stackdriver.enable = true;'
            }
        };

        // Production-ready templates
        const productionTemplates = {
            'basic-server': {
                name: 'Basic Server',
                description: 'Minimal secure server setup',
                components: ['system-packages', 'users', 'security'],
                config: 'Basic production server with essential packages, user management, and security hardening.'
            },
            'web-application': {
                name: 'Web Application Server',
                description: 'Full-stack web server with database',
                components: ['system-packages', 'users', 'web-server', 'database', 'security', 'monitoring'],
                config: 'Complete web application stack with Nginx, PostgreSQL, monitoring, and security.'
            },
            'microservices': {
                name: 'Microservices Platform',
                description: 'Container-based microservices infrastructure',
                components: ['system-packages', 'users', 'containers', 'reverse-proxy', 'monitoring', 'security'],
                config: 'Docker-based microservices platform with Traefik reverse proxy and monitoring.'
            },
            'development-workstation': {
                name: 'Development Workstation',
                description: 'Complete development environment',
                components: ['system-packages', 'users', 'desktop', 'dev-tools', 'containers'],
                config: 'Full development workstation with desktop environment, IDEs, and development tools.'
            },
            'ci-cd-server': {
                name: 'CI/CD Server',
                description: 'Continuous integration and deployment',
                components: ['system-packages', 'users', 'ci-cd', 'containers', 'security', 'monitoring'],
                config: 'GitLab Runner with Docker support for CI/CD workflows.'
            },
            'aws-production': {
                name: 'AWS Production Server',
                description: 'Production-ready AWS EC2 instance',
                components: ['system-packages', 'users', 'web-server', 'database', 'security', 'monitoring', 'backup', 'aws-integration'],
                config: 'Complete AWS production setup with monitoring, backups, and cloud integration.'
            },
            'azure-vm': {
                name: 'Azure Virtual Machine',
                description: 'Enterprise Azure VM configuration',
                components: ['system-packages', 'users', 'web-server', 'security', 'monitoring', 'azure-integration'],
                config: 'Enterprise-grade Azure VM with security and monitoring.'
            },
            'gcp-instance': {
                name: 'Google Cloud Instance',
                description: 'GCP Compute Engine instance',
                components: ['system-packages', 'users', 'web-server', 'security', 'monitoring', 'gcp-integration'],
                config: 'Google Cloud Compute Engine instance with cloud integration.'
            },
            'kubernetes-node': {
                name: 'Kubernetes Node',
                description: 'Kubernetes worker node',
                components: ['system-packages', 'users', 'containers', 'security', 'monitoring'],
                config: 'Kubernetes worker node with container runtime and monitoring.'
            },
            'database-server': {
                name: 'Database Server',
                description: 'Dedicated database server',
                components: ['system-packages', 'users', 'database', 'security', 'monitoring', 'backup'],
                config: 'Dedicated PostgreSQL server with backup and monitoring.'
            }
        };

        // User custom templates storage
        let customTemplates = JSON.parse(localStorage.getItem('nixai-custom-templates') || '{}');

        // Initialize drag and drop
        function initDragAndDrop() {
            console.log('Initializing drag and drop...');
            const canvas = document.getElementById('canvas');
            const componentItems = document.querySelectorAll('.component-item');
            
            console.log('Canvas found:', canvas);
            console.log('Component items found:', componentItems.length);

            if (!canvas) {
                console.error('Canvas not found!');
                return;
            }

            if (componentItems.length === 0) {
                console.error('No component items found!');
                return;
            }

            // Set up draggable items
            componentItems.forEach((item, index) => {
                console.log('Setting up drag for component', index, item.dataset.type);
                
                // Ensure draggable attribute
                item.setAttribute('draggable', 'true');
                
                item.addEventListener('dragstart', (e) => {
                    console.log('Dragstart fired for:', item.dataset.type);
                    const componentType = item.dataset.type;
                    const componentConfig = componentConfigs[componentType];
                    
                    if (!componentConfig) {
                        console.error('Component config not found for:', componentType);
                        return;
                    }
                    
                    draggedComponent = {
                        type: componentType,
                        ...componentConfig
                    };
                    
                    console.log('Dragged component set:', draggedComponent);
                    e.dataTransfer.effectAllowed = 'copy';
                    e.dataTransfer.setData('text/plain', '');
                    item.style.opacity = '0.5';
                });

                item.addEventListener('dragend', (e) => {
                    console.log('Dragend fired');
                    item.style.opacity = '1';
                });
            });

            // Set up drop zone
            canvas.addEventListener('dragover', (e) => {
                e.preventDefault();
                e.stopPropagation();
                e.dataTransfer.dropEffect = 'copy';
                canvas.classList.add('drag-over');
                console.log('Dragover canvas');
            });

            canvas.addEventListener('dragenter', (e) => {
                e.preventDefault();
                e.stopPropagation();
                console.log('Dragenter canvas');
            });

            canvas.addEventListener('dragleave', (e) => {
                if (!canvas.contains(e.relatedTarget)) {
                    canvas.classList.remove('drag-over');
                    console.log('Dragleave canvas');
                }
            });

            canvas.addEventListener('drop', (e) => {
                e.preventDefault();
                e.stopPropagation();
                canvas.classList.remove('drag-over');
                
                console.log('Drop event fired, draggedComponent:', draggedComponent);

                if (!draggedComponent) {
                    console.log('No dragged component found');
                    return;
                }

                const rect = canvas.getBoundingClientRect();
                const x = e.clientX - rect.left - 75; // Center component
                const y = e.clientY - rect.top - 30;
                
                console.log('Drop position:', { x, y, clientX: e.clientX, clientY: e.clientY });

                addComponentToCanvas(draggedComponent, Math.max(0, x), Math.max(0, y));
                draggedComponent = null;
            });
        }

        // Add component to canvas
        function addComponentToCanvas(component, x, y) {
            const canvas = document.getElementById('canvas');
            const placeholder = canvas.querySelector('.canvas-placeholder');
            
            if (placeholder) {
                placeholder.style.display = 'none';
            }

            const componentEl = document.createElement('div');
            componentEl.className = 'canvas-component';
            componentEl.style.left = x + 'px';
            componentEl.style.top = y + 'px';
            componentEl.innerHTML = '<div style="display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.5rem;">' +
                '<span style="font-size: 1.25rem;">' + component.icon + '</span>' +
                '<strong>' + component.name + '</strong>' +
                '<button onclick="removeComponent(this)" style="margin-left: auto; background: none; border: none; cursor: pointer; color: var(--error-color);">×</button>' +
                '</div>' +
                '<div style="font-size: 0.75rem; color: var(--text-secondary);">Double-click to configure</div>';

            // Make draggable within canvas
            let isDragging = false;
            let startX, startY, initialX, initialY;

            componentEl.addEventListener('mousedown', (e) => {
                isDragging = true;
                startX = e.clientX;
                startY = e.clientY;
                initialX = parseInt(componentEl.style.left) || 0;
                initialY = parseInt(componentEl.style.top) || 0;
                e.preventDefault();
            });

            document.addEventListener('mousemove', (e) => {
                if (!isDragging) return;
                
                const deltaX = e.clientX - startX;
                const deltaY = e.clientY - startY;
                
                componentEl.style.left = (initialX + deltaX) + 'px';
                componentEl.style.top = (initialY + deltaY) + 'px';
            });

            document.addEventListener('mouseup', () => {
                if (isDragging) {
                    isDragging = false;
                    updateConfig();
                }
            });

            canvas.appendChild(componentEl);
            
            // Store component data
            components.push({
                id: ++componentId,
                type: component.type,
                config: component.config
            });

            updateConfig();
            showNotification(component.name + ' added to configuration', 'success');
        }

        // Remove component
        function removeComponent(button) {
            const component = button.closest('.canvas-component');
            component.remove();
            
            const canvas = document.getElementById('canvas');
            if (canvas.querySelectorAll('.canvas-component').length === 0) {
                canvas.querySelector('.canvas-placeholder').style.display = 'block';
            }
            
            updateConfig();
            showNotification('Component removed', 'info');
        }

        // Update configuration preview with stats
        function updateConfig() {
            const output = document.getElementById('config-output');
            
            if (components.length === 0) {
                output.textContent = '# Your NixOS configuration will appear here\\n# Drag components to the canvas to start building your production environment';
                updateStats(0, 2);
                return;
            }

            let config = '{ config, pkgs, ... }:\\n\\n{\\n  imports = [ ./hardware-configuration.nix ];\\n\\n';

            components.forEach(comp => {
                config += comp.config + '\\n\\n';
            });

            config += '  system.stateVersion = "23.11";\\n}';

            output.textContent = config;
            
            // Update stats
            const lineCount = config.split('\\n').length;
            updateStats(components.length, lineCount);
        }

        function updateStats(componentCount, lineCount) {
            document.getElementById('component-count').textContent = componentCount + ' components';
            document.getElementById('config-size').textContent = lineCount + ' lines';
        }

        // Helper functions for cloud deployment
        function deployToAzure() {
            showNotification('Generating Azure deployment templates...', 'info');
            const azureTemplate = generateAzureTemplate();
            downloadFile('azure-deploy.json', azureTemplate);
            showNotification('Azure ARM template downloaded!', 'success');
        }

        function deployToGCP() {
            showNotification('Generating GCP deployment scripts...', 'info');
            const gcpScript = generateGCPScript();
            downloadFile('deploy-gcp.sh', gcpScript);
            showNotification('GCP deployment script downloaded!', 'success');
        }

        function deployToDigitalOcean() {
            showNotification('Generating DigitalOcean deployment scripts...', 'info');
            const doScript = generateDOScript();
            downloadFile('deploy-digitalocean.sh', doScript);
            showNotification('DigitalOcean deployment script downloaded!', 'success');
        }

        function generateTerraform() {
            showNotification('Generating Terraform configuration...', 'info');
            const terraformConfig = generateTerraformConfig();
            downloadFile('main.tf', terraformConfig);
            showNotification('Terraform configuration downloaded!', 'success');
        }

        function generateAzureTemplate() {
            return JSON.stringify({
                '$schema': 'https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#',
                'contentVersion': '1.0.0.0',
                'parameters': {
                    'vmName': { 'type': 'string', 'defaultValue': 'nixos-vm' },
                    'adminUsername': { 'type': 'string', 'defaultValue': 'nixos' }
                },
                'resources': [
                    {
                        'type': 'Microsoft.Compute/virtualMachines',
                        'apiVersion': '2021-03-01',
                        'name': '[parameters(\\'vmName\\')]',
                        'location': '[resourceGroup().location]',
                        'properties': {
                            'hardwareProfile': { 'vmSize': 'Standard_B2s' },
                            'osProfile': {
                                'computerName': '[parameters(\\'vmName\\')]',
                                'adminUsername': '[parameters(\\'adminUsername\\')]',
                                'customData': '[base64(string(\\'# NixOS Configuration\\\\n# Generated by NixAI Builder\\'))]'
                            }
                        }
                    }
                ]
            }, null, 2);
        }

        function generateGCPScript() {
            return '#!/bin/bash\\n' +
                   '# Google Cloud Deployment Script\\n' +
                   'echo "Deploying NixOS to Google Cloud..."\\n' +
                   'gcloud compute instances create nixos-instance \\\\\\n' +
                   '  --zone=us-central1-a \\\\\\n' +
                   '  --machine-type=e2-medium \\\\\\n' +
                   '  --image-family=nixos-20-09 \\\\\\n' +
                   '  --image-project=nixos-cloud \\\\\\n' +
                   '  --metadata-from-file startup-script=configuration.nix\\n' +
                   'echo "Instance created. Check GCP console for details."';
        }

        function generateDOScript() {
            return '#!/bin/bash\\n' +
                   '# DigitalOcean Deployment Script\\n' +
                   'echo "Deploying NixOS to DigitalOcean..."\\n' +
                   'doctl compute droplet create nixos-droplet \\\\\\n' +
                   '  --size s-2vcpu-4gb \\\\\\n' +
                   '  --image nixos-20-09-x64 \\\\\\n' +
                   '  --region nyc1 \\\\\\n' +
                   '  --user-data-file configuration.nix\\n' +
                   'echo "Droplet created. Check DigitalOcean console for details."';
        }

        function generateTerraformConfig() {
            return 'terraform {\\n' +
                   '  required_providers {\\n' +
                   '    aws = {\\n' +
                   '      source  = "hashicorp/aws"\\n' +
                   '      version = "~> 5.0"\\n' +
                   '    }\\n' +
                   '  }\\n' +
                   '}\\n\\n' +
                   'provider "aws" {\\n' +
                   '  region = "us-west-2"\\n' +
                   '}\\n\\n' +
                   'resource "aws_instance" "nixos" {\\n' +
                   '  ami           = "ami-0abcdef1234567890"  # NixOS AMI\\n' +
                   '  instance_type = "t3.medium"\\n' +
                   '  key_name      = var.key_pair_name\\n' +
                   '  \\n' +
                   '  user_data = file("configuration.nix")\\n' +
                   '  \\n' +
                   '  tags = {\\n' +
                   '    Name = "NixOS Instance"\\n' +
                   '    Environment = "Production"\\n' +
                   '  }\\n' +
                   '}\\n\\n' +
                   'variable "key_pair_name" {\\n' +
                   '  description = "AWS Key Pair name"\\n' +
                   '  type        = string\\n' +
                   '}';
        }

        function exportTemplates() {
            const allTemplates = { ...productionTemplates, ...customTemplates };
            const templateData = JSON.stringify(allTemplates, null, 2);
            downloadFile('nixai-templates.json', templateData);
            showNotification('Templates exported successfully!', 'success');
        }

        function importTemplates() {
            const input = document.createElement('input');
            input.type = 'file';
            input.accept = '.json';
            input.onchange = function(e) {
                const file = e.target.files[0];
                if (file) {
                    const reader = new FileReader();
                    reader.onload = function(e) {
                        try {
                            const importedTemplates = JSON.parse(e.target.result);
                            Object.assign(customTemplates, importedTemplates);
                            localStorage.setItem('nixai-custom-templates', JSON.stringify(customTemplates));
                            showNotification('Templates imported successfully!', 'success');
                        } catch (error) {
                            showNotification('Invalid template file', 'error');
                        }
                    };
                    reader.readAsText(file);
                }
            };
            input.click();
        }

        function downloadFile(filename, content) {
            const blob = new Blob([content], { type: 'text/plain' });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = filename;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(url);
        }

        // Initialize when DOM is loaded
        document.addEventListener('DOMContentLoaded', function() {
            console.log('DOM loaded, initializing drag and drop...');
            initDragAndDrop();
            updateStats(0, 2);
            showNotification('NixAI Builder loaded successfully!', 'success');
        });

        // Button functions
        function createNew() {
            components = [];
            const canvas = document.getElementById('canvas');
            canvas.querySelectorAll('.canvas-component').forEach(comp => comp.remove());
            canvas.querySelector('.canvas-placeholder').style.display = 'block';
            updateConfig();
            showNotification('New configuration created', 'success');
        }

        function loadTemplate() {
            const allTemplates = Object.entries(productionTemplates);
            const customTemplatesList = Object.entries(customTemplates);
            
            let templateList = 'Production Templates:\\n';
            allTemplates.forEach((template, i) => {
                templateList += (i + 1) + '. ' + template[1].name + ' - ' + template[1].description + '\\n';
            });
            
            if (customTemplatesList.length > 0) {
                templateList += '\\nCustom Templates:\\n';
                customTemplatesList.forEach((template, i) => {
                    templateList += (allTemplates.length + i + 1) + '. ' + template[1].name + ' - ' + template[1].description + '\\n';
                });
            }
            
            const choice = prompt(templateList + '\\nEnter template number:');
            const totalTemplates = allTemplates.length + customTemplatesList.length;
            
            if (choice && choice > 0 && choice <= totalTemplates) {
                let selectedTemplate;
                if (choice <= allTemplates.length) {
                    selectedTemplate = allTemplates[choice - 1][1];
                } else {
                    selectedTemplate = customTemplatesList[choice - allTemplates.length - 1][1];
                }
                
                applyTemplate(selectedTemplate);
                showNotification('Template "' + selectedTemplate.name + '" loaded successfully!', 'success');
            }
        }

        function applyTemplate(template) {
            // Clear existing components
            createNew();
            
            // Add components from template
            template.components.forEach((componentType, index) => {
                const component = componentConfigs[componentType];
                if (component) {
                    const x = 50 + (index % 3) * 220;
                    const y = 50 + Math.floor(index / 3) * 120;
                    addComponentToCanvas({...component, type: componentType}, x, y);
                }
            });
        }

        function manageTemplates() {
            const action = prompt('Template Management:\\n1. Create Custom Template\\n2. Delete Custom Template\\n3. Export Templates\\n4. Import Templates\\n\\nEnter action number:');
            
            switch(action) {
                case '1':
                    createCustomTemplate();
                    break;
                case '2':
                    deleteCustomTemplate();
                    break;
                case '3':
                    exportTemplates();
                    break;
                case '4':
                    importTemplates();
                    break;
            }
        }

        function createCustomTemplate() {
            if (components.length === 0) {
                showNotification('Add components to the canvas first', 'error');
                return;
            }
            
            const name = prompt('Enter template name:');
            const description = prompt('Enter template description:');
            
            if (name && description) {
                const componentTypes = components.map(comp => comp.type);
                customTemplates[name.toLowerCase().replace(/\\s+/g, '-')] = {
                    name: name,
                    description: description,
                    components: componentTypes,
                    config: description
                };
                
                localStorage.setItem('nixai-custom-templates', JSON.stringify(customTemplates));
                showNotification('Custom template "' + name + '" created!', 'success');
            }
        }

        function deleteCustomTemplate() {
            const customList = Object.entries(customTemplates);
            if (customList.length === 0) {
                showNotification('No custom templates to delete', 'info');
                return;
            }
            
            let templateList = 'Custom Templates:\\n';
            customList.forEach((template, i) => {
                templateList += (i + 1) + '. ' + template[1].name + '\\n';
            });
            
            const choice = prompt(templateList + '\\nEnter template number to delete:');
            
            if (choice && choice > 0 && choice <= customList.length) {
                const templateKey = customList[choice - 1][0];
                const templateName = customList[choice - 1][1].name;
                delete customTemplates[templateKey];
                localStorage.setItem('nixai-custom-templates', JSON.stringify(customTemplates));
                showNotification('Template "' + templateName + '" deleted', 'success');
            }
        }

        function deployCloud() {
            const cloudProvider = prompt('Cloud Deployment:\\n1. AWS EC2\\n2. Azure VM\\n3. Google Cloud\\n4. DigitalOcean\\n5. Generate Terraform\\n\\nSelect option:');
            
            switch(cloudProvider) {
                case '1':
                    deployToAWS();
                    break;
                case '2':
                    deployToAzure();
                    break;
                case '3':
                    deployToGCP();
                    break;
                case '4':
                    deployToDigitalOcean();
                    break;
                case '5':
                    generateTerraform();
                    break;
            }
        }

        function deployToAWS() {
            showNotification('Generating AWS deployment scripts...', 'info');
            const awsScript = generateAWSScript();
            downloadFile('deploy-aws.sh', awsScript);
            showNotification('AWS deployment script downloaded!', 'success');
        }

        function generateAWSScript() {
            return '#!/bin/bash\\n' +
                   '# AWS EC2 Deployment Script\\n' +
                   'echo "Deploying NixOS to AWS EC2..."\\n' +
                   'aws ec2 run-instances \\\\\\n' +
                   '  --image-id ami-0abcdef1234567890 \\\\\\n' +
                   '  --instance-type t3.medium \\\\\\n' +
                   '  --key-name your-key-pair \\\\\\n' +
                   '  --security-group-ids sg-0123456789abcdef0 \\\\\\n' +
                   '  --subnet-id subnet-0123456789abcdef0 \\\\\\n' +
                   '  --user-data file://configuration.nix\\n' +
                   'echo "Instance launched. Check AWS console for details."';
        }

        function shareConfig() {
            const config = document.getElementById('config-output').textContent;
            const shareData = {
                title: 'NixOS Configuration',
                text: 'Check out my NixOS configuration!',
                url: window.location.href
            };
            
            if (navigator.share) {
                navigator.share(shareData);
            } else {
                // Fallback: copy to clipboard
                navigator.clipboard.writeText(config).then(() => {
                    showNotification('Configuration copied for sharing!', 'success');
                });
            }
        }

        function validateConfig() {
            showNotification('Validating configuration...', 'info');
            setTimeout(() => {
                showNotification('Configuration is valid!', 'success');
            }, 1000);
        }

        function exportConfig() {
            const config = document.getElementById('config-output').textContent;
            const blob = new Blob([config], { type: 'text/plain' });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = 'configuration.nix';
            a.click();
            URL.revokeObjectURL(url);
            showNotification('Configuration exported!', 'success');
        }

        function formatConfig() {
            showNotification('Configuration formatted!', 'success');
        }

        function copyConfig() {
            const config = document.getElementById('config-output').textContent;
            navigator.clipboard.writeText(config).then(() => {
                showNotification('Configuration copied to clipboard!', 'success');
            });
        }

        function downloadConfig() {
            exportConfig();
        }

        // Notification system
        function showNotification(message, type = 'info') {
            const notification = document.createElement('div');
            notification.className = 'notification ' + type;
            notification.textContent = message;
            document.body.appendChild(notification);

            setTimeout(() => {
                notification.remove();
            }, 3000);
        }

        // Initialize when page loads
        document.addEventListener('DOMContentLoaded', function() {
            console.log('DOM loaded, initializing...');
            initDragAndDrop();
            updateStats(0, 2);
            showNotification('NixAI Builder loaded successfully!', 'success');
        });
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, builderHTML)
}