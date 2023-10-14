output "zone_ids" {
  description = "Zone id"
  value       = data.cloudflare_zone.manabie_io.id
}