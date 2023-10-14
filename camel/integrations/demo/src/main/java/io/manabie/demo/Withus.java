package io.manabie.demo;

import java.nio.ByteBuffer;
import java.text.SimpleDateFormat;
import java.util.Date;

import org.apache.camel.PropertyInject;
import org.apache.camel.builder.RouteBuilder;
import org.apache.camel.dataformat.csv.CsvDataFormat;

import com.google.protobuf.ByteString;

import io.grpc.CallOptions;
import io.grpc.Channel;
import io.grpc.ClientCall;
import io.grpc.ClientInterceptor;
import io.grpc.ForwardingClientCall;
import io.grpc.Metadata;
import io.grpc.MethodDescriptor;
import io.manabie.demo.auth.AuthManager;
import io.manabie.demo.proto.usermgmt.v2.Student;
import io.manabie.demo.usermgmt.entities.DomainUser;
import io.manabie.demo.usermgmt.entities.ManabieStudent;

// NOTE: Forward port for services if running at localhost
// kubectl -n local-manabie-backend  port-forward service/bob 5050:5050
// kubectl -n local-manabie-backend  port-forward service/usermgmt 6150:6150

public class Withus extends RouteBuilder {

    private String ManagaraBase = "ManagaraBase";

    private String getFileName(String org) {
        Date today = new Date();
        String strToday = new SimpleDateFormat("yyyyMMdd").format(today);
        if (org == ManagaraBase) {
            return String.format("/withus/W2-D6L_users%s.tsv", strToday);
        }
        return String.format("/itee/N1-M1_users%s.tsv", strToday);
    }

    private String getGGStorageEndpoint(String bucketName) {
        String ggEndpointTemplate = "google-storage://{{withusBucket}}?objectName=%s&autoCreateBucket=false&operation=getObject";
        return String.format(ggEndpointTemplate, bucketName);
    }

    @PropertyInject("bob-service.address")
    String AuthServiceAddress;
    @PropertyInject("bob-service.port")
    int AuthServicePort;
    @PropertyInject("gcloud.api-key")
    String GoogleApiKey;
    @PropertyInject("user-credential.managara-base.tenant-id")
    String ManagaraBaseUserCredentialTenantId;
    @PropertyInject("user-credential.managara-base.username")
    String ManagaraBaseUserCredentialUsername;
    @PropertyInject("user-credential.managara-base.password")
    String ManagaraBaseUserCredentialPassword;
    @PropertyInject("user-credential.managara-hs.tenant-id")
    String ManagaraHsUserCredentialTenantId;
    @PropertyInject("user-credential.managara-hs.username")
    String ManagaraHsUserCredentialUsername;
    @PropertyInject("user-credential.managara-hs.password")
    String ManagaraHsUserCredentialPassword;

    // @PropertyInject("usermgmt.db.url")
    // String UsermgmtPostgresUrl;
    // @PropertyInject("usermgmt.db.username")
    // String UsermgmtPostgresUsername;
    // @PropertyInject("usermgmt.db.password")
    // String UsermgmtPostgresPassword;

    @Override
    public void configure() throws Exception {
        log.info("About to start route: http Server -> Log ");
        CsvDataFormat csvDataFormat = new CsvDataFormat();
        char c = '\t';
        csvDataFormat.setDelimiter(c);
        csvDataFormat.setUseMaps(true);

        String ggStorageEndpointManagaraBase = getGGStorageEndpoint(getFileName(this.ManagaraBase));

        getContext().getRegistry().bind("ManabieStudent", ManabieStudent.class);
        getContext().getRegistry().bind("DomainUser", DomainUser.class);

        log.info("ggStorageEndpointManagaraBase: " + ggStorageEndpointManagaraBase);

        log.info("user-credential.managara-base.username: " + ManagaraBaseUserCredentialUsername);
        log.info("user-credential.managara-base.password: " + ManagaraBaseUserCredentialPassword);
        log.info("user-credential.managara-hs.username: " + ManagaraHsUserCredentialUsername);
        log.info("user-credential.managara-hs.password: " + ManagaraHsUserCredentialPassword);
        log.info("AuthServiceAddress: " + AuthServiceAddress);
        log.info("AuthServicePort: " + AuthServicePort);

        AuthManager authManager = new AuthManager(GoogleApiKey, AuthServiceAddress, AuthServicePort);
        String idToken = authManager.LoginFirebaseWithUserCredential(ManagaraBaseUserCredentialTenantId,
                ManagaraBaseUserCredentialUsername, ManagaraBaseUserCredentialPassword);
        log.info("authManager.LoginFirebaseWithUserCredential: " + idToken);
        String manabieToken = authManager.ExchangeManabieToken(idToken);
        log.info("authManager.ExchangeManabieToken: " + manabieToken);

        getContext().getRegistry().bind("grpcMockClientInterceptor", new ClientInterceptor() {
            @Override
            public <ReqT, RespT> ClientCall<ReqT, RespT> interceptCall(
                    MethodDescriptor<ReqT, RespT> methodDescriptor, CallOptions callOptions,
                    Channel channel) {
                return new ForwardingClientCall.SimpleForwardingClientCall<ReqT, RespT>(
                        channel.newCall(methodDescriptor, callOptions)) {
                    @Override
                    public void start(Listener<RespT> responseListener, Metadata headers) {
                        headers.put(Metadata.Key.of("pkg", Metadata.ASCII_STRING_MARSHALLER),
                                "com.manabie.liz");
                        headers.put(Metadata.Key.of("version", Metadata.ASCII_STRING_MARSHALLER),
                                "1.0.0");
                        headers.put(Metadata.Key.of("token", Metadata.ASCII_STRING_MARSHALLER),
                                manabieToken);
                        super.start(responseListener, headers);
                    }
                };
            }
        });

        from("timer://trigger-get-data-withus?fixedRate=true&period=6000000")
                .routeId("TimerRoute")
                .to("log:info")
                .to("direct:downloadFile");

        from("direct:downloadFile")
                .routeId("RunManagaraBase")
                .to(ggStorageEndpointManagaraBase)
                .id("googleStorageEndpoint")
                .to("direct:reaFile");

        // Upload raw csv file to endpoint
        from("direct:reaFile")
                .log("file raw: ${body}")
                .process(exchange -> {
                    System.out.println("body: " + exchange.getMessage().getBody(String.class));
                    Student.ImportWithusManagaraBaseCSVRequest request = io.manabie.demo.proto.usermgmt.v2.Student.ImportWithusManagaraBaseCSVRequest
                            .newBuilder()
                            .setPayload(ByteString.copyFrom(exchange.getMessage().getBody(ByteBuffer.class))).build();
                    exchange.getIn().setBody(request);
                })
                .to("grpc://usermgmt:6150/io.manabie.demo.proto.usermgmt.v2.WithusStudentService?method=ImportWithusManagaraBaseCSV")
                .log("Received After call grpc: ${body}");

        from("direct:saveErrorLog")
                .log("info: message error: ${body}");
    }
}
