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

	SetNodePreviewImage(
		ctx context.Context,
		imageGraphID imagegraph.ImageGraphID,
		nodeID imagegraph.NodeID,
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

// encodeImage encodes an image to bytes based on the format
func (ig *ImageGen) encodeImage(img image.Image, format string) ([]byte, error) {
	var buf bytes.Buffer
	var err error

	switch format {
	case "png":
		err = png.Encode(&buf, img)
	case "jpeg", "jpg":
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	default:
		return nil, fmt.Errorf("unsupported image format: %s", format)
	}

	if err != nil {
		return nil, fmt.Errorf("could not encode image: %w", err)
	}

	return buf.Bytes(), nil
}

// loadImage fetches an image from storage and decodes it
func (ig *ImageGen) loadImage(imageID imagegraph.ImageID) (image.Image, string, error) {
	// Get the image data from storage
	imageData, err := ig.imageStorage.Get(imageID)
	if err != nil {
		return nil, "", fmt.Errorf("could not get image: %w", err)
	}

	// Decode the image
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, "", fmt.Errorf("could not decode image: %w", err)
	}

	return img, format, nil
}

// saveAndSetOutput encodes an image, saves it to storage, and sets it as a node output
func (ig *ImageGen) saveAndSetOutput(
	ctx context.Context,
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	outputName imagegraph.OutputName,
	img image.Image,
	format string,
) error {
	// Encode the image
	imageData, err := ig.encodeImage(img, format)
	if err != nil {
		return err
	}

	// Generate new image ID
	outputImageID, err := imagegraph.NewImageID()
	if err != nil {
		return fmt.Errorf("could not generate image ID: %w", err)
	}

	// Save to storage
	err = ig.imageStorage.Save(outputImageID, imageData)
	if err != nil {
		return fmt.Errorf("could not save image: %w", err)
	}

	// Set the output image on the node
	err = ig.outputSetter.SetNodeOutputImage(ctx, imageGraphID, nodeID, outputName, outputImageID)
	if err != nil {
		return fmt.Errorf("could not set node output image: %w", err)
	}

	return nil
}

func (ig *ImageGen) saveAndSetPreview(
	ctx context.Context,
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	img image.Image,
	format string,
) error {
	bounds := img.Bounds()
	width := uint(bounds.Dx())
	height := uint(bounds.Dy())

	interpolationFunction := resize.Bicubic

	if width < 160 || height < 120 {
		interpolationFunction = resize.NearestNeighbor
	}

	if width > height {
		width = 160
		height = 0
	} else {
		width = 0
		height = 120
	}

	previewImg := resize.Resize(width, height, img, interpolationFunction)

	// Encode the image
	imageData, err := ig.encodeImage(previewImg, format)
	if err != nil {
		return err
	}

	// Generate new image ID
	previewImageID, err := imagegraph.NewImageID()
	if err != nil {
		return fmt.Errorf("could not generate preview image ID: %w", err)
	}

	// Save to storage
	err = ig.imageStorage.Save(previewImageID, imageData)
	if err != nil {
		return fmt.Errorf("could not save preview image: %w", err)
	}

	// Set the output image on the node
	err = ig.outputSetter.SetNodePreviewImage(ctx, imageGraphID, nodeID, previewImageID)
	if err != nil {
		return fmt.Errorf("could not set node preview image: %w", err)
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
	// Load the input image
	img, format, err := ig.loadImage(inputImageID)
	if err != nil {
		return err
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

	err = ig.saveAndSetPreview(ctx, imageGraphID, nodeID, blurredImg, format)

	if err != nil {
		return fmt.Errorf("could not generate outputs for blur node: %w", err)
	}

	err = ig.saveAndSetOutput(ctx, imageGraphID, nodeID, outputName, blurredImg, format)

	if err != nil {
		return fmt.Errorf("could not generate outputs for blur node: %w", err)
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
	interpolation string,
	outputName imagegraph.OutputName,
) error {
	// Load the input image
	img, format, err := ig.loadImage(inputImageID)
	if err != nil {
		return err
	}

	// Get interpolation function
	interpolationFunction, ok := resizeInterpolationFunctions[interpolation]
	if !ok {
		return fmt.Errorf("unsupported interpolation function %q", interpolation)
	}

	// Calculate target dimensions
	var targetWidth, targetHeight uint

	if width != nil && height != nil {
		// Both set: use exact dimensions
		targetWidth = uint(*width)
		targetHeight = uint(*height)
	} else if width != nil {
		// Only width set: calculate height proportionally
		targetWidth = uint(*width)
		targetHeight = 0 // resize library will maintain aspect ratio
	} else if height != nil {
		// Only height set: calculate width proportionally
		targetWidth = 0 // resize library will maintain aspect ratio
		targetHeight = uint(*height)
	} else {
		return fmt.Errorf("at least one of width or height must be set")
	}

	resizedImg := resize.Resize(targetWidth, targetHeight, img, interpolationFunction)

	err = ig.saveAndSetPreview(ctx, imageGraphID, nodeID, resizedImg, format)

	if err != nil {
		return fmt.Errorf("could not generate outputs for resize node: %w", err)
	}

	err = ig.saveAndSetOutput(ctx, imageGraphID, nodeID, outputName, resizedImg, format)

	if err != nil {
		return fmt.Errorf("could not generate outputs for resize node: %w", err)
	}

	return nil
}

var resizeInterpolationFunctions = map[string]resize.InterpolationFunction{
	"NearestNeighbor":   resize.NearestNeighbor,
	"Bilinear":          resize.Bilinear,
	"Bicubic":           resize.Bicubic,
	"MitchellNetravali": resize.MitchellNetravali,
	"Lanczos2":          resize.Lanczos2,
	"Lanczos3":          resize.Lanczos3,
}

func (ig *ImageGen) GenerateOutputsForResizeMatchNode(
	ctx context.Context,
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	originalImageID imagegraph.ImageID,
	sizeMatchImageID imagegraph.ImageID,
	interpolation string,
	outputName imagegraph.OutputName,
) error {
	// Load the original image
	originalImg, format, err := ig.loadImage(originalImageID)
	if err != nil {
		return err
	}

	// Load the size_match image to get dimensions
	sizeMatchImg, _, err := ig.loadImage(sizeMatchImageID)
	if err != nil {
		return err
	}

	// Get target dimensions from size_match image
	targetBounds := sizeMatchImg.Bounds()
	targetWidth := uint(targetBounds.Dx())
	targetHeight := uint(targetBounds.Dy())

	interpolationFunction, ok := resizeInterpolationFunctions[interpolation]

	if !ok {
		return fmt.Errorf("unsupported interpolation function %q", interpolation)
	}

	resizedImg := resize.Resize(
		targetWidth,
		targetHeight,
		originalImg,
		interpolationFunction,
	)

	err = ig.saveAndSetPreview(ctx, imageGraphID, nodeID, resizedImg, format)

	if err != nil {
		return fmt.Errorf("could not generate outputs for resize match node: %w", err)
	}

	err = ig.saveAndSetOutput(ctx, imageGraphID, nodeID, outputName, resizedImg, format)

	if err != nil {
		return fmt.Errorf("could not generate outputs for resize match node: %w", err)
	}

	return nil
}

func (ig *ImageGen) GenerateOutputsForCropNode(
	ctx context.Context,
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	imageID imagegraph.ImageID,
	left, right, top, bottom int,
	outputName imagegraph.OutputName,
) error {
	// Load the original image
	originalImage, format, err := ig.loadImage(imageID)
	if err != nil {
		return err
	}

	// Get the original image bounds
	bounds := originalImage.Bounds()

	// Clamp crop coordinates to actual image bounds
	if left < bounds.Min.X {
		left = bounds.Min.X
	}
	if right > bounds.Max.X {
		right = bounds.Max.X
	}
	if top < bounds.Min.Y {
		top = bounds.Min.Y
	}
	if bottom > bounds.Max.Y {
		bottom = bounds.Max.Y
	}

	// Ensure we still have a valid rectangle after clamping
	if left >= right || top >= bottom {
		return fmt.Errorf("crop rectangle is invalid or outside image bounds")
	}

	// Create the crop rectangle
	cropRect := image.Rect(left, top, right, bottom)

	// Create a sub-image (this is a view, not a copy)
	var croppedImg image.Image
	if subImager, ok := originalImage.(interface {
		SubImage(r image.Rectangle) image.Image
	}); ok {
		croppedImg = subImager.SubImage(cropRect)
	} else {
		return fmt.Errorf("image type does not support cropping")
	}

	err = ig.saveAndSetPreview(ctx, imageGraphID, nodeID, croppedImg, format)

	if err != nil {
		return fmt.Errorf("could not generate outputs for crop node: %w", err)
	}

	err = ig.saveAndSetOutput(ctx, imageGraphID, nodeID, outputName, croppedImg, format)

	if err != nil {
		return fmt.Errorf("could not generate outputs for crop node: %w", err)
	}

	return nil
}

func (ig *ImageGen) GenerateOutputsForOutputNode(
	ctx context.Context,
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	imageID imagegraph.ImageID,
	outputName imagegraph.OutputName,
) error {
	originalImage, format, err := ig.loadImage(imageID)

	if err != nil {
		return err
	}

	err = ig.saveAndSetPreview(ctx, imageGraphID, nodeID, originalImage, format)

	if err != nil {
		return fmt.Errorf("could not generate outputs for output node: %w", err)
	}

	err = ig.saveAndSetOutput(ctx, imageGraphID, nodeID, outputName, originalImage, format)

	if err != nil {
		return fmt.Errorf("could not generate outputs for output node: %w", err)
	}

	return nil
}

func (ig *ImageGen) GenerateOutputsForPixelInflateNode(
	ctx context.Context,
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	inputImageID imagegraph.ImageID,
	width int,
	lineWidth int,
	lineColor string,
	outputName imagegraph.OutputName,
) error {
	// Load the input image
	img, format, err := ig.loadImage(inputImageID)
	if err != nil {
		return err
	}

	// Get original dimensions
	bounds := img.Bounds()
	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()

	// Calculate new height maintaining aspect ratio
	targetWidth := uint(width)
	targetHeight := uint(float64(width) * float64(originalHeight) / float64(originalWidth))

	// Scale the image using NearestNeighbor to preserve pixel appearance
	scaledImg := resize.Resize(targetWidth, targetHeight, img, resize.NearestNeighbor)

	// Create a mutable RGBA image from the scaled image
	scaledBounds := scaledImg.Bounds()
	outputImg := image.NewRGBA(scaledBounds)
	for y := scaledBounds.Min.Y; y < scaledBounds.Max.Y; y++ {
		for x := scaledBounds.Min.X; x < scaledBounds.Max.X; x++ {
			outputImg.Set(x, y, scaledImg.At(x, y))
		}
	}

	// Parse hex color #RRGGBB
	var r, g, b uint8
	fmt.Sscanf(lineColor, "#%02x%02x%02x", &r, &g, &b)
	lineCol := color.RGBA{R: r, G: g, B: b, A: 255}

	// Calculate scale factor
	scaleX := float64(targetWidth) / float64(originalWidth)
	scaleY := float64(targetHeight) / float64(originalHeight)

	// Draw vertical lines (delineating original pixel columns)
	for i := 1; i < originalWidth; i++ {
		x := int(float64(i) * scaleX)
		for lineOffset := 0; lineOffset < lineWidth; lineOffset++ {
			xPos := x + lineOffset - lineWidth/2
			if xPos >= 0 && xPos < int(targetWidth) {
				for y := 0; y < int(targetHeight); y++ {
					outputImg.Set(xPos, y, lineCol)
				}
			}
		}
	}

	// Draw horizontal lines (delineating original pixel rows)
	for i := 1; i < originalHeight; i++ {
		y := int(float64(i) * scaleY)
		for lineOffset := 0; lineOffset < lineWidth; lineOffset++ {
			yPos := y + lineOffset - lineWidth/2
			if yPos >= 0 && yPos < int(targetHeight) {
				for x := 0; x < int(targetWidth); x++ {
					outputImg.Set(x, yPos, lineCol)
				}
			}
		}
	}

	err = ig.saveAndSetPreview(ctx, imageGraphID, nodeID, outputImg, format)

	if err != nil {
		return fmt.Errorf("could not generate outputs for pixel inflate node: %w", err)
	}

	err = ig.saveAndSetOutput(ctx, imageGraphID, nodeID, outputName, outputImg, format)

	if err != nil {
		return fmt.Errorf("could not generate outputs for pixel inflate node: %w", err)
	}

	return nil
}
