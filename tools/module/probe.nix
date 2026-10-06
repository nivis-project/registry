# The structure probe. Applies a nivis module with its configuration POISONED
# (`cfg = throw ...`) and reads only the fields that do not depend on it:
# resource coordinates, data-source coordinates, output names, and any further
# top-level keys a module exposes for composition.
#
# This works because `mkResource` stores provider/type/name as plain fields and
# derives `id` from those coordinates alone, so Nix laziness never forces the
# config. nivis states the property itself in nix/lib/modules.nix: "refAttr only
# needs ids, which are computed from coordinates, not config."
#
# A module that branches its resource LIST on cfg (rather than only its configs)
# forces the throw and fails here. That is intended: the caller records the
# failure rather than presenting a guess.
{
  nivisRef,
  moduleRef,
}:
let
  nivis = (builtins.getFlake nivisRef).lib;
  flake = builtins.getFlake moduleRef;
  known = [
    "resources"
    "dataSources"
    "outputs"
  ];
  coords = r: {
    inherit (r) provider type name;
  };
  probeOne =
    mod:
    let
      m = mod {
        inherit nivis;
        cfg = throw "nivis-registry: module read cfg while probing its structure";
      };
    in
    {
      resources = map coords (m.resources or [ ]);
      dataSources = map coords (m.dataSources or [ ]);
      outputs = builtins.attrNames (m.outputs or { });
      composition = builtins.filter (k: !(builtins.elem k known)) (builtins.attrNames m);
    };
in
{
  src = flake.outPath;
  modules = builtins.mapAttrs (_: probeOne) (flake.nivisModules or { });
}
