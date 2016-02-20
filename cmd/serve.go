package cmd

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/gengo/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/philips/grpc-gateway-example/echopb"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		serve()
	},
}

func init() {
	RootCmd.AddCommand(serveCmd)
}

type myService struct{}

func (m *myService) Echo(c context.Context, s *pb.EchoMessage) (*pb.EchoMessage, error) {
	fmt.Printf("rpc request Echo(%q)\n", s.Value)
	return s, nil
}

func newServer() *myService {
	return new(myService)
}

// grpcHandlerFunc returns an http.Handler that delegates to grpcServer on incoming gRPC
// connections or otherHandler otherwise. Copied from cockroachdb.
func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO(tamird): point to merged gRPC code rather than a PR.
		// This is a partial recreation of gRPC's internal checks https://github.com/grpc/grpc-go/pull/514/files#diff-95e9a25b738459a2d3030e1e6fa2a718R61
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}

func serve() {
	opts := []grpc.ServerOption{
		grpc.Creds(credentials.NewClientTLSFromCert(demoCertPool, "localhost:10000"))}

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterEchoServiceServer(grpcServer, newServer())
	ctx := context.Background()

	dcreds := credentials.NewTLS(&tls.Config{
		ServerName: demoAddr,
		RootCAs:    demoCertPool,
	})
	dopts := []grpc.DialOption{grpc.WithTransportCredentials(dcreds)}

	mux := http.NewServeMux()
	mux.HandleFunc("/swagger.json", func(w http.ResponseWriter, req *http.Request) {
		io.Copy(w, strings.NewReader(pb.Swagger))
	})

	gwmux := runtime.NewServeMux()
	err := pb.RegisterEchoServiceHandlerFromEndpoint(ctx, gwmux, demoAddr, dopts)
	if err != nil {
		fmt.Printf("serve: %v\n", err)
		return
	}

	mux.Handle("/", gwmux)

	conn, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}

	srv := &http.Server{
		Addr:    demoAddr,
		Handler: grpcHandlerFunc(grpcServer, mux),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{*demoKeyPair},
		},
	}

	http2.ConfigureServer(srv, nil)

	fmt.Printf("grpc on port: %d\n", port)
	err = srv.Serve(tls.NewListener(conn, srv.TLSConfig))

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	return
}
