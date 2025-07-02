# NixOS 25.05 Gaming Configuration
# Template: Gaming Setup with Steam and Performance Optimizations
# Description: Optimized gaming configuration with Steam, graphics drivers, and performance tweaks
{
  config,
  pkgs,
  ...
}: {
  imports = [
    ./hardware-configuration.nix
  ];

  # System configuration
  system.stateVersion = "25.05";

  # Boot configuration
  boot.loader.systemd-boot.enable = true;
  boot.loader.efi.canTouchEfiVariables = true;

  # Kernel for gaming performance
  boot.kernelPackages = pkgs.linuxPackages_latest;

  # Networking
  networking.hostName = "nixos-gaming";
  networking.networkmanager.enable = true;

  # Localization
  time.timeZone = "Europe/Amsterdam";
  i18n.defaultLocale = "en_US.UTF-8";

  # X11 and KDE Plasma (great for gaming)
  services.xserver = {
    enable = true;
    displayManager.sddm.enable = true;
    desktopManager.plasma5.enable = true;
    xkb = {
      layout = "us";
      variant = "";
    };
  };

  # Graphics drivers
  services.xserver.videoDrivers = ["nvidia"]; # Change to "amdgpu" for AMD cards

  # NVIDIA specific settings (remove if using AMD)
  hardware.nvidia = {
    modesetting.enable = true;
    powerManagement.enable = false;
    powerManagement.finegrained = false;
    open = false;
    nvidiaSettings = true;
    package = config.boot.kernelPackages.nvidiaPackages.stable;
  };

  # OpenGL
  hardware.opengl = {
    enable = true;
    driSupport = true;
    driSupport32Bit = true;
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
  users.users.gamer = {
    isNormalUser = true;
    description = "Gamer";
    extraGroups = ["wheel" "networkmanager" "audio" "video" "gamemode"];
    packages = with pkgs; [
      # Gaming platforms
      steam
      lutris
      heroic
      bottles

      # Game launchers
      discord
      teamspeak_client

      # Streaming
      obs-studio

      # System monitoring
      mangohud
      gamemode

      # Web browser
      firefox

      # Media
      vlc
      spotify
    ];
  };

  # System packages
  environment.systemPackages = with pkgs; [
    # Essential tools
    vim
    git
    wget
    curl

    # System monitoring
    htop
    btop

    # Gaming tools
    steam-run
    wine
    winetricks

    # Performance monitoring
    mangohud
    gamemode

    # Hardware monitoring
    lm_sensors

    # Archive tools
    unzip
    zip
    p7zip
  ];

  # Gaming services
  programs.steam = {
    enable = true;
    remotePlay.openFirewall = true;
    dedicatedServer.openFirewall = true;
  };

  programs.gamemode.enable = true;

  # Services
  services.openssh.enable = true;

  # Fonts
  fonts.packages = with pkgs; [
    noto-fonts
    noto-fonts-cjk-sans
    noto-fonts-emoji
    liberation_ttf
    fira-code
  ];

  # Performance optimizations
  # Increase file descriptor limits
  security.pam.loginLimits = [
    {
      domain = "*";
      type = "soft";
      item = "nofile";
      value = "1048576";
    }
    {
      domain = "*";
      type = "hard";
      item = "nofile";
      value = "1048576";
    }
  ];

  # Kernel parameters for gaming
  boot.kernel.sysctl = {
    "vm.max_map_count" = 2147483642;
    "fs.file-max" = 2097152;
  };

  # Enable 32-bit support for Steam
  # (driSupport32Bit already enabled in hardware.opengl above)
  hardware.pulseaudio.support32Bit = true;

  # Network optimizations for gaming
  networking.firewall = {
    allowedTCPPorts = [27015 27036]; # Steam
    allowedUDPPorts = [27015 27031 27032 27033 27034 27035 27036]; # Steam
  };

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
