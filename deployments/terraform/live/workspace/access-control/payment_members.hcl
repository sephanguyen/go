locals {
  members = [
    {
      name  = "noritaketakamichi"
      email = "noritake.takamichi@manabie.com"

      jira = {
        disabled = true
      }

      github = {
        account = "noritaketakamichi"
        role    = "member"
      }

      squads = [
        {
          name = "payment"
          role = "member"
        },
        {
          name = "order-management"
          role = "member"
        },
        {
          name = "course-management"
          role = "member"
        },
      ]

      functions = [
        {
          name = "pdm"
          role = "member"
        },
      ]
    },
    {
      name  = "julytacda"
      email = "july.tacda@manabie.com"

      jira = {
        disabled = true
      }

      github = {
        account = "jjtacda"
        role    = "member"
      }

      squads = [
        {
          name = "payment"
          role = "member"
        },
        {
          name = "order-management"
          role = "member"
        },
        {
          name = "course-management"
          role = "member"
        },
        {
          name = "adobo"
          role = "member"
        },
      ]

      functions = [
        {
          name = "pdm"
          role = "member"
        },
      ]
    },
    {
      name  = "quangtienpham"
      email = "quangtien.pham@manabie.com"

      github = {
        account = "quangtienpham"
        role    = "member"
      }

      squads = [
        {
          name = "payment"
          role = "manager"
        },
        {
          name = "order-management"
          role = "manager"
        },
        {
          name = "course-management"
          role = "manager"
        },
        {
          name = "adobo"
          role = "manager"
        },
        {
          name = "ddd"
          role = "member"
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
      name  = "huutrungtruong"
      email = "huutrung.truong@manabie.com"

      github = {
        account = "huutrungtruong"
        role    = "member"
      }

      squads = [
        {
          name = "payment"
          role = "member"
        },
        {
          name = "order-management"
          role = "member"
        },
        {
          name = "course-management"
          role = "member"
        },
      ]

      functions = [
        {
          name         = "qa"
          role         = "member"
          access_level = "high"
        },
        {
          name         = "data"
          role         = "member"
          access_level = "high"
        },
      ]
    },
  ]
}
