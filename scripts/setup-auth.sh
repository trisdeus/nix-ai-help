#!/bin/bash

# NixAI Authentication Setup Script
# This script helps set up user authentication and repository configuration

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NIXAI_PORT=${NIXAI_PORT:-35002}
NIXAI_HOST=${NIXAI_HOST:-localhost}
REPO_BASE_PATH=${REPO_BASE_PATH:-"$HOME/nixai-configs"}

echo -e "${BLUE}🔐 NixAI Authentication Setup${NC}"
echo "=================================="
echo

# Function to print colored output
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Function to check if nixai is running
check_nixai_running() {
    if curl -s "http://$NIXAI_HOST:$NIXAI_PORT/api/health" > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# Function to get admin token
get_admin_token() {
    local response
    response=$(curl -s -X POST "http://$NIXAI_HOST:$NIXAI_PORT/api/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"username":"admin","password":"nixai-admin-2024"}' 2>/dev/null)
    
    if echo "$response" | grep -q '"success":true'; then
        echo "$response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4
    else
        echo ""
    fi
}

# Function to create repository structure
setup_repository_structure() {
    local repo_path="$1"
    
    print_info "Setting up repository structure at $repo_path"
    
    # Create base directories
    mkdir -p "$repo_path/users"
    mkdir -p "$repo_path/teams"
    mkdir -p "$repo_path/shared/modules"
    mkdir -p "$repo_path/templates"
    
    # Create shared common configuration
    cat > "$repo_path/shared/common.nix" << 'EOF'
# Shared common NixOS configuration
{ config, pkgs, ... }:

{
  # Basic system packages available to all users
  environment.systemPackages = with pkgs; [
    git
    curl
    wget
    htop
    vim
    nano
    tree
  ];

  # Common services
  services.openssh = {
    enable = true;
    settings = {
      PermitRootLogin = "no";
      PasswordAuthentication = false;
    };
  };

  # Basic firewall
  networking.firewall = {
    enable = true;
    allowedTCPPorts = [ 22 ];
  };
}
EOF

    # Create user template
    cat > "$repo_path/templates/user-template.nix" << 'EOF'
# NixOS Configuration Template for {{USERNAME}}
{ config, pkgs, ... }:

{
  imports = [
    ../shared/common.nix
    ./hardware-configuration.nix
  ];

  # Boot loader
  boot.loader.systemd-boot.enable = true;
  boot.loader.efi.canTouchEfiVariables = true;

  # Networking
  networking.hostName = "{{USERNAME}}-nixos";
  networking.networkmanager.enable = true;

  # Time zone and locale
  time.timeZone = "UTC";
  i18n.defaultLocale = "en_US.UTF-8";

  # User account
  users.users.{{USERNAME}} = {
    isNormalUser = true;
    description = "{{DISPLAY_NAME}}";
    extraGroups = [ "networkmanager" "wheel" ];
    packages = with pkgs; [
      firefox
      thunderbird
      vscode
    ];
  };

  # System packages
  environment.systemPackages = with pkgs; [
    nixai
  ];

  # Enable the NixAI service
  # services.nixai.enable = true;

  system.stateVersion = "24.05";
}
EOF

    # Create team template
    cat > "$repo_path/templates/team-template.nix" << 'EOF'
# Team Configuration Template
{ config, pkgs, ... }:

{
  imports = [
    ../shared/common.nix
  ];

  # Team-specific packages
  environment.systemPackages = with pkgs; [
    docker
    docker-compose
    kubectl
    terraform
  ];

  # Team services
  virtualisation.docker.enable = true;
  
  # Add team users to docker group
  users.groups.docker.members = [
    # Add team member usernames here
  ];
}
EOF

    # Set permissions
    chmod -R 755 "$repo_path"
    
    print_success "Repository structure created at $repo_path"
}

# Function to create a new user
create_user() {
    local token="$1"
    local username="$2"
    local email="$3"
    local display_name="$4"
    local password="$5"
    local role="${6:-user}"
    
    print_info "Creating user: $username"
    
    local response
    response=$(curl -s -X POST "http://$NIXAI_HOST:$NIXAI_PORT/api/users" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $token" \
        -d "{\"username\":\"$username\",\"email\":\"$email\",\"display_name\":\"$display_name\",\"password\":\"$password\",\"role\":\"$role\"}")
    
    if echo "$response" | grep -q '"success":true\|"id":'; then
        print_success "User $username created successfully"
        
        # Create user repository directory
        local user_repo="$REPO_BASE_PATH/users/$username"
        mkdir -p "$user_repo"
        
        # Create user configuration from template
        sed "s/{{USERNAME}}/$username/g; s/{{DISPLAY_NAME}}/$display_name/g" \
            "$REPO_BASE_PATH/templates/user-template.nix" > "$user_repo/configuration.nix"
        
        # Create basic hardware configuration
        cat > "$user_repo/hardware-configuration.nix" << 'EOF'
# Hardware configuration for this machine
{ config, lib, pkgs, modulesPath, ... }:

{
  imports = [ ];

  # Add your hardware-specific configuration here
  # This is typically generated by nixos-generate-config
}
EOF

        # Initialize git repository if git is available
        if command -v git > /dev/null 2>&1; then
            cd "$user_repo"
            git init
            git add .
            git commit -m "Initial NixOS configuration for $username"
            cd - > /dev/null
            print_success "Git repository initialized for $username"
        fi
        
        print_success "User repository created at $user_repo"
        return 0
    else
        print_error "Failed to create user $username"
        echo "Response: $response"
        return 1
    fi
}

# Main script
main() {
    echo "This script will help you set up NixAI authentication and user management."
    echo "Make sure NixAI web interface is running before proceeding."
    echo

    # Check if NixAI is running
    print_info "Checking if NixAI is running on http://$NIXAI_HOST:$NIXAI_PORT"
    if ! check_nixai_running; then
        print_error "NixAI web interface is not running!"
        echo "Please start it with: nixai web start --port $NIXAI_PORT"
        exit 1
    fi
    print_success "NixAI is running"

    # Get admin token
    print_info "Logging in as default admin user"
    ADMIN_TOKEN=$(get_admin_token)
    if [ -z "$ADMIN_TOKEN" ]; then
        print_error "Failed to login as admin user"
        echo "Please check if the default admin credentials are correct:"
        echo "Username: admin"
        echo "Password: nixai-admin-2024"
        exit 1
    fi
    print_success "Admin login successful"

    # Setup repository structure
    if [ ! -d "$REPO_BASE_PATH" ]; then
        print_warning "Repository base path does not exist"
        read -p "Create repository structure at $REPO_BASE_PATH? (y/n): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            setup_repository_structure "$REPO_BASE_PATH"
        else
            print_info "Skipping repository setup"
        fi
    else
        print_success "Repository structure already exists at $REPO_BASE_PATH"
    fi

    # Interactive user creation
    echo
    print_info "User Creation Wizard"
    echo "You can create new users for the NixAI system."
    echo "Press Ctrl+C to skip user creation."
    echo

    while true; do
        echo "Create a new user:"
        read -p "Username: " username
        [ -z "$username" ] && break
        
        read -p "Email: " email
        read -p "Display Name: " display_name
        read -p "Password: " -s password
        echo
        
        echo "Select role:"
        echo "1) user (default)"
        echo "2) admin"
        read -p "Choice (1-2): " role_choice
        
        case $role_choice in
            2) role="admin" ;;
            *) role="user" ;;
        esac
        
        if create_user "$ADMIN_TOKEN" "$username" "$email" "$display_name" "$password" "$role"; then
            echo
            print_success "User setup complete!"
            echo "User can login at: http://$NIXAI_HOST:$NIXAI_PORT/login"
            echo "Repository location: $REPO_BASE_PATH/users/$username"
        fi
        
        echo
        read -p "Create another user? (y/n): " -n 1 -r
        echo
        [[ ! $REPLY =~ ^[Yy]$ ]] && break
        echo
    done

    # Final instructions
    echo
    print_success "Setup complete!"
    echo
    echo "🎉 Your NixAI authentication system is ready!"
    echo
    echo "Next steps:"
    echo "1. 🔐 Change the default admin password at: http://$NIXAI_HOST:$NIXAI_PORT/login"
    echo "2. 👥 Create teams and add users to them"
    echo "3. 📁 Set up user repositories and configurations"
    echo "4. 🚀 Start building NixOS configurations!"
    echo
    echo "📚 Documentation: $PWD/docs/AUTHENTICATION_GUIDE.md"
    echo "🌐 Web Interface: http://$NIXAI_HOST:$NIXAI_PORT"
    echo
    print_warning "Remember to change the default admin password!"
}

# Run main function
main "$@"
