package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/seanhood/go-vedirect/vedirect"
	"go.bug.st/serial/enumerator"
)

var (
	version = "dev"
	commit  = ""
	date    = ""
)

// DeviceList implements flag.Value for collecting multiple -dev flags
type DeviceList []string

func (d *DeviceList) String() string {
	if d == nil || len(*d) == 0 {
		return ""
	}
	return strings.Join(*d, ",")
}

func (d *DeviceList) Set(value string) error {
	*d = append(*d, value)
	return nil
}

// Config is where we keep the flag vars
type Config struct {
	device  DeviceList
	outFile string
	verbose bool
	ver     bool

	// Auto mode detects & streams from all connected VE.Direct-USB bridges.
	// CAUTION: not hot-plug safe, restart required to detect new devices.
	Auto bool

	// Match filters enumerated ports by name when used with Auto mode.
	// Takes precedence over default product name matching.
	Match string

	// Watchdog exits with error if no MQTT message published in 60 seconds
	Watchdog bool

	// extras is the raw CLI string for additional key-value pairs
	extras string

	MQTT struct {
		Server  string
		Topic   string
		TLSKey  string
		TLSCert string
		TLSCA   string
	}
}

type Service struct {
	Config *Config

	MQTT           mqtt.Client
	OutputFile     *os.File
	fileWriteMux   sync.Mutex // Protect concurrent file writes
	lastPublishMux sync.Mutex
	lastPublish    time.Time
}

// parseExtras parses a comma-separated string of key=value pairs into a map.
// Example input: "hostname=shedtofu,installation=shed"
func parseExtras(extrasStr string) (map[string]string, error) {
	result := make(map[string]string)
	if extrasStr == "" {
		return result, nil
	}

	pairs := strings.Split(extrasStr, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid extras format: %q (expected key=value)", pair)
		}
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])
		if key == "" {
			return nil, fmt.Errorf("empty key in extras pair: %q", pair)
		}
		result[key] = value
	}

	return result, nil
}

func main() {
	c := new(Config)
	flag.Var(&c.device, "dev", "full path to serial device node (can be specified multiple times)")
	flag.StringVar(&c.MQTT.Server, "mqtt.server", "tcp://localhost:1883", "MQTT Server address")
	flag.StringVar(&c.MQTT.Topic, "mqtt.topic", "", "The MQTT Topic to publish messages to")

	flag.StringVar(&c.MQTT.TLSKey, "mqtt.tls_key", "", "MQTT TLS Private Key")
	flag.StringVar(&c.MQTT.TLSCert, "mqtt.tls_cert", "", "MQTT TLS Private Cert")
	flag.StringVar(&c.MQTT.TLSCA, "mqtt.tls_rootca", "", "MQTT TLS Root CA")

	flag.StringVar(&c.outFile, "out-file", "", "File to write json data to")
	flag.BoolVar(&c.verbose, "verbose", false, "Verbose Output")

	flag.BoolVar(&c.Auto, "auto", false, "Auto detect VE.Direct-USB bridges")
	flag.StringVar(&c.Match, "match", "", "Filter enumerated ports by name (used with -auto)")
	flag.BoolVar(&c.Watchdog, "watchdog", false, "Exit with error if no MQTT message published in 60 seconds")
	flag.StringVar(&c.extras, "extras", "", "Additional key-value pairs to include in payload (format: key1=value1,key2=value2)")

	flag.BoolVar(&c.ver, "v", false, "Print Version")
	flag.Parse()

	if c.ver {
		fmt.Println(buildVersion(version, commit, date))
		os.Exit(0)
	}

	// Parse extras
	extrasMap, err := parseExtras(c.extras)
	if err != nil {
		log.Fatalf("Error parsing extras: %v", err)
	}

	// Inject version information into all MQTT messages
	if _, exists := extrasMap["publisher_version"]; !exists {
		extrasMap["publisher_version"] = version
	}
	if commit != "" && commit != "dev" {
		if _, exists := extrasMap["publisher_commit"]; !exists {
			extrasMap["publisher_commit"] = commit
		}
	}
	if date != "" {
		if _, exists := extrasMap["publisher_build_date"]; !exists {
			extrasMap["publisher_build_date"] = date
		}
	}

	// Set default device if none specified and not in auto mode
	if len(c.device) == 0 && !c.Auto {
		c.device = DeviceList{"/dev/ttyUSB0"}
	}

	svc := &Service{
		Config: c,
	}

	// Mqtt Setup
	if c.MQTT.Topic != "" {
		opts := *mqtt.NewClientOptions()
		opts.SetMaxReconnectInterval(1 * time.Second)

		certpool := x509.NewCertPool()
		pemCerts, err := ioutil.ReadFile(c.MQTT.TLSCA)
		if err == nil {
			certpool.AppendCertsFromPEM(pemCerts)
		}

		if c.MQTT.TLSCert != "" && c.MQTT.TLSKey != "" {
			cert, err := tls.LoadX509KeyPair(c.MQTT.TLSCert, c.MQTT.TLSKey)
			if err != nil {
				log.Fatal(err)
			}

			tlsConfig := &tls.Config{
				InsecureSkipVerify: true,
				ClientAuth:         tls.NoClientCert,
				Certificates:       []tls.Certificate{cert},
				ClientCAs:          nil,
				RootCAs:            certpool,
			}
			opts.SetTLSConfig(tlsConfig)
		}
		opts.AddBroker(c.MQTT.Server)

		svc.MQTT = mqtt.NewClient(&opts)
		if token := svc.MQTT.Connect(); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
			return
		}

		log.Printf("Connected to %s\n", c.MQTT.Server)

	}

	// file output setup
	if c.outFile != "" {
		log.Printf("Saving data to: %s", c.outFile)
		var err error
		svc.OutputFile, err = os.OpenFile(c.outFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
		if err != nil {
			log.Fatal(err)
		}

		defer svc.OutputFile.Close()
	}

	if c.outFile == "" && c.MQTT.Topic == "" {
		log.Fatal("No output configured, please set -out-file or -mqtt.topic")
	}

	if c.Watchdog && c.MQTT.Topic == "" {
		log.Fatal("Watchdog requires MQTT to be configured")
	}

	// Start watchdog if enabled
	if c.Watchdog {
		svc.lastPublish = time.Now()
		go svc.watchdog()
	}

	if c.Auto {
		// Create an enumerator, search for all VE.Direct-USB bridges
		ports, err := enumerator.GetDetailedPortsList()
		if err != nil {
			log.Fatal(err)
		}

		// Create a wait group to track streamer goroutines.
		wg := &sync.WaitGroup{}

		for _, port := range ports {
			// Match by port name if -match flag is provided, otherwise find by product name constant
			shouldStart := false
			if c.Match != "" {
				shouldStart = strings.Contains(port.Name, c.Match)
			} else {
				shouldStart = port.Product == "VE Direct cable"
			}

			if shouldStart {
				fmt.Printf("Starting streamer: %s (sn=%s)\n", port.Name, port.SerialNumber)
				wg.Add(1)

				go func(port *enumerator.PortDetails) {
					defer wg.Done()

					deviceExtras := map[string]string{
						"vedirect_serial": port.SerialNumber,
						"vedirect_port":   port.Name,
					}
					// Merge user-provided extras (user extras override auto-detected ones)
					for k, v := range extrasMap {
						deviceExtras[k] = v
					}

					svc.streamFromPath(port.Name, deviceExtras)
				}(port)
			}
		}

		wg.Wait()
	} else if len(c.device) > 1 {
		// Multi-device manual mode
		log.Printf("Starting manual multi-device mode with %d devices\n", len(c.device))

		wg := &sync.WaitGroup{}

		for idx, devicePath := range c.device {
			fmt.Printf("Starting streamer: %s (manual device %d)\n", devicePath, idx)
			wg.Add(1)

			go func(path string, index int) {
				defer wg.Done()

				deviceExtras := map[string]string{
					"vedirect_port":        path,
					"vedirect_device_index": strconv.Itoa(index),
				}
				// Merge user-provided extras
				for k, v := range extrasMap {
					deviceExtras[k] = v
				}

				svc.streamFromPath(path, deviceExtras)
			}(devicePath, idx)
		}

		wg.Wait()
	} else if len(c.device) == 1 {
		// Single device mode (backward compatible)
		svc.streamFromPath(c.device[0], extrasMap)
	} else {
		// No devices specified
		log.Fatal("No device specified, please provide -dev flag or use -auto")
	}
}

func (svc *Service) streamFromPath(path string, extras map[string]string) {
	var reader io.Reader

	stat, err := os.Stat(path)
	if err != nil {
		log.Fatal(err)
	}

	// Should probably be in go-vedirect package
	if stat.Mode().IsRegular() {
		reader = vedirect.OpenFile(path)
	} else {
		reader = vedirect.OpenSerial(path)
	}

	s := vedirect.NewStream(reader)
	for {
		b, checksum := s.ReadBlock()
		if checksum == 0 {

			fields := b.Fields()

			fields["timestamp"] = strconv.FormatInt(time.Now().Unix(), 10)

			// Copy any extra fields into the outgoing payload.
			for k, v := range extras {
				fields[k] = v
			}

			jsonPayload, err := json.Marshal(fields)
			if err != nil {
				log.Fatal(err)
			}

			if svc.Config.verbose {
				log.Println(string(jsonPayload))
			}

			if svc.MQTT != nil {
				token := svc.MQTT.Publish(svc.Config.MQTT.Topic, 1, false, jsonPayload)
				if token.Wait() && token.Error() == nil {
					svc.lastPublishMux.Lock()
					svc.lastPublish = time.Now()
					svc.lastPublishMux.Unlock()
				}
			}

			if svc.OutputFile != nil {
				svc.fileWriteMux.Lock()
				_, err := svc.OutputFile.Write(jsonPayload)
				svc.fileWriteMux.Unlock()
				if err != nil {
					log.Fatal(err)
				}
			}

		} else {
			log.Printf("Bad block on %s, skipping (checksum residue=%d, parsed fields=%v)", path, checksum, b.Fields())
		}
	}
}

func (svc *Service) watchdog() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		svc.lastPublishMux.Lock()
		elapsed := time.Since(svc.lastPublish)
		svc.lastPublishMux.Unlock()

		if elapsed > 60*time.Second {
			log.Printf("ERROR: No MQTT message published in %.0f seconds\n", elapsed.Seconds())
			os.Exit(1)
		}
	}
}

func buildVersion(version, commit, date string) string {
	var result = version
	if commit != "" {
		result = fmt.Sprintf("%s\ncommit: %s", result, commit)
	}
	if date != "" {
		result = fmt.Sprintf("%s\nbuilt at: %s", result, date)
	}
	return result
}
