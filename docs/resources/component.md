---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "prismatic_component Resource - terraform-provider-prismatic"
subcategory: ""
description: |-
  Publish a Component to Prismatic. Use the 'Component Bundle' data source to generate the bundle
---

# prismatic_component (Resource)

Publish a Component to Prismatic. Use the 'Component Bundle' data source to generate the bundle



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `bundle_directory` (String) Bundled directory. Reference the results of the 'Component Bundle' data source.
- `bundle_path` (String) Bundle path. Reference the results of the 'Component Bundle' data source.
- `signature` (String) Bundle signature. Reference the results of the 'Component Bundle' data source.

### Read-Only

- `description` (String) The description of the Component
- `id` (String) The ID of the Component
- `key` (String) The key of the Component
- `label` (String) The label of the Component


