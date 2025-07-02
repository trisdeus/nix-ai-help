{
  config,
  pkgs,
  ...
}: {
  services.nginx.enable = true;
  services.nginx.virtualHosts.localhost = {
    root = "/var/www";
    listen = [
      {
        addr = "127.0.0.1";
        port = 80;
      }
    ];
  };
}
