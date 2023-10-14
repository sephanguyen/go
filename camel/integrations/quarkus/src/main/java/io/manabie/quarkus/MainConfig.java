package io.manabie.quarkus;

import io.manabie.quarkus.common.GlobalConfig;
import io.manabie.quarkus.route1.Route1Config;
import io.manabie.quarkus.usermgmt.UsermgmtConfig;
import io.smallrye.config.ConfigMapping;
import io.smallrye.config.WithName;


/**
 * This is the main configuration properties for our Java application.
 * 
 * References:
 *  - https://smallrye.io/smallrye-config/Main/config/getting-started/
 */
@ConfigMapping(prefix = "main")
public interface MainConfig {
    @WithName("global")
    GlobalConfig Global();

    @WithName("route1")
    Route1Config Route1Config();

    @WithName("usermgmt")
    UsermgmtConfig Usermgmt();
}
