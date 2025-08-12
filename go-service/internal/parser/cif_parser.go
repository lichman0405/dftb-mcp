package parser

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"dftbopt-mcp/go-service/internal/types"
)

// CIFParser handles parsing of CIF (Crystallographic Information File) format
type CIFParser struct{}

// NewCIFParser creates a new CIF parser instance
func NewCIFParser() *CIFParser {
	return &CIFParser{}
}

// ParseFromBase64 parses a CIF file from base64 encoded string
func (p *CIFParser) ParseFromBase64(base64Content string) (*types.CIFFile, error) {
	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(base64Content)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 content: %v", err)
	}

	return p.ParseFromString(string(decoded))
}

// ParseFromString parses a CIF file from string content
func (p *CIFParser) ParseFromString(content string) (*types.CIFFile, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	
	cif := &types.CIFFile{}
	var currentDataBlock *types.CIFFile
	var inLoop bool
	var loopHeaders []string
	var loopData [][]string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle data blocks
		if strings.HasPrefix(line, "data_") {
			// Save previous data block if exists
			if currentDataBlock != nil {
				cif.DataBlock = *currentDataBlock
			}
			
			// Start new data block
			currentDataBlock = &types.CIFFile{}
			currentDataBlock.DataBlock.Name = strings.TrimPrefix(line, "data_")
			currentDataBlock.DataBlock.CellLength = make(map[string]float64)
			currentDataBlock.DataBlock.CellAngle = make(map[string]float64)
			currentDataBlock.DataBlock.Metadata = make(map[string]string)
			inLoop = false
			loopHeaders = nil
			loopData = nil
			continue
		}

		// Handle loops
		if strings.HasPrefix(line, "loop_") {
			inLoop = true
			loopHeaders = nil
			loopData = nil
			continue
		}

		if inLoop {
			// Collect loop headers
			if strings.HasPrefix(line, "_") {
				loopHeaders = append(loopHeaders, line)
				continue
			}

			// Collect loop data
			if !strings.HasPrefix(line, "_") && line != "" {
				fields := strings.Fields(line)
				if len(fields) == len(loopHeaders) {
					loopData = append(loopData, fields)
				}
				continue
			}

			// End of loop
			inLoop = false
			p.processLoopData(currentDataBlock, loopHeaders, loopData)
			continue
		}

		// Handle key-value pairs
		if strings.Contains(line, " ") && !strings.HasPrefix(line, "_") {
			parts := strings.SplitN(line, " ", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				
				// Remove quotes if present
				if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
					value = strings.Trim(value, "\"")
				} else if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
					value = strings.Trim(value, "'")
				}
				
				currentDataBlock.DataBlock.Metadata[key] = value
			}
			continue
		}

		// Handle CIF keywords
		if strings.HasPrefix(line, "_") {
			parts := strings.SplitN(line, " ", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				
				// Remove quotes if present
				if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
					value = strings.Trim(value, "\"")
				} else if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
					value = strings.Trim(value, "'")
				}
				
				switch key {
				case "_cell_length_a", "_cell_length_b", "_cell_length_c":
					if val, err := strconv.ParseFloat(value, 64); err == nil {
						currentDataBlock.DataBlock.CellLength[key] = val
					}
				case "_cell_angle_alpha", "_cell_angle_beta", "_cell_angle_gamma":
					if val, err := strconv.ParseFloat(value, 64); err == nil {
						currentDataBlock.DataBlock.CellAngle[key] = val
					}
				}
			}
		}
	}

	// Save the last data block
	if currentDataBlock != nil {
		cif.DataBlock = *currentDataBlock
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading CIF content: %v", err)
	}

	return cif, nil
}

// processLoopData processes loop data and populates atom sites
func (p *CIFParser) processLoopData(dataBlock *types.CIFFile, headers []string, data [][]string) {
	// Check if this is an atom site loop
	atomSiteHeaders := []string{
		"_atom_site_label",
		"_atom_site_type_symbol",
		"_atom_site_fract_x",
		"_atom_site_fract_y",
		"_atom_site_fract_z",
	}

	if p.containsAllHeaders(headers, atomSiteHeaders) {
		for _, row := range data {
			atomSite := types.AtomSite{}
			
			for i, header := range headers {
				if i >= len(row) {
					continue
				}
				
				value := strings.TrimSpace(row[i])
				
				switch header {
				case "_atom_site_label":
					atomSite.Label = value
				case "_atom_site_type_symbol":
					atomSite.TypeSymbol = value
				case "_atom_site_fract_x":
					if val, err := strconv.ParseFloat(value, 64); err == nil {
						atomSite.FractX = val
					}
				case "_atom_site_fract_y":
					if val, err := strconv.ParseFloat(value, 64); err == nil {
						atomSite.FractY = val
					}
				case "_atom_site_fract_z":
					if val, err := strconv.ParseFloat(value, 64); err == nil {
						atomSite.FractZ = val
					}
				case "_atom_site_u_iso_or_equiv":
					if val, err := strconv.ParseFloat(value, 64); err == nil {
						atomSite.UIsoOrEquiv = val
					}
				case "_atom_site_adp_type":
					atomSite.AdpType = value
				}
			}
			
			dataBlock.DataBlock.AtomSites = append(dataBlock.DataBlock.AtomSites, atomSite)
		}
	}

	// Check if this is a symmetry operation loop
	symmetryHeaders := []string{
		"_symmetry_equiv_pos_as_xyz_x",
		"_symmetry_equiv_pos_as_xyz_y",
		"_symmetry_equiv_pos_as_xyz_z",
	}

	if p.containsAllHeaders(headers, symmetryHeaders) {
		for _, row := range data {
			symmetry := types.SymmetryOperation{}
			
			for i, header := range headers {
				if i >= len(row) {
					continue
				}
				
				value := strings.TrimSpace(row[i])
				
				switch header {
				case "_symmetry_equiv_pos_as_xyz_x":
					symmetry.X = value
				case "_symmetry_equiv_pos_as_xyz_y":
					symmetry.Y = value
				case "_symmetry_equiv_pos_as_xyz_z":
					symmetry.Z = value
				}
			}
			
			dataBlock.DataBlock.Symmetry = append(dataBlock.DataBlock.Symmetry, symmetry)
		}
	}
}

// containsAllHeaders checks if all required headers are present
func (p *CIFParser) containsAllHeaders(headers, required []string) bool {
	headerMap := make(map[string]bool)
	for _, h := range headers {
		headerMap[h] = true
	}
	
	for _, req := range required {
		if !headerMap[req] {
			return false
		}
	}
	
	return true
}

// ToDFTBInput converts CIF file to DFTB+ input format
func (p *CIFParser) ToDFTBInput(cif *types.CIFFile, method string, fmax float64) (*types.DFTBInput, error) {
	if cif == nil || cif.DataBlock.Name == "" {
		return nil, fmt.Errorf("invalid CIF file")
	}

	input := &types.DFTBInput{}
	
	// Set geometry
	input.Geometry.Periodic = true
	
	// Set lattice vectors (simplified - assuming cubic cell for now)
	a := cif.DataBlock.CellLength["_cell_length_a"]
	b := cif.DataBlock.CellLength["_cell_length_b"]
	c := cif.DataBlock.CellLength["_cell_length_c"]
	
	alpha := cif.DataBlock.CellAngle["_cell_angle_alpha"] * 3.141592653589793 / 180.0
	beta := cif.DataBlock.CellAngle["_cell_angle_beta"] * 3.141592653589793 / 180.0
	gamma := cif.DataBlock.CellAngle["_cell_angle_gamma"] * 3.141592653589793 / 180.0
	
	// Simplified lattice vector calculation (for demonstration)
	// In a real implementation, you would need proper crystallographic calculations
	input.Geometry.LatticeVectors = [3][3]float64{
		{a, 0, 0},
		{0, b, 0},
		{0, 0, c},
	}
	
	// Extract elements and coordinates
	elementMap := make(map[string]bool)
	for _, atom := range cif.DataBlock.AtomSites {
		if !elementMap[atom.TypeSymbol] {
			input.Geometry.Elements = append(input.Geometry.Elements, atom.TypeSymbol)
			elementMap[atom.TypeSymbol] = true
		}
		
		// Convert fractional to Cartesian coordinates (simplified)
		x := atom.FractX * a
		y := atom.FractY * b
		z := atom.FractZ * c
		
		input.Geometry.Coordinates = append(input.Geometry.Coordinates, []float64{x, y, z})
	}
	
	// Set Hamiltonian method
	input.Hamiltonian.Method = method
	
	// Enable force calculation
	input.Analysis.Forces = true
	
	// Set convergence threshold
	input.Options.Fmax = fmax
	
	return input, nil
}

// SaveToFile saves CIF content to a file
func (p *CIFParser) SaveToFile(content string, filename string) (string, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %v", err)
	}
	
	// Write content to file
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %v", err)
	}
	
	return filename, nil
}

// ReadFromFile reads CIF content from a file
func (p *CIFParser) ReadFromFile(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}
	
	return string(content), nil
}

// ValidateCIFContent validates basic CIF format
func (p *CIFParser) ValidateCIFContent(content string) error {
	if !strings.Contains(content, "data_") {
		return fmt.Errorf("missing data block declaration")
	}
	
	if !strings.Contains(content, "_cell_length_a") && !strings.Contains(content, "_cell.angle_alpha") {
		return fmt.Errorf("missing cell parameters")
	}
	
	if !strings.Contains(content, "_atom_site") && !strings.Contains(content, "loop_") {
		return fmt.Errorf("missing atom site information")
	}
	
	return nil
}
