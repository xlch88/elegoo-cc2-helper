package main

// Fur Elise, WoO 59 - Ludwig van Beethoven (1810).
// Public-domain score: Breitkopf & Hartel (1888), typeset by Stelios Samelis.
// https://www.mutopiaproject.org/cgibin/piece-info.cgi?id=931
// MIDI: https://www.mutopiaproject.org/ftp/BeethovenLv/WoO59/fur_Elise_WoO59/fur_Elise_WoO59.mid
// Opening theme only, through the first ending; played once without repeats.
// Monophonic: highest right-hand note, left hand during right-hand rests;
// transposed one octave up. Tempo: 72 quarters/minute; duration: 10000 ms.
// Values are frequency in Hz (0 = rest) and duration in milliseconds.
var furElise = [...][2]int{
	{1319, 208}, {1245, 209}, {1319, 208}, {1245, 208}, {1319, 209}, {988, 208},
	{1175, 208}, {1047, 209}, {880, 416}, {440, 209}, {523, 208}, {659, 208},
	{880, 209}, {988, 416}, {415, 209}, {659, 208}, {831, 208}, {988, 209},
	{1047, 416}, {440, 209}, {659, 208}, {1319, 208}, {1245, 209}, {1319, 208},
	{1245, 208}, {1319, 209}, {988, 208}, {1175, 208}, {1047, 209}, {880, 416},
	{440, 209}, {523, 208}, {659, 208}, {880, 209}, {988, 416}, {415, 209},
	{659, 208}, {1047, 208}, {988, 209}, {880, 833},
}
