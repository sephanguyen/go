include {
  path = find_in_parent_folders()
}

terraform {
  source = "../../..//modules/cloudflare-dns"

}

locals {
  records = read_terragrunt_config("cloudflare-records.hcl").locals
  token   = yamldecode(sops_decrypt_file(find_in_parent_folders("token.enc.yaml")))

}

inputs = {
  dns_config         = local.records.dns_config
  cloudflare_api_key = local.token.cloudflare
}

