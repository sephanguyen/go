locals {
  // The mappinng port of Postgresql instances that the sql proxy set to listen on.
  // Whenever a new instance is deployed, its port should be added to this map.
  postgresql_instance_port = {
    manabie-59fd            = "5432"
    jprep-uat               = "5433"
    synersia-228d           = "5434"
    renseikai-83fc          = "5435"
    jprep-6a98              = "5436"
    manabie-2db8            = "5437"
    jp-partners-b04fbb69    = "5438"
    prod-tokyo              = "5439"
    analytics               = "5440"
    manabie-lms-de12e08e    = "5441"
    manabie-common-88e1ee71 = "5442"
    prod-tokyo-lms-b2dc4508 = "5443"
    prod-jprep-d995522c     = "5444"

    prod-tokyo-data-warehouse-251f01f8 = "5445"
    manabie-auth-f2dc7988              = "5446"
    prod-tokyo-auth-42c5a298           = "5447"
  }
}
