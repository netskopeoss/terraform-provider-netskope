## Requirements

No requirements.

## Providers

No providers.

## Modules

| Name | Source | Version |
|------|--------|---------|
| <a name="module_aws-networking"></a> [aws-networking](#module\_aws-networking) | ./modules/aws-networking | n/a |
| <a name="module_ec2"></a> [ec2](#module\_ec2) | ./modules/ec2 | n/a |
| <a name="module_privateapps"></a> [privateapps](#module\_privateapps) | ./modules/privateapps | n/a |
| <a name="module_publishers"></a> [publishers](#module\_publishers) | ./modules/publishers | n/a |
| <a name="module_ssh-key"></a> [ssh-key](#module\_ssh-key) | ./modules/ssh-key | n/a |

## Resources

No resources.

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_project"></a> [project](#input\_project) | The project name; used to name objects. | `string` | `"ns-tf-demo"` | no |
| <a name="input_region"></a> [region](#input\_region) | AWS region | `string` | `"us-east-1"` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_intenral_ec2_connection_string"></a> [intenral\_ec2\_connection\_string](#output\_intenral\_ec2\_connection\_string) | Copy/Paste/Enter - You are in the private ec2 instance |
| <a name="output_pub01_connection_string"></a> [pub01\_connection\_string](#output\_pub01\_connection\_string) | Copy/Paste/Enter - You are in the Publisher01 npa instance |
| <a name="output_pub02_connection_string"></a> [pub02\_connection\_string](#output\_pub02\_connection\_string) | Copy/Paste/Enter - You are in the Publisher01 npa instance |
| <a name="output_publisher01_details"></a> [publisher01\_details](#output\_publisher01\_details) | Name and Token for the Publisher |
| <a name="output_publisher02_details"></a> [publisher02\_details](#output\_publisher02\_details) | Name and Token for the Publisher |
