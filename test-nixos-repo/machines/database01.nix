# Database Server Configuration
{
  config,
  pkgs,
  ...
}: {
  # Set hostname
  networking.hostName = "database01";

  # Enable PostgreSQL
  services.postgresql = {
    enable = true;
    package = pkgs.postgresql_15;
    enableTCPIP = true;
    authentication = pkgs.lib.mkOverride 10 ''
      local all all trust
      host all all 127.0.0.1/32 trust
      host all all ::1/128 trust
      host all all 192.168.1.0/24 md5
    '';
  };

  # Enable Redis
  services.redis = {
    servers."main" = {
      enable = true;
      port = 6379;
    };
  };

  # Enable SSH
  services.openssh = {
    enable = true;
    permitRootLogin = "no";
  };

  # Enable firewall
  networking.firewall = {
    enable = true;
    allowedTCPPorts = [22 5432 6379];
  };

  # Network configuration
  networking.interfaces.eth0.ipv4.addresses = [
    {
      address = "192.168.1.101";
      prefixLength = 24;
    }
  ];
  networking.defaultGateway = "192.168.1.1";
  networking.nameservers = ["8.8.8.8" "1.1.1.1"];

  # Install database packages
  environment.systemPackages = with pkgs; [
    postgresql
    redis
    vim
    htop
    iotop
  ];

  # System optimization for database
  boot.kernel.sysctl = {
    "vm.swappiness" = 10;
    "vm.dirty_ratio" = 15;
    "vm.dirty_background_ratio" = 5;
  };
}
