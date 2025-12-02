package filter

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"mime"
	"net/http"
	"path/filepath"

	"golang.org/x/image/draw"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	"github.com/akikareha/himewiki/internal/config"
)

func imageWithOpenAI(cfg *config.Config, title string, data []byte) ([]byte, error) {
	maxLength := cfg.ImageFilter.MaxLength
	if len(data) > maxLength {
		return nil, fmt.Errorf("image data too long")
	}

	ext := filepath.Ext(title)
	extFound := false
	for _, extension := range cfg.Image.Extensions {
		if ext == "."+extension {
			extFound = true
			break
		}
	}
	if !extFound {
		return nil, fmt.Errorf("invalid image extension: %s", ext)
	}

	mimeType := mime.TypeByExtension(ext)
	contentType := http.DetectContentType(data)
	if contentType != mimeType {
		return nil, fmt.Errorf("Invalid content type")
	}

	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	if "image/"+format != mimeType {
		return nil, fmt.Errorf("invalid image format")
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	newW, newH := width, height
	maxSize := cfg.ImageFilter.MaxSize
	if width > maxSize || height > maxSize {
		scale := float64(maxSize) / float64(width)
		if height > width {
			scale = float64(maxSize) / float64(height)
		}
		newW = int(float64(width) * scale)
		newH = int(float64(height) * scale)
	}

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)

	var buf bytes.Buffer
	switch format {
	case "png":
		err = png.Encode(&buf, dst)
	case "jpeg":
		err = jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 90})
	default:
		return nil, fmt.Errorf("unsupported image format: %s", format)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	imageBytes := buf.Bytes()

	apiKey := cfg.ImageFilter.Key
	if apiKey == "" {
		return nil, fmt.Errorf("Image filter key (OpenAI API key) not set")
	}
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	b64 := base64.StdEncoding.EncodeToString(imageBytes)
	dataURI := "data:" + mimeType + ";base64," + b64

	resp, err := client.Moderations.New(
		context.Background(),
		openai.ModerationNewParams{
			Model: openai.ModerationModelOmniModerationLatest,
			Input: openai.ModerationNewParamsInputUnion{
				OfModerationMultiModalArray: []openai.ModerationMultiModalInputUnionParam{
					openai.ModerationMultiModalInputParamOfText(
						title,
					),
					openai.ModerationMultiModalInputParamOfImageURL(
						openai.ModerationImageURLInputImageURLParam{
							URL: dataURI,
						},
					),
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}

	if len(resp.Results) == 0 {
		return nil, fmt.Errorf("moderation API returned no results")
	}
	result := resp.Results[0]

	if result.Flagged {
		return nil, fmt.Errorf("Image flagged. Categories: %+v\n", result.Categories)
	}

	return imageBytes, nil
}
