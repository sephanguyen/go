package io.manabie.demo.proto.usermgmt.v2;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.56.1)",
    comments = "Source: usermgmt/v2/auth.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class AuthServiceGrpc {

  private AuthServiceGrpc() {}

  public static final String SERVICE_NAME = "usermgmt.v2.AuthService";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Auth.ExchangeCustomTokenRequest,
      io.manabie.demo.proto.usermgmt.v2.Auth.ExchangeCustomTokenResponse> getExchangeCustomTokenMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ExchangeCustomToken",
      requestType = io.manabie.demo.proto.usermgmt.v2.Auth.ExchangeCustomTokenRequest.class,
      responseType = io.manabie.demo.proto.usermgmt.v2.Auth.ExchangeCustomTokenResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Auth.ExchangeCustomTokenRequest,
      io.manabie.demo.proto.usermgmt.v2.Auth.ExchangeCustomTokenResponse> getExchangeCustomTokenMethod() {
    io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Auth.ExchangeCustomTokenRequest, io.manabie.demo.proto.usermgmt.v2.Auth.ExchangeCustomTokenResponse> getExchangeCustomTokenMethod;
    if ((getExchangeCustomTokenMethod = AuthServiceGrpc.getExchangeCustomTokenMethod) == null) {
      synchronized (AuthServiceGrpc.class) {
        if ((getExchangeCustomTokenMethod = AuthServiceGrpc.getExchangeCustomTokenMethod) == null) {
          AuthServiceGrpc.getExchangeCustomTokenMethod = getExchangeCustomTokenMethod =
              io.grpc.MethodDescriptor.<io.manabie.demo.proto.usermgmt.v2.Auth.ExchangeCustomTokenRequest, io.manabie.demo.proto.usermgmt.v2.Auth.ExchangeCustomTokenResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ExchangeCustomToken"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.demo.proto.usermgmt.v2.Auth.ExchangeCustomTokenRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.demo.proto.usermgmt.v2.Auth.ExchangeCustomTokenResponse.getDefaultInstance()))
              .setSchemaDescriptor(new AuthServiceMethodDescriptorSupplier("ExchangeCustomToken"))
              .build();
        }
      }
    }
    return getExchangeCustomTokenMethod;
  }

  private static volatile io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Auth.ValidateUserIPRequest,
      io.manabie.demo.proto.usermgmt.v2.Auth.ValidateUserIPResponse> getValidateUserIPMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ValidateUserIP",
      requestType = io.manabie.demo.proto.usermgmt.v2.Auth.ValidateUserIPRequest.class,
      responseType = io.manabie.demo.proto.usermgmt.v2.Auth.ValidateUserIPResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Auth.ValidateUserIPRequest,
      io.manabie.demo.proto.usermgmt.v2.Auth.ValidateUserIPResponse> getValidateUserIPMethod() {
    io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Auth.ValidateUserIPRequest, io.manabie.demo.proto.usermgmt.v2.Auth.ValidateUserIPResponse> getValidateUserIPMethod;
    if ((getValidateUserIPMethod = AuthServiceGrpc.getValidateUserIPMethod) == null) {
      synchronized (AuthServiceGrpc.class) {
        if ((getValidateUserIPMethod = AuthServiceGrpc.getValidateUserIPMethod) == null) {
          AuthServiceGrpc.getValidateUserIPMethod = getValidateUserIPMethod =
              io.grpc.MethodDescriptor.<io.manabie.demo.proto.usermgmt.v2.Auth.ValidateUserIPRequest, io.manabie.demo.proto.usermgmt.v2.Auth.ValidateUserIPResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ValidateUserIP"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.demo.proto.usermgmt.v2.Auth.ValidateUserIPRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.demo.proto.usermgmt.v2.Auth.ValidateUserIPResponse.getDefaultInstance()))
              .setSchemaDescriptor(new AuthServiceMethodDescriptorSupplier("ValidateUserIP"))
              .build();
        }
      }
    }
    return getValidateUserIPMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static AuthServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<AuthServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<AuthServiceStub>() {
        @java.lang.Override
        public AuthServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new AuthServiceStub(channel, callOptions);
        }
      };
    return AuthServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static AuthServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<AuthServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<AuthServiceBlockingStub>() {
        @java.lang.Override
        public AuthServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new AuthServiceBlockingStub(channel, callOptions);
        }
      };
    return AuthServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static AuthServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<AuthServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<AuthServiceFutureStub>() {
        @java.lang.Override
        public AuthServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new AuthServiceFutureStub(channel, callOptions);
        }
      };
    return AuthServiceFutureStub.newStub(factory, channel);
  }

  /**
   */
  public interface AsyncService {

    /**
     */
    default void exchangeCustomToken(io.manabie.demo.proto.usermgmt.v2.Auth.ExchangeCustomTokenRequest request,
        io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Auth.ExchangeCustomTokenResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getExchangeCustomTokenMethod(), responseObserver);
    }

    /**
     */
    default void validateUserIP(io.manabie.demo.proto.usermgmt.v2.Auth.ValidateUserIPRequest request,
        io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Auth.ValidateUserIPResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getValidateUserIPMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service AuthService.
   */
  public static abstract class AuthServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return AuthServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service AuthService.
   */
  public static final class AuthServiceStub
      extends io.grpc.stub.AbstractAsyncStub<AuthServiceStub> {
    private AuthServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected AuthServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new AuthServiceStub(channel, callOptions);
    }

    /**
     */
    public void exchangeCustomToken(io.manabie.demo.proto.usermgmt.v2.Auth.ExchangeCustomTokenRequest request,
        io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Auth.ExchangeCustomTokenResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getExchangeCustomTokenMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void validateUserIP(io.manabie.demo.proto.usermgmt.v2.Auth.ValidateUserIPRequest request,
        io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Auth.ValidateUserIPResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getValidateUserIPMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service AuthService.
   */
  public static final class AuthServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<AuthServiceBlockingStub> {
    private AuthServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected AuthServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new AuthServiceBlockingStub(channel, callOptions);
    }

    /**
     */
    public io.manabie.demo.proto.usermgmt.v2.Auth.ExchangeCustomTokenResponse exchangeCustomToken(io.manabie.demo.proto.usermgmt.v2.Auth.ExchangeCustomTokenRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getExchangeCustomTokenMethod(), getCallOptions(), request);
    }

    /**
     */
    public io.manabie.demo.proto.usermgmt.v2.Auth.ValidateUserIPResponse validateUserIP(io.manabie.demo.proto.usermgmt.v2.Auth.ValidateUserIPRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getValidateUserIPMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service AuthService.
   */
  public static final class AuthServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<AuthServiceFutureStub> {
    private AuthServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected AuthServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new AuthServiceFutureStub(channel, callOptions);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.demo.proto.usermgmt.v2.Auth.ExchangeCustomTokenResponse> exchangeCustomToken(
        io.manabie.demo.proto.usermgmt.v2.Auth.ExchangeCustomTokenRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getExchangeCustomTokenMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.demo.proto.usermgmt.v2.Auth.ValidateUserIPResponse> validateUserIP(
        io.manabie.demo.proto.usermgmt.v2.Auth.ValidateUserIPRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getValidateUserIPMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_EXCHANGE_CUSTOM_TOKEN = 0;
  private static final int METHODID_VALIDATE_USER_IP = 1;

  private static final class MethodHandlers<Req, Resp> implements
      io.grpc.stub.ServerCalls.UnaryMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ServerStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ClientStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.BidiStreamingMethod<Req, Resp> {
    private final AsyncService serviceImpl;
    private final int methodId;

    MethodHandlers(AsyncService serviceImpl, int methodId) {
      this.serviceImpl = serviceImpl;
      this.methodId = methodId;
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public void invoke(Req request, io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        case METHODID_EXCHANGE_CUSTOM_TOKEN:
          serviceImpl.exchangeCustomToken((io.manabie.demo.proto.usermgmt.v2.Auth.ExchangeCustomTokenRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Auth.ExchangeCustomTokenResponse>) responseObserver);
          break;
        case METHODID_VALIDATE_USER_IP:
          serviceImpl.validateUserIP((io.manabie.demo.proto.usermgmt.v2.Auth.ValidateUserIPRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Auth.ValidateUserIPResponse>) responseObserver);
          break;
        default:
          throw new AssertionError();
      }
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public io.grpc.stub.StreamObserver<Req> invoke(
        io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        default:
          throw new AssertionError();
      }
    }
  }

  public static final io.grpc.ServerServiceDefinition bindService(AsyncService service) {
    return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
        .addMethod(
          getExchangeCustomTokenMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.demo.proto.usermgmt.v2.Auth.ExchangeCustomTokenRequest,
              io.manabie.demo.proto.usermgmt.v2.Auth.ExchangeCustomTokenResponse>(
                service, METHODID_EXCHANGE_CUSTOM_TOKEN)))
        .addMethod(
          getValidateUserIPMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.demo.proto.usermgmt.v2.Auth.ValidateUserIPRequest,
              io.manabie.demo.proto.usermgmt.v2.Auth.ValidateUserIPResponse>(
                service, METHODID_VALIDATE_USER_IP)))
        .build();
  }

  private static abstract class AuthServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    AuthServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return io.manabie.demo.proto.usermgmt.v2.Auth.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("AuthService");
    }
  }

  private static final class AuthServiceFileDescriptorSupplier
      extends AuthServiceBaseDescriptorSupplier {
    AuthServiceFileDescriptorSupplier() {}
  }

  private static final class AuthServiceMethodDescriptorSupplier
      extends AuthServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final String methodName;

    AuthServiceMethodDescriptorSupplier(String methodName) {
      this.methodName = methodName;
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.MethodDescriptor getMethodDescriptor() {
      return getServiceDescriptor().findMethodByName(methodName);
    }
  }

  private static volatile io.grpc.ServiceDescriptor serviceDescriptor;

  public static io.grpc.ServiceDescriptor getServiceDescriptor() {
    io.grpc.ServiceDescriptor result = serviceDescriptor;
    if (result == null) {
      synchronized (AuthServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new AuthServiceFileDescriptorSupplier())
              .addMethod(getExchangeCustomTokenMethod())
              .addMethod(getValidateUserIPMethod())
              .build();
        }
      }
    }
    return result;
  }
}
