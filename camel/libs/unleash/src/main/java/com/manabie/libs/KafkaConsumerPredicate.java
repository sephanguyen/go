
package com.manabie.libs;

import java.util.function.Predicate;

import org.apache.camel.PropertyInject;

public class KafkaConsumerPredicate implements Predicate {

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

    public KafkaConsumerPredicate() {
    }

    public boolean test(java.lang.Object arg0) {
        System.out.println("featureName: " + service + "_" + url + "_" + token + "_" + env + "_" + org);

        ManabieUnleash manabieUnleash = new ManabieUnleash(service, url, token, env, org);

        return manabieUnleash.isEnabled("Architecture_BACKEND_MasterData_Course_TeachingMethod");
    }
}
