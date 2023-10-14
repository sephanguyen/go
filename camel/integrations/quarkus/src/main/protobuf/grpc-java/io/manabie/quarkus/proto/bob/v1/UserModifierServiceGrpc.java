package io.manabie.quarkus.proto.bob.v1;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.56.1)",
    comments = "Source: bob/v1/users.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class UserModifierServiceGrpc {

  private UserModifierServiceGrpc() {}

  public static final String SERVICE_NAME = "bob.v1.UserModifierService";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<io.manabie.quarkus.proto.bob.v1.Users.ExchangeTokenRequest,
      io.manabie.quarkus.proto.bob.v1.Users.ExchangeTokenResponse> getExchangeTokenMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ExchangeToken",
      requestType = io.manabie.quarkus.proto.bob.v1.Users.ExchangeTokenRequest.class,
      responseType = io.manabie.quarkus.proto.bob.v1.Users.ExchangeTokenResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.quarkus.proto.bob.v1.Users.ExchangeTokenRequest,
      io.manabie.quarkus.proto.bob.v1.Users.ExchangeTokenResponse> getExchangeTokenMethod() {
    io.grpc.MethodDescriptor<io.manabie.quarkus.proto.bob.v1.Users.ExchangeTokenRequest, io.manabie.quarkus.proto.bob.v1.Users.ExchangeTokenResponse> getExchangeTokenMethod;
    if ((getExchangeTokenMethod = UserModifierServiceGrpc.getExchangeTokenMethod) == null) {
      synchronized (UserModifierServiceGrpc.class) {
        if ((getExchangeTokenMethod = UserModifierServiceGrpc.getExchangeTokenMethod) == null) {
          UserModifierServiceGrpc.getExchangeTokenMethod = getExchangeTokenMethod =
              io.grpc.MethodDescriptor.<io.manabie.quarkus.proto.bob.v1.Users.ExchangeTokenRequest, io.manabie.quarkus.proto.bob.v1.Users.ExchangeTokenResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ExchangeToken"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.bob.v1.Users.ExchangeTokenRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.bob.v1.Users.ExchangeTokenResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserModifierServiceMethodDescriptorSupplier("ExchangeToken"))
              .build();
        }
      }
    }
    return getExchangeTokenMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static UserModifierServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<UserModifierServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<UserModifierServiceStub>() {
        @java.lang.Override
        public UserModifierServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new UserModifierServiceStub(channel, callOptions);
        }
      };
    return UserModifierServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static UserModifierServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<UserModifierServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<UserModifierServiceBlockingStub>() {
        @java.lang.Override
        public UserModifierServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new UserModifierServiceBlockingStub(channel, callOptions);
        }
      };
    return UserModifierServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static UserModifierServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<UserModifierServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<UserModifierServiceFutureStub>() {
        @java.lang.Override
        public UserModifierServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new UserModifierServiceFutureStub(channel, callOptions);
        }
      };
    return UserModifierServiceFutureStub.newStub(factory, channel);
  }

  /**
   */
  public interface AsyncService {

    /**
     */
    default void exchangeToken(io.manabie.quarkus.proto.bob.v1.Users.ExchangeTokenRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.bob.v1.Users.ExchangeTokenResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getExchangeTokenMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service UserModifierService.
   */
  public static abstract class UserModifierServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return UserModifierServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service UserModifierService.
   */
  public static final class UserModifierServiceStub
      extends io.grpc.stub.AbstractAsyncStub<UserModifierServiceStub> {
    private UserModifierServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected UserModifierServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new UserModifierServiceStub(channel, callOptions);
    }

    /**
     */
    public void exchangeToken(io.manabie.quarkus.proto.bob.v1.Users.ExchangeTokenRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.bob.v1.Users.ExchangeTokenResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getExchangeTokenMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service UserModifierService.
   */
  public static final class UserModifierServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<UserModifierServiceBlockingStub> {
    private UserModifierServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected UserModifierServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new UserModifierServiceBlockingStub(channel, callOptions);
    }

    /**
     */
    public io.manabie.quarkus.proto.bob.v1.Users.ExchangeTokenResponse exchangeToken(io.manabie.quarkus.proto.bob.v1.Users.ExchangeTokenRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getExchangeTokenMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service UserModifierService.
   */
  public static final class UserModifierServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<UserModifierServiceFutureStub> {
    private UserModifierServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected UserModifierServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new UserModifierServiceFutureStub(channel, callOptions);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.quarkus.proto.bob.v1.Users.ExchangeTokenResponse> exchangeToken(
        io.manabie.quarkus.proto.bob.v1.Users.ExchangeTokenRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getExchangeTokenMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_EXCHANGE_TOKEN = 0;

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
        case METHODID_EXCHANGE_TOKEN:
          serviceImpl.exchangeToken((io.manabie.quarkus.proto.bob.v1.Users.ExchangeTokenRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.bob.v1.Users.ExchangeTokenResponse>) responseObserver);
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
          getExchangeTokenMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.quarkus.proto.bob.v1.Users.ExchangeTokenRequest,
              io.manabie.quarkus.proto.bob.v1.Users.ExchangeTokenResponse>(
                service, METHODID_EXCHANGE_TOKEN)))
        .build();
  }

  private static abstract class UserModifierServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    UserModifierServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return io.manabie.quarkus.proto.bob.v1.Users.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("UserModifierService");
    }
  }

  private static final class UserModifierServiceFileDescriptorSupplier
      extends UserModifierServiceBaseDescriptorSupplier {
    UserModifierServiceFileDescriptorSupplier() {}
  }

  private static final class UserModifierServiceMethodDescriptorSupplier
      extends UserModifierServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final String methodName;

    UserModifierServiceMethodDescriptorSupplier(String methodName) {
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
      synchronized (UserModifierServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new UserModifierServiceFileDescriptorSupplier())
              .addMethod(getExchangeTokenMethod())
              .build();
        }
      }
    }
    return result;
  }
}
