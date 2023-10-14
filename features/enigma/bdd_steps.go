package enigma

import (
	"regexp"
	"sync"

	"github.com/manabie-com/backend/features/helper"

	"github.com/cucumber/godog"
)

var (
	buildRegexpMapOnce sync.Once
	regexpMap          map[string]*regexp.Regexp
)

func initSteps(ctx *godog.ScenarioContext, s *suite) {
	steps := map[string]interface{}{
		`^everything is OK$`: s.everythingIsOK,
		`^a random number$`:  s.CommonSuite.ARandomNumber,

		`^health check endpoint called$`:                                                                        s.healthCheckEndpointCalled,
		`^returns "([^"]*)" status code$`:                                                                       s.CommonSuite.ReturnsStatusCode,
		`^a request with (\d+) student and (\d+) staff$`:                                                        s.aRequestWithStudentAndStaff,
		`^a request with (\d+) student and (\d+) staff with payload invalid$`:                                   s.aRequestWithStudentAndStaffInvalidPayload,
		`^a valid JPREP signature in its header$`:                                                               s.stepAValidJPREPSignatureInItsHeader,
		`^the request user registration is performed$`:                                                          s.stepPerformUserRegistrationRequest,
		`^the request master registration is performed$`:                                                        s.stepPerformMasterRegistrationRequest,
		`^a partner "([^"]*)" data sync log already exists in DB$`:                                              s.aPartnerDataSyncAlreadyExistInDB,
		`^a partner "([^"]*)" data sync log not exists in DB$`:                                                  s.aPartnerDataSyncNotExistInDB,
		`^a partner "([^"]*)" data sync logs split already exists (\d+) rows in DB$`:                            s.aPartnerDataSyncSplitAlreadyExistInDB,
		`^a signed in admin$`:                                                                                   s.CommonSuite.ASignedInAdmin,
		`^a signed in as "([^"]*)" with school "([^"]*)"$`:                                                      s.aSignedInWithSchool,
		`^a request with (\d+) course, (\d+) lesson, (\d+) class and (\d+) academic year$`:                      s.requestMasterRegistration,
		`^a request with (\d+) course, (\d+) lesson, (\d+) class and (\d+) academic year with payload invalid$`: s.requestMasterRegistrationInvalidPayload,
		`^a request with (\d+) student lessons with payload invalid$`:                                           s.requestUserCourseRegistrationInvalidPayload,
		`^a request with (\d+) student lessons$`:                                                                s.requestUserCourseRegistration,
		`^the request user course registration is performed$`:                                                   s.stepPerformUserCourseRegistrationRequest,
		`^a payload of "([^"]*)" data sync logs split match with request$`:                                      s.aPayloadLogsMatchWithRequest,
		`^a request get partner data logs report$`:                                                              s.aRequestGetPartnerDataReport,
		`^the request get partner log report is performed$`:                                                     s.theRequestGetPartnerLogReportIsPerformed,
		`^a response of "([^"]*)" partner log report match with DB$`:                                            s.aResponsePartnerLogReportMatchDB,
		`^a data sync log already of "([^"]*)" with "([^"]*)" and (\d+) exists in DB at "([^"]*)"$`:             s.someDataLogSplitOfKindWithStatusAndTry,
		`^request with recover data sync at "([^"]*)"$`:                                                         s.requestWithRecoverDataSync,
		`^the request recover data sync is performed$`:                                                          s.theRequestRecoverDataSyncIsPerformed,
		`^a partner "([^"]*)" data sync log split match with (\d+)$`:                                            s.aLogSplitMatchWithStatusAndRetryTimes,
	}

	buildRegexpMapOnce.Do(func() {
		regexpMap = helper.BuildRegexpMapV2(steps)
	})
	for k, v := range steps {
		ctx.Step(regexpMap[k], v)
	}
}
