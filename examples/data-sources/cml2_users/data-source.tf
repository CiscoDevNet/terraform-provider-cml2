# Get a list of all users on the system or a specific user.
# For a specific user, the length of the list is always one if the
# user is found, zero otherwise.
#
# User names are enforced to be unique!

data "cml2_users" "get_admin" {
  username = "admin"
}

output "admin_id" {
  value = data.cml2_users.get_admin.users[0].id
}
