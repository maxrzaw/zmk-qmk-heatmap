package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"zmk-heatmap/pkg/heatmap"
)

var mergeOutputParam string

func init() {
	rootCmd.AddCommand(mergeCmd)

	mergeCmd.Flags().StringVarP(&mergeOutputParam, "output", "o", "merged_heatmap.json", "Output file for merged heatmap")
	mergeCmd.Args = cobra.MinimumNArgs(2)
}

var mergeCmd = &cobra.Command{
	Use:   "merge INPUT1 INPUT2 [INPUT3...]",
	Short: "Merge two or more heatmap files",
	Long: `Merge multiple heatmap JSON files by combining their key and combo press counts.

Examples:
  zmk-heatmap merge heatmap1.json heatmap2.json
  zmk-heatmap merge heatmap1.json heatmap2.json heatmap3.json --output combined.json`,
	Run: func(cmd *cobra.Command, args []string) {
		// Validate input files exist
		for _, input := range args {
			if _, err := os.Stat(input); os.IsNotExist(err) {
				log.Fatalf("Input file does not exist: %s", input)
			}
		}

		// Load first heatmap as base
		baseHeatmap, err := heatmap.Load(args[0])
		if err != nil {
			log.Fatalf("Cannot load %s: %v", args[0], err)
		}

		// Merge remaining heatmaps
		for _, input := range args[1:] {
			h, err := heatmap.Load(input)
			if err != nil {
				log.Fatalf("Cannot load %s: %v", input, err)
			}

			// Register all presses from h into baseHeatmap
			for _, kp := range h.KeyPresses {
				for i := 0; i < kp.Taps; i++ {
					baseHeatmap.RegisterKeyPress(kp.Layer, kp.Position, heatmap.Tap)
				}
				for i := 0; i < kp.Holds; i++ {
					baseHeatmap.RegisterKeyPress(kp.Layer, kp.Position, heatmap.Hold)
				}
				for i := 0; i < kp.Shifts; i++ {
					baseHeatmap.RegisterKeyPress(kp.Layer, kp.Position, heatmap.Shifted)
				}
			}
			for _, cp := range h.ComboPresses {
				for i := 0; i < cp.Taps; i++ {
					baseHeatmap.RegisterComboPress(cp.Layer, cp.Keys, heatmap.Tap)
				}
				for i := 0; i < cp.Holds; i++ {
					baseHeatmap.RegisterComboPress(cp.Layer, cp.Keys, heatmap.Hold)
				}
				for i := 0; i < cp.Shifts; i++ {
					baseHeatmap.RegisterComboPress(cp.Layer, cp.Keys, heatmap.Shifted)
				}
			}
		}

		// Save merged heatmap
		err = baseHeatmap.Save(mergeOutputParam)
		if err != nil {
			log.Fatalf("Cannot save %s: %v", mergeOutputParam, err)
		}
		log.Printf("Merged heatmap saved to %s", mergeOutputParam)
	},
}
