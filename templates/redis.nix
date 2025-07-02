{
  config,
  pkgs,
  ...
}: {
  services.redis.enable = true;
  services.redis.port = 6379;
}
