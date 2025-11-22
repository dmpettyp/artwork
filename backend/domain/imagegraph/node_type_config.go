package imagegraph

import (
	"fmt"
	"slices"
)

type NodeConfig interface {
	Validate() error
	NodeType() NodeType
}

func NewNodeConfig(nodeType NodeType) NodeConfig {
	switch nodeType {
	case NodeTypeInput:
		return NewNodeConfigInput()
	case NodeTypeOutput:
		return NewNodeConfigOutput()
	case NodeTypeCrop:
		return NewNodeConfigCrop()
	case NodeTypeBlur:
		return NewNodeConfigBlur()
	case NodeTypeResize:
		return NewNodeConfigResize()
	case NodeTypeResizeMatch:
		return NewNodeConfigResizeMatch()
	case NodeTypePixelInflate:
		return NewNodeConfigPixelInflate()
	case NodeTypePaletteExtract:
		return NewNodeConfigPaletteExtract()
	case NodeTypePaletteApply:
		return NewNodeConfigPaletteApply()
	default:
		return nil
	}
}

// NodeConfigInput is the configuration for input nodes.
type NodeConfigInput struct{}

func NewNodeConfigInput() *NodeConfigInput {
	return &NodeConfigInput{}
}

func (c *NodeConfigInput) Validate() error {
	return nil
}

func (c *NodeConfigInput) NodeType() NodeType {
	return NodeTypeInput
}

// NodeConfigOutput is the configuration for output nodes.
type NodeConfigOutput struct{}

func NewNodeConfigOutput() *NodeConfigOutput {
	return &NodeConfigOutput{}
}

func (c *NodeConfigOutput) Validate() error {
	return nil
}

func (c *NodeConfigOutput) NodeType() NodeType {
	return NodeTypeOutput
}

// NodeConfigCrop is the configuration for crop nodes.
type NodeConfigCrop struct {
	Left              *int `json:"left,omitempty"`
	Right             *int `json:"right,omitempty"`
	Top               *int `json:"top,omitempty"`
	Bottom            *int `json:"bottom,omitempty"`
	AspectRatioWidth  *int `json:"aspect_ratio_width,omitempty"`
	AspectRatioHeight *int `json:"aspect_ratio_height,omitempty"`
}

func NewNodeConfigCrop() *NodeConfigCrop {
	return &NodeConfigCrop{}
}

func (c *NodeConfigCrop) Validate() error {
	// If no bounds are provided at all, this is valid (passthrough mode)
	if c.Left == nil && c.Right == nil && c.Top == nil && c.Bottom == nil {
		return nil
	}

	// Validate provided coordinates are non-negative
	if c.Left != nil && *c.Left < 0 {
		return fmt.Errorf("left coordinate must be non-negative")
	}
	if c.Right != nil && *c.Right < 0 {
		return fmt.Errorf("right coordinate must be non-negative")
	}
	if c.Top != nil && *c.Top < 0 {
		return fmt.Errorf("top coordinate must be non-negative")
	}
	if c.Bottom != nil && *c.Bottom < 0 {
		return fmt.Errorf("bottom coordinate must be non-negative")
	}

	// If both left and right are provided, validate their relationship
	if c.Left != nil && c.Right != nil && *c.Left >= *c.Right {
		return fmt.Errorf("left must be less than right")
	}

	// If both top and bottom are provided, validate their relationship
	if c.Top != nil && c.Bottom != nil && *c.Top >= *c.Bottom {
		return fmt.Errorf("top must be less than bottom")
	}

	// Both must be set together or both omitted
	if (c.AspectRatioWidth != nil) != (c.AspectRatioHeight != nil) {
		return fmt.Errorf("aspect_ratio_width and aspect_ratio_height must both be set or both omitted")
	}

	// Only validate aspect ratio if we have all four bounds and aspect ratio is specified
	if c.AspectRatioWidth != nil && c.AspectRatioHeight != nil &&
		c.Left != nil && c.Right != nil && c.Top != nil && c.Bottom != nil {

		aspectWidth := *c.AspectRatioWidth
		aspectHeight := *c.AspectRatioHeight

		// Both must be positive
		if aspectWidth <= 0 || aspectHeight <= 0 {
			return fmt.Errorf("aspect ratio values must be positive integers")
		}

		// Validate that crop dimensions match the aspect ratio (within rounding tolerance)
		cropWidth := *c.Right - *c.Left
		cropHeight := *c.Bottom - *c.Top

		expectedRatio := float64(aspectWidth) / float64(aspectHeight)
		actualRatio := float64(cropWidth) / float64(cropHeight)

		// Allow 1% tolerance for rounding
		tolerance := 0.01
		if actualRatio < expectedRatio*(1-tolerance) || actualRatio > expectedRatio*(1+tolerance) {
			return fmt.Errorf("crop dimensions (%dx%d) do not match specified aspect ratio (%d:%d)",
				cropWidth, cropHeight, aspectWidth, aspectHeight)
		}
	}

	return nil
}

func (c *NodeConfigCrop) NodeType() NodeType {
	return NodeTypeCrop
}

// NodeConfigBlur is the configuration for blur nodes.
type NodeConfigBlur struct {
	Radius int `json:"radius"`
}

func NewNodeConfigBlur() *NodeConfigBlur {
	return &NodeConfigBlur{Radius: 2}
}

func (c *NodeConfigBlur) Validate() error {
	if c.Radius < 1 {
		return fmt.Errorf("radius must be at least 1")
	}
	if c.Radius > 100 {
		return fmt.Errorf("radius must be 100 or less")
	}
	return nil
}

func (c *NodeConfigBlur) NodeType() NodeType {
	return NodeTypeBlur
}

// NodeConfigResize is the configuration for resize nodes.
type NodeConfigResize struct {
	Width         *int   `json:"width,omitempty"`
	Height        *int   `json:"height,omitempty"`
	Interpolation string `json:"interpolation"`
}

func NewNodeConfigResize() *NodeConfigResize {
	return &NodeConfigResize{}
}

func (c *NodeConfigResize) Validate() error {
	// At least one of width or height must be set
	if c.Width == nil && c.Height == nil {
		return fmt.Errorf("at least one of width or height must be set")
	}

	// Validate width if present
	if c.Width != nil {
		if *c.Width < 1 {
			return fmt.Errorf("width must be at least 1")
		}
		if *c.Width > 10000 {
			return fmt.Errorf("width must be 10000 or less")
		}
	}

	// Validate height if present
	if c.Height != nil {
		if *c.Height < 1 {
			return fmt.Errorf("height must be at least 1")
		}
		if *c.Height > 10000 {
			return fmt.Errorf("height must be 10000 or less")
		}
	}

	// Validate interpolation
	if err := validateInterpolation(c.Interpolation); err != nil {
		return err
	}

	return nil
}

func (c *NodeConfigResize) NodeType() NodeType {
	return NodeTypeResize
}

// NodeConfigResizeMatch is the configuration for resize-match nodes.
type NodeConfigResizeMatch struct {
	Interpolation string `json:"interpolation"`
}

func NewNodeConfigResizeMatch() *NodeConfigResizeMatch {
	return &NodeConfigResizeMatch{}
}

func (c *NodeConfigResizeMatch) Validate() error {
	return validateInterpolation(c.Interpolation)
}

func (c *NodeConfigResizeMatch) NodeType() NodeType {
	return NodeTypeResizeMatch
}

// NodeConfigPixelInflate is the configuration for pixel-inflate nodes.
type NodeConfigPixelInflate struct {
	Width     int    `json:"width"`
	LineWidth int    `json:"line_width"`
	LineColor string `json:"line_color"`
}

func NewNodeConfigPixelInflate() *NodeConfigPixelInflate {
	return &NodeConfigPixelInflate{}
}

func (c *NodeConfigPixelInflate) Validate() error {
	if c.Width < 1 {
		return fmt.Errorf("width must be at least 1")
	}
	if c.Width > 10000 {
		return fmt.Errorf("width must be 10000 or less")
	}

	if c.LineWidth < 1 {
		return fmt.Errorf("line_width must be at least 1")
	}
	if c.LineWidth > 100 {
		return fmt.Errorf("line_width must be 100 or less")
	}

	// Validate hex color format #RRGGBB
	if len(c.LineColor) != 7 || c.LineColor[0] != '#' {
		return fmt.Errorf("line_color must be in #RRGGBB format")
	}
	for i := 1; i < 7; i++ {
		ch := c.LineColor[i]
		if !((ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'F') || (ch >= 'a' && ch <= 'f')) {
			return fmt.Errorf("line_color must be in #RRGGBB format")
		}
	}

	return nil
}

func (c *NodeConfigPixelInflate) NodeType() NodeType {
	return NodeTypePixelInflate
}

// NodeConfigPaletteExtract is the configuration for palette-extract nodes.
type NodeConfigPaletteExtract struct {
	NumColors int    `json:"num_colors"`
	ClusterBy string `json:"cluster_by"`
}

func NewNodeConfigPaletteExtract() *NodeConfigPaletteExtract {
	return &NodeConfigPaletteExtract{
		NumColors: 16,
		ClusterBy: "RGB",
	}
}

func (c *NodeConfigPaletteExtract) Validate() error {
	if c.NumColors < 1 {
		return fmt.Errorf("num_colors must be at least 1")
	}
	if c.NumColors > 1000 {
		return fmt.Errorf("num_colors must be 1000 or less")
	}

	validClusterBy := []string{"RGB", "HSL"}
	if !slices.Contains(validClusterBy, c.ClusterBy) {
		return fmt.Errorf("cluster_by must be one of: %v", validClusterBy)
	}

	return nil
}

func (c *NodeConfigPaletteExtract) NodeType() NodeType {
	return NodeTypePaletteExtract
}

// NodeConfigPaletteApply is the configuration for palette-apply nodes.
type NodeConfigPaletteApply struct{}

func NewNodeConfigPaletteApply() *NodeConfigPaletteApply {
	return &NodeConfigPaletteApply{}
}

func (c *NodeConfigPaletteApply) Validate() error {
	return nil
}

func (c *NodeConfigPaletteApply) NodeType() NodeType {
	return NodeTypePaletteApply
}

// Helper function for interpolation validation
func validateInterpolation(interpolation string) error {
	validOptions := []string{
		"NearestNeighbor",
		"Bilinear",
		"Bicubic",
		"MitchellNetravali",
		"Lanczos2",
		"Lanczos3",
	}
	if !slices.Contains(validOptions, interpolation) {
		return fmt.Errorf("interpolation must be one of: %v", validOptions)
	}
	return nil
}
