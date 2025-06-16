# Example home manager configuration for nixai
{
  inputs,
  config,
  lib,
  pkgs,
  ...
}: {
  # Import the nixai home manager module
  imports = [inputs.nixai.homeManagerModules.default];

  # Configure nixai service
  services.nixai = {
    enable = true; # Set to true to enable the service

    mcp = {
      enable = true;
      aiProvider = "copilot"; # or "ollama", "openai", "gemini"
      aiModel = "gpt-4";

      # Optional: Override package if needed
      # package = inputs.nixai.packages.${pkgs.system}.nixai;

      # Optional: Custom socket path
      # socketPath = "$HOME/.local/share/nixai/mcp.sock";

      # Optional: Custom documentation sources
      # documentationSources = [
      #   "https://wiki.nixos.org/wiki/NixOS_Wiki"
      #   "https://nix.dev/"
      #   "https://nixos.org/manual/nixpkgs/stable/"
      # ];
    };
  };

  # Alternatively, you can install nixai as a user package without the service
  # home.packages = with pkgs; [
  #   inputs.nixai.packages.${system}.nixai
  # ];
}
