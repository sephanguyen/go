package com.manabie.libs;

import io.getunleash.DefaultUnleash;
import io.getunleash.Unleash;
import io.getunleash.UnleashContext;
import io.getunleash.util.UnleashConfig;

public class ManabieUnleash {
    private Unleash unleash;

    private String env;
    private String org;

    public ManabieUnleash() {
        super();
    }

    public ManabieUnleash(String serviceName, String url, String apiToken, String env, String org) {
        super();

        this.env = env;
        this.org = org;

        UnleashConfig config = UnleashConfig.builder()
                .appName(serviceName)
                .instanceId(serviceName)
                .unleashAPI(url)
                .apiKey(apiToken)
                .synchronousFetchOnInitialisation(true)
                .build();

        unleash = new DefaultUnleash(config, new ManabieEnvUnleashStrategy(), new ManabieOrgUnleashStrategy());
    }

    public Boolean isEnabled(String name) {
        UnleashContext unleashContext = UnleashContext.builder()
                .addProperty("env", env)
                .addProperty("org", org)
                .build();
        Boolean shouldRun = unleash.isEnabled(name, unleashContext);

        return shouldRun;
    }

    public Unleash getUnleash() {
        return unleash;
    }

    public void setUnleash(Unleash unleash) {
        this.unleash = unleash;
    }

    public String getEnv() {
        return env;
    }

    public String getOrg() {
        return org;
    }

    public void setEnv(String env) {
        this.env = env;
    }

    public void setOrg(String org) {
        this.org = org;
    }

}