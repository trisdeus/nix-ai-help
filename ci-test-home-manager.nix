# CI Test Configuration for Home Manager (Non-flake version)
# This file is used in GitHub Actions to test the nixai Home Manager module
{pkgs ? import <nixpkgs> {}}: let
  home-manager = builtins.fetchTarball "https://github.com/nix-community/home-manager/archive/master.tar.gz";
  nixaiPackage = pkgs.callPackage ./nix/package.nix {
    version = "ci-test";
    src = ./.;
    rev = null;
    gitCommit = "ci-test";
    buildDate = "1970-01-01T00:00:00Z";
  };
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
