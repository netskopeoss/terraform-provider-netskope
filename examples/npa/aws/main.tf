module "publishers" {
  source    = "./modules/publishers"
  project = var.project
}

module "aws-networking" {
  source    = "./modules/aws-networking"
  project = var.project
}

module "ssh-key" {
  source    = "./modules/ssh-key"
  project = var.project
}

module "ec2" {
  source     = "./modules/ec2"
  project        = var.project
  vpc            = module.aws-networking.vpc
  external_sg_id = module.aws-networking.external_sg_id
  internal_sg_id = module.aws-networking.internal_sg_id
  key_name       = module.ssh-key.key_name
  token01        = module.publishers.pub01_token
  token02        = module.publishers.pub02_token
}


module "privateapps" {
  source      = "./modules/privateapps"
  project      = var.project
  pub01_id      = module.publishers.pub01_id
  pub01_name    = module.publishers.pub01_name
  pub02_id      = module.publishers.pub02_id
  pub02_name    = module.publishers.pub02_name
  private_dns   = module.ec2.private_dns
}