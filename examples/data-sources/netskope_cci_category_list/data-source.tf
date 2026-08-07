data "netskope_cci_category_list" "all" {}

locals {
  # Build a name → ID map for portable category references
  cci_category_id = {
    for c in data.netskope_cci_category_list.all.categories :
    c.category_name => c.category_id
  }
}

resource "netskope_custom_category" "genai" {
  name                           = "CAT-GenAI"
  included_predefined_categories = [tostring(local.cci_category_id["Generative AI"])]
}
