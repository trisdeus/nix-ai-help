# Example flake.nix for using nixai in your own project
{
  description = "My NixOS configuration with nixai";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    home-manager = {
      url = "github:nix-community/home-manager";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    # Import nixai flake - make sure to use the correct URL
    nixai = {
      url = "github:olafkfreund/nix-ai-help"; # or path:./path/to/nixai if local
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = {
    self,
    nixpkgs,
    home-manager,
    nixai,
    ...
  } @ inputs: {
    # Example for NixOS configuration
    nixosConfigurations.hostname = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      specialArgs = {inherit inputs;};
      modules = [
        # Your existing configuration
        ./configuration.nix

        # Import nixai NixOS module
        nixai.nixosModules.default
      ];
    };

    # Example for Home Manager configuration
    homeConfigurations."username@hostname" = home-manager.lib.homeManagerConfiguration {
      pkgs = nixpkgs.legacyPackages.x86_64-linux;
      extraSpecialArgs = {inherit inputs;};
      modules = [
        # Your existing home-manager configuration
        ./home.nix

        # Import nixai Home Manager module
        nixai.homeManagerModules.default
      ];
    };
  };
}
