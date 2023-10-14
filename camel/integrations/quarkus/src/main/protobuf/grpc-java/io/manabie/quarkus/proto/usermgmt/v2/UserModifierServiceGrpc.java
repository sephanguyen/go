package io.manabie.quarkus.proto.usermgmt.v2;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 * <pre>
 * services
 * </pre>
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.56.1)",
    comments = "Source: usermgmt/v2/users.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class UserModifierServiceGrpc {

  private UserModifierServiceGrpc() {}

  public static final String SERVICE_NAME = "usermgmt.v2.UserModifierService";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.CreateStudentRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Users.CreateStudentResponse> getCreateStudentMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "CreateStudent",
      requestType = io.manabie.quarkus.proto.usermgmt.v2.Users.CreateStudentRequest.class,
      responseType = io.manabie.quarkus.proto.usermgmt.v2.Users.CreateStudentResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.CreateStudentRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Users.CreateStudentResponse> getCreateStudentMethod() {
    io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.CreateStudentRequest, io.manabie.quarkus.proto.usermgmt.v2.Users.CreateStudentResponse> getCreateStudentMethod;
    if ((getCreateStudentMethod = UserModifierServiceGrpc.getCreateStudentMethod) == null) {
      synchronized (UserModifierServiceGrpc.class) {
        if ((getCreateStudentMethod = UserModifierServiceGrpc.getCreateStudentMethod) == null) {
          UserModifierServiceGrpc.getCreateStudentMethod = getCreateStudentMethod =
              io.grpc.MethodDescriptor.<io.manabie.quarkus.proto.usermgmt.v2.Users.CreateStudentRequest, io.manabie.quarkus.proto.usermgmt.v2.Users.CreateStudentResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "CreateStudent"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Users.CreateStudentRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Users.CreateStudentResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserModifierServiceMethodDescriptorSupplier("CreateStudent"))
              .build();
        }
      }
    }
    return getCreateStudentMethod;
  }

  private static volatile io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.CreateParentsAndAssignToStudentRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Users.CreateParentsAndAssignToStudentResponse> getCreateParentsAndAssignToStudentMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "CreateParentsAndAssignToStudent",
      requestType = io.manabie.quarkus.proto.usermgmt.v2.Users.CreateParentsAndAssignToStudentRequest.class,
      responseType = io.manabie.quarkus.proto.usermgmt.v2.Users.CreateParentsAndAssignToStudentResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.CreateParentsAndAssignToStudentRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Users.CreateParentsAndAssignToStudentResponse> getCreateParentsAndAssignToStudentMethod() {
    io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.CreateParentsAndAssignToStudentRequest, io.manabie.quarkus.proto.usermgmt.v2.Users.CreateParentsAndAssignToStudentResponse> getCreateParentsAndAssignToStudentMethod;
    if ((getCreateParentsAndAssignToStudentMethod = UserModifierServiceGrpc.getCreateParentsAndAssignToStudentMethod) == null) {
      synchronized (UserModifierServiceGrpc.class) {
        if ((getCreateParentsAndAssignToStudentMethod = UserModifierServiceGrpc.getCreateParentsAndAssignToStudentMethod) == null) {
          UserModifierServiceGrpc.getCreateParentsAndAssignToStudentMethod = getCreateParentsAndAssignToStudentMethod =
              io.grpc.MethodDescriptor.<io.manabie.quarkus.proto.usermgmt.v2.Users.CreateParentsAndAssignToStudentRequest, io.manabie.quarkus.proto.usermgmt.v2.Users.CreateParentsAndAssignToStudentResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "CreateParentsAndAssignToStudent"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Users.CreateParentsAndAssignToStudentRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Users.CreateParentsAndAssignToStudentResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserModifierServiceMethodDescriptorSupplier("CreateParentsAndAssignToStudent"))
              .build();
        }
      }
    }
    return getCreateParentsAndAssignToStudentMethod;
  }

  private static volatile io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.ImportParentsAndAssignToStudentRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Users.ImportParentsAndAssignToStudentResponse> getImportParentsAndAssignToStudentMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ImportParentsAndAssignToStudent",
      requestType = io.manabie.quarkus.proto.usermgmt.v2.Users.ImportParentsAndAssignToStudentRequest.class,
      responseType = io.manabie.quarkus.proto.usermgmt.v2.Users.ImportParentsAndAssignToStudentResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.ImportParentsAndAssignToStudentRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Users.ImportParentsAndAssignToStudentResponse> getImportParentsAndAssignToStudentMethod() {
    io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.ImportParentsAndAssignToStudentRequest, io.manabie.quarkus.proto.usermgmt.v2.Users.ImportParentsAndAssignToStudentResponse> getImportParentsAndAssignToStudentMethod;
    if ((getImportParentsAndAssignToStudentMethod = UserModifierServiceGrpc.getImportParentsAndAssignToStudentMethod) == null) {
      synchronized (UserModifierServiceGrpc.class) {
        if ((getImportParentsAndAssignToStudentMethod = UserModifierServiceGrpc.getImportParentsAndAssignToStudentMethod) == null) {
          UserModifierServiceGrpc.getImportParentsAndAssignToStudentMethod = getImportParentsAndAssignToStudentMethod =
              io.grpc.MethodDescriptor.<io.manabie.quarkus.proto.usermgmt.v2.Users.ImportParentsAndAssignToStudentRequest, io.manabie.quarkus.proto.usermgmt.v2.Users.ImportParentsAndAssignToStudentResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ImportParentsAndAssignToStudent"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Users.ImportParentsAndAssignToStudentRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Users.ImportParentsAndAssignToStudentResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserModifierServiceMethodDescriptorSupplier("ImportParentsAndAssignToStudent"))
              .build();
        }
      }
    }
    return getImportParentsAndAssignToStudentMethod;
  }

  private static volatile io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateStudentRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateStudentResponse> getUpdateStudentMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UpdateStudent",
      requestType = io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateStudentRequest.class,
      responseType = io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateStudentResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateStudentRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateStudentResponse> getUpdateStudentMethod() {
    io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateStudentRequest, io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateStudentResponse> getUpdateStudentMethod;
    if ((getUpdateStudentMethod = UserModifierServiceGrpc.getUpdateStudentMethod) == null) {
      synchronized (UserModifierServiceGrpc.class) {
        if ((getUpdateStudentMethod = UserModifierServiceGrpc.getUpdateStudentMethod) == null) {
          UserModifierServiceGrpc.getUpdateStudentMethod = getUpdateStudentMethod =
              io.grpc.MethodDescriptor.<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateStudentRequest, io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateStudentResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UpdateStudent"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateStudentRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateStudentResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserModifierServiceMethodDescriptorSupplier("UpdateStudent"))
              .build();
        }
      }
    }
    return getUpdateStudentMethod;
  }

  private static volatile io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateParentsAndFamilyRelationshipRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateParentsAndFamilyRelationshipResponse> getUpdateParentsAndFamilyRelationshipMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UpdateParentsAndFamilyRelationship",
      requestType = io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateParentsAndFamilyRelationshipRequest.class,
      responseType = io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateParentsAndFamilyRelationshipResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateParentsAndFamilyRelationshipRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateParentsAndFamilyRelationshipResponse> getUpdateParentsAndFamilyRelationshipMethod() {
    io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateParentsAndFamilyRelationshipRequest, io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateParentsAndFamilyRelationshipResponse> getUpdateParentsAndFamilyRelationshipMethod;
    if ((getUpdateParentsAndFamilyRelationshipMethod = UserModifierServiceGrpc.getUpdateParentsAndFamilyRelationshipMethod) == null) {
      synchronized (UserModifierServiceGrpc.class) {
        if ((getUpdateParentsAndFamilyRelationshipMethod = UserModifierServiceGrpc.getUpdateParentsAndFamilyRelationshipMethod) == null) {
          UserModifierServiceGrpc.getUpdateParentsAndFamilyRelationshipMethod = getUpdateParentsAndFamilyRelationshipMethod =
              io.grpc.MethodDescriptor.<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateParentsAndFamilyRelationshipRequest, io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateParentsAndFamilyRelationshipResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UpdateParentsAndFamilyRelationship"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateParentsAndFamilyRelationshipRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateParentsAndFamilyRelationshipResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserModifierServiceMethodDescriptorSupplier("UpdateParentsAndFamilyRelationship"))
              .build();
        }
      }
    }
    return getUpdateParentsAndFamilyRelationshipMethod;
  }

  private static volatile io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.ReissueUserPasswordRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Users.ReissueUserPasswordResponse> getReissueUserPasswordMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ReissueUserPassword",
      requestType = io.manabie.quarkus.proto.usermgmt.v2.Users.ReissueUserPasswordRequest.class,
      responseType = io.manabie.quarkus.proto.usermgmt.v2.Users.ReissueUserPasswordResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.ReissueUserPasswordRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Users.ReissueUserPasswordResponse> getReissueUserPasswordMethod() {
    io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.ReissueUserPasswordRequest, io.manabie.quarkus.proto.usermgmt.v2.Users.ReissueUserPasswordResponse> getReissueUserPasswordMethod;
    if ((getReissueUserPasswordMethod = UserModifierServiceGrpc.getReissueUserPasswordMethod) == null) {
      synchronized (UserModifierServiceGrpc.class) {
        if ((getReissueUserPasswordMethod = UserModifierServiceGrpc.getReissueUserPasswordMethod) == null) {
          UserModifierServiceGrpc.getReissueUserPasswordMethod = getReissueUserPasswordMethod =
              io.grpc.MethodDescriptor.<io.manabie.quarkus.proto.usermgmt.v2.Users.ReissueUserPasswordRequest, io.manabie.quarkus.proto.usermgmt.v2.Users.ReissueUserPasswordResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ReissueUserPassword"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Users.ReissueUserPasswordRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Users.ReissueUserPasswordResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserModifierServiceMethodDescriptorSupplier("ReissueUserPassword"))
              .build();
        }
      }
    }
    return getReissueUserPasswordMethod;
  }

  private static volatile io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.UpsertStudentCoursePackageRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Users.UpsertStudentCoursePackageResponse> getUpsertStudentCoursePackageMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UpsertStudentCoursePackage",
      requestType = io.manabie.quarkus.proto.usermgmt.v2.Users.UpsertStudentCoursePackageRequest.class,
      responseType = io.manabie.quarkus.proto.usermgmt.v2.Users.UpsertStudentCoursePackageResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.UpsertStudentCoursePackageRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Users.UpsertStudentCoursePackageResponse> getUpsertStudentCoursePackageMethod() {
    io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.UpsertStudentCoursePackageRequest, io.manabie.quarkus.proto.usermgmt.v2.Users.UpsertStudentCoursePackageResponse> getUpsertStudentCoursePackageMethod;
    if ((getUpsertStudentCoursePackageMethod = UserModifierServiceGrpc.getUpsertStudentCoursePackageMethod) == null) {
      synchronized (UserModifierServiceGrpc.class) {
        if ((getUpsertStudentCoursePackageMethod = UserModifierServiceGrpc.getUpsertStudentCoursePackageMethod) == null) {
          UserModifierServiceGrpc.getUpsertStudentCoursePackageMethod = getUpsertStudentCoursePackageMethod =
              io.grpc.MethodDescriptor.<io.manabie.quarkus.proto.usermgmt.v2.Users.UpsertStudentCoursePackageRequest, io.manabie.quarkus.proto.usermgmt.v2.Users.UpsertStudentCoursePackageResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UpsertStudentCoursePackage"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Users.UpsertStudentCoursePackageRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Users.UpsertStudentCoursePackageResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserModifierServiceMethodDescriptorSupplier("UpsertStudentCoursePackage"))
              .build();
        }
      }
    }
    return getUpsertStudentCoursePackageMethod;
  }

  private static volatile io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.RemoveParentFromStudentRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Users.RemoveParentFromStudentResponse> getRemoveParentFromStudentMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "RemoveParentFromStudent",
      requestType = io.manabie.quarkus.proto.usermgmt.v2.Users.RemoveParentFromStudentRequest.class,
      responseType = io.manabie.quarkus.proto.usermgmt.v2.Users.RemoveParentFromStudentResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.RemoveParentFromStudentRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Users.RemoveParentFromStudentResponse> getRemoveParentFromStudentMethod() {
    io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.RemoveParentFromStudentRequest, io.manabie.quarkus.proto.usermgmt.v2.Users.RemoveParentFromStudentResponse> getRemoveParentFromStudentMethod;
    if ((getRemoveParentFromStudentMethod = UserModifierServiceGrpc.getRemoveParentFromStudentMethod) == null) {
      synchronized (UserModifierServiceGrpc.class) {
        if ((getRemoveParentFromStudentMethod = UserModifierServiceGrpc.getRemoveParentFromStudentMethod) == null) {
          UserModifierServiceGrpc.getRemoveParentFromStudentMethod = getRemoveParentFromStudentMethod =
              io.grpc.MethodDescriptor.<io.manabie.quarkus.proto.usermgmt.v2.Users.RemoveParentFromStudentRequest, io.manabie.quarkus.proto.usermgmt.v2.Users.RemoveParentFromStudentResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "RemoveParentFromStudent"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Users.RemoveParentFromStudentRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Users.RemoveParentFromStudentResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserModifierServiceMethodDescriptorSupplier("RemoveParentFromStudent"))
              .build();
        }
      }
    }
    return getRemoveParentFromStudentMethod;
  }

  private static volatile io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserProfileRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserProfileResponse> getUpdateUserProfileMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UpdateUserProfile",
      requestType = io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserProfileRequest.class,
      responseType = io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserProfileResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserProfileRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserProfileResponse> getUpdateUserProfileMethod() {
    io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserProfileRequest, io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserProfileResponse> getUpdateUserProfileMethod;
    if ((getUpdateUserProfileMethod = UserModifierServiceGrpc.getUpdateUserProfileMethod) == null) {
      synchronized (UserModifierServiceGrpc.class) {
        if ((getUpdateUserProfileMethod = UserModifierServiceGrpc.getUpdateUserProfileMethod) == null) {
          UserModifierServiceGrpc.getUpdateUserProfileMethod = getUpdateUserProfileMethod =
              io.grpc.MethodDescriptor.<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserProfileRequest, io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserProfileResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UpdateUserProfile"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserProfileRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserProfileResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserModifierServiceMethodDescriptorSupplier("UpdateUserProfile"))
              .build();
        }
      }
    }
    return getUpdateUserProfileMethod;
  }

  private static volatile io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserDeviceTokenRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserDeviceTokenResponse> getUpdateUserDeviceTokenMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UpdateUserDeviceToken",
      requestType = io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserDeviceTokenRequest.class,
      responseType = io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserDeviceTokenResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserDeviceTokenRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserDeviceTokenResponse> getUpdateUserDeviceTokenMethod() {
    io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserDeviceTokenRequest, io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserDeviceTokenResponse> getUpdateUserDeviceTokenMethod;
    if ((getUpdateUserDeviceTokenMethod = UserModifierServiceGrpc.getUpdateUserDeviceTokenMethod) == null) {
      synchronized (UserModifierServiceGrpc.class) {
        if ((getUpdateUserDeviceTokenMethod = UserModifierServiceGrpc.getUpdateUserDeviceTokenMethod) == null) {
          UserModifierServiceGrpc.getUpdateUserDeviceTokenMethod = getUpdateUserDeviceTokenMethod =
              io.grpc.MethodDescriptor.<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserDeviceTokenRequest, io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserDeviceTokenResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UpdateUserDeviceToken"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserDeviceTokenRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserDeviceTokenResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserModifierServiceMethodDescriptorSupplier("UpdateUserDeviceToken"))
              .build();
        }
      }
    }
    return getUpdateUserDeviceTokenMethod;
  }

  private static volatile io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserLastLoginDateRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserLastLoginDateResponse> getUpdateUserLastLoginDateMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UpdateUserLastLoginDate",
      requestType = io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserLastLoginDateRequest.class,
      responseType = io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserLastLoginDateResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserLastLoginDateRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserLastLoginDateResponse> getUpdateUserLastLoginDateMethod() {
    io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserLastLoginDateRequest, io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserLastLoginDateResponse> getUpdateUserLastLoginDateMethod;
    if ((getUpdateUserLastLoginDateMethod = UserModifierServiceGrpc.getUpdateUserLastLoginDateMethod) == null) {
      synchronized (UserModifierServiceGrpc.class) {
        if ((getUpdateUserLastLoginDateMethod = UserModifierServiceGrpc.getUpdateUserLastLoginDateMethod) == null) {
          UserModifierServiceGrpc.getUpdateUserLastLoginDateMethod = getUpdateUserLastLoginDateMethod =
              io.grpc.MethodDescriptor.<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserLastLoginDateRequest, io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserLastLoginDateResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UpdateUserLastLoginDate"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserLastLoginDateRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserLastLoginDateResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserModifierServiceMethodDescriptorSupplier("UpdateUserLastLoginDate"))
              .build();
        }
      }
    }
    return getUpdateUserLastLoginDateMethod;
  }

  private static volatile io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.GenerateImportParentsAndAssignToStudentTemplateRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Users.GenerateImportParentsAndAssignToStudentTemplateResponse> getGenerateImportParentsAndAssignToStudentTemplateMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GenerateImportParentsAndAssignToStudentTemplate",
      requestType = io.manabie.quarkus.proto.usermgmt.v2.Users.GenerateImportParentsAndAssignToStudentTemplateRequest.class,
      responseType = io.manabie.quarkus.proto.usermgmt.v2.Users.GenerateImportParentsAndAssignToStudentTemplateResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.GenerateImportParentsAndAssignToStudentTemplateRequest,
      io.manabie.quarkus.proto.usermgmt.v2.Users.GenerateImportParentsAndAssignToStudentTemplateResponse> getGenerateImportParentsAndAssignToStudentTemplateMethod() {
    io.grpc.MethodDescriptor<io.manabie.quarkus.proto.usermgmt.v2.Users.GenerateImportParentsAndAssignToStudentTemplateRequest, io.manabie.quarkus.proto.usermgmt.v2.Users.GenerateImportParentsAndAssignToStudentTemplateResponse> getGenerateImportParentsAndAssignToStudentTemplateMethod;
    if ((getGenerateImportParentsAndAssignToStudentTemplateMethod = UserModifierServiceGrpc.getGenerateImportParentsAndAssignToStudentTemplateMethod) == null) {
      synchronized (UserModifierServiceGrpc.class) {
        if ((getGenerateImportParentsAndAssignToStudentTemplateMethod = UserModifierServiceGrpc.getGenerateImportParentsAndAssignToStudentTemplateMethod) == null) {
          UserModifierServiceGrpc.getGenerateImportParentsAndAssignToStudentTemplateMethod = getGenerateImportParentsAndAssignToStudentTemplateMethod =
              io.grpc.MethodDescriptor.<io.manabie.quarkus.proto.usermgmt.v2.Users.GenerateImportParentsAndAssignToStudentTemplateRequest, io.manabie.quarkus.proto.usermgmt.v2.Users.GenerateImportParentsAndAssignToStudentTemplateResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GenerateImportParentsAndAssignToStudentTemplate"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Users.GenerateImportParentsAndAssignToStudentTemplateRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  io.manabie.quarkus.proto.usermgmt.v2.Users.GenerateImportParentsAndAssignToStudentTemplateResponse.getDefaultInstance()))
              .setSchemaDescriptor(new UserModifierServiceMethodDescriptorSupplier("GenerateImportParentsAndAssignToStudentTemplate"))
              .build();
        }
      }
    }
    return getGenerateImportParentsAndAssignToStudentTemplateMethod;
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
   * <pre>
   * services
   * </pre>
   */
  public interface AsyncService {

    /**
     */
    @java.lang.Deprecated
    default void createStudent(io.manabie.quarkus.proto.usermgmt.v2.Users.CreateStudentRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.CreateStudentResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getCreateStudentMethod(), responseObserver);
    }

    /**
     */
    default void createParentsAndAssignToStudent(io.manabie.quarkus.proto.usermgmt.v2.Users.CreateParentsAndAssignToStudentRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.CreateParentsAndAssignToStudentResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getCreateParentsAndAssignToStudentMethod(), responseObserver);
    }

    /**
     */
    default void importParentsAndAssignToStudent(io.manabie.quarkus.proto.usermgmt.v2.Users.ImportParentsAndAssignToStudentRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.ImportParentsAndAssignToStudentResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getImportParentsAndAssignToStudentMethod(), responseObserver);
    }

    /**
     */
    @java.lang.Deprecated
    default void updateStudent(io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateStudentRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateStudentResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUpdateStudentMethod(), responseObserver);
    }

    /**
     */
    default void updateParentsAndFamilyRelationship(io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateParentsAndFamilyRelationshipRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateParentsAndFamilyRelationshipResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUpdateParentsAndFamilyRelationshipMethod(), responseObserver);
    }

    /**
     */
    default void reissueUserPassword(io.manabie.quarkus.proto.usermgmt.v2.Users.ReissueUserPasswordRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.ReissueUserPasswordResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getReissueUserPasswordMethod(), responseObserver);
    }

    /**
     */
    default void upsertStudentCoursePackage(io.manabie.quarkus.proto.usermgmt.v2.Users.UpsertStudentCoursePackageRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.UpsertStudentCoursePackageResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUpsertStudentCoursePackageMethod(), responseObserver);
    }

    /**
     */
    default void removeParentFromStudent(io.manabie.quarkus.proto.usermgmt.v2.Users.RemoveParentFromStudentRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.RemoveParentFromStudentResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getRemoveParentFromStudentMethod(), responseObserver);
    }

    /**
     */
    default void updateUserProfile(io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserProfileRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserProfileResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUpdateUserProfileMethod(), responseObserver);
    }

    /**
     */
    default void updateUserDeviceToken(io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserDeviceTokenRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserDeviceTokenResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUpdateUserDeviceTokenMethod(), responseObserver);
    }

    /**
     */
    default void updateUserLastLoginDate(io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserLastLoginDateRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserLastLoginDateResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUpdateUserLastLoginDateMethod(), responseObserver);
    }

    /**
     */
    default void generateImportParentsAndAssignToStudentTemplate(io.manabie.quarkus.proto.usermgmt.v2.Users.GenerateImportParentsAndAssignToStudentTemplateRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.GenerateImportParentsAndAssignToStudentTemplateResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGenerateImportParentsAndAssignToStudentTemplateMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service UserModifierService.
   * <pre>
   * services
   * </pre>
   */
  public static abstract class UserModifierServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return UserModifierServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service UserModifierService.
   * <pre>
   * services
   * </pre>
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
    @java.lang.Deprecated
    public void createStudent(io.manabie.quarkus.proto.usermgmt.v2.Users.CreateStudentRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.CreateStudentResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getCreateStudentMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void createParentsAndAssignToStudent(io.manabie.quarkus.proto.usermgmt.v2.Users.CreateParentsAndAssignToStudentRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.CreateParentsAndAssignToStudentResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getCreateParentsAndAssignToStudentMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void importParentsAndAssignToStudent(io.manabie.quarkus.proto.usermgmt.v2.Users.ImportParentsAndAssignToStudentRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.ImportParentsAndAssignToStudentResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getImportParentsAndAssignToStudentMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    @java.lang.Deprecated
    public void updateStudent(io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateStudentRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateStudentResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUpdateStudentMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void updateParentsAndFamilyRelationship(io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateParentsAndFamilyRelationshipRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateParentsAndFamilyRelationshipResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUpdateParentsAndFamilyRelationshipMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void reissueUserPassword(io.manabie.quarkus.proto.usermgmt.v2.Users.ReissueUserPasswordRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.ReissueUserPasswordResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getReissueUserPasswordMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void upsertStudentCoursePackage(io.manabie.quarkus.proto.usermgmt.v2.Users.UpsertStudentCoursePackageRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.UpsertStudentCoursePackageResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUpsertStudentCoursePackageMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void removeParentFromStudent(io.manabie.quarkus.proto.usermgmt.v2.Users.RemoveParentFromStudentRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.RemoveParentFromStudentResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getRemoveParentFromStudentMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void updateUserProfile(io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserProfileRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserProfileResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUpdateUserProfileMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void updateUserDeviceToken(io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserDeviceTokenRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserDeviceTokenResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUpdateUserDeviceTokenMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void updateUserLastLoginDate(io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserLastLoginDateRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserLastLoginDateResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUpdateUserLastLoginDateMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void generateImportParentsAndAssignToStudentTemplate(io.manabie.quarkus.proto.usermgmt.v2.Users.GenerateImportParentsAndAssignToStudentTemplateRequest request,
        io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.GenerateImportParentsAndAssignToStudentTemplateResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGenerateImportParentsAndAssignToStudentTemplateMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service UserModifierService.
   * <pre>
   * services
   * </pre>
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
    @java.lang.Deprecated
    public io.manabie.quarkus.proto.usermgmt.v2.Users.CreateStudentResponse createStudent(io.manabie.quarkus.proto.usermgmt.v2.Users.CreateStudentRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getCreateStudentMethod(), getCallOptions(), request);
    }

    /**
     */
    public io.manabie.quarkus.proto.usermgmt.v2.Users.CreateParentsAndAssignToStudentResponse createParentsAndAssignToStudent(io.manabie.quarkus.proto.usermgmt.v2.Users.CreateParentsAndAssignToStudentRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getCreateParentsAndAssignToStudentMethod(), getCallOptions(), request);
    }

    /**
     */
    public io.manabie.quarkus.proto.usermgmt.v2.Users.ImportParentsAndAssignToStudentResponse importParentsAndAssignToStudent(io.manabie.quarkus.proto.usermgmt.v2.Users.ImportParentsAndAssignToStudentRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getImportParentsAndAssignToStudentMethod(), getCallOptions(), request);
    }

    /**
     */
    @java.lang.Deprecated
    public io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateStudentResponse updateStudent(io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateStudentRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUpdateStudentMethod(), getCallOptions(), request);
    }

    /**
     */
    public io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateParentsAndFamilyRelationshipResponse updateParentsAndFamilyRelationship(io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateParentsAndFamilyRelationshipRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUpdateParentsAndFamilyRelationshipMethod(), getCallOptions(), request);
    }

    /**
     */
    public io.manabie.quarkus.proto.usermgmt.v2.Users.ReissueUserPasswordResponse reissueUserPassword(io.manabie.quarkus.proto.usermgmt.v2.Users.ReissueUserPasswordRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getReissueUserPasswordMethod(), getCallOptions(), request);
    }

    /**
     */
    public io.manabie.quarkus.proto.usermgmt.v2.Users.UpsertStudentCoursePackageResponse upsertStudentCoursePackage(io.manabie.quarkus.proto.usermgmt.v2.Users.UpsertStudentCoursePackageRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUpsertStudentCoursePackageMethod(), getCallOptions(), request);
    }

    /**
     */
    public io.manabie.quarkus.proto.usermgmt.v2.Users.RemoveParentFromStudentResponse removeParentFromStudent(io.manabie.quarkus.proto.usermgmt.v2.Users.RemoveParentFromStudentRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getRemoveParentFromStudentMethod(), getCallOptions(), request);
    }

    /**
     */
    public io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserProfileResponse updateUserProfile(io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserProfileRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUpdateUserProfileMethod(), getCallOptions(), request);
    }

    /**
     */
    public io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserDeviceTokenResponse updateUserDeviceToken(io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserDeviceTokenRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUpdateUserDeviceTokenMethod(), getCallOptions(), request);
    }

    /**
     */
    public io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserLastLoginDateResponse updateUserLastLoginDate(io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserLastLoginDateRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUpdateUserLastLoginDateMethod(), getCallOptions(), request);
    }

    /**
     */
    public io.manabie.quarkus.proto.usermgmt.v2.Users.GenerateImportParentsAndAssignToStudentTemplateResponse generateImportParentsAndAssignToStudentTemplate(io.manabie.quarkus.proto.usermgmt.v2.Users.GenerateImportParentsAndAssignToStudentTemplateRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGenerateImportParentsAndAssignToStudentTemplateMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service UserModifierService.
   * <pre>
   * services
   * </pre>
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
    @java.lang.Deprecated
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.quarkus.proto.usermgmt.v2.Users.CreateStudentResponse> createStudent(
        io.manabie.quarkus.proto.usermgmt.v2.Users.CreateStudentRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getCreateStudentMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.quarkus.proto.usermgmt.v2.Users.CreateParentsAndAssignToStudentResponse> createParentsAndAssignToStudent(
        io.manabie.quarkus.proto.usermgmt.v2.Users.CreateParentsAndAssignToStudentRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getCreateParentsAndAssignToStudentMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.quarkus.proto.usermgmt.v2.Users.ImportParentsAndAssignToStudentResponse> importParentsAndAssignToStudent(
        io.manabie.quarkus.proto.usermgmt.v2.Users.ImportParentsAndAssignToStudentRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getImportParentsAndAssignToStudentMethod(), getCallOptions()), request);
    }

    /**
     */
    @java.lang.Deprecated
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateStudentResponse> updateStudent(
        io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateStudentRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUpdateStudentMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateParentsAndFamilyRelationshipResponse> updateParentsAndFamilyRelationship(
        io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateParentsAndFamilyRelationshipRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUpdateParentsAndFamilyRelationshipMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.quarkus.proto.usermgmt.v2.Users.ReissueUserPasswordResponse> reissueUserPassword(
        io.manabie.quarkus.proto.usermgmt.v2.Users.ReissueUserPasswordRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getReissueUserPasswordMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.quarkus.proto.usermgmt.v2.Users.UpsertStudentCoursePackageResponse> upsertStudentCoursePackage(
        io.manabie.quarkus.proto.usermgmt.v2.Users.UpsertStudentCoursePackageRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUpsertStudentCoursePackageMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.quarkus.proto.usermgmt.v2.Users.RemoveParentFromStudentResponse> removeParentFromStudent(
        io.manabie.quarkus.proto.usermgmt.v2.Users.RemoveParentFromStudentRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getRemoveParentFromStudentMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserProfileResponse> updateUserProfile(
        io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserProfileRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUpdateUserProfileMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserDeviceTokenResponse> updateUserDeviceToken(
        io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserDeviceTokenRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUpdateUserDeviceTokenMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserLastLoginDateResponse> updateUserLastLoginDate(
        io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserLastLoginDateRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUpdateUserLastLoginDateMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<io.manabie.quarkus.proto.usermgmt.v2.Users.GenerateImportParentsAndAssignToStudentTemplateResponse> generateImportParentsAndAssignToStudentTemplate(
        io.manabie.quarkus.proto.usermgmt.v2.Users.GenerateImportParentsAndAssignToStudentTemplateRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGenerateImportParentsAndAssignToStudentTemplateMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_CREATE_STUDENT = 0;
  private static final int METHODID_CREATE_PARENTS_AND_ASSIGN_TO_STUDENT = 1;
  private static final int METHODID_IMPORT_PARENTS_AND_ASSIGN_TO_STUDENT = 2;
  private static final int METHODID_UPDATE_STUDENT = 3;
  private static final int METHODID_UPDATE_PARENTS_AND_FAMILY_RELATIONSHIP = 4;
  private static final int METHODID_REISSUE_USER_PASSWORD = 5;
  private static final int METHODID_UPSERT_STUDENT_COURSE_PACKAGE = 6;
  private static final int METHODID_REMOVE_PARENT_FROM_STUDENT = 7;
  private static final int METHODID_UPDATE_USER_PROFILE = 8;
  private static final int METHODID_UPDATE_USER_DEVICE_TOKEN = 9;
  private static final int METHODID_UPDATE_USER_LAST_LOGIN_DATE = 10;
  private static final int METHODID_GENERATE_IMPORT_PARENTS_AND_ASSIGN_TO_STUDENT_TEMPLATE = 11;

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
        case METHODID_CREATE_STUDENT:
          serviceImpl.createStudent((io.manabie.quarkus.proto.usermgmt.v2.Users.CreateStudentRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.CreateStudentResponse>) responseObserver);
          break;
        case METHODID_CREATE_PARENTS_AND_ASSIGN_TO_STUDENT:
          serviceImpl.createParentsAndAssignToStudent((io.manabie.quarkus.proto.usermgmt.v2.Users.CreateParentsAndAssignToStudentRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.CreateParentsAndAssignToStudentResponse>) responseObserver);
          break;
        case METHODID_IMPORT_PARENTS_AND_ASSIGN_TO_STUDENT:
          serviceImpl.importParentsAndAssignToStudent((io.manabie.quarkus.proto.usermgmt.v2.Users.ImportParentsAndAssignToStudentRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.ImportParentsAndAssignToStudentResponse>) responseObserver);
          break;
        case METHODID_UPDATE_STUDENT:
          serviceImpl.updateStudent((io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateStudentRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateStudentResponse>) responseObserver);
          break;
        case METHODID_UPDATE_PARENTS_AND_FAMILY_RELATIONSHIP:
          serviceImpl.updateParentsAndFamilyRelationship((io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateParentsAndFamilyRelationshipRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateParentsAndFamilyRelationshipResponse>) responseObserver);
          break;
        case METHODID_REISSUE_USER_PASSWORD:
          serviceImpl.reissueUserPassword((io.manabie.quarkus.proto.usermgmt.v2.Users.ReissueUserPasswordRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.ReissueUserPasswordResponse>) responseObserver);
          break;
        case METHODID_UPSERT_STUDENT_COURSE_PACKAGE:
          serviceImpl.upsertStudentCoursePackage((io.manabie.quarkus.proto.usermgmt.v2.Users.UpsertStudentCoursePackageRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.UpsertStudentCoursePackageResponse>) responseObserver);
          break;
        case METHODID_REMOVE_PARENT_FROM_STUDENT:
          serviceImpl.removeParentFromStudent((io.manabie.quarkus.proto.usermgmt.v2.Users.RemoveParentFromStudentRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.RemoveParentFromStudentResponse>) responseObserver);
          break;
        case METHODID_UPDATE_USER_PROFILE:
          serviceImpl.updateUserProfile((io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserProfileRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserProfileResponse>) responseObserver);
          break;
        case METHODID_UPDATE_USER_DEVICE_TOKEN:
          serviceImpl.updateUserDeviceToken((io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserDeviceTokenRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserDeviceTokenResponse>) responseObserver);
          break;
        case METHODID_UPDATE_USER_LAST_LOGIN_DATE:
          serviceImpl.updateUserLastLoginDate((io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserLastLoginDateRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserLastLoginDateResponse>) responseObserver);
          break;
        case METHODID_GENERATE_IMPORT_PARENTS_AND_ASSIGN_TO_STUDENT_TEMPLATE:
          serviceImpl.generateImportParentsAndAssignToStudentTemplate((io.manabie.quarkus.proto.usermgmt.v2.Users.GenerateImportParentsAndAssignToStudentTemplateRequest) request,
              (io.grpc.stub.StreamObserver<io.manabie.quarkus.proto.usermgmt.v2.Users.GenerateImportParentsAndAssignToStudentTemplateResponse>) responseObserver);
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
          getCreateStudentMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.quarkus.proto.usermgmt.v2.Users.CreateStudentRequest,
              io.manabie.quarkus.proto.usermgmt.v2.Users.CreateStudentResponse>(
                service, METHODID_CREATE_STUDENT)))
        .addMethod(
          getCreateParentsAndAssignToStudentMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.quarkus.proto.usermgmt.v2.Users.CreateParentsAndAssignToStudentRequest,
              io.manabie.quarkus.proto.usermgmt.v2.Users.CreateParentsAndAssignToStudentResponse>(
                service, METHODID_CREATE_PARENTS_AND_ASSIGN_TO_STUDENT)))
        .addMethod(
          getImportParentsAndAssignToStudentMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.quarkus.proto.usermgmt.v2.Users.ImportParentsAndAssignToStudentRequest,
              io.manabie.quarkus.proto.usermgmt.v2.Users.ImportParentsAndAssignToStudentResponse>(
                service, METHODID_IMPORT_PARENTS_AND_ASSIGN_TO_STUDENT)))
        .addMethod(
          getUpdateStudentMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateStudentRequest,
              io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateStudentResponse>(
                service, METHODID_UPDATE_STUDENT)))
        .addMethod(
          getUpdateParentsAndFamilyRelationshipMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateParentsAndFamilyRelationshipRequest,
              io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateParentsAndFamilyRelationshipResponse>(
                service, METHODID_UPDATE_PARENTS_AND_FAMILY_RELATIONSHIP)))
        .addMethod(
          getReissueUserPasswordMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.quarkus.proto.usermgmt.v2.Users.ReissueUserPasswordRequest,
              io.manabie.quarkus.proto.usermgmt.v2.Users.ReissueUserPasswordResponse>(
                service, METHODID_REISSUE_USER_PASSWORD)))
        .addMethod(
          getUpsertStudentCoursePackageMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.quarkus.proto.usermgmt.v2.Users.UpsertStudentCoursePackageRequest,
              io.manabie.quarkus.proto.usermgmt.v2.Users.UpsertStudentCoursePackageResponse>(
                service, METHODID_UPSERT_STUDENT_COURSE_PACKAGE)))
        .addMethod(
          getRemoveParentFromStudentMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.quarkus.proto.usermgmt.v2.Users.RemoveParentFromStudentRequest,
              io.manabie.quarkus.proto.usermgmt.v2.Users.RemoveParentFromStudentResponse>(
                service, METHODID_REMOVE_PARENT_FROM_STUDENT)))
        .addMethod(
          getUpdateUserProfileMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserProfileRequest,
              io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserProfileResponse>(
                service, METHODID_UPDATE_USER_PROFILE)))
        .addMethod(
          getUpdateUserDeviceTokenMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserDeviceTokenRequest,
              io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserDeviceTokenResponse>(
                service, METHODID_UPDATE_USER_DEVICE_TOKEN)))
        .addMethod(
          getUpdateUserLastLoginDateMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserLastLoginDateRequest,
              io.manabie.quarkus.proto.usermgmt.v2.Users.UpdateUserLastLoginDateResponse>(
                service, METHODID_UPDATE_USER_LAST_LOGIN_DATE)))
        .addMethod(
          getGenerateImportParentsAndAssignToStudentTemplateMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              io.manabie.quarkus.proto.usermgmt.v2.Users.GenerateImportParentsAndAssignToStudentTemplateRequest,
              io.manabie.quarkus.proto.usermgmt.v2.Users.GenerateImportParentsAndAssignToStudentTemplateResponse>(
                service, METHODID_GENERATE_IMPORT_PARENTS_AND_ASSIGN_TO_STUDENT_TEMPLATE)))
        .build();
  }

  private static abstract class UserModifierServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    UserModifierServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return io.manabie.quarkus.proto.usermgmt.v2.Users.getDescriptor();
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
              .addMethod(getCreateStudentMethod())
              .addMethod(getCreateParentsAndAssignToStudentMethod())
              .addMethod(getImportParentsAndAssignToStudentMethod())
              .addMethod(getUpdateStudentMethod())
              .addMethod(getUpdateParentsAndFamilyRelationshipMethod())
              .addMethod(getReissueUserPasswordMethod())
              .addMethod(getUpsertStudentCoursePackageMethod())
              .addMethod(getRemoveParentFromStudentMethod())
              .addMethod(getUpdateUserProfileMethod())
              .addMethod(getUpdateUserDeviceTokenMethod())
              .addMethod(getUpdateUserLastLoginDateMethod())
              .addMethod(getGenerateImportParentsAndAssignToStudentTemplateMethod())
              .build();
        }
      }
    }
    return result;
  }
}
