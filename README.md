# ppmAudio

### A PPM audio extractor and a PPM-ready ADPCM audio encoder, all-in-one

----

## Getting and using ppmAudio

At this moment in time, releases will not be uploaded. You must build it yourself for now.

## What does it do?

When you run ppmAudio with appropriate parameters and files, it can do one of two things:
1. Extract either the bgm, se1, se2, or se3 tracks of a given PPM file to WAV
2. Encode a WAV file as PPM-ready ADPCM to manually inject into a PPM and sign yourself (don't forget to modify offsets and lengths accordingly!)

----

## Building it yourself

In order to build ppmAudio locally, you must have already installed a working
Golang environment on your development system and installed the package
dependencies that ppmAudio relies on to function properly.

ppmAudio is currently built using Golang `1.10.2`.

### Dependencies

| Package Name |
| ------------ |
| [adpcm](https://github.com/bovarysme/adpcm) |
| [audio](https://github.com/go-audio/audio) |
| [wav](https://github.com/go-audio/wav) |

### Building

Simply run `go build` in this repo's directory once all dependencies are satisfied.

### Running ppmAudio

Finally, to run ppmAudio, you will need to use a terminal/shell or a command prompt
and give it some parameters to do what it needs to do. You must provide a PPM file
if you wish to export a track to WAV and you must provide a WAV file if you wish to
encode it to PPM-ready ADPCM.

| Parameters | Usage |
| ---------- | ----- |
| `-e` | Points to the WAV file to encode as PPM-ready ADPCM. Cannot be used with `-d` or `-t`, and requires an output file. |
| `-d` | Points to the PPM file to encode a track of to WAV. Cannot be used with `-e`, and requires both a track and an output file. |
| `-t` | Specifies the track to encode as WAV in conjunction with the specified PPM. Cannot be used with `-e`, and requires both a PPM file and an output file. |
| `-o` | Points to the new file to output the resulting audio to. Required to be used with either `-e` and `-d`. |

Examples:
- `./ppmAudio -e audio.wav -o bgm.adpcm`
- `./ppmAudio -d flipnote.ppm -t bgm -o bgm.wav`
- `./ppmAudio -d flipnote.ppm -t se1 -o soundeffect1.wav`
- `./ppmAudio -d flipnote.ppm -t se2 -o soundeffect2.wav`
- `./ppmAudio -d flipnote.ppm -t se3 -o soundeffect3.wav`

### Contributing notes

When pushing to your repo or submitting pull requests to this repo, it is highly
advised that you clean up the working directory to only contain `LICENSE`, `main.go`,
`README.md`, and the `.git` folder. A proper `.gitignore` will be written soon to
mitigate this requirement.

----

## Support
For help and support with ppmAudio, create an issue on the issues page. If you do not have a GitHub account, send me a message on Discord (@JoshuaDoes#1685) or join the [Kaeru Network Discord](https://discord.me/kaeru).

## License
The source code for ppmAudio is released under the MIT License. See LICENSE for more details.

## Donations
All donations are highly appreciated. They help motivate me to continue working on side projects like these, especially when it comes to something you may really want added!

[![Donate](https://img.shields.io/badge/Donate-PayPal-green.svg)](https://paypal.me/JoshuaDoes)
