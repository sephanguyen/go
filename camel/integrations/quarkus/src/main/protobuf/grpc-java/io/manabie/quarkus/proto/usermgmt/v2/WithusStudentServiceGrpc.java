package io.manabie.quarkus.proto.usermgmt.v2;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.56.1)",
    comments = "Source: usermgmt/v2/student.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class WithusStudentServiceGrpc {

  private WithusStudentServiceGrpc() {}

  public static final String SERVICE_NAME = "usermgmt.v2.WithusStudentService";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Student.ImportWithusManagaraBaseCSVRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Student.ImportWithusManagaraBaseCSVResponse> getImportWithusManagaraBaseCSVMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ImportWithusManagaraBaseCSV",
      requestType = io.manabie.quarkus.proto.usermgmt.v2.Student.ImportWithusManagaraBaseCSVRequest.class,
      responseType = io.manabie.quarkus.proto.usermgmt.v2.Student.ImportWithusManagaraBaseCSVResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Student.ImportWithusManagaraBaseCSVRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Student.ImportWithusManagaraBaseCSVResponse> getImportWithusManagaraBaseCSVMethod() {
    io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Student.ImportWithusManagaraBaseCSVRequest, io.manabie.quarkus.proto.usermgmt.v2.Student.ImportWithusManagaraBaseCSVResponse> getImportWithusManagaraBaseCSVMethod;
    if ((getImportWithusManagaraBaseCSVMethod = WithusStudentServiceGrpc.getImportWithusManagaraBaseCSVMethod) == null) {
      synchronized (WithusStudentServiceGrpc.class) {
        if ((getImportWithusManagaraBaseCSVMethod = WithusStudentServiceGrpc.getImportWithusManagaraBaseCSVMethod) == null) {
          WithusStudentServiceGrpc.getImportWithusManagaraBaseCSVMethod = getImportWithusManagaraBaseCSVMethod =
              io.grpc.MethodDescriptor.<io.manabie.quarkus.proto.usermgmt.v2.Student.ImportWithusManagaraBaseCSVRequest, io.manabie.quarkus.proto.usermgmt.v2.Student.ImportWithusManagaraBaseCSVResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ImportWithusManagaraBaseCSV"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Student.ImportWithusManagaraBaseCSVRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Student.ImportWithusManagaraBaseCSVResponse.getDefaultInstance()))
              .setSchemaDescriptor(new WithusStudentServiceMethodDescriptorSupplier("ImportWithusManagaraBaseCSV"))
              .build();
        }
      }
    }
    return getImportWithusManagaraBaseCSVMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static WithusStudentServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<WithusStudentServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<WithusStudentServiceStub>() {
        @java.lang.Override
        public WithusStudentServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new WithusStudentServiceStub(channel, callOptions);
        }
      };
    return WithusStudentServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static WithusStudentServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<WithusStudentServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<WithusStudentServiceBlockingStub>() {
        @java.lang.Override
        public WithusStudentServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new WithusStudentServiceBlockingStub(channel, callOptions);
        }
      };
    return WithusStudentServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static WithusStudentServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<WithusStudentServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<WithusStudentServiceFutureStub>() {
        @java.lang.Override
        public WithusStudentServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new WithusStudentServiceFutureStub(channel, callOptions);
        }
      };
    return WithusStudentServiceFutureStub.newStub(factory, channel);
  }

  /**
   */
  public interface AsyncService {

    /**
     */
    default void importWithusManagaraBaseCSV(io.manabie.quarkus.proto.usermgmt.v2.Student.ImportWithusManagaraBaseCSVRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Student.ImportWithusManagaraBaseCSVResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getImportWithusManagaraBaseCSVMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service WithusStudentService.
   */
  public static abstract class WithusStudentServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return WithusStudentServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service WithusStudentService.
   */
  public static final class WithusStudentServiceStub
      extends io.grpc.stub.AbstractAsyncStub<WithusStudentServiceStub> {
    private WithusStudentServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected WithusStudentServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new WithusStudentServiceStub(channel, callOptions);
    }

    /**
     */
    public void importWithusManagaraBaseCSV(io.manabie.quarkus.proto.usermgmt.v2.Student.ImportWithusManagaraBaseCSVRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Student.ImportWithusManagaraBaseCSVResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getImportWithusManagaraBaseCSVMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service WithusStudentService.
   */
  public static final class WithusStudentServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<WithusStudentServiceBlockingStub> {
    private WithusStudentServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected WithusStudentServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new WithusStudentServiceBlockingStub(channel, callOptions);
    }

    /**
     */
    public io.manabie.quarkus.proto.usermgmt.v2.Student.ImportWithusManagaraBaseCSVResponse importWithusManagaraBaseCSV(io.manabie.quarkus.proto.usermgmt.v2.Student.ImportWithusManagaraBaseCSVRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getImportWithusManagaraBaseCSVMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service WithusStudentService.
   */
  public static final class WithusStudentServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<WithusStudentServiceFutureStub> {
    private WithusStudentServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected WithusStudentServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new WithusStudentServiceFutureStub(channel, callOptions);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.quarkus.proto.usermgmt.v2.Student.ImportWithusManagaraBaseCSVResponse> importWithusManagaraBaseCSV(
        io.manabie.quarkus.proto.usermgmt.v2.Student.ImportWithusManagaraBaseCSVRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getImportWithusManagaraBaseCSVMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_IMPORT_WITHUS_MANAGARA_BASE_CSV = 0;

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
        case METHODID_IMPORT_WITHUS_MANAGARA_BASE_CSV:
          serviceImpl.importWithusManagaraBaseCSV((io.manabie.quarkus.proto.usermgmt.v2.Student.ImportWithusManagaraBaseCSVRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Student.ImportWithusManagaraBaseCSVResponse>) responseObserver);
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
          getImportWithusManagaraBaseCSVMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.quarkus.proto.usermgmt.v2.Student.ImportWithusManagaraBaseCSVRequest,
              io.manabie.quarkus.proto.usermgmt.v2.Student.ImportWithusManagaraBaseCSVResponse>(
                service, METHODID_IMPORT_WITHUS_MANAGARA_BASE_CSV)))
        .build();
  }

  private static abstract class WithusStudentServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    WithusStudentServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return io.manabie.quarkus.proto.usermgmt.v2.Student.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("WithusStudentService");
    }
  }

  private static final class WithusStudentServiceFileDescriptorSupplier
      extends WithusStudentServiceBaseDescriptorSupplier {
    WithusStudentServiceFileDescriptorSupplier() {}
  }

  private static final class WithusStudentServiceMethodDescriptorSupplier
      extends WithusStudentServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final String methodName;

    WithusStudentServiceMethodDescriptorSupplier(String methodName) {
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
      synchronized (WithusStudentServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new WithusStudentServiceFileDescriptorSupplier())
              .addMethod(getImportWithusManagaraBaseCSVMethod())
              .build();
        }
      }
    }
    return result;
  }
}
