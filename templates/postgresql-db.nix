services.postgresql = {
  enable = true;
  package = pkgs.postgresql_15;
  initialDatabases = [
    { name = "myapp"; }
  ];
  authentication = pkgs.lib.mkOverride 10 ''
    local all all trust
    host all all 127.0.0.1/32 trust
  '';
};
