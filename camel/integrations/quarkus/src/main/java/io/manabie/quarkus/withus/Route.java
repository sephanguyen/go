package io.manabie.quarkus.withus;

import static org.apache.camel.builder.endpoint.StaticEndpointBuilders.googleStorage;

import java.io.IOException;
import java.nio.ByteBuffer;
import java.text.SimpleDateFormat;
import java.util.Date;
import java.util.HashMap;

import org.apache.camel.Exchange;
import org.apache.camel.builder.RouteBuilder;
import org.apache.camel.component.google.storage.GoogleCloudStorageOperations;
import org.apache.http.HttpEntity;
import org.apache.http.ParseException;
import org.apache.http.util.EntityUtils;

import com.google.gson.Gson;
import com.google.gson.JsonObject;
import com.google.protobuf.ByteString;

import io.grpc.CallOptions;
import io.grpc.Channel;
import io.grpc.ClientCall;
import io.grpc.ClientInterceptor;
import io.grpc.ForwardingClientCall;
import io.grpc.Metadata;
import io.grpc.MethodDescriptor;
import io.manabie.quarkus.common.GlobalConfig;
import io.manabie.quarkus.proto.bob.v1.Users;
import io.manabie.quarkus.proto.usermgmt.v2.Student;
import io.manabie.quarkus.usermgmt.UsermgmtConfig;

public class Route extends RouteBuilder {
	static private String managaraBase = "ManagaraBase";

	private String ggsWithusBucket;
	private String googleApiKey;
	private String managaraBaseTenantId;
	private String managaraBaseUsername;
	private String managaraBasePassword;
	private String bobGRPCAddress;
	private String usermgmtGRPCAddress;

	static private String _manabiePkgName = "com.manabie.liz";
	static private String _manabiePkgVersion = "1.0.0";
	static private String _manabieToken = ""; // will be updated at runtime after we fetch token

	public Route(GlobalConfig gc, UsermgmtConfig c) {
		this.ggsWithusBucket = c.GoogleStorageWithusBucket();
		this.googleApiKey = c.GoogleAPIKey();
		this.managaraBaseTenantId = c.ManagaraBase().TenantID();
		this.managaraBaseUsername = c.ManagaraBase().Username();
		this.managaraBasePassword = c.ManagaraBase().Password();

		this.bobGRPCAddress = gc.BobAddress();
		this.usermgmtGRPCAddress = gc.UsermgmtAddress();
	}

	@Override
	public void configure() throws Exception {
		// Register outbounnd GRPC interceptor for authentication
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
								Route._manabiePkgName);
						headers.put(Metadata.Key.of("version", Metadata.ASCII_STRING_MARSHALLER),
								Route._manabiePkgVersion);
						headers.put(Metadata.Key.of("token", Metadata.ASCII_STRING_MARSHALLER),
								Route._manabieToken);
						super.start(responseListener, headers);
					}
				};
			}
		});

		// Route entrypoint
		from("timer:entrypoint?fixedRate=true&period=600000").to("direct:getGoogleToken");

		from("direct:getGoogleToken").id("getGoogleToken")
				.process(this::setIdentityToolkitRequestPayload)
				.process(this::setIdentityToolkitRequestHeaders)
				.to("https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword")
				.process(this::transformToExchangeManabieTokenRequest)
				.to("direct:exchangeManabieToken");

		from("direct:exchangeManabieToken").id("exchangeManabieToken")
				.log("exchaging manabie token: ${body}").process(exchange -> {
					exchange.getMessage().setBody(Users.ExchangeTokenRequest.newBuilder()
							.setToken(exchange.getIn().getBody(String.class)).build());
				})
				.to(String.format("grpc://%s/io.manabie.quarkus.proto.bob.v1.UserModifierService"
						+ "?method=ExchangeToken" + "&synchronous=true", this.bobGRPCAddress))
				.process(this::setupAuthInterceptorForOutboundGRPC).to("direct:downloadFile");

		from("direct:downloadFile").id("downloadFile")
				.to(googleStorage(this.ggsWithusBucket).objectName(getFileName(Route.managaraBase))
						.autoCreateBucket(false).operation(GoogleCloudStorageOperations.getObject))
				.to("direct:readFile");

		from("direct:readFile").id("readFile").process(exchange -> {
			Student.ImportWithusManagaraBaseCSVRequest request =
					Student.ImportWithusManagaraBaseCSVRequest.newBuilder()
							.setPayload(ByteString
									.copyFrom(exchange.getMessage().getBody(ByteBuffer.class)))
							.build();
			exchange.getIn().setBody(request);
		}).to(String.format(
				"grpc:%s/io.manabie.quarkus.proto.usermgmt.v2.WithusStudentService?method=ImportWithusManagaraBaseCSV",
				this.usermgmtGRPCAddress)).log("Received after call grpc: ${body}");
	}

	private String getFileName(String org) {
		Date today = new Date();
		String strToday = new SimpleDateFormat("yyyyMMdd").format(today);
		if (org == Route.managaraBase) {
			return String.format("/withus/W2-D6L_users%s.tsv", strToday);
		}
		return String.format("/itee/N1-M1_users%s.tsv", strToday);
	}

	private void setIdentityToolkitRequestPayload(Exchange exchange) {
		Gson body = new Gson();
		HashMap<String, String> data = new HashMap<String, String>();
		data.put("tenantId", this.managaraBaseTenantId);
		data.put("email", this.managaraBaseUsername);
		data.put("password", this.managaraBasePassword);
		data.put("returnSecureToken", "true");
		exchange.getMessage().setBody(body.toJson(data));
	}

	private void transformToExchangeManabieTokenRequest(Exchange exchange)
			throws IOException, ParseException {
		Gson gson = new Gson();
		JsonObject data = gson.fromJson(
				EntityUtils.toString(exchange.getIn().getBody(HttpEntity.class)), JsonObject.class);
		String googleToken = data.get("idToken").getAsString();
		exchange.getMessage().setBody(googleToken);
	}

	private void setIdentityToolkitRequestHeaders(Exchange exchange) {
		exchange.getIn().setHeader(Exchange.HTTP_QUERY, String.format("key=%s", this.googleApiKey));
		exchange.getIn().setHeader(Exchange.HTTP_METHOD,
				constant(org.apache.camel.component.http.HttpMethods.POST));
		exchange.getIn().setHeader(Exchange.CONTENT_TYPE, constant("application/json"));
	}

	private void setupAuthInterceptorForOutboundGRPC(Exchange exchange) {
		Route._manabieToken =
				exchange.getIn().getBody(Users.ExchangeTokenResponse.class).getToken();
	}
}
