# this example shows the use of user and group resources
# to control access to a lab.
# Note that these resources will be removed at destroy time!

resource "cml2_user" "student1" {
  username    = "student1"
  password    = "secret"
  fullname    = "Joe Learner"
  email       = "student1@cml.lab"
  description = "This is the Student 1 account"
  is_admin    = false
}

resource "cml2_user" "student2" {
  username    = "student2"
  password    = "secret"
  fullname    = "Jane Learner"
  email       = "student2@cml.lab"
  description = "This is the Student 2 account"
  is_admin    = false
}

resource "cml2_lab" "student_lab" {
  title = "Student Lab"
}

resource "cml2_group" "students" {
  description = "Permission group for all students"
  name        = "students"
  members     = [cml2_user.student1.id, cml2_user.student2.id]
  labs = [
    {
      id         = cml2_lab.student_lab.id
      permission = "read_write"
    },
  ]
}
