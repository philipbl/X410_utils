# X410 Utils

This program reads information from the Microprocessor built into the USRP X410. Currently, it only supports three commands:

- `power-status`: Checks the power status of the X410. The possible values are off, on, or unknown.

- `start`: Starts the X410. This is the same as pressing the power button of the X410's front panel. The X410 must be plugged in for this command to work.

- `shutdown`: Shuts down the X410. This is the same as pressing the power button of the X410's front panel. The X410 must be plugged in for this command to work.

For the `start` and `shutdown` command, it checks the current status and only runs the command if it makes sense. For example, if you start the X410 but it is already started, then the command will not run.

This utility is using the [serial connection](https://files.ettus.com/manual/page_usrp_x4xx.html#x4xx_getting_started_serial) built into the X410. There are many other commands that you can run over serial, but this utility does not support them at that time.
