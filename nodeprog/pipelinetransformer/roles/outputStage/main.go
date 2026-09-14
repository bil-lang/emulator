//go:build js && wasm

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:76
package main

import "sync"

import "emulator/nodeprog/bilink"

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:77

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:78
import (

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:79
	"math"

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:80
	"time"
	//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:81
)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:82

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:83
const generations = 6

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:84
const vocabSize = 4

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:85
const dModel = 4

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:86
const dFF = 8

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:87
const stageRows = 4 // embed, attention, feed-forward, output

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:88

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:89
// hopDelay is purely cosmetic -- the rendezvous itself already forces

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:90
// correct pacing -- it just makes each hop visible on the grid.

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:91
const hopDelay = 60 * time.Millisecond

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:92

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:93
var vocab = [vocabSize]string{"the", "cat", "sat", "mat"}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:94

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:95
// embedding, wq/wk/wv, w1/b1/w2/b2 and wo are the same fixed,

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:96
// hand-picked weights Example 21 used -- illustrative numbers for a

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:97
// from-scratch demo, not a trained model. The point of this example

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:98
// is the distributed dataflow architecture, not prediction quality.

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:99
var embedding = [vocabSize][dModel]float64{

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:100
	{0.10, 0.20, 0.30, 0.40},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:101
	{0.40, 0.10, 0.20, 0.30},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:102
	{0.30, 0.40, 0.10, 0.20},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:103
	{0.20, 0.30, 0.40, 0.10},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:104
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:105

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:106
var wq = [dModel][dModel]float64{

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:107
	{1, 0, 0, 1},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:108
	{0, 1, 1, 0},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:109
	{1, 1, 0, 0},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:110
	{0, 0, 1, 1},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:111
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:112

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:113
var wk = [dModel][dModel]float64{

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:114
	{1, 1, 0, 0},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:115
	{0, 0, 1, 1},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:116
	{1, 0, 1, 0},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:117
	{0, 1, 0, 1},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:118
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:119

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:120
var wv = [dModel][dModel]float64{

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:121
	{1, 0, 1, 0},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:122
	{0, 1, 0, 1},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:123
	{1, 1, 1, 1},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:124
	{0, 0, 0, 1},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:125
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:126

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:127
var w1 = [dModel][dFF]float64{

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:128
	{0.2, -0.1, 0.3, 0.0, -0.2, 0.1, 0.0, 0.3},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:129
	{-0.1, 0.2, 0.0, 0.3, 0.1, -0.2, 0.3, 0.0},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:130
	{0.3, 0.0, -0.1, 0.2, 0.0, 0.3, -0.2, 0.1},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:131
	{0.0, 0.3, 0.2, -0.1, 0.3, 0.0, 0.1, -0.2},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:132
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:133
var b1 = [dFF]float64{0.05, -0.05, 0.05, -0.05, 0.05, -0.05, 0.05, -0.05}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:134

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:135
var w2 = [dFF][dModel]float64{

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:136
	{0.2, -0.1, 0.1, 0.0},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:137
	{-0.1, 0.2, 0.0, 0.1},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:138
	{0.1, 0.0, 0.2, -0.1},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:139
	{0.0, 0.1, -0.1, 0.2},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:140
	{0.2, 0.0, -0.1, 0.1},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:141
	{0.0, 0.2, 0.1, -0.1},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:142
	{-0.1, 0.1, 0.0, 0.2},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:143
	{0.1, -0.1, 0.2, 0.0},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:144
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:145
var b2 = [dModel]float64{0, 0, 0, 0}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:146

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:147
var wo = [dModel][vocabSize]float64{

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:148
	{0.5, -0.2, 0.1, 0.0},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:149
	{0.0, 0.5, -0.2, 0.1},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:150
	{0.1, 0.0, 0.5, -0.2},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:151
	{-0.2, 0.1, 0.0, 0.5},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:152
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:153

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:154
func wordAt(tok int) string { return vocab[tok%vocabSize] }

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:155

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:156
func posEncode(pos int) [dModel]float64 {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:157
	var pe [dModel]float64

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:158
	for i := 0; i < dModel; i += 2 {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:159
		freq := 1.0 / math.Pow(10000, float64(i)/float64(dModel))

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:160
		pe[i] = math.Sin(float64(pos) * freq)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:161
		if i+1 < dModel {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:162
			pe[i+1] = math.Cos(float64(pos) * freq)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:163
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:164
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:165
	return pe

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:166
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:167

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:168
func add4(a, b [dModel]float64) [dModel]float64 {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:169
	var y [dModel]float64

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:170
	for i := range dModel {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:171
		y[i] = a[i] + b[i]

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:172
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:173
	return y

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:174
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:175

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:176
func matVec4(w [dModel][dModel]float64, x [dModel]float64) [dModel]float64 {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:177
	var y [dModel]float64

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:178
	for i := range dModel {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:179
		sum := 0.0

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:180
		for j := range dModel {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:181
			sum += w[i][j] * x[j]

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:182
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:183
		y[i] = sum

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:184
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:185
	return y

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:186
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:187

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:188
func dot4(a, b [dModel]float64) float64 {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:189
	sum := 0.0

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:190
	for i := range dModel {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:191
		sum += a[i] * b[i]

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:192
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:193
	return sum

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:194
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:195

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:196
func softmax(scores []float64) []float64 {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:197
	max := scores[0]

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:198
	for _, s := range scores {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:199
		if s > max {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:200
			max = s

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:201
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:202
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:203
	out := make([]float64, len(scores))

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:204
	sum := 0.0

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:205
	for i, s := range scores {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:206
		out[i] = math.Exp(s - max)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:207
		sum += out[i]

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:208
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:209
	for i := range out {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:210
		out[i] /= sum

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:211
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:212
	return out

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:213
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:214

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:215
func relu(x float64) float64 {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:216
	if x > 0 {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:217
		return x

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:218
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:219
	return 0

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:220
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:221

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:222
func feedForward(x [dModel]float64) [dModel]float64 {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:223
	var hidden [dFF]float64

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:224
	for j := range dFF {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:225
		sum := b1[j]

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:226
		for i := range dModel {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:227
			sum += x[i] * w1[i][j]

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:228
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:229
		hidden[j] = relu(sum)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:230
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:231
	var out [dModel]float64

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:232
	for k := range dModel {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:233
		sum := b2[k]

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:234
		for j := range dFF {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:235
			sum += hidden[j] * w2[j][k]

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:236
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:237
		out[k] = sum

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:238
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:239
	return out

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:240
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:241

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:242
func outputLogits(x [dModel]float64) [vocabSize]float64 {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:243
	var logits [vocabSize]float64

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:244
	for k := range vocabSize {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:245
		sum := 0.0

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:246
		for i := range dModel {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:247
			sum += x[i] * wo[i][k]

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:248
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:249
		logits[k] = sum

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:250
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:251
	return logits

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:252
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:253

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:254
func argmax4(v [vocabSize]float64) int {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:255
	best := 0

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:256
	for i := 1; i < vocabSize; i++ {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:257
		if v[i] > v[best] {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:258
			best = i

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:259
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:260
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:261
	return best

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:262
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:263

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:264
// runAttention runs self-attention plus the first residual connection

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:265
// over a whole row's worth of already-embedded vectors -- exactly the

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:266
// attention half of Example 21's runTransformer, stopping short of

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:267
// the feed-forward/output steps, which happen two and three rows

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:268
// further south instead.

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:269
func runAttention(xs [][dModel]float64) [][dModel]float64 {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:270
	cols := len(xs)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:271

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:272
	q := make([][dModel]float64, cols)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:273
	k := make([][dModel]float64, cols)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:274
	v := make([][dModel]float64, cols)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:275
	for c := range cols {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:276
		q[c] = matVec4(wq, xs[c])

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:277
		k[c] = matVec4(wk, xs[c])

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:278
		v[c] = matVec4(wv, xs[c])

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:279
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:280

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:281
	resid1 := make([][dModel]float64, cols)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:282
	for c := range cols {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:283
		scores := make([]float64, cols)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:284
		for j := range cols {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:285
			scores[j] = dot4(q[c], k[j]) / math.Sqrt(float64(dModel))

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:286
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:287
		weights := softmax(scores)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:288

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:289
		var attnOut [dModel]float64

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:290
		for j := range cols {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:291
			for i := range dModel {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:292
				attnOut[i] += weights[j] * v[j][i]

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:293
			}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:294
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:295

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:296
		resid1[c] = add4(xs[c], attnOut)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:297
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:298

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:299
	return resid1

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:300
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:301

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:302
// f32bits/f32val convert one [dModel]float64 vector's component to and

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:303
// from the raw int32 bit pattern that actually crosses a physical

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:304
// link -- the link fabric only carries a plain int32 per hop (see

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:305
// bilink.Send/Recv), the same "one int32 at a time" idiom Example 21

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:306
// used for scalar tokens. Pure bit-twiddling, no channel involved, so

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:307
// -- unlike a send/receive loop itself -- these are safe to factor out

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:308
// into ordinary top-level funcs.

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:309
func f32bits(x float64) int32 { return int32(math.Float32bits(float32(x))) }

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:310
func f32val(bits int32) float64 { return float64(math.Float32frombits(uint32(bits))) }

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:311

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:312
// embedStage runs on every column of row 0: computes this

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:313
// generation's synthetic input token locally (the same

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:314
// (col+gen)%vocabSize convention Example 21's relay/rowEnd used),

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:315
// embeds it, adds its positional encoding, and hands the result south

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:316
// to row 1. Pure per-column work, no lateral communication at all --

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:317
// every column of row 0 runs this in parallel with every other

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:318
// column, unlike Example 21 where one processor did every column's

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:319
// embedding serially. south is the only link this role ever touches.

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:320
//

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:321
// sendSouth is a closure, not a top-level func: `place`-bound channel

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:322
// parameters are resolved by bilc at compile time, purely from their

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:323
// literal name appearing directly inside the *placed-called proc's own

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:324
// body* -- a channel handed to a separately-compiled helper function is

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:325
// just a nil value at runtime (nothing ever fills it in), so every

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:326
// proc below defines its own tiny send/recv closures over its own

// channel parameters instead of sharing one.
//
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:327
func
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:328
embedStage(south chan<- int32) {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:329
	sendSouth := func(v [dModel]float64) {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:330
		for i := range dModel {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:331
			bilink.Send(2, f32bits(v[i]))
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:331

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:332
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:333
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:334

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:335
	c := bilink.Col()

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:336
	bilink.Screenf("embed ready at col %d", c)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:337

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:338
	for gen := range generations {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:339
		tok := (c + gen) % vocabSize

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:340
		x0 := add4(embedding[tok], posEncode(c))

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:341
		time.Sleep(hopDelay)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:342
		sendSouth(x0)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:343
		bilink.Screenf("gen %d: embedded %q", gen, wordAt(tok))

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:344
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:345

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:346
	bilink.Screenf("embed done")

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:347
	time.Sleep(1<<63 - 1)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:347

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:348
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:349

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:350
// attnController runs at (1,0): the west end of the attention row.

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:351
// It gets its own column's embedded vector straight from the north

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:352
// link, gathers every other column's vector via the row's own

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:353
// west/east relay chain (Example 21's relay idiom, generalized from

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:354
// one int32 per round to a whole [dModel]float64 per round), runs

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:355
// self-attention plus the first residual connection once the whole

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:356
// row has arrived, scatters each column's own result back out along

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:357
// the same relay chain, then forwards its own result south to row

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:358
// 2 -- exactly what every other attention-row column does with

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:359
// theirs. East is used both ways (gather in, scatter out), so it

// needs two directional parameters bound to the same physical link.
//
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:360
func
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:361
attnController(north <-chan int32, eastIn <-chan int32, eastOut chan<- int32, south chan<- int32) {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:362
	recvNorth := func() [dModel]float64 {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:363
		var v [dModel]float64

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:364
		for i := range dModel {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:365
			var bits int32

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:366
			bits = bilink.Recv(0)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:366

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:367
			v[i] = f32val(bits)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:368
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:369
		return v

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:370
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:371
	recvEast := func() [dModel]float64 {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:372
		var v [dModel]float64

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:373
		for i := range dModel {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:374
			var bits int32

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:375
			bits = bilink.Recv(1)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:375

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:376
			v[i] = f32val(bits)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:377
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:378
		return v

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:379
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:380
	sendEast := func(v [dModel]float64) {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:381
		for i := range dModel {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:382
			bilink.Send(1, f32bits(v[i]))
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:382

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:383
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:384
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:385
	sendSouth := func(v [dModel]float64) {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:386
		for i := range dModel {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:387
			bilink.Send(2, f32bits(v[i]))
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:387

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:388
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:389
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:390

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:391
	cols := bilink.NumCols()

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:392
	bilink.Screenf("attn controller ready, %d cols", cols)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:393

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:394
	for gen := range generations {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:395
		xs := make([][dModel]float64, cols)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:396
		xs[0] = recvNorth()

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:397

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:398
		for r := cols - 1; r >= 1; r-- {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:399
			time.Sleep(hopDelay)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:400
			xs[r] = recvEast()

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:401
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:402

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:403
		resid1 := runAttention(xs)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:404

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:405
		for s := 1; s < cols; s++ {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:406
			time.Sleep(hopDelay)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:407
			sendEast(resid1[s])

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:408
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:409

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:410
		time.Sleep(hopDelay)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:411
		sendSouth(resid1[0])

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:412
		bilink.Screenf("gen %d: attended", gen)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:413
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:414

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:415
	bilink.Screenf("attn controller done")

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:416
	time.Sleep(1<<63 - 1)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:416

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:417
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:418

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:419
// attnRelay runs on every other attention-row column. For gather

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:420
// round r: if r is its own column it originates (its own vector,

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:421
// already in hand from row 0); if r is further east it forwards

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:422
// (receive east, forward west); otherwise that round isn't on its

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:423
// path. Scatter mirrors this going back east. Once it has its own

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:424
// final result it forwards that south to row 2. This is Example 21's

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:425
// relay proc, generalized from one int32 per round to a whole

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:426
// [dModel]float64 per round. Both east and west are used bidirectionally,

// so each gets two directional parameters bound to the same link.
//
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:427
func
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:428
attnRelay(north <-chan int32, eastIn <-chan int32, westOut chan<- int32, westIn <-chan int32, eastOut chan<- int32, south chan<- int32) {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:429
	recvNorth := func() [dModel]float64 {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:430
		var v [dModel]float64

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:431
		for i := range dModel {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:432
			var bits int32

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:433
			bits = bilink.Recv(0)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:433

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:434
			v[i] = f32val(bits)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:435
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:436
		return v

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:437
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:438
	recvEast := func() [dModel]float64 {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:439
		var v [dModel]float64

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:440
		for i := range dModel {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:441
			var bits int32

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:442
			bits = bilink.Recv(1)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:442

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:443
			v[i] = f32val(bits)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:444
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:445
		return v

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:446
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:447
	recvWest := func() [dModel]float64 {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:448
		var v [dModel]float64

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:449
		for i := range dModel {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:450
			var bits int32

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:451
			bits = bilink.Recv(3)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:451

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:452
			v[i] = f32val(bits)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:453
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:454
		return v

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:455
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:456
	sendWest := func(v [dModel]float64) {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:457
		for i := range dModel {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:458
			bilink.Send(3, f32bits(v[i]))
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:458

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:459
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:460
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:461
	sendEast := func(v [dModel]float64) {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:462
		for i := range dModel {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:463
			bilink.Send(1, f32bits(v[i]))
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:463

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:464
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:465
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:466
	sendSouth := func(v [dModel]float64) {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:467
		for i := range dModel {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:468
			bilink.Send(2, f32bits(v[i]))
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:468

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:469
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:470
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:471

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:472
	c := bilink.Col()

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:473
	cols := bilink.NumCols()

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:474
	bilink.Screenf("attn relay ready at col %d", c)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:475

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:476
	for gen := range generations {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:477
		myVec := recvNorth()

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:478

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:479
		for r := cols - 1; r >= 1; r-- {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:480
			switch {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:481
			case r == c:

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:482
				time.Sleep(hopDelay)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:483
				sendWest(myVec)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:484
			case r > c:

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:485
				v := recvEast()

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:486
				time.Sleep(hopDelay)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:487
				sendWest(v)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:488
			}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:489
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:490

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:491
		var mine [dModel]float64

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:492
		for s := 1; s < cols; s++ {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:493
			switch {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:494
			case s == c:

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:495
				mine = recvWest()

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:496
			case s > c:

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:497
				v := recvWest()

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:498
				time.Sleep(hopDelay)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:499
				sendEast(v)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:500
			}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:501
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:502

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:503
		time.Sleep(hopDelay)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:504
		sendSouth(mine)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:505
		bilink.Screenf("gen %d: relayed", gen)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:506
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:507

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:508
	bilink.Screenf("attn relay done")

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:509
	time.Sleep(1<<63 - 1)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:509

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:510
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:511

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:512
// attnRowEnd runs at (1,cols-1): the east end of the attention row.

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:513
// Source for exactly one gather round (its own column) and

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:514
// destination for exactly one scatter round, then forwards its own

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:515
// result south -- unlike attnRelay it needs no round loop at all. West

// is used both ways, so it gets two directional parameters.
//
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:516
func
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:517
attnRowEnd(north <-chan int32, westOut chan<- int32, westIn <-chan int32, south chan<- int32) {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:518
	recvNorth := func() [dModel]float64 {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:519
		var v [dModel]float64

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:520
		for i := range dModel {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:521
			var bits int32

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:522
			bits = bilink.Recv(0)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:522

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:523
			v[i] = f32val(bits)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:524
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:525
		return v

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:526
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:527
	recvWest := func() [dModel]float64 {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:528
		var v [dModel]float64

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:529
		for i := range dModel {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:530
			var bits int32

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:531
			bits = bilink.Recv(3)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:531

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:532
			v[i] = f32val(bits)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:533
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:534
		return v

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:535
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:536
	sendWest := func(v [dModel]float64) {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:537
		for i := range dModel {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:538
			bilink.Send(3, f32bits(v[i]))
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:538

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:539
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:540
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:541
	sendSouth := func(v [dModel]float64) {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:542
		for i := range dModel {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:543
			bilink.Send(2, f32bits(v[i]))
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:543

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:544
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:545
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:546

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:547
	bilink.Screenf("attn row-end ready at col %d", bilink.Col())

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:548

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:549
	for gen := range generations {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:550
		myVec := recvNorth()

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:551
		time.Sleep(hopDelay)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:552
		sendWest(myVec)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:553

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:554
		mine := recvWest()

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:555
		time.Sleep(hopDelay)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:556
		sendSouth(mine)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:557
		bilink.Screenf("gen %d: row-end relayed", gen)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:558
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:559

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:560
	bilink.Screenf("attn row-end done")

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:561
	time.Sleep(1<<63 - 1)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:561

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:562
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:563

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:564
// ffnStage runs on every column of row 2: purely per-column work

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:565
// again, like embedStage -- receive the attention stage's residual

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:566
// output from north, run the feed-forward sub-layer plus its own

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:567
// residual connection, hand the result south to row 3. No lateral

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:568
// communication: every column of row 2 runs this fully in parallel

// with every other column.
//
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:569
func
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:570
ffnStage(north <-chan int32, south chan<- int32) {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:571
	recvNorth := func() [dModel]float64 {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:572
		var v [dModel]float64

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:573
		for i := range dModel {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:574
			var bits int32

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:575
			bits = bilink.Recv(0)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:575

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:576
			v[i] = f32val(bits)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:577
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:578
		return v

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:579
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:580
	sendSouth := func(v [dModel]float64) {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:581
		for i := range dModel {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:582
			bilink.Send(2, f32bits(v[i]))
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:582

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:583
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:584
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:585

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:586
	bilink.Screenf("ffn ready at col %d", bilink.Col())

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:587

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:588
	for range generations {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:589
		resid1 := recvNorth()

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:590
		resid2 := add4(resid1, feedForward(resid1))

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:591
		time.Sleep(hopDelay)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:592
		sendSouth(resid2)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:593
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:594

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:595
	bilink.Screenf("ffn done")

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:596
	time.Sleep(1<<63 - 1)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:596

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:597
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:598

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:599
// outputStage runs on every column of row 3, the pipeline's final

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:600
// stage: receive the feed-forward stage's output from north, project

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:601
// to vocabulary logits, argmax to a predicted token, done. The

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:602
// prediction never needs to travel anywhere else, since embedStage

// already knows (and logged) what token it started from.
//
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:603
func
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:604
outputStage(north <-chan int32) {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:605
	recvNorth := func() [dModel]float64 {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:606
		var v [dModel]float64

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:607
		for i := range dModel {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:608
			var bits int32

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:609
			bits = bilink.Recv(0)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:609

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:610
			v[i] = f32val(bits)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:611
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:612
		return v

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:613
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:614

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:615
	c := bilink.Col()

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:616
	bilink.Screenf("output ready at col %d", c)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:617

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:618
	for gen := range generations {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:619
		resid2 := recvNorth()

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:620
		predicted := argmax4(outputLogits(resid2))

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:621
		bilink.Screenf("gen %d: col %d -> %q", gen, c, wordAt(predicted))

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:622
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:623

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:624
	bilink.Screenf("output done")

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:625
	time.Sleep(1<<63 - 1)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:625

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:626
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:627

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:628
// idle runs on every node nothing more specific matched -- every row

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:629
// at or beyond stageRows, since this demo's pipeline is only 4 rows

// deep.
//
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:630
func
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:631
idle() {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:632
	bilink.Screenf("idle -- beyond the pipeline's %d rows", stageRows)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:633
	time.Sleep(1<<63 - 1)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:633

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:634
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:635

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:636
func main() {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:637
	cols := bilink.NumCols()

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:638

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:639
	_ = (1)
	_ = (0)
	_ = (1)
	_ = (cols - 1)
	_ = (0)
	_ = (1)
	_ = (2)
	_ = (3)
	outputStage(nil)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:679

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/22-pipeline-transformer.bil:680
}

func par(branches ...func()) {
	var wg sync.WaitGroup
	wg.Add(len(branches))
	for _, b := range branches {
		go func(b func()) {
			defer wg.Done()
			b()
		}(b)
	}
	wg.Wait()
}

func parFor(n int, body func(int)) {
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			body(i)
		}(i)
	}
	wg.Wait()
}

func makeChans[T any](n int) []chan T {
	cs := make([]chan T, n)
	for i := range cs {
		cs[i] = make(chan T)
	}
	return cs
}

func splitN[T any](s []T, n int) [][]T {
	chunk := len(s) / n
	out := make([][]T, n)
	for i := range n {
		lo := i * chunk
		hi := lo + chunk
		if i == n-1 {
			hi = len(s)
		}
		out[i] = s[lo:hi:hi]
	}
	return out
}
