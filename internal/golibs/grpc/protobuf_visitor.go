package grpc

import (
	"github.com/yoheimuta/go-protoparser/parser"
)

// ProtoBufVisitor implement parser.Visitor but just handle VisitPackage, VisitService and VisitRPC currently
type ProtoBufVisitor struct {
	VisitPackageFunc func(*parser.Package) (next bool)
	VisitRPCFunc     func(*parser.RPC) (next bool)
	VisitServiceFunc func(*parser.Service) (next bool)
}

func (p *ProtoBufVisitor) VisitComment(*parser.Comment) {}
func (p *ProtoBufVisitor) VisitEmptyStatement(*parser.EmptyStatement) (next bool) {
	return true
}
func (p *ProtoBufVisitor) VisitEnum(*parser.Enum) (next bool) {
	return true
}
func (p *ProtoBufVisitor) VisitEnumField(*parser.EnumField) (next bool) {
	return true
}
func (p *ProtoBufVisitor) VisitExtend(*parser.Extend) (next bool) {
	return true
}
func (p *ProtoBufVisitor) VisitExtensions(*parser.Extensions) (next bool) {
	return true
}
func (p *ProtoBufVisitor) VisitField(*parser.Field) (next bool) {
	return true
}
func (p *ProtoBufVisitor) VisitGroupField(*parser.GroupField) (next bool) {
	return true
}
func (p *ProtoBufVisitor) VisitImport(*parser.Import) (next bool) {
	return true
}
func (p *ProtoBufVisitor) VisitMapField(*parser.MapField) (next bool) {
	return true
}
func (p *ProtoBufVisitor) VisitMessage(*parser.Message) (next bool) {
	return true
}
func (p *ProtoBufVisitor) VisitOneof(*parser.Oneof) (next bool) {
	return true
}
func (p *ProtoBufVisitor) VisitOneofField(*parser.OneofField) (next bool) {
	return true
}
func (p *ProtoBufVisitor) VisitOption(*parser.Option) (next bool) {
	return true
}
func (p *ProtoBufVisitor) VisitPackage(pkg *parser.Package) (next bool) {
	if p.VisitPackageFunc == nil {
		return true
	}
	return p.VisitPackageFunc(pkg)
}
func (p *ProtoBufVisitor) VisitReserved(*parser.Reserved) (next bool) {
	return true
}
func (p *ProtoBufVisitor) VisitRPC(rpc *parser.RPC) (next bool) {
	if p.VisitRPCFunc == nil {
		return true
	}
	return p.VisitRPCFunc(rpc)
}
func (p *ProtoBufVisitor) VisitService(s *parser.Service) (next bool) {
	if p.VisitServiceFunc == nil {
		return true
	}
	return p.VisitServiceFunc(s)
}
func (p *ProtoBufVisitor) VisitSyntax(*parser.Syntax) (next bool) {
	return true
}

func getPackageName(proto *parser.Proto) string {
	var pkgName string
	for _, v := range proto.ProtoBody {
		v.Accept(&ProtoBufVisitor{
			VisitPackageFunc: func(p *parser.Package) (next bool) {
				pkgName = p.Name
				return false
			},
		})
	}

	return pkgName
}

func gRPCServicesFromProto(proto *parser.Proto) Services {
	pkgName := getPackageName(proto)
	res := make(Services)
	for _, v := range proto.ProtoBody {
		var name string
		v.Accept(&ProtoBufVisitor{
			VisitServiceFunc: func(service *parser.Service) (next bool) {
				name = service.ServiceName
				return true
			},
			VisitRPCFunc: func(rpc *parser.RPC) (next bool) {
				res.AddMethodByService(pkgName, name, rpc.RPCName)
				return false
			},
		})
	}

	return res
}
