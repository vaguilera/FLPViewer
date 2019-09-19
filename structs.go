package main

type header struct {
	Magic    uint32
	Length   uint32
	_        uint16
	Channels uint16
	PPQ      uint16
}

type Project struct {
	title    string
	tempo    uint32
	version  string
	style    string
	author   string
	comments string
	URL      string
}

type Channel struct {
	id       uint16
	name     string
	sample   string
	ctype    uint8
	fileName string
	plugin   Plugin
}

type Insert struct {
	name  string
	slots [10]Slot
}
type Slot struct {
	id     uint16
	name   string
	plugin Plugin
}

type Plugin struct {
	pluginType string
	name       string
	filename   string
	vendor     string
}

var channelType = map[uint8]string{
	0: "Sample",
	4: "Sample",
	5: "Automation",
	2: "VST",
}

var VSTType = map[uint8]string{
	8: "VST External (8)",
	9: "VST External (9)",
	3: "VST Internal",
}

const (
	PluginMidi       = 1
	PluginFlags      = 2
	PluginIo         = 30
	PluginInputInfo  = 31
	PluginOutputInfo = 32
	PluginPluginInfo = 50
	PluginVstPlugin  = 51
	PluginGuid       = 52
	PluginState      = 53
	PluginName       = 54
	PluginFilename   = 55
	PluginVendorName = 56
)

const (
	ByteChanType = 21

	Undef              = 192
	Text               = 192
	TextChanName       = 192
	TextPatName        = 193
	TextTitle          = 194
	TextComment        = 195
	TextSampleFileName = 196
	TextURL            = 197
	TextCommentRtf     = 198
	TextVersion        = 199
	GeneratorName      = 201
	TextComments       = 202
	TextPluginName     = 203
	TextInsertName     = 204
	TextMidiCtrls      = 208
	TextDelay          = 209
	TextStyle          = 206
	TextAuthor         = 207

	DWordFineTempo = 156

	Word               = 64
	WordNewChan        = 64
	WordCurrentSlotNum = 98
	Int                = 128
	Data               = 210

	DataPatternNotes = 224
	DataNewPlugin    = 212
	DataPluginParams = 213
	DataInsertRoutes = 235
)
