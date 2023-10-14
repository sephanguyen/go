output "postgresql_read_only_role_name" {
  value = postgresql_role.read_only_role.name
}

output "postgresql_read_write_role_name" {
  value = postgresql_role.read_write_role.name
}

output "postgresql_bypass_rls_role" {
  value = postgresql_role.bypass_rls_role.name
}
