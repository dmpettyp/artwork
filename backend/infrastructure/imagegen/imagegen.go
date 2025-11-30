package imagegen

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"log/slog"
	"math"
	"math/rand"
	"sort"
	"strings"

	"github.com/anthonynsimon/bild/blur"
	"github.com/dmpettyp/artwork/domain/imagegraph"
	"github.com/nfnt/resize"
)

type imageStorage interface {
	Save(imageID imagegraph.ImageID, imageData []byte) error
	Get(imageID imagegraph.ImageID) ([]byte, error)
}

type nodeUpdater interface {
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

	SetNodeConfig(
		ctx context.Context,
		imageGraphID imagegraph.ImageGraphID,
		nodeID imagegraph.NodeID,
		config imagegraph.NodeConfig,
	) error
}

type ImageGen struct {
	imageStorage imageStorage
	nodeUpdater  nodeUpdater
	logger       *slog.Logger
}

func NewImageGen(
	imageStorage imageStorage,
	nodeUpdater nodeUpdater,
	logger *slog.Logger,
) *ImageGen {
	if logger == nil {
		logger = slog.Default()
	}

	return &ImageGen{
		imageStorage: imageStorage,
		nodeUpdater:  nodeUpdater,
		logger:       logger,
	}
}

func (ig *ImageGen) logGeneration(
	nodeType string,
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	nodeVersion imagegraph.NodeVersion,
	attrs ...any,
) {
	if ig.logger == nil {
		return
	}

	args := []any{
		"node_type", nodeType,
		"graph_id", imageGraphID.String(),
		"node_id", nodeID.String(),
		"node_version", int64(nodeVersion),
	}
	args = append(args, attrs...)
	ig.logger.Info("generate_node", args...)
}

func (ig *ImageGen) encodeImage(img image.Image) ([]byte, error) {
	var buf bytes.Buffer

	err := png.Encode(&buf, img)

	if err != nil {
		return nil, fmt.Errorf("could not encode image: %w", err)
	}

	return buf.Bytes(), nil
}

func (ig *ImageGen) loadImage(imageID imagegraph.ImageID) (image.Image, error) {
	imageData, err := ig.imageStorage.Get(imageID)

	if err != nil {
		return nil, fmt.Errorf("could not get image: %w", err)
	}

	img, _, err := image.Decode(bytes.NewReader(imageData))

	if err != nil {
		return nil, fmt.Errorf("could not decode image: %w", err)
	}

	return img, nil
}

// saveAndSetOutput encodes an image, saves it to storage, and sets it as a node output
func (ig *ImageGen) saveAndSetOutput(
	ctx context.Context,
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	outputName imagegraph.OutputName,
	img image.Image,
) error {
	// Encode the image
	imageData, err := ig.encodeImage(img)
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
	err = ig.nodeUpdater.SetNodeOutputImage(ctx, imageGraphID, nodeID, outputName, outputImageID)
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
) error {
	bounds := img.Bounds()
	width := uint(bounds.Dx())
	height := uint(bounds.Dy())

	interpolationFunction := resize.Lanczos2

	if width < 300 || height < 300 {
		interpolationFunction = resize.NearestNeighbor
	}

	if width > height {
		width = 300
		height = 0
	} else {
		width = 0
		height = 300
	}

	previewImg := resize.Resize(width, height, img, interpolationFunction)

	imageData, err := ig.encodeImage(previewImg)

	if err != nil {
		return err
	}

	previewImageID, err := imagegraph.NewImageID()

	if err != nil {
		return fmt.Errorf("could not generate preview image ID: %w", err)
	}

	err = ig.imageStorage.Save(previewImageID, imageData)

	if err != nil {
		return fmt.Errorf("could not save preview image: %w", err)
	}

	err = ig.nodeUpdater.SetNodePreviewImage(ctx, imageGraphID, nodeID, previewImageID)

	if err != nil {
		return fmt.Errorf("could not set node preview image: %w", err)
	}

	return nil
}

func (ig *ImageGen) GeneratePreviewForInputNode(
	ctx context.Context,
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	outputImageID imagegraph.ImageID,
) error {
	// Load the input image
	outputImage, err := ig.loadImage(outputImageID)
	if err != nil {
		return err
	}

	err = ig.saveAndSetPreview(ctx, imageGraphID, nodeID, outputImage)

	if err != nil {
		return fmt.Errorf("could not generate outputs for blur node: %w", err)
	}

	return nil
}

func (ig *ImageGen) GenerateOutputsForBlurNode(
	ctx context.Context,
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	nodeVersion imagegraph.NodeVersion,
	inputImageID imagegraph.ImageID,
	radius int,
) error {
	ig.logGeneration("blur", imageGraphID, nodeID, nodeVersion, "radius", radius)

	// Load the input image
	img, err := ig.loadImage(inputImageID)
	if err != nil {
		return err
	}

	blurredImg := blur.Gaussian(img, float64(radius))

	err = ig.saveAndSetPreview(ctx, imageGraphID, nodeID, blurredImg)

	if err != nil {
		return fmt.Errorf("could not generate outputs for blur node: %w", err)
	}

	err = ig.saveAndSetOutput(ctx, imageGraphID, nodeID, "blurred", blurredImg)

	if err != nil {
		return fmt.Errorf("could not generate outputs for blur node: %w", err)
	}

	return nil
}

func (ig *ImageGen) GenerateOutputsForResizeNode(
	ctx context.Context,
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	nodeVersion imagegraph.NodeVersion,
	inputImageID imagegraph.ImageID,
	width *int,
	height *int,
	interpolation string,
) error {
	ig.logGeneration("resize", imageGraphID, nodeID, nodeVersion,
		"width", width,
		"height", height,
		"interpolation", interpolation,
	)

	// Load the input image
	img, err := ig.loadImage(inputImageID)
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

	err = ig.saveAndSetPreview(ctx, imageGraphID, nodeID, resizedImg)

	if err != nil {
		return fmt.Errorf("could not generate outputs for resize node: %w", err)
	}

	err = ig.saveAndSetOutput(ctx, imageGraphID, nodeID, "resized", resizedImg)

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
	nodeVersion imagegraph.NodeVersion,
	originalImageID imagegraph.ImageID,
	sizeMatchImageID imagegraph.ImageID,
	interpolation string,
) error {
	ig.logGeneration("resize_match", imageGraphID, nodeID, nodeVersion,
		"interpolation", interpolation,
	)

	// Load the original image
	originalImg, err := ig.loadImage(originalImageID)
	if err != nil {
		return err
	}

	// Load the size_match image to get dimensions
	sizeMatchImg, err := ig.loadImage(sizeMatchImageID)
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

	err = ig.saveAndSetPreview(ctx, imageGraphID, nodeID, resizedImg)

	if err != nil {
		return fmt.Errorf("could not generate outputs for resize match node: %w", err)
	}

	err = ig.saveAndSetOutput(ctx, imageGraphID, nodeID, "resized", resizedImg)

	if err != nil {
		return fmt.Errorf("could not generate outputs for resize match node: %w", err)
	}

	return nil
}

// createCropPreviewImage creates a preview image showing the crop region overlay
func (ig *ImageGen) createCropPreviewImage(originalImage image.Image, left, top, right, bottom int) image.Image {
	bounds := originalImage.Bounds()

	// Create a new RGBA image
	previewImg := image.NewRGBA(bounds)

	// Copy original image to preview
	draw.Draw(previewImg, bounds, originalImage, bounds.Min, draw.Src)

	// Define overlay color (semi-transparent black)
	overlayColor := color.RGBA{R: 0, G: 0, B: 0, A: 128}

	// Draw semi-transparent overlay on areas outside crop region
	// Top rectangle
	if top > bounds.Min.Y {
		topRect := image.Rect(bounds.Min.X, bounds.Min.Y, bounds.Max.X, top)
		draw.Draw(previewImg, topRect, &image.Uniform{overlayColor}, image.Point{}, draw.Over)
	}

	// Bottom rectangle
	if bottom < bounds.Max.Y {
		bottomRect := image.Rect(bounds.Min.X, bottom, bounds.Max.X, bounds.Max.Y)
		draw.Draw(previewImg, bottomRect, &image.Uniform{overlayColor}, image.Point{}, draw.Over)
	}

	// Left rectangle (only the crop region height to avoid overlapping corners)
	if left > bounds.Min.X {
		leftRect := image.Rect(bounds.Min.X, top, left, bottom)
		draw.Draw(previewImg, leftRect, &image.Uniform{overlayColor}, image.Point{}, draw.Over)
	}

	// Right rectangle (only the crop region height to avoid overlapping corners)
	if right < bounds.Max.X {
		rightRect := image.Rect(right, top, bounds.Max.X, bottom)
		draw.Draw(previewImg, rightRect, &image.Uniform{overlayColor}, image.Point{}, draw.Over)
	}

	// Draw white border around crop rectangle
	borderColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	borderWidth := int(float64(bounds.Dx()) * 0.02)

	// Draw border by drawing thick lines on each side of the crop rectangle
	for offset := range borderWidth {
		// Top border
		for x := left; x < right; x++ {
			y := top + offset
			if y >= bounds.Min.Y && y < bounds.Max.Y && x >= bounds.Min.X && x < bounds.Max.X {
				previewImg.Set(x, y, borderColor)
			}
		}

		// Bottom border
		for x := left; x < right; x++ {
			y := bottom - offset - 1
			if y >= bounds.Min.Y && y < bounds.Max.Y && x >= bounds.Min.X && x < bounds.Max.X {
				previewImg.Set(x, y, borderColor)
			}
		}

		// Left border
		for y := top; y < bottom; y++ {
			x := left + offset
			if x >= bounds.Min.X && x < bounds.Max.X && y >= bounds.Min.Y && y < bounds.Max.Y {
				previewImg.Set(x, y, borderColor)
			}
		}

		// Right border
		for y := top; y < bottom; y++ {
			x := right - offset - 1
			if x >= bounds.Min.X && x < bounds.Max.X && y >= bounds.Min.Y && y < bounds.Max.Y {
				previewImg.Set(x, y, borderColor)
			}
		}
	}

	return previewImg
}

func (ig *ImageGen) GenerateOutputsForCropNode(
	ctx context.Context,
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	nodeVersion imagegraph.NodeVersion,
	imageID imagegraph.ImageID,
	left, right, top, bottom *int,
) error {
	ig.logGeneration("crop", imageGraphID, nodeID, nodeVersion,
		"left", left,
		"right", right,
		"top", top,
		"bottom", bottom,
	)

	originalImage, err := ig.loadImage(imageID)

	if err != nil {
		return err
	}

	bounds := originalImage.Bounds()

	// If no crop bounds are provided, pass through the original image
	if left == nil && right == nil && top == nil && bottom == nil {
		err = ig.saveAndSetPreview(ctx, imageGraphID, nodeID, originalImage)

		if err != nil {
			return fmt.Errorf("could not generate outputs for crop node: %w", err)
		}

		err = ig.saveAndSetOutput(ctx, imageGraphID, nodeID, "cropped", originalImage)

		if err != nil {
			return fmt.Errorf("could not generate outputs for crop node: %w", err)
		}

		return nil
	}

	// Fill in missing bounds with defaults based on image dimensions
	actualLeft := bounds.Min.X
	actualRight := bounds.Max.X
	actualTop := bounds.Min.Y
	actualBottom := bounds.Max.Y

	if left != nil {
		actualLeft = *left
	}
	if right != nil {
		actualRight = *right
	}
	if top != nil {
		actualTop = *top
	}
	if bottom != nil {
		actualBottom = *bottom
	}

	// Clamp crop coordinates to actual image bounds
	if actualLeft < bounds.Min.X {
		actualLeft = bounds.Min.X
	}
	if actualRight > bounds.Max.X {
		actualRight = bounds.Max.X
	}
	if actualTop < bounds.Min.Y {
		actualTop = bounds.Min.Y
	}
	if actualBottom > bounds.Max.Y {
		actualBottom = bounds.Max.Y
	}

	// Ensure we still have a valid rectangle after clamping
	if actualLeft >= actualRight || actualTop >= actualBottom {
		return fmt.Errorf("crop rectangle is invalid or outside image bounds")
	}

	// Create the crop rectangle
	cropRect := image.Rect(actualLeft, actualTop, actualRight, actualBottom)

	// Create a sub-image (this is a view, not a copy)
	var croppedImg image.Image
	if subImager, ok := originalImage.(interface {
		SubImage(r image.Rectangle) image.Image
	}); ok {
		croppedImg = subImager.SubImage(cropRect)
	} else {
		return fmt.Errorf("image type does not support cropping")
	}

	// Generate preview with crop overlay visualization
	previewImg := ig.createCropPreviewImage(originalImage, actualLeft, actualTop, actualRight, actualBottom)

	err = ig.saveAndSetPreview(ctx, imageGraphID, nodeID, previewImg)

	if err != nil {
		return fmt.Errorf("could not generate outputs for crop node: %w", err)
	}

	err = ig.saveAndSetOutput(ctx, imageGraphID, nodeID, "cropped", croppedImg)

	if err != nil {
		return fmt.Errorf("could not generate outputs for crop node: %w", err)
	}

	return nil
}

func (ig *ImageGen) GenerateOutputsForOutputNode(
	ctx context.Context,
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	nodeVersion imagegraph.NodeVersion,
	imageID imagegraph.ImageID,
) error {
	ig.logGeneration("output", imageGraphID, nodeID, nodeVersion)

	originalImage, err := ig.loadImage(imageID)

	if err != nil {
		return err
	}

	err = ig.saveAndSetPreview(ctx, imageGraphID, nodeID, originalImage)

	if err != nil {
		return fmt.Errorf("could not generate outputs for output node: %w", err)
	}

	err = ig.saveAndSetOutput(ctx, imageGraphID, nodeID, "final", originalImage)

	if err != nil {
		return fmt.Errorf("could not generate outputs for output node: %w", err)
	}

	return nil
}

func (ig *ImageGen) GenerateOutputsForPixelInflateNode(
	ctx context.Context,
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	nodeVersion imagegraph.NodeVersion,
	inputImageID imagegraph.ImageID,
	width int,
	lineWidth int,
	lineColor string,
) error {
	ig.logGeneration("pixel_inflate", imageGraphID, nodeID, nodeVersion,
		"width", width,
		"line_width", lineWidth,
		"line_color", lineColor,
	)

	// Load the input image
	img, err := ig.loadImage(inputImageID)
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
	for i := range originalWidth - 1 {
		x := int(float64(i+1) * scaleX)
		for lineOffset := range lineWidth {
			xPos := x + lineOffset - lineWidth/2
			if xPos >= 0 && xPos < int(targetWidth) {
				for y := range int(targetHeight) {
					outputImg.Set(xPos, y, lineCol)
				}
			}
		}
	}

	// Draw horizontal lines (delineating original pixel rows)
	for i := range originalHeight - 1 {
		y := int(float64(i+1) * scaleY)
		for lineOffset := range lineWidth {
			yPos := y + lineOffset - lineWidth/2
			if yPos >= 0 && yPos < int(targetHeight) {
				for x := range int(targetWidth) {
					outputImg.Set(x, yPos, lineCol)
				}
			}
		}
	}

	err = ig.saveAndSetPreview(ctx, imageGraphID, nodeID, outputImg)

	if err != nil {
		return fmt.Errorf("could not generate outputs for pixel inflate node: %w", err)
	}

	err = ig.saveAndSetOutput(ctx, imageGraphID, nodeID, "inflated", outputImg)

	if err != nil {
		return fmt.Errorf("could not generate outputs for pixel inflate node: %w", err)
	}

	return nil
}

func (ig *ImageGen) GenerateOutputsForPaletteExtractNode(
	ctx context.Context,
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	nodeVersion imagegraph.NodeVersion,
	sourceImageID imagegraph.ImageID,
	numColors int,
	method string,
) error {
	ig.logGeneration("palette_extract", imageGraphID, nodeID, nodeVersion,
		"method", method,
		"num_colors", numColors,
	)

	// Load source image
	sourceImg, err := ig.loadImage(sourceImageID)
	if err != nil {
		return err
	}

	var palette []color.Color
	switch method {
	case "dominant_frequency":
		palette = mostCommonColors(sourceImg, numColors)
	default: // "oklab_clusters" and fallback
		// Extract colors from the image (ignoring alpha)
		colors := extractColorsFromImage(sourceImg)
		palette = kmeansClusteringOKLab(colors, numColors)
	}

	// No sorting - use colors as returned by clustering

	paletteImg := createPaletteImage(palette)

	err = ig.saveAndSetPreview(ctx, imageGraphID, nodeID, paletteImg)
	if err != nil {
		return fmt.Errorf("could not generate outputs for palette extract node: %w", err)
	}

	err = ig.saveAndSetOutput(ctx, imageGraphID, nodeID, "palette", paletteImg)
	if err != nil {
		return fmt.Errorf("could not generate outputs for palette extract node: %w", err)
	}

	return nil
}

func (ig *ImageGen) GenerateOutputsForPaletteApplyNode(
	ctx context.Context,
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	nodeVersion imagegraph.NodeVersion,
	sourceImageID imagegraph.ImageID,
	paletteImageID imagegraph.ImageID,
	config *imagegraph.NodeConfigPaletteApply,
) error {
	normalizeMode := ""
	if config != nil {
		normalizeMode = config.Normalize
	}
	ig.logGeneration("palette_apply", imageGraphID, nodeID, nodeVersion,
		"normalize", normalizeMode,
	)

	// Load source image
	sourceImg, err := ig.loadImage(sourceImageID)
	if err != nil {
		return err
	}

	// Load palette image
	paletteImg, err := ig.loadImage(paletteImageID)
	if err != nil {
		return err
	}

	// Extract palette colors (all non-transparent unique colors)
	paletteColors := extractPaletteColors(paletteImg)

	if len(paletteColors) == 0 {
		return fmt.Errorf("palette image contains no colors")
	}

	// Normalize palette lightness if requested
	if config != nil && config.Normalize == "lightness" {
		paletteColors = normalizePaletteLightness(paletteColors)
	}

	// Map source image to palette
	outputImg := mapImageToPalette(sourceImg, paletteColors)

	// Save preview
	err = ig.saveAndSetPreview(ctx, imageGraphID, nodeID, outputImg)
	if err != nil {
		return fmt.Errorf("could not generate outputs for palette apply node: %w", err)
	}

	// Save output
	err = ig.saveAndSetOutput(ctx, imageGraphID, nodeID, "mapped", outputImg)
	if err != nil {
		return fmt.Errorf("could not generate outputs for palette apply node: %w", err)
	}

	return nil
}

func (ig *ImageGen) GenerateOutputsForPaletteCreateNode(
	ctx context.Context,
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	nodeVersion imagegraph.NodeVersion,
	colorStrings []string,
) error {
	ig.logGeneration("palette_create", imageGraphID, nodeID, nodeVersion,
		"colors_count", len(colorStrings),
	)

	colors := make([]color.Color, 0, len(colorStrings))
	for _, hex := range colorStrings {
		col, err := parseHexColor(hex)
		if err != nil {
			return fmt.Errorf("invalid color %q: %w", hex, err)
		}
		colors = append(colors, col)
	}

	paletteImg := createPaletteImage(colors)

	if err := ig.saveAndSetPreview(ctx, imageGraphID, nodeID, paletteImg); err != nil {
		return fmt.Errorf("could not generate palette create preview: %w", err)
	}

	if err := ig.saveAndSetOutput(ctx, imageGraphID, nodeID, "palette", paletteImg); err != nil {
		return fmt.Errorf("could not generate palette create output: %w", err)
	}

	return nil
}

func (ig *ImageGen) GenerateOutputsForPaletteEditNode(
	ctx context.Context,
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	nodeVersion imagegraph.NodeVersion,
	sourceImageID imagegraph.ImageID,
	existingColors []string,
	currentConfig string,
) error {
	ig.logGeneration("palette_edit", imageGraphID, nodeID, nodeVersion,
		"existing_colors", len(existingColors),
	)

	// Load source image
	sourceImg, err := ig.loadImage(sourceImageID)
	if err != nil {
		return err
	}

	extracted := extractColorsFromImage(sourceImg)
	if len(extracted) > 100 {
		return fmt.Errorf("palette edit: source image contains more than 100 unique colors")
	}

	// Map existing colors (with disabled flag)
	existingMap := make(map[string]bool)
	disabledMap := make(map[string]bool)
	for _, raw := range existingColors {
		base := strings.TrimPrefix(raw, "!")
		existingMap[base] = true
		if strings.HasPrefix(raw, "!") {
			disabledMap[base] = true
		}
	}

	// Add extracted colors if not present
	for _, c := range extracted {
		hex := colorToHex(c)
		if _, ok := existingMap[hex]; ok {
			continue
		}
		existingMap[hex] = true
	}

	// Build combined list with disabled flags
	combined := make([]string, 0, len(existingMap))
	for colorHex := range existingMap {
		if disabledMap[colorHex] {
			combined = append(combined, "!"+colorHex)
		} else {
			combined = append(combined, colorHex)
		}
	}

	// Sort deterministically
	sort.SliceStable(combined, func(i, j int) bool {
		ci, _ := parseHexColor(strings.TrimPrefix(combined[i], "!"))
		cj, _ := parseHexColor(strings.TrimPrefix(combined[j], "!"))
		return lessByLuminanceHue(ci, cj)
	})

	// Build enabled palette image
	enabledColors := make([]color.Color, 0, len(combined))
	for _, raw := range combined {
		if strings.HasPrefix(raw, "!") {
			continue
		}
		col, _ := parseHexColor(raw)
		enabledColors = append(enabledColors, col)
	}

	paletteImg := createPaletteImage(enabledColors)

	// Update config (only if changed to avoid loops)
	newConfigStr := strings.Join(combined, ",")
	if newConfigStr != currentConfig {
		cfg := imagegraph.NewNodeConfigPaletteEdit()
		cfg.Colors = newConfigStr
		if err := ig.nodeUpdater.SetNodeConfig(ctx, imageGraphID, nodeID, cfg); err != nil {
			return fmt.Errorf("could not update palette edit config: %w", err)
		}
	}

	if err := ig.saveAndSetPreview(ctx, imageGraphID, nodeID, paletteImg); err != nil {
		return fmt.Errorf("could not generate palette edit preview: %w", err)
	}

	if err := ig.saveAndSetOutput(ctx, imageGraphID, nodeID, "palette", paletteImg); err != nil {
		return fmt.Errorf("could not generate palette edit output: %w", err)
	}

	return nil
}

// extractPaletteColors extracts all non-transparent unique colors from a palette image
func extractPaletteColors(img image.Image) []color.Color {
	bounds := img.Bounds()
	colorMap := make(map[uint32]color.Color)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			r, g, b, a := c.RGBA()

			// Skip transparent pixels
			if a>>8 == 0 {
				continue
			}

			// Convert to 8-bit
			r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)
			key := uint32(r8)<<16 | uint32(g8)<<8 | uint32(b8)
			colorMap[key] = color.RGBA{R: r8, G: g8, B: b8, A: 255}
		}
	}

	// Convert map to slice
	colors := make([]color.Color, 0, len(colorMap))
	for _, c := range colorMap {
		colors = append(colors, c)
	}

	return colors
}

// mapImageToPalette maps each pixel in the source image to the nearest color in the palette
func mapImageToPalette(sourceImg image.Image, palette []color.Color) image.Image {
	bounds := sourceImg.Bounds()
	outputImg := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			sourceColor := sourceImg.At(x, y)
			nearestColor := findNearestColor(sourceColor, palette)
			outputImg.Set(x, y, nearestColor)
		}
	}

	return outputImg
}

// normalizePaletteLightness scales palette colors in OKLab so the lightness range spans [0,1].
func normalizePaletteLightness(palette []color.Color) []color.Color {
	if len(palette) == 0 {
		return palette
	}

	minL := math.MaxFloat64
	maxL := -math.MaxFloat64
	labs := make([][3]float64, len(palette))

	for i, c := range palette {
		l, a, b := rgbToOKLab(c)
		labs[i] = [3]float64{l, a, b}
		if l < minL {
			minL = l
		}
		if l > maxL {
			maxL = l
		}
	}

	if maxL <= minL {
		return palette
	}

	scaled := make([]color.Color, len(palette))
	for i, lab := range labs {
		lNorm := (lab[0] - minL) / (maxL - minL)
		scaled[i] = okLabToRGBA(lNorm, lab[1], lab[2])
	}
	return scaled
}

// findNearestColor finds the nearest color in the palette using Euclidean distance in RGB space
func findNearestColor(c color.Color, palette []color.Color) color.Color {
	r1, g1, b1, _ := c.RGBA()
	r1_8, g1_8, b1_8 := float64(r1>>8), float64(g1>>8), float64(b1>>8)

	minDist := float64(1000000)
	var nearestColor color.Color = palette[0]

	for _, pc := range palette {
		r2, g2, b2, _ := pc.RGBA()
		r2_8, g2_8, b2_8 := float64(r2>>8), float64(g2>>8), float64(b2>>8)

		// Euclidean distance in RGB space
		dr := r1_8 - r2_8
		dg := g1_8 - g2_8
		db := b1_8 - b2_8
		dist := dr*dr + dg*dg + db*db

		if dist < minDist {
			minDist = dist
			nearestColor = pc
		}
	}

	return nearestColor
}

// extractColorsFromImage extracts all unique RGB colors from an image
func extractColorsFromImage(img image.Image) []color.Color {
	bounds := img.Bounds()
	colorMap := make(map[uint32]color.Color)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			r, g, b, _ := c.RGBA()
			// Convert to 8-bit and ignore alpha
			r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)
			key := uint32(r8)<<16 | uint32(g8)<<8 | uint32(b8)
			colorMap[key] = color.RGBA{R: r8, G: g8, B: b8, A: 255}
		}
	}

	// Convert map to slice
	colors := make([]color.Color, 0, len(colorMap))
	for _, c := range colorMap {
		colors = append(colors, c)
	}

	return colors
}

// mostCommonColors returns the top-k most frequent colors in an image (alpha ignored)
func mostCommonColors(img image.Image, k int) []color.Color {
	if k <= 0 {
		return []color.Color{}
	}

	// Colors within this OKLab distance are considered duplicates
	const proximityThreshold = 0.01

	bounds := img.Bounds()
	colorCounts := make(map[uint32]int)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			r, g, b, _ := c.RGBA()
			// Convert to 8-bit and ignore alpha
			r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)
			key := uint32(r8)<<16 | uint32(g8)<<8 | uint32(b8)
			colorCounts[key]++
		}
	}

	type colorCount struct {
		key   uint32
		count int
	}

	sorted := make([]colorCount, 0, len(colorCounts))
	for key, count := range colorCounts {
		sorted = append(sorted, colorCount{key: key, count: count})
	}

	// Sort by frequency (desc), then by key for determinism
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].count == sorted[j].count {
			return sorted[i].key < sorted[j].key
		}
		return sorted[i].count > sorted[j].count
	})

	if k > len(sorted) {
		k = len(sorted)
	}

	// Deduplicate visually-close colors in frequency order
	type labColor struct {
		col color.Color
		lab [3]float64
	}

	selected := make([]labColor, 0, k)
	for _, entry := range sorted {
		if len(selected) >= k {
			break
		}

		c := color.RGBA{
			R: uint8(entry.key >> 16),
			G: uint8((entry.key >> 8) & 0xFF),
			B: uint8(entry.key & 0xFF),
			A: 255,
		}
		l, a, b := rgbToOKLab(c)

		tooClose := false
		for _, chosen := range selected {
			dl := chosen.lab[0] - l
			da := chosen.lab[1] - a
			db := chosen.lab[2] - b
			if dl*dl+da*da+db*db < proximityThreshold*proximityThreshold {
				tooClose = true
				break
			}
		}

		if !tooClose {
			selected = append(selected, labColor{col: c, lab: [3]float64{l, a, b}})
		}
	}

	// Order visually: luminance/hue only for a pleasing, stable layout
	sort.SliceStable(selected, func(i, j int) bool {
		return lessByLuminanceHue(selected[i].col, selected[j].col)
	})

	palette := make([]color.Color, 0, k)
	for _, entry := range selected {
		palette = append(palette, entry.col)
	}

	return palette
}

type labColor struct {
	l, a, b float64
	src     color.Color
}

// createPaletteImage creates a near-square image from palette colors
func createPaletteImage(colors []color.Color) image.Image {
	if len(colors) == 0 {
		// Return a 1x1 transparent image if no colors
		img := image.NewRGBA(image.Rect(0, 0, 1, 1))
		return img
	}

	// Calculate near-square dimensions
	numColors := len(colors)
	width := int(math.Ceil(math.Sqrt(float64(numColors))))
	height := (numColors + width - 1) / width // Ceiling division

	// Create image
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with colors
	idx := 0
	for y := range height {
		for x := range width {
			if idx < len(colors) {
				img.Set(x, y, colors[idx])
				idx++
			} else {
				// Fill remaining with transparent pixels
				img.Set(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 0})
			}
		}
	}

	return img
}

func parseHexColor(hex string) (color.Color, error) {
	var r, g, b uint8
	if _, err := fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b); err != nil {
		return nil, fmt.Errorf("failed to parse hex color: %w", err)
	}
	return color.RGBA{R: r, G: g, B: b, A: 255}, nil
}

func colorToHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", uint8(r>>8), uint8(g>>8), uint8(b>>8))
}

func lessByLuminanceHue(a, b color.Color) bool {
	la, aa, ba := rgbToOKLab(a)
	lb, ab, bb := rgbToOKLab(b)
	if la == lb {
		ha := math.Atan2(aa, ba)
		hb := math.Atan2(ab, bb)
		return ha < hb
	}
	return la < lb
}

// kmeansClusteringOKLab performs k-means clustering in OKLab space for better perceptual grouping.
func kmeansClusteringOKLab(colors []color.Color, k int) []color.Color {
	if len(colors) == 0 {
		return []color.Color{}
	}

	if len(colors) <= k {
		return colors
	}

	labColors := make([]labColor, len(colors))
	for i, c := range colors {
		l, a, b := rgbToOKLab(c)
		labColors[i] = labColor{l: l, a: a, b: b, src: c}
	}

	rng := rand.New(rand.NewSource(42))

	bestPalette := make([]color.Color, k)
	bestInertia := math.MaxFloat64

	const maxIterations = 30
	const restarts = 3

	for range restarts {
		centroids := initCentroidsKMeansPP(labColors, k, rng)
		assignments := make([]int, len(labColors))

		for range maxIterations {
			changed := false

			for i, lc := range labColors {
				minDist := math.MaxFloat64
				best := 0
				for j, c := range centroids {
					dl := lc.l - c[0]
					da := lc.a - c[1]
					db := lc.b - c[2]
					dist := dl*dl + da*da + db*db
					if dist < minDist {
						minDist = dist
						best = j
					}
				}
				if assignments[i] != best {
					assignments[i] = best
					changed = true
				}
			}

			newCentroids := make([][3]float64, k)
			counts := make([]int, k)
			for i, lc := range labColors {
				cluster := assignments[i]
				newCentroids[cluster][0] += lc.l
				newCentroids[cluster][1] += lc.a
				newCentroids[cluster][2] += lc.b
				counts[cluster]++
			}

			for i := range counts {
				if counts[i] > 0 {
					newCentroids[i][0] /= float64(counts[i])
					newCentroids[i][1] /= float64(counts[i])
					newCentroids[i][2] /= float64(counts[i])
				} else {
					idx := i % len(labColors)
					newCentroids[i] = [3]float64{labColors[idx].l, labColors[idx].a, labColors[idx].b}
				}
			}

			centroids = newCentroids

			if !changed {
				break
			}
		}

		inertia := 0.0
		for i, lc := range labColors {
			c := centroids[assignments[i]]
			dl := lc.l - c[0]
			da := lc.a - c[1]
			db := lc.b - c[2]
			inertia += dl*dl + da*da + db*db
		}

		if inertia < bestInertia {
			bestInertia = inertia
			for i, c := range centroids {
				bestPalette[i] = okLabToRGBA(c[0], c[1], c[2])
			}
		}
	}

	sort.SliceStable(bestPalette, func(i, j int) bool {
		li, ai, bi := rgbToOKLab(bestPalette[i])
		lj, aj, bj := rgbToOKLab(bestPalette[j])
		if li == lj {
			hi := math.Atan2(ai, bi)
			hj := math.Atan2(aj, bj)
			return hi < hj
		}
		return li < lj
	})

	return bestPalette
}

// initCentroidsKMeansPP initializes centroids using k-means++ in OKLab space.
func initCentroidsKMeansPP(colors []labColor, k int, rng *rand.Rand) [][3]float64 {
	centroids := make([][3]float64, 0, k)

	first := colors[rng.Intn(len(colors))]
	centroids = append(centroids, [3]float64{first.l, first.a, first.b})

	for len(centroids) < k {
		dists := make([]float64, len(colors))
		sum := 0.0
		for i, c := range colors {
			minDist := math.MaxFloat64
			for _, cent := range centroids {
				dl := c.l - cent[0]
				da := c.a - cent[1]
				db := c.b - cent[2]
				dist := dl*dl + da*da + db*db
				if dist < minDist {
					minDist = dist
				}
			}
			dists[i] = minDist
			sum += minDist
		}

		target := rng.Float64() * sum
		acc := 0.0
		for i, d := range dists {
			acc += d
			if acc >= target {
				c := colors[i]
				centroids = append(centroids, [3]float64{c.l, c.a, c.b})
				break
			}
		}
	}

	return centroids
}

// rgbToOKLab converts an sRGB color to OKLab.
func rgbToOKLab(c color.Color) (float64, float64, float64) {
	r, g, b, _ := c.RGBA()
	rf := srgbToLinear(float64(r) / 65535.0)
	gf := srgbToLinear(float64(g) / 65535.0)
	bf := srgbToLinear(float64(b) / 65535.0)

	l := 0.4122214708*rf + 0.5363325363*gf + 0.0514459929*bf
	m := 0.2119034982*rf + 0.6806995451*gf + 0.1073969566*bf
	s := 0.0883024619*rf + 0.2817188376*gf + 0.6299787005*bf

	l_ := math.Cbrt(l)
	m_ := math.Cbrt(m)
	s_ := math.Cbrt(s)

	lOK := 0.2104542553*l_ + 0.7936177850*m_ - 0.0040720468*s_
	aOK := 1.9779984951*l_ - 2.4285922050*m_ + 0.4505937099*s_
	bOK := 0.0259040371*l_ + 0.7827717662*m_ - 0.8086757660*s_

	return lOK, aOK, bOK
}

// okLabToRGBA converts OKLab to sRGB and clamps to byte range.
func okLabToRGBA(l, a, b float64) color.Color {
	l_ := l + 0.3963377774*a + 0.2158037573*b
	m_ := l - 0.1055613458*a - 0.0638541728*b
	s_ := l - 0.0894841775*a - 1.2914855480*b

	l3 := l_ * l_ * l_
	m3 := m_ * m_ * m_
	s3 := s_ * s_ * s_

	r := +4.0767416621*l3 - 3.3077115913*m3 + 0.2309699292*s3
	g := -1.2684380046*l3 + 2.6097574011*m3 - 0.3413193965*s3
	bc := -0.0041960863*l3 - 0.7034186147*m3 + 1.7076147010*s3

	return color.RGBA{
		R: floatToByte(linearToSRGB(r)),
		G: floatToByte(linearToSRGB(g)),
		B: floatToByte(linearToSRGB(bc)),
		A: 255,
	}
}

func srgbToLinear(c float64) float64 {
	if c <= 0.04045 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

func linearToSRGB(c float64) float64 {
	if c <= 0.0 {
		return 0.0
	}
	if c >= 1.0 {
		return 1.0
	}
	if c <= 0.0031308 {
		return 12.92 * c
	}
	return 1.055*math.Pow(c, 1.0/2.4) - 0.055
}

func floatToByte(f float64) uint8 {
	if f < 0 {
		f = 0
	}
	if f > 1 {
		f = 1
	}
	return uint8(f * 255.0)
}
