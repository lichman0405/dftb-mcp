import express, { Request, Response } from 'express';
import { Server as WebSocketServer } from 'ws';
import { createServer } from 'http';
import { v4 as uuidv4 } from 'uuid';
import * as YAML from 'yamljs';
import { serve, setup } from 'swagger-ui-express';
import { 
  MCPRequest, 
  MCPResponse, 
  MCPError, 
  ServerConfig,
  DFTBOptimizationRequest,
  DFTBOptimizationResponse,
  CacheConfig 
} from '../types';
import { GoServiceClient } from '../client/GoServiceClient';
import { DFTBOptimizationTool } from '../tools/DFTBOptimizationTool';
import { CacheService } from '../services/CacheService';

export class MCPServer {
  private app: express.Application;
  private httpServer: any;
  private wss: WebSocketServer;
  private config: ServerConfig;
  private goServiceClient: GoServiceClient;
  private dftbTool: DFTBOptimizationTool;
  private tools: Map<string, DFTBOptimizationTool>;
  private cacheService: CacheService | null;

  constructor(config: ServerConfig) {
    this.config = config;
    this.app = express();
    this.httpServer = createServer(this.app);
    this.wss = new WebSocketServer({ server: this.httpServer });
    this.goServiceClient = new GoServiceClient(config.goServiceUrl);
    
    // Initialize cache service if enabled
    if (config.enableCache && config.redisUrl) {
      const cacheConfig: CacheConfig = {
        enabled: true,
        url: config.redisUrl,
        ttl: config.cacheTTL || 86400, // 24 hours default
        prefix: 'dftbopt'
      };
      this.cacheService = new CacheService(cacheConfig);
    } else {
      this.cacheService = null;
    }
    
    this.dftbTool = new DFTBOptimizationTool(this.goServiceClient, this.cacheService || undefined);
    this.tools = new Map();
    
    this.setupMiddleware();
    this.setupRoutes();
    this.setupWebSocket();
    this.registerTools();
  }

  /**
   * Setup Express middleware
   */
  private setupMiddleware(): void {
    this.app.use(express.json({ limit: this.config.maxRequestSize }));
    this.app.use(express.urlencoded({ extended: true, limit: this.config.maxRequestSize }));
    
    // CORS headers
    this.app.use((req, res, next) => {
      res.header('Access-Control-Allow-Origin', '*');
      res.header('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS');
      res.header('Access-Control-Allow-Headers', 'Origin, X-Requested-With, Content-Type, Accept, Authorization');
      if (req.method === 'OPTIONS') {
        res.sendStatus(200);
      } else {
        next();
      }
    });

    // Request logging
    this.app.use((req, res, next) => {
      console.log(`${new Date().toISOString()} - ${req.method} ${req.path}`);
      next();
    });
  }

  /**
   * Setup HTTP routes
   */
  private setupRoutes(): void {
    // Load OpenAPI specification
    const openApiSpec = YAML.load(__dirname + '/../../api-docs.yaml');
    
    // Serve Swagger UI
    this.app.use('/api-docs', serve);
    this.app.get('/api-docs', setup(openApiSpec));
    
    // Serve OpenAPI JSON
    this.app.get('/api-docs.json', (req: Request, res: Response) => {
      res.setHeader('Content-Type', 'application/json');
      res.send(openApiSpec);
    });

    // Health check endpoint
    this.app.get('/health', (req: Request, res: Response) => {
      res.json({
        status: 'healthy',
        timestamp: new Date().toISOString(),
        version: '1.0.0'
      });
    });

    // MCP tools list endpoint
    this.app.get('/api/v1/tools', (req: Request, res: Response) => {
      const tools = Array.from(this.tools.values()).map(tool => tool.getDefinition());
      res.json({ tools });
    });

    // MCP HTTP endpoint (for REST-based MCP communication)
    this.app.post('/api/v1/mcp', async (req: Request, res: Response) => {
      try {
        const mcpRequest: MCPRequest = req.body;
        const response = await this.handleMCPRequest(mcpRequest);
        res.json(response);
      } catch (error) {
        const errorResponse: MCPResponse = {
          jsonrpc: '2.0',
          id: req.body.id || 'unknown',
          error: {
            code: -32603,
            message: 'Internal error',
            data: error instanceof Error ? error.message : String(error)
          }
        };
        res.status(500).json(errorResponse);
      }
    });

    // Direct optimization endpoint (simplified)
    this.app.post('/api/v1/optimize', async (req: Request, res: Response) => {
      try {
        const request: DFTBOptimizationRequest = req.body;
        const response = await this.dftbTool.execute(request);
        res.json(response);
      } catch (error) {
        res.status(500).json({
          status: 'error',
          error_message: error instanceof Error ? error.message : 'Unknown error'
        });
      }
    });

    // Root endpoint
    this.app.get('/', (req: Request, res: Response) => {
      res.json({
        message: 'DFTB+ MCP Server',
        version: '1.0.0',
        endpoints: {
          health: '/health',
          tools: '/api/v1/tools',
          mcp: '/api/v1/mcp',
          optimize: '/api/v1/optimize'
        }
      });
    });
  }

  /**
   * Setup WebSocket server for real-time MCP communication
   */
  private setupWebSocket(): void {
    this.wss.on('connection', (ws) => {
      console.log('WebSocket client connected');
      
      ws.on('message', async (data) => {
        try {
          const mcpRequest: MCPRequest = JSON.parse(data.toString());
          const response = await this.handleMCPRequest(mcpRequest);
          ws.send(JSON.stringify(response));
        } catch (error) {
          const errorResponse: MCPResponse = {
            jsonrpc: '2.0',
            id: mcpRequest?.id || 'unknown',
            error: {
              code: -32700,
              message: 'Parse error',
              data: error instanceof Error ? error.message : String(error)
            }
          };
          ws.send(JSON.stringify(errorResponse));
        }
      });

      ws.on('close', () => {
        console.log('WebSocket client disconnected');
      });

      ws.on('error', (error) => {
        console.error('WebSocket error:', error);
      });

      // Send welcome message
      ws.send(JSON.stringify({
        jsonrpc: '2.0',
        id: null,
        result: {
          message: 'Connected to DFTB+ MCP Server',
          tools: Array.from(this.tools.values()).map(tool => tool.getDefinition())
        }
      }));
    });
  }

  /**
   * Register available tools
   */
  private registerTools(): void {
    this.tools.set('dftb_optimize', this.dftbTool);
    console.log(`Registered ${this.tools.size} tools`);
  }

  /**
   * Handle MCP request and route to appropriate tool
   */
  private async handleMCPRequest(request: MCPRequest): Promise<MCPResponse> {
    // Validate MCP request format
    if (!request.jsonrpc || request.jsonrpc !== '2.0') {
      throw new Error('Invalid JSON-RPC version');
    }

    if (!request.method) {
      throw new Error('Method is required');
    }

    if (!request.id) {
      throw new Error('Request ID is required');
    }

    // Handle tool calls
    if (request.method === 'tools/call') {
      const { name, arguments: args } = request.params;
      
      if (!name) {
        return this.createErrorResponse(request.id, -32602, 'Tool name is required');
      }

      const tool = this.tools.get(name);
      if (!tool) {
        return this.createErrorResponse(request.id, -32601, `Tool '${name}' not found`);
      }

      try {
        const result = await tool.execute(args);
        return {
          jsonrpc: '2.0',
          id: request.id,
          result
        };
      } catch (error) {
        return this.createErrorResponse(
          request.id, 
          -32603, 
          'Tool execution failed', 
          error instanceof Error ? error.message : String(error)
        );
      }
    }

    // Handle tools list
    if (request.method === 'tools/list') {
      const tools = Array.from(this.tools.values()).map(tool => tool.getDefinition());
      return {
        jsonrpc: '2.0',
        id: request.id,
        result: { tools }
      };
    }

    // Unknown method
    return this.createErrorResponse(request.id, -32601, 'Method not found');
  }

  /**
   * Create MCP error response
   */
  private createErrorResponse(id: string | number, code: number, message: string, data?: any): MCPResponse {
    const error: MCPError = { code, message };
    if (data) {
      error.data = data;
    }
    
    return {
      jsonrpc: '2.0',
      id,
      error
    };
  }

  /**
   * Start the server
   */
  public start(): void {
    this.httpServer.listen(this.config.port, () => {
      console.log(`DFTB+ MCP Server running on port ${this.config.port}`);
      console.log(`Go service URL: ${this.config.goServiceUrl}`);
      console.log(`WebSocket support: ${this.config.enableWebSocket ? 'enabled' : 'disabled'}`);
    });
  }

  /**
   * Stop the server
   */
  public stop(): void {
    this.wss.close();
    this.httpServer.close();
    console.log('DFTB+ MCP Server stopped');
  }

  /**
   * Get server configuration
   */
  public getConfig(): ServerConfig {
    return { ...this.config };
  }
}
