locals {
  members = [
    {
      name  = "truongtuManabie"
      email = "quangtu.truong@manabie.com"

      github = {
        account = "truongtuManabie"
        role    = "member"
      }

      squads = [
        {
          name = "payment"
          role = "member"
        },
        {
          name = "student-course-management"
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
      name  = "Phan Duc Anh"
      email = "ducanh.phan@manabie.com"

      github = {
        account = "AnhPhan49"
        role    = "member"
      }

      squads = [
        {
          name = "payment"
          role = "member"
        },
        {
          name = "student-course-management"
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
      name  = "Nguyen Thanh Tuan"
      email = "thanhtuan.nguyen@manabie.com"

      github = {
        account = "nttuan98"
        role    = "member"
      }

      squads = [
        {
          name = "payment"
          role = "member"
        },
        {
          name = "student-course-management"
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
      name  = "nphdiem"
      email = "hoangdiem.nguyen@manabie.com"

      github = {
        account = "nphdiem"
        role    = "member"
      }

      squads = [
        {
          name = "payment"
          role = "member"
        },
        {
          name = "student-course-management"
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
      name  = "vuongManabie"
      email = "minhvuong.do@manabie.com"

      github = {
        account = "vuongManabie"
        role    = "member"
      }

      squads = [
        {
          name = "payment"
          role = "member"
        },
        {
          name = "student-course-management"
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
  ]
}
