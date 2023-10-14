package io.manabie.quarkus.proto.usermgmt.v2;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.56.1)",
    comments = "Source: usermgmt/v2/user_groups.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class UserGroupMgmtServiceGrpc {

  private UserGroupMgmtServiceGrpc() {}

  public static final String SERVICE_NAME = "usermgmt.v2.UserGroupMgmtService";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.UserGroups.CreateUserGroupRequest,
      io.manabie.quarkus.proto.usermgmt.v2.UserGroups.CreateUserGroupResponse> getCreateUserGroupMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "CreateUserGroup",
      requestType = io.manabie.quarkus.proto.usermgmt.v2.UserGroups.CreateUserGroupRequest.class,
      responseType = io.manabie.quarkus.proto.usermgmt.v2.UserGroups.CreateUserGroupResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.UserGroups.CreateUserGroupRequest,
      io.manabie.quarkus.proto.usermgmt.v2.UserGroups.CreateUserGroupResponse> getCreateUserGroupMethod() {
    io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.UserGroups.CreateUserGroupRequest, io.manabie.quarkus.proto.usermgmt.v2.UserGroups.CreateUserGroupResponse> getCreateUserGroupMethod;
    if ((getCreateUserGroupMethod = UserGroupMgmtServiceGrpc.getCreateUserGroupMethod) == null) {
      synchronized (UserGroupMgmtServiceGrpc.class) {
        if ((getCreateUserGroupMethod = UserGroupMgmtServiceGrpc.getCreateUserGroupMethod) == null) {
          UserGroupMgmtServiceGrpc.getCreateUserGroupMethod = getCreateUserGroupMethod =
              io.grpc.MethodDescriptor.<io.manabie.quarkus.proto.usermgmt.v2.UserGroups.CreateUserGroupRequest, io.manabie.quarkus.proto.usermgmt.v2.UserGroups.CreateUserGroupResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "CreateUserGroup"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.UserGroups.CreateUserGroupRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.UserGroups.CreateUserGroupResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserGroupMgmtServiceMethodDescriptorSupplier("CreateUserGroup"))
              .build();
        }
      }
    }
    return getCreateUserGroupMethod;
  }

  private static volatile io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.UserGroups.UpdateUserGroupRequest,
      io.manabie.quarkus.proto.usermgmt.v2.UserGroups.UpdateUserGroupResponse> getUpdateUserGroupMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UpdateUserGroup",
      requestType = io.manabie.quarkus.proto.usermgmt.v2.UserGroups.UpdateUserGroupRequest.class,
      responseType = io.manabie.quarkus.proto.usermgmt.v2.UserGroups.UpdateUserGroupResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.UserGroups.UpdateUserGroupRequest,
      io.manabie.quarkus.proto.usermgmt.v2.UserGroups.UpdateUserGroupResponse> getUpdateUserGroupMethod() {
    io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.UserGroups.UpdateUserGroupRequest, io.manabie.quarkus.proto.usermgmt.v2.UserGroups.UpdateUserGroupResponse> getUpdateUserGroupMethod;
    if ((getUpdateUserGroupMethod = UserGroupMgmtServiceGrpc.getUpdateUserGroupMethod) == null) {
      synchronized (UserGroupMgmtServiceGrpc.class) {
        if ((getUpdateUserGroupMethod = UserGroupMgmtServiceGrpc.getUpdateUserGroupMethod) == null) {
          UserGroupMgmtServiceGrpc.getUpdateUserGroupMethod = getUpdateUserGroupMethod =
              io.grpc.MethodDescriptor.<io.manabie.quarkus.proto.usermgmt.v2.UserGroups.UpdateUserGroupRequest, io.manabie.quarkus.proto.usermgmt.v2.UserGroups.UpdateUserGroupResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UpdateUserGroup"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.UserGroups.UpdateUserGroupRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.UserGroups.UpdateUserGroupResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserGroupMgmtServiceMethodDescriptorSupplier("UpdateUserGroup"))
              .build();
        }
      }
    }
    return getUpdateUserGroupMethod;
  }

  private static volatile io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.UserGroups.ValidateUserLoginRequest,
      io.manabie.quarkus.proto.usermgmt.v2.UserGroups.ValidateUserLoginResponse> getValidateUserLoginMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ValidateUserLogin",
      requestType = io.manabie.quarkus.proto.usermgmt.v2.UserGroups.ValidateUserLoginRequest.class,
      responseType = io.manabie.quarkus.proto.usermgmt.v2.UserGroups.ValidateUserLoginResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.UserGroups.ValidateUserLoginRequest,
      io.manabie.quarkus.proto.usermgmt.v2.UserGroups.ValidateUserLoginResponse> getValidateUserLoginMethod() {
    io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.UserGroups.ValidateUserLoginRequest, io.manabie.quarkus.proto.usermgmt.v2.UserGroups.ValidateUserLoginResponse> getValidateUserLoginMethod;
    if ((getValidateUserLoginMethod = UserGroupMgmtServiceGrpc.getValidateUserLoginMethod) == null) {
      synchronized (UserGroupMgmtServiceGrpc.class) {
        if ((getValidateUserLoginMethod = UserGroupMgmtServiceGrpc.getValidateUserLoginMethod) == null) {
          UserGroupMgmtServiceGrpc.getValidateUserLoginMethod = getValidateUserLoginMethod =
              io.grpc.MethodDescriptor.<io.manabie.quarkus.proto.usermgmt.v2.UserGroups.ValidateUserLoginRequest, io.manabie.quarkus.proto.usermgmt.v2.UserGroups.ValidateUserLoginResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ValidateUserLogin"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.UserGroups.ValidateUserLoginRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.UserGroups.ValidateUserLoginResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserGroupMgmtServiceMethodDescriptorSupplier("ValidateUserLogin"))
              .build();
        }
      }
    }
    return getValidateUserLoginMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static UserGroupMgmtServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<UserGroupMgmtServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<UserGroupMgmtServiceStub>() {
        @java.lang.Override
        public UserGroupMgmtServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new UserGroupMgmtServiceStub(channel, callOptions);
        }
      };
    return UserGroupMgmtServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static UserGroupMgmtServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<UserGroupMgmtServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<UserGroupMgmtServiceBlockingStub>() {
        @java.lang.Override
        public UserGroupMgmtServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new UserGroupMgmtServiceBlockingStub(channel, callOptions);
        }
      };
    return UserGroupMgmtServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static UserGroupMgmtServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<UserGroupMgmtServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<UserGroupMgmtServiceFutureStub>() {
        @java.lang.Override
        public UserGroupMgmtServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new UserGroupMgmtServiceFutureStub(channel, callOptions);
        }
      };
    return UserGroupMgmtServiceFutureStub.newStub(factory, channel);
  }

  /**
   */
  public interface AsyncService {

    /**
     */
    default void createUserGroup(io.manabie.quarkus.proto.usermgmt.v2.UserGroups.CreateUserGroupRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.UserGroups.CreateUserGroupResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getCreateUserGroupMethod(), responseObserver);
    }

    /**
     */
    default void updateUserGroup(io.manabie.quarkus.proto.usermgmt.v2.UserGroups.UpdateUserGroupRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.UserGroups.UpdateUserGroupResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUpdateUserGroupMethod(), responseObserver);
    }

    /**
     */
    default void validateUserLogin(io.manabie.quarkus.proto.usermgmt.v2.UserGroups.ValidateUserLoginRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.UserGroups.ValidateUserLoginResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getValidateUserLoginMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service UserGroupMgmtService.
   */
  public static abstract class UserGroupMgmtServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return UserGroupMgmtServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service UserGroupMgmtService.
   */
  public static final class UserGroupMgmtServiceStub
      extends io.grpc.stub.AbstractAsyncStub<UserGroupMgmtServiceStub> {
    private UserGroupMgmtServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected UserGroupMgmtServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new UserGroupMgmtServiceStub(channel, callOptions);
    }

    /**
     */
    public void createUserGroup(io.manabie.quarkus.proto.usermgmt.v2.UserGroups.CreateUserGroupRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.UserGroups.CreateUserGroupResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getCreateUserGroupMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void updateUserGroup(io.manabie.quarkus.proto.usermgmt.v2.UserGroups.UpdateUserGroupRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.UserGroups.UpdateUserGroupResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUpdateUserGroupMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void validateUserLogin(io.manabie.quarkus.proto.usermgmt.v2.UserGroups.ValidateUserLoginRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.UserGroups.ValidateUserLoginResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getValidateUserLoginMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service UserGroupMgmtService.
   */
  public static final class UserGroupMgmtServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<UserGroupMgmtServiceBlockingStub> {
    private UserGroupMgmtServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected UserGroupMgmtServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new UserGroupMgmtServiceBlockingStub(channel, callOptions);
    }

    /**
     */
    public io.manabie.quarkus.proto.usermgmt.v2.UserGroups.CreateUserGroupResponse createUserGroup(io.manabie.quarkus.proto.usermgmt.v2.UserGroups.CreateUserGroupRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getCreateUserGroupMethod(), getCallOptions(), request);
    }

    /**
     */
    public io.manabie.quarkus.proto.usermgmt.v2.UserGroups.UpdateUserGroupResponse updateUserGroup(io.manabie.quarkus.proto.usermgmt.v2.UserGroups.UpdateUserGroupRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUpdateUserGroupMethod(), getCallOptions(), request);
    }

    /**
     */
    public io.manabie.quarkus.proto.usermgmt.v2.UserGroups.ValidateUserLoginResponse validateUserLogin(io.manabie.quarkus.proto.usermgmt.v2.UserGroups.ValidateUserLoginRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getValidateUserLoginMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service UserGroupMgmtService.
   */
  public static final class UserGroupMgmtServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<UserGroupMgmtServiceFutureStub> {
    private UserGroupMgmtServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected UserGroupMgmtServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new UserGroupMgmtServiceFutureStub(channel, callOptions);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.quarkus.proto.usermgmt.v2.UserGroups.CreateUserGroupResponse> createUserGroup(
        io.manabie.quarkus.proto.usermgmt.v2.UserGroups.CreateUserGroupRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getCreateUserGroupMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.quarkus.proto.usermgmt.v2.UserGroups.UpdateUserGroupResponse> updateUserGroup(
        io.manabie.quarkus.proto.usermgmt.v2.UserGroups.UpdateUserGroupRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUpdateUserGroupMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.quarkus.proto.usermgmt.v2.UserGroups.ValidateUserLoginResponse> validateUserLogin(
        io.manabie.quarkus.proto.usermgmt.v2.UserGroups.ValidateUserLoginRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getValidateUserLoginMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_CREATE_USER_GROUP = 0;
  private static final int METHODID_UPDATE_USER_GROUP = 1;
  private static final int METHODID_VALIDATE_USER_LOGIN = 2;

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
        case METHODID_CREATE_USER_GROUP:
          serviceImpl.createUserGroup((io.manabie.quarkus.proto.usermgmt.v2.UserGroups.CreateUserGroupRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.UserGroups.CreateUserGroupResponse>) responseObserver);
          break;
        case METHODID_UPDATE_USER_GROUP:
          serviceImpl.updateUserGroup((io.manabie.quarkus.proto.usermgmt.v2.UserGroups.UpdateUserGroupRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.UserGroups.UpdateUserGroupResponse>) responseObserver);
          break;
        case METHODID_VALIDATE_USER_LOGIN:
          serviceImpl.validateUserLogin((io.manabie.quarkus.proto.usermgmt.v2.UserGroups.ValidateUserLoginRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.UserGroups.ValidateUserLoginResponse>) responseObserver);
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
          getCreateUserGroupMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.quarkus.proto.usermgmt.v2.UserGroups.CreateUserGroupRequest,
              io.manabie.quarkus.proto.usermgmt.v2.UserGroups.CreateUserGroupResponse>(
                service, METHODID_CREATE_USER_GROUP)))
        .addMethod(
          getUpdateUserGroupMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.quarkus.proto.usermgmt.v2.UserGroups.UpdateUserGroupRequest,
              io.manabie.quarkus.proto.usermgmt.v2.UserGroups.UpdateUserGroupResponse>(
                service, METHODID_UPDATE_USER_GROUP)))
        .addMethod(
          getValidateUserLoginMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.quarkus.proto.usermgmt.v2.UserGroups.ValidateUserLoginRequest,
              io.manabie.quarkus.proto.usermgmt.v2.UserGroups.ValidateUserLoginResponse>(
                service, METHODID_VALIDATE_USER_LOGIN)))
        .build();
  }

  private static abstract class UserGroupMgmtServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    UserGroupMgmtServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return io.manabie.quarkus.proto.usermgmt.v2.UserGroups.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("UserGroupMgmtService");
    }
  }

  private static final class UserGroupMgmtServiceFileDescriptorSupplier
      extends UserGroupMgmtServiceBaseDescriptorSupplier {
    UserGroupMgmtServiceFileDescriptorSupplier() {}
  }

  private static final class UserGroupMgmtServiceMethodDescriptorSupplier
      extends UserGroupMgmtServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final String methodName;

    UserGroupMgmtServiceMethodDescriptorSupplier(String methodName) {
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
      synchronized (UserGroupMgmtServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new UserGroupMgmtServiceFileDescriptorSupplier())
              .addMethod(getCreateUserGroupMethod())
              .addMethod(getUpdateUserGroupMethod())
              .addMethod(getValidateUserLoginMethod())
              .build();
        }
      }
    }
    return result;
  }
}
