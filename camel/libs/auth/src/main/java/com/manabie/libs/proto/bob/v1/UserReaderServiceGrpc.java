package com.manabie.libs.proto.bob.v1;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.56.1)",
    comments = "Source: bob/v1/users.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class UserReaderServiceGrpc {

  private UserReaderServiceGrpc() {}

  public static final String SERVICE_NAME = "bob.v1.UserReaderService";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.GetCurrentUserProfileRequest,
      com.manabie.libs.proto.bob.v1.Users.GetCurrentUserProfileResponse> getGetCurrentUserProfileMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetCurrentUserProfile",
      requestType = com.manabie.libs.proto.bob.v1.Users.GetCurrentUserProfileRequest.class,
      responseType = com.manabie.libs.proto.bob.v1.Users.GetCurrentUserProfileResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.GetCurrentUserProfileRequest,
      com.manabie.libs.proto.bob.v1.Users.GetCurrentUserProfileResponse> getGetCurrentUserProfileMethod() {
    io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.GetCurrentUserProfileRequest, com.manabie.libs.proto.bob.v1.Users.GetCurrentUserProfileResponse> getGetCurrentUserProfileMethod;
    if ((getGetCurrentUserProfileMethod = UserReaderServiceGrpc.getGetCurrentUserProfileMethod) == null) {
      synchronized (UserReaderServiceGrpc.class) {
        if ((getGetCurrentUserProfileMethod = UserReaderServiceGrpc.getGetCurrentUserProfileMethod) == null) {
          UserReaderServiceGrpc.getGetCurrentUserProfileMethod = getGetCurrentUserProfileMethod =
              io.grpc.MethodDescriptor.<com.manabie.libs.proto.bob.v1.Users.GetCurrentUserProfileRequest, com.manabie.libs.proto.bob.v1.Users.GetCurrentUserProfileResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetCurrentUserProfile"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.manabie.libs.proto.bob.v1.Users.GetCurrentUserProfileRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.manabie.libs.proto.bob.v1.Users.GetCurrentUserProfileResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserReaderServiceMethodDescriptorSupplier("GetCurrentUserProfile"))
              .build();
        }
      }
    }
    return getGetCurrentUserProfileMethod;
  }

  private static volatile io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.RetrieveTeacherProfilesRequest,
      com.manabie.libs.proto.bob.v1.Users.RetrieveTeacherProfilesResponse> getRetrieveTeacherProfilesMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "RetrieveTeacherProfiles",
      requestType = com.manabie.libs.proto.bob.v1.Users.RetrieveTeacherProfilesRequest.class,
      responseType = com.manabie.libs.proto.bob.v1.Users.RetrieveTeacherProfilesResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.RetrieveTeacherProfilesRequest,
      com.manabie.libs.proto.bob.v1.Users.RetrieveTeacherProfilesResponse> getRetrieveTeacherProfilesMethod() {
    io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.RetrieveTeacherProfilesRequest, com.manabie.libs.proto.bob.v1.Users.RetrieveTeacherProfilesResponse> getRetrieveTeacherProfilesMethod;
    if ((getRetrieveTeacherProfilesMethod = UserReaderServiceGrpc.getRetrieveTeacherProfilesMethod) == null) {
      synchronized (UserReaderServiceGrpc.class) {
        if ((getRetrieveTeacherProfilesMethod = UserReaderServiceGrpc.getRetrieveTeacherProfilesMethod) == null) {
          UserReaderServiceGrpc.getRetrieveTeacherProfilesMethod = getRetrieveTeacherProfilesMethod =
              io.grpc.MethodDescriptor.<com.manabie.libs.proto.bob.v1.Users.RetrieveTeacherProfilesRequest, com.manabie.libs.proto.bob.v1.Users.RetrieveTeacherProfilesResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "RetrieveTeacherProfiles"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.manabie.libs.proto.bob.v1.Users.RetrieveTeacherProfilesRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.manabie.libs.proto.bob.v1.Users.RetrieveTeacherProfilesResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserReaderServiceMethodDescriptorSupplier("RetrieveTeacherProfiles"))
              .build();
        }
      }
    }
    return getRetrieveTeacherProfilesMethod;
  }

  private static volatile io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.RetrieveBasicProfileRequest,
      com.manabie.libs.proto.bob.v1.Users.RetrieveBasicProfileResponse> getRetrieveBasicProfileMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "RetrieveBasicProfile",
      requestType = com.manabie.libs.proto.bob.v1.Users.RetrieveBasicProfileRequest.class,
      responseType = com.manabie.libs.proto.bob.v1.Users.RetrieveBasicProfileResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.RetrieveBasicProfileRequest,
      com.manabie.libs.proto.bob.v1.Users.RetrieveBasicProfileResponse> getRetrieveBasicProfileMethod() {
    io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.RetrieveBasicProfileRequest, com.manabie.libs.proto.bob.v1.Users.RetrieveBasicProfileResponse> getRetrieveBasicProfileMethod;
    if ((getRetrieveBasicProfileMethod = UserReaderServiceGrpc.getRetrieveBasicProfileMethod) == null) {
      synchronized (UserReaderServiceGrpc.class) {
        if ((getRetrieveBasicProfileMethod = UserReaderServiceGrpc.getRetrieveBasicProfileMethod) == null) {
          UserReaderServiceGrpc.getRetrieveBasicProfileMethod = getRetrieveBasicProfileMethod =
              io.grpc.MethodDescriptor.<com.manabie.libs.proto.bob.v1.Users.RetrieveBasicProfileRequest, com.manabie.libs.proto.bob.v1.Users.RetrieveBasicProfileResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "RetrieveBasicProfile"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.manabie.libs.proto.bob.v1.Users.RetrieveBasicProfileRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.manabie.libs.proto.bob.v1.Users.RetrieveBasicProfileResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserReaderServiceMethodDescriptorSupplier("RetrieveBasicProfile"))
              .build();
        }
      }
    }
    return getRetrieveBasicProfileMethod;
  }

  private static volatile io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.SearchBasicProfileRequest,
      com.manabie.libs.proto.bob.v1.Users.SearchBasicProfileResponse> getSearchBasicProfileMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "SearchBasicProfile",
      requestType = com.manabie.libs.proto.bob.v1.Users.SearchBasicProfileRequest.class,
      responseType = com.manabie.libs.proto.bob.v1.Users.SearchBasicProfileResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.SearchBasicProfileRequest,
      com.manabie.libs.proto.bob.v1.Users.SearchBasicProfileResponse> getSearchBasicProfileMethod() {
    io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.SearchBasicProfileRequest, com.manabie.libs.proto.bob.v1.Users.SearchBasicProfileResponse> getSearchBasicProfileMethod;
    if ((getSearchBasicProfileMethod = UserReaderServiceGrpc.getSearchBasicProfileMethod) == null) {
      synchronized (UserReaderServiceGrpc.class) {
        if ((getSearchBasicProfileMethod = UserReaderServiceGrpc.getSearchBasicProfileMethod) == null) {
          UserReaderServiceGrpc.getSearchBasicProfileMethod = getSearchBasicProfileMethod =
              io.grpc.MethodDescriptor.<com.manabie.libs.proto.bob.v1.Users.SearchBasicProfileRequest, com.manabie.libs.proto.bob.v1.Users.SearchBasicProfileResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "SearchBasicProfile"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.manabie.libs.proto.bob.v1.Users.SearchBasicProfileRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.manabie.libs.proto.bob.v1.Users.SearchBasicProfileResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserReaderServiceMethodDescriptorSupplier("SearchBasicProfile"))
              .build();
        }
      }
    }
    return getSearchBasicProfileMethod;
  }

  private static volatile io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.CheckProfileRequest,
      com.manabie.libs.proto.bob.v1.Users.CheckProfileResponse> getCheckProfileMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "CheckProfile",
      requestType = com.manabie.libs.proto.bob.v1.Users.CheckProfileRequest.class,
      responseType = com.manabie.libs.proto.bob.v1.Users.CheckProfileResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.CheckProfileRequest,
      com.manabie.libs.proto.bob.v1.Users.CheckProfileResponse> getCheckProfileMethod() {
    io.grpc.MethodDescriptor<com.manabie.libs.proto.bob.v1.Users.CheckProfileRequest, com.manabie.libs.proto.bob.v1.Users.CheckProfileResponse> getCheckProfileMethod;
    if ((getCheckProfileMethod = UserReaderServiceGrpc.getCheckProfileMethod) == null) {
      synchronized (UserReaderServiceGrpc.class) {
        if ((getCheckProfileMethod = UserReaderServiceGrpc.getCheckProfileMethod) == null) {
          UserReaderServiceGrpc.getCheckProfileMethod = getCheckProfileMethod =
              io.grpc.MethodDescriptor.<com.manabie.libs.proto.bob.v1.Users.CheckProfileRequest, com.manabie.libs.proto.bob.v1.Users.CheckProfileResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "CheckProfile"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.manabie.libs.proto.bob.v1.Users.CheckProfileRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.manabie.libs.proto.bob.v1.Users.CheckProfileResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserReaderServiceMethodDescriptorSupplier("CheckProfile"))
              .build();
        }
      }
    }
    return getCheckProfileMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static UserReaderServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<UserReaderServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<UserReaderServiceStub>() {
        @java.lang.Override
        public UserReaderServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new UserReaderServiceStub(channel, callOptions);
        }
      };
    return UserReaderServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static UserReaderServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<UserReaderServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<UserReaderServiceBlockingStub>() {
        @java.lang.Override
        public UserReaderServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new UserReaderServiceBlockingStub(channel, callOptions);
        }
      };
    return UserReaderServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static UserReaderServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<UserReaderServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<UserReaderServiceFutureStub>() {
        @java.lang.Override
        public UserReaderServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new UserReaderServiceFutureStub(channel, callOptions);
        }
      };
    return UserReaderServiceFutureStub.newStub(factory, channel);
  }

  /**
   */
  public interface AsyncService {

    /**
     */
    default void getCurrentUserProfile(com.manabie.libs.proto.bob.v1.Users.GetCurrentUserProfileRequest request,
        io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.GetCurrentUserProfileResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetCurrentUserProfileMethod(), responseObserver);
    }

    /**
     */
    default void retrieveTeacherProfiles(com.manabie.libs.proto.bob.v1.Users.RetrieveTeacherProfilesRequest request,
        io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.RetrieveTeacherProfilesResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getRetrieveTeacherProfilesMethod(), responseObserver);
    }

    /**
     */
    default void retrieveBasicProfile(com.manabie.libs.proto.bob.v1.Users.RetrieveBasicProfileRequest request,
        io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.RetrieveBasicProfileResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getRetrieveBasicProfileMethod(), responseObserver);
    }

    /**
     */
    default void searchBasicProfile(com.manabie.libs.proto.bob.v1.Users.SearchBasicProfileRequest request,
        io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.SearchBasicProfileResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getSearchBasicProfileMethod(), responseObserver);
    }

    /**
     */
    default void checkProfile(com.manabie.libs.proto.bob.v1.Users.CheckProfileRequest request,
        io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.CheckProfileResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getCheckProfileMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service UserReaderService.
   */
  public static abstract class UserReaderServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return UserReaderServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service UserReaderService.
   */
  public static final class UserReaderServiceStub
      extends io.grpc.stub.AbstractAsyncStub<UserReaderServiceStub> {
    private UserReaderServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected UserReaderServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new UserReaderServiceStub(channel, callOptions);
    }

    /**
     */
    public void getCurrentUserProfile(com.manabie.libs.proto.bob.v1.Users.GetCurrentUserProfileRequest request,
        io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.GetCurrentUserProfileResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetCurrentUserProfileMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void retrieveTeacherProfiles(com.manabie.libs.proto.bob.v1.Users.RetrieveTeacherProfilesRequest request,
        io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.RetrieveTeacherProfilesResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getRetrieveTeacherProfilesMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void retrieveBasicProfile(com.manabie.libs.proto.bob.v1.Users.RetrieveBasicProfileRequest request,
        io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.RetrieveBasicProfileResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getRetrieveBasicProfileMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void searchBasicProfile(com.manabie.libs.proto.bob.v1.Users.SearchBasicProfileRequest request,
        io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.SearchBasicProfileResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getSearchBasicProfileMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void checkProfile(com.manabie.libs.proto.bob.v1.Users.CheckProfileRequest request,
        io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.CheckProfileResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getCheckProfileMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service UserReaderService.
   */
  public static final class UserReaderServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<UserReaderServiceBlockingStub> {
    private UserReaderServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected UserReaderServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new UserReaderServiceBlockingStub(channel, callOptions);
    }

    /**
     */
    public com.manabie.libs.proto.bob.v1.Users.GetCurrentUserProfileResponse getCurrentUserProfile(com.manabie.libs.proto.bob.v1.Users.GetCurrentUserProfileRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetCurrentUserProfileMethod(), getCallOptions(), request);
    }

    /**
     */
    public com.manabie.libs.proto.bob.v1.Users.RetrieveTeacherProfilesResponse retrieveTeacherProfiles(com.manabie.libs.proto.bob.v1.Users.RetrieveTeacherProfilesRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getRetrieveTeacherProfilesMethod(), getCallOptions(), request);
    }

    /**
     */
    public com.manabie.libs.proto.bob.v1.Users.RetrieveBasicProfileResponse retrieveBasicProfile(com.manabie.libs.proto.bob.v1.Users.RetrieveBasicProfileRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getRetrieveBasicProfileMethod(), getCallOptions(), request);
    }

    /**
     */
    public com.manabie.libs.proto.bob.v1.Users.SearchBasicProfileResponse searchBasicProfile(com.manabie.libs.proto.bob.v1.Users.SearchBasicProfileRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getSearchBasicProfileMethod(), getCallOptions(), request);
    }

    /**
     */
    public com.manabie.libs.proto.bob.v1.Users.CheckProfileResponse checkProfile(com.manabie.libs.proto.bob.v1.Users.CheckProfileRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getCheckProfileMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service UserReaderService.
   */
  public static final class UserReaderServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<UserReaderServiceFutureStub> {
    private UserReaderServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected UserReaderServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new UserReaderServiceFutureStub(channel, callOptions);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<com.manabie.libs.proto.bob.v1.Users.GetCurrentUserProfileResponse> getCurrentUserProfile(
        com.manabie.libs.proto.bob.v1.Users.GetCurrentUserProfileRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetCurrentUserProfileMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<com.manabie.libs.proto.bob.v1.Users.RetrieveTeacherProfilesResponse> retrieveTeacherProfiles(
        com.manabie.libs.proto.bob.v1.Users.RetrieveTeacherProfilesRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getRetrieveTeacherProfilesMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<com.manabie.libs.proto.bob.v1.Users.RetrieveBasicProfileResponse> retrieveBasicProfile(
        com.manabie.libs.proto.bob.v1.Users.RetrieveBasicProfileRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getRetrieveBasicProfileMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<com.manabie.libs.proto.bob.v1.Users.SearchBasicProfileResponse> searchBasicProfile(
        com.manabie.libs.proto.bob.v1.Users.SearchBasicProfileRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getSearchBasicProfileMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<com.manabie.libs.proto.bob.v1.Users.CheckProfileResponse> checkProfile(
        com.manabie.libs.proto.bob.v1.Users.CheckProfileRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getCheckProfileMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_GET_CURRENT_USER_PROFILE = 0;
  private static final int METHODID_RETRIEVE_TEACHER_PROFILES = 1;
  private static final int METHODID_RETRIEVE_BASIC_PROFILE = 2;
  private static final int METHODID_SEARCH_BASIC_PROFILE = 3;
  private static final int METHODID_CHECK_PROFILE = 4;

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
        case METHODID_GET_CURRENT_USER_PROFILE:
          serviceImpl.getCurrentUserProfile((com.manabie.libs.proto.bob.v1.Users.GetCurrentUserProfileRequest) request,
              (io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.GetCurrentUserProfileResponse>) responseObserver);
          break;
        case METHODID_RETRIEVE_TEACHER_PROFILES:
          serviceImpl.retrieveTeacherProfiles((com.manabie.libs.proto.bob.v1.Users.RetrieveTeacherProfilesRequest) request,
              (io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.RetrieveTeacherProfilesResponse>) responseObserver);
          break;
        case METHODID_RETRIEVE_BASIC_PROFILE:
          serviceImpl.retrieveBasicProfile((com.manabie.libs.proto.bob.v1.Users.RetrieveBasicProfileRequest) request,
              (io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.RetrieveBasicProfileResponse>) responseObserver);
          break;
        case METHODID_SEARCH_BASIC_PROFILE:
          serviceImpl.searchBasicProfile((com.manabie.libs.proto.bob.v1.Users.SearchBasicProfileRequest) request,
              (io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.SearchBasicProfileResponse>) responseObserver);
          break;
        case METHODID_CHECK_PROFILE:
          serviceImpl.checkProfile((com.manabie.libs.proto.bob.v1.Users.CheckProfileRequest) request,
              (io.grpc.stub.StreamObserver<com.manabie.libs.proto.bob.v1.Users.CheckProfileResponse>) responseObserver);
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
          getGetCurrentUserProfileMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.manabie.libs.proto.bob.v1.Users.GetCurrentUserProfileRequest,
              com.manabie.libs.proto.bob.v1.Users.GetCurrentUserProfileResponse>(
                service, METHODID_GET_CURRENT_USER_PROFILE)))
        .addMethod(
          getRetrieveTeacherProfilesMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.manabie.libs.proto.bob.v1.Users.RetrieveTeacherProfilesRequest,
              com.manabie.libs.proto.bob.v1.Users.RetrieveTeacherProfilesResponse>(
                service, METHODID_RETRIEVE_TEACHER_PROFILES)))
        .addMethod(
          getRetrieveBasicProfileMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.manabie.libs.proto.bob.v1.Users.RetrieveBasicProfileRequest,
              com.manabie.libs.proto.bob.v1.Users.RetrieveBasicProfileResponse>(
                service, METHODID_RETRIEVE_BASIC_PROFILE)))
        .addMethod(
          getSearchBasicProfileMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.manabie.libs.proto.bob.v1.Users.SearchBasicProfileRequest,
              com.manabie.libs.proto.bob.v1.Users.SearchBasicProfileResponse>(
                service, METHODID_SEARCH_BASIC_PROFILE)))
        .addMethod(
          getCheckProfileMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              com.manabie.libs.proto.bob.v1.Users.CheckProfileRequest,
              com.manabie.libs.proto.bob.v1.Users.CheckProfileResponse>(
                service, METHODID_CHECK_PROFILE)))
        .build();
  }

  private static abstract class UserReaderServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    UserReaderServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return com.manabie.libs.proto.bob.v1.Users.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("UserReaderService");
    }
  }

  private static final class UserReaderServiceFileDescriptorSupplier
      extends UserReaderServiceBaseDescriptorSupplier {
    UserReaderServiceFileDescriptorSupplier() {}
  }

  private static final class UserReaderServiceMethodDescriptorSupplier
      extends UserReaderServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final String methodName;

    UserReaderServiceMethodDescriptorSupplier(String methodName) {
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
      synchronized (UserReaderServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new UserReaderServiceFileDescriptorSupplier())
              .addMethod(getGetCurrentUserProfileMethod())
              .addMethod(getRetrieveTeacherProfilesMethod())
              .addMethod(getRetrieveBasicProfileMethod())
              .addMethod(getSearchBasicProfileMethod())
              .addMethod(getCheckProfileMethod())
              .build();
        }
      }
    }
    return result;
  }
}
