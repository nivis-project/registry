{
  description = "Nivis Registry — OpenTofu-compatible providers with Nix-native documentation";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    # nivis provides the `nivis` CLI (incl. `nivis gen`), used to extract
    # schema-derived Nix constructors from provider binaries.
    nivis.url = "github:wearetechnative/nivis";
  };

  outputs =
    {
      self,
      nixpkgs,
      nivis,
    }:
    let
      # Plain-nix system enumeration — NO flake-utils. Copied from the nivis flake.
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems =
        f:
        builtins.listToAttrs (
          map (system: {
            name = system;
            value = f system;
          }) systems
        );
      pkgsFor = system: import nixpkgs { inherit system; };

      # The nivis CLI package for a given system (exposes `nivis gen`).
      nivisCliFor = system: nivis.packages.${system}.nivis or nivis.packages.${system}.default;
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

      # The backend generator (tools/seed, tools/extract, tools/compat,
      # tools/generate) builds as a single Go module. No external deps yet, so
      # vendorHash = null; revisit when a dependency is added.
      packages = forAllSystems (
        system:
        let
          pkgs = pkgsFor system;
        in
        {
          default = pkgs.buildGoModule {
            pname = "nivis-registry-tools";
            version = "0.1.0";
            src = ./tools;
            vendorHash = null;
            # Build the whole module (libraries seed/extract/compat/generate +
            # the cmd/* drivers). go test runs in the sandbox; the
            # network-touching paths are gated behind env (NIVIS_SRC) and skip.
            doCheck = true;
          };
        }
      );

      # `nix flake check` runs the Go test suite hermetically (no network) plus a
      # gofmt gate, so the harness is green from day one and stays green.
      checks = forAllSystems (
        system:
        let
          pkgs = pkgsFor system;
        in
        {
          # Compiles tools/ and runs `go test ./...` in the Nix sandbox.
          go-tests = pkgs.buildGoModule {
            pname = "nivis-registry-tools-tests";
            version = "0.1.0";
            src = ./tools;
            vendorHash = null;
            doCheck = true;
          };

          # Fail the check if any tracked Go file is not gofmt-clean.
          gofmt = pkgs.runCommand "gofmt-check" { nativeBuildInputs = [ pkgs.go ]; } ''
            cd ${./tools}
            unformatted=$(gofmt -l .)
            if [ -n "$unformatted" ]; then
              echo "gofmt found unformatted files:" >&2
              echo "$unformatted" >&2
              exit 1
            fi
            echo ok > $out
          '';
        }
      );

      formatter = forAllSystems (system: (pkgsFor system).nixfmt);
    };
}
