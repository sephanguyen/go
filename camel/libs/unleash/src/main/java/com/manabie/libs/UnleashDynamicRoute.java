package com.manabie.libs;

import java.util.Map;

import org.apache.camel.ExchangeProperties;
import org.apache.camel.Handler;
import org.apache.camel.Header;
import org.apache.camel.PropertyInject;

public class UnleashDynamicRoute {
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

    public UnleashDynamicRoute() {
    }

    private int incr = 0;

    public String isEnabled(String featureName) {
        System.out.println("featureName: " + featureName + " " + incr);

        ManabieUnleash manabieUnleash = new ManabieUnleash(service, url, token, env,
                org);
        boolean isEnabled = manabieUnleash.isEnabled(featureName);

        System.out.println("isEnabled: " + isEnabled);
        if (isEnabled && incr == 0) {
            incr++;
            return "direct:unleashEnabledSte1";
        } else if (isEnabled && incr == 1) {
            incr++;
            return "direct:unleashEnabledSte2";
        }
        return null;
    }

}
