# Test your nixai flake import
# Run this with: nix eval -f test-nixai-import.nix version
let
  flake = builtins.getFlake "github:olafkfreund/nix-ai-help";
in rec {
  # Test that the package builds and has correct version
  package = flake.packages.x86_64-linux.nixai;

  # Test that modules are available
  nixosModule = flake.nixosModules.default;
  homeManagerModule = flake.homeManagerModules.default;

  # Extract version for verification
  version = package.version or "unknown";
}
