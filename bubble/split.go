package bubble

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/TylerBrock/saw/config"
	"github.com/fatih/color"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
)

// Styles
var (
	activeStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62"))

	inactiveStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("63")).
		Padding(0, 1)
)

type logMessage struct {
	timestamp time.Time
	message   string
	stream    string
}

type pane struct {
	title        string
	logGroup     string
	messages     []logMessage
	scroll       int
	client       *cloudwatchlogs.Client
	config       *config.Configuration
	output       *config.OutputConfiguration
	lastEventID  map[string]bool
	lastSeenTime *int64
	mu           sync.Mutex
}

type model struct {
	panes      [2]*pane
	activePane int
	width      int
	height     int
	ready      bool
	awsConfig  *config.AWSConfiguration
	quitting   bool
}

type tickMsg time.Time
type logUpdateMsg struct {
	paneIndex int
	messages  []logMessage
}

func NewModel(logGroup1, logGroup2 string, awsConfig *config.AWSConfiguration, cfg1, cfg2 *config.Configuration, out1, out2 *config.OutputConfiguration) model {
	m := model{
		activePane: 0,
		awsConfig:  awsConfig,
	}

	m.panes[0] = &pane{
		title:       logGroup1,
		logGroup:    logGroup1,
		messages:    make([]logMessage, 0),
		config:      cfg1,
		output:      out1,
		lastEventID: make(map[string]bool),
	}

	m.panes[1] = &pane{
		title:       logGroup2,
		logGroup:    logGroup2,
		messages:    make([]logMessage, 0),
		config:      cfg2,
		output:      out2,
		lastEventID: make(map[string]bool),
	}

	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		fetchLogsCmd(0, m.panes[0], m.awsConfig),
		fetchLogsCmd(1, m.panes[1], m.awsConfig),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func fetchLogsCmd(paneIndex int, p *pane, awsConfig *config.AWSConfiguration) tea.Cmd {
	return func() tea.Msg {
		if p.client == nil {
			ctx := context.Background()
			configOpts := []func(*awsconfig.LoadOptions) error{}

			if awsConfig.Region != "" {
				configOpts = append(configOpts, awsconfig.WithRegion(awsConfig.Region))
			}

			if awsConfig.Profile != "" {
				configOpts = append(configOpts, awsconfig.WithSharedConfigProfile(awsConfig.Profile))
			}

			cfg, err := awsconfig.LoadDefaultConfig(ctx, configOpts...)
			if err != nil {
				return logUpdateMsg{paneIndex: paneIndex, messages: nil}
			}

			if awsConfig.Endpoint != "" {
				cfg.BaseEndpoint = aws.String(awsConfig.Endpoint)
			}

			p.client = cloudwatchlogs.NewFromConfig(cfg)
		}

		messages := fetchLogs(p)
		return logUpdateMsg{paneIndex: paneIndex, messages: messages}
	}
}

func fetchLogs(p *pane) []logMessage {
	p.mu.Lock()
	defer p.mu.Unlock()

	ctx := context.Background()
	input := &cloudwatchlogs.FilterLogEventsInput{
		LogGroupName: aws.String(p.logGroup),
		Interleaved:  aws.Bool(true),
	}

	if p.lastSeenTime != nil {
		input.StartTime = p.lastSeenTime
	} else if p.config.Start != "" {
		// Parse start time if provided
		currentTime := time.Now()
		relative, err := time.ParseDuration(p.config.Start)
		if err == nil {
			startTime := currentTime.Add(relative)
			input.StartTime = aws.Int64(startTime.UnixMilli())
		}
	} else {
		// Tail mode: default to last 5 minutes if no start time specified
		startTime := time.Now().Add(-5 * time.Minute)
		input.StartTime = aws.Int64(startTime.UnixMilli())
	}

	if p.config.Filter != "" {
		input.FilterPattern = aws.String(p.config.Filter)
	}

	if len(p.config.Streams) > 0 {
		streamNames := make([]string, 0, len(p.config.Streams))
		for _, stream := range p.config.Streams {
			streamNames = append(streamNames, aws.ToString(stream.LogStreamName))
		}
		input.LogStreamNames = streamNames
	}

	messages := make([]logMessage, 0)
	paginator := cloudwatchlogs.NewFilterLogEventsPaginator(p.client, input)

	// Only fetch first page initially, subsequent ticks will get more
	if paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err == nil {
			for _, event := range page.Events {
				eventID := aws.ToString(event.EventId)
				if _, seen := p.lastEventID[eventID]; !seen {
					// Strip newlines and trim spaces from message to prevent extra line feeds
					rawMessage := aws.ToString(event.Message)
					rawMessage = strings.ReplaceAll(rawMessage, "\n", " ")
					rawMessage = strings.ReplaceAll(rawMessage, "\r", " ")
					rawMessage = strings.TrimSpace(rawMessage)

					msg := logMessage{
						timestamp: time.UnixMilli(aws.ToInt64(event.Timestamp)),
						message:   colorizeLogLevel(rawMessage),
						stream:    aws.ToString(event.LogStreamName),
					}

					if p.output != nil && p.output.Shorten {
						msg.message = shortenLine(msg.message)
					}

					messages = append(messages, msg)
					p.lastEventID[eventID] = true

					// Update last seen time
					ts := event.Timestamp
					if p.lastSeenTime == nil || *ts > *p.lastSeenTime {
						p.lastSeenTime = ts
					}
				}
			}
		}
	}

	return messages
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "tab":
			m.activePane = (m.activePane + 1) % 2
		case "up", "k":
			if m.panes[m.activePane].scroll > 0 {
				m.panes[m.activePane].scroll--
			}
		case "down", "j":
			paneHeight := m.height/2 - 1
			contentHeight := paneHeight - 4
			maxScroll := len(m.panes[m.activePane].messages) - contentHeight
			if maxScroll < 0 {
				maxScroll = 0
			}
			if m.panes[m.activePane].scroll < maxScroll {
				m.panes[m.activePane].scroll++
			}
		case "g":
			// Jump to top
			m.panes[m.activePane].scroll = 0
		case "G":
			// Jump to bottom
			paneHeight := m.height/2 - 1
			contentHeight := paneHeight - 4
			maxScroll := len(m.panes[m.activePane].messages) - contentHeight
			if maxScroll < 0 {
				maxScroll = 0
			}
			m.panes[m.activePane].scroll = maxScroll
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

	case tickMsg:
		cmds = append(cmds, tickCmd())
		cmds = append(cmds, fetchLogsCmd(0, m.panes[0], m.awsConfig))
		cmds = append(cmds, fetchLogsCmd(1, m.panes[1], m.awsConfig))

	case logUpdateMsg:
		if len(msg.messages) > 0 {
			m.panes[msg.paneIndex].mu.Lock()
			m.panes[msg.paneIndex].messages = append(m.panes[msg.paneIndex].messages, msg.messages...)
			// Keep only last 1000 messages to avoid memory issues
			if len(m.panes[msg.paneIndex].messages) > 1000 {
				m.panes[msg.paneIndex].messages = m.panes[msg.paneIndex].messages[len(m.panes[msg.paneIndex].messages)-1000:]
			}
			// Auto-scroll to bottom (tail mode)
			paneHeight := m.height/2 - 1
			contentHeight := paneHeight - 4
			maxScroll := len(m.panes[msg.paneIndex].messages) - contentHeight
			if maxScroll < 0 {
				maxScroll = 0
			}
			m.panes[msg.paneIndex].scroll = maxScroll
			m.panes[msg.paneIndex].mu.Unlock()
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	if m.quitting {
		return "Goodbye!\n"
	}

	// Calculate pane height - make each pane one line smaller to fit both in window
	paneHeight := m.height/2 - 1

	// Render top pane
	topPane := m.renderPane(m.panes[0], paneHeight, m.activePane == 0, false)

	// Render bottom pane with help text in title
	bottomPane := m.renderPane(m.panes[1], paneHeight, m.activePane == 1, true)

	// Join panes vertically - exact half-screen split
	return lipgloss.JoinVertical(
		lipgloss.Left,
		topPane,
		bottomPane,
	)
}

func (m model) renderPane(p *pane, height int, active bool, showHelp bool) string {
	p.mu.Lock()
	defer p.mu.Unlock()

	style := inactiveStyle
	if active {
		style = activeStyle
	}

	// Title with optional help text
	var title string
	if showHelp {
		title = titleStyle.Render(fmt.Sprintf(" %s (%d) • q: quit • tab: switch • ↑↓/jk: scroll • g/G: top/bottom ", p.title, len(p.messages)))
	} else {
		title = titleStyle.Render(fmt.Sprintf(" %s (%d messages) ", p.title, len(p.messages)))
	}

	// Content area (account for title line and borders)
	contentHeight := height - 4 // -4 for top border, title, bottom border, and padding
	if contentHeight < 1 {
		contentHeight = 1
	}

	lines := make([]string, 0, contentHeight)

	// Calculate visible range
	totalMessages := len(p.messages)
	start := p.scroll
	end := start + contentHeight

	if start > totalMessages {
		start = totalMessages
	}
	if end > totalMessages {
		end = totalMessages
	}

	// Render visible messages
	for i := start; i < end; i++ {
		msg := p.messages[i]
		timestamp := msg.timestamp.Format("15:04:05")
		line := fmt.Sprintf("%s %s", timestamp, msg.message)

		// Truncate to width if needed
		if len(line) > m.width-6 {
			line = line[:m.width-6] + "..."
		}

		lines = append(lines, line)
	}

	// Fill remaining space with empty lines
	for len(lines) < contentHeight {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")

	// Combine title and content
	paneContent := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		content,
	)

	// Render with exact height and width to maintain consistent dimensions
	return style.Height(height).Width(m.width).Render(paneContent)
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

// shortenLine truncates lines exceeding 512 characters and appends "..."
func shortenLine(line string) string {
	const maxLength = 512
	if len(line) > maxLength {
		return line[:maxLength] + "..."
	}
	return line
}

// Run starts the bubble tea program
func Run(logGroup1, logGroup2 string, awsConfig *config.AWSConfiguration, cfg1, cfg2 *config.Configuration, out1, out2 *config.OutputConfiguration) error {
	m := NewModel(logGroup1, logGroup2, awsConfig, cfg1, cfg2, out1, out2)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running program: %w", err)
	}

	return nil
}