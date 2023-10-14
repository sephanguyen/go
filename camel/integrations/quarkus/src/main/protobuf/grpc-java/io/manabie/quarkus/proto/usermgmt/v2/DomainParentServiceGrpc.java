package io.manabie.quarkus.proto.usermgmt.v2;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.56.1)",
    comments = "Source: usermgmt/v2/parents.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class DomainParentServiceGrpc {

  private DomainParentServiceGrpc() {}

  public static final String SERVICE_NAME = "usermgmt.v2.DomainParentService";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Parents.UpsertParentRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Parents.UpsertParentResponse> getUpsertParentMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UpsertParent",
      requestType = io.manabie.quarkus.proto.usermgmt.v2.Parents.UpsertParentRequest.class,
      responseType = io.manabie.quarkus.proto.usermgmt.v2.Parents.UpsertParentResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Parents.UpsertParentRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Parents.UpsertParentResponse> getUpsertParentMethod() {
    io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Parents.UpsertParentRequest, io.manabie.quarkus.proto.usermgmt.v2.Parents.UpsertParentResponse> getUpsertParentMethod;
    if ((getUpsertParentMethod = DomainParentServiceGrpc.getUpsertParentMethod) == null) {
      synchronized (DomainParentServiceGrpc.class) {
        if ((getUpsertParentMethod = DomainParentServiceGrpc.getUpsertParentMethod) == null) {
          DomainParentServiceGrpc.getUpsertParentMethod = getUpsertParentMethod =
              io.grpc.MethodDescriptor.<io.manabie.quarkus.proto.usermgmt.v2.Parents.UpsertParentRequest, io.manabie.quarkus.proto.usermgmt.v2.Parents.UpsertParentResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UpsertParent"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Parents.UpsertParentRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Parents.UpsertParentResponse.getDefaultInstance()))
              .setSchemaDescriptor(new DomainParentServiceMethodDescriptorSupplier("UpsertParent"))
              .build();
        }
      }
    }
    return getUpsertParentMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static DomainParentServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<DomainParentServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<DomainParentServiceStub>() {
        @java.lang.Override
        public DomainParentServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new DomainParentServiceStub(channel, callOptions);
        }
      };
    return DomainParentServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static DomainParentServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<DomainParentServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<DomainParentServiceBlockingStub>() {
        @java.lang.Override
        public DomainParentServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new DomainParentServiceBlockingStub(channel, callOptions);
        }
      };
    return DomainParentServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static DomainParentServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<DomainParentServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<DomainParentServiceFutureStub>() {
        @java.lang.Override
        public DomainParentServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new DomainParentServiceFutureStub(channel, callOptions);
        }
      };
    return DomainParentServiceFutureStub.newStub(factory, channel);
  }

  /**
   */
  public interface AsyncService {

    /**
     */
    default void upsertParent(io.manabie.quarkus.proto.usermgmt.v2.Parents.UpsertParentRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Parents.UpsertParentResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUpsertParentMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service DomainParentService.
   */
  public static abstract class DomainParentServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return DomainParentServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service DomainParentService.
   */
  public static final class DomainParentServiceStub
      extends io.grpc.stub.AbstractAsyncStub<DomainParentServiceStub> {
    private DomainParentServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected DomainParentServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new DomainParentServiceStub(channel, callOptions);
    }

    /**
     */
    public void upsertParent(io.manabie.quarkus.proto.usermgmt.v2.Parents.UpsertParentRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Parents.UpsertParentResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUpsertParentMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service DomainParentService.
   */
  public static final class DomainParentServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<DomainParentServiceBlockingStub> {
    private DomainParentServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected DomainParentServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new DomainParentServiceBlockingStub(channel, callOptions);
    }

    /**
     */
    public io.manabie.quarkus.proto.usermgmt.v2.Parents.UpsertParentResponse upsertParent(io.manabie.quarkus.proto.usermgmt.v2.Parents.UpsertParentRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUpsertParentMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service DomainParentService.
   */
  public static final class DomainParentServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<DomainParentServiceFutureStub> {
    private DomainParentServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected DomainParentServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new DomainParentServiceFutureStub(channel, callOptions);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.quarkus.proto.usermgmt.v2.Parents.UpsertParentResponse> upsertParent(
        io.manabie.quarkus.proto.usermgmt.v2.Parents.UpsertParentRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUpsertParentMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_UPSERT_PARENT = 0;

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
        case METHODID_UPSERT_PARENT:
          serviceImpl.upsertParent((io.manabie.quarkus.proto.usermgmt.v2.Parents.UpsertParentRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Parents.UpsertParentResponse>) responseObserver);
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
          getUpsertParentMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.quarkus.proto.usermgmt.v2.Parents.UpsertParentRequest,
              io.manabie.quarkus.proto.usermgmt.v2.Parents.UpsertParentResponse>(
                service, METHODID_UPSERT_PARENT)))
        .build();
  }

  private static abstract class DomainParentServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    DomainParentServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return io.manabie.quarkus.proto.usermgmt.v2.Parents.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("DomainParentService");
    }
  }

  private static final class DomainParentServiceFileDescriptorSupplier
      extends DomainParentServiceBaseDescriptorSupplier {
    DomainParentServiceFileDescriptorSupplier() {}
  }

  private static final class DomainParentServiceMethodDescriptorSupplier
      extends DomainParentServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final String methodName;

    DomainParentServiceMethodDescriptorSupplier(String methodName) {
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
      synchronized (DomainParentServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new DomainParentServiceFileDescriptorSupplier())
              .addMethod(getUpsertParentMethod())
              .build();
        }
      }
    }
    return result;
  }
}
