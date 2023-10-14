resource "cloudflare_certificate_pack" "manabieio_certificates" {
  for_each = { for k, v in var.dns_config : k => v if v.proxied && v.certificate_pack != null }

  certificate_authority   = each.value.certificate_pack.certificate_authority
  hosts                   = each.value.certificate_pack.hosts
  type                    = each.value.certificate_pack.type
  validation_method       = each.value.certificate_pack.validation_method
  validity_days           = each.value.certificate_pack.validity_days
  zone_id                 = data.cloudflare_zone.manabie_io.id
  wait_for_active_status = each.value.certificate_pack.wait_for_active_status
}
