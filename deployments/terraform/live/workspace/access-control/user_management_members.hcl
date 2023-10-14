locals {
  members = [
    {
      name  = "crissang"
      email = "xuansang.pham@manabie.com"

      github = {
        account = "crissang"
        role    = "member"
      }

      squads = [
        {
          name = "user-management"
          role = "member"
        },
        {
          name = "auth"
          role = "member"
        },
      ]

      functions = [
        {
          name         = "web"
          role         = "member"
          access_level = "high"
        },
      ]
    },
    {
      name  = "hoangtule"
      email = "hoangtu.le@manabie.com"

      github = {
        account = "hoangtule"
        role    = "member"
      }

      squads = [
        {
          name = "user-management"
          role = "member"
        },
      ]

      functions = [
        {
          name         = "web"
          role         = "member"
          access_level = "low"
        },
      ]
    },
    {
      name  = "minhthao56"
      email = "minhthao.nguyen@manabie.com"

      github = {
        account = "minhthao56"
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
      name  = "phucgaoxam"
      email = "phuc.chau@manabie.com"

      github = {
        account = "phucgaoxam"
        role    = "member"
      }

      squads = [
        {
          name = "auth"
          role = "manager"
        },
        {
          name = "user-management"
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
      name  = "nohattee-manabie"
      email = "huutuan.nguyen@manabie.com"

      github = {
        account = "nohattee-manabie"
        role    = "member"
      }

      squads = [
        {
          name = "user-management"
          role = "member"
        },
        {
          name = "auth"
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
    # {
    #   name  = "KhanhLe91"
    #   email = "vankhanh.le@manabie.com"
    #
    #   github = {
    #     account = "KhanhLe91"
    #     role    = "member"
    #   }
    #
    #   squads = [
    #     {
    #       name = "auth"
    #       role = "member"
    #     },
    #     {
    #       name = "user-management"
    #       role = "member"
    #     },
    #   ]
    #
    #   functions = [
    #     {
    #       name         = "qa"
    #       role         = "member"
    #       access_level = "high"
    #     },
    #     {
    #       name         = "data"
    #       role         = "member"
    #       access_level = "high"
    #     },
    #   ]
    # },
    {
      name  = "dvbtrung2302"
      email = "baotrung.vo@manabie.com"

      github = {
        account = "dvbtrung2302"
        role    = "member"
      }

      squads = [
        {
          name = "user-management"
          role = "member"
        },
        {
          name = "auth"
          role = "member"
        },
      ]

      functions = [
        {
          name         = "web"
          role         = "member"
          access_level = "high"
        },
      ]
    },
    {
      name  = "alexander-manabie"
      email = "alexander.teo@manabie.com"

      github = {
        account = "alexander-manabie"
        role    = "member"
      }

      squads = [
        {
          name = "user-management"
          role = "member"
        },
        {
          name = "auth"
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
    {
      name  = "QuachTrungHieu"
      email = "trunghieu.quach@manabie.com"

      github = {
        account = "QuachTrungHieu"
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
