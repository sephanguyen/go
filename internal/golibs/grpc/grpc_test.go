package grpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGRPCServicesFromFullMethods(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		name     string
		methods  []string
		hasError bool
		expected Services
	}{
		{
			name: "happy case",
			methods: []string{
				"bob.v1.LessonModifierService/GetLiveLessonState",
				"bob.v1.LessonModifierService/ModifyLiveLessonState",
				"manabie.bob.Course/RetrieveLiveLesson",
				"bob.v1.CourseReaderService/ListCourses",
				"bob.v1.ClassModifierService/JoinLesson",
				"bob.v1.ClassReaderService/ListStudentsByLesson",
				"bob.v1.CourseReaderService/ListLessonMedias",
				"bob.v1.ClassModifierService/LeaveLesson",
				"bob.v1.ClassModifierService/EndLiveLesson",
				"manabie.bob.Student/GetStudentProfile",
				"bob.v1.LessonModifierService/PreparePublish",
				"bob.v1.LessonModifierService/Unpublish",
			},
			expected: Services{
				"bob.v1.LessonModifierService": &Service{
					ServiceName: "LessonModifierService",
					MethodNamesMap: map[string]bool{
						"GetLiveLessonState":    true,
						"ModifyLiveLessonState": true,
						"PreparePublish":        true,
						"Unpublish":             true,
					},
				},
				"manabie.bob.Course": &Service{
					ServiceName: "Course",
					MethodNamesMap: map[string]bool{
						"RetrieveLiveLesson": true,
					},
				},
				"bob.v1.CourseReaderService": &Service{
					ServiceName: "CourseReaderService",
					MethodNamesMap: map[string]bool{
						"ListCourses":      true,
						"ListLessonMedias": true,
					},
				},
				"bob.v1.ClassModifierService": &Service{
					ServiceName: "ClassModifierService",
					MethodNamesMap: map[string]bool{
						"JoinLesson":    true,
						"LeaveLesson":   true,
						"EndLiveLesson": true,
					},
				},
				"bob.v1.ClassReaderService": &Service{
					ServiceName: "ClassReaderService",
					MethodNamesMap: map[string]bool{
						"ListStudentsByLesson": true,
					},
				},
				"manabie.bob.Student": &Service{
					ServiceName: "Student",
					MethodNamesMap: map[string]bool{
						"GetStudentProfile": true,
					},
				},
			},
		},
		{
			name: "invalid grpc method",
			methods: []string{
				"bob.v1.LessonModifierService-GetLiveLessonState",
			},
			hasError: true,
		},
		{
			name: "invalid case: missing method name",
			methods: []string{
				"bob.v1.LessonModifierService/",
			},
			hasError: true,
		},
		{
			name: "invalid case: missing service name",
			methods: []string{
				"bob.v1./GetLiveLessonState",
			},
			hasError: true,
		},
		{
			name: "invalid case: missing package name",
			methods: []string{
				".LessonModifierService/GetLiveLessonState",
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := ServicesFromFullMethods(tc.methods)
			if tc.hasError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, actual, len(tc.expected))
			for serviceName, service := range actual {
				require.NotNil(t, tc.expected[serviceName])
				for _, methodName := range service.MethodNames() {
					assert.True(t, tc.expected[serviceName].MethodNamesMap[methodName])
				}
			}
		})
	}
}

func TestGRPCServices_RemoveBuFullMethodNames(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		name            string
		deletedMethods  []string
		originList      Services
		expectedMethods []string
		hasError        bool
	}{
		{
			name: "happy case",
			deletedMethods: []string{
				"bob.v1.LessonModifierService/GetLiveLessonState",
				"bob.v1.ClassReaderService/ListStudentsByLesson",
				"bob.v1.CourseReaderService/ListLessonMedias",
				"bob.v1.LessonModifierService/Unpublish",
			},
			originList: Services{
				"bob.v1.LessonModifierService": &Service{
					ServiceName: "LessonModifierService",
					MethodNamesMap: map[string]bool{
						"GetLiveLessonState":    true,
						"ModifyLiveLessonState": true,
						"PreparePublish":        true,
						"Unpublish":             true,
					},
				},
				"manabie.bob.Course": &Service{
					ServiceName: "Course",
					MethodNamesMap: map[string]bool{
						"RetrieveLiveLesson": true,
					},
				},
				"bob.v1.CourseReaderService": &Service{
					ServiceName: "CourseReaderService",
					MethodNamesMap: map[string]bool{
						"ListCourses":      true,
						"ListLessonMedias": true,
					},
				},
				"bob.v1.ClassModifierService": &Service{
					ServiceName: "ClassModifierService",
					MethodNamesMap: map[string]bool{
						"JoinLesson":    true,
						"LeaveLesson":   true,
						"EndLiveLesson": true,
					},
				},
				"bob.v1.ClassReaderService": &Service{
					ServiceName: "ClassReaderService",
					MethodNamesMap: map[string]bool{
						"ListStudentsByLesson": true,
					},
				},
				"manabie.bob.Student": &Service{
					ServiceName: "Student",
					MethodNamesMap: map[string]bool{
						"GetStudentProfile": true,
					},
				},
			},
			expectedMethods: []string{
				"bob.v1.LessonModifierService/ModifyLiveLessonState",
				"manabie.bob.Course/RetrieveLiveLesson",
				"bob.v1.CourseReaderService/ListCourses",
				"bob.v1.ClassModifierService/JoinLesson",
				"bob.v1.ClassModifierService/LeaveLesson",
				"bob.v1.ClassModifierService/EndLiveLesson",
				"manabie.bob.Student/GetStudentProfile",
				"bob.v1.LessonModifierService/PreparePublish",
			},
		},
		{
			name:           "empty deleted methods",
			deletedMethods: []string{},
			originList: Services{
				"bob.v1.LessonModifierService": &Service{
					ServiceName: "LessonModifierService",
					MethodNamesMap: map[string]bool{
						"GetLiveLessonState":    true,
						"ModifyLiveLessonState": true,
						"PreparePublish":        true,
						"Unpublish":             true,
					},
				},
				"manabie.bob.Course": &Service{
					ServiceName: "Course",
					MethodNamesMap: map[string]bool{
						"RetrieveLiveLesson": true,
					},
				},
				"bob.v1.CourseReaderService": &Service{
					ServiceName: "CourseReaderService",
					MethodNamesMap: map[string]bool{
						"ListCourses":      true,
						"ListLessonMedias": true,
					},
				},
				"bob.v1.ClassModifierService": &Service{
					ServiceName: "ClassModifierService",
					MethodNamesMap: map[string]bool{
						"JoinLesson":    true,
						"LeaveLesson":   true,
						"EndLiveLesson": true,
					},
				},
				"bob.v1.ClassReaderService": &Service{
					ServiceName: "ClassReaderService",
					MethodNamesMap: map[string]bool{
						"ListStudentsByLesson": true,
					},
				},
				"manabie.bob.Student": &Service{
					ServiceName: "Student",
					MethodNamesMap: map[string]bool{
						"GetStudentProfile": true,
					},
				},
			},
			expectedMethods: []string{
				"bob.v1.LessonModifierService/GetLiveLessonState",
				"bob.v1.LessonModifierService/ModifyLiveLessonState",
				"manabie.bob.Course/RetrieveLiveLesson",
				"bob.v1.CourseReaderService/ListCourses",
				"bob.v1.ClassModifierService/JoinLesson",
				"bob.v1.ClassReaderService/ListStudentsByLesson",
				"bob.v1.CourseReaderService/ListLessonMedias",
				"bob.v1.ClassModifierService/LeaveLesson",
				"bob.v1.ClassModifierService/EndLiveLesson",
				"manabie.bob.Student/GetStudentProfile",
				"bob.v1.LessonModifierService/PreparePublish",
				"bob.v1.LessonModifierService/Unpublish",
			},
		},
		{
			name: "deleted method more origin list",
			deletedMethods: []string{
				"bob.v1.LessonModifierService/GetLiveLessonState",
				"bob.v1.LessonModifierService/ModifyLiveLessonState",
				"manabie.bob.Course/RetrieveLiveLesson",
				"bob.v1.CourseReaderService/ListCourses",
				"bob.v1.ClassModifierService/JoinLesson",
				"bob.v1.ClassReaderService/ListStudentsByLesson",
				"bob.v1.CourseReaderService/ListLessonMedias",
				"bob.v1.ClassModifierService/LeaveLesson",
				"bob.v1.ClassModifierService/EndLiveLesson",
				"manabie.bob.Student/GetStudentProfile",
				"bob.v1.LessonModifierService/PreparePublish",
			},
			originList: Services{
				"bob.v1.LessonModifierService": &Service{
					ServiceName: "LessonModifierService",
					MethodNamesMap: map[string]bool{
						"GetLiveLessonState": true,
						"Unpublish":          true,
					},
				},
			},
			expectedMethods: []string{
				"bob.v1.LessonModifierService/Unpublish",
			},
		},
		{
			name: "invalid deleted methods",
			deletedMethods: []string{
				"bob.v1.LessonModifierService/",
			},
			originList: Services{
				"bob.v1.LessonModifierService": &Service{
					ServiceName: "LessonModifierService",
					MethodNamesMap: map[string]bool{
						"GetLiveLessonState":    true,
						"ModifyLiveLessonState": true,
						"PreparePublish":        true,
						"Unpublish":             true,
					},
				},
				"manabie.bob.Course": &Service{
					ServiceName: "Course",
					MethodNamesMap: map[string]bool{
						"RetrieveLiveLesson": true,
					},
				},
				"bob.v1.CourseReaderService": &Service{
					ServiceName: "CourseReaderService",
					MethodNamesMap: map[string]bool{
						"ListCourses":      true,
						"ListLessonMedias": true,
					},
				},
				"bob.v1.ClassModifierService": &Service{
					ServiceName: "ClassModifierService",
					MethodNamesMap: map[string]bool{
						"JoinLesson":    true,
						"LeaveLesson":   true,
						"EndLiveLesson": true,
					},
				},
				"bob.v1.ClassReaderService": &Service{
					ServiceName: "ClassReaderService",
					MethodNamesMap: map[string]bool{
						"ListStudentsByLesson": true,
					},
				},
				"manabie.bob.Student": &Service{
					ServiceName: "Student",
					MethodNamesMap: map[string]bool{
						"GetStudentProfile": true,
					},
				},
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.originList.RemoveByFullMethodNames(tc.deletedMethods)
			if tc.hasError {
				require.Error(t, err)
				return
			}
			assert.ElementsMatch(t, tc.originList.GRPCMethods(), tc.expectedMethods)
		})
	}
}
