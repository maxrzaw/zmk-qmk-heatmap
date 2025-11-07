package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"zmk-heatmap/pkg/heatmap"
)

var input1Param string
var input2Param string
var mergeOutputParam string

func init() {
	rootCmd.AddCommand(mergeCmd)

	mergeCmd.Flags().StringVarP(&input1Param, "input1", "1", "", "First heatmap file to merge")
	mergeCmd.Flags().StringVarP(&input2Param, "input2", "2", "", "Second heatmap file to merge")
	mergeCmd.Flags().StringVarP(&mergeOutputParam, "output", "o", "merged_heatmap.json", "Output file for merged heatmap")
	mergeCmd.MarkFlagRequired("input1")
	mergeCmd.MarkFlagRequired("input2")
}

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge two heatmap files",
	Long:  "Merge two heatmap JSON files by combining their key and combo press counts",
	Run: func(cmd *cobra.Command, args []string) {
		// Load first heatmap
		h1, err := heatmap.Load(input1Param)
		if err != nil {
			log.Fatalf("Cannot load %s: %v", input1Param, err)
		}

		// Load second heatmap
		h2, err := heatmap.Load(input2Param)
		if err != nil {
			log.Fatalf("Cannot load %s: %v", input2Param, err)
		}

		// Merge: register all presses from h2 into h1
		for _, kp := range h2.KeyPresses {
			for i := 0; i < kp.Taps; i++ {
				h1.RegisterKeyPress(kp.Layer, kp.Position, heatmap.Tap)
			}
			for i := 0; i < kp.Holds; i++ {
				h1.RegisterKeyPress(kp.Layer, kp.Position, heatmap.Hold)
			}
			for i := 0; i < kp.Shifts; i++ {
				h1.RegisterKeyPress(kp.Layer, kp.Position, heatmap.Shifted)
			}
		}
		for _, cp := range h2.ComboPresses {
			for i := 0; i < cp.Taps; i++ {
				h1.RegisterComboPress(cp.Layer, cp.Keys, heatmap.Tap)
			}
			for i := 0; i < cp.Holds; i++ {
				h1.RegisterComboPress(cp.Layer, cp.Keys, heatmap.Hold)
			}
			for i := 0; i < cp.Shifts; i++ {
				h1.RegisterComboPress(cp.Layer, cp.Keys, heatmap.Shifted)
			}
		}

		// Save merged heatmap
		err = h1.Save(mergeOutputParam)
		if err != nil {
			log.Fatalf("Cannot save %s: %v", mergeOutputParam, err)
		}
		log.Printf("Merged heatmap saved to %s", mergeOutputParam)
	},
}
