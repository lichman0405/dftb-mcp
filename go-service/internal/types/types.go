package types

// OptimizationRequest represents the request for DFTB+ optimization
type OptimizationRequest struct {
	RequestID       string  `json:"request_id" binding:"required"`
	StructureFile   string  `json:"structure_file" binding:"required"`   // Base64 encoded CIF content
	Method          string  `json:"method" binding:"required"`           // "GFN1-xTB" or "GFN2-xTB"
	Fmax            float64 `json:"fmax" binding:"required,min=0.001"`   // Force convergence threshold
	OriginalFilename string `json:"original_filename,omitempty"`         // Optional original filename
}

// OptimizationResponse represents the response from DFTB+ optimization
type OptimizationResponse struct {
	Status        string                 `json:"status"`                   // "success" or "error"
	RequestID     string                 `json:"request_id"`
	ParsedData    map[string]interface{} `json:"parsed_data,omitempty"`    // Parsed DFTB+ output
	OutputCIFPath string                 `json:"output_cif_path,omitempty"` // Path to optimized CIF file (base64 encoded)
	ErrorMessage  string                 `json:"error_message,omitempty"`   // Error message if failed
}

// DFTBOutput represents the parsed output from DFTB+ calculation
type DFTBOutput struct {
	Summary struct {
		Warnings           []string `json:"warnings"`
		ConvergenceStatus  string   `json:"convergence_status"`
		CalculationStatus  string   `json:"calculation_status"`
		Error              string   `json:"error,omitempty"`
	} `json:"summary"`
	
	ConvergenceInfo struct {
		SCCConverged bool `json:"scc_converged"`
	} `json:"convergence_info"`
	
	ElectronicProperties struct {
		FermiLevelEV      float64 `json:"fermi_level_eV,omitempty"`
		TotalCharge       float64 `json:"total_charge,omitempty"`
		DipoleMomentDebye struct {
			X float64 `json:"x"`
			Y float64 `json:"y"`
			Z float64 `json:"z"`
		} `json:"dipole_moment_debye,omitempty"`
	} `json:"electronic_properties,omitempty"`
	
	EnergiesEV     map[string]float64 `json:"energies_eV"`
	EnergiesHartree map[string]float64 `json:"energies_hartree"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}

// ServiceInfo represents the service information
type ServiceInfo struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Capabilities []string `json:"capabilities"`
}

// ServerConfig represents the server configuration
type ServerConfig struct {
	Port         int    `json:"port"`
	WorkDir      string `json:"work_dir"`
	DFTBPath     string `json:"dftb_path"`
	MaxRequests  int    `json:"max_requests"`
	Timeout      int    `json:"timeout"` // in seconds
}

// CIFFile represents a parsed CIF file structure
type CIFFile struct {
	DataBlock struct {
		Name        string                 `json:"name"`
		CellLength  map[string]float64     `json:"cell_length"`
		CellAngle   map[string]float64     `json:"cell_angle"`
		AtomSites   []AtomSite            `json:"atom_sites"`
		Symmetry    []SymmetryOperation   `json:"symmetry,omitempty"`
		Metadata    map[string]string     `json:"metadata,omitempty"`
	} `json:"data_block"`
}

// AtomSite represents an atomic site in CIF format
type AtomSite struct {
	Label        string  `json:"label"`
	TypeSymbol   string  `json:"type_symbol"`
	FractX       float64 `json:"fract_x"`
	FractY       float64 `json:"fract_y"`
	FractZ       float64 `json:"fract_z"`
	UIsoOrEquiv  float64 `json:"u_iso_or_equiv,omitempty"`
	AdpType      string  `json:"adp_type,omitempty"`
}

// SymmetryOperation represents a symmetry operation in CIF format
type SymmetryOperation struct {
	X string `json:"x"`
	Y string `json:"y"`
	Z string `json:"z"`
}

// DFTBInput represents the input for DFTB+ calculation
type DFTBInput struct {
	Geometry struct {
		Periodic      bool     `json:"periodic"`
		LatticeVectors [3][3]float64 `json:"lattice_vectors"`
		Elements      []string `json:"elements"`
		Coordinates   [][]float64 `json:"coordinates"`
	} `json:"geometry"`
	
	Hamiltonian struct {
		Method string `json:"method"` // "GFN1-xTB" or "GFN2-xTB"
	} `json:"hamiltonian"`
	
	Analysis struct {
		Forces bool `json:"forces"`
	} `json:"analysis"`
	
	Options struct {
		Fmax float64 `json:"fmax"` // Force convergence threshold
	} `json:"options"`
}
