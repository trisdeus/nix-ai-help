{
  config,
  pkgs,
  ...
}: {
  services.postgresql.enable = true;
  services.postgresql.package = pkgs.postgresql_14;
  services.postgresql.dataDir = "/var/lib/postgresql/data";
}
