resource "terraform_scim_user" "my_scimuser" {
  active      = true
  external_id = "User-Ext_id"
  user_name   = "upn1"
}