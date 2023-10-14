package interceptors

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/mod/semver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	packageNameKey = "pkg"
	versionKey     = "version"
)

type checkAppVersion struct {
	versions        map[string]uint
	minVersionCheck uint
	decider         func(methodName string) bool
}

// NewCheckAppVersion creates new checkAppVersion
func NewCheckAppVersion(cfgClientVersion string, minVersionCheck string, methodIgnores []string) (*checkAppVersion, error) {
	versions, err := getAppVersion(cfgClientVersion)
	if err != nil {
		return nil, errors.Wrap(err, "getAppVersion")
	}

	a := &checkAppVersion{}
	a.versions = versions

	a.minVersionCheck, err = versionStringToInt(minVersionCheck)
	if err != nil {
		return nil, err
	}

	whiteList := map[string]bool{}
	for _, method := range methodIgnores {
		whiteList[method] = true
	}

	a.decider = func(methodName string) bool {
		_, ok := whiteList[methodName]
		return !ok
	}

	return a, nil
}

func (rcv *checkAppVersion) UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	ctx, span := StartSpan(ctx, "checkAppVersion.UnaryServerInterceptor")
	defer span.End()

	if !rcv.decider(info.FullMethod) {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot check gRPC metadata")
	}

	if err := rcv.checkAppVersion(md); err != nil {
		return nil, err
	}

	return handler(ctx, req)
}

func (rcv *checkAppVersion) StreamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if !rcv.decider(info.FullMethod) {
		return handler(srv, ss)
	}

	md, ok := metadata.FromIncomingContext(ss.Context())
	if !ok {
		return status.Error(codes.Internal, "cannot check gRPC metadata")
	}

	if err := rcv.checkAppVersion(md); err != nil {
		return err
	}

	return handler(srv, ss)
}

func (rcv *checkAppVersion) checkAppVersion(md metadata.MD) error {
	pkgs := md.Get(packageNameKey)
	if len(pkgs) == 0 || len(pkgs[0]) == 0 {
		return status.Error(codes.InvalidArgument, "missing package name")
	}

	requiredVersion, ok := rcv.versions[pkgs[0]]
	if !ok {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("invalid package name: %s", pkgs[0]))
	}

	vers := md.Get(versionKey)
	if len(vers) == 0 || len(vers[0]) == 0 {
		return status.Error(codes.InvalidArgument, "missing client version")
	}

	appVersion, err := versionStringToInt(vers[0])
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	if checkForceUpdate(requiredVersion, appVersion) {
		return status.Error(codes.Aborted, "force update")
	}

	return nil
}

func versionStringToInt(version string) (uint, error) {
	versionArr := strings.Split(version, ".")
	if len(versionArr) < 3 {
		return 0, errors.New("invalid client version")
	}

	major, _ := strconv.Atoi(versionArr[0])
	minor, _ := strconv.Atoi(versionArr[1])
	patch, _ := strconv.Atoi(versionArr[2])

	appVersion := uint(major*10000 + minor*100 + patch)
	return appVersion, nil
}

func parseVersionConfig(cfgClientVersion string) (map[string]string, error) {
	clientVersions := make(map[string]string)

	versions := strings.Split(cfgClientVersion, ",")
	for _, ver := range versions {
		parts := strings.Split(ver, ":")
		if len(parts) != 2 {
			return nil, errors.New("invalid version, must match pattern <pkg_name>:<required_version>")
		}

		clientVersions[parts[0]] = parts[1]
	}

	if len(clientVersions) == 0 {
		return nil, errors.New("no client version given")
	}

	return clientVersions, nil
}

func getAppVersion(cfgClientVersion string) (map[string]uint, error) {
	clientVersions := make(map[string]uint)

	versions := strings.Split(cfgClientVersion, ",")
	for _, ver := range versions {
		parts := strings.Split(ver, ":")
		if len(parts) != 2 {
			return nil, errors.New("invalid version, must match pattern <pkg_name>:<required_version>")
		}

		requiredVersion, err := versionStringToInt(parts[1])
		if err != nil {
			return nil, errors.Wrap(err, "invalid required version")
		}

		clientVersions[parts[0]] = requiredVersion
	}

	if len(clientVersions) == 0 {
		return nil, errors.New("no client version given")
	}

	return clientVersions, nil
}

func compareStringNumber(a, b string) int {
	if len(a) == len(b) {
		return strings.Compare(a, b)
	}
	if len(a) < len(b) {
		return -1
	}
	return 1
}

func checkForceUpdate(requireVersion, appVersion uint) bool {
	requireVersionMajor := (requireVersion % 10000) / 100
	appVersionMajor := (appVersion % 10000) / 100

	return appVersionMajor < requireVersionMajor
}

func compareVersion(requiredVersion, appVersion string) error {
	rvs := strings.Split(requiredVersion, ".")
	apvs := strings.Split(appVersion, ".")
	if len(rvs) != 3 || len(apvs) != 3 {
		return fmt.Errorf("invalid version, must match pattern <major>.<minor>.<patch>")
	}

	requireMn := semver.MajorMinor(fmt.Sprintf("v%s", requiredVersion))
	appMn := semver.MajorMinor(fmt.Sprintf("v%s", appVersion))
	compareResult := semver.Compare(requireMn, appMn)

	if compareResult == 1 {
		return status.Error(codes.Aborted, "force update")
	} else if compareResult == 0 {
		if compareStringNumber(rvs[2], apvs[2]) == 1 {
			return status.Error(codes.Aborted, "force update")
		}
	}
	return nil
}

func CheckForceUpdateApp(ctx context.Context, cfgClientVersion string) error {
	versions, err := parseVersionConfig(cfgClientVersion)
	if err != nil {
		return errors.Wrap(err, "getAppVersion")
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Internal, "cannot check gRPC metadata")
	}

	pkgs := md.Get(packageNameKey)
	if len(pkgs) == 0 || len(pkgs[0]) == 0 {
		return status.Error(codes.InvalidArgument, "missing package name")
	}

	requiredVersion, ok := versions[pkgs[0]]
	if !ok {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("invalid package name: %s", pkgs[0]))
	}

	vers := md.Get(versionKey)
	if len(vers) == 0 || len(vers[0]) == 0 {
		return status.Error(codes.InvalidArgument, "missing client version")
	}

	return compareVersion(requiredVersion, vers[0])
}
