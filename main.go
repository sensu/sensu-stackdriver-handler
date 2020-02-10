package main

import (
	"context"
	"fmt"
	"log"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	"github.com/sensu/sensu-go/types"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoredres "google.golang.org/genproto/googleapis/api/monitoredres"
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
			Keyspace: "sensu.io/plugins/sensu-stackdriver-handler/config",
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

func writeTimeSeriesValue(projectID, metricType string, value int) error {
	ctx := context.Background()
	c, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return err
	}
	now := &timestamp.Timestamp{
		Seconds: time.Now().Unix(),
	}
	req := &monitoringpb.CreateTimeSeriesRequest{
		Name: "projects/" + handlerConfig.ProjectID,
		TimeSeries: []*monitoringpb.TimeSeries{{
			Metric: &metricpb.Metric{
				Type: metricType,
			},
			Resource: &monitoredres.MonitoredResource{
				Type: "gce_instance",
				Labels: map[string]string{
					"project_id": handlerConfig.ProjectID,
				},
			},
			Points: []*monitoringpb.Point{{
				Interval: &monitoringpb.TimeInterval{
					StartTime: now,
					EndTime:   now,
				},
				Value: &monitoringpb.TypedValue{
					Value: &monitoringpb.TypedValue_Int64Value{
						Int64Value: int64(value),
					},
				},
			}},
		}},
	}
	log.Printf("writeTimeseriesRequest: %+v\n", req)

	err = c.CreateTimeSeries(ctx, req)
	if err != nil {
		return fmt.Errorf("could not write time series value, %v ", err)
	}
	return nil
}

func executeHandler(event *types.Event) error {
	log.Println("executing handler with --project-id", handlerConfig.ProjectID)
	return nil
}
