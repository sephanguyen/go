package com.manabie.libs;

import com.google.gson.Gson;
import com.google.gson.JsonObject;

import com.manabie.libs.proto.bob.v1.UserModifierServiceGrpc;
import com.manabie.libs.proto.bob.v1.Users;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.entity.StringEntity;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClientBuilder;
import org.apache.http.util.EntityUtils;

import java.io.IOException;

public class AuthManager {

    private String googleApiKey;
    private String authServiceAddress;
    private int authServicePort;

    public AuthManager() {
        super();
    }

    public AuthManager(String googleApiKey, String authServiceAddress, int authServicePort) {
        super();

        this.googleApiKey = googleApiKey;
        this.authServiceAddress = authServiceAddress;
        this.authServicePort = authServicePort;
    }

    public String LoginFirebaseWithUserCredential(String tenantId, String username, String password)
            throws IOException {
        JsonObject json = new JsonObject();

        json.addProperty("tenantId", tenantId);
        json.addProperty("email", username);
        json.addProperty("password", password);
        json.addProperty("returnSecureToken", true);

        String url = "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=" + googleApiKey;

        try (CloseableHttpClient httpClient = HttpClientBuilder.create().build()) {
            CloseableHttpResponse response;
            HttpPost request = new HttpPost(url);
            StringEntity params = new StringEntity(json.toString());
            request.addHeader("content-type", "application/json");
            request.setEntity(params);
            response = httpClient.execute(request);
            Gson gson = new Gson();
            JsonObject object = gson.fromJson(EntityUtils.toString(response.getEntity()), JsonObject.class);
            return object.get("idToken").getAsString();
        } catch (Exception ex) {
            return ex.toString();
        }
    }

    public String ExchangeManabieToken(String idToken) {
        ManagedChannel channel = ManagedChannelBuilder.forAddress(authServiceAddress,
                        authServicePort).usePlaintext()
                .build();

        UserModifierServiceGrpc.UserModifierServiceBlockingStub stub = UserModifierServiceGrpc.newBlockingStub(channel);

        Users.ExchangeTokenResponse response = stub
                .exchangeToken(Users.ExchangeTokenRequest.newBuilder().setToken(idToken).build());
        return response.getToken();
    }
}
