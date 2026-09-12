//revive:disable:package-comments
package service

import (
	"fmt"

	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip" // Experimental gzip initialization
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"git.sonicoriginal.software/grpc-foundation/methods"
	foundation "git.sonicoriginal.software/grpc-foundation/server"
	diagpb "git.sonicoriginal.software/grpc-protos/diagnostics"
	infopb "git.sonicoriginal.software/grpc-protos/info"

	"git.sonicoriginal.software/grpc-service/diagnostics"
)

// Register attaches the caller's services to srv, along with the diagnostics,
// info, health, and reflection services every service exposes. It marks each
// of the caller's services SERVING and returns their fully qualified method
// names, which is what the server exposes beyond that infrastructure.
//
// checks are the dependencies the diagnostics service reports on. Every
// dependency is the caller's to name; nil means none.
//
// Register only assembles. Serving, shutdown, and any later health status
// change are the caller's to make.
func Register(
	srv Server,
	healthSrv Health,
	checks diagnostics.Checks,
	registerFn func(grpc.ServiceRegistrar),
) ([]string, error) {
	if srv == nil {
		return nil, fmt.Errorf("srv cannot be nil")
	}

	if healthSrv == nil {
		return nil, fmt.Errorf("healthSrv cannot be nil")
	}

	if registerFn != nil {
		registerFn(srv)
	}

	diagpb.RegisterDiagnosticsServiceServer(srv, diagnostics.NewServer(checks))
	infopb.RegisterInfoServiceServer(srv, &infoServer{version: foundation.Version()})
	grpc_health_v1.RegisterHealthServer(srv, healthSrv)
	reflection.Register(srv)

	methodNames := methods.Extract(srv, methods.NewPatternFilter(nil, infrastructurePrefixes))

	for _, serviceName := range methods.ServiceNames(methodNames) {
		healthSrv.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_SERVING)
	}

	return methodNames, nil
}
