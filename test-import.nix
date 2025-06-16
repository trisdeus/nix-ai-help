{
  description = "Test nixai import";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    nixai.url = "path:/home/olafkfreund/Source/NIX/nix-ai-help";
  };

  outputs = {
    self,
    nixpkgs,
    nixai,
  }: {
    packages.x86_64-linux.test = nixai.packages.x86_64-linux.nixai;
  };
}
