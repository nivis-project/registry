{
  description = "Nivis Registry — OpenTofu-compatible providers with Nix-native documentation";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    # nivis provides the `nivis` CLI (incl. `nivis gen`), used to extract
    # schema-derived Nix constructors from provider binaries.
    nivis.url = "github:wearetechnative/nivis";
  };

  outputs =
    { self, nixpkgs, nivis }:
    let
      # Plain-nix system enumeration — NO flake-utils. Copied from the nivis flake.
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = f: builtins.listToAttrs (map (system: { name = system; value = f system; }) systems);
      pkgsFor = system: import nixpkgs { inherit system; };

      # The nivis CLI package for a given system (exposes `nivis gen`).
      nivisCliFor =
        system:
        nivis.packages.${system}.nivis or nivis.packages.${system}.default;
    in
    {
      # Dev shell: the full toolchain for building the registry.
      devShells = forAllSystems (
        system:
        let
          pkgs = pkgsFor system;
        in
        {
          default = pkgs.mkShell {
            # beans + openspec are provided by the host environment (not in nixpkgs
            # here); the registry build only needs go/node/jj + the nivis CLI.
            packages = [
              pkgs.go
              pkgs.nodejs
              pkgs.pnpm
              pkgs.jujutsu # jj
              (nivisCliFor system)
            ];
            shellHook = ''
              echo "nivis-registry dev shell — go, pnpm, jj, nivis on PATH."
              echo "  nivis gen --help   # schema-derived Nix constructors"
              echo "  beans list --json  # the board (beans/openspec from host)"
            '';
          };
        }
      );

      # Backend-generator packages live here as they are implemented
      # (tools/seed, tools/extract, tools/compat, tools/generate). Stubbed for now.
      packages = forAllSystems (
        system:
        let
          pkgs = pkgsFor system;
        in
        {
          # default = pkgs.buildGoModule { ... };  # TODO: wire tools/generate
        }
      );

      # Keep `nix flake check` meaningful from day one.
      checks = forAllSystems (
        system:
        let
          pkgs = pkgsFor system;
        in
        {
          # Placeholder check; replaced by `go test ./...` wrapper in foundations-scaffold.
          devshell-evaluates = pkgs.runCommand "devshell-evaluates" { } "echo ok > $out";
        }
      );

      formatter = forAllSystems (system: (pkgsFor system).nixfmt-rfc-style);
    };
}
