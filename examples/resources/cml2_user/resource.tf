# Create a user account

resource "cml2_user" "student1" {
  username    = "student1"
  password    = "secret"
  fullname    = "Joe Learner"
  email       = "student1@cml.lab"
  description = "This is the Student 1 account"
  is_admin    = false
}
