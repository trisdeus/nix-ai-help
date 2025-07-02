services.nginx = {
  enable = true;
  virtualHosts."example.com" = {
    enableACME = true;
    forceSSL = true;
    root = "/var/www/html";
  };
};
