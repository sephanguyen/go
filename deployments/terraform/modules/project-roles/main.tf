module "custom-roles" {
  source  = "terraform-google-modules/iam/google//modules/custom_role_iam"
  version = "7.3.0"

  count = length(var.roles)

  target_level = "project"
  target_id    = var.project_id

  role_id     = var.roles[count.index].id
  title       = var.roles[count.index].title
  description = var.roles[count.index].description

  base_roles  = var.roles[count.index].base_roles
  permissions = var.roles[count.index].permissions

  members = []
}
