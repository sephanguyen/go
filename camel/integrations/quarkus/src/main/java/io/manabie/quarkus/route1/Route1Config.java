package io.manabie.quarkus.route1;

import io.smallrye.config.WithName;

public interface Route1Config {
    @WithName("message")
    String message();
}
