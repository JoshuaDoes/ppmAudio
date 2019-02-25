package main

//ppmaudio -e audio.wav -o audio.adpcm
//ppmaudio -d flipnote.ppm -t bgm -o bgm.wav

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/JoshuaDoes/adpcm-go"
	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

type Offset struct {
	Offset uint32
	Length int
}
type SoundData struct {
	SoundMeta    SoundMeta
	BGM          []int // PCM audio
	SoundEffect1 []int // PCM audio
	SoundEffect2 []int // PCM audio
	SoundEffect3 []int // PCM audio
	Size         int
}
type SoundMeta struct {
	BGM          Offset
	SoundEffect1 Offset
	SoundEffect2 Offset
	SoundEffect3 Offset
}

var (
	channels       = 1
	sampleRate     = 8192
	sourceBitDepth = 2

	ppmMagic = []byte("PARA")
)

func main() {
	fmt.Println("> Initializing parameters")
	flagEncodeWav := flag.String("e", "", "The WAV file to encode")
	flagDecodePPM := flag.String("d", "", "The PPM to decode the audio of")
	flagTrack := flag.String("t", "", "The track to decode (bgm, se1, se2, se3)")
	flagOutputFile := flag.String("o", "", "The file to output to")

	fmt.Println("> Parsing parameters")
	flag.Parse()
	encodeWav := *flagEncodeWav
	decodePPM := *flagDecodePPM
	track := *flagTrack
	outputFile := *flagOutputFile

	fmt.Println("> Checking for invalid parameters")
	if encodeWav == "" {
		if decodePPM == "" || outputFile == "" {
			fmt.Println("Error: You must either specify the WAV file to encode to PPM ADPCM or the PPM file to decode the specified track of, along with the output file.")
			fmt.Println("Examples:")
			fmt.Println("> " + os.Args[0] + " -e audio.wav -o audio.adpcm")
			fmt.Println("> " + os.Args[0] + " -d flipnote.ppm -t bgm -o audio.wav")
			os.Exit(0)
		}
	}
	if encodeWav != "" && track != "" {
		fmt.Println("Error: You cannot specify a track to decode when trying to encode a WAV file to ADPCM.")
		fmt.Println("Examples:")
		fmt.Println("> " + os.Args[0] + " -e audio.wav -o audio.adpcm")
		fmt.Println("> " + os.Args[0] + " -d flipnote.ppm -t bgm -o audio.wav")
		os.Exit(0)
	}
	if encodeWav != "" && decodePPM != "" {
		fmt.Println("Error: You cannot encode a WAV file and decode a PPM audio track in the same command.")
		fmt.Println("Examples:")
		fmt.Println("> " + os.Args[0] + " -e audio.wav -o audio.adpcm")
		fmt.Println("> " + os.Args[0] + " -d flipnote.ppm -t bgm -o audio.wav")
		os.Exit(0)
	}
	if decodePPM != "" && track == "" {
		fmt.Println("Error: You must specify a PPM audio track to decode.")
		fmt.Println("Examples:")
		fmt.Println("> " + os.Args[0] + " -e audio.wav -o audio.adpcm")
		fmt.Println("> " + os.Args[0] + " -d flipnote.ppm -t bgm -o audio.wav")
		os.Exit(0)
	}
	if decodePPM != "" {
		if track != "bgm" && track != "se1" && track != "se2" && track != "se3" {
			fmt.Println("Error: Invalid PPM audio track.")
			fmt.Println("Available tracks: bgm | se1 | se2 | se3")
			os.Exit(0)
		}
	}

	if encodeWav != "" {
		fmt.Println("> Opening specified WAV file")
		wavFile, err := os.Open(encodeWav)
		if err != nil {
			fmt.Println("Error: Could not open the specified WAV file.")
			fmt.Println("Additional details: " + fmt.Sprintf("%v", err))
			os.Exit(0)
		}

		fmt.Println("> Initializing WAV decoder for specified WAV file")
		wavDecoder := wav.NewDecoder(wavFile)

		fmt.Println("> Getting PCM buffer of specified WAV file")
		audioBuffer, err := wavDecoder.FullPCMBuffer()
		if err != nil {
			fmt.Println("Error: Could not get the full PCM buffer of the specified WAV file.")
			fmt.Println("Additional details: " + fmt.Sprintf("%v", err))
			os.Exit(0)
		}

		fmt.Println("> Setting format of audio")
		audioBuffer.Format = &audio.Format{NumChannels: channels, SampleRate: sampleRate}
		audioBuffer.SourceBitDepth = sourceBitDepth

		fmt.Println("> Encoding PCM to nibble-flipped ADPCM")
		encodedAudio := encodeAudio(audioBuffer.Data)

		fmt.Println("> Creating output file")
		output, err := os.Create(outputFile)
		if err != nil {
			fmt.Println("Error: Could not create output file.")
			fmt.Println("Additional details: " + fmt.Sprintf("%v", err))
			os.Exit(0)
		}

		fmt.Println("> Writing encoded audio to output file")
		_, err = output.Write(encodedAudio)
		if err != nil {
			fmt.Println("Error: Could not write encoded ADPCM audio to output file.")
			fmt.Println("Additional details: " + fmt.Sprintf("%v", err))
			os.Exit(0)
		}
	}

	if decodePPM != "" {
		fmt.Println("> Opening specified PPM file")
		ppmFile, err := os.Open(decodePPM)
		if err != nil {
			fmt.Println("Error: Could not open specified PPM file.")
			fmt.Println("Additional details: " + fmt.Sprintf("%v", err))
			os.Exit(0)
		}

		fmt.Println("> Checking PPM magic")
		magic := make([]byte, 4)
		ppmFile.ReadAt(magic, 0x0)
		if !bytes.Equal(ppmMagic, magic) {
			fmt.Println("Error: PPM magic incorrect.")
			os.Exit(0)
		}

		fmt.Println("> Initializing sound data struct")
		soundData := &SoundData{}

		fmt.Println("> Reading animation size")
		animationSize := make([]byte, 4)
		ppmFile.ReadAt(animationSize, 0x4)
		frameDataSize := int(binaryReadLE_uint32(animationSize))

		fmt.Println("> Reading frame count")
		frameCountBytes := make([]byte, 2)
		ppmFile.ReadAt(frameCountBytes, 0xC)
		frameCount := int(binaryReadLE_uint16(frameCountBytes)) + 1

		fmt.Println("> Reading audio size")
		audioSize := make([]byte, 4)
		ppmFile.ReadAt(audioSize, 0x8)
		soundData.Size = hex2int(binaryReadLE(audioSize))

		fmt.Println("> Calculating sound header offset")
		soundHeaderOffset := 0x06A0 + frameDataSize + frameCount
		if (soundHeaderOffset % 4) != 0 {
			soundHeaderOffset += 4 - (soundHeaderOffset % 4)
		}

		fmt.Println("> Seeking to sound header offset")
		ppmFile.Seek(int64(soundHeaderOffset), 0)

		fmt.Println("> Reading size of BGM, sound effect 1, sound effect 2, and sound effect 3")
		bgmSizeBytes := make([]byte, 4)
		sec1SizeBytes := make([]byte, 4)
		sec2SizeBytes := make([]byte, 4)
		sec3SizeBytes := make([]byte, 4)
		ppmFile.Read(bgmSizeBytes)
		ppmFile.Read(sec1SizeBytes)
		ppmFile.Read(sec2SizeBytes)
		ppmFile.Read(sec3SizeBytes)
		bgmSize := binaryReadLE_uint32(bgmSizeBytes)
		sec1Size := binaryReadLE_uint32(sec1SizeBytes)
		sec2Size := binaryReadLE_uint32(sec2SizeBytes)
		sec3Size := binaryReadLE_uint32(sec3SizeBytes)

		soundHeaderOffset += 32
		fmt.Println("> Calculating BGM offset and length")
		soundData.SoundMeta.BGM.Offset = uint32(soundHeaderOffset)
		soundData.SoundMeta.BGM.Length = int(bgmSize)
		soundHeaderOffset += int(bgmSize)
		fmt.Println("> Calculating sound effect 1 offset and length")
		soundData.SoundMeta.SoundEffect1.Offset = uint32(soundHeaderOffset)
		soundData.SoundMeta.SoundEffect1.Length = int(sec1Size)
		fmt.Println("> Calculating sound effect 2 offset and length")
		soundHeaderOffset += int(sec1Size)
		soundData.SoundMeta.SoundEffect2.Offset = uint32(soundHeaderOffset)
		soundData.SoundMeta.SoundEffect2.Length = int(sec2Size)
		fmt.Println("> Calculating sound effect 3 offset and length")
		soundHeaderOffset += int(sec2Size)
		soundData.SoundMeta.SoundEffect3.Offset = uint32(soundHeaderOffset)
		soundData.SoundMeta.SoundEffect3.Length = int(sec3Size)

		fmt.Println("> Creating output file")
		output, err := os.Create(outputFile)
		if err != nil {
			fmt.Println("Error: Could not create output file.")
			fmt.Println("Additional details: " + fmt.Sprintf("%v", err))
			os.Exit(0)
		}

		switch track {
		case "bgm":
			fmt.Println("> Decoding BGM")
			soundData.BGM = decodeAudio(ppmFile, soundData.SoundMeta.BGM.Offset, soundData.SoundMeta.BGM.Length)

			fmt.Println("> Setting format of audio")
			audioBuffer := &audio.IntBuffer{Format: &audio.Format{NumChannels: channels, SampleRate: sampleRate}, Data: soundData.BGM, SourceBitDepth: sourceBitDepth}

			fmt.Println("> Initializing WAV encoder for BGM")
			wavEncoder := wav.NewEncoder(output, sampleRate, 16, 1, 1)

			fmt.Println("> Writing encoded audio to output file")
			err := wavEncoder.Write(audioBuffer.AsIntBuffer())
			if err != nil {
				fmt.Println("Error: Could not write WAV encoded audio.")
				fmt.Println("Additional details: " + fmt.Sprintf("%v", err))
				os.Exit(0)
			}
			err = wavEncoder.Close()
			if err != nil {
				fmt.Println("Error: Could not close WAV encoder.")
				fmt.Println("Additional details: " + fmt.Sprintf("%v", err))
			}
		case "se1":
			fmt.Println("> Decoding sound effect 1")
			soundData.SoundEffect1 = decodeAudio(ppmFile, soundData.SoundMeta.SoundEffect1.Offset, soundData.SoundMeta.SoundEffect1.Length)

			fmt.Println("> Setting format of audio")
			audioBuffer := &audio.IntBuffer{Format: &audio.Format{NumChannels: channels, SampleRate: sampleRate}, Data: soundData.SoundEffect1, SourceBitDepth: sourceBitDepth}

			fmt.Println("> Initializing WAV encoder for sound effect 1")
			wavEncoder := wav.NewEncoder(output, sampleRate, 16, 1, 1)

			fmt.Println("> Writing encoded audio to output file")
			err := wavEncoder.Write(audioBuffer.AsIntBuffer())
			if err != nil {
				fmt.Println("Error: Could not write WAV encoded audio.")
				fmt.Println("Additional details: " + fmt.Sprintf("%v", err))
				os.Exit(0)
			}
			err = wavEncoder.Close()
			if err != nil {
				fmt.Println("Error: Could not close WAV encoder.")
				fmt.Println("Additional details: " + fmt.Sprintf("%v", err))
			}
		case "se2":
			fmt.Println("> Decoding sound effect 2")
			soundData.SoundEffect2 = decodeAudio(ppmFile, soundData.SoundMeta.SoundEffect2.Offset, soundData.SoundMeta.SoundEffect2.Length)

			fmt.Println("> Setting format of audio")
			audioBuffer := &audio.IntBuffer{Format: &audio.Format{NumChannels: channels, SampleRate: sampleRate}, Data: soundData.SoundEffect2, SourceBitDepth: sourceBitDepth}

			fmt.Println("> Initializing WAV encoder for sound effect 2")
			wavEncoder := wav.NewEncoder(output, sampleRate, 16, 1, 1)

			fmt.Println("> Writing encoded audio to output file")
			err := wavEncoder.Write(audioBuffer.AsIntBuffer())
			if err != nil {
				fmt.Println("Error: Could not write WAV encoded audio.")
				fmt.Println("Additional details: " + fmt.Sprintf("%v", err))
				os.Exit(0)
			}
			err = wavEncoder.Close()
			if err != nil {
				fmt.Println("Error: Could not close WAV encoder.")
				fmt.Println("Additional details: " + fmt.Sprintf("%v", err))
			}
		case "se3":
			fmt.Println("> Decoding sound effect 3")
			soundData.SoundEffect3 = decodeAudio(ppmFile, soundData.SoundMeta.SoundEffect3.Offset, soundData.SoundMeta.SoundEffect3.Length)

			fmt.Println("> Setting format of audio")
			audioBuffer := &audio.IntBuffer{Format: &audio.Format{NumChannels: channels, SampleRate: sampleRate}, Data: soundData.SoundEffect3, SourceBitDepth: sourceBitDepth}

			fmt.Println("> Initializing WAV encoder for sound effect 3")
			wavEncoder := wav.NewEncoder(output, sampleRate, 16, 1, 1)

			fmt.Println("> Writing encoded audio to output file")
			err := wavEncoder.Write(audioBuffer.AsIntBuffer())
			if err != nil {
				fmt.Println("Error: Could not write WAV encoded audio.")
				fmt.Println("Additional details: " + fmt.Sprintf("%v", err))
				os.Exit(0)
			}
			err = wavEncoder.Close()
			if err != nil {
				fmt.Println("Error: Could not close WAV encoder.")
				fmt.Println("Additional details: " + fmt.Sprintf("%v", err))
			}
		}
	}

	fmt.Println("Done!")
}

func decodeAudio(ppmFile *os.File, trackOffset uint32, trackLength int) []int {
	ppmFile.Seek(int64(trackOffset), 0)

	buffer := make([]byte, trackLength)
	ppmFile.Read(buffer)
	for i := 0; i < trackLength; i++ {
		buffer[i] = (buffer[i]&0xF)<<4 | (buffer[i] >> 4) // Flipnote Studio's adpcm data uses reverse nibble order
	}
	audio := make([]int, 0)
	decoder := adpcm.NewDecoder(1)
	decoder.Decode(buffer, &audio)
	return audio
}

func encodeAudio(audio []int) []byte {
	encodedAudio := make([]byte, 0)
	adpcm.Encode(audio, &encodedAudio)
	for i := 0; i < len(encodedAudio); i++ {
		encodedAudio[i] = (encodedAudio[i]&0xF)<<4 | (encodedAudio[i] >> 4)
	}
	return encodedAudio
}

func binaryReadLE(byteArray []byte) []byte {
	for i, j := 0, len(byteArray)-1; i < j; i, j = i+1, j-1 {
		byteArray[i], byteArray[j] = byteArray[j], byteArray[i]
	}

	return byteArray
}
func binaryReadLE_uint16(byteArray []byte) uint16 {
	var result uint16
	buffer := bytes.NewReader(byteArray)
	binary.Read(buffer, binary.LittleEndian, &result)
	return result
}
func binaryReadLE_uint32(byteArray []byte) uint32 {
	var result uint32
	buffer := bytes.NewReader(byteArray)
	binary.Read(buffer, binary.LittleEndian, &result)
	return result
}
func hex2int(hexBytes []byte) int {
	hexStr := "0x" + hex.EncodeToString(hexBytes)
	result, _ := strconv.ParseInt(hexStr, 0, 64)
	return int(result)
}
