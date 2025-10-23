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
	"github.com/nfnt/resize"
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

func (ig *ImageGen) GenerateOutputsForResizeNode(
	ctx context.Context,
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	inputImageID imagegraph.ImageID,
	width *int,
	height *int,
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

	// Calculate target dimensions
	bounds := img.Bounds()
	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()

	var newWidth, newHeight int

	if width != nil && height != nil {
		// Both set: use exact dimensions
		newWidth = *width
		newHeight = *height
	} else if width != nil {
		// Only width set: calculate height proportionally
		newWidth = *width
		aspectRatio := float64(originalHeight) / float64(originalWidth)
		newHeight = int(float64(newWidth) * aspectRatio)
	} else if height != nil {
		// Only height set: calculate width proportionally
		newHeight = *height
		aspectRatio := float64(originalWidth) / float64(originalHeight)
		newWidth = int(float64(newHeight) * aspectRatio)
	} else {
		return fmt.Errorf("at least one of width or height must be set")
	}

	// Create resized image
	resizedImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Simple nearest-neighbor scaling
	scaleX := float64(originalWidth) / float64(newWidth)
	scaleY := float64(originalHeight) / float64(newHeight)

	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := int(float64(x) * scaleX)
			srcY := int(float64(y) * scaleY)
			resizedImg.Set(x, y, img.At(srcX, srcY))
		}
	}

	// Encode the resized image
	var buf bytes.Buffer
	switch format {
	case "png":
		err = png.Encode(&buf, resizedImg)
	case "jpeg", "jpg":
		err = jpeg.Encode(&buf, resizedImg, &jpeg.Options{Quality: 90})
	default:
		return fmt.Errorf("unsupported image format: %s", format)
	}
	if err != nil {
		return fmt.Errorf("could not encode resized image: %w", err)
	}

	// Generate new image ID and save
	outputImageID, err := imagegraph.NewImageID()
	if err != nil {
		return fmt.Errorf("could not generate image ID: %w", err)
	}

	err = ig.imageStorage.Save(outputImageID, buf.Bytes())
	if err != nil {
		return fmt.Errorf("could not save resized image: %w", err)
	}

	// Set the output image on the node
	err = ig.outputSetter.SetNodeOutputImage(ctx, imageGraphID, nodeID, outputName, outputImageID)
	if err != nil {
		return fmt.Errorf("could not set node output image: %w", err)
	}

	return nil
}

func (ig *ImageGen) GenerateOutputsForResizeMatchNode(
	ctx context.Context,
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	originalImageID imagegraph.ImageID,
	sizeMatchImageID imagegraph.ImageID,
	outputName imagegraph.OutputName,
) error {
	// Get the original image
	originalImageData, err := ig.imageStorage.Get(originalImageID)
	if err != nil {
		return fmt.Errorf("could not get original image: %w", err)
	}

	// Get the size_match image
	sizeMatchImageData, err := ig.imageStorage.Get(sizeMatchImageID)
	if err != nil {
		return fmt.Errorf("could not get size_match image: %w", err)
	}

	// Decode the original image
	originalImg, format, err := image.Decode(bytes.NewReader(originalImageData))
	if err != nil {
		return fmt.Errorf("could not decode original image: %w", err)
	}

	// Decode the size_match image to get dimensions
	sizeMatchImg, _, err := image.Decode(bytes.NewReader(sizeMatchImageData))
	if err != nil {
		return fmt.Errorf("could not decode size_match image: %w", err)
	}

	// Get target dimensions from size_match image
	targetBounds := sizeMatchImg.Bounds()
	targetWidth := uint(targetBounds.Dx())
	targetHeight := uint(targetBounds.Dy())

	// Resize original image to match size_match dimensions using LANCZOS
	resizedImg := resize.Resize(targetWidth, targetHeight, originalImg, resize.Lanczos3)

	// Encode the resized image
	var buf bytes.Buffer
	switch format {
	case "png":
		err = png.Encode(&buf, resizedImg)
	case "jpeg", "jpg":
		err = jpeg.Encode(&buf, resizedImg, &jpeg.Options{Quality: 90})
	default:
		return fmt.Errorf("unsupported image format: %s", format)
	}
	if err != nil {
		return fmt.Errorf("could not encode resized image: %w", err)
	}

	// Generate new image ID and save
	outputImageID, err := imagegraph.NewImageID()
	if err != nil {
		return fmt.Errorf("could not generate image ID: %w", err)
	}

	err = ig.imageStorage.Save(outputImageID, buf.Bytes())
	if err != nil {
		return fmt.Errorf("could not save resized image: %w", err)
	}

	// Set the output image on the node
	err = ig.outputSetter.SetNodeOutputImage(ctx, imageGraphID, nodeID, outputName, outputImageID)
	if err != nil {
		return fmt.Errorf("could not set node output image: %w", err)
	}

	return nil
}
