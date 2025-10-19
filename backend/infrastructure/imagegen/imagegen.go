package imagegen

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"

	"github.com/dmpettyp/artwork/domain/imagegraph"
)

type imageStorage interface {
	Save(imageID imagegraph.ImageID, imageData []byte) error
	Get(imageID imagegraph.ImageID) ([]byte, error)
}

type outputSetter interface {
	SetNodeOutputImage(
		ctx context.Context,
		imageGraphID imagegraph.ImageGraphID,
		nodeID imagegraph.NodeID,
		outputName imagegraph.OutputName,
		imageID imagegraph.ImageID,
	) error
}

type ImageGen struct {
	imageStorage imageStorage
	outputSetter outputSetter
}

func NewImageGen(
	imageStorage imageStorage,
	outputSetter outputSetter,
) *ImageGen {
	return &ImageGen{
		imageStorage: imageStorage,
		outputSetter: outputSetter,
	}
}

func (ig *ImageGen) GenerateOutputsForScaleNode(
	ctx context.Context,
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	inputImageID imagegraph.ImageID,
	scaleFactor float64,
	outputName imagegraph.OutputName,
) error {
	// Get the input image
	imageData, err := ig.imageStorage.Get(inputImageID)
	if err != nil {
		return fmt.Errorf("could not get input image: %w", err)
	}

	// Decode the image
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return fmt.Errorf("could not decode image: %w", err)
	}

	// Scale the image
	bounds := img.Bounds()
	newWidth := int(float64(bounds.Dx()) * scaleFactor)
	newHeight := int(float64(bounds.Dy()) * scaleFactor)

	scaledImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Simple nearest-neighbor scaling
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := int(float64(x) / scaleFactor)
			srcY := int(float64(y) / scaleFactor)
			scaledImg.Set(x, y, img.At(srcX, srcY))
		}
	}

	// Encode the scaled image
	var buf bytes.Buffer
	switch format {
	case "png":
		err = png.Encode(&buf, scaledImg)
	case "jpeg", "jpg":
		err = jpeg.Encode(&buf, scaledImg, &jpeg.Options{Quality: 90})
	default:
		return fmt.Errorf("unsupported image format: %s", format)
	}
	if err != nil {
		return fmt.Errorf("could not encode scaled image: %w", err)
	}

	// Generate new image ID and save
	outputImageID, err := imagegraph.NewImageID()
	if err != nil {
		return fmt.Errorf("could not generate image ID: %w", err)
	}

	err = ig.imageStorage.Save(outputImageID, buf.Bytes())
	if err != nil {
		return fmt.Errorf("could not save scaled image: %w", err)
	}

	// Set the output image on the node
	err = ig.outputSetter.SetNodeOutputImage(ctx, imageGraphID, nodeID, outputName, outputImageID)
	if err != nil {
		return fmt.Errorf("could not set node output image: %w", err)
	}

	return nil
}

func (ig *ImageGen) GenerateOutputsForBlurNode(
	ctx context.Context,
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	inputImageID imagegraph.ImageID,
	radius int,
	outputName imagegraph.OutputName,
) error {
	// Get the input image
	imageData, err := ig.imageStorage.Get(inputImageID)
	if err != nil {
		return fmt.Errorf("could not get input image: %w", err)
	}

	// Decode the image
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return fmt.Errorf("could not decode image: %w", err)
	}

	// Apply Gaussian blur
	bounds := img.Bounds()
	blurredImg := image.NewRGBA(bounds)

	// Simple box blur approximation of Gaussian blur
	// Apply horizontal blur pass
	tempImg := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			var r, g, b, a uint32
			count := 0

			for dx := -radius; dx <= radius; dx++ {
				px := x + dx
				if px >= bounds.Min.X && px < bounds.Max.X {
					c := img.At(px, y)
					rr, gg, bb, aa := c.RGBA()
					r += rr
					g += gg
					b += bb
					a += aa
					count++
				}
			}

			tempImg.Set(x, y, color.RGBA{
				R: uint8(r / uint32(count) >> 8),
				G: uint8(g / uint32(count) >> 8),
				B: uint8(b / uint32(count) >> 8),
				A: uint8(a / uint32(count) >> 8),
			})
		}
	}

	// Apply vertical blur pass
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			var r, g, b, a uint32
			count := 0

			for dy := -radius; dy <= radius; dy++ {
				py := y + dy
				if py >= bounds.Min.Y && py < bounds.Max.Y {
					c := tempImg.At(x, py)
					rr, gg, bb, aa := c.RGBA()
					r += rr
					g += gg
					b += bb
					a += aa
					count++
				}
			}

			blurredImg.Set(x, y, color.RGBA{
				R: uint8(r / uint32(count) >> 8),
				G: uint8(g / uint32(count) >> 8),
				B: uint8(b / uint32(count) >> 8),
				A: uint8(a / uint32(count) >> 8),
			})
		}
	}

	// Encode the blurred image
	var buf bytes.Buffer
	switch format {
	case "png":
		err = png.Encode(&buf, blurredImg)
	case "jpeg", "jpg":
		err = jpeg.Encode(&buf, blurredImg, &jpeg.Options{Quality: 90})
	default:
		return fmt.Errorf("unsupported image format: %s", format)
	}
	if err != nil {
		return fmt.Errorf("could not encode blurred image: %w", err)
	}

	// Generate new image ID and save
	outputImageID, err := imagegraph.NewImageID()
	if err != nil {
		return fmt.Errorf("could not generate image ID: %w", err)
	}

	err = ig.imageStorage.Save(outputImageID, buf.Bytes())
	if err != nil {
		return fmt.Errorf("could not save blurred image: %w", err)
	}

	// Set the output image on the node
	err = ig.outputSetter.SetNodeOutputImage(ctx, imageGraphID, nodeID, outputName, outputImageID)
	if err != nil {
		return fmt.Errorf("could not set node output image: %w", err)
	}

	return nil
}
