package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/jacobsa/go-serial/serial"
	"github.com/urfave/cli/v2"
)

type PowerStatus int64

const (
	OFF    PowerStatus = 0
	ON                 = 3
	UNKOWN             = -1
)

func discoverSerialDevice() (string, error) {
	prefix := "/dev/serial/by-id"
	files, err := os.ReadDir(prefix)
	if err != nil {
		return "", fmt.Errorf("error reading /dev/serial/by-id: %v", err)
	}

	var matchingDevices []string
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "usb-Digilent_Digilent_USB_Device_") && strings.Contains(file.Name(), "if02") {
			matchingDevices = append(matchingDevices, fmt.Sprintf("%s/%s", prefix, file.Name()))
		}
	}

	switch len(matchingDevices) {
	case 0:
		return "", fmt.Errorf("no matching serial devices found")
	case 1:
		return matchingDevices[0], nil
	default:
		return "", fmt.Errorf("multiple matching serial devices found. Please specify the address using --addr")
	}
}

func checkPowerStatus(addr string) (PowerStatus, error) {
	pattern := `power state (\d+)`
	re := regexp.MustCompile(pattern)

	response, err := sendSerialCommand(addr, "powerinfo")
	if err != nil {
		return UNKOWN, err
	}

	matches := re.FindStringSubmatch(response)
	if len(matches) < 2 {
		return UNKOWN, fmt.Errorf("No power state found in the input string")
	}

	// Parse the captured group as an integer
	num, err := strconv.Atoi(matches[1])
	if err != nil {
		return UNKOWN, err
	}

	// Map to the correct PowerStatus value
	if num == 0 {
		return OFF, nil
	} else if num == 3 {
		return ON, nil
	} else {
		return UNKOWN, fmt.Errorf("Unknown power status: %d", num)
	}
}

func sendSerialCommand(addr string, command string) (string, error) {
	options := serial.OpenOptions{
		PortName:        addr,
		BaudRate:        115200,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}

	// Remove newline if needed
	command = strings.TrimSuffix(command, "\n")

	port, err := serial.Open(options)
	if err != nil {
		return "", fmt.Errorf("error opening serial port: %v", err)
	}
	defer port.Close()

	log.Printf("Sending '%s' command to %s\n", command, addr)

	_, err = port.Write([]byte(command + "\n"))
	if err != nil {
		return "", fmt.Errorf("error writing to serial port: %v", err)
	}

	reader := bufio.NewReader(port)

	// For some reason, we always read back the message sent first...
	response, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("Error reading from serial port: %v", err)
	}

	// Read the response
	response, err = reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("Error reading from serial port: %v", err)
	}

	log.Printf("Response: %s", response)
	return response, nil
}

func main() {
	app := &cli.App{
		Name:  "x410_utils",
		Usage: "CLI for controlling an USRP X410 device",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "addr",
				Aliases: []string{"a"},
				Value:   "",
				Usage:   "Serial device address",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Turn on verbose flag",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "power-status",
				Usage: "Check power status",
				Action: func(c *cli.Context) error {
					addr := c.String("addr")
					verbose := c.Bool("verbose")

					if !verbose {
						log.SetOutput(io.Discard)
					}

					if addr == "" {
						var err error
						addr, err = discoverSerialDevice()
						if err != nil {
							return err
						}
					}

					status, err := checkPowerStatus(addr)
					if err != nil {
						return err
					}

					if status == ON {
						fmt.Println("on")
					} else if status == OFF {
						fmt.Println("off")
					} else {
						fmt.Println("unknown")
					}

					return nil
				},
			},
			{
				Name:  "start",
				Usage: "Turn on device",
				Action: func(c *cli.Context) error {
					addr := c.String("addr")
					verbose := c.Bool("verbose")

					if !verbose {
						log.SetOutput(io.Discard)
					}

					if addr == "" {
						var err error
						addr, err = discoverSerialDevice()
						if err != nil {
							return err
						}
					}

					status, err := checkPowerStatus(addr)
					if err != nil {
						return err
					}

					if status == ON {
						fmt.Println("Device is already on...")
						return nil
					}

					_, err = sendSerialCommand(addr, "powerbtn")
					if err != nil {
						return err
					}

					return nil
				},
			},
			{
				Name:  "shutdown",
				Usage: "Turn off device",
				Action: func(c *cli.Context) error {
					addr := c.String("addr")
					if addr == "" {
						var err error
						addr, err = discoverSerialDevice()
						if err != nil {
							return err
						}
					}

					status, err := checkPowerStatus(addr)
					if err != nil {
						return err
					}

					if status == OFF {
						fmt.Println("Device is already off...")
						return nil
					}

					_, err = sendSerialCommand(addr, "reboot")
					if err != nil {
						return err
					}

					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
