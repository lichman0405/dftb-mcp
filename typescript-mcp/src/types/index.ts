// MCP Protocol Types
export interface MCPRequest {
  jsonrpc: "2.0";
  id: string | number;
  method: string;
  params?: any;
}

export interface MCPResponse {
  jsonrpc: "2.0";
  id: string | number;
  result?: any;
  error?: MCPError;
}

export interface MCPError {
  code: number;
  message: string;
  data?: any;
}

// Tool Definitions
export interface ToolDefinition {
  name: string;
  description: string;
  inputSchema: {
    type: "object";
    properties: Record<string, any>;
    required: string[];
  };
}

// DFTB Optimization Types
export interface DFTBOptimizationRequest {
  structure_file: string; // Base64 encoded CIF content
  method?: "GFN1-xTB" | "GFN2-xTB";
  fmax?: number;
  original_filename?: string;
}

export interface DFTBOptimizationResponse {
  status: "success" | "error";
  request_id: string;
  input_parameters: {
    original_filename?: string;
    method: string;
    fmax_eV_A: number;
  };
  detailed_results?: {
    summary: {
      warnings: string[];
      convergence_status: string;
      calculation_status: string;
      error?: string;
    };
    convergence_info: {
      scc_converged: boolean;
    };
    electronic_properties?: {
      fermi_level_eV?: number;
      total_charge?: number;
      dipole_moment_debye?: {
        x: number;
        y: number;
        z: number;
      };
    };
    energies_eV: Record<string, number>;
    energies_hartree: Record<string, number>;
  };
  optimized_structure_cif_b64?: string;
  error_message?: string;
}

// Go Service API Types
export interface GoServiceRequest {
  request_id: string;
  structure_file: string; // Base64 encoded
  method: "GFN1-xTB" | "GFN2-xTB";
  fmax: number;
  original_filename?: string;
}

export interface GoServiceResponse {
  status: "success" | "error";
  request_id: string;
  parsed_data?: any;
  output_cif_path?: string;
  error_message?: string;
}

// Server Configuration
export interface ServerConfig {
  port: number;
  goServiceUrl: string;
  enableWebSocket: boolean;
  maxRequestSize: number;
  redisUrl?: string;
  cacheTTL?: number;
  enableCache?: boolean;
}

// Cache Configuration
export interface CacheConfig {
  enabled: boolean;
  url: string;
  ttl: number; // in seconds
  prefix: string;
}

// Cache Entry
export interface CacheEntry {
  key: string;
  value: DFTBOptimizationResponse;
  timestamp: number;
  ttl: number;
}

// Cache Stats
export interface CacheStats {
  hits: number;
  misses: number;
  size: number;
}
