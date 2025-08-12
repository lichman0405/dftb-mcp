import axios, { AxiosInstance } from 'axios';
import { v4 as uuidv4 } from 'uuid';
import { GoServiceRequest, GoServiceResponse } from '../types';

export class GoServiceClient {
  private client: AxiosInstance;
  private baseUrl: string;

  constructor(baseUrl: string = 'http://localhost:8080') {
    this.baseUrl = baseUrl;
    this.client = axios.create({
      baseURL: this.baseUrl,
      timeout: 300000, // 5 minutes timeout for long calculations
      headers: {
        'Content-Type': 'application/json',
      },
    });
  }

  /**
   * Send optimization request to Go service
   */
  async optimizeStructure(
    structureFile: string,
    method: 'GFN1-xTB' | 'GFN2-xTB' = 'GFN1-xTB',
    fmax: number = 0.1,
    originalFilename?: string
  ): Promise<GoServiceResponse> {
    const requestId = uuidv4();
    
    const request: GoServiceRequest = {
      request_id: requestId,
      structure_file: structureFile,
      method,
      fmax,
      original_filename: originalFilename,
    };

    try {
      const response = await this.client.post<GoServiceResponse>('/api/v1/optimize', request);
      return response.data;
    } catch (error) {
      if (axios.isAxiosError(error)) {
        throw new Error(`Go service request failed: ${error.response?.data?.error || error.message}`);
      }
      throw new Error(`Unexpected error: ${error instanceof Error ? error.message : String(error)}`);
    }
  }

  /**
   * Health check for Go service
   */
  async healthCheck(): Promise<{ status: string; timestamp: string }> {
    try {
      const response = await this.client.get('/health');
      return response.data;
    } catch (error) {
      throw new Error(`Health check failed: ${error instanceof Error ? error.message : String(error)}`);
    }
  }

  /**
   * Get service information
   */
  async getServiceInfo(): Promise<{ name: string; version: string; capabilities: string[] }> {
    try {
      const response = await this.client.get('/info');
      return response.data;
    } catch (error) {
      throw new Error(`Failed to get service info: ${error instanceof Error ? error.message : String(error)}`);
    }
  }

  /**
   * Update base URL
   */
  setBaseUrl(url: string): void {
    this.baseUrl = url;
    this.client.defaults.baseURL = url;
  }
}
