//revive:disable:package-comments
package service

// infrastructurePrefixes name the services Register excludes from the returned
// method list and never reports a health status for. They are up exactly when
// the process is, which the health service's own "" entry already reports.
var infrastructurePrefixes = []string{
	"grpc.",        // gRPC infrastructure (health, reflection)
	"info.",        // info endpoint
	"diagnostics.", // diagnostics endpoint
}
