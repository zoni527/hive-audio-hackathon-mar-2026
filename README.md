# Hive audio hachathon March 2026: Sound visualizer
This project was coded during one weekend.
## Description
This audio visualizer consists of four different visualizers that
hook into the computer's default audio input through portaudio.
The top visualizer is a loudness graph that represents the maximum
intensity in a sample over time, the bottom is a fast fourier
transform frequency plot (0 - ~10000 Hz), and the middle two both
are plots of the current sample, one as a vertical bar graph, the
other as a circular plot with sample intensity offsetting the
distance from the center of the circle.
## Dependencies & requirements
- Linux (developed on Ubuntu 24.04.1)
- Go (developed with 1.25.5)
- Portaudio (19.6.0-1.2)
## Installation & running
To run without producing a named executable:
```bash
go run .
```
To get the default named executable and run it:
```bash
go build
./visualizer
```
Custom named executable:
```bash
go build -o <custom_name>
./<custom_name>
```
Ensure that your default sound input device is correctly set up
and is not muted.

For routing different sound sources to the default audio input I
can recommend [Helvum](https://github.com/knyipab/helvum) which
in turn was recommended to me by a peer who also took part in the
audio hackathon.
## AI use
Google Gemini was used to help with the correct scaling of the
FFT graph, verifying logic, helping with dependency management,
compilation, and hunting for bugs.
