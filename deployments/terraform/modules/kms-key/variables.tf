variable "project_id" {
  type = string
}

variable "key_ring" {
  type = object({
    name     = string
    location = string
  })
}

variable "kms_keys" {
  type = map(object({
    rotation_period     = string
    owner               = string
    encrypter_decrypter = string
    encrypters          = list(string)
    decrypters          = list(string)
  }))

  default = {}
}

variable "create_google_groups" {
  type        = bool
  description = "Whether to create Google Groups. Useful for GCP projects that share Google Groups with other projects."
}
