output "postgresql_databases" {
  value = var.postgresql.databases
}

output "postgresql_user_passwords" {
  value = {
    for user, pw in random_password.postgresql_users :
    user => pw.result
  }
  sensitive = true
}
