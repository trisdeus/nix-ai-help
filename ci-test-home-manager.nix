# CI Test Configuration for Home Manager (Non-flake version)
# This file is used in GitHub Actions to test the nixai Home Manager module
{pkgs ? import <nixpkgs> {}}: let
  home-manager = builtins.fetchTarball "https://github.com/nix-community/home-manager/archive/master.tar.gz";
  # Create a simple mock package for CI testing that doesn't require building Go modules
  nixaiPackage = pkgs.writeShellScriptBin "nixai" ''
    echo "nixai ci-test version"
    exit 0
  '';
in
  import "${home-manager}/modules" {
    pkgs = pkgs;
    configuration = {
      imports = [./nix/modules/home-manager.nix];

      home.username = "ci-test";
      home.homeDirectory = "/home/ci-test";
      home.stateVersion = "25.05";

      # Minimal nixai configuration for CI testing
      services.nixai = {
        enable = true;
        mcp = {
          enable = true;
          package = nixaiPackage;
          aiProvider = "ollama";
          aiModel = "llama3";
          socketPath = "/tmp/nixai-ci-test.sock";
        };

        neovimIntegration = {
          enable = false; # Disable for CI to avoid complexity
        };

        vscodeIntegration = {
          enable = false; # Disable for CI
        };
      };

      # Minimal required Home Manager config
      home.packages = []; # Empty for CI test
    };
  }
