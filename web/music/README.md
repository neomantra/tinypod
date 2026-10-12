# music

Put WAV files here to build them into the web page. Each folder is an album, as on the device. With no WAV files the page plays a few demo tones.

```
music/Album One/01 Song.wav
music/Single.wav
```

Keep the files small. The browser downloads all of them with the page. 22.05 kHz mono IMA ADPCM uses about 11 KB per second.

```
ffmpeg -i in.mp3 -ac 1 -ar 22050 -c:a adpcm_ima_wav out.wav
```
