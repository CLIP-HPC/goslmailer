package main

var (
	ic_tc = []ic_test_case{
		{
			name:   "TestMissingOtherCmdArgs",
			defcon: "msteams",
			invocationContext: invocationContext{
				CmdParams{
					Subject: "Slurm subject line",
					Other:   []string{},
				},
				Receivers{},
			},
			want: Receivers{},
		},
		{
			name:   "TestEmptyStringArg",
			defcon: "msteams",
			invocationContext: invocationContext{
				CmdParams{
					Subject: "Slurm subject line",
					Other: []string{
						"",
					},
				},
				Receivers{},
			},
			want: Receivers{},
		},
		{
			name:   "TestLongSingleArg",
			defcon: "conX",
			invocationContext: invocationContext{
				CmdParams{
					Subject: "Slurm subject line",
					Other: []string{
						",conX,conX:pj,msteams:pja,:::xxx,,msteams::,petarj,matrix:!channelid:server.org,:xxx",
					},
				},
				Receivers{},
			},
			want: Receivers{
				{
					scheme: "conX",
					target: "conX",
				},
				{
					scheme: "conX",
					target: "pj",
				},
				{
					scheme: "msteams",
					target: "pja",
				},
				{
					scheme: "conX",
					target: "petarj",
				},
				{
					scheme: "matrix",
					target: "!channelid:server.org",
				},
			},
		},
		{
			name:   "TestMultipleArgs",
			defcon: "mailto",
			invocationContext: invocationContext{
				CmdParams{
					Subject: "Slurm subject line",
					Other: []string{
						"",
						"msteams:::",
						"mailto:pja@bla.bla,,,",
						"pja@bla.bla",
						"msteams:pja",
						":::pja",
						"matrix:!channelid:server.org",
						"::pja",
					},
				},
				Receivers{},
			},
			want: Receivers{
				{
					scheme: "mailto",
					target: "pja@bla.bla",
				},
				{
					scheme: "mailto",
					target: "pja@bla.bla",
				},
				{
					scheme: "msteams",
					target: "pja",
				},
				{
					scheme: "matrix",
					target: "!channelid:server.org",
				},
			},
		},
	}
)
