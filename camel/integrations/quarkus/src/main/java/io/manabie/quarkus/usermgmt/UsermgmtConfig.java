package io.manabie.quarkus.usermgmt;

import java.util.Optional;

import io.smallrye.config.WithName;

public interface UsermgmtConfig {
    @WithName("google-api-key")
    String GoogleAPIKey();

    @WithName("withus-bucket")
    String GoogleStorageWithusBucket();

    @WithName("managara-base")
    IdentityPlatformCredentialConfig ManagaraBase();

    @WithName("managara-hs")
    Optional<IdentityPlatformCredentialConfig> ManagaraHS();
}
