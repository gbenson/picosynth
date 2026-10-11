package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

func main() {
	if err := Main(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	fmt.Println("Firmware updated")
}

func Main() error {
	if len(os.Args) != 3 || os.Args[1] != "flash" {
		cmd := filepath.Base(os.Args[0])
		fmt.Fprintf(os.Stderr, "usage: %s flash FILENAME\n", cmd)
		os.Exit(2)
	}

	filename := os.Args[2]
	filebytes, err := os.ReadFile(filename)
	if err != nil {
		return err
	} else if len(filebytes) == 0 {
		return fmt.Errorf("%s: empty file", filename)
	} else if len(filebytes)&511 != 0 {
		return fmt.Errorf("%s: not a valid UF2", filename)
	}
	//fmt.Println("Read", filename)

	outdir, err := GetBootloaderRoot()
	if err != nil {
		return err
	}
	fmt.Println("Bootloader mounted at", outdir)

	filename = filepath.Join(outdir, "flash.uf2")
	return os.WriteFile(filename, filebytes, 0666)
}

func GetBootloaderRoot() (string, error) {
	path, err := getBootloaderRoot()
	if err != nil || path != "" {
		return path, err
	}

	if err := EnterBootloader(); err != nil {
		return "", err
	}

	const maxAttempts = 10
	for _ = range maxAttempts {
		time.Sleep(time.Second / 4)

		path, err = getBootloaderRoot()
		if err != nil {
			return "", err
		} else if path != "" {
			return path, nil
		}
	}
	return "", errors.New("bootloader root not found")
}

func getBootloaderRoot() (string, error) {
	paths, err := filepath.Glob("/mnt/chromeos/removable/*/INFO_UF2.TXT")
	if err != nil || len(paths) == 0 {
		return "", err
	} else if len(paths) > 1 {
		return "", errors.New("multiple bootloader root candidates")
	}

	path := paths[0]
	// if b, err := os.ReadFile(path); err != nil {
	// 	fmt.Fprintln(os.Stderr, "warning:", err)
	// } else {
	// 	fmt.Println("Found", strings.TrimSpace(string(b)))
	// }

	return filepath.Dir(path), nil
}

func EnterBootloader() error {
	port, err := GetDefaultPort()
	if err != nil {
		return err
	}

	fmt.Println("Entering bootloader")
	p, err := serial.Open(port, &serial.Mode{BaudRate: 1200})
	if err != nil {
		return err
	}
	defer p.Close()

	if err := p.SetDTR(false); err != nil {
		fmt.Fprintln(os.Stderr, "warning:", err)
	}

	return nil
}

func GetDefaultPort() (string, error) {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return "", err
	}

	var result string
	for _, p := range ports {
		if !p.IsUSB {
			continue
		}

		vid, err := strconv.ParseUint(p.VID, 16, 16)
		if err != nil {
			return "", fmt.Errorf("%s: bad VID: %w", p.VID, err)
		}

		pid, err := strconv.ParseUint(p.PID, 16, 16)
		if err != nil {
			return "", fmt.Errorf("%s: bad PID: %w", p.PID, err)
		}

		switch (vid << 16) | pid {
		case 0x2e8a_000a: // Raspberry Pi Pico, Raspberry Pi Pico2
		default:
			fmt.Println("ignoring", (vid<<16)|pid)
			continue
		}

		fmt.Printf("Found %s %s on %s\n", p.Manufacturer, p.Product, p.Name)

		if result != "" {
			return "", errors.New("multiple port candidates")
		}

		result = p.Name
	}

	if result == "" {
		return "", errors.New("port not found")
	}

	return result, nil
}
