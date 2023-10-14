locals {
  dns_config = {

    prep-aic = {
      dns_name = "*.prep.aic"
      dns_ip   = "34.146.208.247"
      dns_type = "A"
      proxied  = true

      certificate_pack = {
        hosts = ["manabie.io", "*.prep.aic.manabie.io"]
      }
    },

    prod-aic = {
      dns_name = "*.prod.aic"
      dns_ip   = "34.146.208.247"
      dns_type = "A"
      proxied  = true

      certificate_pack = {
        hosts = ["manabie.io", "*.prod.aic.manabie.io"]
      }
    },

    atlantis = {
      dns_name = "atlantis"
      dns_ip   = "35.247.138.181"
      dns_type = "A"
      proxied  = false
    },

    blog = {
      dns_name = "blog"
      dns_ip   = "manabie-com.github.io."
      dns_type = "CNAME"
      proxied  = false
    },

    prep-ga = {
      dns_name = "*.prep.ga"
      dns_ip   = "34.146.208.247"
      dns_type = "A"
      proxied  = true

      certificate_pack = {
        hosts = ["manabie.io", "*.prep.ga.manabie.io"]
      }
    },

    prod-ga = {
      dns_name = "*.prod.ga"
      dns_ip   = "34.146.208.247"
      dns_type = "A"
      proxied  = true

      certificate_pack = {
        hosts = ["manabie.io", "*.prod.ga.manabie.io"]
      }
    },

    grafana = {
      dns_name = "grafana"
      dns_ip   = "35.247.138.181"
      dns_type = "A"
      proxied  = true
    },

    grafana9 = {
      dns_name = "grafana9"
      dns_ip   = "35.247.138.181"
      dns_type = "A"
      proxied  = false
    },

    grafana-oncall = {
      dns_name = "oncall"
      dns_ip   = "35.247.138.181"
      dns_type = "A"
      proxied  = false
    },

    jp-partners = {
      dns_name = "*.jp-partners"
      dns_ip   = "34.146.208.247"
      dns_type = "A"
      proxied  = false
    },

    prep-jprep = {
      dns_name = "*.prep.jprep"
      dns_ip   = "35.221.127.164"
      dns_type = "A"
      proxied  = true

      certificate_pack = {
        hosts = ["manabie.io", "*.prep.jprep.manabie.io"]
      }
    },

    prod-jprep = {
      dns_name = "*.prod.jprep"
      dns_ip   = "34.146.213.37"
      dns_type = "A"
      proxied  = false

      certificate_pack = {
        hosts = ["manabie.io", "*.prod.jprep.manabie.io"]
      }
    },

    staging-jprep = {
      dns_name = "*.staging.jprep"
      dns_ip   = "34.124.226.59"
      dns_type = "A"
      proxied  = true

      certificate_pack = {
        hosts = ["manabie.io", "*.staging.jprep.manabie.io"]
      }
    },

    uat-jprep = {
      dns_name = "*.uat.jprep"
      dns_ip   = "34.124.226.59"
      dns_type = "A"
      proxied  = true

      certificate_pack = {
        hosts = ["manabie.io", "*.uat.jprep.manabie.io"]
      }
    },

    learner = {
      dns_name = "learner"
      dns_ip   = "34.146.213.37"
      dns_type = "A"
      proxied  = true

      certificate_pack = {
        hosts = ["manabie.io", "learner.manabie.io"]
      }
    },

    stag-manabie-vn = {
      dns_name = "*.stag.manabie-vn"
      dns_ip   = "34.124.226.59"
      dns_type = "A"
      proxied  = false
    },

    stag-green-manabie-vn = {
      dns_name = "*.stag-green.manabie-vn"
      dns_ip   = "34.124.226.59"
      dns_type = "A"
      proxied  = false
    },

    portal = {
      dns_name = "portal"
      dns_ip   = "34.146.213.37"
      dns_type = "A"
      proxied  = true

      certificate_pack = {
        hosts = ["manabie.io", "portal.manabie.io"]
      }
    },

    production = {
      dns_name = "*.production"
      dns_ip   = "35.247.138.181"
      dns_type = "A"
      proxied  = false
    },

    production-blue = {
      dns_name = "*.production-blue"
      dns_ip   = "35.247.138.181"
      dns_type = "A"
      proxied  = false
    },

    production-green = {
      dns_name = "*.production-green"
      dns_ip   = "35.247.138.181"
      dns_type = "A"
      proxied  = false
    },

    production2 = {
      dns_name = "*.production2"
      dns_ip   = "35.247.138.181"
      dns_type = "A"
      proxied  = false
    },

    redash = {
      dns_name = "redash"
      dns_ip   = "35.247.138.181"
      dns_type = "A"
      proxied  = false
    },

    redash-green = {
      dns_name = "redash-green"
      dns_ip   = "35.247.138.181"
      dns_type = "A"
      proxied  = false
    },

    prep-renseikai = {
      dns_name = "*.prep.renseikai"
      dns_ip   = "34.146.208.247"
      dns_type = "A"
      proxied  = true

      certificate_pack = {
        hosts = ["manabie.io", "*.prep.renseikai.manabie.io"]
      }
    },

    prod-renseikai = {
      dns_name = "*.prod.renseikai"
      dns_ip   = "34.146.208.247"
      dns_type = "A"
      proxied  = true

      certificate_pack = {
        hosts = ["manabie.io", "*.prod.renseikai.manabie.io"]
      }
    },

    staging = {
      dns_name = "*.staging"
      dns_ip   = "34.124.226.59"
      dns_type = "A"
      proxied  = true

      certificate_pack = {
        hosts = ["manabie.io", "*.staging.manabie.io"]
      }
    },

    staging-blue = {
      dns_name = "*.staging-blue"
      dns_ip   = "34.124.226.59"
      dns_type = "A"
      proxied  = true

      certificate_pack = {
        hosts = ["manabie.io", "*.staging-blue.manabie.io"]
      }
    },

    staging-green = {
      dns_name = "*.staging-green"
      dns_ip   = "34.124.226.59"
      dns_type = "A"
      proxied  = true

      certificate_pack = {
        hosts = ["manabie.io", "*.staging-green.manabie.io"]
      }
    },

    synersia = {
      dns_name = "*.synersia"
      dns_ip   = "34.146.208.247"
      dns_type = "A"
      proxied  = true

      certificate_pack = {
        hosts = ["manabie.io", "*.synersia.manabie.io"]
      }
    },

    prep-synersia = {
      dns_name = "*.prep.synersia"
      dns_ip   = "34.146.208.247"
      dns_type = "A"
      proxied  = true

      certificate_pack = {
        hosts = ["manabie.io", "*.prep.synersia.manabie.io"]
      }
    },

    prod-synersia = {
      dns_name = "*.prod.synersia"
      dns_ip   = "34.146.208.247"
      dns_type = "A"
      proxied  = true

      certificate_pack = {
        hosts = ["manabie.io", "*.prod.synersia.manabie.io"]
      }
    },

    teacher = {
      dns_name = "teacher"
      dns_ip   = "34.146.213.37"
      dns_type = "A"
      proxied  = true

      certificate_pack = {
        hosts = ["manabie.io", "teacher.manabie.io"]
      }
    },

    prometheus-pushgateway = {
      dns_name = "prometheus-pushgateway"
      dns_ip   = "34.124.226.59"
      dns_type = "A"
      proxied  = false
    }

    thanos = {
      dns_name = "thanos"
      dns_ip   = "35.247.138.181"
      dns_type = "A"
      proxied  = false
    },

    thanos-store = {
      dns_name = "thanos-store"
      dns_ip   = "35.247.138.181"
      dns_type = "A"
      proxied  = false
    },

    prep-tokyo = {
      dns_name = "*.prep.tokyo"
      dns_ip   = "34.146.213.37"
      dns_type = "A"
      proxied  = true

      certificate_pack = {
        hosts = ["manabie.io", "*.prep.tokyo.manabie.io"]
      }
    },

    prod-tokyo = {
      dns_name = "*.prod.tokyo"
      dns_ip   = "34.146.213.37"
      dns_type = "A"
      proxied  = true

      certificate_pack = {
        hosts = ["manabie.io", "*.prod.tokyo.manabie.io"]
      }
    },

    uat = {
      dns_name = "*.uat"
      dns_ip   = "34.124.226.59"
      dns_type = "A"
      proxied  = true

      certificate_pack = {
        hosts = ["manabie.io", "*.uat.manabie.io"]
      }
    },

    vcluster = {
      dns_name = "*.vcluster"
      dns_ip   = "35.247.134.42"
      dns_type = "A"
      proxied  = false
    },

    actions-controller-webhook = {
      dns_name = "actions-controller-webhook"
      dns_ip   = "34.124.226.59"
      dns_type = "A"
      proxied  = false
    },

    sendgrid-cname1 = {
      dns_name = "em3045"
      dns_ip   = "u32246145.wl146.sendgrid.net"
      dns_type = "CNAME"
      proxied  = false
    },

    sendgrid-cname2 = {
      dns_name = "s1._domainkey"
      dns_ip   = "s1.domainkey.u32246145.wl146.sendgrid.net"
      dns_type = "CNAME"
      proxied  = false
    },

    sendgrid-cname3 = {
      dns_name = "s2._domainkey"
      dns_ip   = "s2.domainkey.u32246145.wl146.sendgrid.net"
      dns_type = "CNAME"
      proxied  = false
    },

    actions-runner-metrics = {
      dns_name = "actions-runner-metrics"
      dns_ip   = "34.124.226.59"
      dns_type = "A"
      proxied  = false
    }
  }
}
