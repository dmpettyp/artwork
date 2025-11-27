package imagegraph

import (
	"fmt"
	"slices"
	"strings"
)

type FieldType string

const (
	FieldTypeInt    FieldType = "int"
	FieldTypeString FieldType = "string"
	FieldTypeFloat  FieldType = "float"
	FieldTypeBool   FieldType = "bool"
	FieldTypeOption FieldType = "option"
	FieldTypeColor  FieldType = "color"
)

// FieldSchema describes a configuration field for API schema generation
type FieldSchema struct {
	Name     string    `json:"name"`
	Type     FieldType `json:"type"`
	Required bool      `json:"required"`
	Options  []string  `json:"options,omitempty"`
	Default  any       `json:"default,omitempty"`
}

type NodeConfig interface {
	Validate() error
	NodeType() NodeType
	Schema() []FieldSchema
}

// Shared options for interpolation fields
var interpolationOptions = []string{
	"NearestNeighbor",
	"Bilinear",
	"Bicubic",
	"MitchellNetravali",
	"Lanczos2",
	"Lanczos3",
}

// Shared options for cluster_by fields
var clusterByOptions = []string{"RGB", "Perceptual"}

func isValidHexColor(color string) bool {
	if len(color) != 7 || color[0] != '#' {
		return false
	}
	for i := 1; i < 7; i++ {
		ch := color[i]
		if !((ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'F') || (ch >= 'a' && ch <= 'f')) {
			return false
		}
	}
	return true
}

func NewNodeConfig(nodeType NodeType) NodeConfig {
	cfg, ok := NodeTypeDefs[nodeType]
	if !ok || cfg.NewConfig == nil {
		return nil
	}
	return cfg.NewConfig()
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

func (c *NodeConfigInput) Schema() []FieldSchema {
	return []FieldSchema{}
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

func (c *NodeConfigOutput) Schema() []FieldSchema {
	return []FieldSchema{}
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

func (c *NodeConfigCrop) Schema() []FieldSchema {
	return []FieldSchema{
		{Name: "left", Type: FieldTypeInt, Required: false},
		{Name: "right", Type: FieldTypeInt, Required: false},
		{Name: "top", Type: FieldTypeInt, Required: false},
		{Name: "bottom", Type: FieldTypeInt, Required: false},
		{Name: "aspect_ratio_width", Type: FieldTypeInt, Required: false},
		{Name: "aspect_ratio_height", Type: FieldTypeInt, Required: false},
	}
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

func (c *NodeConfigBlur) Schema() []FieldSchema {
	return []FieldSchema{
		{Name: "radius", Type: FieldTypeInt, Required: true, Default: 2},
	}
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

	if !slices.Contains(interpolationOptions, c.Interpolation) {
		return fmt.Errorf("interpolation must be one of: %v", interpolationOptions)
	}

	return nil
}

func (c *NodeConfigResize) NodeType() NodeType {
	return NodeTypeResize
}

func (c *NodeConfigResize) Schema() []FieldSchema {
	return []FieldSchema{
		{Name: "width", Type: FieldTypeInt, Required: false},
		{Name: "height", Type: FieldTypeInt, Required: false},
		{Name: "interpolation", Type: FieldTypeOption, Required: true, Options: interpolationOptions},
	}
}

// NodeConfigResizeMatch is the configuration for resize-match nodes.
type NodeConfigResizeMatch struct {
	Interpolation string `json:"interpolation"`
}

func NewNodeConfigResizeMatch() *NodeConfigResizeMatch {
	return &NodeConfigResizeMatch{}
}

func (c *NodeConfigResizeMatch) Validate() error {
	if !slices.Contains(interpolationOptions, c.Interpolation) {
		return fmt.Errorf("interpolation must be one of: %v", interpolationOptions)
	}
	return nil
}

func (c *NodeConfigResizeMatch) NodeType() NodeType {
	return NodeTypeResizeMatch
}

func (c *NodeConfigResizeMatch) Schema() []FieldSchema {
	return []FieldSchema{
		{Name: "interpolation", Type: FieldTypeOption, Required: true, Options: interpolationOptions},
	}
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

	if !isValidHexColor(c.LineColor) {
		return fmt.Errorf("line_color must be in #RRGGBB format")
	}

	return nil
}

func (c *NodeConfigPixelInflate) NodeType() NodeType {
	return NodeTypePixelInflate
}

func (c *NodeConfigPixelInflate) Schema() []FieldSchema {
	return []FieldSchema{
		{Name: "width", Type: FieldTypeInt, Required: true, Default: 500},
		{Name: "line_width", Type: FieldTypeInt, Required: true, Default: 3},
		{Name: "line_color", Type: FieldTypeColor, Required: true, Default: "#FFFFFF"},
	}
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

	if !slices.Contains(clusterByOptions, c.ClusterBy) {
		return fmt.Errorf("cluster_by must be one of: %v", clusterByOptions)
	}

	return nil
}

func (c *NodeConfigPaletteExtract) NodeType() NodeType {
	return NodeTypePaletteExtract
}

func (c *NodeConfigPaletteExtract) Schema() []FieldSchema {
	return []FieldSchema{
		{Name: "num_colors", Type: FieldTypeInt, Required: true, Default: 16},
		{Name: "cluster_by", Type: FieldTypeOption, Required: true, Options: clusterByOptions, Default: "RGB"},
	}
}

// NodeConfigPaletteApply is the configuration for palette-apply nodes.
type NodeConfigPaletteApply struct {
	Normalize string `json:"normalize"`
}

func NewNodeConfigPaletteApply() *NodeConfigPaletteApply {
	return &NodeConfigPaletteApply{Normalize: "none"}
}

func (c *NodeConfigPaletteApply) Validate() error {
	if c.Normalize == "" {
		c.Normalize = "none"
	}
	if !slices.Contains([]string{"none", "lightness"}, c.Normalize) {
		return fmt.Errorf("normalize must be one of: none, lightness")
	}
	return nil
}

func (c *NodeConfigPaletteApply) NodeType() NodeType {
	return NodeTypePaletteApply
}

func (c *NodeConfigPaletteApply) Schema() []FieldSchema {
	return []FieldSchema{
		{Name: "normalize", Type: FieldTypeOption, Required: false, Options: []string{"none", "lightness"}, Default: "none"},
	}
}

// parseColorsList splits a comma-separated string, trims whitespace, and
// validates each entry is a #RRGGBB color.
func parseColorsList(list string) ([]string, error) {
	raw := strings.Split(list, ",")
	parts := make([]string, 0, len(raw))
	for _, part := range raw {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}

	for _, c := range parts {
		val := strings.TrimPrefix(c, "!")
		if !isValidHexColor(val) {
			return nil, fmt.Errorf("color %q must be in #RRGGBB format (prefix ! allowed to disable)", c)
		}
	}

	return parts, nil
}

// NodeConfigPaletteCreate is the configuration for palette-create nodes.
type NodeConfigPaletteCreate struct {
	Colors string `json:"colors"`
}

func NewNodeConfigPaletteCreate() *NodeConfigPaletteCreate {
	return &NodeConfigPaletteCreate{}
}

func (c *NodeConfigPaletteCreate) Validate() error {
	_, err := parseColorsList(c.Colors)
	return err
}

func (c *NodeConfigPaletteCreate) NodeType() NodeType {
	return NodeTypePaletteCreate
}

func (c *NodeConfigPaletteCreate) Schema() []FieldSchema {
	return []FieldSchema{
		{Name: "colors", Type: FieldTypeString, Required: true},
	}
}

// ColorsList returns the parsed list of colors from the config.
func (c *NodeConfigPaletteCreate) ColorsList() ([]string, error) {
	all, err := parseColorsList(c.Colors)
	if err != nil {
		return nil, err
	}

	enabled := make([]string, 0, len(all))
	for _, col := range all {
		if strings.HasPrefix(col, "!") {
			continue
		}
		enabled = append(enabled, col)
	}

	return enabled, nil
}
