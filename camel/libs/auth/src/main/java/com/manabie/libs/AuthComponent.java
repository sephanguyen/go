package com.manabie.libs;

import org.apache.camel.CamelContext;
import org.apache.camel.Endpoint;
import org.apache.camel.support.DefaultComponent;

import java.io.IOException;
import java.util.Map;

@org.apache.camel.spi.annotations.Component("auth")
public class AuthComponent extends DefaultComponent {

    public AuthComponent() {
        this(null);
    }

    public AuthComponent(CamelContext context) {
        super(context);
    }

    @Override
    protected Endpoint createEndpoint(String uri, String remaining, Map<String, Object> parameters) throws Exception {
        Endpoint endpoint = new AuthEndpoint(uri, this, remaining);
        setProperties(endpoint, parameters);
        return endpoint;
    }

    @Override
    public void close() throws IOException {
        super.close();
    }
}
