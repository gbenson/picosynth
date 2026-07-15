package picosynth

var ParameterGroups = []ParameterGroup{
	// ParameterGroup{
	// 	Name: "global/voice",
	// 	Hotkey: ButtonSong, // the red one
	// 	Parameters: []ParameterSpec{
	// 		RegisterParameter("pitch bend"),
	// 		RegisterParameter("glide"),
	// 	},
	// },
	ParameterGroup{
		Name:   "osc1",
		Hotkey: ButtonRhythm, // first yellow
		Parameters: []ParameterSpec{
			RegisterParameter(Osc1Pitch), // or, only ±octave for osc1?
			//RegisterParameter(Osc1Shaper),
			RegisterParameter(Osc1Shape),
			RegisterParameter(Osc1Level),
		},
	},
	ParameterGroup{
		Name:   "osc2",
		Hotkey: ButtonAccomp, // second yellow
		Parameters: []ParameterSpec{
			RegisterParameter(Osc2Pitch), // more detune options for osc2
			//OctaveParameter(Osc2Pitch),
			//SemitoneParameter(Osc2Pitch), => or on long press menu?
			//RegisterParameter(Osc2Shaper),
			RegisterParameter(Osc2Shape),
			RegisterParameter(Osc2Level),
		},
	},
	ParameterGroup{
		Name:   "filt1",
		Hotkey: ButtonFunny, // third yellow
		Parameters: []ParameterSpec{
			RegisterParameter(Filt1Mode),
			//RegisterParameter(Filt1KeyTrack), // a matrix cell?
			//RegisterParameter(Filt1EnvDepth), // a matrix cell?
		},
	},
}
