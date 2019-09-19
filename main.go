package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/fatih/color"
)

var project Project
var channels []Channel
var mixer []Insert
var currentInsert = 0
var currentSlot uint16 = 0xFFFF
var lastGenerator string

func parseHead(file *os.File) {

	FLPHeader := header{}
	readStruct(file, &FLPHeader, 14)
}

func getBufferLen(r io.Reader) uint32 {
	data, _ := readInt8(r)
	dataLen := uint32(data & 0x7F)
	var shift uint8

	for (data & 0x80) != 0 {
		data, _ = readInt8(r)
		shift += 7
		dataLen = dataLen | (uint32(data&0x7F) << shift)
	}
	return dataLen
}

func parseByte(event uint8, r io.Reader) {
	data, _ := readInt8(r)

	switch event {
	case ByteChanType:
		channels[len(channels)-1].ctype = data
	default:
	}
	//fmt.Printf("Byte (%#v): %d\n", event, data)
}

func parseWord(event uint8, r io.Reader) {
	data := readInt16(r)

	switch event {
	case WordNewChan:
		newChan := Channel{id: data}
		channels = append(channels, newChan)
	case WordCurrentSlotNum:
		//fmt.Println("CURRENT SLOT ", data)
		currentSlot = data
		mixer[currentInsert].slots[currentSlot].id = data
	default:
	}

}
func parseDWord(event uint8, r io.Reader) {
	data := readInt32(r)

	switch event {
	case DWordFineTempo:
		project.tempo = data
	default:

	}
	//fmt.Printf("DWord (%#v): %d\n", event, data)
}

func parsePlugin(r io.Reader) Plugin {
	plugin := Plugin{}
	pluginType := readInt32(r) & 0x0F
	plugin.pluginType = VSTType[uint8(pluginType)]

	if pluginType < 8 {
		plugin.name = lastGenerator
		return plugin
	}

	for {
		event := readInt32(r)
		size := uint32(readInt64(r))

		if event == 0 {
			return plugin
		}
		data, _ := readNextBytes(r, size)
		switch event {
		case PluginPluginInfo:
			break
		case PluginVendorName:
			plugin.vendor = string(data)
			break
		case PluginFilename:
			plugin.filename = string(data)

			break
		case PluginName:
			plugin.name = string(data)
			break
		case PluginState:

			break
		default:

			break
		}

	}
}

func parseData(event uint8, r io.Reader) {
	size := getBufferLen(r)

	switch event {

	case DataPatternNotes:
		readNextBytes(r, size)
	case DataNewPlugin:
		readNextBytes(r, size)
	//fmt.Printf("Data [%d] New Plugin Data: % x\n", event, data)

	case DataPluginParams:
		data, _ := readNextBytes(r, size)
		plugin := parsePlugin(bytes.NewReader(data))

		if currentSlot == 0xFFFF {
			channels[len(channels)-1].plugin = plugin
		} else {
			mixer[currentInsert].slots[currentSlot].plugin = plugin
		}

	case DataInsertRoutes: // We have this at the end of every mixer channel
		//fmt.Println("Current Insert", currentInsert)
		//		if mixer[currentInsert].name != "" {
		newInsert := Insert{}
		mixer = append(mixer, newInsert)
		currentInsert++

		readNextBytes(r, size)
	default:
		readNextBytes(r, size)
	}

	//fmt.Printf("data - Event: %d - Size: %d\n", event, size)
}

func parseText(event uint8, r io.Reader) {
	size := getBufferLen(r)
	data, _ := readNextBytes(r, size)

	switch event {
	case TextVersion:
		project.version = string(data)
	case TextTitle:
		project.title = unicodeToString(data)
	case TextPluginName:
		//fmt.Printf("Plugin name: %s\n", unicodeToString(data))

		if currentSlot == 0xFFFF {
			channels[len(channels)-1].name = unicodeToString(data)
		} else {
			mixer[currentInsert].slots[currentSlot].name = unicodeToString(data)
		}

	case GeneratorName:
		//fmt.Printf("Generator name: %s\n", unicodeToString(data))
		lastGenerator = unicodeToString(data)
	case TextPatName:
		//fmt.Printf("Pattern name: %s\n", unicodeToString(data))
	case TextURL:
		project.URL = unicodeToString(data)
	case TextStyle:
		project.style = unicodeToString(data)
	case TextAuthor:
		project.author = unicodeToString(data)
	case TextComment:
		project.comments = unicodeToString(data)
	case TextSampleFileName:
		//fmt.Printf("FileName: %s\n", unicodeToString(data))
		channels[len(channels)-1].fileName = unicodeToString(data)
	case TextInsertName:
		mixer[currentInsert].name = unicodeToString(data)

	default:
		//fmt.Printf("Text Unknown event (%d) - Size (%d) - %s\n", event, size, data)
	}

}

func parseChunk(file *os.File) {

	magic, _ := readNextBytes(file, 4)
	if string(magic) != "FLdt" {
		log.Fatal("Error reading data chunk")
	}
	size := readInt32(file)
	data, _ := readNextBytes(file, size)
	buffer := bytes.NewReader(data)

	count := 0
	for {

		event, err := readInt8(buffer)
		if err != nil {
			return
		}
		//fmt.Printf("(%#v) ", event)

		if event < Word {
			parseByte(event, buffer)
		} else if event < Int {
			parseWord(event, buffer)
		} else if event < Text {
			parseDWord(event, buffer)
		} else if event < Data {
			parseText(event, buffer)
		} else {
			parseData(event, buffer)
		}
		count++

	}

}

func isEmptyInsert(insert Insert) bool {

	for _, slot := range insert.slots {
		if slot.plugin != (Plugin{}) {
			return false
		}
	}
	return true
}

func prettyPrintMixer() {
	green := color.New(color.FgGreen)

	for i, insert := range mixer {
		if isEmptyInsert(insert) {
			continue
		}
		fmt.Printf("Insert %d ", i)
		green.Printf("[ %s]:\n", insert.name)
		for _, slot := range insert.slots {
			if slot.plugin != (Plugin{}) {
				fmt.Printf("\t%s %s %s\n", slot.plugin.name, slot.plugin.vendor, slot.plugin.filename)

			}
		}
	}

}

func prettyPrintChannels() {
	red := color.New(color.FgRed)
	green := color.New(color.FgGreen)

	for i, channel := range channels {

		fmt.Printf("Channel %d ", i)

		switch channel.ctype {
		case 0, 4:
			red.Printf("[ %s]: ", channel.name)
			fmt.Println(channel.fileName)

		case 5:
			fmt.Printf("[ %s] Automation channel\n", channel.name)

		case 2:
			green.Printf("[ %s]:\n", channel.name)
			fmt.Printf("\t%s\n\t%s\n\t%s\n", channel.plugin.name, channel.plugin.filename, channel.plugin.vendor)

		default:

		}

	}

}

func prettyPrintProject() {
	fmt.Printf("Title: %s\nTempo: %.3f\nVersion: %s\n\n", project.title, float32(project.tempo)/1000.0, project.version)
}

func main() {

	log.SetFlags(0)
	if len(os.Args) < 2 {
		fmt.Printf("FLP-READER\nUsage: %s <filename>\n", os.Args[0])
		os.Exit(1)
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal("Error while opening file", err)
	}

	defer file.Close()

	masterInsert := Insert{name: "Master"}
	mixer = append(mixer, masterInsert)

	fmt.Printf("Reading: %s\n", os.Args[1])
	parseHead(file)
	parseChunk(file)

	prettyPrintProject()
	prettyPrintChannels()
	fmt.Println("================ MIXER ================")
	prettyPrintMixer()

}
