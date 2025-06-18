# Standalone installation for nixai from local source
# Usage: cd nix && nix-build standalone-install.nix && result/bin/nixai --help
let
  pkgs = import <nixpkgs> {};
in
  pkgs.callPackage ./package.nix {
    # Use parent directory as source (the root of the repo)
    srcOverride = ../.;
    version = "latest";
    buildDate = "2025-06-08T00:00:00Z";
  }
