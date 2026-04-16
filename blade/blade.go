package blade

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/TylerBrock/colorjson"
	"github.com/TylerBrock/saw/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/fatih/color"
)

// A Blade is a Saw execution instance
type Blade struct {
	config *config.Configuration
	aws    *config.AWSConfiguration
	output *config.OutputConfiguration
	cwl    *cloudwatchlogs.Client
}

// NewBlade creates a new Blade with CloudWatchLogs instance from provided config
func NewBlade(
	config *config.Configuration,
	awsConfig *config.AWSConfiguration,
	outputConfig *config.OutputConfiguration,
) *Blade {
	blade := Blade{}
	ctx := context.Background()

	// Build config options
	configOpts := []func(*awsconfig.LoadOptions) error{}

	if awsConfig.Region != "" {
		configOpts = append(configOpts, awsconfig.WithRegion(awsConfig.Region))
	}

	if awsConfig.Profile != "" {
		configOpts = append(configOpts, awsconfig.WithSharedConfigProfile(awsConfig.Profile))
	}

	// Load AWS config
	cfg, err := awsconfig.LoadDefaultConfig(ctx, configOpts...)
	if err != nil {
		fmt.Printf("unable to load SDK config, %v\n", err)
		os.Exit(1)
	}

	// Override endpoint if specified
	if awsConfig.Endpoint != "" {
		cfg.BaseEndpoint = aws.String(awsConfig.Endpoint)
	}

	blade.cwl = cloudwatchlogs.NewFromConfig(cfg)
	blade.config = config
	blade.output = outputConfig

	return &blade
}

// GetLogGroups gets the log groups from AWS given the blade configuration
func (b *Blade) GetLogGroups() []types.LogGroup {
	ctx := context.Background()
	input := b.config.DescribeLogGroupsInput()
	groups := make([]types.LogGroup, 0)

	paginator := cloudwatchlogs.NewDescribeLogGroupsPaginator(b.cwl, input)
	for paginator.HasMorePages() {
		out, err := paginator.NextPage(ctx)
		if err != nil {
			fmt.Printf("Error getting log groups: %v\n", err)
			os.Exit(2)
		}
		groups = append(groups, out.LogGroups...)
	}

	return groups
}

// GetLogStreams gets the log streams from AWS given the blade configuration
func (b *Blade) GetLogStreams() []types.LogStream {
	ctx := context.Background()
	input := b.config.DescribeLogStreamsInput()
	streams := make([]types.LogStream, 0)

	paginator := cloudwatchlogs.NewDescribeLogStreamsPaginator(b.cwl, input)
	for paginator.HasMorePages() {
		out, err := paginator.NextPage(ctx)
		if err != nil {
			fmt.Printf("Error getting log streams: %v\n", err)
			os.Exit(2)
		}
		streams = append(streams, out.LogStreams...)
	}

	return streams
}

// GetEvents gets events from AWS given the blade configuration
func (b *Blade) GetEvents() {
	ctx := context.Background()
	formatter := b.output.Formatter()
	input := b.config.FilterLogEventsInput()

	paginator := cloudwatchlogs.NewFilterLogEventsPaginator(b.cwl, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(2)
		}

		for _, event := range page.Events {
			var message string
			if b.output.Pretty {
				message = formatEvent(formatter, event)
			} else {
				message = aws.ToString(event.Message)
			}
			message = colorizeLogLevel(message)
			if b.output.Shorten {
				message = shortenLine(message)
			}
			fmt.Println(message)
		}
	}
}

// StreamEvents continuously prints log events to the console
func (b *Blade) StreamEvents() {
	ctx := context.Background()
	var lastSeenTime *int64
	var seenEventIDs map[string]bool
	formatter := b.output.Formatter()
	input := b.config.FilterLogEventsInput()

	clearSeenEventIds := func() {
		seenEventIDs = make(map[string]bool, 0)
	}

	addSeenEventIDs := func(id *string) {
		seenEventIDs[aws.ToString(id)] = true
	}

	updateLastSeenTime := func(ts *int64) {
		if lastSeenTime == nil || *ts > *lastSeenTime {
			lastSeenTime = ts
			clearSeenEventIds()
		}
	}

	for {
		paginator := cloudwatchlogs.NewFilterLogEventsPaginator(b.cwl, input)
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(2)
			}

			for _, event := range page.Events {
				updateLastSeenTime(event.Timestamp)
				if _, seen := seenEventIDs[aws.ToString(event.EventId)]; !seen {
					var message string
					if b.output.Raw {
						message = aws.ToString(event.Message)
					} else {
						message = formatEvent(formatter, event)
					}
					message = strings.TrimRight(message, "\n")
					message = colorizeLogLevel(message)
					if b.output.Shorten {
						message = shortenLine(message)
					}
					fmt.Println(message)
					addSeenEventIDs(event.EventId)
				}
			}
		}

		if lastSeenTime != nil {
			input.StartTime = lastSeenTime
		}
		time.Sleep(1 * time.Second)
	}
}

// shortenLine truncates lines exceeding 512 characters and appends "..."
func shortenLine(line string) string {
	const maxLength = 512
	if len(line) > maxLength {
		return line[:maxLength] + "..."
	}
	return line
}

// colorizeLogLevel colorizes INFO and ERROR log levels in the message
func colorizeLogLevel(message string) string {
	// Define color formatters
	infoColor := color.New(color.FgBlack, color.BgGreen).SprintFunc()
	warnColor := color.New(color.FgBlack, color.BgYellow).SprintFunc()
	errorColor := color.New(color.FgBlack, color.BgRed).SprintFunc()
	lambdaColor := color.New(color.FgBlack, color.BgHiBlue).SprintFunc()

	// Replace INFO with colored version
	message = strings.ReplaceAll(message, "INFO", infoColor("INFO"))
	message = strings.ReplaceAll(message, "WARN", warnColor("WARN"))
	message = strings.ReplaceAll(message, "START RequestId", lambdaColor("START RequestId"))
	message = strings.ReplaceAll(message, "END RequestId", lambdaColor("END RequestId"))

	// Replace ERROR with colored version
	message = strings.ReplaceAll(message, "ERROR", errorColor("ERROR"))

	return message
}

// formatEvent returns a CloudWatch log event as a formatted string using the provided formatter
func formatEvent(formatter *colorjson.Formatter, event types.FilteredLogEvent) string {
	red := color.New(color.FgRed).SprintFunc()
	white := color.New(color.FgWhite).SprintFunc()

	str := aws.ToString(event.Message)
	bytes := []byte(str)
	date := time.UnixMilli(aws.ToInt64(event.Timestamp))
	dateStr := date.Format(time.RFC3339)
	streamStr := aws.ToString(event.LogStreamName)
	jl := map[string]interface{}{}

	if err := json.Unmarshal(bytes, &jl); err != nil {
		return fmt.Sprintf("[%s] (%s) %s", red(dateStr), white(streamStr), str)
	}

	output, _ := formatter.Marshal(jl)
	return fmt.Sprintf("[%s] (%s) %s", red(dateStr), white(streamStr), output)
}
