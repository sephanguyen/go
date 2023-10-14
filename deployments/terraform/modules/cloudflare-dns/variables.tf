variable "dns_config" {
  type = map(object({
    dns_name = string
    dns_ip   = string
    dns_type = string
    proxied  = bool

    certificate_pack = optional(object({
      hosts                  = list(string)

      certificate_authority  = optional(string, "digicert")
      type                   = optional(string, "advanced")
      validation_method      = optional(string, "txt")
      validity_days          = optional(number, 365) # 1 year
      wait_for_active_status = optional(bool, true)
    }))
  }))
}

variable "cloudflare_api_key" {
  type      = string
  sensitive = true
}

variable "cloudflare_email" {
  type    = string
  default = "bao.nguyen@manabie.com"
}


