package com.manabie.libs;

import java.util.Arrays;
import java.util.List;
import java.util.Map;
import java.util.Optional;

import io.getunleash.UnleashContext;
import io.getunleash.strategy.Strategy;

public class ManabieOrgUnleashStrategy implements Strategy {

    protected static final String PARAM = "organizations";

    @Override
    public String getName() {
        return "strategy_organization";
    }

    @Override
    public boolean isEnabled(Map<String, String> parameters, UnleashContext unleashContext) {
        Optional<String> currentEnv = Optional.ofNullable(unleashContext.getProperties().get("org"));
        if (currentEnv.isPresent()) {
            System.out.println(parameters);
            List<String> configuredTenants = Arrays.asList(parameters.get(PARAM).split(",\\s?"));
            return configuredTenants.contains(currentEnv.get());
        } else {
            return false;
        }
    }

    @Override
    public boolean isEnabled(Map<String, String> parameters) {
        throw new UnsupportedOperationException("Unimplemented method 'isEnabled'");
    }
}
