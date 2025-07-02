# NixOS 25.05 Server Configuration
# Template: Minimal Server Setup
# Description: A secure minimal server configuration with SSH, firewall, and basic monitoring
{pkgs, ...}: {
  imports = [
    ./hardware-configuration.nix
  ];

  # System configuration
  system.stateVersion = "25.05";

  # Boot configuration
  boot.loader.systemd-boot.enable = true;
  boot.loader.efi.canTouchEfiVariables = true;

  # Networking
  networking.hostName = "nixos-server";
  networking.networkmanager.enable = true;

  # Firewall
  networking.firewall = {
    enable = true;
    allowedTCPPorts = [22 80 443];
    allowedUDPPorts = [];
  };

  # Localization
  time.timeZone = "UTC";
  i18n.defaultLocale = "en_US.UTF-8";

  # No desktop environment - server only
  services.xserver.enable = false;

  # User account
  users.users.admin = {
    isNormalUser = true;
    description = "Server Administrator";
    extraGroups = ["wheel" "networkmanager"];
    openssh.authorizedKeys.keys = [
      # Add your SSH public key here
      # "ssh-rsa AAAAB3NzaC1yc2E... your-key@example.com"
    ];
  };

  # System packages
  environment.systemPackages = with pkgs; [
    # Essential tools
    vim
    git
    wget
    curl
    unzip
    zip

    # System monitoring
    htop
    btop
    iotop

    # Network tools
    nmap
    tcpdump

    # Security tools
    fail2ban

    # System administration
    tmux
    screen
  ];

  # Services
  services.openssh = {
    enable = true;
    settings = {
      PasswordAuthentication = false;
      PermitRootLogin = "no";
      X11Forwarding = false;
    };
  };

  # Fail2ban for SSH protection
  services.fail2ban = {
    enable = true;
    jails = {
      sshd = {
        settings = {
          enabled = true;
          port = "ssh";
          filter = "sshd";
          logpath = "/var/log/auth.log";
          maxretry = 5;
          bantime = 3600;
        };
      };
    };
  };

  # Automatic updates
  system.autoUpgrade = {
    enable = true;
    dates = "daily";
    allowReboot = false;
  };

  # Automatic garbage collection
  nix.gc = {
    automatic = true;
    dates = "daily";
    options = "--delete-older-than 7d";
  };

  # Enable flakes
  nix.settings.experimental-features = ["nix-command" "flakes"];

  # Security hardening
  security.sudo.wheelNeedsPassword = true;

  # System monitoring
  services.prometheus = {
    exporters = {
      node = {
        enable = true;
        enabledCollectors = ["systemd"];
        port = 9100;
      };
    };
  };
}
