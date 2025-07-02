# NixOS 25.05 Development Environment
# Template: Complete Development Setup
# Description: Full development environment with multiple languages, tools, and IDEs
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
  networking.hostName = "nixos-dev";
  networking.networkmanager.enable = true;

  # Localization
  time.timeZone = "Europe/Amsterdam";
  i18n.defaultLocale = "en_US.UTF-8";

  # X11 and i3 (lightweight for development)
  services.xserver = {
    enable = true;
    displayManager.lightdm.enable = true;
    windowManager.i3.enable = true;
    xkb = {
      layout = "us";
      variant = "";
    };
  };

  # Audio
  hardware.pulseaudio.enable = false;
  security.rtkit.enable = true;
  services.pipewire = {
    enable = true;
    alsa.enable = true;
    alsa.support32Bit = true;
    pulse.enable = true;
  };

  # User account
  users.users.developer = {
    isNormalUser = true;
    description = "Developer";
    extraGroups = ["wheel" "networkmanager" "docker" "audio" "video"];
    shell = pkgs.zsh;
  };

  # Enable ZSH
  programs.zsh.enable = true;

  # Development packages
  environment.systemPackages = with pkgs; [
    # Editors and IDEs
    vscode
    vim
    neovim
    emacs

    # Version control
    git
    gh
    gitlab-runner

    # Programming languages
    nodejs_20
    python311
    python311Packages.pip
    python311Packages.virtualenv
    go
    rustc
    cargo
    gcc
    gnumake
    cmake

    # Package managers
    yarn
    npm

    # Databases
    postgresql
    sqlite
    redis

    # Containers
    docker
    docker-compose
    podman

    # Cloud tools
    awscli2
    kubectl
    terraform

    # System tools
    htop
    btop
    tree
    jq
    yq
    curl
    wget
    unzip
    zip

    # Terminal tools
    alacritty
    tmux
    zsh
    oh-my-zsh
    starship

    # Network tools
    nmap
    wireshark

    # Development utilities
    postman
    insomnia

    # Fonts for coding
    jetbrains-mono
    fira-code
    source-code-pro
  ];

  # Services
  services.docker.enable = true;
  services.postgresql = {
    enable = true;
    package = pkgs.postgresql_15;
    authentication = pkgs.lib.mkOverride 10 ''
      local all all trust
      host all all 127.0.0.1/32 trust
      host all all ::1/128 trust
    '';
  };

  services.redis.servers.main = {
    enable = true;
    port = 6379;
  };

  services.openssh.enable = true;

  # Fonts
  fonts.packages = with pkgs; [
    noto-fonts
    noto-fonts-cjk-sans
    noto-fonts-emoji
    liberation_ttf
    fira-code
    fira-code-symbols
    jetbrains-mono
    source-code-pro
  ];

  # Shell configuration
  programs.starship.enable = true;

  # Virtualization
  virtualisation.docker.enable = true;
  virtualisation.virtualbox.host.enable = true;

  # Development environment variables
  environment.variables = {
    EDITOR = "code";
    BROWSER = "firefox";
  };

  # Enable flakes
  nix.settings.experimental-features = ["nix-command" "flakes"];

  # Security
  security.sudo.wheelNeedsPassword = false;

  # Automatic garbage collection
  nix.gc = {
    automatic = true;
    dates = "weekly";
    options = "--delete-older-than 30d";
  };
}
