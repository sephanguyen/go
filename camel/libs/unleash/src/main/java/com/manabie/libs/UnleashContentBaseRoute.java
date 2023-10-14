package com.manabie.libs;

import org.apache.camel.PropertyInject;

public class UnleashContentBaseRoute {
    @PropertyInject("unleash.env")
    String env;

    @PropertyInject("unleash.org")
    String org;

    @PropertyInject("unleash.url")
    String url;

    @PropertyInject("unleash.token")
    String token;

    @PropertyInject("unleash.service")
    String service;

    public UnleashContentBaseRoute() {
    }

    private ManabieUnleash manabieUnleash;

    private ManabieUnleash getUnleash() {
        if (manabieUnleash == null) {
            manabieUnleash = new ManabieUnleash(service, url, token, env, org);
        }

        return manabieUnleash;
    }

    public boolean isEnabled(String featureName) {
        System.out.println("featureName: " + service + "_" + url + "_" + token + "_" + env + "_" + org);

        ManabieUnleash manabieUnleash = getUnleash();

        return manabieUnleash.isEnabled(featureName);
    }
}
