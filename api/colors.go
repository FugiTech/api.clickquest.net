package api

import "strings"

var colors = []Color{
	{Name: "red", Normal: "#FF0000", Dark: "#990000", Light: "#FF6565"},     //   0deg 100% sat 50,30,70% light
	{Name: "yellow", Normal: "#FFFF00", Dark: "#999900", Light: "#FFFF65"},  //  60deg 100% sat 50,30,70% light
	{Name: "green", Normal: "#00FF00", Dark: "#009900", Light: "#65FF65"},   // 120deg 100% sat 50,30,70% light
	{Name: "cyan", Normal: "#00FFFF", Dark: "#009999", Light: "#65FFFF"},    // 180deg 100% sat 50,30,70% light
	{Name: "blue", Normal: "#0000FF", Dark: "#000099", Light: "#6565FF"},    // 240deg 100% sat 50,30,70% light
	{Name: "magenta", Normal: "#FF00FF", Dark: "#990099", Light: "#FF65FF"}, // 300deg 100% sat 50,30,70% light

	// Legacy Pack
	{Name: "orange (legacy)", Normal: "#FF5721", Dark: "#B13E0F", Light: "#FF9912"}, // Variable
	{Name: "purple (legacy)", Normal: "#9900CC", Dark: "#660099", Light: "#9933CC"}, // Variable
	{Name: "gray (legacy)", Normal: "#666666", Dark: "#333333", Light: "#CCCCCC"},   // Variable

	// Obscure Color Pack
	{Name: "orange", Normal: "#FF8000", Dark: "#994D00", Light: "#FFB366"},  //  30deg 100% sat 50,30,70% light
	{Name: "seafoam", Normal: "#00FF80", Dark: "#00994D", Light: "#66FFB3"}, // 150deg 100% sat 50,30,70% light
	{Name: "purple", Normal: "#8000FF", Dark: "#4C0099", Light: "#B366FF"},  // 270deg 100% sat 50,30,70% light

	// Monochrome Color Pack
	{Name: "gray", Normal: "#808080", Dark: "#000000", Light: "#FFFFFF"},   //    0deg   0% sat 50,0,100% light
	{Name: "slate", Normal: "#404040", Dark: "#262626", Light: "#595959"},  //    0deg   0% sat 25,15,35% light
	{Name: "silver", Normal: "#BFBFBF", Dark: "#A6A6A6", Light: "#D9D9D9"}, //    0deg   0% sat 75,65,85% light
}

func GetColor(hex string) (Color, string) {
	hex = strings.ToUpper(hex)
	for _, c := range colors {
		switch hex {
		case c.Normal:
			return c, "Normal"
		case c.Dark:
			return c, "Dark"
		case c.Light:
			return c, "Light"
		}
	}
	return Color{Name: "default", Normal: "#FFFFFF"}, "Normal"
}
