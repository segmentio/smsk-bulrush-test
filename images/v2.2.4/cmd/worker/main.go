/*
Description: Simple bulrush worker
- bulrush worker that consumes from a topic in a kafka and send messages to a topic in a different kafka
- This worker does not collect any datadog metrics
- No profiling is enabled

Usage: (local Desktop run)
	$ go run cmd/worker/main.go -h
	$ go run cmd/worker/main.go \
	  -debug=true \
	  -source.connect=localhost:9092 \
	  -source.group-id=bulrush-test-cg-1 \
	  -source.topic=bulrush-test-1 \
	  -sink.connect=localhost:9092 \
	  -output-topic=bulrush-test-2 \
	  -address=:3000
*/

package main

import (
	"context"
	"github.com/pkg/errors"
	"github.com/segmentio/bulrush/v2"
	"github.com/segmentio/bulrush/v2/transport/kafka"
	"github.com/segmentio/bulrush/v2/workload/streaming"
	"github.com/segmentio/conf"
	"github.com/segmentio/events/v2"
	_ "github.com/segmentio/events/v2/ecslogs"
	_ "github.com/segmentio/events/v2/log"
	_ "github.com/segmentio/events/v2/sigevents"
	_ "github.com/segmentio/events/v2/text"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var config = struct {
	Debug       bool               `conf:"debug" help:"enables debug logging"`
	Source      kafka.SourceConfig `conf:"source"`
	Sink        kafka.SinkConfig   `conf:"sink"`
	Address     string             `conf:"address" help:"address on which the server should listen"`
	OutputTopic string             `conf:"output-topic" help:"output topic where messages will be sent" validate:"min=1"`
}{}

var (
	// Version represents the version of the program at runtime. The default value
	// is set as a placeholder. Our build script will inject a real value during
	// compilation based on the Git tags.
	Version = "x.x.x"
)

func init() {
	config.Debug = true

	// Consumer config
	config.Source = kafka.DefaultSourceConfig
	config.Source.Connect = "localhost:9092"
	config.Source.GroupId = "bulrush-test-cg-1"
	config.Source.Topic = "bulrush-test-1"

	// Producer config
	config.Sink = kafka.DefaultSinkConfig
	config.Sink.Connect = "localhost:9092"
	config.OutputTopic = "bulrush-test-2"

	config.Address = ":3000"
}

func main() { // start main
	conf.LoadWith(&config, conf.Loader{
		Args: os.Args[1:],
		Sources: []conf.Source{
			conf.NewEnvSource("", os.Environ()...),
		},
	})

	events.DefaultLogger.EnableDebug = config.Debug
	events.Log("Starting service", events.Args{
		{Name: "version", Value: Version},
		{Name: "config.debug", Value: config.Debug},
		{Name: "config.Source.Connect", Value: config.Source.Connect},
		{Name: "config.Sink.Connect", Value: config.Sink.Connect},
		{Name: "config.Address", Value: config.Address},
	})

	// Kafka Consumer config
	source, err := kafka.NewSource(config.Source)
	if err != nil {
		panic(err)
	}
	defer source.Close()

	// Kafka Producer config
	sink, err := kafka.NewSink(config.Sink)
	if err != nil {
		panic(err)
	}
	defer sink.Close()

	worker, err := streaming.New(streaming.Config{
		Source: source,
		Sink:   sink,
		Handler: func(ctx context.Context, msg bulrush.Message) ([]bulrush.Message, error) {
			// route the message to somewhere else.  this function could do all
			// sorts of things such as fan-out, filter, transform, repartition by
			// assigning a Key, etc.
			// Print the consumed message
			events.Debug("bulrush.Message: %s", msg)
			return []bulrush.Message{{
				Topic: config.OutputTopic,
				Value: msg.Value,
			}}, nil
		},
	})
	if err != nil {
		panic(err)
	}

	// start the worker along with capturing errors
	errCh := make(chan error)
	go func() {
		errCh <- worker.Run()
	}()

	// Health check handler
	go func() {
		http.HandleFunc("/internal/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		})
		errCh <- errors.Wrapf(http.ListenAndServe(config.Address, nil), "healthcheck handler")
	}()

	// Handle signals
	sigchan := make(chan os.Signal)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-sigchan:
		events.Log("stopping in response to signal %{signal}s.", sig)
	case err := <-errCh:
		events.Log("stopping in response to error %+{error}v.", err)
	}

	if err := worker.Close(); err != nil {
		events.Log("error shutting down worker: %+{error}v", err)
	}
} // end main
