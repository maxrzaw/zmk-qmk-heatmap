package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tarm/goserial"

	log_parser "zmk-heatmap/pkg/collector"
	"zmk-heatmap/pkg/heatmap"
	"zmk-heatmap/pkg/keymap"
)

var (
	keyboardParam string
	outputParam   string
	keymapParam   string
	numSensors    int
)

func init() {
	rootCmd.AddCommand(collectCmd)

	collectCmd.Flags().StringVarP(&keyboardParam, "keyboard", "k", "auto", "e.g. /dev/tty.usbmodem144001")
	collectCmd.Flags().StringVarP(&outputParam, "output", "o", "heatmap.json", "e.g. ~/heatmap.json")
	collectCmd.Flags().StringVarP(&keymapParam, "keymap", "m", "keymap.yaml", "e.g. ~/keymap.yaml")
	collectCmd.Flags().IntVarP(&numSensors, "sensors", "s", 0, "the number of sensors (encoders) of the keyboard, this value is needed to properly collect combos")
	collectCmd.MarkFlagRequired("keymap")
}

var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Process the keystrokes from your keyboard",
	Long:  "Process the keystrokes from your keyboard and save the aggregated result in a file that can be used to generate the heatmap",
	Run: func(cmd *cobra.Command, args []string) {
		// Remove the timestamp from the log messages
		log.SetFlags(log.Flags() &^ (log.Ldate))

		keyboardPath, err := findKeyboard(keyboardParam)
		if err != nil {
			log.Fatalln(err.Error())
		}

		_, err = os.Stat(keyboardPath)
		if err != nil {
			log.Fatalln("Cannot connect to the keyboardPath:", err)
		}

		log.Println("Connecting to the keyboardPath at:", keyboardPath)

		c := &serial.Config{Name: keyboardPath, Baud: 115200}
		s, err := serial.OpenPort(c)
		if err != nil {
			log.Fatalln(err)
		}

		heatmap, err := loadHeatMap(outputParam)
		if err != nil {
			log.Fatalln(err)
		}
		if heatmap.GetPressCount() > 0 {
			log.Println("Loaded", outputParam, "with", heatmap.GetPressCount(), "key presses")
		}

		keymapFile := keymapParam
		keymapp, err := keymap.Load(keymapFile, numSensors)
		if err != nil {
			log.Fatalln("Cannot load the keymap:", err)
		}
		log.Println("Loading keymap", keymapFile)

		parser := log_parser.NewZmkLogParser(keymapp)

		// Store the collected keystrokes every 5 seconds
		ticker := time.NewTicker(5 * time.Second)
		go storeKeyStrokes(ticker, heatmap)

		// Start the key scanner
		scanner := bufio.NewScanner(s)
		for scanner.Scan() {
			_ = parser.Parse(scanner.Text(), heatmap)
		}
		if scanner.Err() != nil {
			log.Fatal(scanner.Err())
		}
	},
}

func loadHeatMap(heatmapFile string) (heatmap_ *heatmap.Heatmap, err error) {
	if _, err := os.Stat(heatmapFile); os.IsNotExist(err) {
		return heatmap.New(), nil
	}

	return heatmap.Load(heatmapFile)
}

// isZMKKeyboard attempts to detect if a serial port is a ZMK keyboard
// by reading a few lines and checking for ZMK-style logging patterns
func isZMKKeyboard(portPath string) bool {
	c := &serial.Config{Name: portPath, Baud: 115200, ReadTimeout: time.Second * 2}
	s, err := serial.OpenPort(c)
	if err != nil {
		return false
	}
	defer s.Close()

	// Try to read a few lines to see if it looks like ZMK output
	scanner := bufio.NewScanner(s)
	linesRead := 0
	maxLines := 10

	for scanner.Scan() && linesRead < maxLines {
		line := scanner.Text()
		linesRead++

		// Check for common ZMK logging patterns
		if strings.Contains(line, "[") &&
			(strings.Contains(line, "<dbg>") ||
				strings.Contains(line, "<inf>") ||
				strings.Contains(line, "position_state_changed") ||
				strings.Contains(line, "zmk")) {
			return true
		}
	}

	return false
}

// Scan for possible keyboards and returns the path for the keyboard if one and only one is connected
// returns an error otherwise
func findKeyboard(keyboardPath string) (k string, err error) {
	if keyboardPath != "" && keyboardPath != "auto" {
		return keyboardPath, nil
	}

	var candidates []string

	if runtime.GOOS == "windows" {
		// On Windows, scan common COM ports
		log.Println("Scanning for ZMK keyboards on COM ports...")
		for i := 1; i <= 20; i++ {
			portPath := fmt.Sprintf("COM%d", i)
			if isZMKKeyboard(portPath) {
				candidates = append(candidates, portPath)
				log.Printf("Found ZMK keyboard on %s\n", portPath)
			}
		}
	} else {
		// On Unix/macOS, scan /dev/tty*
		log.Println("Scanning for ZMK keyboards on /dev/tty.usbmodem*...")
		matches, _ := filepath.Glob("/dev/tty.usbmodem*")
		for _, match := range matches {
			if isZMKKeyboard(match) {
				candidates = append(candidates, match)
				log.Printf("Found ZMK keyboard on %s\n", match)
			}
		}
	}

	if len(candidates) == 0 {
		return k, errors.New("No ZMK keyboard found. Please make sure that the keyboard is connected via USB and that the firmware has USB_LOGGING enabled. See: https://zmk.dev/docs/development/usb-logging")
	}

	if len(candidates) > 1 {
		return k, errors.New("Multiple ZMK keyboards found: " + (strings.Join(candidates, ", ")) + ". Please specify the wanted keyboard with the -k flag.")
	}

	return candidates[0], nil
}

func storeKeyStrokes(ticker *time.Ticker, heatmap *heatmap.Heatmap) {
	for {
		select {
		case <-ticker.C:
			log.Println("Collected", heatmap.GetPressCount(), "key presses")
			heatmap.Save(outputParam)

			// case <- quit:
			//	ticker.Stop()
			//	return
		}
	}
}
