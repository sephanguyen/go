package grpc

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yoheimuta/go-protoparser"
)

func TestProtoBufVisitor(t *testing.T) {
	f, err := os.Open("testdata/classes.proto")
	require.NoError(t, err)

	got, err := protoparser.Parse(f)
	require.NoError(t, err)

	pgkName := getPackageName(got)
	assert.Equal(t, "bob.v1", pgkName)

	srv := gRPCServicesFromProto(got)
	actualSrc := make(Services)
	actualSrc.AddMethodByService("bob.v1", "ClassReaderService", "ListClass")
	actualSrc.AddMethodByService("bob.v1", "ClassReaderService", "RetrieveClassMembers")
	actualSrc.AddMethodByService("bob.v1", "ClassReaderService", "RetrieveClassLearningStatistics")
	actualSrc.AddMethodByService("bob.v1", "ClassReaderService", "RetrieveStudentLearningStatistics")
	actualSrc.AddMethodByService("bob.v1", "ClassReaderService", "RetrieveClassByIDs")
	actualSrc.AddMethodByService("bob.v1", "ClassReaderService", "ListStudentsByLesson")

	actualSrc.AddMethodByService("bob.v1", "ClassModifierService", "CreateClass")
	actualSrc.AddMethodByService("bob.v1", "ClassModifierService", "EditClass")
	actualSrc.AddMethodByService("bob.v1", "ClassModifierService", "UpdateClassCode")
	actualSrc.AddMethodByService("bob.v1", "ClassModifierService", "JoinClass")
	actualSrc.AddMethodByService("bob.v1", "ClassModifierService", "LeaveClass")
	actualSrc.AddMethodByService("bob.v1", "ClassModifierService", "AddClassOwner")
	actualSrc.AddMethodByService("bob.v1", "ClassModifierService", "AddClassMember")
	actualSrc.AddMethodByService("bob.v1", "ClassModifierService", "RemoveClassMember")
	actualSrc.AddMethodByService("bob.v1", "ClassModifierService", "EndLiveLesson")
	actualSrc.AddMethodByService("bob.v1", "ClassModifierService", "JoinLesson")
	actualSrc.AddMethodByService("bob.v1", "ClassModifierService", "LeaveLesson")
	actualSrc.AddMethodByService("bob.v1", "ClassModifierService", "ConvertMedia")

	assert.Len(t, srv, 2)
	for serviceName, service := range actualSrc {
		assert.NotNil(t, srv[serviceName])
		for _, methodName := range service.MethodNames() {
			assert.True(t, srv[serviceName].MethodNamesMap[methodName])
		}
	}
}
