locals {
  members = [
    {
      name  = "QuangMinh-Tran"
      email = "quangminh.tran@manabie.com"

      github = {
        account = "QuangMinh-Tran"
        role    = "member"
      }

      squads = [
        {
          name = "calendar"
          role = "manager"
        },
      ]

      functions = [
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
      name  = "7inh"
      email = "quoclinh.tran@manabie.com"

      github = {
        account = "7inh"
        role    = "member"
      }

      squads = [
        {
          name = "calendar"
          role = "member"
        },
      ]

      functions = [
        {
          name         = "web"
          role         = "member"
          access_level = "high"
        },
        {
          name         = "backend"
          role         = "member"
          access_level = "high"
        },
      ]
    },
    {
      name  = "qgdomingo"
      email = "gio.domingo@manabie.com"

      github = {
        account = "qgdomingo"
        role    = "member"
      }

      squads = [
        {
          name = "calendar"
          role = "member"
        },
      ]

      functions = [
        {
          name         = "backend"
          role         = "member"
          access_level = "high"
        },
      ]
    },
    {
      name  = "hoangnguyen147"
      email = "huyhoang.nguyen@manabie.com"

      github = {
        account = "hoangnguyen147"
        role    = "member"
      }

      squads = [
        {
          name = "calendar"
          role = "member"
        },
        {
          name = "lesson"
          role = "member"
        },
      ]

      functions = [
        {
          name         = "web"
          role         = "member"
          access_level = "high"
        },
        {
          name         = "backend"
          role         = "member"
          access_level = "high"
        },
      ]
    },
  ]
}
