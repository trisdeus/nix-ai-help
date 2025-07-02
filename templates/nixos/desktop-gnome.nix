# NixOS 25.05 Desktop Configuration with GNOME
# Template: Desktop GNOME Environment
# Description: A complete desktop setup with GNOME desktop environment, multimedia support, and development tools
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
  networking.hostName = "nixos-desktop";
  networking.networkmanager.enable = true;

  # Localization
  time.timeZone = "Europe/Amsterdam";
  i18n.defaultLocale = "en_US.UTF-8";
  i18n.extraLocaleSettings = {
    LC_ADDRESS = "nl_NL.UTF-8";
    LC_IDENTIFICATION = "nl_NL.UTF-8";
    LC_MEASUREMENT = "nl_NL.UTF-8";
    LC_MONETARY = "nl_NL.UTF-8";
    LC_NAME = "nl_NL.UTF-8";
    LC_NUMERIC = "nl_NL.UTF-8";
    LC_PAPER = "nl_NL.UTF-8";
    LC_TELEPHONE = "nl_NL.UTF-8";
    LC_TIME = "nl_NL.UTF-8";
  };

  # X11 and GNOME
  services.xserver = {
    enable = true;
    displayManager.gdm.enable = true;
    desktopManager.gnome.enable = true;
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
  users.users.nixuser = {
    isNormalUser = true;
    description = "NixOS User";
    extraGroups = ["networkmanager" "wheel" "audio" "video"];
    packages = with pkgs; [
      # Desktop applications
      firefox
      thunderbird
      libreoffice
      gimp
      vlc

      # Development tools
      vscode
      git
      gh

      # System utilities
      htop
      neofetch
      tree
      wget
      curl
    ];
  };

  # System packages
  environment.systemPackages = with pkgs; [
    # Essential tools
    vim
    nano
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
    networkmanager
    networkmanagerapplet

    # Development
    gcc
    gnumake

    # GNOME extensions
    gnome-tweaks
    gnome-extension-manager
  ];

  # Services
  services.printing.enable = true;
  services.flatpak.enable = true;
  services.openssh.enable = true;

  # Fonts
  fonts.packages = with pkgs; [
    noto-fonts
    noto-fonts-cjk-sans
    noto-fonts-emoji
    liberation_ttf
    fira-code
    fira-code-symbols
  ];

  # Security
  security.sudo.wheelNeedsPassword = false;

  # Automatic garbage collection
  nix.gc = {
    automatic = true;
    dates = "weekly";
    options = "--delete-older-than 30d";
  };

  # Enable flakes
  nix.settings.experimental-features = ["nix-command" "flakes"];
}
