#!/usr/bin/env node

/**
 * DFTB+ MCP Server Entry Point
 * This is the main entry point for the MCP server that handles DFTB+ optimization requests
 */

import { MCPServer } from './server/MCPServer';
import { ServerConfig } from './types';

// Server configuration
const config: ServerConfig = {
  port: parseInt(process.env.PORT || '3000'),
  goServiceUrl: process.env.GO_SERVICE_URL || 'http://localhost:8080',
  enableWebSocket: process.env.ENABLE_WEBSOCKET !== 'false',
  maxRequestSize: process.env.MAX_REQUEST_SIZE || '50mb'
};

// Create and start the server
const server = new MCPServer(config);

// Graceful shutdown handling
process.on('SIGINT', () => {
  console.log('Received SIGINT, shutting down gracefully...');
  server.stop();
  process.exit(0);
});

process.on('SIGTERM', () => {
  console.log('Received SIGTERM, shutting down gracefully...');
  server.stop();
  process.exit(0);
});

// Handle uncaught exceptions
process.on('uncaughtException', (error) => {
  console.error('Uncaught Exception:', error);
  server.stop();
  process.exit(1);
});

process.on('unhandledRejection', (reason, promise) => {
  console.error('Unhandled Rejection at:', promise, 'reason:', reason);
  server.stop();
  process.exit(1);
});

// Start the server
server.start();

console.log('🚀 DFTB+ MCP Server started successfully!');
console.log('📋 Configuration:', JSON.stringify(config, null, 2));
console.log('🔧 Environment Variables:');
console.log(`   PORT: ${config.port}`);
console.log(`   GO_SERVICE_URL: ${config.goServiceUrl}`);
console.log(`   ENABLE_WEBSOCKET: ${config.enableWebSocket}`);
console.log(`   MAX_REQUEST_SIZE: ${config.maxRequestSize}`);
console.log('📖 Available endpoints:');
console.log('   GET  /health        - Health check');
console.log('   GET  /              - Server information');
console.log('   GET  /api/v1/tools  - List available tools');
console.log('   POST /api/v1/mcp    - MCP protocol endpoint');
console.log('   POST /api/v1/optimize - Direct optimization endpoint');
console.log('🔌 WebSocket support available on the same port');
