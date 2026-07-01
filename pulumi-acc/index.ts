import * as prismatic from "@pulumi/prismatic"; // alias emitted by `pulumi package add`; confirmed against the generated SDK

// A fixed public key so the program is deterministic across evaluations. Generating a fresh key each
// run would make every plan re-plan a (RequiresReplace) change on public_key — in Pulumi and in
// Terraform alike — which is not what we want to assert. This key is distinct from the Terraform
// acceptance test's key so the two suites never collide on the organization's global signing-key list.
const publicKey = `-----BEGIN RSA PUBLIC KEY-----
MIIBigKCAYEA14voB1hPCYWJuBqIlruM47Cj4QQ5EI6s0deocqVgQa3hzV+bQe9F
x3+YCwld+T3kA8+hnZ6rYTt5wUWNmMw9zIbbKtN2jWUdacUQp96vklr7+kzk4NrG
LBdvQoSxbcyo+c6MOGykR7PTmuJf53o53qh95QHmUZuBx09dV6OO2M9T3WTjr7lC
qKJNWzXmuB21JujbPKxF+uGsn0PwRZsW/Bzif2JJSjlzYgTgLgxUMdAxl42mnni7
eAtd7hChd8QfrRCpleEGwfFsSLR9I4fsoQM1DdAhbfiuC259L+hWis/DXLZsLvBC
uDUQgQb0/4Gq0D/ldS0T+HOmPTgQblNnaJH4eyCBzkyo/ui08pyPXig2GX0YvC1m
x9QWb9ERF4xUnFOcaFdGtV9rM3n7Q7QnzYnPU6GNvAuz9iY0QZX8F3EEipOJqCHr
ZeY5dR1PFODLsEexAXqi2WEr3OX8nQjeFtY7sa6AxW/a1UYWOWWNwWRVLUoJAny6
vJcZUe/bRJwrAgMBAAE=
-----END RSA PUBLIC KEY-----`;

// Provider auth flows entirely through the PRISMATIC_* env vars the provider already reads
// (internal/provider/provider.go), so no explicit prismatic.Provider or `pulumi config set` is needed.
const key = new prismatic.OrganizationSigningKey("acc", {
  publicKey,
});

export const id = key.id;
export const imported = key.imported;
export const issuedAt = key.issuedAt;
