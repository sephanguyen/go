package com.manabie.libs;

import org.apache.camel.Exchange;
import org.apache.camel.Message;
import org.apache.camel.support.DefaultProducer;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;

public class AuthProducer extends DefaultProducer {
    private AuthEndpoint endpoint;

    public AuthProducer(AuthEndpoint endpoint) {
        super(endpoint);
        this.endpoint = endpoint;
    }

    @Override
    public void process(Exchange exchange) throws Exception {
        int authServicePort = Integer.parseInt(endpoint.getAuthServicePort());
        AuthManager authManager = new AuthManager(endpoint.getGoogleApiKey(), endpoint.getAuthServiceAddress(), authServicePort);
        String idToken = authManager.LoginFirebaseWithUserCredential(endpoint.getTenantId(), endpoint.getUsername().replace(" ", "+"), endpoint.getPassword());
        String manabieToken = authManager.ExchangeManabieToken(idToken);

        Message messageIn = exchange.getIn();
        messageIn.setHeader("Authorization", "Bearer " + manabieToken);
    }

    @Override
    public void close() throws IOException {
        super.close();
    }
}
