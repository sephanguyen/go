package io.manabie.demo.proto.usermgmt.v2;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.56.1)",
    comments = "Source: usermgmt/v2/student.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class StudentServiceGrpc {

  private StudentServiceGrpc() {}

  public static final String SERVICE_NAME = "usermgmt.v2.StudentService";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Student.GetStudentProfileRequest,
      io.manabie.demo.proto.usermgmt.v2.Student.GetStudentProfileResponse> getGetStudentProfileMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetStudentProfile",
      requestType = io.manabie.demo.proto.usermgmt.v2.Student.GetStudentProfileRequest.class,
      responseType = io.manabie.demo.proto.usermgmt.v2.Student.GetStudentProfileResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Student.GetStudentProfileRequest,
      io.manabie.demo.proto.usermgmt.v2.Student.GetStudentProfileResponse> getGetStudentProfileMethod() {
    io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Student.GetStudentProfileRequest, io.manabie.demo.proto.usermgmt.v2.Student.GetStudentProfileResponse> getGetStudentProfileMethod;
    if ((getGetStudentProfileMethod = StudentServiceGrpc.getGetStudentProfileMethod) == null) {
      synchronized (StudentServiceGrpc.class) {
        if ((getGetStudentProfileMethod = StudentServiceGrpc.getGetStudentProfileMethod) == null) {
          StudentServiceGrpc.getGetStudentProfileMethod = getGetStudentProfileMethod =
              io.grpc.MethodDescriptor.<io.manabie.demo.proto.usermgmt.v2.Student.GetStudentProfileRequest, io.manabie.demo.proto.usermgmt.v2.Student.GetStudentProfileResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetStudentProfile"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.demo.proto.usermgmt.v2.Student.GetStudentProfileRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.demo.proto.usermgmt.v2.Student.GetStudentProfileResponse.getDefaultInstance()))
              .setSchemaDescriptor(new StudentServiceMethodDescriptorSupplier("GetStudentProfile"))
              .build();
        }
      }
    }
    return getGetStudentProfileMethod;
  }

  private static volatile io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentCommentRequest,
      io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentCommentResponse> getUpsertStudentCommentMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UpsertStudentComment",
      requestType = io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentCommentRequest.class,
      responseType = io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentCommentResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentCommentRequest,
      io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentCommentResponse> getUpsertStudentCommentMethod() {
    io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentCommentRequest, io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentCommentResponse> getUpsertStudentCommentMethod;
    if ((getUpsertStudentCommentMethod = StudentServiceGrpc.getUpsertStudentCommentMethod) == null) {
      synchronized (StudentServiceGrpc.class) {
        if ((getUpsertStudentCommentMethod = StudentServiceGrpc.getUpsertStudentCommentMethod) == null) {
          StudentServiceGrpc.getUpsertStudentCommentMethod = getUpsertStudentCommentMethod =
              io.grpc.MethodDescriptor.<io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentCommentRequest, io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentCommentResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UpsertStudentComment"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentCommentRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentCommentResponse.getDefaultInstance()))
              .setSchemaDescriptor(new StudentServiceMethodDescriptorSupplier("UpsertStudentComment"))
              .build();
        }
      }
    }
    return getUpsertStudentCommentMethod;
  }

  private static volatile io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Student.DeleteStudentCommentsRequest,
      io.manabie.demo.proto.usermgmt.v2.Student.DeleteStudentCommentsResponse> getDeleteStudentCommentsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "DeleteStudentComments",
      requestType = io.manabie.demo.proto.usermgmt.v2.Student.DeleteStudentCommentsRequest.class,
      responseType = io.manabie.demo.proto.usermgmt.v2.Student.DeleteStudentCommentsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Student.DeleteStudentCommentsRequest,
      io.manabie.demo.proto.usermgmt.v2.Student.DeleteStudentCommentsResponse> getDeleteStudentCommentsMethod() {
    io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Student.DeleteStudentCommentsRequest, io.manabie.demo.proto.usermgmt.v2.Student.DeleteStudentCommentsResponse> getDeleteStudentCommentsMethod;
    if ((getDeleteStudentCommentsMethod = StudentServiceGrpc.getDeleteStudentCommentsMethod) == null) {
      synchronized (StudentServiceGrpc.class) {
        if ((getDeleteStudentCommentsMethod = StudentServiceGrpc.getDeleteStudentCommentsMethod) == null) {
          StudentServiceGrpc.getDeleteStudentCommentsMethod = getDeleteStudentCommentsMethod =
              io.grpc.MethodDescriptor.<io.manabie.demo.proto.usermgmt.v2.Student.DeleteStudentCommentsRequest, io.manabie.demo.proto.usermgmt.v2.Student.DeleteStudentCommentsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "DeleteStudentComments"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.demo.proto.usermgmt.v2.Student.DeleteStudentCommentsRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.demo.proto.usermgmt.v2.Student.DeleteStudentCommentsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new StudentServiceMethodDescriptorSupplier("DeleteStudentComments"))
              .build();
        }
      }
    }
    return getDeleteStudentCommentsMethod;
  }

  private static volatile io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Student.GenerateImportStudentTemplateRequest,
      io.manabie.demo.proto.usermgmt.v2.Student.GenerateImportStudentTemplateResponse> getGenerateImportStudentTemplateMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GenerateImportStudentTemplate",
      requestType = io.manabie.demo.proto.usermgmt.v2.Student.GenerateImportStudentTemplateRequest.class,
      responseType = io.manabie.demo.proto.usermgmt.v2.Student.GenerateImportStudentTemplateResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Student.GenerateImportStudentTemplateRequest,
      io.manabie.demo.proto.usermgmt.v2.Student.GenerateImportStudentTemplateResponse> getGenerateImportStudentTemplateMethod() {
    io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Student.GenerateImportStudentTemplateRequest, io.manabie.demo.proto.usermgmt.v2.Student.GenerateImportStudentTemplateResponse> getGenerateImportStudentTemplateMethod;
    if ((getGenerateImportStudentTemplateMethod = StudentServiceGrpc.getGenerateImportStudentTemplateMethod) == null) {
      synchronized (StudentServiceGrpc.class) {
        if ((getGenerateImportStudentTemplateMethod = StudentServiceGrpc.getGenerateImportStudentTemplateMethod) == null) {
          StudentServiceGrpc.getGenerateImportStudentTemplateMethod = getGenerateImportStudentTemplateMethod =
              io.grpc.MethodDescriptor.<io.manabie.demo.proto.usermgmt.v2.Student.GenerateImportStudentTemplateRequest, io.manabie.demo.proto.usermgmt.v2.Student.GenerateImportStudentTemplateResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GenerateImportStudentTemplate"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.demo.proto.usermgmt.v2.Student.GenerateImportStudentTemplateRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.demo.proto.usermgmt.v2.Student.GenerateImportStudentTemplateResponse.getDefaultInstance()))
              .setSchemaDescriptor(new StudentServiceMethodDescriptorSupplier("GenerateImportStudentTemplate"))
              .build();
        }
      }
    }
    return getGenerateImportStudentTemplateMethod;
  }

  private static volatile io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentRequest,
      io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentResponse> getImportStudentMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ImportStudent",
      requestType = io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentRequest.class,
      responseType = io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentRequest,
      io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentResponse> getImportStudentMethod() {
    io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentRequest, io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentResponse> getImportStudentMethod;
    if ((getImportStudentMethod = StudentServiceGrpc.getImportStudentMethod) == null) {
      synchronized (StudentServiceGrpc.class) {
        if ((getImportStudentMethod = StudentServiceGrpc.getImportStudentMethod) == null) {
          StudentServiceGrpc.getImportStudentMethod = getImportStudentMethod =
              io.grpc.MethodDescriptor.<io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentRequest, io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ImportStudent"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentResponse.getDefaultInstance()))
              .setSchemaDescriptor(new StudentServiceMethodDescriptorSupplier("ImportStudent"))
              .build();
        }
      }
    }
    return getImportStudentMethod;
  }

  private static volatile io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentRequest,
      io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentResponse> getImportStudentV2Method;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ImportStudentV2",
      requestType = io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentRequest.class,
      responseType = io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentRequest,
      io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentResponse> getImportStudentV2Method() {
    io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentRequest, io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentResponse> getImportStudentV2Method;
    if ((getImportStudentV2Method = StudentServiceGrpc.getImportStudentV2Method) == null) {
      synchronized (StudentServiceGrpc.class) {
        if ((getImportStudentV2Method = StudentServiceGrpc.getImportStudentV2Method) == null) {
          StudentServiceGrpc.getImportStudentV2Method = getImportStudentV2Method =
              io.grpc.MethodDescriptor.<io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentRequest, io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ImportStudentV2"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentResponse.getDefaultInstance()))
              .setSchemaDescriptor(new StudentServiceMethodDescriptorSupplier("ImportStudentV2"))
              .build();
        }
      }
    }
    return getImportStudentV2Method;
  }

  private static volatile io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentRequest,
      io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentResponse> getUpsertStudentMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UpsertStudent",
      requestType = io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentRequest.class,
      responseType = io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentRequest,
      io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentResponse> getUpsertStudentMethod() {
    io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentRequest, io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentResponse> getUpsertStudentMethod;
    if ((getUpsertStudentMethod = StudentServiceGrpc.getUpsertStudentMethod) == null) {
      synchronized (StudentServiceGrpc.class) {
        if ((getUpsertStudentMethod = StudentServiceGrpc.getUpsertStudentMethod) == null) {
          StudentServiceGrpc.getUpsertStudentMethod = getUpsertStudentMethod =
              io.grpc.MethodDescriptor.<io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentRequest, io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UpsertStudent"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentResponse.getDefaultInstance()))
              .setSchemaDescriptor(new StudentServiceMethodDescriptorSupplier("UpsertStudent"))
              .build();
        }
      }
    }
    return getUpsertStudentMethod;
  }

  private static volatile io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Student.RetrieveStudentCommentRequest,
      io.manabie.demo.proto.usermgmt.v2.Student.RetrieveStudentCommentResponse> getRetrieveStudentCommentMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "RetrieveStudentComment",
      requestType = io.manabie.demo.proto.usermgmt.v2.Student.RetrieveStudentCommentRequest.class,
      responseType = io.manabie.demo.proto.usermgmt.v2.Student.RetrieveStudentCommentResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Student.RetrieveStudentCommentRequest,
      io.manabie.demo.proto.usermgmt.v2.Student.RetrieveStudentCommentResponse> getRetrieveStudentCommentMethod() {
    io.grpc.MethodDescriptor<io.manabie.demo.proto.usermgmt.v2.Student.RetrieveStudentCommentRequest, io.manabie.demo.proto.usermgmt.v2.Student.RetrieveStudentCommentResponse> getRetrieveStudentCommentMethod;
    if ((getRetrieveStudentCommentMethod = StudentServiceGrpc.getRetrieveStudentCommentMethod) == null) {
      synchronized (StudentServiceGrpc.class) {
        if ((getRetrieveStudentCommentMethod = StudentServiceGrpc.getRetrieveStudentCommentMethod) == null) {
          StudentServiceGrpc.getRetrieveStudentCommentMethod = getRetrieveStudentCommentMethod =
              io.grpc.MethodDescriptor.<io.manabie.demo.proto.usermgmt.v2.Student.RetrieveStudentCommentRequest, io.manabie.demo.proto.usermgmt.v2.Student.RetrieveStudentCommentResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "RetrieveStudentComment"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.demo.proto.usermgmt.v2.Student.RetrieveStudentCommentRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.demo.proto.usermgmt.v2.Student.RetrieveStudentCommentResponse.getDefaultInstance()))
              .setSchemaDescriptor(new StudentServiceMethodDescriptorSupplier("RetrieveStudentComment"))
              .build();
        }
      }
    }
    return getRetrieveStudentCommentMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static StudentServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<StudentServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<StudentServiceStub>() {
        @java.lang.Override
        public StudentServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new StudentServiceStub(channel, callOptions);
        }
      };
    return StudentServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static StudentServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<StudentServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<StudentServiceBlockingStub>() {
        @java.lang.Override
        public StudentServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new StudentServiceBlockingStub(channel, callOptions);
        }
      };
    return StudentServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static StudentServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<StudentServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<StudentServiceFutureStub>() {
        @java.lang.Override
        public StudentServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new StudentServiceFutureStub(channel, callOptions);
        }
      };
    return StudentServiceFutureStub.newStub(factory, channel);
  }

  /**
   */
  public interface AsyncService {

    /**
     */
    default void getStudentProfile(io.manabie.demo.proto.usermgmt.v2.Student.GetStudentProfileRequest request,
        io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Student.GetStudentProfileResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetStudentProfileMethod(), responseObserver);
    }

    /**
     */
    default void upsertStudentComment(io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentCommentRequest request,
        io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentCommentResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUpsertStudentCommentMethod(), responseObserver);
    }

    /**
     */
    default void deleteStudentComments(io.manabie.demo.proto.usermgmt.v2.Student.DeleteStudentCommentsRequest request,
        io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Student.DeleteStudentCommentsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getDeleteStudentCommentsMethod(), responseObserver);
    }

    /**
     */
    default void generateImportStudentTemplate(io.manabie.demo.proto.usermgmt.v2.Student.GenerateImportStudentTemplateRequest request,
        io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Student.GenerateImportStudentTemplateResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGenerateImportStudentTemplateMethod(), responseObserver);
    }

    /**
     */
    @java.lang.Deprecated
    default void importStudent(io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentRequest request,
        io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getImportStudentMethod(), responseObserver);
    }

    /**
     */
    default void importStudentV2(io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentRequest request,
        io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getImportStudentV2Method(), responseObserver);
    }

    /**
     */
    default void upsertStudent(io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentRequest request,
        io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUpsertStudentMethod(), responseObserver);
    }

    /**
     */
    default void retrieveStudentComment(io.manabie.demo.proto.usermgmt.v2.Student.RetrieveStudentCommentRequest request,
        io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Student.RetrieveStudentCommentResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getRetrieveStudentCommentMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service StudentService.
   */
  public static abstract class StudentServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return StudentServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service StudentService.
   */
  public static final class StudentServiceStub
      extends io.grpc.stub.AbstractAsyncStub<StudentServiceStub> {
    private StudentServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected StudentServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new StudentServiceStub(channel, callOptions);
    }

    /**
     */
    public void getStudentProfile(io.manabie.demo.proto.usermgmt.v2.Student.GetStudentProfileRequest request,
        io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Student.GetStudentProfileResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetStudentProfileMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void upsertStudentComment(io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentCommentRequest request,
        io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentCommentResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUpsertStudentCommentMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void deleteStudentComments(io.manabie.demo.proto.usermgmt.v2.Student.DeleteStudentCommentsRequest request,
        io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Student.DeleteStudentCommentsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getDeleteStudentCommentsMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void generateImportStudentTemplate(io.manabie.demo.proto.usermgmt.v2.Student.GenerateImportStudentTemplateRequest request,
        io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Student.GenerateImportStudentTemplateResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGenerateImportStudentTemplateMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    @java.lang.Deprecated
    public void importStudent(io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentRequest request,
        io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getImportStudentMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void importStudentV2(io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentRequest request,
        io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getImportStudentV2Method(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void upsertStudent(io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentRequest request,
        io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUpsertStudentMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void retrieveStudentComment(io.manabie.demo.proto.usermgmt.v2.Student.RetrieveStudentCommentRequest request,
        io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Student.RetrieveStudentCommentResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getRetrieveStudentCommentMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service StudentService.
   */
  public static final class StudentServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<StudentServiceBlockingStub> {
    private StudentServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected StudentServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new StudentServiceBlockingStub(channel, callOptions);
    }

    /**
     */
    public io.manabie.demo.proto.usermgmt.v2.Student.GetStudentProfileResponse getStudentProfile(io.manabie.demo.proto.usermgmt.v2.Student.GetStudentProfileRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetStudentProfileMethod(), getCallOptions(), request);
    }

    /**
     */
    public io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentCommentResponse upsertStudentComment(io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentCommentRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUpsertStudentCommentMethod(), getCallOptions(), request);
    }

    /**
     */
    public io.manabie.demo.proto.usermgmt.v2.Student.DeleteStudentCommentsResponse deleteStudentComments(io.manabie.demo.proto.usermgmt.v2.Student.DeleteStudentCommentsRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getDeleteStudentCommentsMethod(), getCallOptions(), request);
    }

    /**
     */
    public io.manabie.demo.proto.usermgmt.v2.Student.GenerateImportStudentTemplateResponse generateImportStudentTemplate(io.manabie.demo.proto.usermgmt.v2.Student.GenerateImportStudentTemplateRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGenerateImportStudentTemplateMethod(), getCallOptions(), request);
    }

    /**
     */
    @java.lang.Deprecated
    public io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentResponse importStudent(io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getImportStudentMethod(), getCallOptions(), request);
    }

    /**
     */
    public io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentResponse importStudentV2(io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getImportStudentV2Method(), getCallOptions(), request);
    }

    /**
     */
    public io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentResponse upsertStudent(io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUpsertStudentMethod(), getCallOptions(), request);
    }

    /**
     */
    public io.manabie.demo.proto.usermgmt.v2.Student.RetrieveStudentCommentResponse retrieveStudentComment(io.manabie.demo.proto.usermgmt.v2.Student.RetrieveStudentCommentRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getRetrieveStudentCommentMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service StudentService.
   */
  public static final class StudentServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<StudentServiceFutureStub> {
    private StudentServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected StudentServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new StudentServiceFutureStub(channel, callOptions);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.demo.proto.usermgmt.v2.Student.GetStudentProfileResponse> getStudentProfile(
        io.manabie.demo.proto.usermgmt.v2.Student.GetStudentProfileRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetStudentProfileMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentCommentResponse> upsertStudentComment(
        io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentCommentRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUpsertStudentCommentMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.demo.proto.usermgmt.v2.Student.DeleteStudentCommentsResponse> deleteStudentComments(
        io.manabie.demo.proto.usermgmt.v2.Student.DeleteStudentCommentsRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getDeleteStudentCommentsMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.demo.proto.usermgmt.v2.Student.GenerateImportStudentTemplateResponse> generateImportStudentTemplate(
        io.manabie.demo.proto.usermgmt.v2.Student.GenerateImportStudentTemplateRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGenerateImportStudentTemplateMethod(), getCallOptions()), request);
    }

    /**
     */
    @java.lang.Deprecated
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentResponse> importStudent(
        io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getImportStudentMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentResponse> importStudentV2(
        io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getImportStudentV2Method(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentResponse> upsertStudent(
        io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUpsertStudentMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.demo.proto.usermgmt.v2.Student.RetrieveStudentCommentResponse> retrieveStudentComment(
        io.manabie.demo.proto.usermgmt.v2.Student.RetrieveStudentCommentRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getRetrieveStudentCommentMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_GET_STUDENT_PROFILE = 0;
  private static final int METHODID_UPSERT_STUDENT_COMMENT = 1;
  private static final int METHODID_DELETE_STUDENT_COMMENTS = 2;
  private static final int METHODID_GENERATE_IMPORT_STUDENT_TEMPLATE = 3;
  private static final int METHODID_IMPORT_STUDENT = 4;
  private static final int METHODID_IMPORT_STUDENT_V2 = 5;
  private static final int METHODID_UPSERT_STUDENT = 6;
  private static final int METHODID_RETRIEVE_STUDENT_COMMENT = 7;

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
        case METHODID_GET_STUDENT_PROFILE:
          serviceImpl.getStudentProfile((io.manabie.demo.proto.usermgmt.v2.Student.GetStudentProfileRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Student.GetStudentProfileResponse>) responseObserver);
          break;
        case METHODID_UPSERT_STUDENT_COMMENT:
          serviceImpl.upsertStudentComment((io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentCommentRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentCommentResponse>) responseObserver);
          break;
        case METHODID_DELETE_STUDENT_COMMENTS:
          serviceImpl.deleteStudentComments((io.manabie.demo.proto.usermgmt.v2.Student.DeleteStudentCommentsRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Student.DeleteStudentCommentsResponse>) responseObserver);
          break;
        case METHODID_GENERATE_IMPORT_STUDENT_TEMPLATE:
          serviceImpl.generateImportStudentTemplate((io.manabie.demo.proto.usermgmt.v2.Student.GenerateImportStudentTemplateRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Student.GenerateImportStudentTemplateResponse>) responseObserver);
          break;
        case METHODID_IMPORT_STUDENT:
          serviceImpl.importStudent((io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentResponse>) responseObserver);
          break;
        case METHODID_IMPORT_STUDENT_V2:
          serviceImpl.importStudentV2((io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentResponse>) responseObserver);
          break;
        case METHODID_UPSERT_STUDENT:
          serviceImpl.upsertStudent((io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentResponse>) responseObserver);
          break;
        case METHODID_RETRIEVE_STUDENT_COMMENT:
          serviceImpl.retrieveStudentComment((io.manabie.demo.proto.usermgmt.v2.Student.RetrieveStudentCommentRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.demo.proto.usermgmt.v2.Student.RetrieveStudentCommentResponse>) responseObserver);
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
          getGetStudentProfileMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.demo.proto.usermgmt.v2.Student.GetStudentProfileRequest,
              io.manabie.demo.proto.usermgmt.v2.Student.GetStudentProfileResponse>(
                service, METHODID_GET_STUDENT_PROFILE)))
        .addMethod(
          getUpsertStudentCommentMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentCommentRequest,
              io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentCommentResponse>(
                service, METHODID_UPSERT_STUDENT_COMMENT)))
        .addMethod(
          getDeleteStudentCommentsMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.demo.proto.usermgmt.v2.Student.DeleteStudentCommentsRequest,
              io.manabie.demo.proto.usermgmt.v2.Student.DeleteStudentCommentsResponse>(
                service, METHODID_DELETE_STUDENT_COMMENTS)))
        .addMethod(
          getGenerateImportStudentTemplateMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.demo.proto.usermgmt.v2.Student.GenerateImportStudentTemplateRequest,
              io.manabie.demo.proto.usermgmt.v2.Student.GenerateImportStudentTemplateResponse>(
                service, METHODID_GENERATE_IMPORT_STUDENT_TEMPLATE)))
        .addMethod(
          getImportStudentMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentRequest,
              io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentResponse>(
                service, METHODID_IMPORT_STUDENT)))
        .addMethod(
          getImportStudentV2Method(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.demo.proto.usermgmt.v2.Student.ImportStudentRequest,
              io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentResponse>(
                service, METHODID_IMPORT_STUDENT_V2)))
        .addMethod(
          getUpsertStudentMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentRequest,
              io.manabie.demo.proto.usermgmt.v2.Student.UpsertStudentResponse>(
                service, METHODID_UPSERT_STUDENT)))
        .addMethod(
          getRetrieveStudentCommentMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.demo.proto.usermgmt.v2.Student.RetrieveStudentCommentRequest,
              io.manabie.demo.proto.usermgmt.v2.Student.RetrieveStudentCommentResponse>(
                service, METHODID_RETRIEVE_STUDENT_COMMENT)))
        .build();
  }

  private static abstract class StudentServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    StudentServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return io.manabie.demo.proto.usermgmt.v2.Student.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("StudentService");
    }
  }

  private static final class StudentServiceFileDescriptorSupplier
      extends StudentServiceBaseDescriptorSupplier {
    StudentServiceFileDescriptorSupplier() {}
  }

  private static final class StudentServiceMethodDescriptorSupplier
      extends StudentServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final String methodName;

    StudentServiceMethodDescriptorSupplier(String methodName) {
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
      synchronized (StudentServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new StudentServiceFileDescriptorSupplier())
              .addMethod(getGetStudentProfileMethod())
              .addMethod(getUpsertStudentCommentMethod())
              .addMethod(getDeleteStudentCommentsMethod())
              .addMethod(getGenerateImportStudentTemplateMethod())
              .addMethod(getImportStudentMethod())
              .addMethod(getImportStudentV2Method())
              .addMethod(getUpsertStudentMethod())
              .addMethod(getRetrieveStudentCommentMethod())
              .build();
        }
      }
    }
    return result;
  }
}
