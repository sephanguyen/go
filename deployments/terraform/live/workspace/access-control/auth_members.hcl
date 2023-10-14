locals {
  members = [
    {
      name  = "an-tang"
      email = "hoangan.tang@manabie.com"

      github = {
        account = "an-tang"
        role    = "member"
      }

      squads = [
        {
          name = "auth"
          role = "member"
        },
        {
          name = "user-management"
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
      name  = "sonh"
      email = "phison.huynh@manabie.com"

      github = {
        account = "sonh"
        role    = "member"
      }

      squads = [
        {
          name = "auth"
          role = "member"
        },
        {
          name = "ddd"
          role = "member"
        },
        {
          name = "user-management"
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
      name  = "danhngt"
      email = "thanhdanh.nguyen@manabie.com"

      github = {
        account = "shanenoi"
        role    = "member"
      }

      squads = [
        {
          name = "auth"
          role = "member"
        },
        {
          name = "user-management"
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
      name  = "sondevmanabie"
      email = "truongson.huynh@manabie.com"

      github = {
        account = "sondevmanabie"
        role    = "member"
      }

      squads = [
        {
          name = "auth"
          role = "member"
        },
        {
          name = "user-management"
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
      name  = "DarrenManabie"
      email = "darren.fong@manabie.com"

      github = {
        account = "DarrenManabie"
        role    = "member"
      }

      squads = [
        {
          name = "auth"
          role = "member"
        },
        {
          name = "user-management"
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
