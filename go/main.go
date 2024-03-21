package main

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/samwho/streamdeck"
)

func main() {
	now := time.Now()
	fileName := fmt.Sprintf("streamdeck-inboxes-%s-*.log", now.Format("2006-01-02t15h04m05s"))
	f, err := os.CreateTemp("logs", fileName)
	if err != nil {
		log.Fatalf("error creating temp file: %v", err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Printf("unable to close file “%s”: %v\n", fileName, err)
		}
	}(f)

	log.SetOutput(f)

	ctx := context.Background()
	if err := run(ctx); err != nil {
		log.Printf("%v\n", err)

		return
	}
}

func run(ctx context.Context) error {
	params, err := streamdeck.ParseRegistrationParams(os.Args)
	if err != nil {
		return err
	}

	client := streamdeck.NewClient(ctx, params)
	setup(client)

	return client.Run()
}

const (
	GhubBatterySection        = "percentage"
	GhubBatteryWarningSection = "warning"
)

const updateFrequency = 5 * time.Second

func setup(client *streamdeck.Client) {
	dbFilePath, err := getDbFilepath()
	if err != nil {
		return
	}

	if _, err := os.Stat(dbFilePath); os.IsNotExist(err) {
		log.Println("File does not exist", err)

		return
	}

	seeBatteryAction := client.Action("ca.michaelabon.logitech-battery.see")

	var quit chan struct{}

	seeBatteryAction.RegisterHandler(
		streamdeck.WillAppear,
		func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
			ticker := time.NewTicker(updateFrequency)
			quit = make(chan struct{})
			go func() {
				for {
					select {
					case <-ticker.C:
						doUpdate(ctx, client, dbFilePath)
					case <-quit:
						ticker.Stop()

						return
					}
				}
			}()

			doUpdate(ctx, client, dbFilePath)

			return nil
		},
	)

	seeBatteryAction.RegisterHandler(
		streamdeck.WillDisappear,
		func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
			close(quit)

			return nil
		},
	)

	seeBatteryAction.RegisterHandler(
		streamdeck.KeyUp,
		func(ctx context.Context, client *streamdeck.Client, event streamdeck.Event) error {
			err := launchGHub()
			if err != nil {
				log.Println("unable to launch G Hub", err)
			}

			return err
		},
	)
}

type BatteryStat struct {
	IsCharging bool
	Percentage float64
}

func doUpdate(ctx context.Context, client *streamdeck.Client, dbFilePath string) {
	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		log.Println("unable to open sqlite3 database", err)

		return
	}

	var file []byte
	err = db.QueryRow(`SELECT FILE FROM DATA ORDER BY _id DESC LIMIT 1;`).Scan(&file)
	if err != nil {
		log.Println("unable to read from database", err)
		if err = db.Close(); err != nil {
			log.Println("unable to close database", err)
		}

		return
	}

	var dev map[string]json.RawMessage

	err = json.Unmarshal(file, &dev)
	if err != nil {
		log.Println("unable to unmarshal the raw bytes", string(file))
		if err = db.Close(); err != nil {
			log.Println("unable to close database", err)
		}

		return
	}

	batteryStats := make(map[string]BatteryStat)
	expectedBatteryStatsPerLine := 3

	for key, value := range dev {
		if !strings.HasPrefix(key, "battery") {
			continue
		}

		splitName := strings.Split(key, "/")
		if len(splitName) != expectedBatteryStatsPerLine {
			continue
		}

		if splitName[2] != GhubBatterySection && splitName[2] != GhubBatteryWarningSection {
			continue
		}

		if _, ok := batteryStats[splitName[1]]; ok {
			if splitName[2] == GhubBatteryWarningSection {
				continue
			}
		}

		var batteryValue BatteryStat
		err = json.Unmarshal(value, &batteryValue)
		if err != nil {
			log.Println("unable to unmarshal the battery value", string(value))
			if err = db.Close(); err != nil {
				log.Println("unable to close database", err)
			}

			return
		}

		batteryStats[splitName[1]] = batteryValue
	}

	log.Println("Success", batteryStats)

	encodedSvg := DrawIcon(batteryStats)
	err = client.SetImage(ctx, encodedSvg, streamdeck.HardwareAndSoftware)
	if err != nil {
		log.Println("Unable to set image", err)
		if err = db.Close(); err != nil {
			log.Println("unable to close database", err)
		}

		return
	}

	if err = db.Close(); err != nil {
		log.Println("unable to close database", err)
	}
}
