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

	for _, p := range event.Metrics.Points {
		l := make(map[string]string)

		if event.Entity.Labels != nil {
			for k, v := range event.Entity.Labels {
				l[k] = v
			}
		}
		l["sensu_entity_name"] = event.Entity.Name

		if event.HasCheck() {
			if event.Check.Labels != nil {
				for k, v := range event.Check.Labels {
					l[k] = v
				}
			}
			l["sensu_check_name"] = event.Check.Name
		}

		for _, t := range p.Tags {
			l[t.Name] = t.Value
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
