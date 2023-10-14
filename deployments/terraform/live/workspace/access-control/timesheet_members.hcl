locals {
  members = [
    {
      name  = "Thmtran"
      email = "manhtung.tran@manabie.com"

      github = {
        account = "Thmtran"
        role    = "member"
      }

      squads = [
        {
          name = "timesheet"
          role = "member"
        },
        {
          name = "hermes"
          role = "member"
        },
      ]

      functions = [
        {
          name = "qa"
          role = "member"
          access_level = "high"
        },
        {
          name         = "data"
          role         = "member"
          access_level = "high"
        },
      ]
    },
    {
      name  = "tuanduongnguyen209"
      email = "tuanduong.nguyen@manabie.com"

      github = {
        account = "tuanduongnguyen209"
        role    = "member"
      }

      squads = [
        {
          name = "timesheet"
          role = "member"
        },
        {
          name = "hermes"
          role = "member"
        },
      ]

      functions = [
        {
          name = "web"
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
      name  = "samceracas-manabie"
      email = "ezequielsam.ceracas@manabie.com"

      github = {
        account = "samceracas-manabie"
        role    = "member"
      }

      squads = [
        {
          name = "timesheet"
          role = "member"
        },
      ]

      functions = [
        {
          name = "web"
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
      name  = "danh996"
      email = "congdanh.le@manabie.com"

      github = {
        account = "danh996"
        role    = "member"
      }

      squads = [
        {
          name = "timesheet"
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
      name  = "rbmrclo"
      email = "robbie.marcelo@manabie.com"

      github = {
        account = "rbmrclo"
        role    = "member"
      }

      squads = [
        {
          name = "timesheet"
          role = "manager"
        },
        {
          name = "syllabus"
          role = "manager"
        },
      ]

      functions = [
        {
          name = "backend"
          role = "member"
        },
      ]
    },
  ]
}
