package imagegen

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"math"
	"sort"

	"github.com/anthonynsimon/bild/blur"
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

	imageData, err := ig.encodeImage(previewImg, format)

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

	err = ig.outputSetter.SetNodePreviewImage(ctx, imageGraphID, nodeID, previewImageID)

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
	outputImage, format, err := ig.loadImage(outputImageID)
	if err != nil {
		return err
	}

	err = ig.saveAndSetPreview(ctx, imageGraphID, nodeID, outputImage, format)

	if err != nil {
		return fmt.Errorf("could not generate outputs for blur node: %w", err)
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

	blurredImg := blur.Gaussian(img, float64(radius))

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
	for offset := 0; offset < borderWidth; offset++ {
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
	imageID imagegraph.ImageID,
	left, right, top, bottom *int,
	outputName imagegraph.OutputName,
) error {
	originalImage, format, err := ig.loadImage(imageID)

	if err != nil {
		return err
	}

	bounds := originalImage.Bounds()

	// If no crop bounds are provided, pass through the original image
	if left == nil && right == nil && top == nil && bottom == nil {
		err = ig.saveAndSetPreview(ctx, imageGraphID, nodeID, originalImage, format)

		if err != nil {
			return fmt.Errorf("could not generate outputs for crop node: %w", err)
		}

		err = ig.saveAndSetOutput(ctx, imageGraphID, nodeID, outputName, originalImage, format)

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

	err = ig.saveAndSetPreview(ctx, imageGraphID, nodeID, previewImg, format)

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

func (ig *ImageGen) GenerateOutputsForPaletteExtractNode(
	ctx context.Context,
	imageGraphID imagegraph.ImageGraphID,
	nodeID imagegraph.NodeID,
	sourceImageID imagegraph.ImageID,
	numColors int,
	clusterBy string,
	outputName imagegraph.OutputName,
) error {
	// Load source image
	sourceImg, format, err := ig.loadImage(sourceImageID)
	if err != nil {
		return err
	}

	// Extract colors from the image (ignoring alpha)
	colors := extractColorsFromImage(sourceImg)

	// Apply k-means clustering to get dominant colors
	var palette []color.Color
	if clusterBy == "HSL" {
		palette = kmeansClusteringHSL(colors, numColors)
	} else {
		// Default to RGB clustering
		palette = kmeansClusteringRGB(colors, numColors)
	}

	// Sort by hue
	sortColorsByHue(palette)

	// Create output image with near-square dimensions
	outputImg := createPaletteImage(palette)

	// Save preview
	err = ig.saveAndSetPreview(ctx, imageGraphID, nodeID, outputImg, format)
	if err != nil {
		return fmt.Errorf("could not generate outputs for palette extract node: %w", err)
	}

	// Save output
	err = ig.saveAndSetOutput(ctx, imageGraphID, nodeID, outputName, outputImg, format)
	if err != nil {
		return fmt.Errorf("could not generate outputs for palette extract node: %w", err)
	}

	return nil
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

// kmeansClusteringRGB performs k-means clustering in RGB space to find dominant colors
func kmeansClusteringRGB(colors []color.Color, k int) []color.Color {
	if len(colors) == 0 {
		return []color.Color{}
	}

	// If we have fewer colors than k, return all colors
	if len(colors) <= k {
		return colors
	}

	// Initialize centroids by evenly spacing through sorted colors
	centroids := make([][3]float64, k)
	step := len(colors) / k
	for i := 0; i < k; i++ {
		idx := i * step
		if idx >= len(colors) {
			idx = len(colors) - 1
		}
		r, g, b, _ := colors[idx].RGBA()
		centroids[i] = [3]float64{float64(r >> 8), float64(g >> 8), float64(b >> 8)}
	}

	// Run k-means iterations
	const maxIterations = 20
	for iteration := 0; iteration < maxIterations; iteration++ {
		// Assign colors to nearest centroid
		assignments := make([]int, len(colors))
		for i, c := range colors {
			r, g, b, _ := c.RGBA()
			r8, g8, b8 := float64(r>>8), float64(g>>8), float64(b>>8)

			minDist := float64(1000000)
			bestCluster := 0
			for j, centroid := range centroids {
				dr := r8 - centroid[0]
				dg := g8 - centroid[1]
				db := b8 - centroid[2]
				dist := dr*dr + dg*dg + db*db

				if dist < minDist {
					minDist = dist
					bestCluster = j
				}
			}
			assignments[i] = bestCluster
		}

		// Update centroids
		newCentroids := make([][3]float64, k)
		counts := make([]int, k)

		for i, c := range colors {
			cluster := assignments[i]
			r, g, b, _ := c.RGBA()
			newCentroids[cluster][0] += float64(r >> 8)
			newCentroids[cluster][1] += float64(g >> 8)
			newCentroids[cluster][2] += float64(b >> 8)
			counts[cluster]++
		}

		for i := 0; i < k; i++ {
			if counts[i] > 0 {
				newCentroids[i][0] /= float64(counts[i])
				newCentroids[i][1] /= float64(counts[i])
				newCentroids[i][2] /= float64(counts[i])
			}
		}

		centroids = newCentroids
	}

	// Convert centroids to colors
	result := make([]color.Color, k)
	for i, centroid := range centroids {
		result[i] = color.RGBA{
			R: uint8(centroid[0]),
			G: uint8(centroid[1]),
			B: uint8(centroid[2]),
			A: 255,
		}
	}

	return result
}

// kmeansClusteringHSL performs k-means clustering in HSL space to find perceptually distributed colors
func kmeansClusteringHSL(colors []color.Color, k int) []color.Color {
	if len(colors) == 0 {
		return []color.Color{}
	}

	// If we have fewer colors than k, return all colors
	if len(colors) <= k {
		return colors
	}

	// Convert all colors to HSL
	colorData := make([]colorWithHSL, len(colors))
	for i, c := range colors {
		h, s, l := rgbToHSL(c)
		colorData[i] = colorWithHSL{color: c, h: h, s: s, l: l}
	}

	// Initialize centroids by evenly spacing through hue spectrum
	centroids := make([][3]float64, k)
	for i := 0; i < k; i++ {
		// Distribute evenly across hue (0-360), mid saturation, mid lightness
		centroids[i] = [3]float64{
			float64(i) * 360.0 / float64(k), // Hue evenly distributed
			0.5,                              // Mid saturation
			0.5,                              // Mid lightness
		}
	}

	// Run k-means iterations
	const maxIterations = 20
	for iteration := 0; iteration < maxIterations; iteration++ {
		// Assign colors to nearest centroid in HSL space
		assignments := make([]int, len(colorData))
		for i, cd := range colorData {
			minDist := float64(1000000)
			bestCluster := 0
			for j, centroid := range centroids {
				// Calculate distance in HSL space
				// Hue is circular, so we need to handle wraparound
				dh := math.Abs(cd.h - centroid[0])
				if dh > 180 {
					dh = 360 - dh
				}
				// Weight hue more heavily (scale by 2)
				dh *= 2.0

				ds := cd.s - centroid[1]
				dl := cd.l - centroid[2]
				dist := dh*dh + ds*ds + dl*dl

				if dist < minDist {
					minDist = dist
					bestCluster = j
				}
			}
			assignments[i] = bestCluster
		}

		// Update centroids
		newCentroids := make([][3]float64, k)
		counts := make([]int, k)

		for i, cd := range colorData {
			cluster := assignments[i]
			newCentroids[cluster][0] += cd.h
			newCentroids[cluster][1] += cd.s
			newCentroids[cluster][2] += cd.l
			counts[cluster]++
		}

		for i := 0; i < k; i++ {
			if counts[i] > 0 {
				newCentroids[i][0] /= float64(counts[i])
				newCentroids[i][1] /= float64(counts[i])
				newCentroids[i][2] /= float64(counts[i])
			}
		}

		centroids = newCentroids
	}

	// Convert centroids back to RGB
	result := make([]color.Color, k)
	for i, centroid := range centroids {
		result[i] = hslToRGB(centroid[0], centroid[1], centroid[2])
	}

	return result
}

// hslToRGB converts HSL color to RGB
func hslToRGB(h, s, l float64) color.Color {
	var r, g, b float64

	if s == 0 {
		// Achromatic (gray)
		r, g, b = l, l, l
	} else {
		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q

		r = hueToRGB(p, q, h+120)
		g = hueToRGB(p, q, h)
		b = hueToRGB(p, q, h-120)
	}

	return color.RGBA{
		R: uint8(r * 255),
		G: uint8(g * 255),
		B: uint8(b * 255),
		A: 255,
	}
}

// hueToRGB is a helper for HSL to RGB conversion
func hueToRGB(p, q, t float64) float64 {
	// Normalize hue to 0-360
	for t < 0 {
		t += 360
	}
	for t > 360 {
		t -= 360
	}

	if t < 60 {
		return p + (q-p)*t/60
	}
	if t < 180 {
		return q
	}
	if t < 240 {
		return p + (q-p)*(240-t)/60
	}
	return p
}

// rgbToHSL converts RGB color to HSL
func rgbToHSL(c color.Color) (h, s, l float64) {
	r, g, b, _ := c.RGBA()
	r8, g8, b8 := float64(r>>8)/255.0, float64(g>>8)/255.0, float64(b>>8)/255.0

	max := r8
	if g8 > max {
		max = g8
	}
	if b8 > max {
		max = b8
	}

	min := r8
	if g8 < min {
		min = g8
	}
	if b8 < min {
		min = b8
	}

	l = (max + min) / 2.0

	if max == min {
		h = 0
		s = 0
		return
	}

	d := max - min
	if l > 0.5 {
		s = d / (2.0 - max - min)
	} else {
		s = d / (max + min)
	}

	switch max {
	case r8:
		h = (g8 - b8) / d
		if g8 < b8 {
			h += 6
		}
	case g8:
		h = (b8-r8)/d + 2
	case b8:
		h = (r8-g8)/d + 4
	}

	h *= 60
	return
}

// colorWithHSL holds a color and its HSL values for sorting
type colorWithHSL struct {
	color color.Color
	h, s, l float64
}

// sortColorsByHue sorts colors by their hue value
// Separates grayscale colors and sorts them by lightness at the end
func sortColorsByHue(colors []color.Color) {
	const saturationThreshold = 0.1 // Colors below this saturation are considered grayscale

	// Convert all colors to HSL
	colorData := make([]colorWithHSL, len(colors))
	for i, c := range colors {
		h, s, l := rgbToHSL(c)
		colorData[i] = colorWithHSL{color: c, h: h, s: s, l: l}
	}

	// Sort using standard library
	sort.SliceStable(colorData, func(i, j int) bool {
		c1, c2 := colorData[i], colorData[j]

		isGray1 := c1.s < saturationThreshold
		isGray2 := c2.s < saturationThreshold

		// Grayscale colors go to the end
		if isGray1 != isGray2 {
			return !isGray1 // chromatic colors (false) come before grayscale (true)
		}

		// Both are grayscale - sort by lightness (dark to light)
		if isGray1 && isGray2 {
			return c1.l < c2.l
		}

		// Both are chromatic - sort by hue, then saturation, then lightness
		if c1.h != c2.h {
			return c1.h < c2.h
		}
		if c1.s != c2.s {
			return c1.s > c2.s // Higher saturation first
		}
		return c1.l < c2.l // Darker first
	})

	// Copy sorted colors back
	for i, cd := range colorData {
		colors[i] = cd.color
	}
}

// createPaletteImage creates a near-square image from palette colors
func createPaletteImage(colors []color.Color) image.Image {
	if len(colors) == 0 {
		// Return a 1x1 black image if no colors
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
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
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
