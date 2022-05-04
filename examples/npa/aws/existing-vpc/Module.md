## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_netskope"></a> [netskope](#requirement\_netskope) | 0.1.0 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws"></a> [aws](#provider\_aws) | 4.12.1 |
| <a name="provider_netskope"></a> [netskope](#provider\_netskope) | 0.1.0 |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [aws_instance.NpaPublisher](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/instance) | resource |
| netskope_privateapps.PrivateApp | resource |
| netskope_publishers.Publisher | resource |
| [aws_ami.npa-publisher](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/ami) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_aws_instance_type"></a> [aws\_instance\_type](#input\_aws\_instance\_type) | AWS Instance Type | `string` | `"t3.medium"` | no |
| <a name="input_aws_key_name"></a> [aws\_key\_name](#input\_aws\_key\_name) | AWS SSH Key Name | `string` | n/a | yes |
| <a name="input_aws_security_group"></a> [aws\_security\_group](#input\_aws\_security\_group) | AWS SG Id | `string` | n/a | yes |
| <a name="input_aws_subnet"></a> [aws\_subnet](#input\_aws\_subnet) | AWS Subnet Id | `string` | n/a | yes |
| <a name="input_privateapp"></a> [privateapp](#input\_privateapp) | Private Application Name | `string` | `"privateapp-tf"` | no |
| <a name="input_privateapp_host"></a> [privateapp\_host](#input\_privateapp\_host) | Private Application Host | `string` | `"*.ec2.internal"` | no |
| <a name="input_privateapp_tcp"></a> [privateapp\_tcp](#input\_privateapp\_tcp) | TCP Ports | `string` | `"22, 443"` | no |
| <a name="input_privateapp_udp"></a> [privateapp\_udp](#input\_privateapp\_udp) | UDP Ports | `string` | `"123-124"` | no |
| <a name="input_publisher"></a> [publisher](#input\_publisher) | Publisher Name | `string` | `"ns-publisher-tf"` | no |
| <a name="input_region"></a> [region](#input\_region) | AWS region | `string` | `"us-east-1"` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_publisher_details"></a> [publisher\_details](#output\_publisher\_details) | Name/IP and Token for the Publisher |
