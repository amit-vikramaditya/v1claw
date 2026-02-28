//go:build !linux

package tools

// transfer is a stub for non-Linux platforms.
func (t *SPITool) transfer(tc ToolContext, args map[string]interface{}) *ToolResult {
	return ErrorResult("SPI is only supported on Linux")
}

// readDevice is a stub for non-Linux platforms.
func (t *SPITool) readDevice(tc ToolContext, args map[string]interface{}) *ToolResult {
	return ErrorResult("SPI is only supported on Linux")
}
