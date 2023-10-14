locals {
  members = [
    {
      name  = "DavidSonNguyen"
      email = "vison.nguyen@manabie.com"

      github = {
        account = "DavidSonNguyen"
        role    = "member"
      }

      squads = [
        {
          name = "virtual-classroom"
          role = "manager"
        },
      ]

      functions = [
        {
          name         = "mobile"
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
      name  = "phukieu93"
      email = "minhphu.kieu@manabie.com"

      github = {
        account = "phukieu93"
        role    = "member"
      }

      squads = [
        {
          name = "virtual-classroom"
          role = "member"
        },
      ]

      functions = [
        {
          name         = "mobile"
          role         = "member"
          access_level = "high"
        },
      ]
    },
    {
      name  = "nguyenhoangvannha"
      email = "vannha.nguyen@manabie.com"

      github = {
        account = "nguyenhoangvannha"
        role    = "member"
      }

      squads = [
        {
          name = "virtual-classroom"
          role = "member"
        },
      ]

      functions = [
        {
          name         = "mobile"
          role         = "member"
          access_level = "high"
        },
      ]
    },
    {
      name  = "trungtkradium"
      email = "trungkim.tran@manabie.com"

      github = {
        account = "trungtkradium"
        role    = "member"
      }

      squads = [
        {
          name = "lesson"
          role = "member"
        },
        {
          name = "virtual-classroom"
          role = "member"
        },
      ]

      functions = [
        {
          name         = "mobile"
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
      name  = "To-NganPham"
      email = "tongan.pham@manabie.com"

      github = {
        account = "To-NganPham"
        role    = "member"
      }

      squads = [
        {
          name = "lesson"
          role = "member"
        },
        {
          name = "virtual-classroom"
          role = "member"
        },
        {
          name = "calendar"
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
