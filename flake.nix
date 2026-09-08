{
  description = "f2md — Fast codebase to Markdown context packer";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      supportedSystems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
      pkgsFor = system: nixpkgs.legacyPackages.${system};
    in {
      packages = forAllSystems (system: {
        f2md = (pkgsFor system).buildGoModule {
          pname = "f2md";
          version = "0.2.0";
          src = ./.;
          vendorHash = null;
          postInstall = ''
            ln -s $out/bin/f2md $out/bin/folder2md
          '';
        };
        default = self.packages.${system}.f2md;
      });

      devShells = forAllSystems (system: {
        default = (pkgsFor system).mkShell {
          packages = [ (pkgsFor system).go (pkgsFor system).gnumake ];
        };
      });
    };
}
