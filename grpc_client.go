package main

import (
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/descriptorpb"
	"log"
)

const (
	configPackage   = protogen.GoImportPath("github.com/shengyjs/utility/config")
	logPackage      = protogen.GoImportPath("log")
	errorsPackage   = protogen.GoImportPath("errors")
	insecurePackage = protogen.GoImportPath("google.golang.org/grpc/credentials/insecure")
	fmtPackage      = protogen.GoImportPath("fmt")
	timePackage     = protogen.GoImportPath("time")
)

// 生成一个包含有client功能封装的 _grpc_client.pb.go文件
func generateClientFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	if len(file.Services) == 0 {
		return nil
	}

	filename := file.GeneratedFilenamePrefix + "_grpc_client.pb.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	g.P("// Code generated by protoc-gen-go-grpc. DO NOT EDIT.")
	g.P("// versions:")
	g.P("// - protoc-gen-go-grpc v", version)
	g.P("// - protoc             ", protocVersion(gen))
	if file.Proto.GetOptions().GetDeprecated() {
		g.P("// ", file.Desc.Path(), " is a deprecated file.")
	} else {
		g.P("// source: ", file.Desc.Path())
	}
	g.P()
	g.P("package ", file.GoPackageName)
	g.P()
	generateClientFileContent(gen, file, g)
	return g
}

// generateClientFileContent generates the gRPC service definitions, excluding the package statement.
func generateClientFileContent(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile) {
	if len(file.Services) == 0 {
		return
	}

	g.P("// This is a compile-time assertion to ensure that this generated file")
	g.P("// is compatible with the grpc package it is being compiled against.")
	g.P("// Requires gRPC-Go v1.32.0 or later.")
	g.P("const _ = ", grpcPackage.Ident("SupportPackageIsVersion7")) // When changing, update version number above.
	g.P()

	// 定义配置选项的结构
	g.P("// 配置文件结构")
	g.P("type ", "serviceConfig ", "struct {")
	g.P("IP string `ini:\"ip\"`")
	g.P("Port string `ini:\"port\"`")
	g.P("}")

	// 处理每一个service
	for _, service := range file.Services {
		genClientStub(gen, file, g, service)
	}
}

func genClientStub(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, service *protogen.Service) {
	clientName := service.GoName + "ServiceClient"

	g.P("// ", clientName, " is the client API for ", service.GoName, " service.")
	g.P("//")
	g.P("// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.")
	g.P()

	// 客户端结构
	if service.Desc.Options().(*descriptorpb.ServiceOptions).GetDeprecated() {
		g.P("//")
		g.P(deprecationComment)
	}
	g.Annotate(clientName, service.Location)
	g.P("// 客户端结构")
	g.P("type ", clientName, " struct {")
	g.P("conf *serviceConfig")
	g.P("conn *grpc.ClientConn")
	g.P("client ", service.GoName, "Client")
	g.P("}")
	g.P()

	// 生成初始化Init函数
	g.P("func (cc *", clientName, ") Init(configFile string, configSectionName string) error {")
	g.P("// 设置配置信息")
	g.P("if cc.conf == nil {")
	g.P("// 读取配置信息")
	g.P("conf := new(serviceConfig)")
	g.P("errCode := ", configPackage.Ident("GetConfigItem"), "(configFile, configSectionName, conf)")
	g.P("if errCode != config.ERR_OK {")
	g.P(logPackage.Ident("Printf"), "(\"GetConfigItem Fail. errCode = %v\", errCode)")
	g.P("return ", errorsPackage.Ident("New"), "(\"GetConfigItem Fail.\")")
	g.P("}")
	g.P(logPackage.Ident("Printf"), "(\"conf is %v\", conf)")
	g.P("cc.conf = conf")
	g.P("} else {")
	g.P(logPackage.Ident("Printf"), "(\"conf is exit. conf = %v \", cc.conf)")
	g.P("}")
	g.P()
	g.P("// 建立连接")
	g.P("if cc.conn == nil {")
	g.P("var opts[] ", grpcPackage.Ident("DialOption"))
	g.P("opts = append(opts, ", grpcPackage.Ident("WithTransportCredentials"), "(", insecurePackage.Ident("NewCredentials"), "()", "))")
	g.P("conn, err := ", grpcPackage.Ident("Dial"), "(", fmtPackage.Ident("Sprintf"), "(\"%s:%s\", cc.conf.IP, cc.conf.Port), opts...)")
	g.P("if err != nil {")
	g.P(logPackage.Ident("Printf"), "(\"Dail Fail. err = %v\", err)")
	g.P("return err")
	g.P("}")
	g.P(logPackage.Ident("Printf"), "(\"conn is %v\", conn)")
	g.P("cc.conn = conn")
	g.P("} else {")
	g.P(logPackage.Ident("Printf"), "(\"conn is exit. conn = %v\", cc.conn)")
	g.P("}")
	g.P()
	g.P("// 创建客户端 ")
	g.P("cc.client = New", service.GoName, "Client(cc.conn)")
	g.P("return nil")
	g.P("}")
	g.P()

	// 关闭连接
	g.P("//关闭连接")
	g.P("func (cc *", clientName, ") Close() error{")
	g.P(logPackage.Ident("Printf"), "(\"Close. conf = %v, conn = %v\", cc.conf, cc.conn)")
	g.P("// 连接存在的话，就关闭")
	g.P("if cc.conn != nil {")
	g.P("err := cc.conn.Close()")
	g.P("if err != nil {")
	g.P(logPackage.Ident("Printf"), "(\"conn.Close Fail. err = %v\", err)")
	g.P("return err")
	g.P("}")
	g.P("} else {")
	g.P(logPackage.Ident("Printf"), "(\"Do not have conn\")")
	g.P("}")
	g.P()
	g.P("cc.conf = nil")
	g.P("cc.conn = nil")
	g.P("cc.client = nil")
	g.P("return nil")
	g.P("}")
	g.P()

	// 生成接口函数
	var methodIndex, streamIndex int = 0, 0
	for _, method := range service.Methods {
		if !method.Desc.IsStreamingServer() && !method.Desc.IsStreamingClient() {
			// 普通RPC函数
			genClientStubMethod(gen, file, g, method, methodIndex)
			methodIndex++
		} else {
			// TODO: 暂不能生成stream方法
			log.Printf("TODO:Can not generate stream method yet!")
			streamIndex++
		}
	}
}

func genClientStubMethod(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, method *protogen.Method, index int) {
	service := method.Parent
	clientName := service.GoName + "ServiceClient"
	g.Annotate(clientName+"."+method.GoName, method.Location)

	if method.Desc.Options().(*descriptorpb.MethodOptions).GetDeprecated() {
		g.P(deprecationComment)
	}

	// TODO: 暂时还不能生成stream方法
	if method.Desc.IsStreamingServer() || method.Desc.IsStreamingClient() {
		log.Printf("TODO:Can not generate stream method yet!")
		return
	}

	g.P(method.Comments.Leading, "func (cc *", clientName, ") ", method.GoName, "(req *", g.QualifiedGoIdent(method.Input.GoIdent), ") (*", g.QualifiedGoIdent(method.Output.GoIdent), ", error) {")
	g.P(logPackage.Ident("Printf"), "(\"", method.GoName, ". req = %v\", req)")
	g.P()
	g.P("// 创建context")
	g.P("ctx, cancel := ", contextPackage.Ident("WithTimeout"), "(", contextPackage.Ident("Background"), "(), 10*", timePackage.Ident("Second"), ")")
	g.P("defer cancel()")
	g.P()
	g.P("// 调用")
	g.P("resp, err := cc.client.", method.GoName, "(ctx, req)")
	g.P(logPackage.Ident("Printf"), "(\"resp = %v\", resp)")
	g.P("if err != nil {")
	g.P(logPackage.Ident("Printf"), "(\"client.", method.GoName, " Fail. req = %v, err = %v\", req, err)")
	g.P("return resp, err")
	g.P("}")
	g.P("return resp, nil")
	g.P("}")
}
