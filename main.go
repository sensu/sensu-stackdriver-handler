package main

import (
	"context"
	"fmt"
	"log"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	"github.com/sensu/sensu-go/types"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

type HandlerConfig struct {
	sensu.PluginConfig
	ProjectID string
}

type ConfigOptions struct {
	ProjectID sensu.PluginConfigOption
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

	handlerConfigOptions = ConfigOptions{
		ProjectID: sensu.PluginConfigOption{
			Path:      "project-id",
			Env:       "STACKDRIVER_PROJECTID",
			Argument:  "project-id",
			Shorthand: "p",
			Default:   "",
			Usage:     "The Google Cloud Project ID",
			Value:     &handlerConfig.ProjectID,
		},
	}

	options = []*sensu.PluginConfigOption{
		&handlerConfigOptions.ProjectID,
	}
)

func main() {
	handler := sensu.NewGoHandler(&handlerConfig.PluginConfig, options, checkArgs, executeHandler)
	handler.Execute()
}

func checkArgs(_ *types.Event) error {
	if len(handlerConfig.ProjectID) == 0 {
		return fmt.Errorf("--project-id or STACKDRIVER_PROJECTID environment variable is required")
	}
	return nil
}

func writeTimeSeries(projectID string, timeSeries []*monitoringpb.TimeSeries) error {
	ctx := context.Background()
	c, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return err
	}
	req := &monitoringpb.CreateTimeSeriesRequest{
		Name:       "projects/" + projectID,
		TimeSeries: timeSeries,
	}
	log.Printf("writeTimeseriesRequest: %+v\n", req)

	err = c.CreateTimeSeries(ctx, req)
	if err != nil {
		return fmt.Errorf("could not write time series, %v ", err)
	}
	return nil
}

func createTimeSeries(event *types.Event) []*monitoringpb.TimeSeries {
	timeSeries := []*monitoringpb.TimeSeries{}

	for _, p := range event.Metrics.Points {
		if len(timeSeries) == 200 {
			log.Print("reached maximum number of time series per request (200)")
			// TODO: Support multi-request.
			break
		}

		l := make(map[string]string)
		for _, t := range p.Tags {
			l[t.Name] = t.Value
		}
		l["sensu_entity_name"] = event.Entity.Name

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

	timeSeries := createTimeSeries(event)
	err := writeTimeSeries(handlerConfig.ProjectID, timeSeries)

	return err
}
