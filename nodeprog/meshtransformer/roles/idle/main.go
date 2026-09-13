//go:build js && wasm

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:43
package main

import "sync"

import "emulator/nodeprog/bilink"

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:44

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:45
import (

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:46
	"math"

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:47
	"time"
	//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:48
)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:49

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:50
const generations = 2

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:51
const vocabSize = 4

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:52
const dModel = 4

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:53
const dFF = 8

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:54

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:55
// hopDelay is purely cosmetic -- the rendezvous itself already forces

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:56
// correct pacing -- it just makes each hop visible on the grid.

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:57
const hopDelay = 80 * time.Millisecond

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:58

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:59
var vocab = [vocabSize]string{"the", "cat", "sat", "mat"}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:60

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:61
// embedding, wq/wk/wv, w1/b1/w2/b2 and wo are fixed, hand-picked

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:62
// numbers -- illustrative weights for a from-scratch demo, not a

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:63
// trained model. The point of this example is the architecture

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:64
// (embed -> attend -> residual -> feed-forward -> residual -> project)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:65
// and the fact that it runs for real, distributed across the mesh's

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:66
// own link fabric, not the quality of its predictions.

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:67
var embedding = [vocabSize][dModel]float64{

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:68
	{0.10, 0.20, 0.30, 0.40},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:69
	{0.40, 0.10, 0.20, 0.30},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:70
	{0.30, 0.40, 0.10, 0.20},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:71
	{0.20, 0.30, 0.40, 0.10},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:72
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:73

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:74
var wq = [dModel][dModel]float64{

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:75
	{1, 0, 0, 1},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:76
	{0, 1, 1, 0},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:77
	{1, 1, 0, 0},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:78
	{0, 0, 1, 1},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:79
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:80

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:81
var wk = [dModel][dModel]float64{

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:82
	{1, 1, 0, 0},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:83
	{0, 0, 1, 1},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:84
	{1, 0, 1, 0},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:85
	{0, 1, 0, 1},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:86
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:87

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:88
var wv = [dModel][dModel]float64{

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:89
	{1, 0, 1, 0},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:90
	{0, 1, 0, 1},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:91
	{1, 1, 1, 1},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:92
	{0, 0, 0, 1},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:93
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:94

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:95
var w1 = [dModel][dFF]float64{

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:96
	{0.2, -0.1, 0.3, 0.0, -0.2, 0.1, 0.0, 0.3},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:97
	{-0.1, 0.2, 0.0, 0.3, 0.1, -0.2, 0.3, 0.0},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:98
	{0.3, 0.0, -0.1, 0.2, 0.0, 0.3, -0.2, 0.1},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:99
	{0.0, 0.3, 0.2, -0.1, 0.3, 0.0, 0.1, -0.2},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:100
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:101
var b1 = [dFF]float64{0.05, -0.05, 0.05, -0.05, 0.05, -0.05, 0.05, -0.05}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:102

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:103
var w2 = [dFF][dModel]float64{

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:104
	{0.2, -0.1, 0.1, 0.0},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:105
	{-0.1, 0.2, 0.0, 0.1},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:106
	{0.1, 0.0, 0.2, -0.1},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:107
	{0.0, 0.1, -0.1, 0.2},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:108
	{0.2, 0.0, -0.1, 0.1},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:109
	{0.0, 0.2, 0.1, -0.1},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:110
	{-0.1, 0.1, 0.0, 0.2},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:111
	{0.1, -0.1, 0.2, 0.0},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:112
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:113
var b2 = [dModel]float64{0, 0, 0, 0}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:114

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:115
var wo = [dModel][vocabSize]float64{

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:116
	{0.5, -0.2, 0.1, 0.0},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:117
	{0.0, 0.5, -0.2, 0.1},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:118
	{0.1, 0.0, 0.5, -0.2},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:119
	{-0.2, 0.1, 0.0, 0.5},

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:120
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:121

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:122
func wordAt(tok int) string { return vocab[tok%vocabSize] }

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:123

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:124
func posEncode(pos int) [dModel]float64 {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:125
	var pe [dModel]float64

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:126
	for i := 0; i < dModel; i += 2 {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:127
		freq := 1.0 / math.Pow(10000, float64(i)/float64(dModel))

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:128
		pe[i] = math.Sin(float64(pos) * freq)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:129
		if i+1 < dModel {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:130
			pe[i+1] = math.Cos(float64(pos) * freq)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:131
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:132
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:133
	return pe

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:134
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:135

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:136
func add4(a, b [dModel]float64) [dModel]float64 {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:137
	var y [dModel]float64

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:138
	for i := 0; i < dModel; i++ {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:139
		y[i] = a[i] + b[i]

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:140
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:141
	return y

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:142
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:143

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:144
func matVec4(w [dModel][dModel]float64, x [dModel]float64) [dModel]float64 {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:145
	var y [dModel]float64

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:146
	for i := 0; i < dModel; i++ {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:147
		sum := 0.0

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:148
		for j := 0; j < dModel; j++ {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:149
			sum += w[i][j] * x[j]

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:150
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:151
		y[i] = sum

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:152
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:153
	return y

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:154
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:155

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:156
func dot4(a, b [dModel]float64) float64 {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:157
	sum := 0.0

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:158
	for i := 0; i < dModel; i++ {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:159
		sum += a[i] * b[i]

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:160
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:161
	return sum

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:162
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:163

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:164
func softmax(scores []float64) []float64 {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:165
	max := scores[0]

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:166
	for _, s := range scores {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:167
		if s > max {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:168
			max = s

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:169
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:170
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:171
	out := make([]float64, len(scores))

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:172
	sum := 0.0

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:173
	for i, s := range scores {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:174
		out[i] = math.Exp(s - max)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:175
		sum += out[i]

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:176
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:177
	for i := range out {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:178
		out[i] /= sum

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:179
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:180
	return out

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:181
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:182

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:183
func relu(x float64) float64 {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:184
	if x > 0 {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:185
		return x

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:186
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:187
	return 0

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:188
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:189

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:190
func feedForward(x [dModel]float64) [dModel]float64 {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:191
	var hidden [dFF]float64

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:192
	for j := 0; j < dFF; j++ {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:193
		sum := b1[j]

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:194
		for i := 0; i < dModel; i++ {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:195
			sum += x[i] * w1[i][j]

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:196
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:197
		hidden[j] = relu(sum)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:198
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:199
	var out [dModel]float64

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:200
	for k := 0; k < dModel; k++ {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:201
		sum := b2[k]

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:202
		for j := 0; j < dFF; j++ {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:203
			sum += hidden[j] * w2[j][k]

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:204
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:205
		out[k] = sum

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:206
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:207
	return out

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:208
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:209

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:210
func outputLogits(x [dModel]float64) [vocabSize]float64 {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:211
	var logits [vocabSize]float64

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:212
	for k := 0; k < vocabSize; k++ {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:213
		sum := 0.0

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:214
		for i := 0; i < dModel; i++ {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:215
			sum += x[i] * wo[i][k]

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:216
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:217
		logits[k] = sum

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:218
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:219
	return logits

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:220
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:221

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:222
func argmax4(v [vocabSize]float64) int {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:223
	best := 0

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:224
	for i := 1; i < vocabSize; i++ {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:225
		if v[i] > v[best] {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:226
			best = i

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:227
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:228
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:229
	return best

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:230
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:231

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:232
// runTransformer runs one tiny single-block transformer encoder pass

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:233
// -- embedding + positional encoding, self-attention over the whole

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:234
// row, a residual connection, a feed-forward layer, a second

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:235
// residual, an output projection and an argmax -- and returns each

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:236
// position's predicted token. This is the only place in the program

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:237
// that does any "ML"; everywhere else just moves tokens/predictions

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:238
// across the mesh.

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:239
func runTransformer(tokens []int) []int {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:240
	cols := len(tokens)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:241

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:242
	x0 := make([][dModel]float64, cols)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:243
	for c := 0; c < cols; c++ {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:244
		x0[c] = add4(embedding[tokens[c]%vocabSize], posEncode(c))

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:245
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:246

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:247
	q := make([][dModel]float64, cols)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:248
	k := make([][dModel]float64, cols)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:249
	v := make([][dModel]float64, cols)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:250
	for c := 0; c < cols; c++ {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:251
		q[c] = matVec4(wq, x0[c])

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:252
		k[c] = matVec4(wk, x0[c])

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:253
		v[c] = matVec4(wv, x0[c])

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:254
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:255

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:256
	predicted := make([]int, cols)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:257
	for c := 0; c < cols; c++ {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:258
		scores := make([]float64, cols)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:259
		for j := 0; j < cols; j++ {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:260
			scores[j] = dot4(q[c], k[j]) / math.Sqrt(float64(dModel))

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:261
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:262
		weights := softmax(scores)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:263

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:264
		var attnOut [dModel]float64

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:265
		for j := 0; j < cols; j++ {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:266
			for i := 0; i < dModel; i++ {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:267
				attnOut[i] += weights[j] * v[j][i]

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:268
			}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:269
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:270

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:271
		resid1 := add4(x0[c], attnOut)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:272
		resid2 := add4(resid1, feedForward(resid1))

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:273
		predicted[c] = argmax4(outputLogits(resid2))

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:274
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:275

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:276
	return predicted

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:277
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:278

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:279
// controller runs at exactly one processor, (0,0): gathers every row-0

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:280
// column's token, runs the transformer block once the whole row has

// arrived, then scatters each column's prediction back out.
//
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:281
func
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:282
controller(toEast chan<- int32, fromEast <-chan int32) {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:283
	cols := bilink.NumCols()

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:284
	bilink.Screenf("controller ready, %d cols", cols)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:285

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:286
	for gen := 0; gen < generations; gen++ {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:287
		tokens := make([]int, cols)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:288
		tokens[0] = gen % vocabSize

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:289

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:290
		for r := cols - 1; r >= 1; r-- {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:291
			time.Sleep(hopDelay)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:292
			var v int32

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:293
			v = bilink.Recv(1)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:293

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:294
			tokens[r] = int(v)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:295
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:296

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:297
		predicted := runTransformer(tokens)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:298

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:299
		for s := 1; s < cols; s++ {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:300
			time.Sleep(hopDelay)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:301
			bilink.Send(1, int32(predicted[s]))
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:301

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:302
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:303

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:304
		bilink.Screenf("gen %d: in=%q out=%q", gen, wordAt(tokens[0]), wordAt(predicted[0]))

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:305
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:306

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:307
	bilink.Screenf("controller done")

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:308
	time.Sleep(1<<63 - 1)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:308

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:309
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:310

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:311
// rowEnd runs at (0,cols-1): the far east end of row 0. It's the

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:312
// source for exactly one gather round (its own column) and the

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:313
// destination for exactly one scatter round -- everything else about

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:314
// gather/scatter happens strictly west of it, so unlike relay it

// needs no round loop at all.
//
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:315
func
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:316
rowEnd(toWest chan<- int32, fromWest <-chan int32) {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:317
	c := bilink.NumCols() - 1

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:318
	bilink.Screenf("row-end ready at col %d", c)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:319

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:320
	for gen := 0; gen < generations; gen++ {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:321
		myTok := (c + gen) % vocabSize

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:322
		time.Sleep(hopDelay)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:323
		bilink.Send(3, int32(myTok))
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:323

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:324

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:325
		var predicted int32

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:326
		predicted = bilink.Recv(3)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:326

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:327
		bilink.Screenf("gen %d: %q -> %q", gen, wordAt(myTok), wordAt(int(predicted)))

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:328
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:329

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:330
	bilink.Screenf("row-end done")

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:331
	time.Sleep(1<<63 - 1)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:331

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:332
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:333

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:334
// relay runs on every other row-0 processor. For each gather round r:

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:335
// if r is its own column it originates (sends its token west); if r

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:336
// is further east it forwards (recv east, send west); otherwise that

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:337
// round isn't on its path at all. Scatter mirrors this exactly,

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:338
// flowing east instead of west. This is Example 20's relay idiom,

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:339
// just run `cols-1` times per generation instead of once. Both

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:340
// physical links (east, west) are used in both directions across the

// two phases, so each becomes a pair of directional parameters.
//
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:341
func
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:342
relay(fromEast <-chan int32, toWest chan<- int32, fromWest <-chan int32, toEast chan<- int32) {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:343
	cols := bilink.NumCols()

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:344
	c := bilink.Col()

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:345
	bilink.Screenf("relay ready at col %d", c)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:346

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:347
	for gen := 0; gen < generations; gen++ {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:348
		myTok := (c + gen) % vocabSize

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:349

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:350
		for r := cols - 1; r >= 1; r-- {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:351
			switch {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:352
			case r == c:

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:353
				time.Sleep(hopDelay)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:354
				bilink.Send(3, int32(myTok))
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:354

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:355
			case r > c:

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:356
				var v int32

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:357
				v = bilink.Recv(1)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:357

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:358
				time.Sleep(hopDelay)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:359
				bilink.Send(3, v)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:359

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:360
			}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:361
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:362

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:363
		var predicted int32

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:364

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:365
		for s := 1; s < cols; s++ {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:366
			switch {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:367
			case s == c:

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:368
				predicted = bilink.Recv(3)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:368

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:369
			case s > c:

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:370
				var v int32

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:371
				v = bilink.Recv(3)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:371

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:372
				time.Sleep(hopDelay)

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:373
				bilink.Send(1, v)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:373

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:374
			}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:375
		}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:376

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:377
		bilink.Screenf("gen %d: %q -> %q", gen, wordAt(myTok), wordAt(int(predicted)))

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:378
	}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:379

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:380
	bilink.Screenf("relay done")

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:381
	time.Sleep(1<<63 - 1)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:381

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:382
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:383

// idle runs everywhere off row 0 -- this demo only uses one row.
//
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:384
func
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:385
idle() {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:386
	bilink.Screenf("idle -- not part of row 0")

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:387
	time.Sleep(1<<63 - 1)
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:387

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:388
}

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:389

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:390
func main() {

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:391
	cols := bilink.NumCols()

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:392

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:393
	_ = (0)
	_ = (0)
	_ = (0)
	_ = (cols - 1)
	_ = (0)
	idle()
//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:414

//line /Users/stephenroe/Library/CloudStorage/Dropbox/ClaudeZone/bil/examples/21-mesh-transformer.bil:415
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
