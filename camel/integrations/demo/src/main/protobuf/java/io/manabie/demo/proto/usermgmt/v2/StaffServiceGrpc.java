package io.manabie.demo.proto.usermgmt.v2;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.56.1)",
    comments = "Source: usermgmt/v2/users.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class StaffServiceGrpc {

  private StaffServiceGrpc() {}

  public static final String SERVICE_NAME = "usermgmt.v2.StaffService";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Users.CreateStaffRequest,
      io.manabie.demo.proto.usermgmt.v2.Users.CreateStaffResponse> getCreateStaffMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "CreateStaff",
      requestType = io.manabie.demo.proto.usermgmt.v2.Users.CreateStaffRequest.class,
      responseType = io.manabie.demo.proto.usermgmt.v2.Users.CreateStaffResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Users.CreateStaffRequest,
      io.manabie.demo.proto.usermgmt.v2.Users.CreateStaffResponse> getCreateStaffMethod() {
    io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Users.CreateStaffRequest, io.manabie.demo.proto.usermgmt.v2.Users.CreateStaffResponse> getCreateStaffMethod;
    if ((getCreateStaffMethod = StaffServiceGrpc.getCreateStaffMethod) == null) {
      synchronized (StaffServiceGrpc.class) {
        if ((getCreateStaffMethod = StaffServiceGrpc.getCreateStaffMethod) == null) {
          StaffServiceGrpc.getCreateStaffMethod = getCreateStaffMethod =
              io.grpc.MethodDescriptor.<io.manabie.demo.proto.usermgmt.v2.Users.CreateStaffRequest, io.manabie.demo.proto.usermgmt.v2.Users.CreateStaffResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "CreateStaff"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.demo.proto.usermgmt.v2.Users.CreateStaffRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.demo.proto.usermgmt.v2.Users.CreateStaffResponse.getDefaultInstance()))
              .setSchemaDescriptor(new StaffServiceMethodDescriptorSupplier("CreateStaff"))
              .build();
        }
      }
    }
    return getCreateStaffMethod;
  }

  private static volatile io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffRequest,
      io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffResponse> getUpdateStaffMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UpdateStaff",
      requestType = io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffRequest.class,
      responseType = io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffRequest,
      io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffResponse> getUpdateStaffMethod() {
    io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffRequest, io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffResponse> getUpdateStaffMethod;
    if ((getUpdateStaffMethod = StaffServiceGrpc.getUpdateStaffMethod) == null) {
      synchronized (StaffServiceGrpc.class) {
        if ((getUpdateStaffMethod = StaffServiceGrpc.getUpdateStaffMethod) == null) {
          StaffServiceGrpc.getUpdateStaffMethod = getUpdateStaffMethod =
              io.grpc.MethodDescriptor.<io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffRequest, io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UpdateStaff"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffResponse.getDefaultInstance()))
              .setSchemaDescriptor(new StaffServiceMethodDescriptorSupplier("UpdateStaff"))
              .build();
        }
      }
    }
    return getUpdateStaffMethod;
  }

  private static volatile io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffSettingRequest,
      io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffSettingResponse> getUpdateStaffSettingMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UpdateStaffSetting",
      requestType = io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffSettingRequest.class,
      responseType = io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffSettingResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffSettingRequest,
      io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffSettingResponse> getUpdateStaffSettingMethod() {
    io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffSettingRequest, io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffSettingResponse> getUpdateStaffSettingMethod;
    if ((getUpdateStaffSettingMethod = StaffServiceGrpc.getUpdateStaffSettingMethod) == null) {
      synchronized (StaffServiceGrpc.class) {
        if ((getUpdateStaffSettingMethod = StaffServiceGrpc.getUpdateStaffSettingMethod) == null) {
          StaffServiceGrpc.getUpdateStaffSettingMethod = getUpdateStaffSettingMethod =
              io.grpc.MethodDescriptor.<io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffSettingRequest, io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffSettingResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UpdateStaffSetting"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffSettingRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffSettingResponse.getDefaultInstance()))
              .setSchemaDescriptor(new StaffServiceMethodDescriptorSupplier("UpdateStaffSetting"))
              .build();
        }
      }
    }
    return getUpdateStaffSettingMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static StaffServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<StaffServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<StaffServiceStub>() {
        @java.lang.Override
        public StaffServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new StaffServiceStub(channel, callOptions);
        }
      };
    return StaffServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static StaffServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<StaffServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<StaffServiceBlockingStub>() {
        @java.lang.Override
        public StaffServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new StaffServiceBlockingStub(channel, callOptions);
        }
      };
    return StaffServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static StaffServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<StaffServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<StaffServiceFutureStub>() {
        @java.lang.Override
        public StaffServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new StaffServiceFutureStub(channel, callOptions);
        }
      };
    return StaffServiceFutureStub.newStub(factory, channel);
  }

  /**
   */
  public interface AsyncService {

    /**
     */
    default void createStaff(io.manabie.demo.proto.usermgmt.v2.Users.CreateStaffRequest request,
        io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Users.CreateStaffResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getCreateStaffMethod(), responseObserver);
    }

    /**
     */
    default void updateStaff(io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffRequest request,
        io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUpdateStaffMethod(), responseObserver);
    }

    /**
     */
    default void updateStaffSetting(io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffSettingRequest request,
        io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffSettingResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUpdateStaffSettingMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service StaffService.
   */
  public static abstract class StaffServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return StaffServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service StaffService.
   */
  public static final class StaffServiceStub
      extends io.grpc.stub.AbstractAsyncStub<StaffServiceStub> {
    private StaffServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected StaffServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new StaffServiceStub(channel, callOptions);
    }

    /**
     */
    public void createStaff(io.manabie.demo.proto.usermgmt.v2.Users.CreateStaffRequest request,
        io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Users.CreateStaffResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getCreateStaffMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void updateStaff(io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffRequest request,
        io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUpdateStaffMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void updateStaffSetting(io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffSettingRequest request,
        io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffSettingResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUpdateStaffSettingMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service StaffService.
   */
  public static final class StaffServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<StaffServiceBlockingStub> {
    private StaffServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected StaffServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new StaffServiceBlockingStub(channel, callOptions);
    }

    /**
     */
    public io.manabie.demo.proto.usermgmt.v2.Users.CreateStaffResponse createStaff(io.manabie.demo.proto.usermgmt.v2.Users.CreateStaffRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getCreateStaffMethod(), getCallOptions(), request);
    }

    /**
     */
    public io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffResponse updateStaff(io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUpdateStaffMethod(), getCallOptions(), request);
    }

    /**
     */
    public io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffSettingResponse updateStaffSetting(io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffSettingRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUpdateStaffSettingMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service StaffService.
   */
  public static final class StaffServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<StaffServiceFutureStub> {
    private StaffServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected StaffServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new StaffServiceFutureStub(channel, callOptions);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.demo.proto.usermgmt.v2.Users.CreateStaffResponse> createStaff(
        io.manabie.demo.proto.usermgmt.v2.Users.CreateStaffRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getCreateStaffMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffResponse> updateStaff(
        io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUpdateStaffMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffSettingResponse> updateStaffSetting(
        io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffSettingRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUpdateStaffSettingMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_CREATE_STAFF = 0;
  private static final int METHODID_UPDATE_STAFF = 1;
  private static final int METHODID_UPDATE_STAFF_SETTING = 2;

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
        case METHODID_CREATE_STAFF:
          serviceImpl.createStaff((io.manabie.demo.proto.usermgmt.v2.Users.CreateStaffRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Users.CreateStaffResponse>) responseObserver);
          break;
        case METHODID_UPDATE_STAFF:
          serviceImpl.updateStaff((io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffResponse>) responseObserver);
          break;
        case METHODID_UPDATE_STAFF_SETTING:
          serviceImpl.updateStaffSetting((io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffSettingRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffSettingResponse>) responseObserver);
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
          getCreateStaffMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.demo.proto.usermgmt.v2.Users.CreateStaffRequest,
              io.manabie.demo.proto.usermgmt.v2.Users.CreateStaffResponse>(
                service, METHODID_CREATE_STAFF)))
        .addMethod(
          getUpdateStaffMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffRequest,
              io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffResponse>(
                service, METHODID_UPDATE_STAFF)))
        .addMethod(
          getUpdateStaffSettingMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffSettingRequest,
              io.manabie.demo.proto.usermgmt.v2.Users.UpdateStaffSettingResponse>(
                service, METHODID_UPDATE_STAFF_SETTING)))
        .build();
  }

  private static abstract class StaffServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    StaffServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return io.manabie.demo.proto.usermgmt.v2.Users.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("StaffService");
    }
  }

  private static final class StaffServiceFileDescriptorSupplier
      extends StaffServiceBaseDescriptorSupplier {
    StaffServiceFileDescriptorSupplier() {}
  }

  private static final class StaffServiceMethodDescriptorSupplier
      extends StaffServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final String methodName;

    StaffServiceMethodDescriptorSupplier(String methodName) {
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
      synchronized (StaffServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new StaffServiceFileDescriptorSupplier())
              .addMethod(getCreateStaffMethod())
              .addMethod(getUpdateStaffMethod())
              .addMethod(getUpdateStaffSettingMethod())
              .build();
        }
      }
    }
    return result;
  }
}
