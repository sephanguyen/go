package grafanabuilder

import (
	"bytes"
	"crypto/sha256"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const ListMethodsOfClassProto = `bob.v1.ClassReaderService/RetrieveStudentLearningStatistics|bob.v1.ClassReaderService/RetrieveStudyPlanItemEventLogs|bob.v1.ClassReaderService/RetrieveClassByIDs|bob.v1.ClassReaderService/ListStudentsByLesson|bob.v1.ClassReaderService/ListClass|bob.v1.ClassReaderService/RetrieveClassMembers|bob.v1.ClassReaderService/RetrieveClassLearningStatistics|bob.v1.ClassModifierService/LeaveClass|bob.v1.ClassModifierService/AddClassOwner|bob.v1.ClassModifierService/AddClassMember|bob.v1.ClassModifierService/LeaveLesson|bob.v1.ClassModifierService/EditClass|bob.v1.ClassModifierService/UpdateClassCode|bob.v1.ClassModifierService/JoinClass|bob.v1.ClassModifierService/RemoveClassMember|bob.v1.ClassModifierService/EndLiveLesson|bob.v1.ClassModifierService/JoinLesson|bob.v1.ClassModifierService/ConvertMedia|bob.v1.ClassModifierService/CreateClass`

func TestGetBasicGrafanaDashboardConfig(t *testing.T) {
	t.Run("empty have service names", func(t *testing.T) {
		main, extension, err := NewGRPCDashboardConfig().
			AddServiceNames("bob", "tom").
			Build()
		require.NoError(t, err)
		assert.NotNil(t, main)
		assert.NotNil(t, extension)

		// validate new dashboard config json file
		f1, err := os.Open("config/default_dashboard_cfg.jsonnet")
		require.NoError(t, err)
		defer f1.Close()

		h1 := sha256.New()
		_, err = io.Copy(h1, f1)
		require.NoError(t, err)

		h2 := sha256.New()
		_, err = io.Copy(h2, main)
		require.NoError(t, err)

		assert.Equal(t, 0, bytes.Compare(h1.Sum(nil), h2.Sum(nil)))

		// validate extension file
		assert.NotNil(t, extension["panel_target.jsonnet"])
		ext := extension["panel_target.jsonnet"]
		data := &bytes.Buffer{}
		_, err = data.ReadFrom(ext)
		require.NoError(t, err)

		dataString := data.String()
		assert.True(t, strings.Contains(dataString, `app_kubernetes_io_name=~"bob|tom"`))
		assert.True(t, strings.Contains(dataString, `pod=~"^(bob|tom).+"`))
		assert.False(t, strings.Contains(dataString, `grpc_server_method=~`))
		assert.False(t, strings.Contains(dataString, `grpc_server_method!~`))
	})

	t.Run("there is only exception methods", func(t *testing.T) {
		main, extension, err := NewGRPCDashboardConfig().
			AddServiceNames("bob", "tom").
			AddExceptionMethods("bob.service1/method1", "bob.service1/method2").
			Build()
		require.NoError(t, err)
		assert.NotNil(t, main)
		assert.NotNil(t, extension)

		// validate new dashboard config json file
		f1, err := os.Open("config/default_dashboard_cfg.jsonnet")
		require.NoError(t, err)
		defer f1.Close()

		h1 := sha256.New()
		_, err = io.Copy(h1, f1)
		require.NoError(t, err)

		h2 := sha256.New()
		_, err = io.Copy(h2, main)
		require.NoError(t, err)

		assert.Equal(t, 0, bytes.Compare(h1.Sum(nil), h2.Sum(nil)))

		// validate extension file
		assert.NotNil(t, extension["panel_target.jsonnet"])
		ext := extension["panel_target.jsonnet"]
		data := &bytes.Buffer{}
		_, err = data.ReadFrom(ext)
		require.NoError(t, err)

		dataString := data.String()
		assert.True(t, strings.Contains(dataString, `app_kubernetes_io_name=~"bob|tom"`))
		assert.True(t, strings.Contains(dataString, `pod=~"^(bob|tom).+"`))
		assert.False(t, strings.Contains(dataString, `grpc_server_method=~`))
		assert.True(t, strings.Contains(dataString, `grpc_server_method!~"bob.service1/`))
		assert.True(t, strings.Contains(dataString, `bob.service1/method1`))
		assert.True(t, strings.Contains(dataString, `bob.service1/method2`))
	})

	t.Run("there is only methods", func(t *testing.T) {
		main, extension, err := NewGRPCDashboardConfig().
			AddServiceNames("bob", "tom").
			AddGRPCMethods("bob.service1/method1", "bob.service1/method2").
			Build()
		require.NoError(t, err)
		assert.NotNil(t, main)
		assert.NotNil(t, extension)

		// validate new dashboard config json file
		f1, err := os.Open("config/default_dashboard_cfg.jsonnet")
		require.NoError(t, err)
		defer f1.Close()

		h1 := sha256.New()
		_, err = io.Copy(h1, f1)
		require.NoError(t, err)

		h2 := sha256.New()
		_, err = io.Copy(h2, main)
		require.NoError(t, err)

		assert.Equal(t, 0, bytes.Compare(h1.Sum(nil), h2.Sum(nil)))

		// validate extension file
		assert.NotNil(t, extension["panel_target.jsonnet"])
		ext := extension["panel_target.jsonnet"]
		data := &bytes.Buffer{}
		_, err = data.ReadFrom(ext)
		require.NoError(t, err)

		dataString := data.String()
		assert.True(t, strings.Contains(dataString, `app_kubernetes_io_name=~"bob|tom"`))
		assert.True(t, strings.Contains(dataString, `pod=~"^(bob|tom).+"`))
		assert.True(t, strings.Contains(dataString, `grpc_server_method=~"bob.service1`))
		assert.True(t, strings.Contains(dataString, `bob.service1/method1`))
		assert.True(t, strings.Contains(dataString, `bob.service1/method2`))
		assert.False(t, strings.Contains(dataString, `grpc_server_method!~`))
	})

	t.Run("there is only proto file", func(t *testing.T) {
		main, extension, err := NewGRPCDashboardConfig().
			AddServiceNames("bob", "tom").
			AddProtoFile("testdata/classes.proto").
			Build()
		require.NoError(t, err)
		assert.NotNil(t, main)
		assert.NotNil(t, extension)

		// validate new dashboard config json file
		f1, err := os.Open("config/default_dashboard_cfg.jsonnet")
		require.NoError(t, err)
		defer f1.Close()

		h1 := sha256.New()
		_, err = io.Copy(h1, f1)
		require.NoError(t, err)

		h2 := sha256.New()
		_, err = io.Copy(h2, main)
		require.NoError(t, err)

		assert.Equal(t, 0, bytes.Compare(h1.Sum(nil), h2.Sum(nil)))

		// validate extension file
		assert.NotNil(t, extension["panel_target.jsonnet"])
		ext := extension["panel_target.jsonnet"]
		data := &bytes.Buffer{}
		_, err = data.ReadFrom(ext)
		require.NoError(t, err)

		dataString := data.String()
		assert.True(t, strings.Contains(dataString, `app_kubernetes_io_name=~"bob|tom"`))
		assert.True(t, strings.Contains(dataString, `pod=~"^(bob|tom).+"`))
		expectedMethods := strings.Split(ListMethodsOfClassProto, "|")
		for _, expectedMethod := range expectedMethods {
			assert.True(t, strings.Contains(dataString, expectedMethod))
		}
		assert.True(t, strings.Contains(dataString, `grpc_server_method=~"`))
		assert.False(t, strings.Contains(dataString, `grpc_server_method!~`))
	})

	t.Run("just get exception methods when there are exception methods, methods and protoFiles", func(t *testing.T) {
		main, extension, err := NewGRPCDashboardConfig().
			AddServiceNames("bob", "tom").
			AddExceptionMethods("bob.service1/method1", "bob.service1/method2").
			AddGRPCMethods("bob.service2/method1", "bob.service2/method2").
			AddProtoFile("testdata/classes.proto").
			Build()
		require.NoError(t, err)
		assert.NotNil(t, main)
		assert.NotNil(t, extension)

		// validate new dashboard config json file
		f1, err := os.Open("config/default_dashboard_cfg.jsonnet")
		require.NoError(t, err)
		defer f1.Close()

		h1 := sha256.New()
		_, err = io.Copy(h1, f1)
		require.NoError(t, err)

		h2 := sha256.New()
		_, err = io.Copy(h2, main)
		require.NoError(t, err)

		assert.Equal(t, 0, bytes.Compare(h1.Sum(nil), h2.Sum(nil)))

		// validate extension file
		assert.NotNil(t, extension["panel_target.jsonnet"])
		ext := extension["panel_target.jsonnet"]
		data := &bytes.Buffer{}
		_, err = data.ReadFrom(ext)
		require.NoError(t, err)

		// just get exception methods
		dataString := data.String()
		assert.True(t, strings.Contains(dataString, `app_kubernetes_io_name=~"bob|tom"`))
		assert.True(t, strings.Contains(dataString, `pod=~"^(bob|tom).+"`))
		assert.False(t, strings.Contains(dataString, `grpc_server_method=~`))
		assert.True(t, strings.Contains(dataString, `grpc_server_method!~"bob.service1/`))
		assert.True(t, strings.Contains(dataString, `bob.service1/method1`))
		assert.True(t, strings.Contains(dataString, `bob.service1/method2`))
	})

	t.Run("just get exception regex methods when there are regex methods, exception methods, methods and protoFiles", func(t *testing.T) {
		main, extension, err := NewGRPCDashboardConfig().
			AddServiceNames("bob", "tom").
			AddExceptionRegexMethods("grpc.health.v1.Health/Check|.+TopicIcon.+").
			AddExceptionMethods("bob.service1/method1", "bob.service1/method2").
			AddGRPCMethods("bob.service2/method1", "bob.service2/method2").
			AddProtoFile("testdata/classes.proto").
			Build()
		require.NoError(t, err)
		assert.NotNil(t, main)
		assert.NotNil(t, extension)

		// validate new dashboard config json file
		f1, err := os.Open("config/default_dashboard_cfg.jsonnet")
		require.NoError(t, err)
		defer f1.Close()

		h1 := sha256.New()
		_, err = io.Copy(h1, f1)
		require.NoError(t, err)

		h2 := sha256.New()
		_, err = io.Copy(h2, main)
		require.NoError(t, err)

		assert.Equal(t, 0, bytes.Compare(h1.Sum(nil), h2.Sum(nil)))

		// validate extension file
		assert.NotNil(t, extension["panel_target.jsonnet"])
		ext := extension["panel_target.jsonnet"]
		data := &bytes.Buffer{}
		_, err = data.ReadFrom(ext)
		require.NoError(t, err)

		// just get regex exception methods
		dataString := data.String()
		assert.True(t, strings.Contains(dataString, `app_kubernetes_io_name=~"bob|tom"`))
		assert.True(t, strings.Contains(dataString, `pod=~"^(bob|tom).+"`))
		assert.False(t, strings.Contains(dataString, `grpc_server_method=~`))
		assert.False(t, strings.Contains(dataString, `grpc_server_method!~"bob.service1/`))
		assert.False(t, strings.Contains(dataString, `bob.service1/method1`))
		assert.False(t, strings.Contains(dataString, `bob.service1/method2`))
		assert.True(t, strings.Contains(dataString, `grpc_server_method!~"grpc.health.v1.Health/Check|.+TopicIcon.+`))
	})

	t.Run("input both methods and protoFiles, with these methods are defined in protoFiles", func(t *testing.T) {
		main, extension, err := NewGRPCDashboardConfig().
			AddServiceNames("bob", "tom").
			AddGRPCMethods("bob.v1.ClassReaderService/RetrieveStudyPlanItemEventLogs", "bob.v1.ClassReaderService/RetrieveClassByIDs").
			AddProtoFile("testdata/classes.proto").
			Build()
		require.NoError(t, err)
		assert.NotNil(t, main)
		assert.NotNil(t, extension)

		// validate new dashboard config json file
		f1, err := os.Open("config/default_dashboard_cfg.jsonnet")
		require.NoError(t, err)
		defer f1.Close()

		h1 := sha256.New()
		_, err = io.Copy(h1, f1)
		require.NoError(t, err)

		h2 := sha256.New()
		_, err = io.Copy(h2, main)
		require.NoError(t, err)

		assert.Equal(t, 0, bytes.Compare(h1.Sum(nil), h2.Sum(nil)))

		//validate extension file
		assert.NotNil(t, extension["panel_target.jsonnet"])
		ext := extension["panel_target.jsonnet"]
		data := &bytes.Buffer{}
		_, err = data.ReadFrom(ext)
		require.NoError(t, err)

		dataString := data.String()
		assert.True(t, strings.Contains(dataString, `app_kubernetes_io_name=~"bob|tom"`))
		assert.True(t, strings.Contains(dataString, `pod=~"^(bob|tom).+"`))
		assert.True(t, strings.Contains(dataString, `grpc_server_method=~"bob.v1.ClassReaderService/`))
		assert.True(t, strings.Contains(dataString, `bob.v1.ClassReaderService/RetrieveStudyPlanItemEventLogs`))
		assert.True(t, strings.Contains(dataString, `bob.v1.ClassReaderService/RetrieveClassByIDs`))
		assert.False(t, strings.Contains(dataString, `grpc_server_method!~`))
		expectedMethods := strings.Split(ListMethodsOfClassProto, "|")
		for _, expectedMethod := range expectedMethods {
			if expectedMethod == "bob.v1.ClassReaderService/RetrieveStudyPlanItemEventLogs" || expectedMethod == "bob.v1.ClassReaderService/RetrieveClassByIDs" {
				continue
			}
			assert.False(t, strings.Contains(dataString, expectedMethod))
		}
	})

	t.Run("input both methods and protoFiles, but these methods are not defined in protoFiles", func(t *testing.T) {
		main, extension, err := NewGRPCDashboardConfig().
			AddServiceNames("bob", "tom").
			AddGRPCMethods("bob.v1.ClassReaderService/RetrieveStudyPlanItemEventLogs", "bob.v1.ClassReaderService/RetrieveClassByIDs", "bob.v1.ClassModifierService/NotDefinedMethod").
			AddProtoFile("testdata/classes.proto").
			Build()
		require.Error(t, err)
		assert.Nil(t, main)
		assert.Nil(t, extension)
	})

	t.Run("Add Requests per seconds by methods panel", func(t *testing.T) {
		main, extension, err := NewGRPCDashboardConfig().
			AddServiceNames("bob", "tom").
			AddRequestsPerSecondsByMethodsPanel().
			Build()
		require.NoError(t, err)
		assert.NotNil(t, main)
		assert.NotNil(t, extension)

		// validate extension file
		assert.NotNil(t, extension["panel_target.jsonnet"])
		ext := extension["panel_target.jsonnet"]
		data := &bytes.Buffer{}
		_, err = data.ReadFrom(ext)
		require.NoError(t, err)

		dataString := data.String()
		assert.True(t, strings.Contains(dataString, `app_kubernetes_io_name=~"bob|tom"`))
		assert.True(t, strings.Contains(dataString, `pod=~"^(bob|tom).+"`))
		assert.False(t, strings.Contains(dataString, `grpc_server_method=~`))
		assert.False(t, strings.Contains(dataString, `grpc_server_method!~`))

		// validate file main have Requests per seconds by methods panel or not
		data = &bytes.Buffer{}
		_, err = data.ReadFrom(main)
		require.NoError(t, err)
		dataString = data.String()
		assert.True(t, strings.Contains(dataString, `Requests per seconds by methods`))
	})
}

func TestUseNewGRPCDashboardConfigToGenConfigJson(t *testing.T) {
	methods := strings.Split("bob.v1.LessonModifierService/GetLiveLessonState|bob.v1.LessonModifierService/ModifyLiveLessonState|manabie.bob.Course/RetrieveLiveLesson|bob.v1.CourseReaderService/ListCourses|bob.v1.ClassModifierService/JoinLesson|bob.v1.ClassReaderService/ListStudentsByLesson|bob.v1.CourseReaderService/ListLessonMedias|bob.v1.ClassModifierService/LeaveLesson|bob.v1.ClassModifierService/EndLiveLesson|manabie.bob.Student/GetStudentProfile|bob.v1.LessonModifierService/PreparePublish|bob.v1.LessonModifierService/Unpublish", "|")
	main, extension, err := NewGRPCDashboardConfig().
		AddUIDAndTitle("id-111", "hello world").
		AddServiceNames("bob", "virtualclassroom").
		AddRequestsPerSecondsByMethodsPanel().
		AddGRPCMethods(methods...).
		Build()
	require.NoError(t, err)
	assert.NotNil(t, main)
	assert.NotNil(t, extension)

	tmpFolder := "tmp_" + idutil.ULIDNow()
	err = os.Mkdir(tmpFolder, os.ModePerm)
	require.NoError(t, err)
	defer os.RemoveAll(tmpFolder)

	err = (&Builder{}).
		AddDashboardConfigFiles(main, extension).
		AddDestinationFilePath(tmpFolder + "/res.json").
		Build()
	require.NoError(t, err)

	f1, err := os.Open(tmpFolder + "/res.json")
	require.NoError(t, err)
	data := &bytes.Buffer{}
	_, err = data.ReadFrom(f1)
	require.NoError(t, err)

	// validate data of res.json file
	dataString := data.String()
	for _, method := range methods {
		assert.True(t, strings.Contains(dataString, method))
	}
	assert.True(t, strings.Contains(dataString, "id-111"))
	assert.True(t, strings.Contains(dataString, "hello world"))
	assert.True(t, strings.Contains(dataString, `app_kubernetes_io_name=~\"bob|virtualclassroom\"`))
	assert.True(t, strings.Contains(dataString, `pod=~\"^(bob|virtualclassroom).+\"`))
	assert.False(t, strings.Contains(dataString, `grpc_server_method!~`))
	assert.True(t, strings.Contains(dataString, `grpc_server_method=~\"`))

}

func TestUseHasuraDashboardConfigToGenConfigJson(t *testing.T) {
	main, exts, err := GetHasuraDashboard("", "bob-hasura")
	require.NoError(t, err)
	assert.NotNil(t, main)
	assert.NotNil(t, exts)

	tmpFolder := "tmp_" + idutil.ULIDNow()
	err = os.Mkdir(tmpFolder, os.ModePerm)
	require.NoError(t, err)
	defer os.RemoveAll(tmpFolder)

	err = (&Builder{}).
		AddDashboardConfigFiles(main, exts).
		AddDestinationFilePath(tmpFolder + "/res.json").
		Build()
	require.NoError(t, err)
}
