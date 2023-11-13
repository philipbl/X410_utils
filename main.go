package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/jacobsa/go-serial/serial"
	"github.com/urfave/cli/v2"
)

func discoverSerialDevice() (string, error) {
	files, err := ioutil.ReadDir("/dev/serial/by-id")
	if err != nil {
		return "", fmt.Errorf("error reading /dev/serial/by-id: %v", err)
	}

	var matchingDevices []string
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "usb-Digilent_Digilent_USB_Device_") && strings.Contains(file.Name(), "if02") {
			matchingDevices = append(matchingDevices, fmt.Sprintf("/dev/serial/by-id/%s", file.Name()))
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

func sendSerialCommand(addr string, command string) error {
	options := serial.OpenOptions{
		PortName:        addr,
		BaudRate:        115200,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}

	port, err := serial.Open(options)
	if err != nil {
		return fmt.Errorf("error opening serial port: %v", err)
	}
	defer port.Close()

	_, err = port.Write([]byte(command))
	if err != nil {
		return fmt.Errorf("error writing to serial port: %v", err)
	}

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "x410_util"
	app.Usage = "CLI for controlling an USRP X410 device"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "addr",
			Aliases: []string{"a"},
			Value:   "",
			Usage:   "Serial device address",
		},
	}
	app.Commands = []*cli.Command{
		{
			Name:  "start",
			Usage: "Turn on device",
			Action: func(c *cli.Context) error {
				addr := c.String("addr")
				if addr == "" {
					var err error
					addr, err = discoverSerialDevice()
					if err != nil {
						return err
					}
				}

				err := sendSerialCommand(addr, "powerbtn")
				if err != nil {
					return err
				}

				fmt.Printf("Sent 'powerbtn' command to %s\n", addr)
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

				err := sendSerialCommand(addr, "reboot")
				if err != nil {
					return err
				}

				fmt.Printf("Sent 'reboot' command to %s\n", addr)
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
