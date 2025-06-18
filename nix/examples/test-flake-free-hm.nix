# Test flake-free home-manager installation
{
  config,
  pkgs,
  ...
}: let
  nixai-src = pkgs.fetchFromGitHub {
    owner = "olafkfreund";
    repo = "nix-ai-help";
    rev = "main";
    sha256 = "0000000000000000000000000000000000000000000000000000"; # Fake hash for testing
  };
  nixai-package = pkgs.callPackage (nixai-src + "/nix/package.nix") {};
in {
  imports = [
    (nixai-src + "/nix/modules/home-manager.nix")
  ];

  services.nixai = {
    enable = true;
    mcp = {
      enable = true;
      package = nixai-package;
      socketPath = "\${config.home.homeDirectory}/.local/share/nixai/mcp.sock";
    };
  };
}
