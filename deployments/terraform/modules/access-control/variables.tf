variable "members" {
  type = list(object({
    name  = string
    email = string

    jira = optional(object({
      disabled = optional(bool)
    }))

    google_group = optional(
      object({
        disabled = optional(bool, false)
      }),
      { disabled = false },
    )

    slack = optional(
      object({
        disabled       = optional(bool, false)
        email_override = optional(string)
      }),
      {
        disabled = false,
      },
    )

    github = object({
      disabled = optional(bool, false)
      account  = optional(string)
      role     = optional(string)
    })

    squads = list(object({
      name = string
      role = string
    }))

    functions = list(object({
      name         = string
      role         = string
      access_level = optional(string)
    }))

    })
  )
  validation {
    condition = alltrue([
      for member in var.members : alltrue(concat(
        [
          for squad in member.squads : contains([
            "dev",
            "admin",
            "ddd",
            "release",
            "user-management",
            "syllabus",
            "syllabus-lm",
            "syllabus-sp",
            "lesson",
            "calendar",
            "communication",
            "adobo",
            "payment",
            "order-management",
            "course-management",
            "student-course-management",
            "cse",
            "data",
            "architecture",
            "automation",
            "platform",
            "timesheet",
            "internship",
            "virtual-classroom",
            "auth",
            "hermes",
            "adobo-fe",
            "communication-fe",
            "lesson-fe",
            "calendar-fe",
            "payment-fe",
            "syllabus-fe",
            "user-management-fe",
            "timesheet-fe",
          ], squad.name)
        ],
        [
          for function in member.functions : contains([
            "automation",
            "platform",
            "qa",
            "backend",
            "web",
            "mobile",
            "pdm",
            "cse",
            "data",
            "techlead",
          ], function.name)
        ]
        )
      )
    ])
    error_message = "A member with invalid data."
  }
}

variable "project_id" {
  type = string
}

variable "github_token" {
  type = string
}

variable "slack_token" {
  type = string
}

variable "service_ownerships" {
  description = "Describe which squad own which services"

  # Using map would be a more efficient method, but I'm sacrificing it for readability
  type = list(object({
    service = string
    env     = string
    squad   = string
  }))
  default = []
}
