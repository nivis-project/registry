# Runbook: the public registry site

The site is `registry.nivis.tf`, served by AWS Amplify from `main` of this
repository. The provider and module contract is produced out-of-band and
bundled into each build.

## How it fits together

```
 .github/workflows/contract-publish.yml   (schedule / manual)
   nix develop -> scripts/scale.sh        extract providers + modules, generate
   scripts/verify-contract.sh             refuse an incomplete contract
   aws s3 sync                            documents first, catalogues last
   aws amplify start-job                  make it live
                                            |
 amplify.yml                                v
   scripts/fetch-contract.sh              sync the bucket into frontend/public/registry
     scripts/verify-contract.sh           refuse again, at build time
   pnpm build                             Vite bundles the contract into dist/
```

Two separate clocks: frontend changes go live on push, contract changes go live
on the schedule. Neither can block the other.

## Force a contract refresh

Run the `contract-publish` workflow via **Run workflow** (`workflow_dispatch`).
It extracts, verifies, publishes and triggers the site build. Nothing else is
needed; the deploy is part of the run.

If the run fails, nothing is deployed and the currently live contract keeps
serving. That is deliberate.

## Roll back the site

Amplify keeps deploy history. Redeploy a previous build from the Amplify
console. The contract rolls back with it, because it is part of the artifact
rather than a separate origin.

## A build failed on the contract gate

`scripts/verify-contract.sh` failed, which means the staged contract is absent
or incomplete. It is doing its job: the previous deploy is still serving.

Causes, in order of likelihood:

- the publication workflow has never run, so the bucket is empty
- a publication run died between uploading documents and uploading catalogues
- `CONTRACT_BUCKET` is unset or points at the wrong prefix

Fix by re-running `contract-publish`. The gate's message names what was missing.

Reproduce it locally against any contract directory:

```sh
scripts/verify-contract.sh frontend/public/registry
```

## Configuration

Repository variables (non-secret), used by the workflow:

| variable | meaning |
|--------------------|------------------------------------------|
| `AWS_ROLE_ARN`     | role assumed through GitHub OIDC         |
| `AWS_REGION`       | `eu-central-1`                           |
| `CONTRACT_BUCKET`  | `s3://<bucket>/<prefix>`                 |
| `AMPLIFY_APP_ID`   | the `registry-nivis-tf` app              |
| `AMPLIFY_BRANCH`   | `main`                                   |

Amplify environment variable:

| variable | meaning |
|--------------------|------------------------------------------|
| `CONTRACT_BUCKET`  | the same `s3://<bucket>/<prefix>`        |

The workflow no-ops while `CONTRACT_BUCKET` or `AWS_ROLE_ARN` is unset, so a
scheduled run cannot fail nightly before the account side exists.

There are no long-lived AWS keys. The Amplify GitHub source token is an SSM
SecureString (`/amplify/github_token`) read at apply time by the nivis stack.

## Build locally the way Amplify does

```sh
CONTRACT_BUCKET=s3://<bucket>/<prefix> scripts/fetch-contract.sh
cd frontend && pnpm install --frozen-lockfile && pnpm build
```

Without AWS access, generate a contract instead and skip the fetch:

```sh
nix develop --command scripts/scale.sh      # or proof.sh for a small set
```
