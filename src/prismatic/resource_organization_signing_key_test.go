package prismatic

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/shurcooL/graphql"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	expectedPubKey = `-----BEGIN RSA PUBLIC KEY-----
MIIBigKCAYEAnBkwVe3rRtng5q3OPgb4FmKEy8259zmJg7EoU1TCKleGoO9Mo/PA
uKya3ln2tcLUxuOlU1UfdeduEC4H6tZXMmfvlDIp5GeuFHY9kkdApWvkp+w6k1PZ
rgzPmq6G3pPvlbekH2+/0wJGZRmLBNtogPfxH9qk1C1RK7gzlTeCmVrhR0V8bBg+
zEG715/luUQiOhcf/x23DgRJB4b/M+G2WoUVQ6b06IHegJswF6+x3YeGCSY3uAni
ayR26LcaEyTRQqtJXJZnKBzAMj83bjVgsHNpMG1skGIK4t/dCJoF0CehfxmIpJ5b
5wJEr5JqtDBBylIp1A8tSp9QzX3mPA5hAlVmoBw+wJrCagN26qF/8VpLZr9w2ij/
7jBFImuIRR2lFMZwOPfLxMH9vaZ2ZEHE2VeQ+n02gvowNISkBq5Oa2AR0opW4T4A
2kyuUjgD3G2U3ne1fPdaGXYDs8XMtV/Ek2LxWu17s7F8/6gvb1JJbXvinE0Y+l9w
Y/35T0CGASyTAgMBAAE=
-----END RSA PUBLIC KEY-----`
)

func resourceWithPubkey(definition string) string {
	return fmt.Sprintf(`
resource "prismatic_organization_signing_key" "key" {
  public_key = <<EOF
%s
EOF
}`, definition)
}

func TestAccResourceOrganizationSigningKey_basic(t *testing.T) {

	resourceName   := "prismatic_organization_signing_key.key"
	baseDefinition := `
		public_key: <<-EOT
		-----BEGIN RSA PUBLIC KEY-----
		MIIBigKCAYEAnBkwVe3rRtng5q3OPgb4FmKEy8259zmJg7EoU1TCKleGoO9Mo/PA
		uKya3ln2tcLUxuOlU1UfdeduEC4H6tZXMmfvlDIp5GeuFHY9kkdApWvkp+w6k1PZ
		rgzPmq6G3pPvlbekH2+/0wJGZRmLBNtogPfxH9qk1C1RK7gzlTeCmVrhR0V8bBg+
		zEG715/luUQiOhcf/x23DgRJB4b/M+G2WoUVQ6b06IHegJswF6+x3YeGCSY3uAni
		ayR26LcaEyTRQqtJXJZnKBzAMj83bjVgsHNpMG1skGIK4t/dCJoF0CehfxmIpJ5b
		5wJEr5JqtDBBylIp1A8tSp9QzX3mPA5hAlVmoBw+wJrCagN26qF/8VpLZr9w2ij/
		7jBFImuIRR2lFMZwOPfLxMH9vaZ2ZEHE2VeQ+n02gvowNISkBq5Oa2AR0opW4T4A
		2kyuUjgD3G2U3ne1fPdaGXYDs8XMtV/Ek2LxWu17s7F8/6gvb1JJbXvinE0Y+l9w
		Y/35T0CGASyTAgMBAAE=
		-----END RSA PUBLIC KEY-----
		EOT`

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckOrganizationSigningKeyResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: resourceWithDefinition(baseDefinition),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "public_key", expectedPubKey),
				),
			},
		},
	})
}

func testAccCheckOrganizationSigningKeyResourceDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*graphql.Client)

	var query struct {
		Organization struct {
			SigningKeys struct {
				Nodes []struct {
					Id        string
					PublicKey string
				}
			}
		}
	}
	if err := client.Query(context.Background(), &query, nil); err != nil {
		return err
	}

	// API does not support filtering on this query so we will instead
	for _, signingKey := range query.Organization.SigningKeys.Nodes {
		if signingKey.PublicKey == expectedPubKey {
			return nil
			break
		}
	}

	return errors.New("found pubkey that should have been deleted")

}
