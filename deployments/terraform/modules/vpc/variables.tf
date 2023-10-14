variable "project_id" {
  type = string
}

variable "network_name" {
  type = string
}

variable "subnets" {
  type = list(map(string))
}

variable "secondary_ranges" {
  type = map(list(object({ range_name = string, ip_cidr_range = string })))
}

variable "auto_create_subnetworks" {
  type = bool
}
