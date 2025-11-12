package heatmap

import (
	"image/color"
	"strconv"
)

type RGB struct {
	R uint8
	G uint8
	B uint8
}

func FromHex(hex string) RGB {
	var r, g, b uint8
	if len(hex) == 7 && hex[0] == '#' {
		rv, err := strconv.ParseUint(hex[1:3], 16, 8)
		if err != nil {
			return RGB{}
		}
		gv, err := strconv.ParseUint(hex[3:5], 16, 8)
		if err != nil {
			return RGB{}
		}
		bv, err := strconv.ParseUint(hex[5:7], 16, 8)
		if err != nil {
			return RGB{}
		}
		r = uint8(rv)
		g = uint8(gv)
		b = uint8(bv)
		return RGB{R: r, G: g, B: b}
	}

	panic("invalid hex color format")
}

func ToHex(color RGB) string {
	return "#" + strconv.FormatUint(uint64(color.R), 16) + strconv.FormatUint(uint64(color.G), 16) + strconv.FormatUint(uint64(color.B), 16)
}

func RgbaToRgb(color color.RGBA, background RGB) RGB {
	source := toFloat(color)
	bg := background.toFloat()
	luminance := float32(color.A) / 255

	target := floatColor{
		r: ((1 - luminance) * bg.r) + (luminance * source.r),
		g: ((1 - luminance) * bg.g) + (luminance * source.g),
		b: ((1 - luminance) * bg.b) + (luminance * source.b),
	}
	return target.toRGB()
}

type floatColor struct {
	r float32
	g float32
	b float32
}

func toFloat(rgba color.RGBA) floatColor {
	return floatColor{
		r: float32(rgba.R) / 255,
		g: float32(rgba.G) / 255,
		b: float32(rgba.B) / 255,
	}
}

func (rgb RGB) toFloat() floatColor {
	return floatColor{
		r: float32(rgb.R) / 255,
		g: float32(rgb.G) / 255,
		b: float32(rgb.B) / 255,
	}
}

func (c floatColor) toRGB() RGB {
	return RGB{
		R: uint8(255 * c.r),
		G: uint8(255 * c.g),
		B: uint8(255 * c.b),
	}
}
