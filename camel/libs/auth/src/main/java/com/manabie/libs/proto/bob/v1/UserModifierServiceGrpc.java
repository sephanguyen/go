package com.manabie.libs.proto.bob.v1;

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
  private static volatile io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.UpdateUserProfileRequest,
      com.manabie.libs.proto.bob.v1.Users.UpdateUserProfileResponse> getUpdateUserProfileMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UpdateUserProfile",
      requestType = com.manabie.libs.proto.bob.v1.Users.UpdateUserProfileRequest.class,
      responseType = com.manabie.libs.proto.bob.v1.Users.UpdateUserProfileResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.UpdateUserProfileRequest,
      com.manabie.libs.proto.bob.v1.Users.UpdateUserProfileResponse> getUpdateUserProfileMethod() {
    io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.UpdateUserProfileRequest, com.manabie.libs.proto.bob.v1.Users.UpdateUserProfileResponse> getUpdateUserProfileMethod;
    if ((getUpdateUserProfileMethod = UserModifierServiceGrpc.getUpdateUserProfileMethod) == null) {
      synchronized (UserModifierServiceGrpc.class) {
        if ((getUpdateUserProfileMethod = UserModifierServiceGrpc.getUpdateUserProfileMethod) == null) {
          UserModifierServiceGrpc.getUpdateUserProfileMethod = getUpdateUserProfileMethod =
              io.grpc.MethodDescriptor.<com.manabie.libs.proto.bob.v1.Users.UpdateUserProfileRequest, com.manabie.libs.proto.bob.v1.Users.UpdateUserProfileResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UpdateUserProfile"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.manabie.libs.proto.bob.v1.Users.UpdateUserProfileRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.manabie.libs.proto.bob.v1.Users.UpdateUserProfileResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserModifierServiceMethodDescriptorSupplier("UpdateUserProfile"))
              .build();
        }
      }
    }
    return getUpdateUserProfileMethod;
  }

  private static volatile io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.UpdateUserDeviceTokenRequest,
      com.manabie.libs.proto.bob.v1.Users.UpdateUserDeviceTokenResponse> getUpdateUserDeviceTokenMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UpdateUserDeviceToken",
      requestType = com.manabie.libs.proto.bob.v1.Users.UpdateUserDeviceTokenRequest.class,
      responseType = com.manabie.libs.proto.bob.v1.Users.UpdateUserDeviceTokenResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.UpdateUserDeviceTokenRequest,
      com.manabie.libs.proto.bob.v1.Users.UpdateUserDeviceTokenResponse> getUpdateUserDeviceTokenMethod() {
    io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.UpdateUserDeviceTokenRequest, com.manabie.libs.proto.bob.v1.Users.UpdateUserDeviceTokenResponse> getUpdateUserDeviceTokenMethod;
    if ((getUpdateUserDeviceTokenMethod = UserModifierServiceGrpc.getUpdateUserDeviceTokenMethod) == null) {
      synchronized (UserModifierServiceGrpc.class) {
        if ((getUpdateUserDeviceTokenMethod = UserModifierServiceGrpc.getUpdateUserDeviceTokenMethod) == null) {
          UserModifierServiceGrpc.getUpdateUserDeviceTokenMethod = getUpdateUserDeviceTokenMethod =
              io.grpc.MethodDescriptor.<com.manabie.libs.proto.bob.v1.Users.UpdateUserDeviceTokenRequest, com.manabie.libs.proto.bob.v1.Users.UpdateUserDeviceTokenResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UpdateUserDeviceToken"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.manabie.libs.proto.bob.v1.Users.UpdateUserDeviceTokenRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.manabie.libs.proto.bob.v1.Users.UpdateUserDeviceTokenResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserModifierServiceMethodDescriptorSupplier("UpdateUserDeviceToken"))
              .build();
        }
      }
    }
    return getUpdateUserDeviceTokenMethod;
  }

  private static volatile io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.ExchangeTokenRequest,
      com.manabie.libs.proto.bob.v1.Users.ExchangeTokenResponse> getExchangeTokenMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ExchangeToken",
      requestType = com.manabie.libs.proto.bob.v1.Users.ExchangeTokenRequest.class,
      responseType = com.manabie.libs.proto.bob.v1.Users.ExchangeTokenResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.ExchangeTokenRequest,
      com.manabie.libs.proto.bob.v1.Users.ExchangeTokenResponse> getExchangeTokenMethod() {
    io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.ExchangeTokenRequest, com.manabie.libs.proto.bob.v1.Users.ExchangeTokenResponse> getExchangeTokenMethod;
    if ((getExchangeTokenMethod = UserModifierServiceGrpc.getExchangeTokenMethod) == null) {
      synchronized (UserModifierServiceGrpc.class) {
        if ((getExchangeTokenMethod = UserModifierServiceGrpc.getExchangeTokenMethod) == null) {
          UserModifierServiceGrpc.getExchangeTokenMethod = getExchangeTokenMethod =
              io.grpc.MethodDescriptor.<com.manabie.libs.proto.bob.v1.Users.ExchangeTokenRequest, com.manabie.libs.proto.bob.v1.Users.ExchangeTokenResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ExchangeToken"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.manabie.libs.proto.bob.v1.Users.ExchangeTokenRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.manabie.libs.proto.bob.v1.Users.ExchangeTokenResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserModifierServiceMethodDescriptorSupplier("ExchangeToken"))
              .build();
        }
      }
    }
    return getExchangeTokenMethod;
  }

  private static volatile io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.RegisterRequest,
      com.manabie.libs.proto.bob.v1.Users.RegisterResponse> getRegisterMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Register",
      requestType = com.manabie.libs.proto.bob.v1.Users.RegisterRequest.class,
      responseType = com.manabie.libs.proto.bob.v1.Users.RegisterResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.RegisterRequest,
      com.manabie.libs.proto.bob.v1.Users.RegisterResponse> getRegisterMethod() {
    io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.RegisterRequest, com.manabie.libs.proto.bob.v1.Users.RegisterResponse> getRegisterMethod;
    if ((getRegisterMethod = UserModifierServiceGrpc.getRegisterMethod) == null) {
      synchronized (UserModifierServiceGrpc.class) {
        if ((getRegisterMethod = UserModifierServiceGrpc.getRegisterMethod) == null) {
          UserModifierServiceGrpc.getRegisterMethod = getRegisterMethod =
              io.grpc.MethodDescriptor.<com.manabie.libs.proto.bob.v1.Users.RegisterRequest, com.manabie.libs.proto.bob.v1.Users.RegisterResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Register"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.manabie.libs.proto.bob.v1.Users.RegisterRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.manabie.libs.proto.bob.v1.Users.RegisterResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserModifierServiceMethodDescriptorSupplier("Register"))
              .build();
        }
      }
    }
    return getRegisterMethod;
  }

  private static volatile io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.ExchangeCustomTokenRequest,
      com.manabie.libs.proto.bob.v1.Users.ExchangeCustomTokenResponse> getExchangeCustomTokenMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ExchangeCustomToken",
      requestType = com.manabie.libs.proto.bob.v1.Users.ExchangeCustomTokenRequest.class,
      responseType = com.manabie.libs.proto.bob.v1.Users.ExchangeCustomTokenResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.ExchangeCustomTokenRequest,
      com.manabie.libs.proto.bob.v1.Users.ExchangeCustomTokenResponse> getExchangeCustomTokenMethod() {
    io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.ExchangeCustomTokenRequest, com.manabie.libs.proto.bob.v1.Users.ExchangeCustomTokenResponse> getExchangeCustomTokenMethod;
    if ((getExchangeCustomTokenMethod = UserModifierServiceGrpc.getExchangeCustomTokenMethod) == null) {
      synchronized (UserModifierServiceGrpc.class) {
        if ((getExchangeCustomTokenMethod = UserModifierServiceGrpc.getExchangeCustomTokenMethod) == null) {
          UserModifierServiceGrpc.getExchangeCustomTokenMethod = getExchangeCustomTokenMethod =
              io.grpc.MethodDescriptor.<com.manabie.libs.proto.bob.v1.Users.ExchangeCustomTokenRequest, com.manabie.libs.proto.bob.v1.Users.ExchangeCustomTokenResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ExchangeCustomToken"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.manabie.libs.proto.bob.v1.Users.ExchangeCustomTokenRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.manabie.libs.proto.bob.v1.Users.ExchangeCustomTokenResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserModifierServiceMethodDescriptorSupplier("ExchangeCustomToken"))
              .build();
        }
      }
    }
    return getExchangeCustomTokenMethod;
  }

  private static volatile io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.UpdateUserLastLoginDateRequest,
      com.manabie.libs.proto.bob.v1.Users.UpdateUserLastLoginDateResponse> getUpdateUserLastLoginDateMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UpdateUserLastLoginDate",
      requestType = com.manabie.libs.proto.bob.v1.Users.UpdateUserLastLoginDateRequest.class,
      responseType = com.manabie.libs.proto.bob.v1.Users.UpdateUserLastLoginDateResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.UpdateUserLastLoginDateRequest,
      com.manabie.libs.proto.bob.v1.Users.UpdateUserLastLoginDateResponse> getUpdateUserLastLoginDateMethod() {
    io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.UpdateUserLastLoginDateRequest, com.manabie.libs.proto.bob.v1.Users.UpdateUserLastLoginDateResponse> getUpdateUserLastLoginDateMethod;
    if ((getUpdateUserLastLoginDateMethod = UserModifierServiceGrpc.getUpdateUserLastLoginDateMethod) == null) {
      synchronized (UserModifierServiceGrpc.class) {
        if ((getUpdateUserLastLoginDateMethod = UserModifierServiceGrpc.getUpdateUserLastLoginDateMethod) == null) {
          UserModifierServiceGrpc.getUpdateUserLastLoginDateMethod = getUpdateUserLastLoginDateMethod =
              io.grpc.MethodDescriptor.<com.manabie.libs.proto.bob.v1.Users.UpdateUserLastLoginDateRequest, com.manabie.libs.proto.bob.v1.Users.UpdateUserLastLoginDateResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UpdateUserLastLoginDate"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.manabie.libs.proto.bob.v1.Users.UpdateUserLastLoginDateRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.manabie.libs.proto.bob.v1.Users.UpdateUserLastLoginDateResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserModifierServiceMethodDescriptorSupplier("UpdateUserLastLoginDate"))
              .build();
        }
      }
    }
    return getUpdateUserLastLoginDateMethod;
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
    default void updateUserProfile(com.manabie.libs.proto.bob.v1.Users.UpdateUserProfileRequest request,
        io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.UpdateUserProfileResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUpdateUserProfileMethod(), responseObserver);
    }

    /**
     */
    default void updateUserDeviceToken(com.manabie.libs.proto.bob.v1.Users.UpdateUserDeviceTokenRequest request,
        io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.UpdateUserDeviceTokenResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUpdateUserDeviceTokenMethod(), responseObserver);
    }

    /**
     */
    default void exchangeToken(com.manabie.libs.proto.bob.v1.Users.ExchangeTokenRequest request,
        io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.ExchangeTokenResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getExchangeTokenMethod(), responseObserver);
    }

    /**
     */
    default void register(com.manabie.libs.proto.bob.v1.Users.RegisterRequest request,
        io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.RegisterResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getRegisterMethod(), responseObserver);
    }

    /**
     */
    default void exchangeCustomToken(com.manabie.libs.proto.bob.v1.Users.ExchangeCustomTokenRequest request,
        io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.ExchangeCustomTokenResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getExchangeCustomTokenMethod(), responseObserver);
    }

    /**
     */
    default void updateUserLastLoginDate(com.manabie.libs.proto.bob.v1.Users.UpdateUserLastLoginDateRequest request,
        io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.UpdateUserLastLoginDateResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUpdateUserLastLoginDateMethod(), responseObserver);
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
    public void updateUserProfile(com.manabie.libs.proto.bob.v1.Users.UpdateUserProfileRequest request,
        io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.UpdateUserProfileResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUpdateUserProfileMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void updateUserDeviceToken(com.manabie.libs.proto.bob.v1.Users.UpdateUserDeviceTokenRequest request,
        io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.UpdateUserDeviceTokenResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUpdateUserDeviceTokenMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void exchangeToken(com.manabie.libs.proto.bob.v1.Users.ExchangeTokenRequest request,
        io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.ExchangeTokenResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getExchangeTokenMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void register(com.manabie.libs.proto.bob.v1.Users.RegisterRequest request,
        io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.RegisterResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getRegisterMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void exchangeCustomToken(com.manabie.libs.proto.bob.v1.Users.ExchangeCustomTokenRequest request,
        io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.ExchangeCustomTokenResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getExchangeCustomTokenMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void updateUserLastLoginDate(com.manabie.libs.proto.bob.v1.Users.UpdateUserLastLoginDateRequest request,
        io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.UpdateUserLastLoginDateResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUpdateUserLastLoginDateMethod(), getCallOptions()), request, responseObserver);
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
    public com.manabie.libs.proto.bob.v1.Users.UpdateUserProfileResponse updateUserProfile(com.manabie.libs.proto.bob.v1.Users.UpdateUserProfileRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUpdateUserProfileMethod(), getCallOptions(), request);
    }

    /**
     */
    public com.manabie.libs.proto.bob.v1.Users.UpdateUserDeviceTokenResponse updateUserDeviceToken(com.manabie.libs.proto.bob.v1.Users.UpdateUserDeviceTokenRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUpdateUserDeviceTokenMethod(), getCallOptions(), request);
    }

    /**
     */
    public com.manabie.libs.proto.bob.v1.Users.ExchangeTokenResponse exchangeToken(com.manabie.libs.proto.bob.v1.Users.ExchangeTokenRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getExchangeTokenMethod(), getCallOptions(), request);
    }

    /**
     */
    public com.manabie.libs.proto.bob.v1.Users.RegisterResponse register(com.manabie.libs.proto.bob.v1.Users.RegisterRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getRegisterMethod(), getCallOptions(), request);
    }

    /**
     */
    public com.manabie.libs.proto.bob.v1.Users.ExchangeCustomTokenResponse exchangeCustomToken(com.manabie.libs.proto.bob.v1.Users.ExchangeCustomTokenRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getExchangeCustomTokenMethod(), getCallOptions(), request);
    }

    /**
     */
    public com.manabie.libs.proto.bob.v1.Users.UpdateUserLastLoginDateResponse updateUserLastLoginDate(com.manabie.libs.proto.bob.v1.Users.UpdateUserLastLoginDateRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUpdateUserLastLoginDateMethod(), getCallOptions(), request);
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
    public com.google.common.util.concurrent.ListenableFuture<com.manabie.libs.proto.bob.v1.Users.UpdateUserProfileResponse> updateUserProfile(
        com.manabie.libs.proto.bob.v1.Users.UpdateUserProfileRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUpdateUserProfileMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<com.manabie.libs.proto.bob.v1.Users.UpdateUserDeviceTokenResponse> updateUserDeviceToken(
        com.manabie.libs.proto.bob.v1.Users.UpdateUserDeviceTokenRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUpdateUserDeviceTokenMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<com.manabie.libs.proto.bob.v1.Users.ExchangeTokenResponse> exchangeToken(
        com.manabie.libs.proto.bob.v1.Users.ExchangeTokenRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getExchangeTokenMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<com.manabie.libs.proto.bob.v1.Users.RegisterResponse> register(
        com.manabie.libs.proto.bob.v1.Users.RegisterRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getRegisterMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<com.manabie.libs.proto.bob.v1.Users.ExchangeCustomTokenResponse> exchangeCustomToken(
        com.manabie.libs.proto.bob.v1.Users.ExchangeCustomTokenRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getExchangeCustomTokenMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<com.manabie.libs.proto.bob.v1.Users.UpdateUserLastLoginDateResponse> updateUserLastLoginDate(
        com.manabie.libs.proto.bob.v1.Users.UpdateUserLastLoginDateRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUpdateUserLastLoginDateMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_UPDATE_USER_PROFILE = 0;
  private static final int METHODID_UPDATE_USER_DEVICE_TOKEN = 1;
  private static final int METHODID_EXCHANGE_TOKEN = 2;
  private static final int METHODID_REGISTER = 3;
  private static final int METHODID_EXCHANGE_CUSTOM_TOKEN = 4;
  private static final int METHODID_UPDATE_USER_LAST_LOGIN_DATE = 5;

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
        case METHODID_UPDATE_USER_PROFILE:
          serviceImpl.updateUserProfile((com.manabie.libs.proto.bob.v1.Users.UpdateUserProfileRequest) request,
              (io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.UpdateUserProfileResponse>) responseObserver);
          break;
        case METHODID_UPDATE_USER_DEVICE_TOKEN:
          serviceImpl.updateUserDeviceToken((com.manabie.libs.proto.bob.v1.Users.UpdateUserDeviceTokenRequest) request,
              (io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.UpdateUserDeviceTokenResponse>) responseObserver);
          break;
        case METHODID_EXCHANGE_TOKEN:
          serviceImpl.exchangeToken((com.manabie.libs.proto.bob.v1.Users.ExchangeTokenRequest) request,
              (io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.ExchangeTokenResponse>) responseObserver);
          break;
        case METHODID_REGISTER:
          serviceImpl.register((com.manabie.libs.proto.bob.v1.Users.RegisterRequest) request,
              (io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.RegisterResponse>) responseObserver);
          break;
        case METHODID_EXCHANGE_CUSTOM_TOKEN:
          serviceImpl.exchangeCustomToken((com.manabie.libs.proto.bob.v1.Users.ExchangeCustomTokenRequest) request,
              (io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.ExchangeCustomTokenResponse>) responseObserver);
          break;
        case METHODID_UPDATE_USER_LAST_LOGIN_DATE:
          serviceImpl.updateUserLastLoginDate((com.manabie.libs.proto.bob.v1.Users.UpdateUserLastLoginDateRequest) request,
              (io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.UpdateUserLastLoginDateResponse>) responseObserver);
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
          getUpdateUserProfileMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.manabie.libs.proto.bob.v1.Users.UpdateUserProfileRequest,
              com.manabie.libs.proto.bob.v1.Users.UpdateUserProfileResponse>(
                service, METHODID_UPDATE_USER_PROFILE)))
        .addMethod(
          getUpdateUserDeviceTokenMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.manabie.libs.proto.bob.v1.Users.UpdateUserDeviceTokenRequest,
              com.manabie.libs.proto.bob.v1.Users.UpdateUserDeviceTokenResponse>(
                service, METHODID_UPDATE_USER_DEVICE_TOKEN)))
        .addMethod(
          getExchangeTokenMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.manabie.libs.proto.bob.v1.Users.ExchangeTokenRequest,
              com.manabie.libs.proto.bob.v1.Users.ExchangeTokenResponse>(
                service, METHODID_EXCHANGE_TOKEN)))
        .addMethod(
          getRegisterMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.manabie.libs.proto.bob.v1.Users.RegisterRequest,
              com.manabie.libs.proto.bob.v1.Users.RegisterResponse>(
                service, METHODID_REGISTER)))
        .addMethod(
          getExchangeCustomTokenMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.manabie.libs.proto.bob.v1.Users.ExchangeCustomTokenRequest,
              com.manabie.libs.proto.bob.v1.Users.ExchangeCustomTokenResponse>(
                service, METHODID_EXCHANGE_CUSTOM_TOKEN)))
        .addMethod(
          getUpdateUserLastLoginDateMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.manabie.libs.proto.bob.v1.Users.UpdateUserLastLoginDateRequest,
              com.manabie.libs.proto.bob.v1.Users.UpdateUserLastLoginDateResponse>(
                service, METHODID_UPDATE_USER_LAST_LOGIN_DATE)))
        .build();
  }

  private static abstract class UserModifierServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    UserModifierServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return com.manabie.libs.proto.bob.v1.Users.getDescriptor();
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
              .addMethod(getUpdateUserProfileMethod())
              .addMethod(getUpdateUserDeviceTokenMethod())
              .addMethod(getExchangeTokenMethod())
              .addMethod(getRegisterMethod())
              .addMethod(getExchangeCustomTokenMethod())
              .addMethod(getUpdateUserLastLoginDateMethod())
              .build();
        }
      }
    }
    return result;
  }
}
