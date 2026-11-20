#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

tools="$(nix build --no-link --print-out-paths --impure --expr '
  let
    pkgs = import (builtins.getFlake "path:'"$ROOT"'").inputs.nixpkgs { system = builtins.currentSystem; };
    rev = "5174b3333331c966c38f4355d50b03ca1c1df2f9";
    files = {
      "Hind-Regular.ttf" = "sha256-Ad4VgCL1MHe1IwPkbeOwq1+yRSIqf/4loqV/3Z6WkWI=";
      "Hind-Medium.ttf" = "sha256-cPnBFYOQ0Mk6i8JLNbS+SyQ5ZTNyu85zgaMsaG+5Gpg=";
      "Hind-SemiBold.ttf" = "sha256-F0KhsKupe0naHT7jr1aPqZfFmjz+c4Kfe+wpP0FtJd0=";
    };
    fonts = pkgs.linkFarm "hind" (pkgs.lib.mapAttrsToList (name: hash: {
      inherit name;
      path = pkgs.fetchurl { url = "https://raw.githubusercontent.com/google/fonts/${rev}/ofl/hind/${name}"; inherit hash; };
    }) files);
    python = pkgs.python3.withPackages (ps: [ ps.coloraide ]);
  in pkgs.linkFarm "render-social-tools" [
    { name = "fonts"; path = fonts; }
    { name = "python"; path = python; }
    { name = "resvg"; path = pkgs.resvg; }
  ]')"

PATH="$tools/resvg/bin:$PATH" "$tools/python/bin/python3" scripts/render-social.py \
  frontend/social.svg frontend/public/social.png "$tools/fonts"
