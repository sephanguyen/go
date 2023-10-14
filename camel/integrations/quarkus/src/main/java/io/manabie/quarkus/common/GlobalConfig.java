package io.manabie.quarkus.common;

import io.smallrye.config.WithName;

public interface GlobalConfig {
    @WithName("route-name")
    String RouteName();

    @WithName("bob-address")
    String BobAddress();

    @WithName("usermgmt-address")
    String UsermgmtAddress();
}
