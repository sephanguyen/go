locals {
  members = [
    {
      name  = "anhpngt"
      email = "tuananh.pham@manabie.com"

      github = {
        account = "anhpngt"
        role    = "member"
      }

      squads = [
        {
          name = "platform"
          role = "maintainer"
        },
        {
          name = "release"
          role = "maintainer"
        },
      ]

      functions = [
        {
          name         = "platform"
          role         = "member"
          access_level = "super"
        },
      ]
    },
    {
      name  = "nguyenhoaibao"
      email = "bao.nguyen@manabie.com"

      github = {
        account = "nguyenhoaibao"
        role    = "admin"
      }

      squads = [
        {
          name = "admin"
          role = "maintainer"
        },
        {
          name = "platform"
          role = "member"
        },
        {
          name = "release"
          role = "member"
        },
      ]

      functions = [
        {
          name         = "platform"
          role         = "member"
          access_level = "super"
        },
      ]
    },
    {
      name  = "bangnh1"
      email = "huubang.nguyen@manabie.com"

      github = {
        account = "bangnh1"
        role    = "member"
      }

      squads = [
        {
          name = "platform"
          role = "member"
        },
      ]

      functions = [
        {
          name         = "platform"
          role         = "member"
          access_level = "high"
        },
      ]
    },
    {
      name  = "vctqs1"
      email = "thu.vo@manabie.com"

      github = {
        account = "vctqs1"
        role    = "member"
      }

      squads = [
        {
          name = "platform"
          role = "member"
        },
        {
          name = "automation"
          role = "manager"
        },
        {
          name = "release"
          role = "manager"
        },
      ]

      functions = [
        {
          name = "automation"
          role = "member"
        },
        {
          name         = "web"
          role         = "manager"
          access_level = "high"
        },
        {
          name = "techlead"
          role = "member"
        },
      ]
    },
    {
      name  = "nploi"
      email = "phucloi.nguyen@manabie.com"

      github = {
        account = "nploi"
        role    = "member"
      }

      squads = [
        {
          name = "platform"
          role = "member"
        },
        {
          name = "automation"
          role = "member"
        },
        {
          name = "release"
          role = "member"
        },
      ]

      functions = [
        {
          name = "mobile"
          role = "member"
        },
        {
          name = "automation"
          role = "member"
        },
        {
          name         = "platform"
          role         = "member"
          access_level = "high"
        },
      ]
    },
    {
      name  = "quanghuyhoang-manabie"
      email = "quanghuy.hoang@manabie.com"

      github = {
        account = "quanghuyhoang-manabie"
        role    = "member"
      }

      squads = [
        {
          name = "platform"
          role = "member"
        },
        {
          name = "automation"
          role = "member"
        },
      ]

      functions = [
        {
          name = "automation"
          role = "member"
        },
      ]
    },
    {
      name  = "hoanhvong"
      email = "anhvong.ho@manabie.com"

      github = {
        account = "hoanhvong"
        role    = "member"
      }

      squads = [
        {
          name = "data"
          role = "member"
        },
        {
          name = "platform"
          role = "member"
        },
      ]

      functions = [
        {
          name         = "data"
          role         = "member"
          access_level = "moderate"
        },
        {
          name         = "platform"
          role         = "member"
          access_level = "high"
        },
      ]
    },
    {
      name  = "chivy1204"
      email = "chivy.nguyen@manabie.com"

      github = {
        account = "chivy1204"
        role    = "member"
      }

      squads = [
        {
          name = "platform"
          role = "member"
        },
      ]

      functions = [
        {
          name         = "platform"
          role         = "member"
          access_level = "high"
        },
      ]
    },
  ]
}
