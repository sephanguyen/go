package io.manabie.quarkus.proto.usermgmt.v2;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.56.1)",
    comments = "Source: usermgmt/v2/school_info.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class SchoolInfoServiceGrpc {

  private SchoolInfoServiceGrpc() {}

  public static final String SERVICE_NAME = "usermgmt.v2.SchoolInfoService";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.SchoolInfo.ImportSchoolInfoRequest,
      io.manabie.quarkus.proto.usermgmt.v2.SchoolInfo.ImportSchoolInfoResponse> getImportSchoolInfoMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ImportSchoolInfo",
      requestType = io.manabie.quarkus.proto.usermgmt.v2.SchoolInfo.ImportSchoolInfoRequest.class,
      responseType = io.manabie.quarkus.proto.usermgmt.v2.SchoolInfo.ImportSchoolInfoResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.SchoolInfo.ImportSchoolInfoRequest,
      io.manabie.quarkus.proto.usermgmt.v2.SchoolInfo.ImportSchoolInfoResponse> getImportSchoolInfoMethod() {
    io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.SchoolInfo.ImportSchoolInfoRequest, io.manabie.quarkus.proto.usermgmt.v2.SchoolInfo.ImportSchoolInfoResponse> getImportSchoolInfoMethod;
    if ((getImportSchoolInfoMethod = SchoolInfoServiceGrpc.getImportSchoolInfoMethod) == null) {
      synchronized (SchoolInfoServiceGrpc.class) {
        if ((getImportSchoolInfoMethod = SchoolInfoServiceGrpc.getImportSchoolInfoMethod) == null) {
          SchoolInfoServiceGrpc.getImportSchoolInfoMethod = getImportSchoolInfoMethod =
              io.grpc.MethodDescriptor.<io.manabie.quarkus.proto.usermgmt.v2.SchoolInfo.ImportSchoolInfoRequest, io.manabie.quarkus.proto.usermgmt.v2.SchoolInfo.ImportSchoolInfoResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ImportSchoolInfo"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.SchoolInfo.ImportSchoolInfoRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.SchoolInfo.ImportSchoolInfoResponse.getDefaultInstance()))
              .setSchemaDescriptor(new SchoolInfoServiceMethodDescriptorSupplier("ImportSchoolInfo"))
              .build();
        }
      }
    }
    return getImportSchoolInfoMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static SchoolInfoServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<SchoolInfoServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<SchoolInfoServiceStub>() {
        @java.lang.Override
        public SchoolInfoServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new SchoolInfoServiceStub(channel, callOptions);
        }
      };
    return SchoolInfoServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static SchoolInfoServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<SchoolInfoServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<SchoolInfoServiceBlockingStub>() {
        @java.lang.Override
        public SchoolInfoServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new SchoolInfoServiceBlockingStub(channel, callOptions);
        }
      };
    return SchoolInfoServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static SchoolInfoServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<SchoolInfoServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<SchoolInfoServiceFutureStub>() {
        @java.lang.Override
        public SchoolInfoServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new SchoolInfoServiceFutureStub(channel, callOptions);
        }
      };
    return SchoolInfoServiceFutureStub.newStub(factory, channel);
  }

  /**
   */
  public interface AsyncService {

    /**
     */
    default void importSchoolInfo(io.manabie.quarkus.proto.usermgmt.v2.SchoolInfo.ImportSchoolInfoRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.SchoolInfo.ImportSchoolInfoResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getImportSchoolInfoMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service SchoolInfoService.
   */
  public static abstract class SchoolInfoServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return SchoolInfoServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service SchoolInfoService.
   */
  public static final class SchoolInfoServiceStub
      extends io.grpc.stub.AbstractAsyncStub<SchoolInfoServiceStub> {
    private SchoolInfoServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected SchoolInfoServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new SchoolInfoServiceStub(channel, callOptions);
    }

    /**
     */
    public void importSchoolInfo(io.manabie.quarkus.proto.usermgmt.v2.SchoolInfo.ImportSchoolInfoRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.SchoolInfo.ImportSchoolInfoResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getImportSchoolInfoMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service SchoolInfoService.
   */
  public static final class SchoolInfoServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<SchoolInfoServiceBlockingStub> {
    private SchoolInfoServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected SchoolInfoServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new SchoolInfoServiceBlockingStub(channel, callOptions);
    }

    /**
     */
    public io.manabie.quarkus.proto.usermgmt.v2.SchoolInfo.ImportSchoolInfoResponse importSchoolInfo(io.manabie.quarkus.proto.usermgmt.v2.SchoolInfo.ImportSchoolInfoRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getImportSchoolInfoMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service SchoolInfoService.
   */
  public static final class SchoolInfoServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<SchoolInfoServiceFutureStub> {
    private SchoolInfoServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected SchoolInfoServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new SchoolInfoServiceFutureStub(channel, callOptions);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.quarkus.proto.usermgmt.v2.SchoolInfo.ImportSchoolInfoResponse> importSchoolInfo(
        io.manabie.quarkus.proto.usermgmt.v2.SchoolInfo.ImportSchoolInfoRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getImportSchoolInfoMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_IMPORT_SCHOOL_INFO = 0;

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
        case METHODID_IMPORT_SCHOOL_INFO:
          serviceImpl.importSchoolInfo((io.manabie.quarkus.proto.usermgmt.v2.SchoolInfo.ImportSchoolInfoRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.SchoolInfo.ImportSchoolInfoResponse>) responseObserver);
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
          getImportSchoolInfoMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.quarkus.proto.usermgmt.v2.SchoolInfo.ImportSchoolInfoRequest,
              io.manabie.quarkus.proto.usermgmt.v2.SchoolInfo.ImportSchoolInfoResponse>(
                service, METHODID_IMPORT_SCHOOL_INFO)))
        .build();
  }

  private static abstract class SchoolInfoServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    SchoolInfoServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return io.manabie.quarkus.proto.usermgmt.v2.SchoolInfo.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("SchoolInfoService");
    }
  }

  private static final class SchoolInfoServiceFileDescriptorSupplier
      extends SchoolInfoServiceBaseDescriptorSupplier {
    SchoolInfoServiceFileDescriptorSupplier() {}
  }

  private static final class SchoolInfoServiceMethodDescriptorSupplier
      extends SchoolInfoServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final String methodName;

    SchoolInfoServiceMethodDescriptorSupplier(String methodName) {
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
      synchronized (SchoolInfoServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new SchoolInfoServiceFileDescriptorSupplier())
              .addMethod(getImportSchoolInfoMethod())
              .build();
        }
      }
    }
    return result;
  }
}
