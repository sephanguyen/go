package io.manabie.quarkus.proto.usermgmt.v2;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.56.1)",
    comments = "Source: usermgmt/v2/users.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class UserReaderServiceGrpc {

  private UserReaderServiceGrpc() {}

  public static final String SERVICE_NAME = "usermgmt.v2.UserReaderService";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.SearchBasicProfileRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Users.SearchBasicProfileResponse> getSearchBasicProfileMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "SearchBasicProfile",
      requestType = io.manabie.quarkus.proto.usermgmt.v2.Users.SearchBasicProfileRequest.class,
      responseType = io.manabie.quarkus.proto.usermgmt.v2.Users.SearchBasicProfileResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.SearchBasicProfileRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Users.SearchBasicProfileResponse> getSearchBasicProfileMethod() {
    io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.SearchBasicProfileRequest, io.manabie.quarkus.proto.usermgmt.v2.Users.SearchBasicProfileResponse> getSearchBasicProfileMethod;
    if ((getSearchBasicProfileMethod = UserReaderServiceGrpc.getSearchBasicProfileMethod) == null) {
      synchronized (UserReaderServiceGrpc.class) {
        if ((getSearchBasicProfileMethod = UserReaderServiceGrpc.getSearchBasicProfileMethod) == null) {
          UserReaderServiceGrpc.getSearchBasicProfileMethod = getSearchBasicProfileMethod =
              io.grpc.MethodDescriptor.<io.manabie.quarkus.proto.usermgmt.v2.Users.SearchBasicProfileRequest, io.manabie.quarkus.proto.usermgmt.v2.Users.SearchBasicProfileResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "SearchBasicProfile"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Users.SearchBasicProfileRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Users.SearchBasicProfileResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserReaderServiceMethodDescriptorSupplier("SearchBasicProfile"))
              .build();
        }
      }
    }
    return getSearchBasicProfileMethod;
  }

  private static volatile io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.RetrieveStudentAssociatedToParentAccountRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Users.RetrieveStudentAssociatedToParentAccountResponse> getRetrieveStudentAssociatedToParentAccountMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "RetrieveStudentAssociatedToParentAccount",
      requestType = io.manabie.quarkus.proto.usermgmt.v2.Users.RetrieveStudentAssociatedToParentAccountRequest.class,
      responseType = io.manabie.quarkus.proto.usermgmt.v2.Users.RetrieveStudentAssociatedToParentAccountResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.RetrieveStudentAssociatedToParentAccountRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Users.RetrieveStudentAssociatedToParentAccountResponse> getRetrieveStudentAssociatedToParentAccountMethod() {
    io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.RetrieveStudentAssociatedToParentAccountRequest, io.manabie.quarkus.proto.usermgmt.v2.Users.RetrieveStudentAssociatedToParentAccountResponse> getRetrieveStudentAssociatedToParentAccountMethod;
    if ((getRetrieveStudentAssociatedToParentAccountMethod = UserReaderServiceGrpc.getRetrieveStudentAssociatedToParentAccountMethod) == null) {
      synchronized (UserReaderServiceGrpc.class) {
        if ((getRetrieveStudentAssociatedToParentAccountMethod = UserReaderServiceGrpc.getRetrieveStudentAssociatedToParentAccountMethod) == null) {
          UserReaderServiceGrpc.getRetrieveStudentAssociatedToParentAccountMethod = getRetrieveStudentAssociatedToParentAccountMethod =
              io.grpc.MethodDescriptor.<io.manabie.quarkus.proto.usermgmt.v2.Users.RetrieveStudentAssociatedToParentAccountRequest, io.manabie.quarkus.proto.usermgmt.v2.Users.RetrieveStudentAssociatedToParentAccountResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "RetrieveStudentAssociatedToParentAccount"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Users.RetrieveStudentAssociatedToParentAccountRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Users.RetrieveStudentAssociatedToParentAccountResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserReaderServiceMethodDescriptorSupplier("RetrieveStudentAssociatedToParentAccount"))
              .build();
        }
      }
    }
    return getRetrieveStudentAssociatedToParentAccountMethod;
  }

  private static volatile io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.GetBasicProfileRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Users.GetBasicProfileResponse> getGetBasicProfileMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetBasicProfile",
      requestType = io.manabie.quarkus.proto.usermgmt.v2.Users.GetBasicProfileRequest.class,
      responseType = io.manabie.quarkus.proto.usermgmt.v2.Users.GetBasicProfileResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.GetBasicProfileRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Users.GetBasicProfileResponse> getGetBasicProfileMethod() {
    io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.GetBasicProfileRequest, io.manabie.quarkus.proto.usermgmt.v2.Users.GetBasicProfileResponse> getGetBasicProfileMethod;
    if ((getGetBasicProfileMethod = UserReaderServiceGrpc.getGetBasicProfileMethod) == null) {
      synchronized (UserReaderServiceGrpc.class) {
        if ((getGetBasicProfileMethod = UserReaderServiceGrpc.getGetBasicProfileMethod) == null) {
          UserReaderServiceGrpc.getGetBasicProfileMethod = getGetBasicProfileMethod =
              io.grpc.MethodDescriptor.<io.manabie.quarkus.proto.usermgmt.v2.Users.GetBasicProfileRequest, io.manabie.quarkus.proto.usermgmt.v2.Users.GetBasicProfileResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetBasicProfile"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Users.GetBasicProfileRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Users.GetBasicProfileResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserReaderServiceMethodDescriptorSupplier("GetBasicProfile"))
              .build();
        }
      }
    }
    return getGetBasicProfileMethod;
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
    default void searchBasicProfile(io.manabie.quarkus.proto.usermgmt.v2.Users.SearchBasicProfileRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.SearchBasicProfileResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getSearchBasicProfileMethod(), responseObserver);
    }

    /**
     */
    default void retrieveStudentAssociatedToParentAccount(io.manabie.quarkus.proto.usermgmt.v2.Users.RetrieveStudentAssociatedToParentAccountRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.RetrieveStudentAssociatedToParentAccountResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getRetrieveStudentAssociatedToParentAccountMethod(), responseObserver);
    }

    /**
     */
    default void getBasicProfile(io.manabie.quarkus.proto.usermgmt.v2.Users.GetBasicProfileRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.GetBasicProfileResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetBasicProfileMethod(), responseObserver);
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
    public void searchBasicProfile(io.manabie.quarkus.proto.usermgmt.v2.Users.SearchBasicProfileRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.SearchBasicProfileResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getSearchBasicProfileMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void retrieveStudentAssociatedToParentAccount(io.manabie.quarkus.proto.usermgmt.v2.Users.RetrieveStudentAssociatedToParentAccountRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.RetrieveStudentAssociatedToParentAccountResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getRetrieveStudentAssociatedToParentAccountMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void getBasicProfile(io.manabie.quarkus.proto.usermgmt.v2.Users.GetBasicProfileRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.GetBasicProfileResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetBasicProfileMethod(), getCallOptions()), request, responseObserver);
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
    public io.manabie.quarkus.proto.usermgmt.v2.Users.SearchBasicProfileResponse searchBasicProfile(io.manabie.quarkus.proto.usermgmt.v2.Users.SearchBasicProfileRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getSearchBasicProfileMethod(), getCallOptions(), request);
    }

    /**
     */
    public io.manabie.quarkus.proto.usermgmt.v2.Users.RetrieveStudentAssociatedToParentAccountResponse retrieveStudentAssociatedToParentAccount(io.manabie.quarkus.proto.usermgmt.v2.Users.RetrieveStudentAssociatedToParentAccountRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getRetrieveStudentAssociatedToParentAccountMethod(), getCallOptions(), request);
    }

    /**
     */
    public io.manabie.quarkus.proto.usermgmt.v2.Users.GetBasicProfileResponse getBasicProfile(io.manabie.quarkus.proto.usermgmt.v2.Users.GetBasicProfileRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetBasicProfileMethod(), getCallOptions(), request);
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
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.quarkus.proto.usermgmt.v2.Users.SearchBasicProfileResponse> searchBasicProfile(
        io.manabie.quarkus.proto.usermgmt.v2.Users.SearchBasicProfileRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getSearchBasicProfileMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.quarkus.proto.usermgmt.v2.Users.RetrieveStudentAssociatedToParentAccountResponse> retrieveStudentAssociatedToParentAccount(
        io.manabie.quarkus.proto.usermgmt.v2.Users.RetrieveStudentAssociatedToParentAccountRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getRetrieveStudentAssociatedToParentAccountMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.quarkus.proto.usermgmt.v2.Users.GetBasicProfileResponse> getBasicProfile(
        io.manabie.quarkus.proto.usermgmt.v2.Users.GetBasicProfileRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetBasicProfileMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_SEARCH_BASIC_PROFILE = 0;
  private static final int METHODID_RETRIEVE_STUDENT_ASSOCIATED_TO_PARENT_ACCOUNT = 1;
  private static final int METHODID_GET_BASIC_PROFILE = 2;

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
        case METHODID_SEARCH_BASIC_PROFILE:
          serviceImpl.searchBasicProfile((io.manabie.quarkus.proto.usermgmt.v2.Users.SearchBasicProfileRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.SearchBasicProfileResponse>) responseObserver);
          break;
        case METHODID_RETRIEVE_STUDENT_ASSOCIATED_TO_PARENT_ACCOUNT:
          serviceImpl.retrieveStudentAssociatedToParentAccount((io.manabie.quarkus.proto.usermgmt.v2.Users.RetrieveStudentAssociatedToParentAccountRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.RetrieveStudentAssociatedToParentAccountResponse>) responseObserver);
          break;
        case METHODID_GET_BASIC_PROFILE:
          serviceImpl.getBasicProfile((io.manabie.quarkus.proto.usermgmt.v2.Users.GetBasicProfileRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.GetBasicProfileResponse>) responseObserver);
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
          getSearchBasicProfileMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.quarkus.proto.usermgmt.v2.Users.SearchBasicProfileRequest,
              io.manabie.quarkus.proto.usermgmt.v2.Users.SearchBasicProfileResponse>(
                service, METHODID_SEARCH_BASIC_PROFILE)))
        .addMethod(
          getRetrieveStudentAssociatedToParentAccountMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.quarkus.proto.usermgmt.v2.Users.RetrieveStudentAssociatedToParentAccountRequest,
              io.manabie.quarkus.proto.usermgmt.v2.Users.RetrieveStudentAssociatedToParentAccountResponse>(
                service, METHODID_RETRIEVE_STUDENT_ASSOCIATED_TO_PARENT_ACCOUNT)))
        .addMethod(
          getGetBasicProfileMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.quarkus.proto.usermgmt.v2.Users.GetBasicProfileRequest,
              io.manabie.quarkus.proto.usermgmt.v2.Users.GetBasicProfileResponse>(
                service, METHODID_GET_BASIC_PROFILE)))
        .build();
  }

  private static abstract class UserReaderServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    UserReaderServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return io.manabie.quarkus.proto.usermgmt.v2.Users.getDescriptor();
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
              .addMethod(getSearchBasicProfileMethod())
              .addMethod(getRetrieveStudentAssociatedToParentAccountMethod())
              .addMethod(getGetBasicProfileMethod())
              .build();
        }
      }
    }
    return result;
  }
}
