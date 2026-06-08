// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package geminiclient нь Google Gemini API-ийн generateContent дуудлагыг
// нимгэн боож HTTP клиент болгоно (pkg/aiclient-тэй ижил загвар). SDK
// хамаарал нэмэхгүйгээр зөвхөн стандарт сангаар REST дуудлага хийнэ.
//
// Энэ клиент нь дуу хоолойн орчуулгын pipeline-ийн хоёр алхмыг хангана:
//  1. TranscribeAndTranslate — аудиог бичвэрлэж (STT) нөгөө хэл рүү
//     орчуулна (нэг multimodal дуудлагаар — хурд + кост хэмнэлт);
//  2. Synthesize — орчуулсан бичвэрийг яриа болгоно (TTS), үр дүнг
//     browser шууд тоглуулж болохуйц WAV болгон боож буцаана.
//
// Аюулгүй байдал: API түлхүүр зөвхөн энэ процессын санах ойд байна —
// browser/клиент рүү хэзээ ч дамжихгүй (BFF → backend → Gemini).
package geminiclient

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DefaultBaseURL нь Gemini API-ийн production endpoint юм.
const DefaultBaseURL = "https://generativelanguage.googleapis.com"

const (
	// defaultSTTModel нь аудио ойлгож, орчуулах multimodal модель.
	defaultSTTModel = "gemini-2.5-flash"
	// defaultTTSModel нь яриа үүсгэх (text-to-speech) модель.
	defaultTTSModel = "gemini-2.5-flash-preview-tts"
	// defaultVoice нь TTS-ийн өгөгдмөл дуу хоолой (Gemini-ийн prebuilt
	// дуу хоолой нь олон хэл дээр ажилладаг тул хэл бүрт нэгийг ашиглана).
	defaultVoice = "Kore"

	// Gemini TTS-ийн гаргадаг аудио формат — түүхий PCM. WAV толгой
	// үүсгэхэд ашиглана.
	ttsSampleRate    = 24000
	ttsBitsPerSample = 16
	ttsChannels      = 1
)

// Usage нь нэг дуудлагын токен зарцуулалт — кост хяналт, метерингд
// зориулж дуудагч руу буцаана.
type Usage struct {
	InputTokens  int
	OutputTokens int
}

// TranslationResult нь TranscribeAndTranslate-ийн үр дүн.
type TranslationResult struct {
	SourceText     string // эх хэл дээрх бичвэр (STT)
	TranslatedText string // зорилтот хэл рүү орчуулсан бичвэр
	Usage          Usage
}

// Voicer нь usecase давхаргын харьцах хил (boundary) — тестэд бодит HTTP
// алхалгүй мокчлох боломж олгоно (aiclient.Streamer-тэй ижил санаа).
type Voicer interface {
	// Transcribe нь аудиог тэр хэл дээрээ бичвэрлэнэ (орчуулгагүй STT) —
	// чатад дуугаар асуухад ашиглана. lang нь "mn"/"en" зөвлөмж.
	Transcribe(ctx context.Context, audio []byte, mimeType, lang string) (text string, usage Usage, err error)
	// TranscribeAndTranslate нь аудиог бичвэрлэж sourceLang → targetLang
	// орчуулна. mimeType нь аудионы MIME (жнь "audio/webm").
	TranscribeAndTranslate(ctx context.Context, audio []byte, mimeType, sourceLang, targetLang string) (TranslationResult, error)
	// Synthesize нь бичвэрийг яриа болгож, WAV-ээр боосон аудио байт буцаана.
	Synthesize(ctx context.Context, text string) (wav []byte, usage Usage, err error)
}

// Client нь Gemini generateContent API-ийн HTTP клиент юм.
type Client struct {
	baseURL  string
	apiKey   string
	sttModel string
	ttsModel string
	voice    string
	http     *http.Client
}

// NewClient нь шинэ Client үүсгэнэ. apiKey хоосон үед клиент бүтэх боловч
// дуудлага бүр "api key is not configured" алдаа буцаана — операторт
// чимээгүй буруу тохиргоо үлдээхгүй (aiclient.NewClient-тэй ижил зарчим).
// timeout нь нэг REST дуудлагын дээд хугацаа.
func NewClient(apiKey, sttModel, ttsModel, voice string, timeout time.Duration) *Client {
	if sttModel == "" {
		sttModel = defaultSTTModel
	}
	if ttsModel == "" {
		ttsModel = defaultTTSModel
	}
	if voice == "" {
		voice = defaultVoice
	}
	if timeout <= 0 {
		timeout = 25 * time.Second
	}
	return &Client{
		baseURL:  DefaultBaseURL,
		apiKey:   apiKey,
		sttModel: sttModel,
		ttsModel: ttsModel,
		voice:    voice,
		http:     &http.Client{Timeout: timeout},
	}
}

// Model нь STT/орчуулгын модель нэрийг буцаана (usage метерингд бичигдэнэ).
func (c *Client) Model() string { return c.sttModel }

// Configured нь API түлхүүр тохируулагдсан эсэхийг буцаана.
func (c *Client) Configured() bool { return c.apiKey != "" }

// --- generateContent-ийн хүсэлт/хариуны бүтэц ---

type genPart struct {
	Text       string     `json:"text,omitempty"`
	InlineData *genInline `json:"inlineData,omitempty"`
}

type genInline struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type genContent struct {
	Role  string    `json:"role,omitempty"`
	Parts []genPart `json:"parts"`
}

type prebuiltVoice struct {
	VoiceName string `json:"voiceName"`
}
type voiceConfig struct {
	PrebuiltVoiceConfig prebuiltVoice `json:"prebuiltVoiceConfig"`
}
type speechConfig struct {
	VoiceConfig voiceConfig `json:"voiceConfig"`
}

type genConfig struct {
	ResponseMimeType   string        `json:"responseMimeType,omitempty"`
	ResponseModalities []string      `json:"responseModalities,omitempty"`
	Temperature        *float64      `json:"temperature,omitempty"`
	SpeechConfig       *speechConfig `json:"speechConfig,omitempty"`
}

type genRequest struct {
	Contents         []genContent `json:"contents"`
	GenerationConfig *genConfig   `json:"generationConfig,omitempty"`
}

type genResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text       string `json:"text"`
				InlineData *struct {
					MimeType string `json:"mimeType"`
					Data     string `json:"data"`
				} `json:"inlineData"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
	} `json:"usageMetadata"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// doGenerate нь generateContent дуудлагыг хийж хариуг задлана.
func (c *Client) doGenerate(ctx context.Context, model string, body genRequest) (*genResponse, error) {
	if c.apiKey == "" {
		return nil, errors.New("geminiclient: api key is not configured")
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("geminiclient: marshal request: %w", err)
	}
	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent", c.baseURL, model)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("geminiclient: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// AI Studio API түлхүүр нь x-goog-api-key толгойгоор дамжина.
	req.Header.Set("x-goog-api-key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("geminiclient: send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Аудио хариу том байж болзошгүй тул 32 MiB хүртэл уншина.
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return nil, fmt.Errorf("geminiclient: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var er genResponse
		_ = json.Unmarshal(raw, &er)
		msg := ""
		if er.Error != nil {
			msg = strings.TrimSpace(er.Error.Message)
		}
		if msg == "" {
			msg = strings.TrimSpace(string(raw))
		}
		return nil, fmt.Errorf("geminiclient: generateContent returned %d: %s", resp.StatusCode, msg)
	}

	var gr genResponse
	if err := json.Unmarshal(raw, &gr); err != nil {
		return nil, fmt.Errorf("geminiclient: decode response: %w", err)
	}
	if gr.Error != nil && strings.TrimSpace(gr.Error.Message) != "" {
		return nil, fmt.Errorf("geminiclient: api error: %s", gr.Error.Message)
	}
	return &gr, nil
}

// TranscribeAndTranslate нь Voicer интерфейсийг хэрэгжүүлнэ. Нэг multimodal
// дуудлагаар аудиог бичвэрлэж, нөгөө хэл рүү орчуулна; үр дүнг JSON-оор
// (responseMimeType) албадаж авч найдвартай задлана.
func (c *Client) TranscribeAndTranslate(ctx context.Context, audio []byte, mimeType, sourceLang, targetLang string) (TranslationResult, error) {
	var out TranslationResult
	srcName := langName(sourceLang)
	tgtName := langName(targetLang)

	instruction := fmt.Sprintf(
		"You are a professional %s-%s interpreter. The user provides an audio clip spoken in %s. "+
			"First transcribe exactly what is said, then translate it into natural, fluent %s. "+
			"Return ONLY a JSON object with two string fields: "+
			`{"transcript": "<the original %s transcription>", "translation": "<the %s translation>"}. `+
			"Do not add any commentary.",
		srcName, tgtName, srcName, tgtName, srcName, tgtName,
	)
	temp := 0.2
	body := genRequest{
		Contents: []genContent{{
			Role: "user",
			Parts: []genPart{
				{Text: instruction},
				{InlineData: &genInline{MimeType: mimeType, Data: base64.StdEncoding.EncodeToString(audio)}},
			},
		}},
		GenerationConfig: &genConfig{
			ResponseMimeType: "application/json",
			Temperature:      &temp,
		},
	}

	gr, err := c.doGenerate(ctx, c.sttModel, body)
	if err != nil {
		return out, err
	}
	text := firstText(gr)
	if text == "" {
		return out, errors.New("geminiclient: empty transcription response")
	}

	var parsed struct {
		Transcript  string `json:"transcript"`
		Translation string `json:"translation"`
	}
	if err := json.Unmarshal([]byte(stripJSONFence(text)), &parsed); err != nil {
		return out, fmt.Errorf("geminiclient: parse translation json: %w", err)
	}
	out.SourceText = strings.TrimSpace(parsed.Transcript)
	out.TranslatedText = strings.TrimSpace(parsed.Translation)
	out.Usage = Usage{
		InputTokens:  gr.UsageMetadata.PromptTokenCount,
		OutputTokens: gr.UsageMetadata.CandidatesTokenCount,
	}
	if out.TranslatedText == "" {
		return out, errors.New("geminiclient: model returned empty translation")
	}
	return out, nil
}

// Transcribe нь Voicer интерфейсийг хэрэгжүүлнэ. Аудиог тэр хэл дээрээ
// бичвэрлэж (орчуулгагүй), цэвэр текст буцаана.
func (c *Client) Transcribe(ctx context.Context, audio []byte, mimeType, lang string) (string, Usage, error) {
	var usage Usage
	instruction := fmt.Sprintf(
		"Transcribe the following audio in %s exactly as spoken. "+
			"Return ONLY the transcription text with no commentary, labels, or quotation marks.",
		langName(lang),
	)
	temp := 0.0
	body := genRequest{
		Contents: []genContent{{
			Role: "user",
			Parts: []genPart{
				{Text: instruction},
				{InlineData: &genInline{MimeType: mimeType, Data: base64.StdEncoding.EncodeToString(audio)}},
			},
		}},
		GenerationConfig: &genConfig{Temperature: &temp},
	}
	gr, err := c.doGenerate(ctx, c.sttModel, body)
	if err != nil {
		return "", usage, err
	}
	text := strings.TrimSpace(firstText(gr))
	usage = Usage{
		InputTokens:  gr.UsageMetadata.PromptTokenCount,
		OutputTokens: gr.UsageMetadata.CandidatesTokenCount,
	}
	if text == "" {
		return "", usage, errors.New("geminiclient: empty transcription response")
	}
	return text, usage, nil
}

// Synthesize нь Voicer интерфейсийг хэрэгжүүлнэ. Бичвэрийг яриа болгож,
// Gemini-ийн буцаасан түүхий PCM-г WAV болгон боож буцаана (browser-ийн
// <audio> түүхий PCM тоглуулж чадахгүй тул RIFF толгой шаардлагатай).
func (c *Client) Synthesize(ctx context.Context, text string) ([]byte, Usage, error) {
	var usage Usage
	body := genRequest{
		Contents: []genContent{{
			Parts: []genPart{{Text: text}},
		}},
		GenerationConfig: &genConfig{
			ResponseModalities: []string{"AUDIO"},
			SpeechConfig: &speechConfig{
				VoiceConfig: voiceConfig{PrebuiltVoiceConfig: prebuiltVoice{VoiceName: c.voice}},
			},
		},
	}

	// Gemini TTS завсрын байдлаар хоёр янзаар бүтэлгүйтдэг: (а) finishReason
	// "OTHER"-оор хоосон аудио, эсвэл (б) 400 "Model tried to generate text"
	// (текстийг яриа болгохын оронд хариулах гэж оролдох). Хоёулаа ижил оролт
	// дээр заримдаа тохиолддог тул хэд дахин оролдоно. Бусад алдаа (auth,
	// rate limit г.м.) дээр шууд буцна. ctx-ийн deadline нийт хугацааг хязгаарлана.
	var b64 string
	var lastErr error
	for attempt := 0; attempt < 6; attempt++ {
		gr, err := c.doGenerate(ctx, c.ttsModel, body)
		if err != nil {
			if strings.Contains(err.Error(), "should only be used for TTS") {
				lastErr = err
				continue
			}
			return nil, usage, err
		}
		usage = Usage{
			InputTokens:  gr.UsageMetadata.PromptTokenCount,
			OutputTokens: gr.UsageMetadata.CandidatesTokenCount,
		}
		if b64 = firstInlineData(gr); b64 != "" {
			break
		}
		lastErr = errors.New("geminiclient: empty audio response")
	}
	if b64 == "" {
		if lastErr != nil {
			return nil, usage, lastErr
		}
		return nil, usage, errors.New("geminiclient: empty audio response")
	}
	pcm, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, usage, fmt.Errorf("geminiclient: decode audio: %w", err)
	}
	return pcmToWAV(pcm, ttsSampleRate, ttsBitsPerSample, ttsChannels), usage, nil
}

// --- туслахууд ---

func firstText(gr *genResponse) string {
	if len(gr.Candidates) == 0 {
		return ""
	}
	for _, p := range gr.Candidates[0].Content.Parts {
		if p.Text != "" {
			return p.Text
		}
	}
	return ""
}

func firstInlineData(gr *genResponse) string {
	if len(gr.Candidates) == 0 {
		return ""
	}
	for _, p := range gr.Candidates[0].Content.Parts {
		if p.InlineData != nil && p.InlineData.Data != "" {
			return p.InlineData.Data
		}
	}
	return ""
}

// stripJSONFence нь модель JSON-г ```json ... ``` дотор боосон тохиолдолд
// хашилтыг арилгана (responseMimeType-тэй ч ховор тохиолдож болзошгүй).
func stripJSONFence(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSuffix(s, "```")
	}
	return strings.TrimSpace(s)
}

func langName(code string) string {
	switch code {
	case "mn":
		return "Mongolian"
	case "en":
		return "English"
	default:
		return code
	}
}

// pcmToWAV нь signed 16-bit little-endian PCM байтуудад 44 байтын RIFF/WAVE
// толгой нэмж, бие даан тоглуулж болохуйц WAV файл болгоно.
func pcmToWAV(pcm []byte, sampleRate, bitsPerSample, channels int) []byte {
	byteRate := sampleRate * channels * bitsPerSample / 8
	blockAlign := channels * bitsPerSample / 8
	dataLen := len(pcm)

	buf := &bytes.Buffer{}
	buf.Grow(44 + dataLen)
	// RIFF chunk descriptor
	buf.WriteString("RIFF")
	_ = binary.Write(buf, binary.LittleEndian, uint32(36+dataLen))
	buf.WriteString("WAVE")
	// fmt sub-chunk
	buf.WriteString("fmt ")
	_ = binary.Write(buf, binary.LittleEndian, uint32(16)) // PCM fmt chunk size
	_ = binary.Write(buf, binary.LittleEndian, uint16(1))  // audio format = PCM
	_ = binary.Write(buf, binary.LittleEndian, uint16(channels))
	_ = binary.Write(buf, binary.LittleEndian, uint32(sampleRate))
	_ = binary.Write(buf, binary.LittleEndian, uint32(byteRate))
	_ = binary.Write(buf, binary.LittleEndian, uint16(blockAlign))
	_ = binary.Write(buf, binary.LittleEndian, uint16(bitsPerSample))
	// data sub-chunk
	buf.WriteString("data")
	_ = binary.Write(buf, binary.LittleEndian, uint32(dataLen))
	buf.Write(pcm)
	return buf.Bytes()
}
