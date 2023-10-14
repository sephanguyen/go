package io.manabie.quarkus.usermgmt;

import io.smallrye.config.WithName;

public interface IdentityPlatformCredentialConfig {
    @WithName("tenant-id")
    String TenantID();

    @WithName("username")
    String Username();

    @WithName("password")
    String Password();
}
