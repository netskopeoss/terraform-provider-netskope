data "netskope_rbac_role_config" "this" {}

locals {
  # Build a name -> tenant-specific ID map for all API groups.
  # API group IDs are tenant-specific and cannot be hardcoded portably.
  api_group_id = {
    for g in data.netskope_rbac_role_config.this.api_groups : g.name => g.api_group_id
  }

  # Build a map of api_group name -> list of obfuscatable field names, excluding "app"
  # (keeps "App names, URLs, and destination IPs" readable; obfuscates user, source, userip, file)
  obfuscation_fields = {
    for g in data.netskope_rbac_role_config.this.api_groups :
    g.name => [
      for o in data.netskope_rbac_role_config.this.obfuscation :
      o.obfuscation_name if o.api_group_id == g.api_group_id && o.obfuscation_name != "app"
    ]
  }
}

resource "netskope_rbac_role" "helpdesk" {
  name        = "helpdesk-readonly"
  description = "Read-only helpdesk role with privacy obfuscation"

  api_groups = [{
    api_group_id = local.api_group_id["events_alerts"]
    permission   = "r"
    obfuscation = {
      properties = local.obfuscation_fields["events_alerts"]
    }
  }]
}
