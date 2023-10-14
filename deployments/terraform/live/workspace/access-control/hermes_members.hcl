locals {
  members = [
    {
      name  = "linhtran-manabie"
      email = "linh.tran@manabie.com"

      github = {
        account = "linhtran-manabie"
        role    = "member"
      }

      squads = [
        {
          name = "communication"
          role = "manager"
        },
        {
          name = "hermes"
          role = "manager"
        },
        {
          name = "syllabus"
          role = "manager"
        },
        {
          name = "syllabus-lm"
          role = "member"
        },
        {
          name = "syllabus-sp"
          role = "member"
        },
      ]

      functions = [
        {
          name = "mobile"
          role = "member"
        },
        {
          name         = "backend"
          role         = "member"
          access_level = "high"
        },
        {
          name = "techlead"
          role = "member"
        },
      ]
    },
    {
      name  = "Nhinlee"
      email = "chinhin.le@manabie.com"

      github = {
        account = "Nhinlee"
        role    = "member"
      }

      squads = [
        {
          name = "communication"
          role = "member"
        },
        {
          name = "hermes"
          role = "member"
        },
      ]

      functions = [
        {
          name = "mobile"
          role = "member"
        },
        {
          name         = "backend"
          role         = "member"
          access_level = "high"
        },
      ]
    },
    {
      name  = "hohieuu"
      email = "vanhieu.ho@manabie.com"

      github = {
        account = "hohieuu"
        role    = "member"
      }

      squads = [
        {
          name = "communication"
          role = "member"
        },
        {
          name = "hermes"
          role = "member"
        },
      ]

      functions = [
        {
          name = "mobile"
          role = "member"
        },
        {
          name         = "backend"
          role         = "member"
          access_level = "high"
        },
      ]
    },
    {
      name  = "Katy Nguyen"
      email = "katy.nguyen@manabie.com"

      github = {
        account = "katycaphe"
        role    = "member"
      }

      squads = [
        {
          name = "hermes"
          role = "member"
        },
      ]

      functions = [
        {
          name         = "pdm"
          role         = "member"
          access_level = "low"
        },
        {
          name         = "data"
          role         = "member"
          access_level = "low"
        },
      ]
    },
  ]
}
