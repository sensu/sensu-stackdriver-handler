package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	"github.com/sensu/sensu-go/types"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

type HandlerConfig struct {
	sensu.PluginConfig
	ProjectID     string
	IncludeLabels bool
}

var (
	handlerConfig = HandlerConfig{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-stackdriver-handler",
			Short:    "Send Sensu Go collected metrics to Google Stackdriver",
			Timeout:  10,
			Keyspace: "sensu.io/plugins/stackdriver-handler/config",
		},
	}

	handlerConfigOptions = []*sensu.PluginConfigOption{
		{
			Path:      "project-id",
			Env:       "STACKDRIVER_PROJECTID",
			Argument:  "project-id",
			Shorthand: "p",
			Default:   "",
			Usage:     "The Google Cloud Project ID",
			Value:     &handlerConfig.ProjectID,
		},
		{
			Path:      "include-labels",
			Env:       "",
			Argument:  "include-labels",
			Shorthand: "l",
			Default:   false,
			Usage:     "Include entity and check labels in the metrics labels (default false)",
			Value:     &handlerConfig.IncludeLabels,
		},
	}
)

func main() {
	handler := sensu.NewGoHandler(&handlerConfig.PluginConfig, handlerConfigOptions, checkArgs, executeHandler)
	handler.Execute()
}

func checkArgs(_ *types.Event) error {
	if len(handlerConfig.ProjectID) == 0 {
		return fmt.Errorf("--project-id or STACKDRIVER_PROJECTID environment variable is required")
	}
	return nil
}

func chunkTimeSeries(timeSeries []*monitoringpb.TimeSeries) [][]*monitoringpb.TimeSeries {
	var c [][]*monitoringpb.TimeSeries

	for i := 0; i < len(timeSeries); i += 200 {
		e := i + 200
		l := len(timeSeries)
		if e > l {
			e = l
		}

		c = append(c, timeSeries[i:e])
	}

	return c
}

func writeTimeSeries(projectID string, timeSeries []*monitoringpb.TimeSeries) error {
	ctx := context.Background()
	c, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return err
	}

	for _, ts := range chunkTimeSeries(timeSeries) {
		req := &monitoringpb.CreateTimeSeriesRequest{
			Name:       "projects/" + projectID,
			TimeSeries: ts,
		}
		log.Printf("writeTimeseriesRequest: %+v\n", req)

		err = c.CreateTimeSeries(ctx, req)
		if err != nil {
			return fmt.Errorf("could not write time series, %v ", err)
		}
	}

	return nil
}

func createTimeSeries(event *types.Event) []*monitoringpb.TimeSeries {
	timeSeries := []*monitoringpb.TimeSeries{}

	replacer := strings.NewReplacer("/", "_", "-", "_", ".", "_")

	for _, p := range event.Metrics.Points {
		l := make(map[string]string)

		if handlerConfig.IncludeLabels {
			if event.Entity.Labels != nil {
				for k, v := range event.Entity.Labels {
					l[replacer.Replace(k)] = v
				}
			}
		}

		if event.HasCheck() {
			if handlerConfig.IncludeLabels {
				if event.Check.Labels != nil {
					for k, v := range event.Check.Labels {
						l[replacer.Replace(k)] = v
					}
				}
			}
		}

		for _, t := range p.Tags {
			l[replacer.Replace(t.Name)] = t.Value
		}

		ts := &timestamp.Timestamp{
			Seconds: p.Timestamp,
		}

		s := &monitoringpb.TimeSeries{
			Metric: &metricpb.Metric{
				Type:   "custom.googleapis.com/sensu/" + p.Name,
				Labels: l,
			},
			Points: []*monitoringpb.Point{{
				Interval: &monitoringpb.TimeInterval{
					StartTime: ts,
					EndTime:   ts,
				},
				Value: &monitoringpb.TypedValue{
					Value: &monitoringpb.TypedValue_DoubleValue{
						DoubleValue: p.Value,
					},
				},
			}},
		}

		timeSeries = append(timeSeries, s)
	}

	return timeSeries
}

func executeHandler(event *types.Event) error {
	log.Println("executing handler with --project-id", handlerConfig.ProjectID)

	if event.HasMetrics() {
		timeSeries := createTimeSeries(event)
		err := writeTimeSeries(handlerConfig.ProjectID, timeSeries)

		if err != nil {
			return err
		}
	}

	return nil
}
