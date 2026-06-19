// Package extract obtains a provider binary and runs `nivis gen` to capture
// schema-derived Nix constructors plus a normalized schema JSON.
//
// Binary acquisition MUST verify before executing. Reuse the nivis model
// (internal/registry in github.com/wearetechnative/nivis): resolve the version
// via the OpenTofu registry, download, verify the SHA256 checksum, cache — THEN
// run. Either shell out to the nivis CLI for resolution, or mirror its
// verify-before-execute behavior; never execute an unverified binary.
//
// nivis gen contract (verified):
//
//	nivis gen --provider <BINARY_PATH> [--identity <id>] [--out <dir>]
//	  -> emits <out>/<id>/<resource_type>.nix typed constructors with
//	     doc-comment headers (computed outputs, nested-block shapes).
//	  Takes a LOCAL BINARY PATH, not a registry address.
//	  Available via the nivis flake input: nix run github:wearetechnative/nivis#nivis -- gen ...
//
// Batch extraction MUST be resilient: a single provider failure is skipped and
// logged, never hanging or aborting the run.
//
// See openspec/changes/extraction-pipeline/specs/provider-extraction/spec.md.
package extract

// GenArgs builds the `nivis gen` argument vector for a verified binary path.
func GenArgs(binaryPath, identity, outDir string) []string {
	args := []string{"gen", "--provider", binaryPath}
	if identity != "" {
		args = append(args, "--identity", identity)
	}
	if outDir != "" {
		args = append(args, "--out", outDir)
	}
	return args
}

// TODO(extraction-pipeline / provider-extraction epic):
//   - Resolve(address) (binaryPath string, err error): verify-before-execute.
//   - Run(binaryPath, identity, outDir): exec nivis gen; capture .nix + schema JSON.
//   - Batch(seed []string): resilient loop; skip+log failures.
