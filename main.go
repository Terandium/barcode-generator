package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"sync"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"github.com/fogleman/gg"
	"github.com/xuri/excelize/v2"
)

func main() {
	// Check if out folder exists, if not create it
	_, err := os.Stat("out")
	if os.IsNotExist(err) {
		fmt.Printf("Out folder does not exist, creating it...\n")
		if err := os.Mkdir("out", os.ModePerm); err != nil {
			log.Fatal("failed to create out folder")
		}
	}

	// Open the Excel file
	f, err := excelize.OpenFile("data.xlsx")
	if err != nil {
		log.Fatal("Create a excel file called: data.xlsx")
	}

	// Get all the rows in the first sheet
	rows, err := f.GetRows(f.GetSheetName(0))
	if err != nil {
		log.Fatal(err)
	}

	// Create a map to store the data
	data := make(map[string]string)

	// Iterate over the rows and populate the map
	for _, row := range rows {
		if len(row) < 2 {
			continue
		}
		barcode := row[0]
		productName := row[1]
		data[barcode] = productName
	}

	// Create a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Process each entry in the map using goroutines
	for barcode, productName := range data {
		wg.Add(1)
		go func(barcode, productName string) {
			defer wg.Done()
			createBarcode(barcode, productName)
		}(barcode, productName)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	fmt.Println("Stickers generated!")
}

func getDirectory() string {
	_, file, _, ok := runtime.Caller(1)
	if ok {
		return path.Dir(file)
	}

	return ""
}

func createBarcode(barcodeData string, customText string) {
	// Generate the barcode
	code, err := code128.Encode(barcodeData)
	if err != nil {
		log.Fatal(err)
	}

	// Scale the barcode to a suitable size
	scaledCode, err := barcode.Scale(code, 400, 100) // Use scaledCode of type barcode.Barcode
	if err != nil {
		log.Fatal(err)
	}

	// Define padding and dimensions
	const paddingTop = 10
	const width = 400
	const height = 150

	// Create a new image context
	dc := gg.NewContext(width, height)

	// Set background color to white
	dc.SetRGB(1, 1, 1)
	dc.Clear()

	// Draw the barcode onto the context with padding
	dc.DrawImage(scaledCode, 0, paddingTop)

	// Calculate the maximum font size for the custom text
	maxWidth := float64(width - 20) // 10 pixels padding on each side
	maxFontSize := calculateMaxFontSize(dc, customText, maxWidth)

	// Get the address of the current directory
	address := getDirectory()

	// Draw the small barcode text underneath the barcode
	dc.SetRGB(0, 0, 0)
	if err := dc.LoadFontFace(address+"/DejaVuSans-Bold.ttf", 14); err != nil {
		panic(err)
	}

	smallBarcode := trimBarcode(barcodeData)
	dc.DrawStringAnchored(smallBarcode, width/2, height-32, 0.5, 0.5)

	// Set text properties and draw the custom text below the barcode
	dc.SetRGB(0, 0, 0)
	if err := dc.LoadFontFace(address+"/DejaVuSans-Bold.ttf", maxFontSize); err != nil {
		panic(err)
	}
	dc.DrawStringAnchored(customText, width/2, height-15, 0.5, 0.5)

	// Save the image to a file
	err = dc.SavePNG("out/" + smallBarcode + ".png")
	if err != nil {
		log.Fatal(err)
	}
}

func calculateMaxFontSize(dc *gg.Context, text string, maxWidth float64) float64 {
	fontSize := 20.0
	for {
		address := getDirectory()
		if err := dc.LoadFontFace(address+"/DejaVuSans-Bold.ttf", fontSize); err != nil {
			panic(err)
		}
		width, _ := dc.MeasureString(text)
		if width <= maxWidth {
			return fontSize
		}
		fontSize -= 1
		if fontSize <= 0 {
			return 1
		}
	}
}

func trimBarcode(s string) string {
	// Check if the string is long enough
	if len(s) <= 12 {
		panic(fmt.Errorf("barcode is too short: %s", s))
	}

	// Trim 8 characters from the front and 1 character from the end
	return s[8 : len(s)-1]
}
