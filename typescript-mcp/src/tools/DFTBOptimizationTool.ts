import { ToolDefinition, DFTBOptimizationRequest, DFTBOptimizationResponse } from '../types';
import { CacheService } from '../services/CacheService';

export class DFTBOptimizationTool {
  private goServiceClient: any; // Will be injected
  private cacheService: CacheService | null;

  constructor(goServiceClient: any, cacheService?: CacheService) {
    this.goServiceClient = goServiceClient;
    this.cacheService = cacheService || null;
  }

  /**
   * Get tool definition for MCP protocol
   */
  getDefinition(): ToolDefinition {
    return {
      name: "dftb_optimize",
      description: "Perform DFTB+ geometry optimization on crystal structures. This tool takes a crystal structure file in CIF format and performs geometry optimization using the DFTB+ quantum chemistry package.",
      inputSchema: {
        type: "object",
        properties: {
          structure_file: {
            type: "string",
            description: "Base64 encoded content of the CIF structure file to optimize",
          },
          method: {
            type: "string",
            enum: ["GFN1-xTB", "GFN2-xTB"],
            default: "GFN1-xTB",
            description: "The semi-empirical method to use for the calculation. GFN1-xTB is faster and suitable for most systems, while GFN2-xTB provides better accuracy for certain elements."
          },
          fmax: {
            type: "number",
            default: 0.1,
            description: "Force convergence threshold in eV/Angstrom. Lower values result in more accurate optimizations but require more computational time."
          },
          original_filename: {
            type: "string",
            description: "Optional original filename for the structure (used for logging and identification)"
          }
        },
        required: ["structure_file"]
      }
    };
  }

  /**
   * Execute the optimization tool
   */
  async execute(params: DFTBOptimizationRequest): Promise<DFTBOptimizationResponse> {
    try {
      // Validate input
      if (!params.structure_file) {
        throw new Error("Structure file is required");
      }

      if (!params.method || !["GFN1-xTB", "GFN2-xTB"].includes(params.method)) {
        params.method = "GFN1-xTB";
      }

      if (!params.fmax || params.fmax <= 0) {
        params.fmax = 0.1;
      }

      // Call Go service
      const goResponse = await this.goServiceClient.optimizeStructure(
        params.structure_file,
        params.method,
        params.fmax,
        params.original_filename
      );

      // Transform Go service response to MCP response format
      if (goResponse.status === "success") {
        return {
          status: "success",
          request_id: goResponse.request_id,
          input_parameters: {
            original_filename: params.original_filename,
            method: params.method || "GFN1-xTB",
            fmax_eV_A: params.fmax || 0.1
          },
          detailed_results: goResponse.parsed_data,
          optimized_structure_cif_b64: goResponse.output_cif_path // This will be base64 content from Go service
        };
      } else {
        return {
          status: "error",
          request_id: goResponse.request_id,
          input_parameters: {
            original_filename: params.original_filename,
            method: params.method || "GFN1-xTB",
            fmax_eV_A: params.fmax || 0.1
          },
          error_message: goResponse.error_message || "Unknown error occurred"
        };
      }

    } catch (error) {
      return {
        status: "error",
        request_id: "unknown",
        input_parameters: {
          original_filename: params.original_filename,
          method: params.method || "GFN1-xTB",
          fmax_eV_A: params.fmax || 0.1
        },
        error_message: error instanceof Error ? error.message : "Unknown error occurred"
      };
    }
  }

  /**
   * Helper method to validate CIF content
   */
  private validateCIFContent(base64Content: string): boolean {
    try {
      const decodedContent = Buffer.from(base64Content, 'base64').toString('utf-8');
      
      // Basic validation checks for CIF format
      if (!decodedContent.includes('data_')) {
        return false;
      }
      
      if (!decodedContent.includes('_cell_length_a') && !decodedContent.includes('_cell.angle_alpha')) {
        return false;
      }
      
      if (!decodedContent.includes('_atom_site') && !decodedContent.includes('loop_')) {
        return false;
      }
      
      return true;
    } catch (error) {
      return false;
    }
  }

  /**
   * Helper method to extract filename from base64 content if not provided
   */
  private extractFilename(base64Content: string): string {
    try {
      const decodedContent = Buffer.from(base64Content, 'base64').toString('utf-8');
      const dataMatch = decodedContent.match(/data_(\w+)/);
      return dataMatch ? `${dataMatch[1]}.cif` : 'unknown_structure.cif';
    } catch (error) {
      return 'unknown_structure.cif';
    }
  }
}
