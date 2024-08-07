package telemetry

import (
	"github.com/honeycombio/otel-config-go/otelconfig"
	"go.opentelemetry.io/contrib/processors/baggage/baggagetrace"
)

const TracerName = "triple-j-bot-tracer"

func InitTelemetry() (func(), error) {
	// Enable multi-span attributes
	bsp := baggagetrace.New()
	// Use the Honeycomb distro to set up the OpenTelemetry SDK
	otelShutdown, err := otelconfig.ConfigureOpenTelemetry(
		otelconfig.WithSpanProcessor(bsp),
	)

	return otelShutdown, err
}
