data "cloudflare_zone" "manabie_io" {
  account_id = "f5726798a4c3f20b082dc94d5b431341"
  name       = "manabie.io"
}

resource "cloudflare_record" "manabieio_records" {
  for_each = var.dns_config
  name     = each.value.dns_name
  value    = each.value.dns_ip
  type     = each.value.dns_type
  zone_id  = data.cloudflare_zone.manabie_io.id
  proxied  = each.value.proxied
  ttl      = 1
}
