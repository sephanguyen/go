package grpc

import (
	"fmt"
	"os"
	"strings"

	"github.com/yoheimuta/go-protoparser"
)

func ServicesFromProtoFile(protoFiles []string) (Services, error) {
	res := make(Services)
	for _, protoFile := range protoFiles {
		f, err := os.Open(protoFile)
		if err != nil {
			return nil, fmt.Errorf("could not open proto file: %w", err)
		}

		got, err := protoparser.Parse(f)
		if err != nil {
			return nil, fmt.Errorf("failed to parse protofile %s: %w", protoFile, err)
		}
		_ = f.Close()
		srv := gRPCServicesFromProto(got)
		for k := range srv {
			res[k] = srv[k]
		}
	}

	return res, nil
}

func ServicesFromFullMethods(methods []string) (Services, error) {
	res := make(Services)
	for _, method := range methods {
		s1 := strings.Split(method, "/")
		if len(s1) != 2 {
			return nil, fmt.Errorf("%s is not full path of a method", method)
		}
		if len(s1[1]) == 0 {
			return nil, fmt.Errorf("%s missing method name", method)
		}

		// split package name and service name
		s2 := strings.Split(s1[0], ".")
		if len(s2) < 2 {
			return nil, fmt.Errorf("%s is not full path of a method", method)
		}
		pkg := strings.Join(s2[:len(s2)-1], ".")
		serviceName := s2[len(s2)-1]
		if len(pkg) == 0 {
			return nil, fmt.Errorf("%s missing package name", method)
		}
		if len(serviceName) == 0 {
			return nil, fmt.Errorf("%s missing serivce name", method)
		}
		res.AddMethodByService(pkg, serviceName, s1[1])
	}

	return res, nil
}

type Service struct {
	ServiceName    string // not contain package name
	MethodNamesMap map[string]bool
}

func (g *Service) AddMethodName(name string) {
	if g.MethodNamesMap == nil {
		g.MethodNamesMap = make(map[string]bool)
	}
	g.MethodNamesMap[name] = true
}

func (g *Service) MethodNames() []string {
	res := make([]string, 0, len(g.MethodNamesMap))
	for name := range g.MethodNamesMap {
		res = append(res, name)
	}
	return res
}

type Services map[string]*Service

func (g Services) RemoveByFullMethodNames(names []string) error {
	if len(names) == 0 {
		return nil
	}
	deletedList, err := ServicesFromFullMethods(names)
	if err != nil {
		return err
	}

	for pkg, deleted := range deletedList {
		if _, ok := g[pkg]; ok && g[pkg].MethodNamesMap != nil {
			for methodName := range deleted.MethodNamesMap {
				if _, ok = g[pkg].MethodNamesMap[methodName]; ok {
					delete(g[pkg].MethodNamesMap, methodName)
				}
			}
		}
	}

	return nil
}

func (g Services) AddMethodByService(pkg, serviceName, methodName string) {
	fullName := fmt.Sprintf("%s.%s", pkg, serviceName)
	if _, ok := g[fullName]; !ok {
		g[fullName] = &Service{
			ServiceName: serviceName,
		}
	}

	g[fullName].AddMethodName(methodName)
}

func (g Services) GRPCMethods() []string {
	res := make([]string, 0)
	for srvName, v := range g {
		for _, methodNames := range v.MethodNames() {
			res = append(res, fmt.Sprintf("%s/%s", srvName, methodNames))
		}
	}

	return res
}
