# Get a list of all groups on the system or a specific group.
# For a specific group, the length of the list is always one if the
# group is found, zero otherwise.
#
# Group names are enforced to be unique!

data "cml2_groups" "students" {
  name = "students"
}

output "student_group_id" {
  value = data.cml2_groups.students.groups[0].id
}
