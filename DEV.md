# Development and testing

Tooling is managed with [mise](https://mise.jdx.dev/); run `mise install`, then `mise tasks` to list
available tasks. `mise run ci` runs the same build/vet/lint/unit-test/docs checks as CI.

## Compatibility checks

| Task / workflow | What it does | Needs a live API? |
| --- | --- | --- |
| `mise run smoke:schema` (`ENGINE=terraform` or `tofu`) | Builds the provider and confirms it serves its schema under the chosen engine via `dev_overrides`. Runs on every PR (`engine-smoke.yml`). | No |
| `mise run testacc` | Full acceptance suite against Terraform. | Yes |
| `mise run testacc:tofu` | The same acceptance suite driven through OpenTofu (`TF_ACC_TERRAFORM_PATH`). | Yes |
| `mise run testpulumi` | Bridges the freshly-built provider through Pulumi and runs a live create/read/destroy round-trip (see [`pulumi-acc/`](pulumi-acc)). | Yes |
| `mise run testacc:all` | Runs the three suites above serially (Terraform → OpenTofu → Pulumi), stopping at the first failure. | Yes |

The three live-API suites are also wired into gated GitHub Actions workflows
(`testacc-terraform.yml`, `testacc-opentofu.yml`, `testpulumi.yml`) that run on manual dispatch and a
weekly schedule. Because they create and destroy **real** resources, they require repository secrets
`PRISMATIC_URL` and `PRISMATIC_REFRESH_TOKEN` (plus `PRISMATIC_TENANT_ID` if multi-tenant) and should
point at a dedicated test tenant.

Locally, append `:prism` (or set `PRISM_TARGET`) to source URL/token/tenant from the
[`prism` CLI](https://prismatic.io/docs/cli/) instead of setting them by hand:

```sh
mise run testacc:all:prism                         # all three, serially
mise run testacc:prism                             # Terraform only
PRISM_TARGET=testacc:tofu mise run testacc:prism   # OpenTofu only
PRISM_TARGET=testpulumi mise run testacc:prism     # Pulumi only
```

The Pulumi run (a `pulumitest` Go test) needs no extra setup — it provides the local state backend and
config passphrase itself.
