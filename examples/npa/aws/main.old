module "publishers" {
  source    = "./modules/publishers"
  namespace = var.namespace
}

module "networking" {
  source    = "./modules/networking"
  namespace = var.namespace
}

module "ssh-key" {
  source    = "./modules/ssh-key"
  namespace = var.namespace
}

module "ec2" {
  source     = "./modules/ec2"
  namespace  = var.namespace
  vpc        = module.networking.vpc
  sg_pub_id  = module.networking.sg_pub_id
  sg_priv_id = module.networking.sg_priv_id
  key_name   = module.ssh-key.key_name
  token      = module.publishers.pub_token
}

module "privateapps" {
  source      = "./modules/privateapps"
  namespace   = var.namespace
  pub_id      = module.publishers.pub_id
  pub_name    = module.publishers.pub_name
  private_dns = module.ec2.private_dns
}




