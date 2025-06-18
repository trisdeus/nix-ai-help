# Example nixai configurations for different use cases
# This file shows how to properly configure nixai to avoid the issues mentioned
{
  config,
  pkgs,
  ...
}: let
  # Option 1: Build from fetchFromGitHub (recommended for non-flake users)
  nixai-src = pkgs.fetchFromGitHub {
    owner = "olafkfreund";
    repo = "nix-ai-help";
    rev = "main"; # or use a specific commit for reproducibility
    sha256 = ""; # Leave empty for first build, then add the hash from error message
  };
  nixai-package = pkgs.callPackage (nixai-src + "/nix/package.nix") {};
in {
  # Example Home Manager configuration
  services.nixai = {
    enable = true;
    mcp = {
      enable = true;
      package = nixai-package; # Use the explicitly built package
      socketPath = "\${config.home.homeDirectory}/.local/share/nixai/mcp.sock";
      aiProvider = "ollama";
      aiModel = "llama3";

      # Custom socket path is now supported via --socket-path flag
      extraFlags = [
        "--log-level=info"
      ];
    };

    # Enable VS Code integration
    vscodeIntegration = {
      enable = true;
      contextAware = true;
      autoRefreshContext = true;
    };

    # Enable Neovim integration
    neovimIntegration = {
      enable = true;
      useNixVim = true;
      autoStartMcp = true;
    };
  };
}
# Alternative configurations:
## Option 2: For flake users (recommended)
# In your flake.nix inputs:
# inputs.nixai.url = "github:olafkfreund/nix-ai-help";
#
# Then in your home-manager configuration:
# services.nixai = {
#   enable = true;
#   mcp = {
#     enable = true;
#     package = inputs.nixai.packages.${pkgs.system}.nixai;
#     socketPath = "${config.home.homeDirectory}/.local/share/nixai/mcp.sock";
#   };
# };
## Option 3: System-wide NixOS configuration
# For /etc/nixos/configuration.nix:
# environment.systemPackages = [ nixai-package ];
#
# Or with services (if using the NixOS module):
# services.nixai = {
#   enable = true;
#   mcp = {
#     enable = true;
#     package = nixai-package;
#     socketPath = "/run/nixai/mcp.sock";
#   };
# };

