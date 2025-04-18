package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/SecurityGeekIO/zscaler-sdk-go/v3/zscaler/zdx"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v3/zscaler/zdx/services"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v3/zscaler/zdx/services/common"
	"github.com/SecurityGeekIO/zscaler-sdk-go/v3/zscaler/zdx/services/reports/devices"
	"github.com/olekukonko/tablewriter"
)

type Device struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Platform string `json:"platform"`
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	apiKey := os.Getenv("ZDX_API_KEY_ID")
	apiSecret := os.Getenv("ZDX_API_SECRET")

	// Initialize ZDX configuration
	zdxCfg, err := zdx.NewConfiguration(
		zdx.WithZDXAPIKeyID(apiKey),
		zdx.WithZDXAPISecret(apiSecret),
		// Uncomment the line below if connecting to a custom ZDX cloud
		// zdx.WithZDXCloud("zdxbeta"),
		zdx.WithDebug(true),
	)
	if err != nil {
		log.Fatalf("Error creating ZDX configuration: %v", err)
	}

	// Initialize ZDX client
	zdxClient, err := zdx.NewClient(zdxCfg)
	if err != nil {
		log.Fatalf("Error creating ZDX client: %v", err)
	}

	// Wrap the ZDX client in a Service instance
	service := services.New(zdxClient)

	ctx := context.Background()

	// Prompt for from time
	fmt.Print("Enter start time in Unix Epoch (optional, defaults to 2 hours ago): ")
	fromInput, _ := reader.ReadString('\n')
	fromInput = strings.TrimSpace(fromInput)

	// Prompt for to time
	fmt.Print("Enter end time in Unix Epoch (optional, defaults to now): ")
	toInput, _ := reader.ReadString('\n')
	toInput = strings.TrimSpace(toInput)

	// Set default time range to last 2 hours if not provided
	now := time.Now()
	fromTime := now.Add(-2 * time.Hour).Unix()
	toTime := now.Unix()

	var fromInt, toInt int

	if fromInput != "" {
		parsedFrom, err := strconv.ParseInt(fromInput, 10, 64)
		if err != nil {
			log.Fatalf("[ERROR] Invalid start time: %v\n", err)
		}
		fromInt, err = safeIntConversion(parsedFrom)
		if err != nil {
			log.Fatalf("[ERROR] %v\n", err)
		}
	} else {
		fromInt, err = safeIntConversion(fromTime)
		if err != nil {
			log.Fatalf("[ERROR] %v\n", err)
		}
	}

	if toInput != "" {
		parsedTo, err := strconv.ParseInt(toInput, 10, 64)
		if err != nil {
			log.Fatalf("[ERROR] Invalid end time: %v\n", err)
		}
		toInt, err = safeIntConversion(parsedTo)
		if err != nil {
			log.Fatalf("[ERROR] %v\n", err)
		}
	} else {
		toInt, err = safeIntConversion(toTime)
		if err != nil {
			log.Fatalf("[ERROR] %v\n", err)
		}
	}

	// Define filters
	filters := devices.GetDevicesFilters{
		GetFromToFilters: common.GetFromToFilters{
			From: fromInt,
			To:   toInt,
		},
	}

	// Get all devices
	devicesList, _, err := devices.GetAllDevices(ctx, service, filters)
	if err != nil {
		log.Fatalf("[ERROR] getting all devices failed: %v\n", err)
	}

	// Extract device details and display in table format
	var deviceData []Device
	for _, device := range devicesList {
		// Extract platform information from device name
		parts := strings.Split(device.Name, "(")
		name := parts[0]
		platform := ""
		if len(parts) > 1 {
			platform = strings.TrimSuffix(parts[1], ")")
		}
		deviceData = append(deviceData, Device{
			ID:       device.ID,
			Name:     name,
			Platform: platform,
		})
	}

	// Display the data in a formatted table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"device_id", "device_name", "platform"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")

	for _, device := range deviceData {
		table.Append([]string{strconv.Itoa(device.ID), device.Name, device.Platform})
	}

	table.Render()
}

func safeIntConversion(value int64) (int, error) {
	if value > int64(int(^uint(0)>>1)) || value < int64(-int(^uint(0)>>1)-1) {
		return 0, fmt.Errorf("value %d is out of range for int type", value)
	}
	return int(value), nil
}
