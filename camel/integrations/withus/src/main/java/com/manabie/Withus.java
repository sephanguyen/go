package com.manabie;

// camel-k: trait=pod.enabled=true

// camel-k: dependency=camel:csv

// camel-k: dependency=camel:http
// camel-k: dependency=camel:google-storage
// camel-k: dependency=camel:file
// camel-k: dependency=camel:grpc
// camel-k: service-account=local-camel-k-resource
// camel-k: property=file:../../../resources/application.properties
// camel-k: dependency=file://libs/auth/target/auth-1.0-SNAPSHOT.jar
// camel-k: dependency=file://libs/grpc/target/grpc-1.0-SNAPSHOT.jar
// camel-k: dependency=file://libs/common/target/common-1.0.4.jar

// camel-k: dependency=mvn:io.grpc:grpc-stub:1.56.0
// camel-k: dependency=mvn:io.grpc:grpc-netty-shaded:1.56.0
// camel-k: dependency=mvn:io.grpc:grpc-protobuf:1.56.0
// camel-k: dependency=mvn:org.apache.tomcat:annotations-api:6.0.53
// camel-k: dependency=mvn:org.postgresql:postgresql:42.6.0

import com.google.protobuf.ByteString;
import com.manabie.utils.DomainUser;
import com.manabie.libs.proto.usermgmt.v2.Student;
//import com.manabie.utils.entities.Student;
import com.manabie.utils.entities.Students;
import com.manabie.utils.ManabieStudent;
import io.grpc.*;
import org.apache.camel.PropertyInject;

import java.nio.ByteBuffer;
import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.ResultSet;
import java.sql.Statement;
import java.text.SimpleDateFormat;

import com.manabie.libs.AuthManager;
import org.apache.camel.builder.RouteBuilder;
import org.apache.camel.dataformat.csv.CsvDataFormat;

import java.util.*;

import java.sql.*;
import java.util.Date;
import java.util.stream.Collectors;

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
        String ggEndpointTemplate = "google-storage://{{withusBucket}}?serviceAccountKey={{withusBucketKey}}&objectName=%s&autoCreateBucket=false&operation=getObject";
        return String.format(ggEndpointTemplate, bucketName);
    }

    @PropertyInject("auth-service.address")
    String AuthServiceAddress;
    @PropertyInject("auth-service.port")
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

    @PropertyInject("usermgmt.db.url")
    String UsermgmtPostgresUrl;
    @PropertyInject("usermgmt.db.username")
    String UsermgmtPostgresUsername;
    @PropertyInject("usermgmt.db.password")
    String UsermgmtPostgresPassword;

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

        // onException(Exception.class)
        // .to("direct:saveErrorLog");

        log.info("ggStorageEndpointManagaraBase: " + ggStorageEndpointManagaraBase);

        log.info("user-credential.managara-base.username: " + ManagaraBaseUserCredentialUsername);
        log.info("user-credential.managara-base.password: " + ManagaraBaseUserCredentialPassword);
        log.info("user-credential.managara-hs.username: " + ManagaraHsUserCredentialUsername);
        log.info("user-credential.managara-hs.password: " + ManagaraHsUserCredentialPassword);

        System.out.println("AuthServiceAddress: " + AuthServiceAddress);
        System.out.println("AuthServicePort: " + AuthServicePort);

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
                return new ForwardingClientCall.SimpleForwardingClientCall<ReqT, RespT>(channel.newCall(methodDescriptor, callOptions)) {
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

       /* Class.forName("org.postgresql.Driver");
        Properties props = new Properties();
        props.setProperty("user", UsermgmtPostgresUsername);
        props.setProperty("password", UsermgmtPostgresPassword);
        props.setProperty("ssl", "false");
        Connection conn = DriverManager.getConnection(UsermgmtPostgresUrl, props);*/

        from("timer://trigger-get-data-withus?fixedRate=true&period=6000000")
                .routeId("TimerRoute")
                .to("log:info")
                .to("direct:downloadFile");

        from("direct:downloadFile")
                .routeId("RunManagaraBase")
                .to(ggStorageEndpointManagaraBase)
                .id("googleStorageEndpoint")
                .to("direct:reaFile");

        /*from("direct:reaFile")
                .log("file raw: ${body}")
                .convertBodyTo(String.class, "Shift_JIS") // convert the incoming message to Shift_JIS
                .routeId("ReadManagaraBase")
                .convertBodyTo(String.class, "UTF-8")
                .unmarshal(csvDataFormat)
                .log("file converted to UTF: ${body}")
                .split(body())
                .log("CSV line: ${body}")
                .transform().method("ManabieStudent", "exchange")
                .log("Object line: ${body}")
                .process(exchange -> {

                    Students students = exchange.getMessage().getBody(Students.class);
                    log.info("badfsdf: " + Arrays.toString(students.getStudents().stream().map(Student::getStudentNumber).toArray()));

                    for (int i = 0; i < students.getStudents().size(); i++) {
                        log.info("student: " + students.getStudents().get(i).getStudentNumber() + " " + students.getStudents().get(i).getUserIDAttr());
                    }

                    {
                        PreparedStatement st = conn.prepareStatement("SELECT user_id, user_external_id FROM users WHERE user_external_id = ANY(?) AND deleted_at IS NULL");

                        Array arg = conn.createArrayOf("text", students.getStudents().stream().map(Student::getStudentNumber).toArray());
                        st.setArray(1, arg);
                        ResultSet rs = st.executeQuery();

                        while (rs.next()) {
                            log.info("Column 1 returned ");
                            log.info(rs.getString(1));
                            log.info(rs.getString(2));
                            String userId = rs.getString(1);
                            String userExternalId = rs.getString(2);

                            students.getStudents().stream().filter(student ->
                                    Objects.equals(student.getStudentNumber(), userExternalId)
                            ).forEach(student ->
                                    student.setUserId(userId)
                            );
                        }
                        rs.close();
                        st.close();
                    }

                    {
                        PreparedStatement st = conn.prepareStatement("SELECT grade_id, partner_internal_id FROM grade WHERE partner_internal_id = ANY(?) and deleted_at is NULL");
                        st.setArray(1, conn.createArrayOf("text", students.getStudents().stream().map(Student::getTagG5).toArray()));
                        ResultSet rs = st.executeQuery();

                        while (rs.next()) {
                            log.info("Column 1 returned ");
                            log.info(rs.getString(1));
                            log.info(rs.getString(2));
                            String gradeId = rs.getString(1);
                            String partnerInternalId = rs.getString(2);

                            students.getStudents().stream().filter(student ->
                                    Objects.equals(student.getTagG5(), partnerInternalId)
                            ).forEach(student ->
                                    student.setGradeId(gradeId)
                            );
                        }
                        rs.close();
                        st.close();
                    }

                    {
                        PreparedStatement st = conn.prepareStatement("SELECT location_id, partner_internal_id FROM locations WHERE partner_internal_id = ANY(?) and deleted_at is NULL");
                        st.setArray(1, conn.createArrayOf("text", students.getStudents().stream().map(student -> student.getLocations().split(",")).toArray()));
                        ResultSet rs = st.executeQuery();

                        while (rs.next()) {
                            log.info("Column 1 returned ");
                            log.info(rs.getString(1));
                            log.info(rs.getString(2));
                            String locationId = rs.getString(1);
                            String partnerInternalId = rs.getString(2);

                            students.getStudents().stream().filter(student ->
                                    Arrays.asList(student.getLocations().split(",")).contains(partnerInternalId)
                            ).forEach(student ->
                                    student.addLocationId(locationId)
                            );
                        }
                        rs.close();
                        st.close();
                    }

                    for (Student student : students.getStudents()) {
                        log.info("student: " + student.getUserId() + " " + student.getGradeId() + " " + student.getLocationIds());
                    }
                })
                .marshal().json()
                .log("Json line: ${body}");*/

        // Upload raw csv file to endpoint
        from("direct:reaFile")
                .log("file raw: ${body}")
                .process(exchange -> {
                    System.out.println("body: " + exchange.getMessage().getBody(String.class));
                    Student.ImportWithusManagaraBaseCSVRequest request = Student.ImportWithusManagaraBaseCSVRequest.newBuilder().setPayload(ByteString.copyFrom(exchange.getMessage().getBody(ByteBuffer.class))).build();
                    exchange.getIn().setBody(request);
                })
                .to("grpc://usermgmt.local-manabie-backend.svc.cluster.local:6150/com.manabie.libs.proto.usermgmt.v2.WithusStudentService?method=ImportWithusManagaraBaseCSV")
                .log("Received After call grpc: ${body}");


        /*from("direct:reaFile")
                .log("file raw: ${body}")
                .convertBodyTo(String.class, "Shift_JIS") // convert the incoming message to Shift_JIS
                .routeId("ReadManagaraBase")
                .convertBodyTo(String.class, "UTF-8")
                .unmarshal(csvDataFormat)
                .log("file converted to UTF: ${body}")
                .split(body())
                .log("CSV line: ${body}")
                .transform().method("ManabieStudent", "exchange")
                .log("Object line: ${body}")
                .marshal().json()
                .log("Json line: ${body}")
                .process(exchange -> {
                    exchange.getIn().setHeader("token", manabieToken);
                    Student.ImportWithusManagaraBaseCSVRequest request = Student.ImportWithusManagaraBaseCSVRequest.newBuilder().build();//.setPayload(ByteString.fromHex("asdfsdfs")).build();
                    exchange.getIn().setBody(request);
                })
                .to("grpc://bob.local-manabie-backend.svc.cluster.local:6150/com.manabie.libs.proto.usermgmt.v2.StudentService?consumerStrategy=PROPAGATION&method=ImportWithusManagaraBaseCSV&synchronous=true")
                .log("Received After call grpc ${body}");*/
        // .setHeader(Exchange.HTTP_METHOD,
        // constant(org.apache.camel.component.http.HttpMethods.POST))
        // .to("http://mock-rest-student-import/import-student")
        // .log("sending update request: ${body}");

        from("direct:saveErrorLog")
                .log("info: message error: ${body}");
    }
}
